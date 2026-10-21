"""Tests des routes staging/catégories de l'API."""

import asyncio

import pytest
from fastapi.testclient import TestClient

from app import store
from app.config import settings
from app.main import app


@pytest.fixture
def env(monkeypatch, tmp_path):
    """DB + dossiers isolés dans tmp_path."""
    stage = tmp_path / "stage"
    stage.mkdir()
    monkeypatch.setattr(settings, "telegram_db_path", str(tmp_path / "q.db"))
    monkeypatch.setattr(settings, "download_dir", str(stage))
    monkeypatch.setattr(settings, "ebook_dir", str(tmp_path / "ebooks"))
    monkeypatch.setattr(settings, "audiobook_dir", str(tmp_path / "audiobooks"))
    monkeypatch.setattr(settings, "manga_dir", str(tmp_path / "manga"))
    monkeypatch.setattr(settings, "telegram_sidecar_api_key", "")  # pas d'auth en test
    with TestClient(app) as client:
        yield client, tmp_path, stage


def _api(client):
    return client


def _stage_job_done(client, stage, filename):
    """Empile un job et le passe à 'done' avec un dest_path dans le staging."""
    r = _api(client).post("/downloads", json={"dialog_id": 1, "message_id": 2})
    job_id = r.json()["id"]
    asyncio.run(
        store.update_job(
            settings.telegram_db_path,
            job_id,
            status="done",
            filename=filename,
            dest_path=str(stage / filename),
        )
    )
    return job_id


def test_health(env):
    client, _, _ = env
    assert _api(client).get("/health").status_code == 200


def test_categories_lists_the_three_libraries(env):
    client, tmp_path, _ = env
    r = _api(client).get("/categories")
    assert r.status_code == 200
    cats = r.json()
    assert [c["id"] for c in cats] == ["ebooks", "audiobooks", "manga"]
    assert all(c["path"].startswith(str(tmp_path)) for c in cats)


def test_enqueue_without_category_lands_in_staging(env):
    client, _, _ = env
    r = _api(client).post("/downloads", json={"dialog_id": 1, "message_id": 2})
    assert r.status_code == 200
    assert r.json()["category"] == "staging"


def test_staging_hides_partial_downloads(env):
    client, _, stage = env
    (stage / ".tmp.dune.epub").write_bytes(b"partial")
    (stage / "fini.epub").write_bytes(b"done")
    got = _api(client).get("/staging").json()
    assert [f["filename"] for f in got] == ["fini.epub"]


def test_move_bulk_creates_subdirs(env):
    client, tmp_path, stage = env
    (stage / "a.epub").write_bytes(b"a")
    (stage / "b.epub").write_bytes(b"b")

    r = _api(client).post(
        "/staging/move", json={"files": ["a.epub", "b.epub"], "dest": "ebooks/Auteurs"}
    )
    assert r.status_code == 200
    assert r.json()["skipped"] == []
    dest_dir = tmp_path / "ebooks" / "Auteurs"
    assert (dest_dir / "a.epub").read_bytes() == b"a"
    assert (dest_dir / "b.epub").read_bytes() == b"b"
    assert _api(client).get("/staging").json() == []


def test_move_leaves_jobs_alone(env):
    """La file est découplée du tri : un déplacement ne touche pas le job."""
    client, tmp_path, stage = env
    (stage / "dune.epub").write_bytes(b"data")
    _stage_job_done(client, stage, "dune.epub")
    before = _api(client).get("/downloads").json()[0]

    r = _api(client).post(
        "/staging/move", json={"files": ["dune.epub"], "dest": "ebooks"}
    )
    assert r.status_code == 200

    after = _api(client).get("/downloads").json()[0]
    assert after == before


