"""Schémas API du sidecar."""

from enum import Enum
from pydantic import BaseModel


class Category(str, Enum):
    ebooks = "ebooks"
    audiobooks = "audiobooks"
    manga = "manga"


class JobStatus(str, Enum):
    queued = "queued"
    downloading = "downloading"
    done = "done"
    error = "error"
    cancelled = "cancelled"


class DialogOut(BaseModel):
    id: int
    title: str
    username: str | None = None
    kind: str  # "channel" | "group"


class MessageAttachmentOut(BaseModel):
    dialog_id: int
    message_id: int
    date: str | None = None
    text_snippet: str | None = None
    filename: str | None = None
    mime_type: str | None = None
    size_bytes: int | None = None


class EnqueueDownloadIn(BaseModel):
    dialog_id: int
    message_id: int
    category: Category


class JobOut(BaseModel):
    id: str
    dialog_id: int
    message_id: int
    category: Category
    filename: str | None = None
    size_bytes: int | None = None
    status: JobStatus
    progress: float = 0.0
    dest_path: str | None = None
    error: str | None = None


class CodeRequestIn(BaseModel):
    phone: str


class CodeConfirmIn(BaseModel):
    phone: str
    code: str
    password: str | None = None  # mot de passe 2FA si requis
