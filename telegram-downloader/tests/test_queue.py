import asyncio
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
