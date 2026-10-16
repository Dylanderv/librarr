package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Single-file moves for the /api/files/* surface.
//
// Creating a folder (POST /api/files/mkdir) left one gesture missing: a loose
// file sitting at a library root could not join the folder it belongs to.
// Renaming stays in place and merging only handles folders, so moving a file
// into another folder is its own small endpoint.
//
// Same invariants as the other writes: the destination is resolved and
// confined server-side, an occupied name is a 409 rather than an overwrite,
// and both listings (source and destination folders) are invalidated.

// moveRequest names the file and the folder it moves to. The file keeps its
// name — this is a move, not a rename.
type moveRequest struct {
	Path string `json:"path"`
	Dest string `json:"dest"`
}

// moveResponse reports where the file ended up.
type moveResponse struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// handleFilesMove moves one file into another folder of the libraries.
//
// POST /api/files/move
func (s *Server) handleFilesMove(w http.ResponseWriter, r *http.Request) {
	var req moveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	if strings.TrimSpace(req.Path) == "" || strings.TrimSpace(req.Dest) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "file and destination are required"})
		return
	}

	src, err := resolveLibraryPathForWrite(s.cfg, req.Path)
	if err != nil {
		writeLibraryPathError(w, err)
		return
	}
	info, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "file not found"})
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to stat file", err)
		return
	}
	if info.IsDir() {
		// Folders relocate through merge, which knows how to carry their
		// contents and re-plan names along the way.
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "only files can be moved"})
		return
	}

	// The destination may be a library root: adding a file to it does not
	// modify the root itself, like creating a folder under it.
	dest, _, err := resolveLibraryPath(s.cfg, req.Dest)
	if err != nil {
		writeLibraryPathError(w, err)
		return
	}
	destInfo, err := os.Stat(dest)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "destination not found"})
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to stat destination", err)
		return
	}
	if !destInfo.IsDir() {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "destination is not a folder"})
		return
	}
	if filepath.Dir(src) == dest {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "the file is already in this folder"})
		return
	}

	target, err := resolveLibraryPathForWrite(s.cfg, filepath.Join(dest, filepath.Base(src)))
	if err != nil {
		writeLibraryPathError(w, err)
		return
	}
	if _, err := os.Stat(target); err == nil {
		writeJSON(w, http.StatusConflict, map[string]any{"error": "a file already has this name in the destination"})
		return
	}

	if err := os.Rename(src, target); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to move file", err)
		return
	}
	invalidateDirListing(filepath.Dir(src))
	invalidateDirListing(dest)

	slog.Info("file moved", "from", src, "to", target)

	writeJSON(w, http.StatusOK, moveResponse{Name: filepath.Base(target), Path: target})
}
