# Plan — Refonte de la page Library (onglet « Browse »)

Branche : `feat/library-browse`. Le contexte complet du chantier est dans
[LIBRARY-REWORK-PROMPT.md](LIBRARY-REWORK-PROMPT.md) — le lire avant d'exécuter
une étape : il contient les pièges connus (NFS, CSP, cache-buster, umask, i18n)
qui s'appliquent à toutes les étapes.

## Décisions actées

Ces douze points ont été arbitrés avec l'utilisateur avant écriture du plan. Ne
pas les remettre en cause sans le lui demander.

| # | Sujet | Décision |
|---|-------|----------|
| 1 | Découverte disque | Lazy, un niveau par requête (`os.ReadDir`, jamais récursif) + cache mémoire TTL. Pas d'index SQLite, pas de scan de fond. |
| 2 | Racines | `MANGA_DIR` / `EBOOK_DIR` / `AUDIOBOOK_DIR`. Pas de nouvelle variable d'environnement. |
| 3 | Navigation UI | Dossier par dossier + fil d'Ariane. Pas d'arbre dépliable. |
| 4 | Propagation dossier → enfants | Écrase les valeurs existantes, mais prévisualisation obligatoire avant application. |
| 5 | Nom de série cible du renommage | Champ pré-rempli avec le nom du dossier, éditable. |
| 6 | Volume non détecté | Ligne d'aperçu avec numéro saisissable ; vide = fichier ignoré. Jamais de devinette. |
| 7 | API | Nouveau préfixe `/api/files/*`, indépendant. `/api/library/metadata/*` reste intact. |
| 8 | Concurrence | Aucun verrou : les garde-fous de `metadata_rename.go` (refus si la cible existe) suffisent. |
| 9 | Colonnes du listing | Nom, taille, date, format. Métadonnées lues seulement à l'ouverture du panneau d'un fichier. |
| 10 | Édition d'un fichier | Métadonnées + nom de fichier en saisie libre (assaini, extension préservée). |
| 11 | PDF / CBR / audio | Listés, badge « renommage uniquement », champs méta grisés avec la raison, inclus dans les plans de renommage. |
| 12 | Kavita | Bouton manuel « Scanner dans Kavita » par bibliothèque. Jamais automatique. |

## Contrat commun aux étapes

- Chaque étape est un commit autonome sur `feat/library-browse`.
- Avant de committer : `gofmt -l ./internal ./cmd`, `go vet ./...`, `go test ./...`,
  et `node --check web/static/js/app.js` si `app.js` est touché.
  Go n'est pas dans le PATH : `export PATH="$PATH:/opt/homebrew/bin"`.
- Ne jamais faire confiance à un plan envoyé par le client : toute application
  recalcule le plan côté serveur à partir du payload d'entrée.
- Ne pas réintroduire de modes de fichier en dur (`0777`/`0666` + umask, cf. prompt).
- Toute modification de `app.js` ou `app.css` impose de bumper `?v=` dans `web/index.html`.

---

## Étape 0 — Les deux bugs bloquants

**Statut : à faire**

- `internal/rename/rename.go` — `VolumeNumber` : normaliser `_` et `.` en espaces
  avant de matcher (après le `TrimSuffix` de l'extension, donc sans risque pour
  `.pdf`). Tests attendus : `Boruto_Two_Blue_Vortex_T4.pdf` → 4,
  `Berserk.Vol.03.cbz` → 3, non-régression sur les cas déjà couverts.
- `internal/api/metadata_kavita.go` — découpler le renommage de
  `res.FilesWritten > 0` : renommer ne dépend pas de l'écriture. Le rescan devient
  conditionné à `FilesWritten > 0 || PathsRenamed > 0` (aujourd'hui une série
  100 % PDF ne rescanne pas non plus).

Indépendant du reste du chantier.

---

## Étape 1 — Racines et résolution de chemin (socle backend)

**Statut : à faire**

- Nouveau fichier `internal/api/files_roots.go` :
  - `libraryRoots(cfg)` → `[{id, label, path}]` construit depuis `MangaDir`,
    `EbookDir`, `AudiobookDir`, en ne gardant que les répertoires existants.
  - `resolveLibraryPath(cfg, p)` → `filepath.Clean` + confinement sous une racine ;
    refus explicite de la racine elle-même pour toute opération d'écriture.
- Les racines sont résolues (`EvalSymlinks`) **une fois** et mises en cache :
  `internal/api/pathvalidate.go` le refait à chaque appel, ce qui est un
  aller-retour NFS inutile. Ne pas modifier `pathvalidate.go` — l'ancienne voie
  reste telle quelle.
