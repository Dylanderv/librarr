# Chantier — Refonte de la page Library

> Document d'amorçage pour une nouvelle session. Il contient tout le contexte
> nécessaire : rien de la session précédente n'est supposé connu.

## Objectif

Reprendre la page Library de librarr **de zéro**, dans un **nouvel onglet** qui
coexiste avec l'existant. L'ancienne page n'est supprimée qu'une fois la nouvelle
jugée satisfaisante — pas de big-bang.

## Besoin exprimé

1. **Lister les dossiers et les fichiers** des trois bibliothèques : manga, ebooks,
   audiobooks. Une vraie arborescence, pas une liste de séries à plat.
2. **Éditer les métadonnées** à deux granularités :
   - sur un **dossier entier**, en propageant à tous ses enfants ;
   - sur un **fichier isolé**, pour définir manuellement un nom de chapitre ou de
     volume, voire le nom de fichier lui-même.
3. **Renommer**, avec **prévisualisation obligatoire** avant application :
   - en masse sur un dossier et tous ses enfants ;
   - ou sur un fichier seul.

## Contrainte d'architecture

**L'affichage ne doit pas dépendre de Kavita.** La page actuelle est un pur miroir
de Kavita : elle appelle `POST /api/Series/all-v2` et n'a aucune donnée propre, si
bien qu'une série absente de la réponse Kavita est invisible même quand le fichier
est bien sur le disque (constaté : après un renommage, une série a disparu de
librarr pendant plusieurs minutes alors que Kavita l'affichait correctement — le
champ `SortName` sur lequel le tri s'appuie est recalculé de façon asynchrone
après les scans).

La nouvelle page doit donc s'appuyer sur **le système de fichiers comme source de
vérité**.

En revanche, **garder la possibilité de dialoguer avec Kavita** — par exemple
déclencher un scan de bibliothèque après une édition de métadonnées. Ce n'est pas
à implémenter tout de suite, mais l'architecture ne doit pas s'y fermer.

## Méthode attendue

**Questionner avant de coder.** Le besoin ci-dessus est une intention, pas une
spécification. Poser les questions nécessaires pour le cadrer complètement, puis
proposer un plan, et n'écrire du code qu'ensuite.

Quelques axes qui méritent une décision explicite (liste non exhaustive) :

- **Découverte** : parcours du disque à la demande, ou index en base rafraîchi
  périodiquement ? Voir la contrainte NFS ci-dessous — un `WalkDir` récursif à
  chaque requête HTTP sur un partage réseau est probablement rédhibitoire.
- **Modèle d'affichage** : arborescence dépliable, navigation dossier par dossier,
  ou vue hybride ? Que montrer d'un fichier (taille, pages, métadonnées lues) ?
- **Propagation dossier → enfants** : quels champs se propagent, lesquels n'ont de
  sens qu'au fichier (numéro de volume, titre de chapitre) ? Que faire des valeurs
  déjà présentes — écraser, compléter, demander ?
- **Renommage** : quelle convention cible, et jusqu'où va la prévisualisation
  (dossier seul, fichiers, les deux) ? Que faire des fichiers dont on n'arrive pas
  à déduire le volume ?
- **Formats non éditables** : que montrer, et qu'autoriser, pour un PDF ou un CBR
  (voir « Pièges connus ») ?
- **Cohabitation** : la nouvelle page réutilise-t-elle les endpoints existants, ou
  s'appuie-t-elle sur un nouveau préfixe d'API ? Quel critère décidera de la
  suppression de l'ancienne ?
- **Écritures concurrentes** : deux éditions en parallèle sur le même dossier, ou
  une édition pendant un scan Kavita.

## Ce qui existe déjà et doit être réutilisé

Ne pas réécrire ces briques, elles sont testées et en production :

- **`internal/rename`** — planificateur de renommage sans aucun accès disque :
  détection du numéro de volume, assainissement des noms, construction d'un `Plan`
  (`FolderFrom/To`, `Files`, `Skipped`, `Warnings`). C'est ce qui permet d'afficher
  une prévisualisation fidèle : le même plan sert à prévisualiser et à appliquer.
- **`internal/api/metadata_rename.go`** — application d'un plan, avec les
  garde-fous : refus de renommer une racine de bibliothèque, refus si la cible
  existe, repli sur un renommage en place quand le déplacement du dossier échoue.
