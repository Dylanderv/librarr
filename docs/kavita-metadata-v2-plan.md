# Kavita metadata editor — V2 plan (online metadata source)

**Goal:** let the user fetch metadata from an online database to pre-fill the
bulk metadata editor built in V1. Manga → **AniList**; ebooks → reuse the
existing **Open Library** client.

This doc is self-contained so a fresh session can implement without re-research.

## Context: what V1 already shipped

Branch `feat/kavita-metadata-editor` (commit `d501a3a`). Bulk editor that writes
metadata **into the files** (EPUB OPF / CBZ ComicInfo.xml), then rescans Kavita.

- `internal/kavita` — Kavita client (login/token cache, `SeriesMetadata`,
  `FilesForSeries`, `ScanSeries`, `AllSeries`, `SearchSeries`).
- `internal/metadata/filewriter` — atomic file writers. Target type is
  `filewriter.Fields{Title,Series,SeriesIndex,Summary,Publisher,Language,ISBN
  *string; Authors,Genres,Tags []string; AppendLists bool}`.
- `config.ResolveKavitaPath` + `KAVITA_PATH_MAP` — map Kavita paths → librarr fs.
- API (admin): `GET /api/library/manga/series/{id}/metadata`,
  `POST /api/library/metadata/bulk`.
- Frontend: select mode + selection bar on manga tab, bulk-edit modal
  (`#metadata-modal`, rows with `.meta-enable`/`.meta-value`, list-mode radios).
  Actions wired in `CLICK_ACTIONS`; `openMetadataModal()` already prefills from
  the Kavita GET endpoint when exactly one series is selected.

## Reuse points (verified)

**`internal/metadata/openlibrary.go`** — mirror its style for the AniList client:
- `Client` holds one `*http.Client` (timeout 15s); `NewClient(httpClient)`.
- In-memory cache `map[string]cacheEntry` under `sync.RWMutex`, TTL 30 min.
- Requests: `http.NewRequestWithContext`, User-Agent
  `"Librarr/2.0 (book download manager; github.com/JeremiahM37/librarr)"`,
  `defer resp.Body.Close()`, status check, streamed `json.NewDecoder`.
- Entry point: `FetchMetadataCtx(ctx, title, author) (*BookMetadata, error)`
  (returns `(nil,nil)` when nothing found). No by-ISBN / by-ID fetch exists.
- Consumed once: `internal/api/search.go:48` (`handleSearchTab`) attaches a
  `metadata` key to the search JSON. Streaming path does NOT attach metadata.
- Server wiring: `metadataClient *metadata.Client` on `Server`
  (`router.go:42`, instantiated `router.go:114`).

## AniList API facts

- **POST** `https://graphql.anilist.co`, GraphQL, **no auth** for public reads.
- Rate limit ~90 req/min (has been degraded to 30/min); burst limiter; `429`
  returns `Retry-After` + `X-RateLimit-*` headers → cache + one retry-on-429.
- Differs from Open Library's GET style: marshal a JSON body
  `{"query": "...", "variables": {...}}`, set `Content-Type: application/json`.

Search query to use:

```graphql
query ($search: String) {
  Page(perPage: 10) {
    media(search: $search, type: MANGA) {
      id
      title { romaji english native }
      description(asHtml: false)
      genres
      tags { name rank isGeneralSpoiler isAdult }
      staff { edges { role node { name { full } } } }
      startDate { year }
      status
      format
      countryOfOrigin
      chapters
      volumes
      synonyms
      isAdult
      coverImage { large }
      siteUrl
    }
  }
}
```

Field mapping AniList → `filewriter.Fields` / candidate:
- authors = staff edges whose `role` contains "Story" (incl. "Story & Art");
  fall back to "Art" if no story credit.
- Genres = `genres`. Tags = `tags[].name` excluding `isGeneralSpoiler`/`isAdult`
  (and maybe rank < 50 to cut noise).
- Summary = `description` with HTML stripped (`<br>`, tags, unescape entities).
- Year = `startDate.year`. Title: prefer english → romaji.

## Implementation steps

1. **`internal/metadata/anilist/client.go`** — new package mirroring openlibrary:
   - `Client` + `NewClient(*http.Client)`, cache, User-Agent, 429/Retry-After.
   - `SearchManga(ctx, title) ([]MangaMatch, error)` returning up to ~10
     candidates. `MangaMatch{AniListID, Title, Summary, Authors, Genres, Tags,
     Year, CoverURL, SiteURL, Chapters, Volumes, Format}`.
   - Unit test with a stubbed transport (sample AniList JSON) — no live calls.

2. **API (admin):** `GET /api/metadata/search?type=manga|book&q=...`
   → `{candidates: [...]}`. `manga` → AniList; `book` → Open Library (adapt its
   single result into the same candidate shape). Register in
   `registerLibraryRoutes`, keep `openapi.json` in sync.

3. **Frontend (`web/static/js/app.js` + `web/index.html`):**
   - In `#metadata-modal`, add a search input + "Search online" button and a
     candidate list (cover + title + year + siteUrl link).
   - On candidate click, prefill the modal rows (reuse `prefillMetadataField`):
     authors/genres/tags/summary/(publisher/language if present).
   - **Gate "Search online" to exactly one selected series** (online match per
     series only; bulk multi-series stays manual). Show the button only when
     `state.selectedSeries.size === 1`.
   - Add EN + RU i18n keys (both dicts — missing keys render as raw text).

4. Route AniList vs Open Library by the library tab / media type. Consider
   attaching AniList (not Open Library) metadata on the **manga** search tab
   too — current behavior feeds book metadata there, which is a poor fit.

## Decisions locked in

- AniList for manga (free, no key, aligns with Kavita+), Open Library for ebooks.
- Candidate-selection flow, NOT auto-match — avoids wrong matches, esp. in bulk.
- Respect rate limits (cache + 429 handling).
- No Comic Vine in V2 (needs an API key); revisit only for western comics.
