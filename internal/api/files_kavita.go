package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/JeremiahM37/librarr/internal/config"
	"github.com/JeremiahM37/librarr/internal/kavita"
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
// Root is a library id from GET /api/files/roots (manga/ebooks/audiobooks).
// The scan is narrowed to the Kavita libraries that live under that root — see
// matchKavitaLibraries.
type kavitaScanRequest struct {
	Root string `json:"root"`
}

// matchKavitaLibraries returns the Kavita libraries that back a local Browse
// root, and the local paths it derived from Kavita's folders (returned so the
// caller can put them in an error message when nothing matched).
//
// The pairing is done on paths and *only* on paths. Kavita reports each
// library's folders as it sees them inside its own container;
// Config.ResolveKavitaPath applies KAVITA_PATH_MAP to turn those into paths
// librarr sees. A library matches when one of its resolved folders overlaps the
// root — either direction counts, because a Kavita library may sit on a
// subfolder of MANGA_DIR (/books/manga/shonen) just as well as it may contain
// it (/books).
//
// Matching by library *name* is deliberately not attempted, not even as a
// fallback: "Manga" pairing with MANGA_DIR is a guess, and the failure mode of
// a wrong guess is silently rescanning somebody else's library. Returning
// nothing and saying so is strictly better.
//
// Several libraries can legitimately match one root (two Kavita libraries on
// two subfolders of MANGA_DIR); all of them are returned and all get scanned.
func matchKavitaLibraries(cfg *config.Config, libs []kavita.Library, rootPath string) (matched []kavita.Library, seen []string) {
	for _, lib := range libs {
		hit := false
		for _, folder := range lib.Folders {
			folder = strings.TrimSpace(folder)
			if folder == "" {
				continue
			}
			local, err := cfg.ResolveKavitaPath(folder)
			if err != nil {
				// No KAVITA_PATH_MAP / no allowed roots configured, or the
				// folder maps outside them. Fall back to the raw path: when
				// librarr and Kavita share a filesystem view (same host, same
				// mounts) no mapping is needed and the folder is already local.
				local = folder
			}
			seen = append(seen, local)
			if pathsOverlap(local, rootPath) {
				hit = true
			}
		}
		if hit {
			matched = append(matched, lib)
		}
	}
	return matched, seen
}

// handleFilesKavitaScan asks Kavita to rescan the library backing a Browse root.
//
// POST /api/files/kavita/scan  {"root": "manga"}
//
// SCOPE: the scan is narrowed to the Kavita libraries whose folders resolve
// under the requested root. librarr lists Kavita's libraries, translates their
// folders back to local paths through KAVITA_PATH_MAP, and scans the ones that
// overlap the root — see matchKavitaLibraries. When nothing matches, the
// request fails with the paths it compared rather than scanning something
// arbitrary.
//
// force is left false: a normal scan already picks up the renames and metadata
// writes Browse makes, and a forced scan re-reads every file in the library,
// which on an NFS mount is minutes of I/O for no added benefit here.
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
	var target libraryRoot
	for _, root := range roots {
		if root.ID == id {
			target = root
			break
		}
	}
	if target.ID == "" {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "unknown library root"})
		return
	}

	client := s.kavitaClient()
	libs, err := client.Libraries(r.Context())
	if err != nil {
		writeErrorDetail(w, http.StatusBadGateway, "kavita scan failed",
			kavita.PublicErrorDetail(err), err)
		return
	}

	matched, seen := matchKavitaLibraries(s.cfg, libs, target.Path)
	if len(matched) == 0 {
		// A configuration problem, not a Kavita failure: say exactly what was
		// compared so the user can fix KAVITA_PATH_MAP (or the library itself)
		// instead of guessing.
		detail := fmt.Sprintf("no Kavita library matches the local root %s", target.Path)
		switch {
		case len(libs) == 0:
			detail += "; Kavita reported no libraries at all"
		case len(seen) == 0:
			detail += "; Kavita reported no folders for its libraries, which usually means librarr's Kavita account is not an administrator"
		default:
			detail += "; Kavita libraries resolve to: " + strings.Join(seen, ", ") +
				" (check KAVITA_PATH_MAP)"
		}
		slog.Warn("kavita scan: no library matches the browse root",
			"root", id, "root_path", target.Path, "kavita_paths", strings.Join(seen, ","))
		writeErrorDetail(w, http.StatusConflict, "kavita scan failed", detail, nil)
		return
	}

	scanned := make([]map[string]any, 0, len(matched))
	for _, lib := range matched {
		if err := client.ScanLibrary(r.Context(), lib.ID, false); err != nil {
			writeErrorDetail(w, http.StatusBadGateway, "kavita scan failed",
				kavita.PublicErrorDetail(err), err)
			return
		}
		scanned = append(scanned, map[string]any{"id": lib.ID, "name": lib.Name})
	}

	slog.Info("kavita scan requested from browse", "root", id, "libraries", len(scanned))
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "requested",
		"root":      id,
		"libraries": scanned,
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
