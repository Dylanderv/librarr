package api

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/JeremiahM37/librarr/internal/config"
)

// Browsing endpoints for the /api/files/* surface.
//
// Discovery is lazy: one os.ReadDir per request, never recursive. Storage is a
// slow NFS mount, so a WalkDir behind an HTTP handler would be unusable on a
// library of any size. Listings carry only what ReadDir/Stat already give
// (name, size, mtime, extension) — no archive is opened here; metadata is read
// on demand when a file panel is opened.

// dirListingTTL is how long a directory listing stays served from memory. Short
// enough that a change made outside librarr shows up quickly, long enough to
// absorb the back-and-forth of browsing (breadcrumb, back, re-enter).
const dirListingTTL = 30 * time.Second

// dirListingCacheMax bounds the cache so browsing a large library cannot grow
// it without limit. Reaching it triggers a sweep of expired entries, and a full
// flush if that was not enough.
const dirListingCacheMax = 512

// fileEntry is one row of a directory listing.
//
// Kind is empty for directories: it describes a media format, and a directory
// has none. Editable means librarr can write metadata into the file, i.e.
// internal/metadata/filewriter supports it — CBZ (ComicInfo.xml) and EPUB (OPF)
// only. Everything else is rename-only.
type fileEntry struct {
	Name     string `json:"name"`
	IsDir    bool   `json:"is_dir"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
	Kind     string `json:"kind"`
	Editable bool   `json:"editable"`
}

// pathSegment is one clickable step of the breadcrumb.
type pathSegment struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// listResponse is the payload of GET /api/files/list.
type listResponse struct {
	Root      string        `json:"root"`
	RootLabel string        `json:"root_label"`
	RootPath  string        `json:"root_path"`
	Path      string        `json:"path"`
	IsRoot    bool          `json:"is_root"`
	Parent    string        `json:"parent"`
	Segments  []pathSegment `json:"segments"`
	Entries   []fileEntry   `json:"entries"`
}

// dirListingCache memoises directory listings by absolute path.
//
// Writes performed by librarr must invalidate the folders they touch —
// see invalidateDirListing.
type dirListingCacheEntry struct {
	entries []fileEntry
	expires time.Time
}

var (
	dirListingMu    sync.Mutex
	dirListingCache = map[string]dirListingCacheEntry{}
)

// invalidateDirListing drops the cached listing of dir.
//
// COUPLING POINT: every librarr operation that creates, renames, deletes or
// resizes a file must call this for each directory it changed (for a rename
// across folders: both the source and the destination directory), otherwise the
// UI keeps showing the pre-write state for up to dirListingTTL. It is a no-op
// when nothing is cached, so it is always safe to call.
func invalidateDirListing(dir string) {
	if dir == "" {
		return
	}
	clean := filepath.Clean(dir)
	dirListingMu.Lock()
	delete(dirListingCache, clean)
	dirListingMu.Unlock()
}

// resetDirListingCache empties the cache. Tests only.
func resetDirListingCache() {
	dirListingMu.Lock()
	dirListingCache = map[string]dirListingCacheEntry{}
	dirListingMu.Unlock()
}

// cachedDirListing returns a cached listing for dir if it has not expired.
func cachedDirListing(dir string, now time.Time) ([]fileEntry, bool) {
	dirListingMu.Lock()
	defer dirListingMu.Unlock()
	e, ok := dirListingCache[dir]
	if !ok || now.After(e.expires) {
		return nil, false
	}
	return e.entries, true
}

// storeDirListing caches a listing for dir.
func storeDirListing(dir string, entries []fileEntry, now time.Time) {
	dirListingMu.Lock()
	defer dirListingMu.Unlock()
	if len(dirListingCache) >= dirListingCacheMax {
		for k, v := range dirListingCache {
			if now.After(v.expires) {
				delete(dirListingCache, k)
			}
		}
		if len(dirListingCache) >= dirListingCacheMax {
			dirListingCache = map[string]dirListingCacheEntry{}
		}
	}
	dirListingCache[dir] = dirListingCacheEntry{entries: entries, expires: now.Add(dirListingTTL)}
}

// audioExts are the audiobook containers we recognise. Listed only: none of
// them is writable by internal/metadata/filewriter.
var audioExts = map[string]bool{
	".mp3": true, ".m4a": true, ".m4b": true, ".flac": true,
	".ogg": true, ".opus": true, ".aac": true, ".wav": true,
	".wma": true, ".alac": true,
}

// classifyFile maps a file name to its kind and whether librarr can write
// metadata into it.
func classifyFile(name string) (kind string, editable bool) {
	switch ext := strings.ToLower(filepath.Ext(name)); ext {
	case ".cbz":
		return "cbz", true
	case ".epub":
		return "epub", true
	case ".pdf":
		return "pdf", false
	case ".cbr":
		return "cbr", false
	default:
		if audioExts[ext] {
			return "audio", false
		}
		return "other", false
	}
}

// isHiddenEntry reports whether an entry should be kept out of listings.
// Dotfiles are noise, and @eaDir is the thumbnail/index sidecar Synology drops
// in every shared folder — neither is browsable content.
func isHiddenEntry(name string) bool {
	return strings.HasPrefix(name, ".") || name == "@eaDir"
}

// naturalLess compares two names case-insensitively, treating digit runs as
// numbers so "Vol.2" sorts before "Vol.10".
func naturalLess(a, b string) bool {
	la, lb := strings.ToLower(a), strings.ToLower(b)
	i, j := 0, 0
	for i < len(la) && j < len(lb) {
		ca, cb := la[i], lb[j]
		if isDigit(ca) && isDigit(cb) {
			si, sj := i, j
			for i < len(la) && isDigit(la[i]) {
				i++
			}
			for j < len(lb) && isDigit(lb[j]) {
				j++
			}
			na := strings.TrimLeft(la[si:i], "0")
			nb := strings.TrimLeft(lb[sj:j], "0")
			if len(na) != len(nb) {
				return len(na) < len(nb)
			}
			if na != nb {
				return na < nb
			}
			continue
		}
		if ca != cb {
			return ca < cb
		}
		i++
		j++
	}
	if len(la)-i != len(lb)-j {
		return len(la)-i < len(lb)-j
	}
	// Identical once lowercased: fall back to the raw names for a total order.
	return a < b
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

// readDirEntries performs the single, non-recursive read of dir.
func readDirEntries(dir string) ([]fileEntry, error) {
	raw, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	entries := make([]fileEntry, 0, len(raw))
	for _, de := range raw {
		name := de.Name()
		if isHiddenEntry(name) {
			continue
		}
		e := fileEntry{Name: name, IsDir: de.IsDir()}
		// Info() is an lstat on most platforms, so symlinks are described as
		// links rather than followed — consistent with the lexical confinement
		// resolveLibraryPath applies. An entry whose stat fails (vanished
		// mid-listing, permission hole) is still listed, without size or date.
		if info, ierr := de.Info(); ierr == nil {
			e.Modified = info.ModTime().UTC().Format(time.RFC3339)
			if !e.IsDir {
				e.Size = info.Size()
			}
		}
		if !e.IsDir {
			e.Kind, e.Editable = classifyFile(name)
		}
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return naturalLess(entries[i].Name, entries[j].Name)
	})
	return entries, nil
}

// listDir returns the listing of dir, from cache when fresh.
func listDir(dir string) ([]fileEntry, error) {
	now := time.Now()
	if entries, ok := cachedDirListing(dir, now); ok {
		return entries, nil
	}
	entries, err := readDirEntries(dir)
	if err != nil {
		return nil, err
	}
	storeDirListing(dir, entries, now)
	return entries, nil
}

// rootForPath returns the library root containing p.
func rootForPath(cfg *config.Config, p string) (libraryRoot, bool) {
	for _, root := range libraryRoots(cfg) {
		if p == root.Path || strings.HasPrefix(p, root.Path+string(filepath.Separator)) {
			return root, true
		}
	}
	return libraryRoot{}, false
}

// breadcrumb decomposes p into the clickable steps below root.
func breadcrumb(root libraryRoot, p string) []pathSegment {
	rel := strings.TrimPrefix(p, root.Path)
	rel = strings.Trim(rel, string(filepath.Separator))
	segments := []pathSegment{}
	if rel == "" {
		return segments
	}
	acc := root.Path
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		acc = filepath.Join(acc, part)
		segments = append(segments, pathSegment{Name: part, Path: acc})
	}
	return segments
}

// writeLibraryPathError maps a path-resolution sentinel to an HTTP status.
func writeLibraryPathError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errLibraryPathRequired), errors.Is(err, errLibraryPathRelative):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
	case errors.Is(err, errLibraryPathOutside), errors.Is(err, errLibraryRootReadOnly):
		writeJSON(w, http.StatusForbidden, map[string]any{"error": err.Error()})
	case errors.Is(err, errNoLibraryRoots):
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": err.Error()})
	default:
		writeError(w, http.StatusInternalServerError, "failed to resolve path", err)
	}
}

// handleFilesRoots lists the browsable libraries.
//
// GET /api/files/roots
//
// The kavita flag says whether POST /api/files/kavita/scan is usable at all, so
// the front end can hide the scan button instead of offering an action that can
// only answer 503.
func (s *Server) handleFilesRoots(w http.ResponseWriter, _ *http.Request) {
	roots := libraryRoots(s.cfg)
	if roots == nil {
		roots = []libraryRoot{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"roots":  roots,
		"kavita": s.cfg != nil && s.cfg.HasKavita(),
	})
}

// handleFilesList lists one directory of one library.
//
// GET /api/files/list?path=/books/manga/Berserk
func (s *Server) handleFilesList(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Query().Get("path")
	clean, isRoot, err := resolveLibraryPath(s.cfg, p)
	if err != nil {
		writeLibraryPathError(w, err)
		return
	}
	root, ok := rootForPath(s.cfg, clean)
	if !ok {
		// resolveLibraryPath already confined the path, so this cannot happen
		// unless the roots changed in between; treat it as a refusal.
		writeLibraryPathError(w, errLibraryPathOutside)
		return
	}

	info, err := os.Stat(clean)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "directory not found"})
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to stat directory", err)
		return
	}
	if !info.IsDir() {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "path is not a directory"})
		return
	}

	entries, err := listDir(clean)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "directory not found"})
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to read directory", err)
		return
	}

	parent := ""
	if !isRoot {
		parent = filepath.Dir(clean)
	}
	writeJSON(w, http.StatusOK, listResponse{
		Root:      root.ID,
		RootLabel: root.Label,
		RootPath:  root.Path,
		Path:      clean,
		IsRoot:    isRoot,
		Parent:    parent,
		Segments:  breadcrumb(root, clean),
		Entries:   entries,
	})
}
