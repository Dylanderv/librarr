"""API FastAPI du sidecar Telegram (read-only + queue 1 à la fois)."""

import asyncio
from contextlib import asynccontextmanager

from fastapi import Depends, FastAPI, HTTPException, Query
from fastapi.responses import HTMLResponse

from . import store
from . import telegram_service as tg
from .config import settings
from .schemas import (
    CodeConfirmIn,
    CodeRequestIn,
    EnqueueDownloadIn,
    JobOut,
)
from .security import require_api_key


worker_task: asyncio.Task | None = None
stop_event = asyncio.Event()
# Jobs annulés en mémoire : le worker vérifie avant/après et pendant le download.
cancelled: set[str] = set()


async def worker_loop():
    """Consomme la file en FIFO, 1 download actif, backoff sur FloodWait."""
    from telethon.errors import FloodWaitError

    while not stop_event.is_set():
        try:
            job = await store.next_queued(settings.telegram_db_path)
            if not job:
                await asyncio.sleep(settings.poll_interval_seconds)
                continue
            if job["id"] in cancelled:
                await store.update_job(
                    settings.telegram_db_path, job["id"], status="cancelled"
                )
                continue
            await store.update_job(
                settings.telegram_db_path, job["id"], status="downloading", progress=0.0
            )

            def _cb(done, total):
                if job["id"] in cancelled:
                    raise tg.CancelledDownload()
                # Push progress sans bloquer (fire-and-forget via tâche).
                asyncio.create_task(
                    store.update_job(
                        settings.telegram_db_path,
                        job["id"],
                        progress=(done / total if total else 0),
                    )
                )

            try:
                res = await tg.download_attachment(
                    job["dialog_id"],
                    job["message_id"],
                    job["category"],
                    progress_cb=_cb,
                )
                if job["id"] in cancelled:
                    # Annulé pendant le download : le partiel est déjà supprimé.
                    await store.update_job(
                        settings.telegram_db_path, job["id"], status="cancelled"
                    )
                else:
                    await store.update_job(
                        settings.telegram_db_path,
                        job["id"],
                        status="done",
                        progress=1.0,
                        filename=res["filename"],
                        size_bytes=res["size_bytes"],
                        dest_path=res["dest_path"],
                        error=None,
                    )
            except tg.CancelledDownload:
                await store.update_job(
                    settings.telegram_db_path, job["id"], status="cancelled"
                )
            except FloodWaitError as e:
                # Requeue : reste "queued", retry après la pénalité.
                await store.update_job(
                    settings.telegram_db_path,
                    job["id"],
                    status="queued",
                    error=f"flood_wait_{e.seconds}s",
                )
                await asyncio.sleep(e.seconds + 5)
            except asyncio.CancelledError:
                await store.update_job(
                    settings.telegram_db_path, job["id"], status="cancelled"
                )
                raise
            except Exception as e:
                await store.update_job(
                    settings.telegram_db_path,
                    job["id"],
                    status="error",
                    error=str(e)[:500],
                )
        except asyncio.CancelledError:
            break
        except Exception:
            await asyncio.sleep(settings.poll_interval_seconds)


@asynccontextmanager
async def lifespan(app: FastAPI):
    global worker_task
    await store.init_db(settings.telegram_db_path)
    await store.reset_stuck(settings.telegram_db_path)
    stop_event.clear()
    worker_task = asyncio.create_task(worker_loop())
    yield
    stop_event.set()
    if worker_task:
        worker_task.cancel()


app = FastAPI(title="telegram-downloader (sidecar Librarr)", lifespan=lifespan)


@app.get("/health")
async def health():
    return {"ok": True}


@app.get("/", response_class=HTMLResponse, include_in_schema=False)
async def index():
    from pathlib import Path

    return (Path(__file__).parent / "static" / "index.html").read_text()


@app.get("/auth/status")
async def auth_status(_: None = Depends(require_api_key)):
    try:
        return await tg.auth_status()
    except RuntimeError as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.post("/auth/code/request")
async def auth_request(body: CodeRequestIn, _: None = Depends(require_api_key)):
    try:
        await tg.request_code(body.phone)
        return {"sent": True}
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e)[:300])


@app.post("/auth/code/confirm")
async def auth_confirm(body: CodeConfirmIn, _: None = Depends(require_api_key)):
    try:
        return await tg.confirm_code(body.phone, body.code, body.password)
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e)[:300])


@app.get("/dialogs")
async def dialogs(_: None = Depends(require_api_key)):
    try:
        return await tg.list_dialogs()
    except RuntimeError as e:
        raise HTTPException(status_code=401, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e)[:300])


@app.get("/dialogs/{dialog_id}/search")
async def search(
    dialog_id: int,
    q: str = Query(default=""),
    limit: int = Query(default=50, le=200),
    _: None = Depends(require_api_key),
):
    try:
        return await tg.search_attachments(dialog_id, q, limit)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e)[:300])


@app.post("/downloads", response_model=JobOut)
async def enqueue_dl(body: EnqueueDownloadIn, _: None = Depends(require_api_key)):
    job = await store.enqueue(
        settings.telegram_db_path, body.dialog_id, body.message_id, body.category.value
    )
    return job


@app.get("/downloads", response_model=list[JobOut])
async def list_dl(_: None = Depends(require_api_key)):
    return await store.list_jobs(settings.telegram_db_path)


@app.get("/downloads/{job_id}", response_model=JobOut)
async def get_dl(job_id: str, _: None = Depends(require_api_key)):
    job = await store.get_job(settings.telegram_db_path, job_id)
    if not job:
        raise HTTPException(status_code=404, detail="job not found")
    return job


@app.delete("/downloads/{job_id}", response_model=JobOut)
async def cancel_dl(job_id: str, _: None = Depends(require_api_key)):
    job = await store.get_job(settings.telegram_db_path, job_id)
    if not job:
        raise HTTPException(status_code=404, detail="job not found")
    if job["status"] in ("queued", "downloading"):
        cancelled.add(job_id)
        await store.update_job(settings.telegram_db_path, job_id, status="cancelled")
        job = await store.get_job(settings.telegram_db_path, job_id)
    return job
