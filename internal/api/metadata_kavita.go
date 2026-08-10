package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/JeremiahM37/librarr/internal/kavita"
	"github.com/JeremiahM37/librarr/internal/metadata/filewriter"
	"github.com/JeremiahM37/librarr/internal/rename"
)

// handleSeriesMetadata returns a series' current metadata (read from Kavita) to
// pre-fill the bulk-edit form.
//
// GET /api/library/manga/series/{id}/metadata
func (s *Server) handleSeriesMetadata(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.HasKavita() {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Kavita is not configured"})
		return
	}
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid series id"})
		return
	}

	md, err := s.kavitaClient().SeriesMetadata(r.Context(), id)
	if err != nil {
		slog.Error("fetch series metadata failed", "series_id", id, "error", err)
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": "failed to read metadata from Kavita"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"series_id":    md.SeriesID,
		"summary":      md.Summary,
		"language":     md.Language,
		"release_year": md.ReleaseYear,
		"authors":      personNames(md.Writers),
		"publishers":   personNames(md.Publishers),
		"genres":       tagTitles(md.Genres),
		"tags":         tagTitles(md.Tags),
	})
}

// bulkMetaFields uses pointers so the JSON decoder can distinguish "field
// omitted → leave unchanged" from "field present → set (possibly to empty)".
type bulkMetaFields struct {
	Title       *string   `json:"title"`
	Series      *string   `json:"series"`
	SeriesIndex *string   `json:"series_index"`
	Summary     *string   `json:"summary"`
	Publisher   *string   `json:"publisher"`
	Language    *string   `json:"language"`
	ISBN        *string   `json:"isbn"`
	Authors     *[]string `json:"authors"`
	Genres      *[]string `json:"genres"`
	Tags        *[]string `json:"tags"`
}

type bulkMetaRequest struct {
	SeriesIDs []int          `json:"series_ids"`
	Mode      string         `json:"mode"` // "set" (replace lists) or "append"
	Fields    bulkMetaFields `json:"fields"`
	// Rename optionally realigns the folder and file names with the series name
	// being written: "none" (default), "folder" or "folder_and_files".
	Rename string `json:"rename"`
}

type seriesMetaResult struct {
	SeriesID     int      `json:"series_id"`
	FilesTotal   int      `json:"files_total"`
	FilesWritten int      `json:"files_written"`
	Rescanned    bool     `json:"rescanned"`
	PathsRenamed int      `json:"paths_renamed,omitempty"`
	Errors       []string `json:"errors,omitempty"`
}

// handleBulkMetadata writes metadata into the files backing each selected
// series, then triggers a targeted Kavita rescan so the changes surface.
//
// POST /api/library/metadata/bulk
func (s *Server) handleBulkMetadata(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.HasKavita() {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Kavita is not configured"})
		return
	}

	var req bulkMetaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid request body"})
		return
	}
	if len(req.SeriesIDs) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "series_ids is required"})
		return
	}
	if len(req.SeriesIDs) > 500 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "too many series in one request (max 500)"})
		return
	}

	fields := toWriterFields(req.Fields, req.Mode)
	if fields.IsEmpty() {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "no fields to update"})
		return
	}
	if len(s.cfg.KavitaAllowedRoots()) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "no writable library roots configured; set KAVITA_LIBRARY_PATH, KAVITA_MANGA_LIBRARY_PATH or KAVITA_PATH_MAP",
		})
		return
	}

	renameMode, err := rename.ParseMode(req.Rename)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	// Renaming needs a name to rename to, and the series field is the only one
	// that carries it.
	if renameMode != rename.ModeNone && (fields.Series == nil || strings.TrimSpace(*fields.Series) == "") {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "renaming requires the series field to be set",
		})
		return
	}

	ctx := r.Context()
	client := s.kavitaClient()

	results := make([]seriesMetaResult, 0, len(req.SeriesIDs))
	totalWritten, totalRenamed := 0, 0
	for _, id := range req.SeriesIDs {
		res := s.applyMetadataToSeries(ctx, client, id, fields, renameMode)
		totalWritten += res.FilesWritten
		totalRenamed += res.PathsRenamed
		results = append(results, res)
	}

	username, _ := r.Context().Value(ctxUsername).(string)
	s.db.LogActivity(username, "metadata_bulk_edit", "",
		fmt.Sprintf("Bulk metadata edit: %d series, %d files written, %d paths renamed",
			len(req.SeriesIDs), totalWritten, totalRenamed))

	writeJSON(w, http.StatusOK, map[string]any{
		"success":       true,
		"series_count":  len(req.SeriesIDs),
		"files_written": totalWritten,
		"paths_renamed": totalRenamed,
		"results":       results,
	})
}

