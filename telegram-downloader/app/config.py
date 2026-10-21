"""Config du sidecar telegram-downloader."""

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    # .env chargé depuis le CWD (docker passe par env_file du compose, qui
    # a la priorité : les variables de process écrasent toujours le .env).
    model_config = SettingsConfigDict(env_file=".env", extra="ignore")

    telegram_api_id: int = 0
    telegram_api_hash: str = ""
    telegram_sidecar_api_key: str = "change-me"
    telegram_session_path: str = "/data/telegram/session"
    telegram_db_path: str = "/data/telegram/queue.db"

    ebook_dir: str = "/books/ebooks"
    audiobook_dir: str = "/books/audiobooks"
    manga_dir: str = "/books/manga"
    bd_dir: str = "/books/bd"

    # Staging : les downloads aboutissent ici, puis sont déplacés vers une
    # des bibliothèques ci-dessus depuis l'interface.
    download_dir: str = "/downloads"

    max_file_bytes: int = 2_000_000_000
    poll_interval_seconds: float = 2.0


settings = Settings()
