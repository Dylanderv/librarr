package api

import (
	"log/slog"
	"net/http"
)

// metaCandidate is the common shape returned to the bulk-editor "search online"
// UI, regardless of whether it came from AniList (manga) or Open Library
// (books). Empty fields are omitted so the frontend can prefill only what it got.
type metaCandidate struct {
	Source    string   `json:"source"`
	Title     string   `json:"title"`
	Summary   string   `json:"summary,omitempty"`
	Authors   []string `json:"authors,omitempty"`
	Genres    []string `json:"genres,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Year      string   `json:"year,omitempty"`
	Publisher string   `json:"publisher,omitempty"`
	Language  string   `json:"language,omitempty"`
	CoverURL  string   `json:"cover_url,omitempty"`
	SiteURL   string   `json:"site_url,omitempty"`
}

// handleMetadataSearch queries an online metadata source for candidates to
// pre-fill the bulk metadata editor. Manga is routed to AniList; books reuse
// the Open Library client.
//
// GET /api/metadata/search?type=manga|book&q=...
func (s *Server) handleMetadataSearch(w http.ResponseWriter, r *http.Request) {
	q := truncateSearchQuery(r.URL.Query().Get("q"))
	if q == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "q is required"})
		return
	}
	mediaType := r.URL.Query().Get("type")
	if mediaType == "" {
		mediaType = "manga"
	}

	switch mediaType {
	case "manga":
		s.searchAniList(w, r, q)
	case "book":
		s.searchOpenLibrary(w, r, q)
	default:
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "type must be 'manga' or 'book'"})
	}
}

func (s *Server) searchAniList(w http.ResponseWriter, r *http.Request, q string) {
	if s.anilistClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "AniList client not available"})
		return
	}
	matches, err := s.anilistClient.SearchManga(r.Context(), q)
	if err != nil {
		slog.Error("anilist search failed", "query", q, "error", err)
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": "failed to search AniList"})
		return
	}

	candidates := make([]metaCandidate, 0, len(matches))
	for _, m := range matches {
		candidates = append(candidates, metaCandidate{
			Source:   "anilist",
			Title:    m.Title,
			Summary:  m.Summary,
			Authors:  m.Authors,
			Genres:   m.Genres,
			Tags:     m.Tags,
			Year:     m.Year,
			CoverURL: m.CoverURL,
			SiteURL:  m.SiteURL,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"candidates": candidates})
}

func (s *Server) searchOpenLibrary(w http.ResponseWriter, r *http.Request, q string) {
	if s.metadataClient == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "Open Library client not available"})
		return
	}
	meta, err := s.metadataClient.FetchMetadataCtx(r.Context(), q, "")
	if err != nil {
		slog.Error("open library search failed", "query", q, "error", err)
		writeJSON(w, http.StatusBadGateway, map[string]any{"error": "failed to search Open Library"})
		return
	}

	candidates := make([]metaCandidate, 0, 1)
	if meta != nil {
		c := metaCandidate{
			Source:    "openlibrary",
			Title:     meta.Title,
			Summary:   meta.Description,
			Year:      meta.Year,
			Publisher: meta.Publisher,
			Language:  meta.Language,
			CoverURL:  meta.CoverURL,
		}
		if meta.Author != "" {
			c.Authors = []string{meta.Author}
		}
		if meta.OLID != "" {
			c.SiteURL = "https://openlibrary.org" + meta.OLID
		}
		candidates = append(candidates, c)
	}
	writeJSON(w, http.StatusOK, map[string]any{"candidates": candidates})
}
