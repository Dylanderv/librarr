package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/JeremiahM37/librarr/internal/config"
)

// newBrowseFixture builds a server whose manga library contains one series
// folder with a mix of formats, plus a decoy sibling root.
func newBrowseFixture(t *testing.T) (*Server, string) {
	t.Helper()
	resetLibraryRootCache()
	resetDirListingCache()
	t.Cleanup(resetLibraryRootCache)
	t.Cleanup(resetDirListingCache)

	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("EvalSymlinks: %v", err)
	}
	manga := filepath.Join(base, "manga")
	ebooks := filepath.Join(base, "ebooks")
	for _, d := range []string{manga, ebooks} {
		if err := os.MkdirAll(d, 0o777); err != nil {
			t.Fatalf("MkdirAll %s: %v", d, err)
		}
	}
	cfg := &config.Config{MangaDir: manga, EbookDir: ebooks}
	return &Server{cfg: cfg}, manga
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o777); err != nil {
		t.Fatal(err)
	}
}

func mustWriteN(t *testing.T, path string, n int) {
	t.Helper()
	mustMkdir(t, filepath.Dir(path))
	if err := os.WriteFile(path, make([]byte, n), 0o666); err != nil {
		t.Fatal(err)
	}
}

func doList(t *testing.T, s *Server, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/files/list?path="+path, nil)
	rec := httptest.NewRecorder()
	s.handleFilesList(rec, req)
	return rec
}

func decodeList(t *testing.T, rec *httptest.ResponseRecorder) listResponse {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var out listResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v (body %s)", err, rec.Body.String())
	}
	return out
}

