package search

import (
	"context"
	"testing"

	"github.com/JeremiahM37/librarr/internal/config"
	"github.com/JeremiahM37/librarr/internal/models"
)

func TestSinkDemotedLast_PreservesOrder(t *testing.T) {
	in := []models.SearchResult{
		{Title: "b", Demoted: true, DemoteReason: DemoteNoSeeders},
		{Title: "a"},
		{Title: "c", Demoted: true, DemoteReason: DemoteSize},
		{Title: "d"},
	}
	out := sinkDemotedLast(in)
	want := []string{"a", "d", "b", "c"}
	for i, title := range want {
		if out[i].Title != title {
			t.Fatalf("position %d: expected %q, got %q (full: %+v)", i, title, out[i].Title, out)
		}
	}
}

func TestManager_Search_KeepsDemotedAtBottom(t *testing.T) {
	cfg := &config.Config{MinTorrentSizeBytes: 10000, MaxTorrentSizeBytes: 2000000000}
	// 0-seeder torrent would previously have vanished; oversize one too.
	// Both must survive to the bottom of the scored list.
	stub := &overrideStubSearcher{tab: "main", results: []models.SearchResult{
		{Source: "torrent", Title: "Dead Release", Seeders: 0, Size: 50000},
		{Source: "torrent", Title: "Good Release", Seeders: 5, Size: 50000},
	}}
	mgr := NewManager(cfg, []Searcher{stub}, NewHealthTracker(5, 60))

	results, _ := mgr.Search(context.Background(), "main", "release")
	if len(results) != 2 {
		t.Fatalf("expected 2 results (nothing dropped), got %d", len(results))
	}
	if results[0].Title != "Good Release" || results[0].Demoted {
		t.Errorf("expected kept result first, got %+v", results[0])
	}
	if !results[1].Demoted || results[1].DemoteReason != DemoteNoSeeders {
		t.Errorf("expected demoted 0-seeder last, got %+v", results[1])
	}
}