- **`internal/metadata/filewriter`** — écriture des métadonnées dans les fichiers :
  `ComicInfo.xml` pour les CBZ, OPF pour les EPUB. Les CBZ sont **patchés en place**
  (ajout en fin d'archive + nouveau central directory), ce qui rend l'édition
  quasi instantanée même sur 800 Mo. `ErrUnsupported` signale les formats non
  éditables.
- **`internal/kavita`** — client : `AllSeries`, `FilesForSeries`, `SeriesMetadata`,
  `ScanSeries`, `ScanLibrary`.
- **`Config.ResolveKavitaPath` / `Config.KavitaAllowedRoots`** — traduction des
  chemins Kavita en chemins locaux (`KAVITA_PATH_MAP`) et validation qu'un chemin
  reste sous une racine autorisée. **Tout accès disque déclenché par une requête
  HTTP doit passer par là** : c'est la protection contre le path traversal.

## Pièges connus

- **Kavita ne lit pas les métadonnées PDF** (demande ouverte depuis 2021,
  discussion Kareadita/Kavita#2545). Pour un PDF, Kavita se fie uniquement au nom
  de dossier et de fichier : **le renommage est le seul levier**. Ne pas investir
  dans l'écriture de métadonnées PDF.
- **Les CBR sont des archives RAR**, non réécrivables avec la stdlib Go. Non
  éditables, à afficher comme tels.
- **Deux bugs ouverts dans `internal/rename`**, à corriger dans ce chantier (ils
  bloquent précisément le cas PDF) :
  1. `VolumeNumber` ne gère pas `_` et `.` comme séparateurs. Sur
     `Boruto_Two_Blue_Vortex_T4.pdf`, le `T4` est collé à un underscore ; comme `_`
     est un caractère de mot, `\bt` ne trouve pas de frontière et la détection
     échoue. Or les noms de torrents utilisent massivement les points. Il faut
     normaliser `_` et `.` en espaces avant de matcher.
  2. Dans `internal/api/metadata_kavita.go`, le renommage est conditionné à
     `res.FilesWritten > 0`. Si aucune métadonnée n'a pu être écrite (série
     entièrement en PDF), le renommage est sauté alors qu'il était possible.
     Renommer ne dépend pas de l'écriture : découpler les deux.
- **Le stockage est du NFS**, monté depuis un NAS Synology. Les opérations disque
  sont lentes et sujettes à latence ; l'ACL du partage donne `rwx` à tous, et
  librarr crée en `0777`/`0666` en laissant `UMASK` (à `000` en production)
  trancher. Ne pas réintroduire de modes en dur : cela plafonnerait l'ACL héritée
  et casserait l'accès de Kavita aux fichiers.
- **CSP stricte** (`script-src 'self'`) : aucun JS inline, aucun attribut `on*`.
  Le front passe par de la délégation d'événements — `data-action` mappé dans
  `CLICK_ACTIONS` / `CHANGE_ACTIONS` dans `web/static/js/app.js`.
- **Cache-buster manuel** : `web/index.html` référence `app.js?v=YYYYMMDD` et les
  assets sont servis avec `Cache-Control: max-age=3600`. **Bumper ce `?v=` à chaque
  modification de `app.js` ou `app.css`**, sinon les navigateurs exécutent l'ancien
  bundle contre le nouveau HTML (déjà arrivé, symptômes déroutants : libellés i18n
  affichés comme clés brutes, gestionnaires d'événements absents).
- **i18n** : les libellés vivent dans l'objet `I18N` de `app.js` (locales `en` et
  `ru`), consommés via `data-i18n` et `t()`. Une clé absente s'affiche telle
  quelle. Ajouter au moins la locale `en`, le fallback s'en charge.
- **Les 3 tests e2e Playwright échouent déjà sur `main`** (dette antérieure du
  refactor de l'UI de recherche : `data-idx`, sélecteurs de tri ambigus). Le job
  `e2e` est donc rouge en permanence — il n'y a pas de filet de sécurité front.
  Ne pas prendre ces échecs pour une régression, et envisager de les réparer.

## Environnement de développement

- Dépôt : `/Users/dylandervaux/git-perso/librarr` (fork de `JeremiahM37/librarr`,
  remote `origin` = `Dylanderv/librarr`, `upstream` = l'amont).
- **Go n'est pas dans le PATH par défaut** : `export PATH="$PATH:/opt/homebrew/bin"`.
  **Docker n'est pas disponible** sur la machine de dev.
- Vérifications avant de livrer : `gofmt -l ./internal ./cmd`, `go vet ./...`,
  `go test ./...`, `node --check web/static/js/app.js`, et un build
  `GOOS=linux GOARCH=amd64 CGO_ENABLED=0` (la cible de l'image).
- **Livraison** : l'image `ghcr.io/dylanderv/librarr:latest` est construite
  **uniquement sur un tag `v*`** (`.github/workflows/release.yml`) — un push sur
  `main` ne suffit pas. Convention de tags : `v1.1.3-dylan.N`.
- Documenter les nouveaux endpoints dans `internal/api/openapi.json`, et les
  nouvelles variables d'environnement dans le `README.md`.

## Déploiement cible (pour comprendre les contraintes)

Deux machines. Un NAS Synology héberge les données et qBittorrent ; un serveur
Linux fait tourner librarr et Kavita, et accède aux données du NAS **via NFS**.
Les uid diffèrent des deux côtés (le NAS applique un squash vers `admin`), ce qui
est précisément pourquoi les permissions reposent sur l'ACL du partage et l'umask
plutôt que sur un alignement d'uid.
