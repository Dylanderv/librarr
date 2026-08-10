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

**Statut : terminé**

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

**Statut : terminé**

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

**Statut : terminé**

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

**Statut : terminé**

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

## Étape 3b — Repli de renommage quand le déplacement du dossier échoue

**Statut : terminé**

- Bug latent dans `applyPlan` : quand `os.Rename` du dossier échouait, `FolderTo`
  restait renseigné. La boucle des fichiers recalculait alors leur chemin source
  comme si le dossier avait bougé, et tous les renommages échouaient à leur tour.
- La branche « la cible existe déjà » faisait bien le reset : incohérence entre
  deux branches censées faire le même repli.
- `keepFilesInPlace` porte désormais le repli complet (reset de `FolderTo` +
  destinations ramenées à côté de la source), et les deux branches l'appellent.
  Test de non-régression : dossier parent en lecture seule, on vérifie que le
  fichier est bien renommé sur place.

---

## Étape 4 — Métadonnées

**Statut : terminé**

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

Arbitrages pris pendant l'implémentation :

- **Borne de propagation** : `metadataMaxFiles = 500` fichiers par aperçu ou
  application. Lire des métadonnées = ouvrir une archive, soit au moins un
  aller-retour NFS par fichier. Au-delà, la réponse porte `truncated: true` et
  une notice — un aperçu tronqué en silence se lirait comme une couverture
  complète.
- **Champs par fichier vs propageables** : `title`, `series_index` et `isbn`
  identifient un volume ; le serveur les **refuse** (400) quand la cible est un
  dossier, l'UI ne fait que les masquer. Les autres (`series`, `summary`,
  `publisher`, `language`, `authors`, `genres`, `tags`) décrivent l'œuvre et se
  propagent.
- **Informations vs erreurs** : les réponses séparent `notices`
  (`{key, count}`, traduisibles) de `errors` (échecs réels). Le chemin Kavita
  (`applyMetadataToSeries`) continue d'empiler ses `Plan.Skipped` dans `Errors`
  avec le préfixe « not renamed: » ; il n'a pas de test direct, donc il est
  laissé tel quel — dette à traiter avec l'étape 5 ou plus tard.

---

## Étape 5 — Kavita, documentation, vérifications finales

**Statut : terminé**

- `POST /api/files/kavita/scan` — `{root}` → `ScanLibrary`. Déclenché uniquement
  par un bouton, jamais automatiquement.
- Avertissement au démarrage si aucune racine `MANGA_DIR`/`EBOOK_DIR`/`AUDIOBOOK_DIR`
  ne recoupe `KavitaAllowedRoots()` : Browse et l'ancienne page divergeraient.
- Documenter les nouveaux endpoints dans `internal/api/openapi.json`.
  `README.md` seulement si une variable d'environnement apparaît (a priori aucune).
- Vérifications complètes, dont le build `GOOS=linux GOARCH=amd64 CGO_ENABLED=0`.

Arbitrage pris pendant l'implémentation :

- **Portée du scan** : `internal/kavita` n'expose que `ScanLibrary` (POST
  `/api/Library/scan`, sans argument — le même appel que le pipeline d'import) et
  `ScanSeries`, qui réclame un id de série Kavita que Browse n'a pas. Faire
  correspondre `MANGA_DIR` à un id de bibliothèque Kavita exigerait de lister les
  bibliothèques Kavita et d'appliquer `KAVITA_PATH_MAP` à l'envers sur leurs
  dossiers : rien de tel n'existe ici, et apparier par le nom rescannerait
  silencieusement la mauvaise bibliothèque. `root` sert donc à valider que la
  requête vise une bibliothèque réelle, et la réponse annonce `scope: "all"`.
  **Tranché** : le scan global est jugé suffisant, la portée étant annoncée
  explicitement partout (réponse, libellé, openapi, README). Un scan ciblé
  resterait possible en ajoutant `Libraries()` + `ScanLibraryByID()` au client
  Kavita et une résolution inverse de chemin dans `config` — dette identifiée,
  non planifiée.
