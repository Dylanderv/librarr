package search

import (
	"github.com/JeremiahM37/librarr/internal/models"
)

// Demote reasons explain why a result would previously have been filtered
// out. Results are never dropped: they sink to the bottom of the list.
const (
	// DemoteLanguage marks a foreign-language title on a non-multilang source.
	DemoteLanguage = "language"
	// DemoteRelevance marks a title that doesn't match the query.
	DemoteRelevance = "relevance"
	// DemoteSuspicious marks a title with non-book keywords.
	DemoteSuspicious = "suspicious"
	// DemoteNoSeeders marks a torrent with no seeders.
	DemoteNoSeeders = "no_seeders"
	// DemoteSize marks a torrent outside the configured size bounds.
	DemoteSize = "size"
	// DemoteNoSource marks an AudioBookBay row with no usable source.
	DemoteNoSource = "no_source"
)

// demote marks a result as filtered-out without dropping it.
func demote(r *models.SearchResult, reason string) {
	r.Demoted = true
	r.DemoteReason = reason
}

// sinkDemotedLast stable-partitions demoted results to the bottom,
// preserving relative order on both sides.
func sinkDemotedLast(results []models.SearchResult) []models.SearchResult {
	kept := results[:0]
	var demoted []models.SearchResult
	for _, r := range results {
		if r.Demoted {
			demoted = append(demoted, r)
		} else {
			kept = append(kept, r)
		}
	}
	return append(kept, demoted...)
}
