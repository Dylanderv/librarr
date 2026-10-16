package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/JeremiahM37/librarr/internal/convert"
)

// CBR-to-CBZ conversion for the /api/files/* surface.
//
// A CBR is a RAR archive, which librarr cannot rewrite — so the per-file
// metadata panel stops at "rename only" for it. This endpoint automates the
// manual workaround (unpack the RAR, repack it as ZIP, rename to .cbz) with a
// pure-Go RAR reader: no external unrar binary, nothing to install in the
// Docker image.
//
// The conversion creates `Name.cbz` next to the source and only then removes
// the `.cbr`, so a failed conversion never costs the original. Every directory
// touched goes through invalidateDirListing, like the rename and merge
// endpoints.

// convertRequest names the .cbr file to convert.
type convertRequest struct {
	Path string `json:"path"`
}

// convertResponse reports the new file and what the archive held.
type convertResponse struct {
	Name    string   `json:"name"`
	Path    string   `json:"path"`
	Files   int      `json:"files"`
	Skipped []string `json:"skipped,omitempty"`
}

// handleFilesConvert repacks one .cbr as .cbz and removes the source.
//
// POST /api/files/convert
func (s *Server) handleFilesConvert(w http.ResponseWriter, r *http.Request) {
	var req convertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	if strings.TrimSpace(req.Path) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "path is required"})
		return
	}

	src, err := resolveLibraryPathForWrite(s.cfg, req.Path)
	if err != nil {
		writeLibraryPathError(w, err)
		return
	}
	if !strings.EqualFold(filepath.Ext(src), ".cbr") {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "only .cbr files can be converted"})
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
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "path is not a file"})
		return
	}

	res, err := convert.CBRToCBZ(src)
	if err != nil {
		switch {
		case errors.Is(err, convert.ErrExists):
			writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error()})
		case errors.Is(err, convert.ErrEncrypted),
			errors.Is(err, convert.ErrUnsafe),
			errors.Is(err, convert.ErrEmpty):
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		default:
			writeError(w, http.StatusInternalServerError, "failed to convert archive", err)
		}
		return
	}

	// The CBZ is complete and verified on disk; only then does the source go,
	// with a plain remove — it is a file, never a directory.
	dir := filepath.Dir(src)
	if err := os.Remove(src); err != nil {
		writeError(w, http.StatusInternalServerError, "converted but failed to remove the source", err)
		return
	}
	invalidateDirListing(dir)

	slog.Info("cbr converted to cbz", "from", src, "to", res.Path, "files", res.Files)

	writeJSON(w, http.StatusCreated, convertResponse{
		Name:    filepath.Base(res.Path),
		Path:    res.Path,
		Files:   res.Files,
		Skipped: res.Skipped,
	})
}
