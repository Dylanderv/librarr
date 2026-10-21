"""File de téléchargement persistée en sqlite. Un seul worker actif."""

import aiosqlite
import time
import uuid
from pathlib import Path

SCHEMA = """
CREATE TABLE IF NOT EXISTS jobs (
  id TEXT PRIMARY KEY,
  dialog_id INTEGER NOT NULL,
  message_id INTEGER NOT NULL,
  category TEXT NOT NULL,
  filename TEXT,
  size_bytes INTEGER,
  status TEXT NOT NULL DEFAULT 'queued',
  progress REAL NOT NULL DEFAULT 0,
  dest_path TEXT,
  error TEXT,
  created_at REAL NOT NULL,
  updated_at REAL NOT NULL,
  UNIQUE(dialog_id, message_id, category)
);
"""


async def init_db(db_path: str) -> None:
    Path(db_path).parent.mkdir(parents=True, exist_ok=True)
    async with aiosqlite.connect(db_path) as db:
        await db.executescript(SCHEMA)
        # Bases antérieures au découplage queue/triage : 'archived' n'est
        # plus un statut, les jobs correspondants redeviennent de simples
        # téléchargements terminés.
        await db.execute("UPDATE jobs SET status='done' WHERE status='archived'")
        await db.commit()


async def enqueue(
    db_path: str, dialog_id: int, message_id: int, category: str = "staging"
) -> dict:
    """Empile un job. Un même message ne peut être actif (queued/downloading) qu'une fois.

    category vaut toujours 'staging' : la file ne suit pas le tri des fichiers.
    """
    job_id = str(uuid.uuid4())
    now = time.time()
    async with aiosqlite.connect(db_path) as db:
        await db.executescript(SCHEMA)
        try:
            await db.execute(
                "INSERT INTO jobs (id, dialog_id, message_id, category, status, created_at, updated_at)"
                " VALUES (?, ?, ?, ?, 'queued', ?, ?)",
                (job_id, dialog_id, message_id, category, now, now),
            )
            await db.commit()
        except Exception as e:
            # Doublon : réutilise le job s'il est encore actif, sinon replace
            # le job terminé (re-téléchargement du même message autorisé).
            if "UNIQUE" in str(e):
                async with db.execute(
                    "SELECT id, status FROM jobs WHERE dialog_id=? AND message_id=?"
                    " ORDER BY status='queued' DESC, status='downloading' DESC LIMIT 1",
                    (dialog_id, message_id),
                ) as cur:
                    row = await cur.fetchone()
                if row:
                    existing_id, existing_status = row
                    if existing_status in ("queued", "downloading"):
                        return await _fetch_job(db, existing_id)
                    await db.execute(
                        "DELETE FROM jobs WHERE dialog_id=? AND message_id=?",
                        (dialog_id, message_id),
                    )
                    await db.commit()
                    return await enqueue(db_path, dialog_id, message_id, category)
            raise
    return {
        "id": job_id,
        "dialog_id": dialog_id,
        "message_id": message_id,
        "category": category,
        "filename": None,
        "size_bytes": None,
        "status": "queued",
        "progress": 0.0,
        "dest_path": None,
        "error": None,
    }


def _row_to_job(row) -> dict:
    return {
        "id": row[0],
        "dialog_id": row[1],
        "message_id": row[2],
        "category": row[3],
        "filename": row[4],
        "size_bytes": row[5],
        "status": row[6],
        "progress": row[7] or 0.0,
        "dest_path": row[8],
        "error": row[9],
    }


_COLS = (
    "id, dialog_id, message_id, category, filename, size_bytes,"
    " status, progress, dest_path, error"
)


async def _fetch_job(db, job_id: str) -> dict | None:
    async with db.execute(f"SELECT {_COLS} FROM jobs WHERE id=?", (job_id,)) as cur:
        row = await cur.fetchone()
        return _row_to_job(row) if row else None


async def list_jobs(db_path: str) -> list[dict]:
    async with aiosqlite.connect(db_path) as db:
        await db.executescript(SCHEMA)
        async with db.execute(
            f"SELECT {_COLS} FROM jobs ORDER BY created_at DESC LIMIT 200"
        ) as cur:
            return [_row_to_job(r) async for r in cur]


async def get_job(db_path: str, job_id: str) -> dict | None:
    async with aiosqlite.connect(db_path) as db:
        await db.executescript(SCHEMA)
        return await _fetch_job(db, job_id)


async def next_queued(db_path: str) -> dict | None:
    async with aiosqlite.connect(db_path) as db:
        await db.executescript(SCHEMA)
        async with db.execute(
            f"SELECT {_COLS} FROM jobs"
            " WHERE status='queued' ORDER BY created_at ASC LIMIT 1"
        ) as cur:
            row = await cur.fetchone()
            return _row_to_job(row) if row else None


async def update_job(db_path: str, job_id: str, **fields) -> None:
    if not fields:
        return
    fields["updated_at"] = time.time()
    cols = ", ".join(f"{k}=?" for k in fields)
    async with aiosqlite.connect(db_path) as db:
        await db.execute(
            f"UPDATE jobs SET {cols} WHERE id=?", (*fields.values(), job_id)
        )
        await db.commit()


async def reset_stuck(db_path: str) -> int:
    """Remet en file les jobs restés 'downloading' (crash/restart)."""
    async with aiosqlite.connect(db_path) as db:
        await db.executescript(SCHEMA)
        cur = await db.execute(
            "UPDATE jobs SET status='queued', progress=0, updated_at=? WHERE status='downloading'",
            (time.time(),),
        )
        await db.commit()
        return cur.rowcount
