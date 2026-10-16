package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/JeremiahM37/librarr/internal/config"
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

// convertRequest names the .cbr file(s) to convert.
//
// Two shapes: the legacy single {path}, answered with a convertResponse, and
// the batch {sources}, answered with a batchConvertResponse that reports per
// file instead of failing the whole request on one bad archive.
type convertRequest struct {
	Path    string   `json:"path"`
	Sources []string `json:"sources"`
}

// convertResponse reports the new file and what the archive held.
type convertResponse struct {
	Name    string   `json:"name"`
	Path    string   `json:"path"`
	Files   int      `json:"files"`
	Skipped []string `json:"skipped,omitempty"`
}

// batchConvertedFile is one archive that became a .cbz.
type batchConvertedFile struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	Files int    `json:"files"`
}

// batchFailedFile is one source left unconverted, with a reason key the front
// end translates (exists | missing | not_file | not_cbr | encrypted | empty |
// unsafe | failed | remove_failed) and an optional detail.
type batchFailedFile struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
	Detail string `json:"detail,omitempty"`
}

// batchConvertResponse reports a batch conversion.
type batchConvertResponse struct {
	Converted []batchConvertedFile `json:"converted"`
	Failed    []batchFailedFile    `json:"failed,omitempty"`
}

// Reasons a source is left unconverted.
const (
	convertFailExists       = "exists"
	convertFailMissing      = "missing"
	convertFailNotFile      = "not_file"
	convertFailNotSupported = "not_supported"
	convertFailEncrypted    = "encrypted"
	convertFailEmpty        = "empty"
	convertFailUnsafe       = "unsafe"
	convertFailConvert      = "failed"
	convertFailRemove       = "remove_failed"
)

// maxConvertSources bounds one batch request, like maxMoveSources bounds a move.
const maxConvertSources = 100

// handleFilesConvert repacks .cbr files as .cbz and removes the sources.
//
// POST /api/files/convert
func (s *Server) handleFilesConvert(w http.ResponseWriter, r *http.Request) {
	var req convertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	if len(req.Sources) > 0 {
		s.handleFilesConvertBatch(w, req)
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
	kind, ok := convertibleKind(src)
	if !ok {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "only .cbr and .pdf files can be converted"})
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

	res, err := convertByKind(kind, src, pdfOptionsFor(s.cfg))
	if err != nil {
		switch {
		case errors.Is(err, convert.ErrExists):
			writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error()})
		case errors.Is(err, convert.ErrEncrypted),
			errors.Is(err, convert.ErrUnsafe),
			errors.Is(err, convert.ErrEmpty):
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		case errors.Is(err, convert.ErrNoPoppler):
			writeJSON(w, http.StatusNotImplemented, map[string]any{"error": err.Error()})
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

// handleFilesConvertBatch converts every source and reports per file instead
// of failing the whole request on one bad archive. A source that cannot be
// confined aborts everything before anything converts — same fail-fast rule
// as the batch move — while per-file problems (collision, unreadable archive)
// are reported and the rest still converts.
func (s *Server) handleFilesConvertBatch(w http.ResponseWriter, req convertRequest) {
	if len(req.Sources) > maxConvertSources {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "too many files in one conversion"})
		return
	}

	// Confinement first: resolve everything before converting anything.
	srcs := make([]string, 0, len(req.Sources))
	for _, raw := range req.Sources {
		src, err := resolveLibraryPathForWrite(s.cfg, raw)
		if err != nil {
			writeLibraryPathError(w, err)
			return
		}
		srcs = append(srcs, src)
	}

	// A missing renderer is a system problem, not a per-file one: fail the
	// whole request instead of one failed row per PDF.
	for _, src := range srcs {
		if strings.EqualFold(filepath.Ext(src), ".pdf") && !convert.HavePoppler() {
			writeJSON(w, http.StatusNotImplemented, map[string]any{"error": convert.ErrNoPoppler.Error()})
			return
		}
	}

	out := batchConvertResponse{Converted: []batchConvertedFile{}}
	touched := map[string]bool{}
	for _, src := range srcs {
		base := filepath.Base(src)
		fail := func(reason, detail string) {
			out.Failed = append(out.Failed, batchFailedFile{Name: base, Reason: reason, Detail: detail})
		}
		info, err := os.Stat(src)
		if err != nil {
			if os.IsNotExist(err) {
				fail(convertFailMissing, "")
				continue
			}
			writeError(w, http.StatusInternalServerError, "failed to stat file", err)
			return
		}
		if info.IsDir() {
			fail(convertFailNotFile, "")
			continue
		}
		kind, ok := convertibleKind(src)
		if !ok {
			fail(convertFailNotSupported, "")
			continue
		}
		res, err := convertByKind(kind, src, pdfOptionsFor(s.cfg))
		if err != nil {
			switch {
			case errors.Is(err, convert.ErrExists):
				fail(convertFailExists, "")
			case errors.Is(err, convert.ErrEncrypted):
				fail(convertFailEncrypted, "")
			case errors.Is(err, convert.ErrEmpty):
				fail(convertFailEmpty, "")
			case errors.Is(err, convert.ErrUnsafe):
				fail(convertFailUnsafe, "")
			default:
				fail(convertFailConvert, err.Error())
			}
			continue
		}
		dir := filepath.Dir(src)
		if err := os.Remove(src); err != nil {
			fail(convertFailRemove, err.Error())
			continue
		}
		touched[dir] = true
		out.Converted = append(out.Converted, batchConvertedFile{
			Name:  filepath.Base(res.Path),
			Path:  res.Path,
			Files: res.Files,
		})
	}
	for d := range touched {
		invalidateDirListing(d)
	}

	slog.Info("cbr batch converted", "converted", len(out.Converted), "failed", len(out.Failed))

	writeJSON(w, http.StatusOK, out)
}

// convertibleKind reports whether src is convertible and which pipeline it
// needs: "cbr" for the RAR repack, "pdf" for the poppler rendering.
func convertibleKind(src string) (string, bool) {
	switch strings.ToLower(filepath.Ext(src)) {
	case ".cbr":
		return "cbr", true
	case ".pdf":
		return "pdf", true
	default:
		return "", false
	}
}

// convertByKind dispatches one source to its pipeline.
func convertByKind(kind, src string, pdfOpts convert.PDFOptions) (convert.Result, error) {
	if kind == "pdf" {
		return convert.PDFToCBZ(src, pdfOpts)
	}
	return convert.CBRToCBZ(src)
}

// pdfOptionsFor reads the rendering options from the server config, falling
// back to the converter defaults when unset.
func pdfOptionsFor(cfg *config.Config) convert.PDFOptions {
	if cfg == nil {
		return convert.PDFOptions{}
	}
	return convert.PDFOptions{DPI: cfg.PDFDPI, JPEGQuality: cfg.PDFJPEGQuality}
}
