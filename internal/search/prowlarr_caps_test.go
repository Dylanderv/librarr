package search

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JeremiahM37/librarr/internal/config"
)

const capsTreeFixture = `[
  {"id": 7000, "name": "Books", "description": "Books", "subCategories": [
    {"id": 7020, "name": "EBook"},
    {"id": 7030, "name": "Comics"},
    {"id": 7040, "name": "Magazines"}
  ]},
  {"id": 3000, "name": "Audio", "subCategories": [
    {"id": 3030, "name": "Audiobook"}
  ]}
]`

func TestFetchProwlarrCategories_ParsesTree(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/indexer/categories" {
			t.Errorf("expected categories path, got %s", r.URL.Path)
		}
		if r.Header.Get("X-Api-Key") != "test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(capsTreeFixture))
	}))
	defer server.Close()

	cfg := &config.Config{ProwlarrURL: server.URL, ProwlarrAPIKey: "test-key"}
	cats, live, err := FetchProwlarrCategories(context.Background(), cfg, server.Client())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !live {
		t.Error("expected live=true on successful fetch")
	}
	if len(cats) != 2 {
		t.Fatalf("expected 2 top-level categories, got %d", len(cats))
	}
	if cats[0].ID != "7000" || cats[0].Name != "Books" {
		t.Errorf("unexpected first category: %+v", cats[0])
	}
	if len(cats[0].Sub) != 3 || cats[0].Sub[1].ID != "7030" || cats[0].Sub[1].Name != "Comics" {
		t.Errorf("unexpected books subcategories: %+v", cats[0].Sub)
	}
}

func TestFetchProwlarrCategories_FallsBackWhenDown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := &config.Config{ProwlarrURL: server.URL, ProwlarrAPIKey: "test-key"}
	cats, live, err := FetchProwlarrCategories(context.Background(), cfg, server.Client())
	if err != nil {
		t.Fatalf("fallback must not error, got %v", err)
	}
	if live {
		t.Error("expected live=false on fetch failure")
	}
	if len(cats) == 0 {
		t.Fatal("expected non-empty fallback categories")
	}
	// Fallback must cover the tabs' current defaults (no behavior loss offline).
	ids := map[string]bool{}
	var collect func(c []Category)
	collect = func(c []Category) {
		for _, cat := range c {
			ids[cat.ID] = true
			collect(cat.Sub)
		}
	}
	collect(cats)
	for _, want := range []string{"7000", "7020", "7030", "3030"} {
		if !ids[want] {
			t.Errorf("fallback missing category %s", want)
		}
	}
}

func TestFetchProwlarrCategories_RejectsHTML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html>proxy error</html>"))
	}))
	defer server.Close()

	cfg := &config.Config{ProwlarrURL: server.URL, ProwlarrAPIKey: "test-key"}
	cats, live, _ := FetchProwlarrCategories(context.Background(), cfg, server.Client())
	if live || len(cats) == 0 {
		t.Error("expected fallback categories on HTML response")
	}
}

func TestProwlarrCapsCache_CachesAndExpires(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(capsTreeFixture))
	}))
	defer server.Close()

	cfg := &config.Config{ProwlarrURL: server.URL, ProwlarrAPIKey: "test-key"}
	cache := NewProwlarrCapsCache(cfg, server.Client(), time.Hour)

	if _, _, err := cache.Categories(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, _, err := cache.Categories(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls.Load() != 1 {
		t.Errorf("expected 1 HTTP call thanks to cache, got %d", calls.Load())
	}

	stale := NewProwlarrCapsCache(cfg, server.Client(), -time.Second)
	if _, _, err := stale.Categories(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, _, err := stale.Categories(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls.Load() != 3 {
		t.Errorf("expected expired cache to refetch every time, got %d total calls", calls.Load())
	}
}

func TestProwlarrCapsCache_Flatten(t *testing.T) {
	cats := []Category{
		{ID: "7000", Name: "Books", Sub: []Category{{ID: "7020", Name: "EBook"}}},
	}
	flat := FlattenCategories(cats)
	if len(flat) != 2 || flat[0].ID != "7000" || flat[1].ID != "7020" {
		t.Errorf("unexpected flatten output: %+v", flat)
	}
	var _ = json.Marshal
}
