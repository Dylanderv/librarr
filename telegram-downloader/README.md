# telegram-downloader — sidecar Telegram pour Librarr (read-only)

## Principe
- Login **une fois** avec ton compte user (phone + code + 2FA), session persistée sur volume.
- Ensuite : `GET /dialogs` (canaux/groupes déjà rejoints) → `GET /dialogs/{id}/search?q=` → `POST /downloads`.
- Queue **FIFO persistée (sqlite), 1 download actif**, `FloodWait` respecté (sleep + retry).
- Fichiers écrits dans le dossier de **staging** `/downloads` (download en `.tmp.*`,
  renommé une fois complet — `clear_tmp()` purge les partielles au boot).
  Aucune catégorie n'est choisie au moment du download.
- Tri ensuite depuis l'interface, **indépendamment de la file** (le cycle de
  vie des jobs s'arrête à `done` : le tri ne change jamais leur statut) :
  - `GET /staging` : fichiers en attente de tri ; `GET /staging/tree` : sous-dossiers existants par bibliothèque.
  - `POST /staging/move {files: [...], dest: "ebooks/Auteurs/Herbert"}` :
    déplace en masse staging → bibliothèque, **crée les sous-dossiers
    manquants**, refuse d'écraser (résultat par fichier : `moved` / `skipped`).
  - `POST /library/move {files: ["manga/One Piece/tome1.cbz"], dest: "staging"}`
    : remet un fichier mal rangé au staging (ou change un fichier de dossier de
    bibliothèque via un `dest` bibliothèque) ; collision → suffixe ` (1)`.
  - `DELETE /staging/{filename}` : supprime un fichier indésirable.
- **Read-only** : jamais de `send`, `join/leave`, réaction, `mark_read`. Garde-fou : `tests/test_readonly.py`.

## Lancer en local
```bash
cd telegram-downloader
cp .env.example .env   # renseigne TELEGRAM_API_ID / HASH / SIDECAR_API_KEY
python3 -m venv .venv && source .venv/bin/activate
pip install -e ".[test]"
uvicorn app.main:app --port 5051
```

## Auth initiale
```bash
curl -H "X-Api-Key: $KEY" -X POST localhost:5051/auth/code/request -H 'Content-Type: application/json' -d '{"phone":"+336..."}'
curl -H "X-Api-Key: $KEY" -X POST localhost:5051/auth/code/confirm -H 'Content-Type: application/json' -d '{"phone":"+336...","code":"12345","password":"2fa-si-requis"}'
curl -H "X-Api-Key: $KEY" localhost:5051/auth/status
```

## Flow type
```bash
curl -H "X-Api-Key: $KEY" localhost:5051/dialogs
curl -H "X-Api-Key: $KEY" 'localhost:5051/dialogs/-1001234567890/search?q=dune&limit=20'
# Télécharge (vers le staging /downloads)
curl -H "X-Api-Key: $KEY" -X POST localhost:5051/downloads -H 'Content-Type: application/json' -d '{"dialog_id":-1001234567890,"message_id":123}'
curl -H "X-Api-Key: $KEY" localhost:5051/downloads
# Tri depuis le staging
curl -H "X-Api-Key: $KEY" localhost:5051/staging
curl -H "X-Api-Key: $KEY" localhost:5051/staging/tree
# Déplacement en masse : les sous-dossiers manquants sont créés au move.
curl -H "X-Api-Key: $KEY" -X POST localhost:5051/staging/move -H 'Content-Type: application/json' \
  -d '{"files":["dune.epub","autre.epub"],"dest":"ebooks/Auteurs/Herbert"}'
curl -H "X-Api-Key: $KEY" -X DELETE localhost:5051/staging/fichier-indesirable.epub
```

## Docker compose (service intégré au compose racine)
Le service `telegram-downloader` est déclaré dans le `docker-compose.yml` racine :
mêmes volumes `/books/*` que `librarr` + staging `/downloads` (monté depuis
`/data/media/books/.telegram-incoming`, même FS que `/books/*` → rename
atomique) + volume `telegram-data` (session + file sqlite).
Renseigne `TELEGRAM_API_ID` / `TELEGRAM_API_HASH` / `TELEGRAM_SIDECAR_API_KEY`
dans le `.env` racine (voir `.env.example`), puis :
```bash
docker compose up -d --build telegram-downloader
```
### Réutiliser la session locale (pas de re-login)
Copie le `.session` existant dans le volume :
```bash
docker cp telegram-downloader/local-data/telegram/session.session \
  telegram-downloader:/data/telegram/session.session
docker compose restart telegram-downloader
```
Vérifie ensuite `GET :5051/auth/status` → `{"authorized":true}`.
