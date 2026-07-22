# Plan — Refonte de l'interface de recherche

> Plan autonome destiné à être exécuté dans une nouvelle session. Il contient
> tout le contexte nécessaire (findings, décisions, fichiers, étapes, tests).
> **Aucune modification backend requise** — c'est une refonte 100 % front.

## Objectif

Refondre l'affichage des résultats de recherche pour :

1. **Ajouter une vue « Liste » (table)** en plus des cards, avec les mêmes infos
   (vignette, seeders/leechers, source, format, taille, indexer…) mais en table.
2. **Regrouper par source en swimlanes repliables**, chaque lane affichant son
   état de recherche (en cours / terminé / circuit ouvert / erreur), son
   compteur de résultats et son temps de réponse.

Problème UX actuel à résoudre : les résultats s'ajoutent/re-trient « au milieu »
de façon asynchrone → liste instable et difficile à lire. Les swimlanes règlent
ça car chaque source remplit **sa** lane, pas une liste globale.

## Décisions verrouillées (validées avec l'utilisateur)

- **Deux bascules orthogonales** (mémorisées en `localStorage`) :
  - **Disposition** : `grid` (cards) / `list` (table)
  - **Regroupement** : `flat` (à plat) / `source` (swimlanes)
- **Vue par défaut** : `list` + `source` (swimlanes en table).
- **Tri par clic sur les en-têtes de colonnes** en vue table (remplace les
  boutons de tri `relevance/seeders/size` actuels). En swimlane, le tri
  s'applique **dans** chaque lane.

## Findings — état du code actuel

### Backend (déjà prêt, NE PAS modifier)

`internal/api/search.go` — `handleSearchStreamTab` (≈ L71-148) émet en SSE :

- event `started` → `{ results: [], sources: SourceMeta() }`
- event `results` (à chaque source qui répond) →
  `{ source, source_done, circuit_open, results, search_time_ms, sources }`
  - ⚠️ `results` = **liste accumulée complète** re-traitée à chaque frame
    (`ProcessResults(allResults)`), pas un delta.
- event `complete` → `{ results, sources }`

`SourceMeta()` (`internal/search/searcher.go` L253) renvoie par source :
`{ name, label, enabled, search_tab, download_type, health? }`.
⚠️ Renvoie **toutes** les sources (tous onglets) → le front doit filtrer sur
`search_tab === onglet courant` **et** `enabled`.

### Frontend (à refondre) — `web/static/js/app.js` (~2682 lignes)

- `doSearch` (L982) → `doStreamingSearch` (L1026) → SSE. Fallback `doJsonSearch`.
- **Le front jette les métadonnées SSE** : `doStreamingSearch` (L1047-1050) ne lit
  que `evt.data.results`, ignore `source`, `source_done`, `circuit_open`,
  `sources`, `search_time_ms`.
- `updateSearchResults` (L1078) écrase `state.searchResults` et appelle
  `renderSearchResults`.
- `renderSearchResults` (L1152) : **`container.innerHTML = ...` intégral à chaque
  frame** → perte de scroll / état / flicker (cœur du problème).
- `renderBookCard` (L1159) : rendu d'une card. Utilise `data-idx` = index dans
  `state.renderedResults` (tableau trié, L1155).
- Tri : `setSortMode` (L1122), `sortResults` (L1131), `parseSize` (L1141).
- Skeleton : `showSearchSkeleton` (L1100) — 12 cards globales.
- **Câblage des clics** : délégation globale sur `[data-action]` (L2652). Le bouton
  download porte `data-action="startDownload" data-idx="${index}"` (L1214) et le
  handler (L2638-2639) résout via `state.renderedResults[+el.dataset.idx]`.
  → **Ce schéma d'index casse avec plusieurs listes (swimlanes).**
- `getDownloadKey` (L1250) : clé stable d'un résultat (déjà utilisée pour
  `pendingDownloads` / `downloadOutcomes`).
- Placeholders de couverture : `data-ph-idx` / `data-ph-title` (L1168).
- State pertinent : `pendingDownloads` (Set, L475), `downloadOutcomes`
  (Map, L477), `searchResults` (L474), `renderedResults`, `sortMode`, `searchTab`.
- `SOURCE_COLORS` (L490) : couleurs/labels par source.

### DOM — `web/index.html` (zone recherche)

