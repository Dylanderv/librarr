package api

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JeremiahM37/librarr/internal/config"
	"github.com/JeremiahM37/librarr/internal/metadata/anilist"
)

// TestBulkMatch_RequiresKavita: without Kavita configured there's no way to map
// a series to its files, so both bulk endpoints must reject the request rather
// than silently no-op.
func TestBulkMatch_RequiresKavita(t *testing.T) {
	s := &Server{cfg: &config.Config{}} // Kavita unconfigured

	t.Run("bulk-match", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/metadata/bulk-match", strings.NewReader(`{"series_ids":[1]}`))
		rr := httptest.NewRecorder()
		s.handleBulkMatch(rr, req)
		if rr.Code != 400 {
			t.Fatalf("expected 400 without Kavita, got %d: %s", rr.Code, rr.Body.String())
		}
	})

	t.Run("bulk-apply", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/metadata/bulk-apply", strings.NewReader(`{"mode":"set","items":[{"series_id":1,"fields":{}}]}`))
		rr := httptest.NewRecorder()
		s.handleBulkMatchApply(rr, req)
		if rr.Code != 400 {
			t.Fatalf("expected 400 without Kavita, got %d: %s", rr.Code, rr.Body.String())
		}
	})
}

// TestAnilistToCandidate checks the AniList→candidate field mapping the bulk
// review UI depends on.
func TestAnilistToCandidate(t *testing.T) {
	c := anilistToCandidate(anilist.MangaMatch{
		Title:    "Oyasumi Punpun",
		Authors:  []string{"Inio Asano"},
		Genres:   []string{"Drama", "Slice of Life"},
		Tags:     []string{"Coming of Age"},
		Summary:  "A story.",
		Year:     "2007",
		CoverURL: "https://example/cover.jpg",
		SiteURL:  "https://anilist.co/manga/1",
	})
	if c.Source != "anilist" {
		t.Errorf("source = %q, want anilist", c.Source)
	}
	if c.Title != "Oyasumi Punpun" || c.Year != "2007" {
		t.Errorf("unexpected title/year: %q / %q", c.Title, c.Year)
	}
	if len(c.Authors) != 1 || c.Authors[0] != "Inio Asano" {
		t.Errorf("authors = %v", c.Authors)
	}
	if len(c.Genres) != 2 || len(c.Tags) != 1 {
		t.Errorf("genres/tags = %v / %v", c.Genres, c.Tags)
	}
}
