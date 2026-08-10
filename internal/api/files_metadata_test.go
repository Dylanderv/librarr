package api

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/JeremiahM37/librarr/internal/metadata/filewriter"
)

// --- fixtures ------------------------------------------------------------
//
// Real CBZ/EPUB archives, built in memory: the point of these tests is that the
// preview reports what is genuinely inside the files and that the apply changes
// it, which a stub file could not show.
//
// File modes are left to os.Create / MkdirAll and the umask, deliberately: the
// production share is NFS with an inherited ACL, and a hard-coded mode there
// caps what Kavita can read.

func writeCBZ(t *testing.T, path string, comicInfo string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	w, err := zw.Create("001.jpg")
	if err != nil {
		t.Fatal(err)
	}
	w.Write([]byte("fakeimage"))
	if comicInfo != "" {
		w, err = zw.Create("ComicInfo.xml")
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(comicInfo))
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
}

const testContainerXML = `<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles><rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/></rootfiles>
</container>`

func writeEPUB(t *testing.T, path, title string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	mw, _ := zw.CreateHeader(&zip.FileHeader{Name: "mimetype", Method: zip.Store})
	mw.Write([]byte("application/epub+zip"))
	entries := map[string]string{
		"META-INF/container.xml": testContainerXML,
		"OEBPS/content.opf": fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:opf="http://www.idpf.org/2007/opf">
    <dc:title>%s</dc:title>
    <dc:language>en</dc:language>
  </metadata>
  <manifest><item id="c" href="c.xhtml" media-type="application/xhtml+xml"/></manifest>
  <spine><itemref idref="c"/></spine>
</package>`, title),
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

func doMetaGet(t *testing.T, s *Server, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/files/metadata?path="+path, nil)
	rec := httptest.NewRecorder()
	s.handleFilesMetadata(rec, req)
	return rec
}

func doMetaPreview(t *testing.T, s *Server, body any) *httptest.ResponseRecorder {
	t.Helper()
	return postJSON(t, s.handleFilesMetadataPreview, "/api/files/metadata/preview", body)
}

func doMetaApply(t *testing.T, s *Server, body any) *httptest.ResponseRecorder {
	t.Helper()
	return postJSON(t, s.handleFilesMetadataApply, "/api/files/metadata/apply", body)
}

func decodeMetaPreview(t *testing.T, rec *httptest.ResponseRecorder) metadataPreviewResponse {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var out metadataPreviewResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v (body %s)", err, rec.Body.String())
	}
	return out
}

func decodeMetaApply(t *testing.T, rec *httptest.ResponseRecorder) metadataApplyResponse {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var out metadataApplyResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v (body %s)", err, rec.Body.String())
	}
	return out
}

func metaRowByName(t *testing.T, rows []metadataFileRow, name string) metadataFileRow {
	t.Helper()
	for _, r := range rows {
		if r.Name == name {
			return r
		}
	}
	t.Fatalf("no preview row for %q in %+v", name, rows)
	return metadataFileRow{}
}

func changeByField(t *testing.T, row metadataFileRow, field string) metadataChange {
	t.Helper()
	for _, c := range row.Changes {
		if c.Field == field {
			return c
		}
	}
	t.Fatalf("no change for field %q in %+v", field, row.Changes)
	return metadataChange{}
}

func strp(s string) *string { return &s }

// --- GET /api/files/metadata --------------------------------------------

func TestMetadataGetReadsFile(t *testing.T) {
	s, manga := newBrowseFixture(t)
	path := filepath.Join(manga, "Berserk", "Vol.01.cbz")
	writeCBZ(t, path, `<?xml version="1.0"?><ComicInfo><Series>Berserk</Series><Number>1</Number><Genre>Action, Dark Fantasy</Genre></ComicInfo>`)

	rec := doMetaGet(t, s, path)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var out struct {
		Kind     string         `json:"kind"`
		Editable bool           `json:"editable"`
		Reason   string         `json:"reason"`
		Fields   metadataValues `json:"fields"`
	}
	json.Unmarshal(rec.Body.Bytes(), &out)
	if out.Kind != "cbz" || !out.Editable || out.Reason != "" {
		t.Fatalf("unexpected envelope: %+v", out)
	}
	if out.Fields.Series != "Berserk" || out.Fields.SeriesIndex != "1" {
		t.Errorf("fields = %+v", out.Fields)
	}
	if len(out.Fields.Genres) != 2 || out.Fields.Genres[0] != "Action" {
		t.Errorf("genres = %v", out.Fields.Genres)
	}
}

// Decision 11: a non-editable format still opens its panel, and says why its
// fields are locked — with a reason specific to the format, not a generic one.
func TestMetadataGetUnsupportedFormatsCarryDistinctReasons(t *testing.T) {
	s, manga := newBrowseFixture(t)
	cases := map[string]string{
		"Vol.01.pdf": metaReasonPDF,
		"Vol.02.cbr": metaReasonCBR,
		"Vol.03.m4b": metaReasonAudio,
	}
	seen := map[string]bool{}
	for name, wantReason := range cases {
		path := filepath.Join(manga, "Berserk", name)
		mustWriteN(t, path, 4)

		rec := doMetaGet(t, s, path)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status = %d", name, rec.Code)
		}
		var out struct {
			Editable bool   `json:"editable"`
			Reason   string `json:"reason"`
		}
		json.Unmarshal(rec.Body.Bytes(), &out)
		if out.Editable {
			t.Errorf("%s: should not be editable", name)
		}
		if out.Reason != wantReason {
			t.Errorf("%s: reason = %q, want %q", name, out.Reason, wantReason)
		}
		seen[out.Reason] = true
	}
	if len(seen) != 3 {
		t.Errorf("the three formats must not share one generic reason: %v", seen)
	}
}

func TestMetadataGetRejectsDirectoryAndTraversal(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Berserk")
	mustMkdir(t, dir)

	if rec := doMetaGet(t, s, dir); rec.Code != http.StatusBadRequest {
		t.Errorf("directory: status = %d, want 400", rec.Code)
	}
	if rec := doMetaGet(t, s, filepath.Join(manga, "..", "..", "outside.cbz")); rec.Code != http.StatusForbidden {
		t.Errorf("traversal out of the roots: status = %d, want 403", rec.Code)
	}
	if rec := doMetaGet(t, s, "/etc/passwd"); rec.Code != http.StatusForbidden {
		t.Errorf("outside path: status = %d, want 403", rec.Code)
	}
}

// --- preview -------------------------------------------------------------

// Decision 4: propagation overwrites, but the preview must show what is being
// overwritten — the current value of every field the edit touches.
func TestMetadataPreviewShowsOverwrittenValues(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Berserk")
	writeCBZ(t, filepath.Join(dir, "Vol.01.cbz"), `<?xml version="1.0"?><ComicInfo><Series>Old Series</Series><Publisher>Old Pub</Publisher></ComicInfo>`)
	writeCBZ(t, filepath.Join(dir, "Vol.02.cbz"), "")

	out := decodeMetaPreview(t, doMetaPreview(t, s, metadataRequest{
		Path:   dir,
		Fields: bulkMetaFields{Series: strp("Berserk"), Publisher: strp("Dark Horse")},
	}))

	if len(out.Files) != 2 {
		t.Fatalf("got %d rows, want 2: %+v", len(out.Files), out.Files)
	}
	row := metaRowByName(t, out.Files, "Vol.01.cbz")
	if row.Status != metaRowEdit {
		t.Errorf("status = %q", row.Status)
	}
	if c := changeByField(t, row, "series"); c.Current != "Old Series" || c.Target != "Berserk" || !c.Changed {
		t.Errorf("series change = %+v; the value being overwritten must be visible", c)
	}
	if c := changeByField(t, row, "publisher"); c.Current != "Old Pub" || c.Target != "Dark Horse" {
		t.Errorf("publisher change = %+v", c)
	}

	// A file with no ComicInfo at all: nothing to overwrite, still an edit.
	empty := metaRowByName(t, out.Files, "Vol.02.cbz")
	if empty.Status != metaRowEdit {
		t.Errorf("status = %q", empty.Status)
	}
	if c := changeByField(t, empty, "series"); c.Current != "" {
		t.Errorf("current = %q, want empty", c.Current)
	}
}

// A file already holding the target values is reported as unchanged and shows
// up in the notices, never in the errors.
func TestMetadataPreviewUnchangedIsANoticeNotAnError(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Berserk")
	writeCBZ(t, filepath.Join(dir, "Vol.01.cbz"), `<?xml version="1.0"?><ComicInfo><Series>Berserk</Series></ComicInfo>`)

	out := decodeMetaPreview(t, doMetaPreview(t, s, metadataRequest{
		Path:   dir,
		Fields: bulkMetaFields{Series: strp("Berserk")},
	}))
	if row := metaRowByName(t, out.Files, "Vol.01.cbz"); row.Status != metaRowUnchanged {
		t.Errorf("status = %q, want unchanged", row.Status)
	}
	if len(out.Errors) != 0 {
		t.Errorf("errors = %v, an up-to-date file is not a failure", out.Errors)
	}
	if len(out.Notices) == 0 || out.Notices[0].Key != metaNoticeUnchanged {
		t.Errorf("notices = %+v", out.Notices)
	}
}

// Non-editable formats are listed with their reason, and the apply does not
// touch them.
func TestMetadataUnsupportedListedButNotWritten(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Berserk")
	writeCBZ(t, filepath.Join(dir, "Vol.01.cbz"), "")
	pdf := filepath.Join(dir, "Vol.02.pdf")
	mustWriteN(t, pdf, 7)
	before, err := os.ReadFile(pdf)
	if err != nil {
		t.Fatal(err)
	}

	out := decodeMetaPreview(t, doMetaPreview(t, s, metadataRequest{
		Path:   dir,
		Fields: bulkMetaFields{Series: strp("Berserk")},
	}))
	row := metaRowByName(t, out.Files, "Vol.02.pdf")
	if row.Status != metaRowUnsupported || row.Reason != metaReasonPDF || row.Editable {
		t.Errorf("pdf row = %+v", row)
	}

	res := decodeMetaApply(t, doMetaApply(t, s, metadataRequest{
		Path:   dir,
		Fields: bulkMetaFields{Series: strp("Berserk")},
	}))
	if res.Written != 1 || res.Unsupported != 1 {
		t.Errorf("apply = %+v", res)
	}
	if len(res.Errors) != 0 {
		t.Errorf("an unsupported format is a notice, not an error: %v", res.Errors)
	}
	after, err := os.ReadFile(pdf)
	if err != nil {
		t.Fatal(err)
	}
	if len(before) != len(after) {
		t.Errorf("the PDF was written to: %d → %d bytes", len(before), len(after))
	}
}

// Title, series_index and ISBN describe one volume; propagating them over a
// folder is refused server-side, not only hidden in the UI.
func TestMetadataPerFileFieldsRefusedOnAFolder(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Berserk")
	writeCBZ(t, filepath.Join(dir, "Vol.01.cbz"), "")

	for _, f := range []bulkMetaFields{
		{Title: strp("Chapter 1")},
		{SeriesIndex: strp("3")},
		{ISBN: strp("9781234567897")},
	} {
		if rec := doMetaPreview(t, s, metadataRequest{Path: dir, Fields: f}); rec.Code != http.StatusBadRequest {
			t.Errorf("preview %+v: status = %d, want 400", f, rec.Code)
		}
		if rec := doMetaApply(t, s, metadataRequest{Path: dir, Fields: f}); rec.Code != http.StatusBadRequest {
			t.Errorf("apply %+v: status = %d, want 400", f, rec.Code)
		}
	}

	// The same fields on a single file are perfectly legitimate.
	file := filepath.Join(dir, "Vol.01.cbz")
	if rec := doMetaPreview(t, s, metadataRequest{Path: file, Fields: bulkMetaFields{Title: strp("Chapter 1")}}); rec.Code != http.StatusOK {
		t.Errorf("single file: status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

// --- recursion and its bound --------------------------------------------

func TestMetadataPreviewRecursion(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Berserk")
	writeCBZ(t, filepath.Join(dir, "Vol.01.cbz"), "")
	writeCBZ(t, filepath.Join(dir, "Extras", "Omake.cbz"), "")

	flat := decodeMetaPreview(t, doMetaPreview(t, s, metadataRequest{
		Path: dir, Fields: bulkMetaFields{Series: strp("Berserk")},
	}))
	if len(flat.Files) != 1 {
		t.Errorf("non-recursive must stay in the folder: %+v", flat.Files)
	}

	deep := decodeMetaPreview(t, doMetaPreview(t, s, metadataRequest{
		Path: dir, Recursive: true, Fields: bulkMetaFields{Series: strp("Berserk")},
	}))
	if len(deep.Files) != 2 {
		t.Fatalf("recursive must descend: %+v", deep.Files)
	}
	metaRowByName(t, deep.Files, "Omake.cbz")
	if deep.Truncated {
		t.Errorf("two files must not truncate")
	}
}

// A preview that stopped short must say so: reading it as full coverage is how
// a user applies an edit believing every file was accounted for.
func TestMetadataPreviewReportsTruncation(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Big")
	for i := 0; i < metadataMaxFiles+5; i++ {
		mustWriteN(t, filepath.Join(dir, fmt.Sprintf("v%04d.pdf", i)), 1)
	}

	out := decodeMetaPreview(t, doMetaPreview(t, s, metadataRequest{
		Path: dir, Fields: bulkMetaFields{Series: strp("Big")},
	}))
	if !out.Truncated {
		t.Fatal("truncated must be true past the bound")
	}
	if len(out.Files) != metadataMaxFiles || out.Limit != metadataMaxFiles {
		t.Errorf("inspected %d rows (limit %d), want %d", len(out.Files), out.Limit, metadataMaxFiles)
	}
	found := false
	for _, n := range out.Notices {
		if n.Key == metaNoticeTruncated && n.Count == metadataMaxFiles {
			found = true
		}
	}
	if !found {
		t.Errorf("truncation must surface as a notice: %+v", out.Notices)
	}
}

// --- apply ---------------------------------------------------------------

func TestMetadataApplyWritesAndReadsBack(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Berserk")
	cbz := filepath.Join(dir, "Vol.01.cbz")
	epub := filepath.Join(dir, "Vol.02.epub")
	writeCBZ(t, cbz, `<?xml version="1.0"?><ComicInfo><Series>Old</Series></ComicInfo>`)
	writeEPUB(t, epub, "Old Title")

	res := decodeMetaApply(t, doMetaApply(t, s, metadataRequest{
		Path: dir,
		Fields: bulkMetaFields{
			Series:    strp("Berserk"),
			Publisher: strp("Dark Horse"),
			Genres:    &[]string{"Dark Fantasy"},
		},
	}))
	if res.Written != 2 || len(res.Errors) != 0 {
		t.Fatalf("apply = %+v", res)
	}

	for _, path := range []string{cbz, epub} {
		got, err := filewriter.Read(path)
		if err != nil {
			t.Fatalf("%s: read back: %v", path, err)
		}
		if got.Series == nil || *got.Series != "Berserk" {
			t.Errorf("%s: series = %v", path, got.Series)
		}
		if got.Publisher == nil || *got.Publisher != "Dark Horse" {
			t.Errorf("%s: publisher = %v", path, got.Publisher)
		}
		if len(got.Genres) != 1 || got.Genres[0] != "Dark Fantasy" {
			t.Errorf("%s: genres = %v", path, got.Genres)
		}
	}
}

// One unreadable archive must not cost the others their edit.
func TestMetadataApplyCollectsErrorsWithoutAbortingTheBatch(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Berserk")
	good := filepath.Join(dir, "Vol.01.cbz")
	writeCBZ(t, good, "")
	// Not a zip at all: filewriter fails to open it.
	mustWriteN(t, filepath.Join(dir, "Vol.02.cbz"), 32)

	res := decodeMetaApply(t, doMetaApply(t, s, metadataRequest{
		Path: dir, Fields: bulkMetaFields{Series: strp("Berserk")},
	}))
	if res.Written != 1 {
		t.Errorf("written = %d, want 1", res.Written)
	}
	if len(res.Errors) != 1 {
		t.Fatalf("errors = %v, want exactly the broken file", res.Errors)
	}
	got, err := filewriter.Read(good)
	if err != nil || got.Series == nil || *got.Series != "Berserk" {
		t.Errorf("the healthy file must still have been written: %v %v", got.Series, err)
	}
}

// A metadata write changes size and mtime, both of which the listing cache
// carries — the folder must be invalidated or the UI shows stale rows.
func TestMetadataApplyInvalidatesListingCache(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Berserk")
	cbz := filepath.Join(dir, "Vol.01.cbz")
	writeCBZ(t, cbz, "")

	before := decodeList(t, doList(t, s, dir))
	if len(before.Entries) != 1 {
		t.Fatalf("listing = %+v", before.Entries)
	}
	sizeBefore := before.Entries[0].Size

	doMetaApply(t, s, metadataRequest{Path: dir, Fields: bulkMetaFields{
		Summary: strp("A long summary that certainly grows the archive"),
	}})

	after := decodeList(t, doList(t, s, dir))
	if after.Entries[0].Size == sizeBefore {
		t.Errorf("listing still reports the pre-write size %d — cache not invalidated", sizeBefore)
	}
}

// --- confinement ---------------------------------------------------------

func TestMetadataWriteRejectsTraversalAndRoots(t *testing.T) {
	s, manga := newBrowseFixture(t)
	writeCBZ(t, filepath.Join(manga, "Berserk", "Vol.01.cbz"), "")
	fields := bulkMetaFields{Series: strp("x")}

	cases := []struct {
		name string
		path string
		want int
	}{
		{"bare root", manga, http.StatusForbidden},
		{"outside", "/etc", http.StatusForbidden},
		{"traversal", filepath.Join(manga, "..", ".."), http.StatusForbidden},
		{"relative", "manga/Berserk", http.StatusBadRequest},
		{"empty", "", http.StatusBadRequest},
	}
	for _, tc := range cases {
		if rec := doMetaPreview(t, s, metadataRequest{Path: tc.path, Fields: fields}); rec.Code != tc.want {
			t.Errorf("%s preview: status = %d, want %d (%s)", tc.name, rec.Code, tc.want, rec.Body.String())
		}
		if rec := doMetaApply(t, s, metadataRequest{Path: tc.path, Fields: fields}); rec.Code != tc.want {
			t.Errorf("%s apply: status = %d, want %d (%s)", tc.name, rec.Code, tc.want, rec.Body.String())
		}
	}
}

func TestMetadataRejectsEmptyEdit(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Berserk")
	writeCBZ(t, filepath.Join(dir, "Vol.01.cbz"), "")

	if rec := doMetaPreview(t, s, metadataRequest{Path: dir}); rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

// Append mode must preview what would actually be stored, values kept
// included — otherwise the target column hides half the result.
func TestMetadataPreviewAppendShowsMergedTarget(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Berserk")
	writeCBZ(t, filepath.Join(dir, "Vol.01.cbz"), `<?xml version="1.0"?><ComicInfo><Genre>Action</Genre></ComicInfo>`)

	out := decodeMetaPreview(t, doMetaPreview(t, s, metadataRequest{
		Path: dir, Mode: "append", Fields: bulkMetaFields{Genres: &[]string{"Dark Fantasy"}},
	}))
	row := metaRowByName(t, out.Files, "Vol.01.cbz")
	c := changeByField(t, row, "genres")
	if c.Current != "Action" || c.Target != "Action, Dark Fantasy" {
		t.Errorf("append preview = %+v", c)
	}
}