- `#search-spinner` (L173), `#search-sort-bar` (L179) + `#search-result-count`
  (L180) + boutons `.sort-btn` (L182-184), `#search-skeleton` (L189, grid),
  `#search-empty` (L192), `#search-no-results` (L198),
  `#search-results` (L204, `grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4`).

## Pivot technique (à faire en premier)

### A. Identité stable des résultats (remplace `data-idx`)

- Introduire une **clé stable** par résultat : réutiliser `getDownloadKey(result)`
  (ou un hash `title|source|url` si collisions possibles). Vérifier l'unicité ;
  si deux résultats partagent la même clé, suffixer un compteur.
- Remplacer `state.renderedResults` (array indexé) par
  **`state.resultsByKey`** (`Map<key, result>`), reconstruite à chaque update.
- Boutons/éléments cliquables : `data-idx` → **`data-key`**.
- Handler `startDownload` (L2638-2639) et placeholders : résoudre par
  `state.resultsByKey.get(el.dataset.key)`.
- ⚠️ Préserver intégralement la logique download existante
  (`pendingDownloads`, `downloadOutcomes`, retry Anna's Archive).

### B. Rendu incrémental « keyed » (fin du `innerHTML` global)

- Un conteneur **par lane** (`data-lane="<source>"`) ; ne re-render que les lanes
  dont les données ont changé sur le frame courant (`evt.data.source`).
- Préserver l'état replié/déplié (`state.collapsedLanes: Set<source>`) et le
  scroll entre frames.
- En vue `flat` : un seul conteneur, mais éviter le re-render destructif —
  soit diff keyed léger, soit ne re-trier/re-render qu'au `complete` en
  insérant de façon stable pendant le stream.

## Étapes d'implémentation

### 1. Exploiter les métadonnées SSE

- Dans `doStreamingSearch`, traiter les events `started` / `results` / `complete`
  et alimenter :
  - `state.sources` : skeleton des lanes (filtré `search_tab === state.searchTab`
    + `enabled`), dans l'ordre de `SourceMeta`.
  - `state.sourceStatus: Map<source, {status, count, timeMs, error}>` où
    `status ∈ {searching, done, circuit, error}`. Marquer `done` quand un frame
    porte `source_done=true` pour cette source ; `circuit` si `circuit_open` ;
    `error` si `error` présent.
- Conserver le fallback `doJsonSearch` (mode dégradé sans lanes → tout dans une
  pseudo-lane « Résultats » ou vue flat).

### 2. State + persistance des préférences

- Ajouter à `state` : `viewLayout` (`'grid'|'list'`), `viewGroup`
  (`'flat'|'source'`), `collapsedLanes` (Set), `sources`, `sourceStatus`,
  `resultsByKey`.
- Init depuis `localStorage` (`librarr.search.viewLayout`,
  `librarr.search.viewGroup`) avec défaut `list` + `source`.
- Setters qui persistent puis re-render : `setViewLayout`, `setViewGroup`,
  `toggleLane(source)`.

### 3. Barre d'outils (index.html)

- Remplacer/compléter `#search-sort-bar` par deux toggles segmentés :
  - Disposition : `Grille | Liste`
  - Regroupement : `À plat | Par source`
- En vue `list`, masquer les boutons `.sort-btn` (tri par colonne à la place).
  En vue `grid`, garder des boutons de tri (relevance/seeders/size).
- Utiliser le même style de toggle segmenté que l'existant (`setSortMode`).

### 4. Rendu — pivot identité + keyed

- Implémenter A puis B ci-dessus.
- Nouveau routeur de rendu `renderSearchResults()` :
  - construit `state.resultsByKey`,
  - si `viewGroup === 'source'` → `renderSwimlanes()`,
  - sinon → `renderFlat()`,
  - chaque branche respecte `viewLayout` (`grid`→cards, `list`→table).

### 5. Vue table + tri par colonne

- `renderBookRow(result)` → `<tr>` avec colonnes :
  vignette · titre · auteur · source · format · taille · seeders/leechers ·
  indexer · action (bouton download avec `data-key`).
- En-têtes cliquables → tri (`state.sortMode` + sens asc/desc) ; en swimlane, tri
  **intra-lane**. Réutiliser/adapter `sortResults` + `parseSize`.
- Envelopper la table dans `overflow-x-auto` (responsive mobile).
- Vignette : réutiliser `cover_url` sinon placeholder (`makePlaceholderHtml`),
  en petit format (ex. `h-12 w-9`).

### 6. Swimlanes

- `renderSwimlanes()` : pour chaque source de `state.sources` (ordre stable),
  une lane avec :
  - **En-tête** : label source + badge d'état
    (⏳ `lane_searching` / ✓ `lane_done` avec N / ⚠ `lane_circuit` / ✗ `lane_error`)
    + `search_time_ms` + chevron repli (`toggleLane`).
  - **Corps** (masqué si repliée) : cards (grid) ou table (list) des résultats de
    cette source, triés intra-lane.
  - Sources à **0 résultat** : lane affichée **repliée** avec « 0 » (transparence,
    ne pas cacher).
- Ne re-render que la lane concernée par le frame quand c'est possible.

### 7. Skeleton & états vides

- Skeleton **par lane** (petites lignes) au lieu des 12 cards globales.
- `#search-no-results` global affiché seulement si **toutes** les lanes sont
  terminées avec 0 résultat.
- Spinner global : piloté par « au moins une source encore `searching` ».

### 8. i18n

Ajouter les clés dans les blocs de traduction (au minimum `en` ; fallback pour
les autres langues, ex. `ru` déjà présent) :
`view_grid`, `view_list`, `group_flat`, `group_source`,
`lane_searching`, `lane_done`, `lane_circuit`, `lane_error`,
`col_title`, `col_author`, `col_source`, `col_format`, `col_size`,
`col_seeders`, `col_indexer`, `col_action`.

## Fichiers touchés

- `web/static/js/app.js` — gros du travail (rendu, state, SSE, i18n, handlers).
- `web/index.html` — barre d'outils + conteneurs de résultats.
- `web/static/css/app.css` — styles table/lane/badges si besoin (sinon Tailwind
  inline comme le reste).
- **Aucun fichier Go.**

## Risques / points d'attention

- **Ne pas casser le download** : re-câbler `startDownload`, `pendingDownloads`,
  `downloadOutcomes`, retry Anna's Archive sur `data-key` (pas `data-idx`).
  Idem pour les placeholders (`data-ph-*` → clé).
- **Préserver collapse + scroll** entre frames SSE (pas de `innerHTML` global).
- **Fallback JSON** (`doJsonSearch`) : garder un rendu correct sans métadonnées de
  source (une seule lane ou vue flat).
- **Filtrage `SourceMeta`** sur l'onglet courant + `enabled`.
- **Unicité des clés** de résultat.
- Le stream renvoyant la liste **complète** à chaque frame, le regroupement
  client se fait par simple `group by result.source` sur `state.searchResults` —
  pas de gestion de delta ni de dédup côté front.

## Checklist de tests (après implémentation)

- [ ] Recherche ebooks/audiobooks/manga : les lanes apparaissent, se remplissent,
      passent en ✓ ; spinner global s'éteint au `complete`.
- [ ] Bascule Grille/Liste et À plat/Par source : rendu correct, préférence
      persistée après reload.
- [ ] Repli/dépli des lanes conservé pendant que d'autres sources arrivent.
- [ ] Tri par colonne (list) et boutons de tri (grid) fonctionnent, y compris
      intra-lane.
- [ ] Download : clic → état loading → success/erreur, sur toutes les vues ;
      retry Anna's Archive OK ; dédup « Duplicate detected » OK.
- [ ] Source en circuit ouvert / en erreur : badge correct, pas de crash.
- [ ] Source à 0 résultat : lane repliée « 0 ».
- [ ] Aucune source ne renvoie rien : `#search-no-results` global.
- [ ] Responsive mobile : table scroll horizontal, pas de débordement.
- [ ] Changement d'onglet : reset propre (cf. L934-948).

## Rappel environnement (session locale)

- Lancer l'app en local : `./run-local.sh` (charge `.env`, crée `local-data/`).
- Front embarqué via `//go:embed` (`web/embed.go`) → en dev, `go run` sert bien
  les fichiers `web/` du disque au build ; recompiler/redémarrer pour voir les
  changements (pas de hot-reload).
- Le flux « ajout torrent qBittorrent » n'est pas testable proprement en local
  (topologie gluetun/VPN) — sans impact ici, refonte purement UI de recherche.
