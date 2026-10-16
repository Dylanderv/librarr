"""Wrapper Telethon strictement read-only.

Méthodes autorisées uniquement :
- get_dialogs (canaux/groupes déjà rejoints)
- recherche dans l'historique (iter_messages avec filtre document)
- download_media d'un message précis

INTERDIT ici : send_message, send_file, join/leave, réactions, marquage lu,
ou tout appel d'écriture. Voir tests/test_readonly.py qui verrouille ça.
"""

import asyncio
from pathlib import Path

ALLOWED_EXTENSIONS = {
    ".epub",
    ".pdf",
    ".mobi",
    ".azw3",
    ".fb2",
    ".mp3",
    ".m4b",
    ".m4a",
    ".flac",
    ".cbz",
    ".cbr",
    ".zip",
    ".rar",
}

CATEGORY_DIRS: dict[str, str] = {}


class CancelledDownload(Exception):
    """Levée par le progress callback quand le job est annulé."""


def set_category_dirs(mapping: dict[str, str]) -> None:
    CATEGORY_DIRS.update(mapping)


def dest_dir_for(category: str) -> Path:
    from .config import settings

    base = {
        "ebooks": settings.ebook_dir,
        "audiobooks": settings.audiobook_dir,
        "manga": settings.manga_dir,
    }.get(category, settings.ebook_dir)
    if category in CATEGORY_DIRS:
        base = CATEGORY_DIRS[category]
    p = Path(base)
    p.mkdir(parents=True, exist_ok=True)
    return p


def is_allowed_file(filename: str | None, max_bytes: int, size: int | None) -> bool:
    if not filename:
        return False
    if size is not None and size > max_bytes:
        return False
    return Path(filename).suffix.lower() in ALLOWED_EXTENSIONS


_client = None
_client_lock = asyncio.Lock()


async def get_client():
    """Singleton Telethon. Le .session persiste sur volume."""
    global _client
    async with _client_lock:
        if _client is not None:
            return _client
        from telethon import TelegramClient
        from .config import settings

        if not settings.telegram_api_id or not settings.telegram_api_hash:
            raise RuntimeError("TELEGRAM_API_ID / TELEGRAM_API_HASH manquants")
        _client = TelegramClient(
            settings.telegram_session_path,
            settings.telegram_api_id,
            settings.telegram_api_hash,
        )
        await _client.connect()
        return _client


async def auth_status() -> dict:
    client = await get_client()
    if not await client.is_user_authorized():
        return {"authorized": False}
    me = await client.get_me()
    return {"authorized": True, "user_id": me.id if me else None}


async def request_code(phone: str):
    client = await get_client()
    # Seule écriture autorisée : l'envoi du code de login (init, une fois).
    return await client.send_code_request(phone)


async def confirm_code(phone: str, code: str, password: str | None = None):
    client = await get_client()
    try:
        await client.sign_in(phone=phone, code=code)
    except Exception as e:
        # SessionPasswordNeededError -> 2FA
        if (
            "password" in type(e).__name__.lower()
            or "SessionPasswordNeeded" in type(e).__name__
        ):
            if not password:
                raise RuntimeError("2FA requis : renseigne password")
            await client.sign_in(password=password)
        else:
            raise
    return await auth_status()


async def list_dialogs():
    """Ne retourne que les channels/supergroupes déjà rejoints."""
    from telethon.tl.types import Channel, Chat

    client = await get_client()
    out = []
    async for d in client.iter_dialogs():
        e = d.entity
        if isinstance(e, Channel):
            kind = "channel"
        elif isinstance(e, Chat):
            kind = "group"
        else:
            continue
        out.append(
            {
                "id": d.id,
                "title": d.title,
                "username": getattr(e, "username", None),
                "kind": kind,
            }
        )
    return out


async def search_attachments(dialog_id: int, query: str = "", limit: int = 50):
    """Recherche plein-texte dans l'historique + filtre pièces jointes."""
    from telethon.tl.types import InputMessagesFilterDocument

    client = await get_client()
    entity = await client.get_entity(dialog_id)
    out = []
    # 1) messages avec documents correspondant à la requête
    async for msg in client.iter_messages(
        entity, search=query or None, filter=InputMessagesFilterDocument, limit=limit
    ):
        doc = msg.document or msg.file
        name = getattr(msg.file, "name", None) if msg.file else None
        size = getattr(msg.file, "size", None) if msg.file else None
        out.append(
            {
                "dialog_id": dialog_id,
                "message_id": msg.id,
                "date": msg.date.isoformat() if msg.date else None,
                "text_snippet": (msg.text or "")[:300] or None,
                "filename": name,
                "mime_type": getattr(doc, "mime_type", None) if doc else None,
                "size_bytes": size,
            }
        )
    return out


async def peek_attachment(dialog_id: int, message_id: int) -> dict:
    """Nom/taille d'une pièce jointe sans la télécharger. Lecture seule.

    Best-effort : ne lève jamais, renvoie des champs None quand Telegram est
    injoignable ou que le message n'a pas de fichier. Sert à afficher un nom
    dans la file d'attente avant (et pendant) le téléchargement.
    """
    try:
        client = await get_client()
        msg = await client.get_messages(dialog_id, ids=message_id)
        if not msg or not msg.file:
            return {"filename": None, "size_bytes": None}
        return {
            "filename": getattr(msg.file, "name", None),
            "size_bytes": getattr(msg.file, "size", None),
        }
    except Exception:
        return {"filename": None, "size_bytes": None}


async def download_attachment(
    dialog_id: int, message_id: int, category: str, progress_cb=None
) -> dict:
    """Télécharge 1 pièce jointe vers le dossier de sa catégorie. Séquentiel.

    progress_cb(done_bytes, total_bytes) est appelé par Telethon ; s'il lève
    CancelledDownload, le téléchargement est interrompu et le partiel supprimé.
    """
    from telethon.errors import FloodWaitError
    from .config import settings

    client = await get_client()
    msg = await client.get_messages(dialog_id, ids=message_id)
    if not msg or not msg.file:
        raise RuntimeError("message sans pièce jointe")
    filename = msg.file.name
    size = msg.file.size
    if not is_allowed_file(filename, settings.max_file_bytes, size):
        raise RuntimeError(f"fichier refusé (extension/taille): {filename} ({size})")
    dest = dest_dir_for(category) / filename

    # Download séquentiel avec gestion FloodWait (rate limiting Telegram).
    attempt = 0
    try:
        while True:
            try:

                def _cb(done, total):
                    if progress_cb:
                        progress_cb(done, total)

                path = await client.download_media(
                    msg, file=str(dest), progress_callback=_cb
                )
                break
            except FloodWaitError as e:
                attempt += 1
                if attempt > 3:
                    raise
                await asyncio.sleep(e.seconds + 5)
    except CancelledDownload:
        dest.unlink(missing_ok=True)
        raise
    return {"filename": filename, "size_bytes": size, "dest_path": str(path or dest)}
