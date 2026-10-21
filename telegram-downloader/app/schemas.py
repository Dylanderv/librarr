"""Schémas API du sidecar."""

from enum import Enum
from pydantic import BaseModel


class Category(str, Enum):
    """Bibliothèques de destination (choisies après le téléchargement)."""

    ebooks = "ebooks"
    audiobooks = "audiobooks"
    manga = "manga"
    bd = "bd"


class JobStatus(str, Enum):
    """La file ne suit QUE le téléchargement — le tri n'y apparaît pas."""

    queued = "queued"
    downloading = "downloading"
    done = "done"  # fichier complet déposé dans le staging
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
    """Le download aboutit toujours au staging ; le tri vient ensuite."""

    dialog_id: int
    message_id: int


class JobOut(BaseModel):
    id: str
    dialog_id: int
    message_id: int
    # Toujours 'staging' : la file ne suit pas le tri des fichiers.
    category: str
    filename: str | None = None
    size_bytes: int | None = None
    status: JobStatus
    progress: float = 0.0
    dest_path: str | None = None
    error: str | None = None


class StagedFileOut(BaseModel):
    filename: str
    size_bytes: int
    modified: float


class MoveIn(BaseModel):
    """Déplacement en masse du staging vers une bibliothèque.

    dest : 'categorie' ou 'categorie/sous/dossier' — les sous-dossiers
    manquants sont créés. Refus si la cible existe déjà (jamais d'écrasement).
    """

    files: list[str]
    dest: str


class LibraryMoveIn(BaseModel):
    """Déplacement depuis une bibliothèque : retour au staging ('staging')
    ou changement de dossier ('categorie/sous/dossier')."""

    files: list[str]
    dest: str


class DestIn(BaseModel):
    """Création de sous-dossier : 'categorie/sous/dossier'."""

    dest: str


class CodeRequestIn(BaseModel):
    phone: str


class CodeConfirmIn(BaseModel):
    phone: str
    code: str
    password: str | None = None  # mot de passe 2FA si requis