// applyMetadataToSeries writes the fields into every file of a single series,
// optionally realigns its folder and file names, and rescans it. It never
// returns an error; per-series/per-file problems are captured in the result so
// one bad series doesn't abort the batch.
//
// Order matters: metadata is written against the current paths, the rename comes
// next, and the rescan runs last so Kavita only ever sees the final layout.
func (s *Server) applyMetadataToSeries(ctx context.Context, client *kavita.Client, id int, fields filewriter.Fields, renameMode rename.Mode) seriesMetaResult {
	res := seriesMetaResult{SeriesID: id}

	files, err := client.FilesForSeries(ctx, id)
	if err != nil {
		res.Errors = append(res.Errors, "failed to list files: "+err.Error())
		return res
	}
	res.FilesTotal = len(files)

	for _, f := range files {
		localPath, err := s.cfg.ResolveKavitaPath(f.FilePath)
		if err != nil {
			res.Errors = append(res.Errors, fmt.Sprintf("%s: %v", f.FilePath, err))
			continue
		}
		if err := filewriter.Apply(localPath, fields); err != nil {
			var unsup filewriter.ErrUnsupported
			if errors.As(err, &unsup) {
				res.Errors = append(res.Errors, fmt.Sprintf("%s: %v", baseName(f.FilePath), err))
			} else {
				res.Errors = append(res.Errors, fmt.Sprintf("%s: write failed: %v", baseName(f.FilePath), err))
			}
			continue
		}
		res.FilesWritten++
	}

	if renameMode != rename.ModeNone && res.FilesWritten > 0 {
		plan := s.buildRenamePlan(ctx, client, id, *fields.Series, renameMode)
		res.Errors = append(res.Errors, plan.Errors...)
		for _, skipped := range plan.Plan.Skipped {
			res.Errors = append(res.Errors, "not renamed: "+skipped)
		}
		if !plan.Plan.IsEmpty() {
			plan.Errors = nil
			res.PathsRenamed = s.applyRenamePlan(&plan)
			res.Errors = append(res.Errors, plan.Errors...)
		}
	}

	// Only rescan if we actually changed something.
	if res.FilesWritten > 0 {
		libraryID := 0
		if sr, err := client.SeriesForID(ctx, id); err == nil {
			libraryID = sr.LibraryID
		}
		if err := client.ScanSeries(ctx, libraryID, id); err != nil {
			res.Errors = append(res.Errors, "rescan failed: "+err.Error())
		} else {
			res.Rescanned = true
		}
	}
	return res
}

// toWriterFields converts the request payload into a filewriter.Fields,
// treating any present list field as an instruction to apply it.
func toWriterFields(f bulkMetaFields, mode string) filewriter.Fields {
	out := filewriter.Fields{
		Title:       f.Title,
		Series:      f.Series,
		SeriesIndex: f.SeriesIndex,
		Summary:     f.Summary,
		Publisher:   f.Publisher,
		Language:    f.Language,
		ISBN:        f.ISBN,
		AppendLists: strings.EqualFold(mode, "append"),
	}
	if f.Authors != nil {
		out.Authors = nonNil(*f.Authors)
	}
	if f.Genres != nil {
		out.Genres = nonNil(*f.Genres)
	}
	if f.Tags != nil {
		out.Tags = nonNil(*f.Tags)
	}
	return out
}

// nonNil guarantees a non-nil slice so filewriter treats a present-but-empty
// list as "clear these values" rather than "leave unchanged".
func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func personNames(people []kavita.Person) []string {
	out := make([]string, 0, len(people))
	for _, p := range people {
		if p.Name != "" {
			out = append(out, p.Name)
		}
	}
	return out
}

func tagTitles(tags []kavita.Tag) []string {
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		if t.Title != "" {
			out = append(out, t.Title)
		}
	}
	return out
}

func baseName(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}
