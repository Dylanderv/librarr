"""API FastAPI du sidecar Telegram (read-only + queue 1 à la fois)."""

import asyncio
from contextlib import asynccontextmanager

from fastapi import Depends, FastAPI, HTTPException, Query
from fastapi.responses import HTMLResponse

from . import files as fs
from . import store
from . import telegram_service as tg
from .config import settings
from .schemas import (
    CodeConfirmIn,
    CodeRequestIn,
    DestIn,
    EnqueueDownloadIn,
    JobOut,
    LibraryMoveIn,
    MoveIn,
    StagedFileOut,
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
            # Nom/taille connus dès le début du download (le worker les
            # résolvait seulement à la fin, d'où une file aveugle).
            meta = await tg.peek_attachment(job["dialog_id"], job["message_id"])
            fill = {k: v for k, v in meta.items() if v is not None}
            if fill:
                await store.update_job(settings.telegram_db_path, job["id"], **fill)

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
    # Purge des téléchargements partiels (.tmp.*) restés sur disque.
    fs.clear_tmp()
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
        settings.telegram_db_path, body.dialog_id, body.message_id
    )
    meta = await tg.peek_attachment(body.dialog_id, body.message_id)
    fill = {k: v for k, v in meta.items() if v is not None}
    if fill:
        await store.update_job(settings.telegram_db_path, job["id"], **fill)
        job = await store.get_job(settings.telegram_db_path, job["id"])
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


# --- Staging : tri des fichiers téléchargés -------------------------------


@app.get("/categories")
async def list_categories(_: None = Depends(require_api_key)):
    return [{"id": name, "path": str(p)} for name, p in fs.categories().items()]


@app.get("/staging", response_model=list[StagedFileOut])
async def list_staged(_: None = Depends(require_api_key)):
    return fs.list_staged()


@app.get("/staging/tree")
async def staging_tree(_: None = Depends(require_api_key)):
    """Arborescence existante par bibliothèque (pour l'explorateur de destination)."""
    return {name: fs.subdirs(name) for name in fs.categories()}


@app.post("/staging/mkdir")
async def staging_mkdir(body: DestIn, _: None = Depends(require_api_key)):
    """Crée 'categorie/sous/dossier' (idempotent)."""
    try:
        made = fs.make_dir(body.dest)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
    return {"created": True, "dest": str(made)}


@app.get("/library/files")
async def library_files(path: str = Query(), _: None = Depends(require_api_key)):
    """Fichiers directs d'un dossier de bibliothèque (pour l'explorateur)."""
    try:
        return {"path": path, "files": fs.list_files(path)}
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.post("/library/move")
async def move_library(body: LibraryMoveIn, _: None = Depends(require_api_key)):
    """Réorganise les fichiers d'une bibliothèque, indépendamment de la file.

    dest = 'staging' (mauvais rangement -> re-tri) ou 'categorie/sous/dossier'.
    """
    if not body.files:
        raise HTTPException(status_code=422, detail="aucun fichier fourni")
    dest_dir = None
    category = None
    if body.dest != "staging":
        try:
            category, dest_dir = fs.resolve_dest(body.dest)
        except ValueError as e:
            raise HTTPException(status_code=400, detail=str(e))

    moved: list[dict] = []
    skipped: list[dict] = []
    for rel in body.files:
        try:
            src = fs.resolve_library(rel)
            res = (
                fs.move_to_staging(src)
                if dest_dir is None
                else fs.move_into(src, category, dest_dir)
            )
        except FileNotFoundError:
            skipped.append({"filename": rel, "error": "absent de la bibliothèque"})
        except IsADirectoryError:
            skipped.append({"filename": rel, "error": "n'est pas un fichier"})
        except FileExistsError as e:
            skipped.append({"filename": rel, "error": f"existe déjà: {e}"})
        except ValueError as e:
            skipped.append({"filename": rel, "error": str(e)})
        else:
            moved.append(res)
    return {"moved": moved, "skipped": skipped}


@app.post("/staging/move")
async def move_staged(body: MoveIn, _: None = Depends(require_api_key)):
    """Déplace des fichiers du staging vers 'categorie/sous/dossier'.

    Résultat par fichier : moved + skipped (collisions, absents).
    """
    if not body.files:
        raise HTTPException(status_code=422, detail="aucun fichier fourni")
    try:
        category, dest_dir = fs.resolve_dest(body.dest)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

    moved: list[dict] = []
    skipped: list[dict] = []
    for filename in body.files:
        try:
            src = fs.resolve_staged(filename)
            res = fs.move_into(src, category, dest_dir)
        except FileNotFoundError:
            skipped.append({"filename": filename, "error": "absent du staging"})
        except FileExistsError as e:
            skipped.append({"filename": filename, "error": f"existe déjà: {e}"})
        except ValueError as e:
            skipped.append({"filename": filename, "error": str(e)})
        else:
            moved.append(res)
    return {"moved": moved, "skipped": skipped}


@app.delete("/staging/{filename}")
async def delete_staged(filename: str, _: None = Depends(require_api_key)):
    """Supprime du staging un fichier indésirable (la file n'est pas touchée)."""
    try:
        fs.resolve_staged(filename)
    except FileNotFoundError:
        raise HTTPException(status_code=404, detail="fichier absent du staging")
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
    fs.delete_staged(filename)
    return {"deleted": True}