- **`ScanLibrary` sans `libraryId`** : les versions récentes de Kavita
  documentent `?libraryId=&force=`. La forme actuelle est préexistante (pipeline
  d'import), donc hors périmètre ; à vérifier contre l'instance réelle au
  déploiement. Si les scans d'import deviennent muets, c'est le premier endroit
  à regarder.

---

## Étape 6 — Réparation des tests e2e (dette antérieure)

**Statut : terminé**

Les 3 tests Playwright échouaient déjà sur `main` (sélecteurs `data-idx`, tri
ambigu) — ce n'est pas une régression de ce chantier. Échecs reproduits
localement puis corrigés ; la suite passe désormais en entier (11 tests).

Diagnostic — les trois échecs relèvent de la même catégorie, un test rendu
obsolète par un refactor antérieur du front. **Aucun bug du front, aucune
régression du chantier Browse.**

- `data-idx` / `state.renderedResults` n'existent plus : le front identifie un
  résultat par une clé stable (`data-key` → `state.resultsByKey`, cf. le
  commentaire de `buildResultsByKey` dans `app.js`). Les trois tests
  interrogeaient encore l'ancien schéma. Les sondes sont réécrites sur
  `data-key`, l'invariant testé reste le même : le titre affiché est bien celui
  du résultat vers lequel pointe le bouton de téléchargement.
- Tri ambigu : `[data-action="setSortMode"]` matche deux contrôles — les boutons
  de la barre d'outils (visibles en layout `grid` uniquement) et les en-têtes de
  colonne du tableau (layout `list`, le défaut). Playwright cliquait le bouton
  caché. Le test bascule maintenant explicitement de layout et cible le contrôle
  visible dans chacun — il couvre donc les deux rendus au lieu d'un seul.

Ajouts :

- Test de fumée sur l'onglet Browse : racines affichées, navigation dans un
  dossier, fichier listé, aucune erreur. Le dossier témoin est semé par la
  fixture `app` **avant le démarrage du binaire** — `/api/files/list` cache un
  listing ~30 s, un dossier créé en cours de test pourrait rester invisible.
- `browse` ajouté à la boucle de `test_all_tabs_switch`.

Aucun code applicatif touché (donc pas de bump de `?v=`).

---

## Étape 7 — Recherche de métadonnées en ligne dans Browse

**Statut : terminé**

L'ancien onglet Library sait pré-remplir les métadonnées depuis une source en
ligne ; Browse ne le savait pas. Écart comblé.

**Aucun changement backend.** `GET /api/metadata/search?type=manga|book&q=…`
(`internal/api/metadata_search.go`) est déjà générique et sans couplage Kavita,
et déjà décrit dans `openapi.json`. Le front le consomme tel quel.

Décisions actées avec l'utilisateur avant écriture :

- **Source routée par la racine courante, surchargeable** : `manga` → AniList,
  `ebooks`/`audiobooks` → Open Library. Le sélecteur reste manipulable — un
  light novel rangé dans `ebooks` est légitimement une fiche AniList.
- **Choisir un candidat remplit les champs reçus et coche leur case d'armement.**
  Un champ rempli mais non armé ne serait pas envoyé : remplir sans cocher
  reviendrait à ne rien faire. Les champs absents de la réponse restent vides et
  désarmés. Le pré-remplissage se fait en JS, donc sans événement `change` : la
  sélection invalide explicitement l'aperçu (`resetBrowseMetaPreview`).
- **Panneau dossier uniquement.** AniList et Open Library décrivent une œuvre,
  pas un volume : le panneau fichier n'a rien à y gagner.
- **Rien n'est écrit sans le parcours existant** Aperçu → Appliquer.

Correspondance candidat → champs : `title` → **Série** (c'est une série, et le
serveur refuse `title` sur une cible dossier) ; `summary`, `publisher`,
`language`, `authors`, `genres`, `tags` vers leurs homonymes.

**`year` et `cover_url` sont affichés mais jamais écrits** : `filewriter.Fields`
n'a pas de champ année et le filewriter ne gère pas les couvertures. Ils servent
à identifier le bon candidat, et une mention sous le formulaire le dit — un
champ montré qui n'atterrit nulle part se lit sinon comme une promesse.
`site_url` est rendu en lien `target="_blank" rel="noopener noreferrer"` pour
départager deux homonymes ; cliquer le lien ne sélectionne pas le candidat.

Cas dégradés distingués par code HTTP plutôt que fondus en une « erreur » :
503 = source non configurée sur ce serveur, 502 = source injoignable, autre =
échec de la recherche, zéro candidat = message propre. Requête vide : aucun
appel. Le bouton est désarmé pendant la requête et porte l'état « Recherche… ».

Détails d'implémentation : les lignes de candidats sont construites en API DOM
(titres et résumés viennent d'un tiers et peuvent contenir des guillemets, que
`escapeHtml` n'échappe pas), et `cover_url`/`site_url` sont filtrés sur le
schéma `http(s)` avant rendu.

Tests e2e : deux tests de fumée ajoutés (source routée + requête pré-remplie,
sélection d'un candidat qui remplit et arme, 503 correctement libellé). L'appel
amont est stubbé au niveau du navigateur (`page.route`) : **aucun accès réseau
externe**. Suite complète 13/13.
