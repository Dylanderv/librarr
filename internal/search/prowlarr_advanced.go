package search

import (
	"context"

	"github.com/JeremiahM37/librarr/internal/models"
)

// ProwlarrOptions carries per-request overrides for a Prowlarr search,
// typically from the UI's advanced search panel. Zero value means "use the
// tab defaults" (categoriesForTab + searchVariants).
type ProwlarrOptions struct {
	// Query replaces the whole variant list with this exact Prowlarr query
	// (no "manga"/"audiobook"/"comic" suffix appended).
	Query string
	// Categories replaces the tab-default Newznab category IDs. Empty keeps
	// the tab defaults.
	Categories []string
	// UseOverride enables Query/Categories. False behaves like Search.
	UseOverride bool
}

// SearchWithOptions runs a Prowlarr search, honoring ProwlarrOptions when
// UseOverride is set and falling back to the tab-default variants otherwise.
//
// Kept in its own file (rather than inside prowlarr.go) so upstream changes
// to the default variants merge without touching fork lines.
func (p *Prowlarr) SearchWithOptions(ctx context.Context, displayQuery string, opts ProwlarrOptions) ([]models.SearchResult, error) {
	if !opts.UseOverride || opts.Query == "" {
		return p.Search(ctx, displayQuery)
	}
	categories := opts.Categories
	if len(categories) == 0 {
		categories = p.categoriesForTab()
	}
	results, err := p.doSearch(ctx, prowlarrSearchParams{
		query:      opts.Query,
		categories: categories,
		limit:      50,
	})
	if err != nil {
		return nil, err
	}
	return dedupeProwlarrResults(results), nil
}

// dedupeProwlarrResults drops repeated info hashes, preserving order.
// Same rule as Search; duplicated here so SearchWithOptions does not have to
// rewire Search's variant loop.
func dedupeProwlarrResults(results []models.SearchResult) []models.SearchResult {
	seen := make(map[string]bool)
	out := results[:0]
	for _, r := range results {
		if r.InfoHash != "" {
			if seen[r.InfoHash] {
				continue
			}
			seen[r.InfoHash] = true
		}
		out = append(out, r)
	}
	return out
}
