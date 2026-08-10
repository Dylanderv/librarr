package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/JeremiahM37/librarr/internal/config"
)

// Kavita bridge for the /api/files/* surface.
//
// Browse reads the filesystem directly, so it never needs Kavita to display
// anything. The one thing Kavita is needed for is the other direction: telling
// it to re-read files librarr just renamed or re-tagged.
//
// This is exposed as an explicit button and nothing else. No apply path in
// files_rename.go or files_metadata.go calls it: a scan is expensive on a NFS
// library, and firing one after every single edit would make a session of ten
// small fixes queue ten full scans. The user decides when the work is done.

// kavitaScanRequest is the payload of POST /api/files/kavita/scan.
//
// Root is a library id from GET /api/files/roots (manga/ebooks/audiobooks). It
// is validated but it does *not* narrow the scan — see handleFilesKavitaScan.
type kavitaScanRequest struct {
	Root string `json:"root"`
}

// handleFilesKavitaScan asks Kavita to rescan its libraries.
//
// POST /api/files/kavita/scan  {"root": "manga"}
//
// SCOPE: the scan is server-wide, not limited to the library named by Root.
// internal/kavita only exposes ScanLibrary (POST /api/Library/scan with no
// argument, the same call the import pipeline makes) and ScanSeries (which
// needs a Kavita series id — Browse has a directory, not a series). Mapping a
// local root such as MANGA_DIR to a Kavita library id would mean listing
// Kavita's libraries and reverse-applying KAVITA_PATH_MAP to their folders;
// neither exists here, and guessing the pairing by name would silently rescan
// the wrong library. So Root is used to check the request refers to a real
// library, and the response states the scope it actually got.
func (s *Server) handleFilesKavitaScan(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.HasKavita() {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"error": "kavita is not configured",
		})
		return
	}

	var req kavitaScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	id := strings.TrimSpace(req.Root)
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "root is required"})
		return
	}
	roots := libraryRoots(s.cfg)
	if len(roots) == 0 {
		writeLibraryPathError(w, errNoLibraryRoots)
		return
	}
	known := false
	for _, root := range roots {
		if root.ID == id {
			known = true
			break
		}
	}
	if !known {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "unknown library root"})
		return
	}

	if err := s.kavitaClient().ScanLibrary(r.Context()); err != nil {
		writeError(w, http.StatusBadGateway, "kavita scan failed", err)
		return
	}
	slog.Info("kavita scan requested from browse", "root", id)
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "requested",
		"root":   id,
		"scope":  "all",
	})
}

// WarnLibraryRootDivergence logs a startup warning when the directories Browse
// walks (MANGA_DIR / EBOOK_DIR / AUDIOBOOK_DIR) and the roots the Kavita-backed
// Library page writes under (KAVITA_PATH_MAP targets, KAVITA_LIBRARY_PATH…)
// have nothing in common.
//
// The two tabs would then show two disjoint sets of books with no visible
// reason, and an edit made in one would never appear in the other. That is a
// misconfiguration, but only a probable one — hence a warning, and only when
// Kavita is actually configured: an instance without Kavita has no second view
// to diverge from and must stay quiet.
func WarnLibraryRootDivergence(cfg *config.Config) {
	browsePaths, kavitaRoots, diverge := libraryRootsDiverge(cfg)
	if !diverge {
		return
	}
	slog.Warn("browse and kavita library roots do not overlap: the Browse and Library tabs will show different books",
		"browse_roots", strings.Join(browsePaths, ","),
		"kavita_roots", strings.Join(kavitaRoots, ","))
}

// libraryRootsDiverge is the decision behind WarnLibraryRootDivergence, kept
// separate so it can be tested without capturing log output. It returns the two
// sets it compared and whether they are disjoint.
func libraryRootsDiverge(cfg *config.Config) (browsePaths, kavitaRoots []string, diverge bool) {
	if cfg == nil || !cfg.HasKavita() {
		return nil, nil, false
	}
	kavitaRoots = cfg.KavitaAllowedRoots()
	browseRoots := libraryRoots(cfg)
	if len(kavitaRoots) == 0 || len(browseRoots) == 0 {
		return nil, kavitaRoots, false
	}
	for _, b := range browseRoots {
		browsePaths = append(browsePaths, b.Path)
	}
	// Compare on the resolved path *and* on the configured one: libraryRoots
	// returns roots with symlinks resolved, while KavitaAllowedRoots are raw
	// strings, so a symlinked mount would otherwise look like a divergence.
	candidates := append([]string{cfg.MangaDir, cfg.EbookDir, cfg.AudiobookDir}, browsePaths...)
	for _, c := range candidates {
		for _, k := range kavitaRoots {
			if pathsOverlap(c, k) {
				return browsePaths, kavitaRoots, false
			}
		}
	}
	return browsePaths, kavitaRoots, true
}

// pathsOverlap reports whether one path is the other or contains it.
func pathsOverlap(a, b string) bool {
	a = strings.TrimRight(a, "/")
	b = strings.TrimRight(b, "/")
	if a == "" || b == "" {
		return false
	}
	return a == b || strings.HasPrefix(a, b+"/") || strings.HasPrefix(b, a+"/")
}
