"""Tests du module staging : dossier de téléchargement + déplacement vers les catégories."""

import pytest

from app import files


@pytest.fixture
def dirs(monkeypatch, tmp_path):
    """Redirige staging + catégories vers tmp_path."""
    from app.config import settings

    stage = tmp_path / "stage"
    ebooks = tmp_path / "ebooks"
    audiobooks = tmp_path / "audiobooks"
    manga = tmp_path / "manga"
    stage.mkdir()
    monkeypatch.setattr(settings, "download_dir", str(stage))
    monkeypatch.setattr(settings, "ebook_dir", str(ebooks))
    monkeypatch.setattr(settings, "audiobook_dir", str(audiobooks))
    monkeypatch.setattr(settings, "manga_dir", str(manga))
    return {"stage": stage, "ebooks": ebooks, "audiobooks": audiobooks, "manga": manga}


def test_categories(dirs):
    cats = files.categories()
    assert set(cats) == {"ebooks", "audiobooks", "manga"}
    assert cats["ebooks"].name == "ebooks"


def test_list_staged_empty(dirs):
    assert files.list_staged() == []


def test_list_staged_lists_files_with_size_and_mtime(dirs):
    (dirs["stage"] / "a.epub").write_bytes(b"x" * 10)
    (dirs["stage"] / "b.cbz").write_bytes(b"y" * 5)
    (dirs["stage"] / "subdir").mkdir()  # pas un fichier -> ignoré
    (dirs["stage"] / ".hidden").write_text("noise")  # dotfile -> ignoré

    got = files.list_staged()
    names = [f["filename"] for f in got]
    assert names == ["a.epub", "b.cbz"]
    assert got[0]["size_bytes"] == 10
    assert got[0]["modified"] is not None


def test_unique_path_no_collision(dirs):
    p = files.unique_path(dirs["stage"], "book.epub")
    assert p == dirs["stage"] / "book.epub"


def test_unique_path_suffixes_on_collision(dirs):
    (dirs["stage"] / "book.epub").touch()
    p1 = files.unique_path(dirs["stage"], "book.epub")
    p1.touch()
    p2 = files.unique_path(dirs["stage"], "book.epub")
    assert p1.name == "book (1).epub"
    assert p2.name == "book (2).epub"


def test_sanitize_rejects_traversal(dirs):
    with pytest.raises(ValueError):
        files.sanitize("../evil.epub")
    with pytest.raises(ValueError):
        files.sanitize("sub/dir/evil.epub")
    with pytest.raises(ValueError):
        files.sanitize(".")
    with pytest.raises(ValueError):
        files.sanitize("")


def test_resolve_staged_rejects_missing(dirs):
    with pytest.raises(FileNotFoundError):
        files.resolve_staged("nope.epub")


def test_move_to_category(dirs):
    src = dirs["stage"] / "dune.epub"
    src.write_bytes(b"data")

    res = files.move_to_category("dune.epub", "ebooks")

    assert src.exists() is False
    assert (dirs["ebooks"] / "dune.epub").read_bytes() == b"data"
    assert res["dest_path"] == str(dirs["ebooks"] / "dune.epub")
    assert res["category"] == "ebooks"


def test_move_creates_missing_category_dir(dirs):
    assert not dirs["audiobooks"].exists()
    (dirs["stage"] / "lune.m4b").write_bytes(b"")
    files.move_to_category("lune.m4b", "audiobooks")
    assert dirs["audiobooks"].is_dir()


def test_move_unknown_category_rejected(dirs):
    (dirs["stage"] / "x.epub").write_bytes(b"")
    with pytest.raises(ValueError):
        files.move_to_category("x.epub", "comics")


def test_move_collision_returns_file_exists_error_and_keeps_source(dirs):
    (dirs["stage"] / "dup.cbz").write_bytes(b"new")
    dirs["manga"].mkdir()
    (dirs["manga"] / "dup.cbz").write_bytes(b"old")

    with pytest.raises(FileExistsError):
        files.move_to_category("dup.cbz", "manga")

    assert (dirs["stage"] / "dup.cbz").read_bytes() == b"new"
    assert (dirs["manga"] / "dup.cbz").read_bytes() == b"old"


def test_delete_staged(dirs):
    (dirs["stage"] / "junk.epub").write_bytes(b"")
    files.delete_staged("junk.epub")
    assert not (dirs["stage"] / "junk.epub").exists()


