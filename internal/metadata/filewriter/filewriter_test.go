package filewriter

import (
	"archive/zip"
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func ptr(s string) *string { return &s }

// writeZip builds a zip at path with the given entries. If mimetypeFirst is
// set, a stored "mimetype" entry is written first (EPUB requirement).
func writeZip(t *testing.T, path string, entries map[string]string, mimetypeFirst bool) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	if mimetypeFirst {
		h := &zip.FileHeader{Name: "mimetype", Method: zip.Store}
		w, _ := zw.CreateHeader(h)
		w.Write([]byte("application/epub+zip"))
	}
	for name, content := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(content))
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
}

func readEntry(t *testing.T, path, name string) string {
	t.Helper()
	b, err := readZipEntry(path, name)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

const containerXML = `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles><rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/></rootfiles>
</container>`

const opfXML = `<?xml version="1.0" encoding="utf-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:opf="http://www.idpf.org/2007/opf">
    <dc:title>Old Title</dc:title>
    <dc:creator>Old Author</dc:creator>
    <dc:language>en</dc:language>
    <dc:subject>Fiction</dc:subject>
  </metadata>
  <manifest><item id="c" href="c.xhtml" media-type="application/xhtml+xml"/></manifest>
  <spine><itemref idref="c"/></spine>
</package>`

func TestApplyEPUB(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "book.epub")
	writeZip(t, path, map[string]string{
		"META-INF/container.xml": containerXML,
		"OEBPS/content.opf":      opfXML,
		"OEBPS/c.xhtml":          "<html></html>",
	}, true)

	err := Apply(path, Fields{
		Title:       ptr("New Title"),
		Summary:     ptr("A summary & <tale>"),
		Series:      ptr("My Series"),
		SeriesIndex: ptr("3"),
		Publisher:   ptr("Acme"),
		ISBN:        ptr("9781234567897"),
		Authors:     []string{"Jane Doe", "John Roe"},
		Genres:      []string{"Sci-Fi", "Adventure"},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	opf := readEntry(t, path, "OEBPS/content.opf")

	// The OPF must still be well-formed XML after our surgical edit.
	if err := xml.Unmarshal([]byte(opf), new(struct{})); err != nil {
		t.Fatalf("edited OPF is not well-formed: %v\n%s", err, opf)
	}

	wantContains := []string{
		"<dc:title>New Title</dc:title>",
		"A summary &amp; &lt;tale&gt;",
		`<meta name="calibre:series" content="My Series"/>`,
		`<meta name="calibre:series_index" content="3"/>`,
		"<dc:publisher>Acme</dc:publisher>",
		"<dc:creator>Jane Doe</dc:creator>",
		"<dc:creator>John Roe</dc:creator>",
		"<dc:subject>Sci-Fi</dc:subject>",
		`opf:scheme="ISBN">9781234567897`,
	}
	for _, w := range wantContains {
		if !strings.Contains(opf, w) {
			t.Errorf("OPF missing %q\n---\n%s", w, opf)
		}
	}
	// Old author replaced (not appended, since AppendLists=false).
	if strings.Contains(opf, "Old Author") {
		t.Errorf("old author should have been replaced\n%s", opf)
	}
	if strings.Count(opf, "<dc:title>") != 1 {
		t.Errorf("expected exactly one dc:title\n%s", opf)
	}

	// The stored mimetype entry must survive and stay uncompressed & first.
	zr, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	if zr.File[0].Name != "mimetype" || zr.File[0].Method != zip.Store {
		t.Errorf("mimetype not first/stored after rewrite: name=%s method=%d", zr.File[0].Name, zr.File[0].Method)
	}
}

func TestApplyEPUBAppend(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "book.epub")
	writeZip(t, path, map[string]string{
		"META-INF/container.xml": containerXML,
		"OEBPS/content.opf":      opfXML,
	}, true)

	if err := Apply(path, Fields{Genres: []string{"Adventure", "Mystery"}, AppendLists: true}); err != nil {
		t.Fatal(err)
	}
	opf := readEntry(t, path, "OEBPS/content.opf")
	for _, g := range []string{"Fiction", "Adventure", "Mystery"} {
		if !strings.Contains(opf, "<dc:subject>"+g+"</dc:subject>") {
			t.Errorf("append lost genre %q\n%s", g, opf)
		}
	}
}

func TestApplyCBZ(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vol1.cbz")
	writeZip(t, path, map[string]string{
		"001.jpg": "fakeimage",
	}, false)

	err := Apply(path, Fields{
		Series:      ptr("One Piece"),
		SeriesIndex: ptr("1"),
		Summary:     ptr("Pirates!"),
		Authors:     []string{"Eiichiro Oda"},
		Genres:      []string{"Action", "Adventure"},
		Tags:        []string{"shounen"},
	})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	raw := readEntry(t, path, comicInfoName)
	var ci comicInfo
	if err := xml.Unmarshal([]byte(raw), &ci); err != nil {
		t.Fatalf("ComicInfo not well-formed: %v\n%s", err, raw)
	}
	if ci.Series != "One Piece" || ci.Number != "1" || ci.Summary != "Pirates!" {
		t.Errorf("unexpected ComicInfo: %+v", ci)
	}
	if ci.Writer != "Eiichiro Oda" {
		t.Errorf("writer = %q", ci.Writer)
	}
	if ci.Genre != "Action, Adventure" {
		t.Errorf("genre = %q", ci.Genre)
	}
	if ci.Tags != "shounen" {
		t.Errorf("tags = %q", ci.Tags)
	}
	// Original image entry preserved.
	if got := readEntry(t, path, "001.jpg"); got != "fakeimage" {
		t.Errorf("image entry corrupted: %q", got)
	}
}

func TestApplyCBZPreservesUnmanagedFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "vol1.cbz")
	existing := `<?xml version="1.0"?><ComicInfo><Series>Old</Series><Characters>Luffy</Characters><Genre>Action</Genre></ComicInfo>`
	writeZip(t, path, map[string]string{comicInfoName: existing}, false)

	if err := Apply(path, Fields{Series: ptr("New"), Genres: []string{"Comedy"}, AppendLists: true}); err != nil {
		t.Fatal(err)
	}
	var ci comicInfo
	xml.Unmarshal([]byte(readEntry(t, path, comicInfoName)), &ci)
	if ci.Series != "New" {
		t.Errorf("series = %q", ci.Series)
	}
	if ci.Characters != "Luffy" {
		t.Errorf("unmanaged field Characters lost: %q", ci.Characters)
	}
	if ci.Genre != "Action, Comedy" {
		t.Errorf("genre append = %q", ci.Genre)
	}
}

func TestApplyUnsupported(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "book.cbr")
	os.WriteFile(path, []byte("rar"), 0644)
	err := Apply(path, Fields{Title: ptr("x")})
	if _, ok := err.(ErrUnsupported); !ok {
		t.Errorf("expected ErrUnsupported, got %v", err)
	}
}

func TestApplyEmptyNoop(t *testing.T) {
	if err := Apply("/nonexistent/path.epub", Fields{}); err != nil {
		t.Errorf("empty fields should be a no-op, got %v", err)
	}
}