- Tests : traversal (`../`), chemin hors racines, racine nue, symlink sortant.

---

## Étape 2 — Navigation (backend + onglet front)

**Statut : à faire**

- `GET /api/files/roots` — les bibliothèques configurées et existantes.
- `GET /api/files/list?path=…` — un seul `os.ReadDir`, jamais récursif. Par
  entrée : `name`, `is_dir`, `size`, `modified`, `kind`
  (`cbz`/`epub`/`pdf`/`cbr`/`audio`/`other`), `editable`.
- Cache mémoire TTL court (~30 s), clé = chemin, invalidé pour le dossier touché
  après chaque écriture de librarr.
- Front : nouvel onglet **Browse** (`tab-browse`) coexistant avec Library —
  liste + fil d'Ariane, `data-action` enregistrés dans `CLICK_ACTIONS`, clés i18n
  au moins en `en`, bump du `?v=`.

Critère de suppression de l'ancien onglet Library : quand Browse couvre lister +
éditer + renommer sur les trois racines. À revalider à ce moment-là.

---

## Étape 3 — Renommage dossier avec aperçu

**Statut : à faire**

- `internal/rename` : ajouter `BuildPlanFor(series string, files []FileInput, mode)`
  avec `FileInput{Path string; Volume *int}` pour porter les volumes saisis à la
  main. `BuildPlan` devient un wrapper — zéro changement de comportement.
- `internal/api/metadata_rename.go` : `checkRenamableDir` et `applyRenamePlan`
  sont câblés en dur sur `KavitaAllowedRoots()` / `ResolveKavitaPath`, donc
  inutilisables avec des racines `MANGA_DIR`. Les paramétrer par une fonction de
  validation ; l'appelant Kavita passe la sienne. Refactor à comportement
  constant, couvert par `metadata_rename_test.go`.
- `POST /api/files/rename/preview` — `{path, series_name, mode, volumes:{file→n}}`
  → `rename.Plan`. `series_name` pré-rempli côté front avec le nom du dossier.
- `POST /api/files/rename/apply` — même payload, **plan recalculé côté serveur**.
- `POST /api/files/rename/file` — `{path, new_name}` : renommage libre d'un
  fichier, `SanitizeSegment`, extension préservée, refus si la cible existe.

---

## Étape 4 — Métadonnées

**Statut : à faire**

- `filewriter.Read(path) (Fields, error)` — brique manquante : le package n'expose
  aujourd'hui que `Apply`. Réutiliser `readZipEntry` et les parseurs
  ComicInfo/OPF existants. `ErrUnsupported` pour PDF/CBR/audio.
- `GET /api/files/metadata?path=…` — fichier unique, appelé seulement à
  l'ouverture du panneau d'édition (décision 9).
- `POST /api/files/metadata/preview` — `{path, recursive, fields}` → par fichier :
  valeur actuelle, valeur cible, `unsupported`. C'est ici que l'écrasement
  (décision 4) devient visible avant validation.
- `POST /api/files/metadata/apply` — même payload, recalcul serveur,
  `filewriter.Apply` fichier par fichier, erreurs collectées sans interrompre le lot.
- Front : panneau d'édition dossier et panneau fichier, champs grisés + raison
  pour les formats non éditables.

---

## Étape 5 — Kavita, documentation, vérifications finales

**Statut : à faire**

- `POST /api/files/kavita/scan` — `{root}` → `ScanLibrary`. Déclenché uniquement
  par un bouton, jamais automatiquement.
- Avertissement au démarrage si aucune racine `MANGA_DIR`/`EBOOK_DIR`/`AUDIOBOOK_DIR`
  ne recoupe `KavitaAllowedRoots()` : Browse et l'ancienne page divergeraient.
- Documenter les nouveaux endpoints dans `internal/api/openapi.json`.
  `README.md` seulement si une variable d'environnement apparaît (a priori aucune).
- Vérifications complètes, dont le build `GOOS=linux GOARCH=amd64 CGO_ENABLED=0`.

---

## Étape 6 — Réparation des tests e2e (dette antérieure)

**Statut : à faire**

Les 3 tests Playwright échouent déjà sur `main` (sélecteurs `data-idx`, tri
ambigu) — ce n'est pas une régression de ce chantier. À réparer dans un commit
séparé, en fin de parcours, pour ne pas mélanger dette antérieure et code neuf.
