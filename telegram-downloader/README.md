# telegram-downloader — sidecar Telegram pour Librarr (read-only)

## Principe
- Login **une fois** avec ton compte user (phone + code + 2FA), session persistée sur volume.
- Ensuite : `GET /dialogs` (canaux/groupes déjà rejoints) → `GET /dialogs/{id}/search?q=` → `POST /downloads`.
- Queue **FIFO persistée (sqlite), 1 download actif**, `FloodWait` respecté (sleep + retry).
- Fichiers écrits dans les volumes partagés `/books/{ebooks,audiobooks,manga}`.
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
curl -H "X-Api-Key: $KEY" -X POST localhost:5051/downloads -H 'Content-Type: application/json' -d '{"dialog_id":-1001234567890,"message_id":123,"category":"ebooks"}'
curl -H "X-Api-Key: $KEY" localhost:5051/downloads
```

## Docker compose (service intégré au compose racine)
Le service `telegram-downloader` est déclaré dans le `docker-compose.yml` racine :
mêmes volumes `/books/*` que `librarr` + volume `telegram-data` (session + file sqlite).
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