func TestFilesRootsListsConfiguredLibraries(t *testing.T) {
	s, manga := newBrowseFixture(t)

	rec := httptest.NewRecorder()
	s.handleFilesRoots(rec, httptest.NewRequest(http.MethodGet, "/api/files/roots", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var out struct {
		Roots []libraryRoot `json:"roots"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(out.Roots) != 2 {
		t.Fatalf("expected 2 roots, got %+v", out.Roots)
	}
	if out.Roots[0].ID != "manga" || out.Roots[0].Path != manga {
		t.Errorf("unexpected root: %+v", out.Roots[0])
	}
}

func TestFilesRootsWithoutConfiguredLibraries(t *testing.T) {
	resetLibraryRootCache()
	t.Cleanup(resetLibraryRootCache)
	s := &Server{cfg: &config.Config{}}

	rec := httptest.NewRecorder()
	s.handleFilesRoots(rec, httptest.NewRequest(http.MethodGet, "/api/files/roots", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	// An empty JSON array, never null: the front end iterates it directly.
	if got := rec.Body.String(); !jsonHasEmptyRoots(t, got) {
		t.Errorf("expected an empty roots array, got %s", got)
	}
}

func jsonHasEmptyRoots(t *testing.T, body string) bool {
	t.Helper()
	var out map[string]json.RawMessage
	if err := json.Unmarshal([]byte(body), &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return string(out["roots"]) == "[]"
}

func TestFilesListRootLevel(t *testing.T) {
	s, manga := newBrowseFixture(t)
	mustMkdir(t, filepath.Join(manga, "Berserk"))
	mustWriteN(t, filepath.Join(manga, "loose.cbz"), 10)

	out := decodeList(t, doList(t, s, manga))
	if out.Root != "manga" || out.RootPath != manga || !out.IsRoot {
		t.Errorf("unexpected root fields: %+v", out)
	}
	if out.Parent != "" {
		t.Errorf("root must have no parent, got %q", out.Parent)
	}
	if len(out.Segments) != 0 {
		t.Errorf("root breadcrumb must be empty, got %+v", out.Segments)
	}
	if len(out.Entries) != 2 || !out.Entries[0].IsDir || out.Entries[0].Name != "Berserk" {
		t.Fatalf("directories must come first: %+v", out.Entries)
	}
	if out.Entries[0].Kind != "" {
		t.Errorf("a directory has no kind, got %q", out.Entries[0].Kind)
	}
}

func TestFilesListClassifiesKindAndEditable(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Berserk")
	files := map[string]struct {
		kind     string
		editable bool
	}{
		"a.cbz":  {"cbz", true},
		"b.EPUB": {"epub", true},
		"c.pdf":  {"pdf", false},
		"d.cbr":  {"cbr", false},
		"e.m4b":  {"audio", false},
		"f.txt":  {"other", false},
	}
	for name := range files {
		mustWriteN(t, filepath.Join(dir, name), 3)
	}

	out := decodeList(t, doList(t, s, dir))
	if len(out.Entries) != len(files) {
		t.Fatalf("got %d entries, want %d", len(out.Entries), len(files))
	}
	for _, e := range out.Entries {
		want, ok := files[e.Name]
		if !ok {
			t.Fatalf("unexpected entry %q", e.Name)
		}
		if e.Kind != want.kind || e.Editable != want.editable {
			t.Errorf("%s: kind=%q editable=%v, want %q/%v", e.Name, e.Kind, e.Editable, want.kind, want.editable)
		}
		if e.Size != 3 {
			t.Errorf("%s: size = %d, want 3", e.Name, e.Size)
		}
		if e.Modified == "" {
			t.Errorf("%s: missing modified date", e.Name)
		}
	}
}

func TestFilesListNaturalSortAndBreadcrumb(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Berserk", "Deluxe")
	for _, n := range []string{"Vol.10.cbz", "Vol.2.cbz", "vol.1.cbz"} {
		mustWriteN(t, filepath.Join(dir, n), 1)
	}

	out := decodeList(t, doList(t, s, dir))
	got := []string{out.Entries[0].Name, out.Entries[1].Name, out.Entries[2].Name}
	want := []string{"vol.1.cbz", "Vol.2.cbz", "Vol.10.cbz"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("natural sort: got %v, want %v", got, want)
		}
	}
	if out.IsRoot {
		t.Error("a nested directory is not a root")
	}
	if out.Parent != filepath.Join(manga, "Berserk") {
		t.Errorf("parent = %q", out.Parent)
	}
	if len(out.Segments) != 2 ||
		out.Segments[0].Name != "Berserk" || out.Segments[0].Path != filepath.Join(manga, "Berserk") ||
		out.Segments[1].Name != "Deluxe" || out.Segments[1].Path != dir {
		t.Errorf("unexpected breadcrumb: %+v", out.Segments)
	}
}

func TestFilesListHidesDotfilesAndSynologySidecar(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Berserk")
	mustWriteN(t, filepath.Join(dir, "keep.cbz"), 1)
	mustWriteN(t, filepath.Join(dir, ".DS_Store"), 1)
	mustMkdir(t, filepath.Join(dir, "@eaDir"))

	out := decodeList(t, doList(t, s, dir))
	if len(out.Entries) != 1 || out.Entries[0].Name != "keep.cbz" {
		t.Errorf("expected only keep.cbz, got %+v", out.Entries)
	}
}

func TestFilesListRejectsInvalidPaths(t *testing.T) {
	s, manga := newBrowseFixture(t)
	mustWriteN(t, filepath.Join(manga, "a.cbz"), 1)

	cases := []struct {
		name string
		path string
		want int
	}{
		{"missing", "", http.StatusBadRequest},
		{"relative", "manga/Berserk", http.StatusBadRequest},
		{"traversal", filepath.Join(manga, "..", "..", "etc"), http.StatusForbidden},
		{"outside", "/etc", http.StatusForbidden},
		{"not found", filepath.Join(manga, "nope"), http.StatusNotFound},
		{"not a directory", filepath.Join(manga, "a.cbz"), http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := doList(t, s, tc.path).Code; got != tc.want {
				t.Errorf("status = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestFilesListWithoutConfiguredRoots(t *testing.T) {
	resetLibraryRootCache()
	resetDirListingCache()
	t.Cleanup(resetLibraryRootCache)
	t.Cleanup(resetDirListingCache)

	s := &Server{cfg: &config.Config{}}
	if got := doList(t, s, "/books/manga").Code; got != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", got)
	}
}

func TestDirListingCacheHitHidesFilesystemChange(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Berserk")
	mustWriteN(t, filepath.Join(dir, "a.cbz"), 1)

	if n := len(decodeList(t, doList(t, s, dir)).Entries); n != 1 {
		t.Fatalf("first listing: got %d entries, want 1", n)
	}
	mustWriteN(t, filepath.Join(dir, "b.cbz"), 1)
	if n := len(decodeList(t, doList(t, s, dir)).Entries); n != 1 {
		t.Errorf("second listing should be served from cache, got %d entries", n)
	}
}

func TestDirListingCacheInvalidation(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Berserk")
	mustWriteN(t, filepath.Join(dir, "a.cbz"), 1)

	decodeList(t, doList(t, s, dir))
	mustWriteN(t, filepath.Join(dir, "b.cbz"), 1)
	invalidateDirListing(dir)
	if n := len(decodeList(t, doList(t, s, dir)).Entries); n != 2 {
		t.Errorf("after invalidation: got %d entries, want 2", n)
	}
	// Invalidating an uncached (or empty) path is a no-op, not a panic.
	invalidateDirListing("")
	invalidateDirListing(filepath.Join(manga, "never-listed"))
}

func TestDirListingCacheExpires(t *testing.T) {
	resetDirListingCache()
	t.Cleanup(resetDirListingCache)

	dir := t.TempDir()
	now := time.Now()
	storeDirListing(dir, []fileEntry{{Name: "a.cbz"}}, now)

	if _, ok := cachedDirListing(dir, now.Add(dirListingTTL-time.Second)); !ok {
		t.Error("entry must still be fresh before the TTL")
	}
	if _, ok := cachedDirListing(dir, now.Add(dirListingTTL+time.Second)); ok {
		t.Error("entry must be considered stale past the TTL")
	}
}

func TestDirListingCacheIsBounded(t *testing.T) {
	resetDirListingCache()
	t.Cleanup(resetDirListingCache)

	now := time.Now()
	for i := 0; i < dirListingCacheMax+10; i++ {
		storeDirListing(filepath.Join("/books/manga", strconv.Itoa(i)), nil, now)
	}
	dirListingMu.Lock()
	size := len(dirListingCache)
	dirListingMu.Unlock()
	if size > dirListingCacheMax {
		t.Errorf("cache grew to %d entries, max is %d", size, dirListingCacheMax)
	}
}

func TestNaturalLess(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"Vol.2", "Vol.10", true},
		{"Vol.10", "Vol.2", false},
		{"vol.1", "Vol.2", true},
		{"Vol.01", "Vol.1", true}, // equal numerically: falls back to raw byte order
		{"Berserk", "berserk 2", true},
		{"a", "b", true},
	}
	for _, tc := range cases {
		if got := naturalLess(tc.a, tc.b); got != tc.want {
			t.Errorf("naturalLess(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}
