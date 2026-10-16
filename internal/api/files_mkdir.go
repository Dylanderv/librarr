package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/JeremiahM37/librarr/internal/rename"
)

// Folder creation for the /api/files/* surface.
//
// The Browse tab could rename and merge but not create: every new series folder
// had to be made outside librarr (file manager, mkdir over SSH) before it
// showed up. This endpoint closes that gap with the narrowest possible write —
// one empty directory, no recursion beyond the target itself.
//
// Invariants, shared with files_rename.go / files_merge.go:
//
//   - The name is a single segment, never a path: rename.SanitizeSegment strips
//     separators, and the result is joined to the parent rather than trusted.
//   - The parent may be a library root (creating /manga/NewSeries is the main
//     use case); only the resolved target goes through resolveLibraryPathForWrite,
//     so a bare root can never be "created" over itself.
//   - The parent listing is invalidated, otherwise the browser keeps showing the
//     pre-create state for up to dirListingTTL.

// mkdirRequest names the parent directory and the new folder inside it.
type mkdirRequest struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

// mkdirResponse reports where the folder actually lives.
type mkdirResponse struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// handleFilesMkdir creates one empty folder inside a library.
//
// POST /api/files/mkdir
func (s *Server) handleFilesMkdir(w http.ResponseWriter, r *http.Request) {
	var req mkdirRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}

	trimmed := strings.TrimSpace(req.Name)
	if trimmed == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "folder name is required"})
		return
	}
	safe := rename.SanitizeSegment(trimmed)
	// SanitizeSegment falls back to "Unknown" when nothing usable remains
	// (e.g. "...", "///", control chars). Creating that would be a surprise,
	// so refuse unless the user literally asked for it.
	if safe == "Unknown" && trimmed != "Unknown" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "folder name is invalid"})
		return
	}

	// The parent may be a bare library root: creating a child there is the
	// point, and it does not modify the root itself. Reads use the permissive
	// resolver on purpose; the write check below covers the target.
	parent, _, err := resolveLibraryPath(s.cfg, req.Path)
	if err != nil {
		writeLibraryPathError(w, err)
		return
	}
	info, err := os.Stat(parent)
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

	target, err := resolveLibraryPathForWrite(s.cfg, filepath.Join(parent, safe))
	if err != nil {
		writeLibraryPathError(w, err)
		return
	}
	if _, err := os.Stat(target); err == nil {
		writeJSON(w, http.StatusConflict, map[string]any{"error": "folder already exists"})
		return
	}

	// 0o777 minus the umask, like the merge target: the share's inherited ACL
	// is what decides, not a hard-coded mode.
	if err := os.MkdirAll(target, 0o777); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create folder", err)
		return
	}
	invalidateDirListing(parent)

	slog.Info("folder created", "path", target)

	writeJSON(w, http.StatusCreated, mkdirResponse{Name: safe, Path: target})
}
