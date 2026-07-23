package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/JeremiahM37/librarr/internal/metadata/anilist"
)

// --- Bulk online auto-match ---
//
// The single-series flow lets the user search AniList and eyeball one candidate.
// Bulk auto-match extends that to many series at once: for each selected series
// librarr searches AniList by the series' own name and returns the top match, so
// the UI can present a review table before anything is written. Applying is a
// separate call (handleBulkMatchApply) that takes per-series fields — because
// each series gets its own metadata, unlike handleBulkMetadata which applies one
// field set to every series.

type bulkMatchRequest struct {
	SeriesIDs []int `json:"series_ids"`
}

type bulkMatchResult struct {
	SeriesID   int            `json:"series_id"`
	SeriesName string         `json:"series_name"`
	Matched    *metaCandidate `json:"matched"` // nil when AniList had no hit
}

// handleBulkMatch looks up each selected series on AniList by its name and
// returns the best candidate per series for review. It writes nothing.
//
// POST /api/metadata/bulk-match  body: {"series_ids": [1,2,3]}
func (s *Server) handleBulkMatch(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.HasKavita() {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Kavita is not configured"})
		return
	}
	if s.anilistClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "AniList client not available"})
		return
	}

	var req bulkMatchRequest
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

	ctx := r.Context()
	client := s.kavitaClient()

	results := make([]bulkMatchResult, 0, len(req.SeriesIDs))
	for _, id := range req.SeriesIDs {
		res := bulkMatchResult{SeriesID: id}

		sr, err := client.SeriesForID(ctx, id)
		if err != nil {
			slog.Error("bulk-match: series lookup failed", "series_id", id, "error", err)
			results = append(results, res)
			continue
		}
		res.SeriesName = sr.Name

		matches, err := s.anilistClient.SearchManga(ctx, sr.Name)
		if err != nil {
			slog.Error("bulk-match: anilist search failed", "series_id", id, "query", sr.Name, "error", err)
			results = append(results, res)
			continue
		}
		if len(matches) > 0 {
			res.Matched = anilistToCandidate(matches[0])
		}
		results = append(results, res)
	}

	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

// anilistToCandidate maps an AniList match into the shared candidate shape.
func anilistToCandidate(m anilist.MangaMatch) *metaCandidate {
	return &metaCandidate{
		Source:   "anilist",
		Title:    m.Title,
		Summary:  m.Summary,
		Authors:  m.Authors,
		Genres:   m.Genres,
		Tags:     m.Tags,
		Year:     m.Year,
		CoverURL: m.CoverURL,
		SiteURL:  m.SiteURL,
	}
}

type bulkApplyItem struct {
	SeriesID int            `json:"series_id"`
	Fields   bulkMetaFields `json:"fields"`
}

type bulkApplyRequest struct {
	Mode  string          `json:"mode"` // "set" (replace lists) or "append"
	Items []bulkApplyItem `json:"items"`
}

// handleBulkMatchApply writes a distinct field set into each series' files,
// reusing the same per-series write+rescan path as handleBulkMetadata. Used by
// the bulk auto-match review UI once the user has confirmed the matches.
//
// POST /api/metadata/bulk-apply
func (s *Server) handleBulkMatchApply(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.HasKavita() {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "Kavita is not configured"})
		return
	}

	var req bulkApplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid request body"})
		return
	}
	if len(req.Items) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "items is required"})
		return
	}
	if len(req.Items) > 500 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "too many series in one request (max 500)"})
		return
	}
	if len(s.cfg.KavitaAllowedRoots()) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "no writable library roots configured; set KAVITA_LIBRARY_PATH, KAVITA_MANGA_LIBRARY_PATH or KAVITA_PATH_MAP",
		})
		return
	}

	ctx := r.Context()
	client := s.kavitaClient()

	results := make([]seriesMetaResult, 0, len(req.Items))
	totalWritten := 0
	applied := 0
	for _, item := range req.Items {
		fields := toWriterFields(item.Fields, req.Mode)
		if fields.IsEmpty() {
			continue
		}
		res := s.applyMetadataToSeries(ctx, client, item.SeriesID, fields)
		totalWritten += res.FilesWritten
		results = append(results, res)
		applied++
	}

	username, _ := r.Context().Value(ctxUsername).(string)
	s.db.LogActivity(username, "metadata_bulk_match", "",
		fmt.Sprintf("Bulk auto-match: %d series, %d files written", applied, totalWritten))

	writeJSON(w, http.StatusOK, map[string]any{
		"success":       true,
		"series_count":  applied,
		"files_written": totalWritten,
		"results":       results,
	})
}