def test_move_reports_skipped_without_blocking_others(env):
    client, tmp_path, stage = env
    (stage / "a.epub").write_bytes(b"a")

    r = _api(client).post(
        "/staging/move",
        json={"files": ["a.epub", "missing.epub", "../evil"], "dest": "ebooks"},
    )
    assert r.status_code == 200
    payload = r.json()
    assert [m["filename"] for m in payload["moved"]] == ["a.epub"]
    assert [s["filename"] for s in payload["skipped"]] == [
        "missing.epub",
        "../evil",
    ]
    assert (tmp_path / "ebooks" / "a.epub").read_bytes() == b"a"


def test_move_validation_errors(env):
    client, _, stage = env
    (_api(client)).__class__  # noqa: F841

    r = _api(client).post("/staging/move", json={"files": [], "dest": "ebooks"})
    assert r.status_code == 422

    r = _api(client).post("/staging/move", json={"files": ["x"], "dest": "pdftools"})
    assert r.status_code == 400


def test_move_collision_skips_and_keeps_source(env):
    client, tmp_path, stage = env
    (stage / "dup.cbz").write_bytes(b"new")
    (tmp_path / "ebooks").mkdir()
    (tmp_path / "ebooks" / "dup.cbz").write_bytes(b"old")
    r = _api(client).post(
        "/staging/move", json={"files": ["dup.cbz"], "dest": "ebooks"}
    )
    assert r.status_code == 200
    assert r.json()["moved"] == []
    assert r.json()["skipped"][0]["filename"] == "dup.cbz"
    assert (stage / "dup.cbz").read_bytes() == b"new"


def test_tree_lists_existing_subdirs_with_counts(env):
    client, tmp_path, _ = env
    a = tmp_path / "ebooks" / "A"
    (a / "B").mkdir(parents=True)
    (a / "B" / "tome.cbz").write_bytes(b"x")
    (tmp_path / "ebooks" / ".hidden").mkdir()

    r = _api(client).get("/staging/tree")
    assert r.status_code == 200
    tree = r.json()
    assert tree["ebooks"] == [
        {"path": "A", "files": 0},
        {"path": "A/B", "files": 1},
    ]
    assert tree["manga"] == []


def test_mkdir_endpoint_creates_and_validates(env):
    client, tmp_path, _ = env
    r = _api(client).post("/staging/mkdir", json={"dest": "manga/One Piece"})
    assert r.status_code == 200
    assert (tmp_path / "manga" / "One Piece").is_dir()

    # Idempotent.
    assert (_api(client)).post(
        "/staging/mkdir", json={"dest": "manga/One Piece"}
    ).status_code == 200

    r = _api(client).post("/staging/mkdir", json={"dest": "nope"})
    assert r.status_code == 400

    r = _api(client).post("/staging/mkdir", json={"dest": "ebooks/../evil"})
    assert r.status_code == 400


def test_library_move_back_to_staging(env):
    """Un fichier mal rangé repart au staging pour re-tri."""
    client, tmp_path, stage = env
    folder = tmp_path / "manga" / "One Piece"
    folder.mkdir(parents=True)
    (folder / "tome3.cbz").write_bytes(b"c")

    r = _api(client).post(
        "/library/move",
        json={"files": ["manga/One Piece/tome3.cbz"], "dest": "staging"},
    )
    assert r.status_code == 200
    assert (stage / "tome3.cbz").read_bytes() == b"c"
    assert not (folder / "tome3.cbz").exists()
    assert [f["filename"] for f in _api(client).get("/staging").json()] == ["tome3.cbz"]


def test_library_move_within_library(env):
    """Changement de dossier de bibliothèque en un move."""
    client, tmp_path, _ = env
    src = tmp_path / "manga" / "One Piece"
    src.mkdir(parents=True)
    (src / "tome3.cbz").write_bytes(b"c")

    r = _api(client).post(
        "/library/move",
        json={"files": ["manga/One Piece/tome3.cbz"], "dest": "manga/BD"},
    )
    assert r.status_code == 200
    assert (tmp_path / "manga" / "BD" / "tome3.cbz").read_bytes() == b"c"
    assert not (src / "tome3.cbz").exists()


