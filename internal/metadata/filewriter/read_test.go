package filewriter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func deref(t *testing.T, p *string, field string) string {
	t.Helper()
	if p == nil {
		t.Fatalf("%s: expected a value, got nil", field)
	}
	return *p
}

// TestReadCBZRoundTrip is the contract the preview endpoint depends on: what
// Apply writes is exactly what Read gives back, or the "current value" column
// would lie about what is being overwritten.
func TestReadCBZRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vol1.cbz")
	writeZip(t, path, map[string]string{"001.jpg": "fakeimage"}, false)

	want := Fields{
		Title:       ptr("Romance Dawn"),
		Series:      ptr("One Piece"),
		SeriesIndex: ptr("1"),
		Summary:     ptr("Pirates & <friends>"),
		Publisher:   ptr("Shueisha"),
		Language:    ptr("en"),
		Authors:     []string{"Eiichiro Oda"},
		Genres:      []string{"Action", "Adventure"},
		Tags:        []string{"shounen"},
	}
	if err := Apply(path, want); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if v := deref(t, got.Title, "title"); v != "Romance Dawn" {
		t.Errorf("title = %q", v)
	}
	if v := deref(t, got.Series, "series"); v != "One Piece" {
		t.Errorf("series = %q", v)
	}
	if v := deref(t, got.SeriesIndex, "series_index"); v != "1" {
		t.Errorf("series_index = %q", v)
	}
	if v := deref(t, got.Summary, "summary"); v != "Pirates & <friends>" {
		t.Errorf("summary = %q", v)
	}
	if v := deref(t, got.Publisher, "publisher"); v != "Shueisha" {
		t.Errorf("publisher = %q", v)
	}
	if v := deref(t, got.Language, "language"); v != "en" {
		t.Errorf("language = %q", v)
	}
	if strings.Join(got.Authors, "|") != "Eiichiro Oda" {
		t.Errorf("authors = %v", got.Authors)
	}
	if strings.Join(got.Genres, "|") != "Action|Adventure" {
		t.Errorf("genres = %v", got.Genres)
	}
	if strings.Join(got.Tags, "|") != "shounen" {
		t.Errorf("tags = %v", got.Tags)
	}
	// ComicInfo has no ISBN element in the subset we manage.
	if got.ISBN != nil {
		t.Errorf("isbn = %q, want nil", *got.ISBN)
	}
}

func TestReadEPUBRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "book.epub")
	writeZip(t, path, map[string]string{
		"META-INF/container.xml": containerXML,
		"OEBPS/content.opf":      opfXML,
		"OEBPS/c.xhtml":          "<html></html>",
	}, true)

	if err := Apply(path, Fields{
		Title:       ptr("New Title"),
		Summary:     ptr("A summary & <tale>"),
		Series:      ptr("My Series"),
		SeriesIndex: ptr("3"),
		Publisher:   ptr("Acme"),
		Language:    ptr("fr"),
		ISBN:        ptr("9781234567897"),
		Authors:     []string{"Jane Doe", "John Roe"},
		Genres:      []string{"Sci-Fi", "Adventure"},
	}); err != nil {
		t.Fatalf("Apply: %v", err)
	}

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if v := deref(t, got.Title, "title"); v != "New Title" {
		t.Errorf("title = %q", v)
	}
	if v := deref(t, got.Summary, "summary"); v != "A summary & <tale>" {
		t.Errorf("summary = %q", v)
	}
	if v := deref(t, got.Series, "series"); v != "My Series" {
		t.Errorf("series = %q", v)
	}
	if v := deref(t, got.SeriesIndex, "series_index"); v != "3" {
		t.Errorf("series_index = %q", v)
	}
	if v := deref(t, got.Publisher, "publisher"); v != "Acme" {
		t.Errorf("publisher = %q", v)
	}
	if v := deref(t, got.Language, "language"); v != "fr" {
		t.Errorf("language = %q", v)
	}
	if v := deref(t, got.ISBN, "isbn"); v != "9781234567897" {
		t.Errorf("isbn = %q", v)
	}
	if strings.Join(got.Authors, "|") != "Jane Doe|John Roe" {
		t.Errorf("authors = %v", got.Authors)
	}
	if strings.Join(got.Genres, "|") != "Sci-Fi|Adventure" {
		t.Errorf("genres = %v", got.Genres)
	}
}

// A file with no metadata container is a legitimate, empty read — not a
// failure. The editor must be able to open a brand-new CBZ.
func TestReadCBZWithoutComicInfo(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vol1.cbz")
	writeZip(t, path, map[string]string{"001.jpg": "fakeimage"}, false)

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !got.IsEmpty() {
		t.Errorf("expected empty fields, got %+v", got)
	}
}

func TestReadEPUBWithoutMetadataBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "book.epub")
	bare := `<?xml version="1.0"?><package xmlns="http://www.idpf.org/2007/opf" version="3.0"><manifest/></package>`
	writeZip(t, path, map[string]string{
		"META-INF/container.xml": containerXML,
		"OEBPS/content.opf":      bare,
	}, true)

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !got.IsEmpty() {
		t.Errorf("expected empty fields, got %+v", got)
	}
}

// A structurally broken EPUB is a real error: reporting it as "no metadata"
// would hide a file the user needs to look at.
func TestReadEPUBBrokenContainerIsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "book.epub")
	writeZip(t, path, map[string]string{"OEBPS/content.opf": opfXML}, true)

	if _, err := Read(path); err == nil {
		t.Fatal("expected an error for an EPUB without container.xml")
	}
}

func TestReadUnsupported(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"book.pdf", "book.cbr", "book.m4b"} {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("x"), 0o666); err != nil {
			t.Fatal(err)
		}
		_, err := Read(path)
		if _, ok := err.(ErrUnsupported); !ok {
			t.Errorf("%s: expected ErrUnsupported, got %v", name, err)
		}
	}
}

// Read must not resurrect values Apply cleared: an explicitly emptied field
// reads back as absent, so a second edit does not silently re-propose it.
func TestReadAfterClearingAField(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vol1.cbz")
	writeZip(t, path, map[string]string{"001.jpg": "img"}, false)

	if err := Apply(path, Fields{Series: ptr("Old"), Summary: ptr("Blurb")}); err != nil {
		t.Fatal(err)
	}
	if err := Apply(path, Fields{Summary: ptr("")}); err != nil {
		t.Fatal(err)
	}
	got, err := Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Summary != nil {
		t.Errorf("summary = %q, want nil after being cleared", *got.Summary)
	}
	if v := deref(t, got.Series, "series"); v != "Old" {
		t.Errorf("series = %q, an unrelated field must survive", v)
	}
}

func TestMergeListsMatchesApply(t *testing.T) {
	got := MergeLists([]string{"Action"}, []string{"action", "Comedy"}, true)
	if strings.Join(got, "|") != "Action|Comedy" {
		t.Errorf("MergeLists append = %v", got)
	}
	got = MergeLists([]string{"Action"}, []string{"Comedy"}, false)
	if strings.Join(got, "|") != "Comedy" {
		t.Errorf("MergeLists replace = %v", got)
	}
}