def test_delete_staged_missing(dirs):
    with pytest.raises(FileNotFoundError):
        files.delete_staged("nope.epub")


def test_clear_tmp_removes_only_partials(dirs):
    (dirs["stage"] / ".tmp.dune.epub").write_bytes(b"partial")
    (dirs["stage"] / "done.epub").write_bytes(b"done")
    (dirs["stage"] / ".tmp.maybe_dir").mkdir()

    n = files.clear_tmp()

    assert n >= 1
    assert not (dirs["stage"] / ".tmp.dune.epub").exists()
    assert (dirs["stage"] / "done.epub").exists()
    assert files.resolve_staged("done.epub").exists()


def test_resolve_dest_root_and_subdir(dirs):
    cat, target = files.resolve_dest("ebooks")
    assert cat == "ebooks" and target == dirs["ebooks"]
    cat, target = files.resolve_dest("ebooks/Auteurs")
    assert cat == "ebooks" and target == dirs["ebooks"] / "Auteurs"


def test_resolve_dest_rejects_bad_paths(dirs):
    for bad in ["", "nope/sub", "/abs/ebooks", "ebooks/../manga", "../../ebooks", ".."]:
        with pytest.raises(ValueError):
            files.resolve_dest(bad)


def test_resolve_dest_rejects_symlink_escape(monkeypatch, tmp_path, dirs):
    dirs["ebooks"].mkdir()
    escape = tmp_path / "outside"
    escape.mkdir()
    (dirs["ebooks"] / "evil").symlink_to(escape)
    with pytest.raises(ValueError):
        files.resolve_dest("ebooks/evil")


def test_move_into_creates_subdirs(dirs):
    src = dirs["stage"] / "dune.epub"
    src.write_bytes(b"data")
    dest_dir = dirs["ebooks"] / "Auteurs" / "Herbert"

    res = files.move_into(src, "ebooks", dest_dir)

    assert (dest_dir / "dune.epub").read_bytes() == b"data"
    assert res["dest_path"] == str(dest_dir / "dune.epub")
    assert res["category"] == "ebooks"


def test_move_into_collision_preserves_both(dirs):
    src = dirs["stage"] / "dup.cbz"
    src.write_bytes(b"new")
    dest_dir = dirs["manga"] / "Nouvelle"
    dest_dir.mkdir(parents=True)
    (dest_dir / "dup.cbz").write_bytes(b"old")

    with pytest.raises(FileExistsError):
        files.move_into(src, "manga", dest_dir)

    assert src.read_bytes() == b"new"
    assert (dest_dir / "dup.cbz").read_bytes() == b"old"


def test_subdirs_lists_relative_dirs_with_counts(dirs):
    dirs["ebooks"].mkdir()
    a = dirs["ebooks"] / "A"
    b = a / "B"
    b.mkdir(parents=True)
    (b / "tome.cbz").write_bytes(b"x")
    (a / "fichier.txt").touch()
    (dirs["ebooks"] / ".hidden").mkdir()
    (dirs["ebooks"] / "@eaDir").mkdir()

    got = files.subdirs("ebooks")

    assert got == [
        {"path": "A", "files": 1},
        {"path": "A/B", "files": 1},
    ]
    assert files.subdirs("manga") == []


def test_make_dir_creates_nested_and_is_idempotent(dirs):
    made = files.make_dir("manga/One Piece/Tomes")
    assert made.is_dir()
    again = files.make_dir("manga/One Piece/Tomes")
    assert again == made
    with pytest.raises(ValueError):
        files.make_dir("../evil")


def test_list_files_direct_children_only_with_sizes(dirs):
    dirs["manga"].mkdir()
    folder = dirs["manga"] / "One Piece"
    folder.mkdir()
    (folder / "tome1.cbz").write_bytes(b"a" * 64)
    (folder / "tome2.cbz").write_bytes(b"a" * 32)
    (folder / ".hidden.epub").write_bytes(b"x")
    (folder / "@eaDir").write_bytes(b"x")
    sub = folder / "sub"
    sub.mkdir()
    (sub / "profond.epub").write_bytes(b"x")

    got = files.list_files("manga/One Piece")

    assert [f["name"] for f in got] == ["tome1.cbz", "tome2.cbz"]
    assert got[0]["size_bytes"] == 64
    assert files.list_files("manga/Inexistant") == []


def test_list_files_rejects_bad_path(dirs):
    with pytest.raises(ValueError):
        files.list_files("../evil")
