package api

import (
	"encoding/json"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/JeremiahM37/librarr/internal/config"
	"github.com/JeremiahM37/librarr/internal/db"
	"github.com/JeremiahM37/librarr/internal/models"
)

// mangaLibraryTestServer seeds two manga rows carrying source_ids — the field the
// download path's duplicate check keys on — plus one ebook, so media_type
// filtering stays covered. Kavita is configured on purpose: the manga tab must
// list librarr's own rows regardless, since a stale row is exactly what a Kavita
// mirror cannot show.
func mangaLibraryTestServer(t *testing.T) (*Server, map[string]int64) {
	t.Helper()
	database, err := db.New(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("db.New: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	items := []models.LibraryItem{
		{Title: "Boruto Two Blue Vortex Tome 03", Author: "Masashi Kishimoto", FilePath: "/manga/Boruto/t03.cbz", FileFormat: "cbz", MediaType: "manga", Source: "prowlarr_manga", SourceID: "54dcb86e6bc2282d1583e5519dc34f6c692d9aba"},
		{Title: "One Piece Tome 105", Author: "Eiichiro Oda", FilePath: "/manga/One Piece/t105.cbz", FileFormat: "cbz", MediaType: "manga", Source: "prowlarr_manga", SourceID: "aaaa1111"},
		{Title: "An Ebook", Author: "E. Bookworm", FilePath: "/x/ebook.epub", FileFormat: "epub", MediaType: "ebook", Source: "manual_import"},
	}
	ids := make(map[string]int64, len(items))
	for i := range items {
		id, err := database.AddItem(&items[i])
		if err != nil {
			t.Fatalf("AddItem(%s): %v", items[i].Title, err)
		}
		ids[items[i].Title] = id
	}

	return &Server{
		cfg: &config.Config{
			KavitaURL:  "http://kavita.invalid",
			KavitaUser: "user",
			KavitaPass: "pass",
		},
		db: database,
	}, ids
}

func getMangaLibrary(t *testing.T, s *Server, target string) []map[string]interface{} {
	t.Helper()
	req := httptest.NewRequest("GET", target, nil)
	rr := httptest.NewRecorder()
	s.handleLibraryManga(rr, req)

	if rr.Code != 200 {
		t.Fatalf("GET %s: HTTP %d: %s", target, rr.Code, rr.Body.String())
	}
	var resp struct {
		Items []map[string]interface{} `json:"items"`
		Total int                      `json:"total"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode body: %v\nbody=%s", err, rr.Body.String())
	}
	if resp.Total != len(resp.Items) {
		t.Errorf("total=%d disagrees with %d items returned", resp.Total, len(resp.Items))
	}
	return resp.Items
}

// TestLibraryManga_ListsLocalRowsWithKavitaConfigured is the regression test for
// the invisible-duplicate bug: the tab used to mirror Kavita's series list when
// Kavita was configured, so a library_items row left behind after its torrent was
// removed blocked re-downloads (HTTP 409) while appearing nowhere in the UI.
func TestLibraryManga_ListsLocalRowsWithKavitaConfigured(t *testing.T) {
	s, _ := mangaLibraryTestServer(t)

	items := getMangaLibrary(t, s, "/api/library/manga")

	if len(items) != 2 {
		t.Fatalf("expected the 2 local manga rows, got %d: %+v", len(items), items)
	}
	for _, it := range items {
		if mt, _ := it["media_type"].(string); mt != "manga" {
			t.Errorf("expected manga, got media_type=%q", mt)
		}
		// The delete button and the duplicate check both need these.
		if it["id"] == nil {
			t.Errorf("row has no id, cannot be deleted from the UI: %+v", it)
		}
		if sid, _ := it["source_id"].(string); sid == "" {
			t.Errorf("row has no source_id, cannot be tied back to a 409: %+v", it)
		}
	}
}

func TestLibraryManga_SearchFiltersOnTitleAndSourceID(t *testing.T) {
	s, _ := mangaLibraryTestServer(t)

	t.Run("by title", func(t *testing.T) {
		items := getMangaLibrary(t, s, "/api/library/manga?q=boruto")
		if len(items) != 1 {
			t.Fatalf("expected 1 match for q=boruto, got %d: %+v", len(items), items)
		}
	})

	// Searching by info hash is what turns "I got a 409 for this hash" into a
	// row you can find and remove.
	t.Run("by source id", func(t *testing.T) {
		items := getMangaLibrary(t, s, "/api/library/manga?q=54dcb86e6bc2282d1583e5519dc34f6c692d9aba")
		if len(items) != 1 {
			t.Fatalf("expected 1 match for the info hash, got %d: %+v", len(items), items)
		}
		if title, _ := items[0]["title"].(string); title != "Boruto Two Blue Vortex Tome 03" {
			t.Errorf("matched the wrong row: %q", title)
		}
	})

	t.Run("no match", func(t *testing.T) {
		if items := getMangaLibrary(t, s, "/api/library/manga?q=naruto"); len(items) != 0 {
			t.Fatalf("expected no match for q=naruto, got %+v", items)
		}
	})

	// A bare wildcard must be matched literally, not expanded by LIKE.
	t.Run("wildcards are literal", func(t *testing.T) {
		if items := getMangaLibrary(t, s, "/api/library/manga?q=%25"); len(items) != 0 {
			t.Fatalf("expected '%%' to match nothing, got %d rows", len(items))
		}
	})
}

// TestDeleteManga_ReleasesTheDuplicateCheck ties the two ends together: removing
// the row is what lets the same release be downloaded again.
func TestDeleteManga_ReleasesTheDuplicateCheck(t *testing.T) {
	s, ids := mangaLibraryTestServer(t)
	const hash = "54dcb86e6bc2282d1583e5519dc34f6c692d9aba"

	if !s.db.HasSourceID(hash) {
		t.Fatal("precondition: the seeded row should be seen as a duplicate")
	}

	req := httptest.NewRequest("DELETE", "/api/library/manga/{id}", nil)
	req.SetPathValue("id", strconv.FormatInt(ids["Boruto Two Blue Vortex Tome 03"], 10))
	rr := httptest.NewRecorder()
	s.handleDeleteManga(rr, req)

	if rr.Code != 200 {
		t.Fatalf("HTTP %d: %s", rr.Code, rr.Body.String())
	}
	if s.db.HasSourceID(hash) {
		t.Error("the row was deleted but the duplicate check still trips — a re-download would still 409")
	}
	if items := getMangaLibrary(t, s, "/api/library/manga"); len(items) != 1 {
		t.Errorf("expected 1 remaining manga row, got %d", len(items))
	}
}

func TestDeleteManga_RejectsNonNumericID(t *testing.T) {
	s, _ := mangaLibraryTestServer(t)

	req := httptest.NewRequest("DELETE", "/api/library/manga/{id}", nil)
	req.SetPathValue("id", "not-an-id")
	rr := httptest.NewRecorder()
	s.handleDeleteManga(rr, req)

	if rr.Code != 400 {
		t.Errorf("expected HTTP 400 for a non-numeric id, got %d: %s", rr.Code, rr.Body.String())
	}
}
