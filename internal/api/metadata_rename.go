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
			for i := range out.Plan.Files {
				out.Plan.Files[i].To = filepath.Join(
					filepath.Dir(out.Plan.Files[i].From),
					filepath.Base(out.Plan.Files[i].To),
				)
			}
		}
	}

	return out
}

var errIsLibraryRoot = errors.New("path is a library root")

// checkRenamableDir rejects directories librarr must not move: anything outside
// the configured roots, and the roots themselves.
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

// applyRenamePlan performs the folder rename first, then the files, so file
// destinations computed against the new folder are valid. Failures are collected
// rather than aborting: a series whose folder cannot move can still have its
// files renamed in place.
func (s *Server) applyRenamePlan(p *seriesRenamePlan) (renamed int) {
	if p.Plan.FolderFrom != "" {
		if err := s.checkRenamableDir(p.Plan.FolderFrom); err != nil {
			p.Errors = append(p.Errors, fmt.Sprintf("folder not renamable: %v", err))
			return 0
		}
		if _, err := os.Stat(p.Plan.FolderTo); err == nil {
			// Merging into an existing folder is a different operation with its
			// own conflict rules; refuse rather than silently interleave.
			p.Errors = append(p.Errors,
				fmt.Sprintf("target folder already exists: %s", p.Plan.FolderTo))
			p.Plan.FolderTo = ""
			for i := range p.Plan.Files {
				p.Plan.Files[i].To = filepath.Join(
					filepath.Dir(p.Plan.Files[i].From),
					filepath.Base(p.Plan.Files[i].To),
				)
			}
		} else if err := os.Rename(p.Plan.FolderFrom, p.Plan.FolderTo); err != nil {
			p.Errors = append(p.Errors, fmt.Sprintf("rename folder: %v", err))
			// The files' destinations assumed the move happened, so fall back to
			// renaming them where they are.
			for i := range p.Plan.Files {
				p.Plan.Files[i].To = filepath.Join(
					filepath.Dir(p.Plan.Files[i].From),
					filepath.Base(p.Plan.Files[i].To),
				)
			}
		} else {
			renamed++
			slog.Info("series folder renamed", "from", p.Plan.FolderFrom, "to", p.Plan.FolderTo)
		}
	}

	for _, fc := range p.Plan.Files {
		from := fc.From
		// If the folder moved, the file moved with it.
		if p.Plan.FolderTo != "" && strings.HasPrefix(from, p.Plan.FolderFrom+string(filepath.Separator)) {
			from = filepath.Join(p.Plan.FolderTo, strings.TrimPrefix(from, p.Plan.FolderFrom+string(filepath.Separator)))
		}
		if from == fc.To {
			continue
		}
		if _, err := os.Stat(fc.To); err == nil {
			p.Errors = append(p.Errors, fmt.Sprintf("%s: target already exists", filepath.Base(fc.To)))
			continue
		}
		if err := os.Rename(from, fc.To); err != nil {
			p.Errors = append(p.Errors, fmt.Sprintf("%s: %v", filepath.Base(from), err))
			continue
		}
		renamed++
	}

	return renamed
}
