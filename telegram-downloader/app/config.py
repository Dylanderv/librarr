"""Config du sidecar telegram-downloader."""

from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    telegram_api_id: int = 0
    telegram_api_hash: str = ""
    telegram_sidecar_api_key: str = "change-me"
    telegram_session_path: str = "/data/telegram/session"
    telegram_db_path: str = "/data/telegram/queue.db"

    ebook_dir: str = "/books/ebooks"
    audiobook_dir: str = "/books/audiobooks"
    manga_dir: str = "/books/manga"

    max_file_bytes: int = 2_000_000_000
    poll_interval_seconds: float = 2.0


settings = Settings()
