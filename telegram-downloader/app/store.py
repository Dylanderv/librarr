"""File de téléchargement persistée en sqlite. Un seul worker actif."""

import aiosqlite
import time
import uuid

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
    async with aiosqlite.connect(db_path) as db:
        await db.executescript(SCHEMA)
        await db.commit()


async def enqueue(db_path: str, dialog_id: int, message_id: int, category: str) -> dict:
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
            # Doublon (même message + catégorie) -> retourne le job existant.
            if "UNIQUE" in str(e):
                async with db.execute(
                    "SELECT id, dialog_id, message_id, category, filename, size_bytes,"
                    " status, progress, dest_path, error"
                    " FROM jobs WHERE dialog_id=? AND message_id=? AND category=?",
                    (dialog_id, message_id, category),
                ) as cur:
                    row = await cur.fetchone()
                if row:
                    return _row_to_job(row)
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


async def list_jobs(db_path: str) -> list[dict]:
    async with aiosqlite.connect(db_path) as db:
        await db.executescript(SCHEMA)
        async with db.execute(
            "SELECT id, dialog_id, message_id, category, filename, size_bytes,"
            " status, progress, dest_path, error FROM jobs ORDER BY created_at DESC LIMIT 200"
        ) as cur:
            return [_row_to_job(r) async for r in cur]


async def get_job(db_path: str, job_id: str) -> dict | None:
    async with aiosqlite.connect(db_path) as db:
        await db.executescript(SCHEMA)
        async with db.execute(
            "SELECT id, dialog_id, message_id, category, filename, size_bytes,"
            " status, progress, dest_path, error FROM jobs WHERE id=?",
            (job_id,),
        ) as cur:
            row = await cur.fetchone()
            return _row_to_job(row) if row else None


async def next_queued(db_path: str) -> dict | None:
    async with aiosqlite.connect(db_path) as db:
        await db.executescript(SCHEMA)
        async with db.execute(
            "SELECT id, dialog_id, message_id, category, filename, size_bytes,"
            " status, progress, dest_path, error FROM jobs"
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
