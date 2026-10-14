package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/JeremiahM37/librarr/internal/config"
)

// Category is one node of Prowlarr's default indexer category tree
// (GET /api/v1/indexer/categories).
type Category struct {
	ID   string     `json:"id"`
	Name string     `json:"name"`
	Sub  []Category `json:"sub"`
}

// DefaultProwlarrCategories is the fallback tree used when Prowlarr is
// unreachable, so advanced search keeps working offline. IDs mirror the
// tab defaults in categoriesForTab.
func DefaultProwlarrCategories() []Category {
	return []Category{
		{ID: "7000", Name: "Books", Sub: []Category{
			{ID: "7020", Name: "EBook"},
			{ID: "7030", Name: "Comics & Manga"},
		}},
		{ID: "3000", Name: "Audio", Sub: []Category{
			{ID: "3030", Name: "Audiobook"},
		}},
	}
}

// FlattenCategories returns the tree in depth-first order (parents first).
func FlattenCategories(cats []Category) []Category {
	var out []Category
	var walk func(c []Category)
	walk = func(c []Category) {
		for _, cat := range c {
			out = append(out, Category{ID: cat.ID, Name: cat.Name})
			walk(cat.Sub)
		}
	}
	walk(cats)
	return out
}

type prowlarrAPICategory struct {
	ID            int64                 `json:"id"`
	Name          string                `json:"name"`
	Description   string                `json:"description"`
	SubCategories []prowlarrAPICategory `json:"subCategories"`
}

func toCategory(c prowlarrAPICategory) Category {
	out := Category{ID: fmt.Sprintf("%d", c.ID), Name: c.Name}
	if out.Name == "" {
		out.Name = c.Description
	}
	for _, sub := range c.SubCategories {
		out.Sub = append(out.Sub, toCategory(sub))
	}
	return out
}

// FetchProwlarrCategories returns Prowlarr's default indexer category tree.
// It never fails hard: any fetch/parse problem falls back to
// DefaultProwlarrCategories with live=false, so callers (and hermetic tests)
// don't depend on a running Prowlarr.
func FetchProwlarrCategories(ctx context.Context, cfg *config.Config, client *http.Client) (cats []Category, live bool, err error) {
	fallback := func() ([]Category, bool, error) {
		return DefaultProwlarrCategories(), false, nil
	}
	if cfg == nil || !cfg.HasProwlarr() {
		return fallback()
	}

	fetchCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(fetchCtx, "GET", strings.TrimRight(cfg.ProwlarrURL, "/")+"/api/v1/indexer/categories", nil)
	if err != nil {
		slog.Warn("prowlarr categories request failed, using fallback", "error", err)
		return fallback()
	}
	req.Header.Set("X-Api-Key", cfg.ProwlarrAPIKey)

	resp, err := client.Do(req)
	if err != nil {
		slog.Warn("prowlarr categories fetch failed, using fallback", "error", err)
		return fallback()
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		slog.Warn("prowlarr categories bad status, using fallback", "status", resp.StatusCode)
		return fallback()
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		slog.Warn("prowlarr categories read failed, using fallback", "error", err)
		return fallback()
	}
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" || strings.HasPrefix(trimmed, "<") {
		slog.Warn("prowlarr categories returned non-JSON, using fallback")
		return fallback()
	}
	var raw []prowlarrAPICategory
	if err := json.Unmarshal(body, &raw); err != nil {
		slog.Warn("prowlarr categories decode failed, using fallback", "error", err)
		return fallback()
	}
	for _, c := range raw {
		cats = append(cats, toCategory(c))
	}
	if len(cats) == 0 {
		return fallback()
	}
	return cats, true, nil
}

// ProwlarrCapsCache caches the category tree so the advanced search panel
// doesn't hit Prowlarr on every keystroke.
type ProwlarrCapsCache struct {
	mu     sync.Mutex
	cfg    *config.Config
	client *http.Client
	ttl    time.Duration

	cats    []Category
	live    bool
	fetched time.Time
}

// NewProwlarrCapsCache creates a cache with the given TTL. Non-positive TTL
// disables caching (every call refetches; useful in tests).
func NewProwlarrCapsCache(cfg *config.Config, client *http.Client, ttl time.Duration) *ProwlarrCapsCache {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &ProwlarrCapsCache{cfg: cfg, client: client, ttl: ttl}
}

// Categories returns the cached tree when fresh, refetching otherwise.
func (c *ProwlarrCapsCache) Categories(ctx context.Context) ([]Category, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ttl > 0 && !c.fetched.IsZero() && time.Since(c.fetched) < c.ttl && c.cats != nil {
		return c.cats, c.live, nil
	}
	cats, live, err := FetchProwlarrCategories(ctx, c.cfg, c.client)
	if err != nil {
		return nil, false, err
	}
	c.cats, c.live, c.fetched = cats, live, time.Now()
	return cats, live, nil
}
