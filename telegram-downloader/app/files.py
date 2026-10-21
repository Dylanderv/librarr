"""Dossier de staging : les downloads aboutissent ici, catégorisation ensuite.

Séparation stricte :
- download  -> settings.download_dir (staging)
- move      -> staging vers une bibliothèque (ebooks/audiobooks/manga/bd)

Le module ne touche ni sqlite ni Telethon : responsabilité = le disque.
"""

import shutil
from pathlib import Path

from .config import settings

CATEGORIES = ("ebooks", "audiobooks", "manga", "bd")


class StagingError(Exception):
    """Base des erreurs du staging."""


def categories() -> dict[str, Path]:
    """Les dossiers de bibliothèque cibles, du plus stable au plus récent."""
    cfg = {
        "ebooks": settings.ebook_dir,
        "audiobooks": settings.audiobook_dir,
        "manga": settings.manga_dir,
        # BD & comics (Kavita/Komga gèrent les deux formats indifféremment).
        "bd": settings.bd_dir,
    }
    return {name: Path(value) for name, value in cfg.items()}


def staging_dir() -> Path:
    p = Path(settings.download_dir)
    p.mkdir(parents=True, exist_ok=True)
    return p


def sanitize(filename: str) -> str:
    """Refuse chemins relatif/absolu et autres manières de sortir du staging."""
    name = Path(filename).name
    if name in ("", ".", "..") or name != filename:
        raise ValueError(f"nom de fichier invalide: {filename!r}")
    return name


def resolve_dest(dest: str) -> tuple[str, Path]:
    """'ebooks/Auteurs/Herbert' -> ('ebooks', ebooks_root/'Auteurs'/'Herbert').

    Refuse absolu, '..' et tout ce qui n'est pas une bibliothèque connue. Les
    sous-dossiers n'ont pas besoin d'exister : le move les crée.
    """
    p = Path(dest)
    if p.is_absolute() or not p.parts or ".." in p.parts:
        raise ValueError(f"destination invalide: {dest!r}")
    category = p.parts[0]
    dirs = categories()
    if category not in dirs:
        raise ValueError(f"bibliothèque inconnue: {category!r}")
    target: Path = dirs[category]
    for part in p.parts[1:]:
        target = target / part
    # Paradoïe symlink : la cible (même inexistante) doit rester dans la
    # bibliothèque une fois les liens résolus.
    root = dirs[category].resolve()
    if target.resolve() != root and root not in target.resolve().parents:
        raise ValueError(f"destination hors bibliothèque: {dest!r}")
    return category, target


def move_into(src: Path, category: str, dest_dir: Path) -> dict:
    """Déplace un fichier du staging vers dest_dir (créé au besoin).

    Collision : refus explicite (FileExistsError) — ne jamais écraser une
    bibliothèque existante. Le fichier reste dans le staging, à la décision.
    """
    dest_dir.mkdir(parents=True, exist_ok=True)
    dest = dest_dir / src.name
    if dest.exists():
        raise FileExistsError(str(dest))
    shutil.move(str(src), str(dest))
    return {
        "filename": dest.name,
        "dest_path": str(dest),
        "category": category,
    }


def move_to_category(filename: str, category: str) -> dict:
    src = resolve_staged(filename)
    dirs = categories()
    if category not in dirs:
        raise ValueError(f"catégorie inconnue: {category!r}")
    return move_into(src, category, dirs[category])


def subdirs(category: str) -> list[dict]:
    """Arborescence existante d'une bibliothèque (chemins relatifs + nb de fichiers).

    Bruit ignoré : dotfiles et @eaDir (Synology).
    """
    root = categories()[category]
    if not root.exists():
        return []
    out = []
    for path in sorted(root.rglob("*")):
        if not path.is_dir():
            continue
        rel = path.relative_to(root).as_posix()
        if any(part.startswith((".", "@")) for part in rel.split("/")):
            continue
        out.append({"path": rel, "files": _count_files(path)})
    return out


def make_dir(dest: str) -> Path:
    """Crée (idempotent) 'categorie/sous/dossier' — même validation que le move."""
    _category, target = resolve_dest(dest)
    target.mkdir(parents=True, exist_ok=True)
    return target


def list_files(dest: str) -> list[dict]:
    """Fichiers directs d'un dossier de bibliothèque (dotfiles/@eaDir exclus)."""
    _category, target = resolve_dest(dest)
    if not target.exists():
        return []
    out = []
    for p in sorted(target.iterdir()):
        if not p.is_file() or p.name.startswith((".", "@")):
            continue
        out.append({"name": p.name, "size_bytes": p.stat().st_size})
    return out


def _count_files(directory: Path) -> int:
    """Fichiers directement dans le dossier (dotfiles et @eaDir exclus)."""
    try:
        return sum(
            1
            for p in directory.iterdir()
            if p.is_file() and not p.name.startswith((".", "@"))
        )
    except OSError:
        return 0


def clear_tmp() -> int:
    """Purge les partielles '.tmp.*' d'un crash pendant le download."""
    n = 0
    for p in staging_dir().glob(".tmp.*"):
        if p.is_file():
            p.unlink()
            n += 1
    return n


def unique_path(directory: Path, filename: str) -> Path:
    """Chemin de destination inoccupé, suffixé si pris : book (1).epub."""
    p = directory / filename
    if not p.exists():
        return p
    stem, suffix = p.stem, p.suffix
    for i in range(1, 1000):
        candidate = directory / f"{stem} ({i}){suffix}"
        if not candidate.exists():
            return candidate
    raise StagingError(f"trop de collisions pour {filename!r}")


def _staged_path(filename: str) -> Path:
    """Chemin absolu du fichier du staging, après contrôle."""
    return staging_dir() / sanitize(filename)


def list_staged() -> list[dict]:
    """Fichiers en attente de catégorisation (dotfiles et sous-dossiers ignorés)."""
    out = []
    for entry in sorted(staging_dir().iterdir()):
        if entry.name.startswith(".") or not entry.is_file():
            continue
        stat = entry.stat()
        out.append(
            {
                "filename": entry.name,
                "size_bytes": stat.st_size,
                "modified": stat.st_mtime,
            }
        )
    return out


def resolve_staged(filename: str) -> Path:
    src = _staged_path(filename)
    if not src.exists():
        raise FileNotFoundError(filename)
    return src


def resolve_library(rel: str) -> Path:
    """Chemin absolu d'un fichier d'une bibliothèque, après validation.

    Refuse tout ce qui n'est pas 'categorie/sous/dossier/fichier' d'une
    bibliothèque connue, dossiers inclus (un move ne vise qu'un fichier).
    """
    category, target = resolve_dest(rel)
    if not target.exists():
        raise FileNotFoundError(rel)
    if not target.is_file():
        raise IsADirectoryError(rel)
    return target


def move_to_staging(src: Path) -> dict:
    """Renvoie un fichier de bibliothèque au staging ; reste dans le triage.

    Collision au staging : suffixe automatique ('tome (1).cbz') — jamais
    d'écrasement ni de perte.
    """
    dest = unique_path(staging_dir(), src.name)
    shutil.move(str(src), str(dest))
    return {
        "filename": dest.name,
        "dest_path": str(dest),
        "category": "staging",
    }


def delete_staged(filename: str) -> None:
    """Supprime du staging un fichier téléchargé indésirable."""
    resolve_staged(filename).unlink()
