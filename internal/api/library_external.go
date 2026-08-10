package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/JeremiahM37/librarr/internal/kavita"
	"github.com/JeremiahM37/librarr/internal/models"
)

// kavitaClient builds a Kavita client from the current config. A fresh client
// per request keeps runtime settings changes honored while still caching the
// JWT across the multiple calls a single request may make.
func (s *Server) kavitaClient() *kavita.Client {
	return kavita.New(s.cfg.KavitaURL, s.cfg.KavitaUser, s.cfg.KavitaPass)
}

// --- Audiobookshelf Library ---

type absLibraryResponse struct {
	Results  []absItem `json:"results"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	NumPages int       `json:"numPages"`
	Limit    int       `json:"limit"`
}

type absItem struct {
	ID    string   `json:"id"`
	Media absMedia `json:"media"`
}

type absMedia struct {
	Metadata   absMetadata   `json:"metadata"`
	Duration   float64       `json:"duration"`
	AudioFiles []interface{} `json:"audioFiles"`
	CoverPath  string        `json:"coverPath"`
}

type absMetadata struct {
	Title      string      `json:"title"`
	AuthorName string      `json:"authorName"`
	SeriesName string      `json:"seriesName"`
	Authors    []absAuthor `json:"authors"`
	Series     []absSeries `json:"series"`
}

type absAuthor struct {
	Name string `json:"name"`
}

type absSeries struct {
	Name string `json:"name"`
}

func (s *Server) handleLibraryAudiobooks(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.HasAudiobookshelf() {
		// Issue #49: when ABS isn't configured, surface audiobooks that
		// were imported into the local DB (e.g. via /api/import/files
		// or the curl-based bulk import flow) rather than returning an
		// empty list. Without this fallback those items only appear
		// under /api/library and get mislabeled as ebooks in the UI.
		s.serveLocalLibraryByMediaType(w, r, "audiobook")
		return
	}

	page := queryInt(r, "page", 1)
	query := r.URL.Query().Get("q")
	limit := 100

	libraryID := s.cfg.ABSLibraryID
	if libraryID == "" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"items": []interface{}{},
			"total": 0,
			"page":  page,
			"pages": 0,
			"error": "ABS_LIBRARY_ID not configured",
		})
		return
	}

	// Build ABS API URL — sort by series then title for grouping.
	absURL := fmt.Sprintf("%s/api/libraries/%s/items", s.cfg.ABSURL, libraryID)
	params := url.Values{
		"page":  {strconv.Itoa(page - 1)}, // ABS uses 0-indexed pages
		"limit": {strconv.Itoa(limit)},
		"sort":  {"media.metadata.seriesName"},
	}
	if query != "" {
		params.Set("filter", "search="+url.QueryEscape(query))
	}

	req, err := http.NewRequest("GET", absURL+"?"+params.Encode(), nil)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.ABSToken)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("ABS library request failed", "error", err)
		writeJSON(w, http.StatusBadGateway, map[string]interface{}{
			"error": "Failed to reach Audiobookshelf",
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		slog.Error("ABS library non-200", "status", resp.StatusCode, "body", string(body))
		writeJSON(w, http.StatusBadGateway, map[string]interface{}{
			"error": fmt.Sprintf("Audiobookshelf returned HTTP %d", resp.StatusCode),
		})
		return
	}

	var absResp absLibraryResponse
	if err := json.NewDecoder(resp.Body).Decode(&absResp); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"error": "Failed to parse ABS response",
		})
		return
	}

	// Build public base URL for covers and links.
	publicURL := s.cfg.ABSPublicURL
	if publicURL == "" {
		publicURL = s.cfg.ABSURL
	}

	var items []map[string]interface{}
	for _, item := range absResp.Results {
		author := item.Media.Metadata.AuthorName
		if author == "" && len(item.Media.Metadata.Authors) > 0 {
			author = item.Media.Metadata.Authors[0].Name
		}
		series := item.Media.Metadata.SeriesName
		if series == "" && len(item.Media.Metadata.Series) > 0 {
			series = item.Media.Metadata.Series[0].Name
		}

		coverURL := ""
		if item.Media.CoverPath != "" || true {
			// ABS always serves cover via /api/items/{id}/cover
			coverURL = fmt.Sprintf("%s/api/items/%s/cover", publicURL, item.ID)
		}

		items = append(items, map[string]interface{}{
			"id":             item.ID,
			"title":          item.Media.Metadata.Title,
			"author":         author,
			"series":         series,
			"duration_hours": math.Round(item.Media.Duration/3600*10) / 10,
			"num_files":      len(item.Media.AudioFiles),
			"cover_url":      coverURL,
			"abs_url":        fmt.Sprintf("%s/item/%s", publicURL, item.ID),
		})
	}

	if items == nil {
		items = []map[string]interface{}{}
	}

	totalPages := absResp.NumPages
	if totalPages == 0 && absResp.Total > 0 {
		totalPages = int(math.Ceil(float64(absResp.Total) / float64(limit)))
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"total": absResp.Total,
		"page":  page,
		"pages": totalPages,
	})
}

// --- Library (Manga) ---

// handleLibraryManga lists the manga rows of the local library, exactly like the
// ebooks and audiobooks tabs do.
//
// It used to mirror Kavita's series list when Kavita was configured. That made
// the tab unable to show — and therefore unable to delete — librarr's own
// library_items rows: the duplicate check in the download path keys on
// library_items.source_id, so an entry left behind after a torrent was removed
// blocked any re-download while being invisible everywhere in the UI. The local
// DB is the only source that can answer "what does librarr think it already
// has?", so that is what this tab shows. Kavita browsing belongs in Kavita.
func (s *Server) handleLibraryManga(w http.ResponseWriter, r *http.Request) {
	s.serveLocalLibraryByMediaType(w, r, "manga")
}

// --- Delete library items ---

func (s *Server) handleDeleteBook(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	// Try as integer (internal DB item) first.
	if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
		if err := s.db.DeleteItem(id); err != nil {
			writeJSON(w, http.StatusNotFound, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
		username, _ := r.Context().Value(ctxUsername).(string)
		s.db.LogActivity(username, "library_remove", idStr, fmt.Sprintf("Removed library item %s", idStr))
		writeJSON(w, http.StatusOK, map[string]interface{}{"success": true})
		return
	}

	// UUID string — ABS library item. Delete via ABS API.
	if s.cfg.HasAudiobookshelf() && s.cfg.ABSToken != "" {
		absURL := fmt.Sprintf("%s/api/items/%s", s.cfg.ABSURL, idStr)
		req, err := http.NewRequest("DELETE", absURL, nil)
		if err == nil {
			req.Header.Set("Authorization", "Bearer "+s.cfg.ABSToken)
			resp, err := http.DefaultClient.Do(req)
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode < 300 {
					slog.Info("deleted ABS library item", "id", idStr)
				} else {
					slog.Warn("ABS delete non-success", "id", idStr, "status", resp.StatusCode)
				}
			}
		}
	}

	// Also try internal DB by source_id.
	_ = s.db.DeleteItemBySourceID(idStr)

	username, _ := r.Context().Value(ctxUsername).(string)
	s.db.LogActivity(username, "library_remove", idStr, fmt.Sprintf("Removed library item %s", idStr))
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

func (s *Server) handleDeleteAudiobook(w http.ResponseWriter, r *http.Request) {
	// Same as delete book — handles both integer and UUID IDs.
	s.handleDeleteBook(w, r)
}

// handleDeleteManga removes a manga row from librarr's own library. Unlike books
// and audiobooks there is no external system to mirror the delete to: manga rows
// only ever come from the local DB, so the id is always an internal one and a
// non-numeric id is a caller bug rather than a foreign identifier.
//
// The file on disk is deliberately left alone. Removing the row is what releases
// the duplicate check in the download path, which is the whole point of exposing
// this; deleting media is a separate, destructive concern.
func (s *Server) handleDeleteManga(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Invalid library item id",
		})
		return
	}
	if err := s.db.DeleteItem(id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	username, _ := r.Context().Value(ctxUsername).(string)
	s.db.LogActivity(username, "library_remove", idStr, fmt.Sprintf("Removed library item %s", idStr))
	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}

// --- Wishlist ---

func (s *Server) handleGetWishlist(w http.ResponseWriter, _ *http.Request) {
	items, err := s.db.GetWishlist()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}
	if items == nil {
		items = []models.WishlistItem{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
	})
}

func (s *Server) handleAddWishlist(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title     string `json:"title"`
		Author    string `json:"author"`
		MediaType string `json:"media_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	if req.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Title is required",
		})
		return
	}
	// Cap user-supplied strings so a misbehaving client can't bloat the DB.
	if len(req.Title) > 500 || len(req.Author) > 500 || len(req.MediaType) > 50 {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Field exceeds maximum length",
		})
		return
	}

	id, err := s.db.AddWishlistItem(req.Title, req.Author, req.MediaType)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"id":      id,
	})
}

func (s *Server) handleDeleteWishlist(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "Invalid ID",
		})
		return
	}

	if err := s.db.DeleteWishlistItem(id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"success": true})
}
