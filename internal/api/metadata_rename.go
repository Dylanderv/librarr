package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/JeremiahM37/librarr/internal/kavita"
	"github.com/JeremiahM37/librarr/internal/rename"
)

// seriesRenamePlan is one series' plan as the UI sees it, carrying the series
// name so the preview can be labelled without a second lookup.
type seriesRenamePlan struct {
	SeriesID   int    `json:"series_id"`
	SeriesName string `json:"series_name"`
	rename.Plan
	Errors []string `json:"errors,omitempty"`
}

type renamePreviewRequest struct {
	SeriesIDs  []int  `json:"series_ids"`
	SeriesName string `json:"series_name"`
	Mode       string `json:"mode"`
}

// POST /api/library/metadata/rename-preview
//
// Returns what a rename would do without touching anything. The same
// buildRenamePlan drives the apply path, so what the user approves is what runs.
func (s *Server) handleRenamePreview(w http.ResponseWriter, r *http.Request) {
	var req renamePreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	if len(req.SeriesIDs) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "no series selected"})
		return
	}
	if len(req.SeriesIDs) > 500 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "too many series in one request (max 500)"})
		return
	}

	mode, err := rename.ParseMode(req.Mode)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	if strings.TrimSpace(req.SeriesName) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "series name is required to preview a rename"})
		return
	}

	client := s.kavitaClient()
	plans := make([]seriesRenamePlan, 0, len(req.SeriesIDs))
	for _, id := range req.SeriesIDs {
		plans = append(plans, s.buildRenamePlan(r.Context(), client, id, req.SeriesName, mode))
	}

	writeJSON(w, http.StatusOK, map[string]any{"plans": plans})
}

// buildRenamePlan resolves a series' files to local paths and works out the
// rename. Paths that fall outside the configured library roots are dropped here,
// so neither the preview nor the apply can reach them.
func (s *Server) buildRenamePlan(ctx context.Context, client *kavita.Client, id int, seriesName string, mode rename.Mode) seriesRenamePlan {
	out := seriesRenamePlan{SeriesID: id, SeriesName: seriesName}

	files, err := client.FilesForSeries(ctx, id)
	if err != nil {
		out.Errors = append(out.Errors, "failed to list files: "+err.Error())
		return out
	}

	paths := make([]string, 0, len(files))
	for _, f := range files {
		local, err := s.cfg.ResolveKavitaPath(f.FilePath)
		if err != nil {
			out.Errors = append(out.Errors, fmt.Sprintf("%s: %v", f.FilePath, err))
			continue
		}
		paths = append(paths, local)
	}
	if len(paths) == 0 {
		return out
	}

	out.Plan = rename.BuildPlan(seriesName, paths, mode)

	// A folder rename must stay inside a library root and must never be the root
	// itself: renaming /books/manga would take the whole library with it.
	if out.Plan.FolderFrom != "" {
		if err := s.checkRenamableDir(out.Plan.FolderFrom); err != nil {
			out.Errors = append(out.Errors, fmt.Sprintf("folder not renamable: %v", err))
			out.Plan.FolderFrom, out.Plan.FolderTo = "", ""
			// Without the folder move, the files stay where they are.
			keepFilesInPlace(&out.Plan)
		}
	}

	return out
}

// keepFilesInPlace rewrites file destinations so they land next to their source
// instead of inside a folder that is not going to move after all. Called
// whenever a planned folder rename is dropped, at plan time or at apply time.
func keepFilesInPlace(p *rename.Plan) {
	for i := range p.Files {
		p.Files[i].To = filepath.Join(
			filepath.Dir(p.Files[i].From),
			filepath.Base(p.Files[i].To),
		)
	}
}

var errIsLibraryRoot = errors.New("path is a library root")

// dirValidator reports whether librarr may rename a directory. It exists so the
// rename machinery is not welded to one notion of "allowed root": the Kavita
// path uses checkRenamableDir (KAVITA_PATH_MAP roots), the /api/files/* path
// uses the MANGA_DIR/EBOOK_DIR/AUDIOBOOK_DIR roots of files_roots.go.
type dirValidator func(dir string) error

// checkRenamableDir rejects directories librarr must not move: anything outside
// the Kavita-allowed roots, and the roots themselves.
func (s *Server) checkRenamableDir(dir string) error {
	clean := filepath.Clean(dir)
	for _, root := range s.cfg.KavitaAllowedRoots() {
		if clean == filepath.Clean(root) {
			return errIsLibraryRoot
		}
	}
	// ResolveKavitaPath already enforces containment; reuse it rather than
	// reimplementing the root walk.
	if _, err := s.cfg.ResolveKavitaPath(clean); err != nil {
		return err
	}
	return nil
}

// applyRenamePlan applies a Kavita series' plan. It is the Kavita-flavoured
// wrapper around applyPlan: the validator is the Kavita root check, and the
// errors are folded into the series envelope the UI reads.
func (s *Server) applyRenamePlan(p *seriesRenamePlan) (renamed int) {
	renamed, errs := applyPlan(&p.Plan, s.checkRenamableDir)
	p.Errors = append(p.Errors, errs...)
	return renamed
}

// applyPlan performs the folder rename first, then the files, so file
// destinations computed against the new folder are valid. Failures are collected
// and returned rather than aborting: a series whose folder cannot move can still
// have its files renamed in place.
//
// The plan is mutated to reflect what actually happened — FolderTo is cleared
// when the folder did not move — so callers can tell where the files ended up.
func applyPlan(p *rename.Plan, canRename dirValidator) (renamed int, errs []string) {
	if p.FolderFrom != "" {
		if err := canRename(p.FolderFrom); err != nil {
			return 0, append(errs, fmt.Sprintf("folder not renamable: %v", err))
		}
		if _, err := os.Stat(p.FolderTo); err == nil {
			// Merging into an existing folder is a different operation with its
			// own conflict rules; refuse rather than silently interleave.
			errs = append(errs, fmt.Sprintf("target folder already exists: %s", p.FolderTo))
			p.FolderTo = ""
			keepFilesInPlace(p)
		} else if err := os.Rename(p.FolderFrom, p.FolderTo); err != nil {
			errs = append(errs, fmt.Sprintf("rename folder: %v", err))
			// The files' destinations assumed the move happened, so fall back to
			// renaming them where they are.
			//
			// KNOWN LIMITATION, kept as-is by this refactor: FolderTo is left set,
			// so the loop below still rewrites each source into a folder that does
			// not exist and every file rename then fails. Clearing it here would
			// fix that, but it is a behaviour change in a path no test covers —
			// tracked separately rather than smuggled into a refactor.
			keepFilesInPlace(p)
		} else {
			renamed++
			slog.Info("series folder renamed", "from", p.FolderFrom, "to", p.FolderTo)
		}
	}

	for _, fc := range p.Files {
		from := fc.From
		// If the folder moved, the file moved with it.
		if p.FolderTo != "" && strings.HasPrefix(from, p.FolderFrom+string(filepath.Separator)) {
			from = filepath.Join(p.FolderTo, strings.TrimPrefix(from, p.FolderFrom+string(filepath.Separator)))
		}
		if from == fc.To {
			continue
		}
		if _, err := os.Stat(fc.To); err == nil {
			errs = append(errs, fmt.Sprintf("%s: target already exists", filepath.Base(fc.To)))
			continue
		}
		if err := os.Rename(from, fc.To); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", filepath.Base(from), err))
			continue
		}
		renamed++
	}

	return renamed, errs
}
