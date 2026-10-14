package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/JeremiahM37/librarr/internal/search"
)

// Prowlarr advanced search plumbing: per-request query/category overrides
// (?pq=, ?pcat=) plus the live category tree (? sourced from Prowlarr with a
// builtin fallback). Lives here rather than in search.go so upstream edits to
// the shared search handlers stay limited to two call sites.

// maxProwlarrCategories caps how many category IDs a single request may pass.
const maxProwlarrCategories = 20

// parseProwlarrOptions reads the advanced-search overrides from the request.
// pq sets the exact Prowlarr query (no suffix variants); pcat is a
// comma-separated list of Newznab category IDs (digits only, anything else is
// dropped). pcat present-but-empty explicitly clears the category filter;
// pcat absent keeps the tab defaults. No pq and no pcat means "tab defaults".
func parseProwlarrOptions(r *http.Request) search.ProwlarrOptions {
	pq := truncateSearchQuery(strings.TrimSpace(r.URL.Query().Get("pq")))
	_, hasPcat := r.URL.Query()["pcat"]
	var cats []string
	if hasPcat {
		for _, part := range strings.Split(r.URL.Query().Get("pcat"), ",") {
			part = strings.TrimSpace(part)
			if part == "" || len(part) > 6 {
				continue
			}
			digits := true
			for _, c := range part {
				if c < '0' || c > '9' {
					digits = false
					break
				}
			}
			if !digits {
				continue
			}
			if _, err := strconv.Atoi(part); err != nil {
				continue
			}
			cats = append(cats, part)
			if len(cats) >= maxProwlarrCategories {
				break
			}
		}
	}
	if pq == "" && !hasPcat {
		return search.ProwlarrOptions{}
	}
	return search.ProwlarrOptions{Query: pq, Categories: cats, CategoriesSet: hasPcat, UseOverride: true}
}

// handleProwlarrCategories serves Prowlarr's default indexer category tree
// for the advanced search panel. live=false means the builtin fallback (the
// Prowlarr query override still works; only the suggestions are static).
func (s *Server) handleProwlarrCategories(w http.ResponseWriter, r *http.Request) {
	cats, live, err := s.prowlarrCaps.Categories(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, "Failed to load Prowlarr categories", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"live":       live,
		"categories": cats,
	})
}
