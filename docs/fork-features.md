# Fork-specific features

This file documents the features this fork adds on top of upstream librarr
(upstream `README.md` covers the common features; the fork's `README.md` links
here from a short anchor paragraph). Content below was moved verbatim out of the
fork's `README.md` so that future upstream merges to `README.md` stay trivial.

- **Browse tab** -- walk your library folders straight from disk, then rename or edit metadata with a preview first (see below)

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `UMASK` | | Octal umask applied to imported files and directories (e.g. `002`). See [File permissions](#file-permissions). |

### File permissions

Librarr creates imported files and directories with permissive modes (`0666`/`0777`)
and lets the process umask narrow them, the same convention as the linuxserver.io
`*arr` images. This matters when another service has to read or write the same files:

- `UMASK=002` → `0775` directories, `0664` files. Anything in the shared group can
  write. The usual choice when librarr and Kavita/Komga/Calibre run as different
  users but share a group.
- `UMASK=000` → defers entirely to whatever ACL the destination inherits. Useful on
  NAS shares (Synology, unRAID) that already apply an inheritable ACL — a stricter
  mode from librarr would only cap it.
- Unset → the container's default (`022`), i.e. read-only for group and others.

If the media server cannot open what librarr imports, this is almost always the
knob to turn — not a `chown`.

### Renaming on a metadata edit

Kavita names a series after its folder and files whenever the embedded metadata
does not take precedence, so a torrent-derived name
(`Inio.Asano.Anthology.OS.Asano.FR.-TONER-Paprika+`) keeps showing in the library
however good the ComicInfo is. The Browse tab's rename panel can therefore
realign the layout with the **Series** field it writes:

- **Folder only** — `Old.Torrent.Name-GROUP/` → `Bakemono Recchan/`
- **Folder and files** — also renames each file to `<Series> Vol.NN.cbz`, the
  layout Kavita's scanner reads most reliably. A single-file series drops the
  volume suffix.

It is opt-in per edit and shows a preview of every change before saving. Files
whose volume number cannot be read are left alone and listed in the preview
rather than being numbered by guesswork. A folder is never renamed when it is a
library root, when the target already exists, or when the series' files are
spread across several directories.

### Browse tab

The **Browse** tab lists your libraries from the filesystem itself — the
directories set by `MANGA_DIR`, `EBOOK_DIR` and `AUDIOBOOK_DIR` — one folder at a
time. Nothing is read from Kavita, so a series Kavita has not scanned yet, or
has temporarily dropped from its listing, is still visible and still editable.
From there you can rename a folder and its files, rename a single file, and
write metadata into a file or into every file of a folder; both always show a
preview before anything is written. PDF, CBR and audio files are listed and can
be renamed, but their metadata is left alone: no writer exists for them, and
Kavita ignores PDF metadata anyway.

Kavita is only ever contacted on demand, through the **Scan in Kavita** button,
which is hidden when Kavita is not configured. No edit triggers a scan by
itself. The scan is targeted at the library being browsed: librarr lists
Kavita's libraries, translates each one's folders back to local paths through
`KAVITA_PATH_MAP`, and scans those that overlap the browsed root. Libraries are
never matched by name, so a mapping that does not line up fails with an explicit
error naming the paths compared instead of rescanning the wrong library.

For Browse and Kavita to agree, the directories above should overlap the paths
Kavita serves (`KAVITA_LIBRARY_PATH`, `KAVITA_MANGA_LIBRARY_PATH`, the targets of
`KAVITA_PATH_MAP`). When they do not, librarr says so once at startup.

### Library tab

The **Library** tab lists what librarr itself recorded importing — the
`library_items` rows — across its Ebooks, Audiobooks and Manga sub-tabs, with the
filter box searching titles, authors, paths and source ids. The Manga sub-tab used
to mirror Kavita's series list instead; it no longer does, because that made the
rows librarr actually keys its own decisions on unviewable.

That matters for duplicates. A download is refused with HTTP 409 when a
`library_items` row already carries the same source id (info hash, MD5 or GUID),
and removing a torrent from the download client does not remove that row. Such a
row would otherwise be invisible while silently blocking every retry. Removing an
entry here deletes only the row, never the file on disk — which is precisely what
releases the duplicate check.

Browse, by contrast, shows the filesystem. The two answer different questions:
Library says what librarr thinks it has, Browse says what is actually on disk.

## API reference

### Library

Fork delta, complementing the `### Library` table in `README.md`:

All three listing endpoints accept `?q=` to filter on title, author, file path or
source id, and serve librarr's own `library_items` rows (`/api/library` and
`/api/library/audiobooks` defer to Audiobookshelf when it is configured).

| Method | Path | Description |
|--------|------|-------------|
| DELETE | `/api/library/manga/{id}` | Remove manga |

### Files (Browse tab, admin)

Filesystem-backed, independent of Kavita. Full payloads and error codes are in
`/api/openapi.json`.

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/files/roots` | List the browsable library roots |
| GET | `/api/files/list` | List one directory (`?path=`) |
| POST | `/api/files/rename/preview` | Preview a folder + files rename |
| POST | `/api/files/rename/apply` | Apply it (plan recomputed server-side) |
| POST | `/api/files/rename/file` | Rename a single file |
| GET | `/api/files/metadata` | Read one file's metadata (`?path=`) |
| POST | `/api/files/metadata/preview` | Preview a metadata edit, file by file |
| POST | `/api/files/metadata/apply` | Write it |
| POST | `/api/files/kavita/scan` | Ask Kavita to rescan the browsed library (manual only) |
