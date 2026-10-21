import pytest

from app import store


@pytest.mark.asyncio
async def test_enqueue_list_next_update(tmp_path):
    db = str(tmp_path / "q.db")
    await store.init_db(db)
    j = await store.enqueue(db, 123, 456, "ebooks")
    assert j["status"] == "queued"
    nxt = await store.next_queued(db)
    assert nxt and nxt["id"] == j["id"]
    await store.update_job(db, j["id"], status="downloading", progress=0.5)
    got = await store.get_job(db, j["id"])
    assert got["status"] == "downloading" and got["progress"] == 0.5
    jobs = await store.list_jobs(db)
    assert len(jobs) == 1


def test_allowed_extensions():
    from app.telegram_service import is_allowed_file

    assert is_allowed_file("book.epub", 2_000_000_000, 1_000)
    assert is_allowed_file("audio.m4b", 2_000_000_000, 1_000)
    assert not is_allowed_file("movie.mkv", 2_000_000_000, 1_000)
    assert not is_allowed_file("huge.epub", 10, 11)
    assert not is_allowed_file(None, 2_000_000_000, 1_000)


@pytest.mark.asyncio
async def test_enqueue_defaults_to_staging(tmp_path):
    db = str(tmp_path / "q.db")
    await store.init_db(db)
    j = await store.enqueue(db, 1, 2)
    assert j["category"] == "staging"


@pytest.mark.asyncio
async def test_enqueue_reuses_active_job_on_duplicate(tmp_path):
    db = str(tmp_path / "q.db")
    await store.init_db(db)
    j1 = await store.enqueue(db, 10, 20)
    j2 = await store.enqueue(db, 10, 20)
    assert j2["id"] == j1["id"]


@pytest.mark.asyncio
async def test_enqueue_reenqueues_after_done(tmp_path):
    """Un job terminé ne bloque pas un re-téléchargement du même message."""
    db = str(tmp_path / "q.db")
    await store.init_db(db)
    j1 = await store.enqueue(db, 10, 20)
    await store.update_job(db, j1["id"], status="done")
    j2 = await store.enqueue(db, 10, 20)
    assert j2["id"] != j1["id"]
    jobs = await store.list_jobs(db)
    assert len(jobs) == 1 and jobs[0]["id"] == j2["id"]


@pytest.mark.asyncio
async def test_init_db_migrates_legacy_archived_status(tmp_path):
    """Les bases créées avant la découple queue/triage contiennent des jobs
    'archived' : ils ne sont plus un statut valide, convertis en 'done'."""
    legacy = """
    CREATE TABLE jobs (
      id TEXT PRIMARY KEY, dialog_id INTEGER NOT NULL, message_id INTEGER NOT NULL,
      category TEXT NOT NULL, filename TEXT, size_bytes INTEGER,
      status TEXT NOT NULL DEFAULT 'queued', progress REAL NOT NULL DEFAULT 0,
      dest_path TEXT, error TEXT, created_at REAL NOT NULL, updated_at REAL NOT NULL
    );
    """
    db = str(tmp_path / "q.db")
    import aiosqlite

    async with aiosqlite.connect(db) as d:
        await d.executescript(legacy)
        await d.execute(
            "INSERT INTO jobs VALUES ('x', 1, 2, 'ebooks', 'a.epub', 10,"
            " 'archived', 1, '/books/ebooks/a.epub', NULL, 1, 2)"
        )
        await d.commit()

    await store.init_db(db)
    got = await store.get_job(db, "x")
    assert got["status"] == "done"

    # Idempotent : un second init ne casse rien.
    await store.init_db(db)
    assert (await store.get_job(db, "x"))["status"] == "done"


@pytest.mark.asyncio
async def test_reset_stuck_requeues_downloading(tmp_path):
    db = str(tmp_path / "q.db")
    await store.init_db(db)
    j = await store.enqueue(db, 1, 2, "manga")
    await store.update_job(db, j["id"], status="downloading", progress=0.5)
    n = await store.reset_stuck(db)
    assert n == 1
    got = await store.get_job(db, j["id"])
    assert got["status"] == "queued" and got["progress"] == 0


def test_cancel_exception_exists():
    from app.telegram_service import CancelledDownload

    assert issubclass(CancelledDownload, Exception)


class _FakeFile:
    def __init__(self, name, size):
        self.name = name
        self.size = size


class _FakeMsg:
    def __init__(self, file):
        self.file = file


class _FakeClient:
    def __init__(self, msg=None, exc=None):
        self._msg = msg
        self._exc = exc

    async def get_messages(self, dialog_id, ids):
        if self._exc:
            raise self._exc
        return self._msg


def _patch_client(monkeypatch, client):
    from app import telegram_service as tg

    async def fake_get_client():
        return client

    monkeypatch.setattr(tg, "get_client", fake_get_client)


@pytest.mark.asyncio
async def test_peek_returns_filename_and_size(monkeypatch):
    from app import telegram_service as tg

    _patch_client(monkeypatch, _FakeClient(_FakeMsg(_FakeFile("book.cbz", 12345))))
    assert await tg.peek_attachment(1, 2) == {
        "filename": "book.cbz",
        "size_bytes": 12345,
    }


@pytest.mark.asyncio
async def test_peek_no_file_or_unreachable_returns_nones(monkeypatch):
    from app import telegram_service as tg

    _patch_client(monkeypatch, _FakeClient(_FakeMsg(None)))
    assert await tg.peek_attachment(1, 2) == {"filename": None, "size_bytes": None}
    _patch_client(monkeypatch, _FakeClient(None))
    assert await tg.peek_attachment(1, 2) == {"filename": None, "size_bytes": None}
    _patch_client(monkeypatch, _FakeClient(exc=RuntimeError("offline")))
    assert await tg.peek_attachment(1, 2) == {"filename": None, "size_bytes": None}
