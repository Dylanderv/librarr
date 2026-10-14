package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JeremiahM37/librarr/internal/config"
	"github.com/JeremiahM37/librarr/internal/search"
)

func TestParseProwlarrOptions_Empty(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/search/manga?q=berserk", nil)
	opts := parseProwlarrOptions(req)
	if opts.UseOverride {
		t.Error("expected no override without pq/pcat")
	}
}

func TestParseProwlarrOptions_QueryAndCategories(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/search/manga?q=berserk&pq=berserk+deluxe&pcat=7020%2C7030", nil)
	opts := parseProwlarrOptions(req)
	if !opts.UseOverride {
		t.Fatal("expected override when pq/pcat present")
	}
	if opts.Query != "berserk deluxe" {
		t.Errorf("expected trimmed pq, got %q", opts.Query)
	}
	if len(opts.Categories) != 2 || opts.Categories[0] != "7020" || opts.Categories[1] != "7030" {
		t.Errorf("unexpected categories: %v", opts.Categories)
	}
}

func TestParseProwlarrOptions_SanitizesCategories(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/search/manga?q=x&pq=y&pcat=7020%2Cabc%2C-1%2C%2C9999999999999999999999", nil)
	opts := parseProwlarrOptions(req)
	for _, c := range opts.Categories {
		for _, r := range c {
			if r < '0' || r > '9' {
				t.Errorf("non-digit category leaked through: %q", opts.Categories)
			}
		}
	}
	if len(opts.Categories) != 1 || opts.Categories[0] != "7020" {
		t.Errorf("expected only clean categories, got %v", opts.Categories)
	}
}

func TestHandleProwlarrCategories_FallbackWhenDown(t *testing.T) {
	cfg := &config.Config{ProwlarrURL: "http://127.0.0.1:1", ProwlarrAPIKey: "x"}
	s := &Server{cfg: cfg, prowlarrCaps: search.NewProwlarrCapsCache(cfg, http.DefaultClient, 0)}

	req := httptest.NewRequest("GET", "/api/prowlarr/categories", nil)
	rr := httptest.NewRecorder()
	s.handleProwlarrCategories(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 with fallback, got %d", rr.Code)
	}
	var body struct {
		Live       bool `json:"live"`
		Categories []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"categories"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body.Live {
		t.Error("expected live=false when Prowlarr is down")
	}
	if len(body.Categories) == 0 {
		t.Error("expected fallback categories")
	}
}

func TestHandleSearchManga_PassesProwlarrOverride(t *testing.T) {
	var gotQuery string
	var gotCats []string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		gotQuery = q.Get("query")
		gotCats = q["categories"]
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"title":"Berserk Deluxe","size":0,"seeders":1,"indexer":"i","infoHash":"h","guid":"g"}]`))
	}))
	defer upstream.Close()

	cfg := &config.Config{ProwlarrURL: upstream.URL, ProwlarrAPIKey: "k"}
	prow := search.NewProwlarr(cfg, upstream.Client(), "manga")
	mgr := search.NewManager(cfg, []search.Searcher{prow}, search.NewHealthTracker(5, 60))
	s := &Server{cfg: cfg, searchMgr: mgr}

	req := httptest.NewRequest("GET", "/api/search/manga?q=berserk&pq=berserk+deluxe&pcat=7020", nil)
	rr := httptest.NewRecorder()
	s.handleSearchManga(rr, req.WithContext(context.WithValue(req.Context(), ctxUsername, "tester")))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if gotQuery != "berserk deluxe" {
		t.Errorf("expected override query at Prowlarr, got %q", gotQuery)
	}
	if len(gotCats) != 1 || gotCats[0] != "7020" {
		t.Errorf("expected categories [7020], got %v", gotCats)
	}
	if !strings.Contains(rr.Body.String(), "Berserk Deluxe") {
		t.Errorf("expected result in response: %s", rr.Body.String())
	}
}