def test_library_move_validations(env):
    client, tmp_path, stage = env
    folder = tmp_path / "ebooks" / "SF"
    folder.mkdir(parents=True)
    (folder / "sub").mkdir()
    (folder / "sub" / "in.ebook").write_bytes(b"i")
    (stage / "tome.cbz").write_bytes(b"old")  # collision au retour au staging
    (folder / "dune.epub").write_bytes(b"d")

    r = _api(client).post("/library/move", json={"dest": "staging", "files": []})
    assert r.status_code == 422

    r = _api(client).post("/library/move", json={"dest": "../evil", "files": ["d"]})
    assert r.status_code == 400

    r = _api(client).post(
        "/library/move",
        json={
            "dest": "staging",
            "files": [
                "ebooks/SF/dune.epub",  # ok
                "ebooks/SF/sub",  # un dossier -> refusé
                "ebooks/SF/absent.epub",  # absent -> refusé
                "manga/evil",  # pas dans un dossier de bibliothèque ->
                "nope/x",  # bibliothèque inconnue ->
                "../../etc/passwd",  # absolu/traversal ->
            ],
        },
    )
    assert r.status_code == 200
    payload = r.json()
    assert [m["filename"] for m in payload["moved"]] == ["dune.epub"]
    skipped_by_name = {s["filename"]: s["error"] for s in payload["skipped"]}
    assert set(skipped_by_name) == {
        "ebooks/SF/sub",
        "ebooks/SF/absent.epub",
        "manga/evil",
        "nope/x",
        "../../etc/passwd",
    }

    # Collision au retour : suffixe, jamais d'écrasement.
    r = _api(client).post(
        "/library/move", json={"files": ["ebooks/SF/dune.epub"], "dest": "staging"}
    )
    assert r.json()["skipped"][0]["error"].startswith("absent")


def test_library_move_collision_gets_unique_name(env):
    client, tmp_path, stage = env
    (stage / "tome.cbz").write_bytes(b"old")
    folder = tmp_path / "ebooks" / "SF"
    folder.mkdir(parents=True)
    (folder / "tome.cbz").write_bytes(b"new")

    r = _api(client).post(
        "/library/move", json={"files": ["ebooks/SF/tome.cbz"], "dest": "staging"}
    )
    payload = r.json()
    assert [m["filename"] for m in payload["moved"]] == ["tome (1).cbz"]
    assert (stage / "tome.cbz").read_bytes() == b"old"
    assert (stage / "tome (1).cbz").read_bytes() == b"new"


def test_library_files_endpoint(env):
    client, tmp_path, _ = env
    folder = tmp_path / "manga" / "One Piece"
    folder.mkdir(parents=True)
    (folder / "tome1.cbz").write_bytes(b"a" * 64)

    r = _api(client).get("/library/files", params={"path": "manga/One Piece"})
    assert r.status_code == 200
    assert r.json()["files"] == [{"name": "tome1.cbz", "size_bytes": 64}]

    # Dossier inexistant -> liste vide.
    assert (
        _api(client)
        .get("/library/files", params={"path": "manga/Berserk"})
        .json()
        .get("files")
        == []
    )

    # Chemins invalides refusés.
    assert (
        _api(client).get("/library/files", params={"path": "../evil"}).status_code
        == 400
    )


def test_staging_list_and_delete_flow(env):
    client, tmp_path, stage = env
    (stage / "dune.epub").write_bytes(b"data")

    r = _api(client).get("/staging")
    assert r.status_code == 200
    assert r.json() == [
        {"filename": "dune.epub", "size_bytes": 4, "modified": r.json()[0]["modified"]}
    ]

    # Supprime le fichier → back to empty.
    assert _api(client).delete("/staging/dune.epub").status_code == 200
    assert not (stage / "dune.epub").exists()
    assert _api(client).delete("/staging/nope.epub").status_code == 404
