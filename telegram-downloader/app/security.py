"""Auth simple du sidecar par clé partagée avec Librarr."""

from fastapi import Header, HTTPException

from .config import settings


async def require_api_key(x_api_key: str | None = Header(default=None)) -> None:
    if not settings.telegram_sidecar_api_key:
        return
    if x_api_key != settings.telegram_sidecar_api_key:
        raise HTTPException(status_code=401, detail="invalid api key")
