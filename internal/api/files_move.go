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

// moveRequest names the file(s) and the folder they move to. Files keep
// their names — this is a move, not a rename.
//
// Two shapes: the legacy single {path, dest}, answered with a moveResponse,
// and the batch {sources, dest}, answered with a batchMoveResponse. The
// single shape behaves exactly as it always has; the batch shape moves what
// it can and reports the rest per file instead of failing the whole request.
type moveRequest struct {
	Path    string   `json:"path"`
	Sources []string `json:"sources"`
	Dest    string   `json:"dest"`
}

// moveResponse reports where the file ended up.
type moveResponse struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// batchMovedFile is one file that reached the destination.
type batchMovedFile struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// batchSkippedFile is one file left where it was, with a reason key the front
// end translates (exists | missing | not_file | in_place).
type batchSkippedFile struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

// batchMoveResponse reports a batch move: what landed and what did not.
type batchMoveResponse struct {
	Path    string             `json:"path"`
	Moved   []batchMovedFile   `json:"moved"`
	Skipped []batchSkippedFile `json:"skipped,omitempty"`
}

// Reasons a file is skipped rather than moved.
const (
	moveSkipExists  = "exists"
	moveSkipMissing = "missing"
	moveSkipNotFile = "not_file"
	moveSkipInPlace = "in_place"
)

// maxMoveSources bounds one batch request, like mergeMaxSources bounds a merge.
const maxMoveSources = 500

// handleFilesMove moves one file ({path, dest}) or a batch of files
// ({sources, dest}) into another folder of the libraries.
//
// POST /api/files/move
func (s *Server) handleFilesMove(w http.ResponseWriter, r *http.Request) {
	var req moveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	if len(req.Sources) > 0 {
		s.handleFilesMoveBatch(w, req)
		return
	}
	s.handleFilesMoveSingle(w, req)
}

// handleFilesMoveBatch moves every source into the shared destination,
// reporting per file instead of failing the whole request on a collision or
// a file that vanished in between.
func (s *Server) handleFilesMoveBatch(w http.ResponseWriter, req moveRequest) {
	if strings.TrimSpace(req.Dest) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "destination is required"})
		return
	}

	// The destination is shared by every source, so it is resolved once. It
	// may be a library root: adding files to it does not modify the root
	// itself, like creating a folder under it.
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

	if len(req.Sources) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "no files to move"})
		return
	}
	if len(req.Sources) > maxMoveSources {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "too many files in one move"})
		return
	}

	// Every source is confined before anything moves: a traversal attempt
	// aborts the whole request instead of partially applying it.
	srcs := make([]string, 0, len(req.Sources))
	for _, raw := range req.Sources {
		src, err := resolveLibraryPathForWrite(s.cfg, raw)
		if err != nil {
			writeLibraryPathError(w, err)
			return
		}
		srcs = append(srcs, src)
	}

	out := batchMoveResponse{Path: dest}
	touched := map[string]bool{dest: true}
	for _, src := range srcs {
		base := filepath.Base(src)
		info, err := os.Stat(src)
		if err != nil {
			if os.IsNotExist(err) {
				out.Skipped = append(out.Skipped, batchSkippedFile{Name: base, Reason: moveSkipMissing})
				continue
			}
			writeError(w, http.StatusInternalServerError, "failed to stat file", err)
			return
		}
		if info.IsDir() {
			out.Skipped = append(out.Skipped, batchSkippedFile{Name: base, Reason: moveSkipNotFile})
			continue
		}
		if filepath.Dir(src) == dest {
			out.Skipped = append(out.Skipped, batchSkippedFile{Name: base, Reason: moveSkipInPlace})
			continue
		}
		target, err := resolveLibraryPathForWrite(s.cfg, filepath.Join(dest, base))
		if err != nil {
			writeLibraryPathError(w, err)
			return
		}
		if _, err := os.Stat(target); err == nil {
			out.Skipped = append(out.Skipped, batchSkippedFile{Name: base, Reason: moveSkipExists})
			continue
		}
		if err := os.Rename(src, target); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to move file", err)
			return
		}
		touched[filepath.Dir(src)] = true
		out.Moved = append(out.Moved, batchMovedFile{Name: base, Path: target})
	}
	for d := range touched {
		invalidateDirListing(d)
	}

	slog.Info("files moved", "dest", dest, "moved", len(out.Moved), "skipped", len(out.Skipped))

	if out.Moved == nil {
		out.Moved = []batchMovedFile{}
	}
	writeJSON(w, http.StatusOK, out)
}

// handleFilesMoveSingle moves one file, the legacy {path, dest} shape. Kept
// byte-for-byte compatible: same statuses, same response.
func (s *Server) handleFilesMoveSingle(w http.ResponseWriter, req moveRequest) {
	if strings.TrimSpace(req.Path) == "" {
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
