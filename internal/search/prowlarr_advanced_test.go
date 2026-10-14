package search

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/JeremiahM37/librarr/internal/config"
	"github.com/JeremiahM37/librarr/internal/models"
)

func recordProwlarrRequests(t *testing.T, hits *[]map[string][]string, counter *atomic.Int32) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if counter != nil {
			counter.Add(1)
		}
		if hits != nil {
			q := r.URL.Query()
			*hits = append(*hits, map[string][]string{
				"query":      {q.Get("query")},
				"categories": q["categories"],
			})
		}
		json.NewEncoder(w).Encode([]prowlarrItem{
			{Title: "Berserk Deluxe Edition", Size: 100, Seeders: 3, Indexer: "idx", InfoHash: "aa", GUID: "g1"},
		})
	}))
}

func TestProwlarr_SearchWithOptions_UsesExactQueryAndCategories(t *testing.T) {
	var hits []map[string][]string
	server := recordProwlarrRequests(t, &hits, nil)
	defer server.Close()

	cfg := &config.Config{ProwlarrURL: server.URL, ProwlarrAPIKey: "test-key"}
	p := NewProwlarr(cfg, server.Client(), "manga")

	results, err := p.SearchWithOptions(context.Background(), "berserk", ProwlarrOptions{
		Query:       "berserk deluxe edition",
		Categories:  []string{"7020", "7030"},
		UseOverride: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if len(hits) != 1 {
		t.Fatalf("expected exactly 1 Prowlarr request (no suffix variants), got %d", len(hits))
	}
	if got := hits[0]["query"][0]; got != "berserk deluxe edition" {
		t.Errorf("expected exact override query, got %q", got)
	}
	cats := strings.Join(hits[0]["categories"], ",")
	if cats != "7020,7030" {
		t.Errorf("expected categories [7020 7030], got %v", hits[0]["categories"])
	}
}

func TestProwlarr_SearchWithOptions_EmptyFallsBackToVariants(t *testing.T) {
	var hits []map[string][]string
	server := recordProwlarrRequests(t, &hits, nil)
	defer server.Close()

	cfg := &config.Config{ProwlarrURL: server.URL, ProwlarrAPIKey: "test-key"}
	p := NewProwlarr(cfg, server.Client(), "manga")

	// UseOverride=false must behave exactly like Search (3 variants on manga tab).
	if _, err := p.SearchWithOptions(context.Background(), "berserk", ProwlarrOptions{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hits) != 3 {
		t.Fatalf("expected 3 variant requests, got %d", len(hits))
	}
}

func TestProwlarr_MangaVariants_IncludeComicSuffix(t *testing.T) {
	cfg := &config.Config{}
	p := NewProwlarr(cfg, http.DefaultClient, "manga")

	var queries []string
	for _, v := range p.searchVariants("berserk", []string{"7020", "7030"}) {
		queries = append(queries, v.query)
	}
	found := false
	for _, q := range queries {
		if q == "berserk comic" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a \"berserk comic\" variant for comics support, got %v", queries)
	}
}

// overrideStubSearcher is a minimal Searcher for Manager-level tests.
type overrideStubSearcher struct {
	tab     string
	seen    []string
	results []models.SearchResult
}

func (s *overrideStubSearcher) Name() string      { return "stub_" + s.tab }
func (s *overrideStubSearcher) Label() string     { return "Stub" }
func (s *overrideStubSearcher) Enabled() bool     { return true }
func (s *overrideStubSearcher) SearchTab() string { return s.tab }
func (s *overrideStubSearcher) DownloadType() string {
	return "direct"
}
func (s *overrideStubSearcher) Search(_ context.Context, query string) ([]models.SearchResult, error) {
	s.seen = append(s.seen, query)
	return s.results, nil
}

func TestManager_SearchAdvanced_RoutesOverrideToProwlarrOnly(t *testing.T) {
	var hits []map[string][]string
	server := recordProwlarrRequests(t, &hits, nil)
	defer server.Close()

	cfg := &config.Config{
		ProwlarrURL:    server.URL,
		ProwlarrAPIKey: "test-key",
	}
	prow := NewProwlarr(cfg, server.Client(), "manga")
	stub := &overrideStubSearcher{tab: "manga"}
	mgr := NewManager(cfg, []Searcher{prow, stub}, NewHealthTracker(5, 60))

	results, _ := mgr.SearchAdvanced(context.Background(), "manga", "berserk", "", ProwlarrOptions{
		Query:       "berserk deluxe",
		Categories:  []string{"7020"},
		UseOverride: true,
	})
	_ = results

	if len(hits) != 1 {
		t.Fatalf("expected 1 Prowlarr request, got %d", len(hits))
	}
	if got := hits[0]["query"][0]; got != "berserk deluxe" {
		t.Errorf("expected override query at Prowlarr, got %q", got)
	}
	if len(stub.seen) != 1 || stub.seen[0] != "berserk" {
		t.Errorf("expected other sources to receive the display query, got %v", stub.seen)
	}
}
