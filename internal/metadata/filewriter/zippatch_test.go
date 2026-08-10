package filewriter

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// entryNames lists the archive's entries in central-directory order.
func entryNames(t *testing.T, path string) []string {
	t.Helper()
	zr, err := zip.OpenReader(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer zr.Close()
	names := make([]string, 0, len(zr.File))
	for _, f := range zr.File {
		names = append(names, f.Name)
	}
	return names
}

func TestPatchZipEntryAddsEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "book.cbz")
	writeZip(t, path, map[string]string{"001.jpg": "page one", "002.jpg": "page two"}, false)

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := patchZipEntry(path, comicInfoName, []byte("<ComicInfo/>")); err != nil {
		t.Fatalf("patch: %v", err)
	}

	// The original bytes must survive untouched — that is what makes an
	// interrupted patch recoverable.
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(after, before) {
		t.Error("patch modified existing bytes; it must only append")
	}

	if got := readEntry(t, path, comicInfoName); got != "<ComicInfo/>" {
		t.Errorf("ComicInfo content = %q", got)
	}
	// Pre-existing entries stay readable, which proves their central directory
	// records still point at valid local headers.
	if got := readEntry(t, path, "001.jpg"); got != "page one" {
		t.Errorf("001.jpg = %q, want %q", got, "page one")
	}
	if got := readEntry(t, path, "002.jpg"); got != "page two" {
		t.Errorf("002.jpg = %q, want %q", got, "page two")
	}
	if names := entryNames(t, path); len(names) != 3 {
		t.Errorf("entries = %v, want 3", names)
	}
}

func TestPatchZipEntryReplacesExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "book.cbz")
	writeZip(t, path, map[string]string{
		"001.jpg":     "page one",
		comicInfoName: "<ComicInfo><Series>old</Series></ComicInfo>",
	}, false)

	if err := patchZipEntry(path, comicInfoName, []byte("<ComicInfo><Series>new</Series></ComicInfo>")); err != nil {
		t.Fatalf("patch: %v", err)
	}

	names := entryNames(t, path)
	if len(names) != 2 {
		t.Fatalf("entries = %v, want 2 (the superseded record must be dropped)", names)
	}
	if got := readEntry(t, path, comicInfoName); got != "<ComicInfo><Series>new</Series></ComicInfo>" {
		t.Errorf("ComicInfo content = %q", got)
	}
}

func TestPatchZipEntryRepeatable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "book.cbz")
	writeZip(t, path, map[string]string{"001.jpg": "page one"}, false)

	for i, want := range []string{"<a/>", "<b/>", "<c/>"} {
		if err := patchZipEntry(path, comicInfoName, []byte(want)); err != nil {
			t.Fatalf("patch %d: %v", i, err)
		}
		if got := readEntry(t, path, comicInfoName); got != want {
			t.Fatalf("after patch %d: content = %q, want %q", i, got, want)
		}
		if names := entryNames(t, path); len(names) != 2 {
			t.Fatalf("after patch %d: entries = %v, want 2", i, names)
		}
	}
}

func TestPatchZipEntryPreservesCompressedEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "book.cbz")

	// Deflated content, as a real CBZ has: the patch must not disturb it.
	body := bytes.Repeat([]byte("compress me "), 4096)
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, err := zw.CreateHeader(&zip.FileHeader{Name: "001.jpg", Method: zip.Deflate})
	if err != nil {
		t.Fatal(err)
	}
	w.Write(body)
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	f.Close()

	if err := patchZipEntry(path, comicInfoName, []byte("<ComicInfo/>")); err != nil {
		t.Fatalf("patch: %v", err)
	}
	if got := readEntry(t, path, "001.jpg"); got != string(body) {
		t.Error("deflated entry did not survive the patch intact")
	}
}

func TestPatchZipEntryRejectsArchiveWithComment(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "book.cbz")

	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	w, _ := zw.Create("001.jpg")
	w.Write([]byte("page one"))
	// A trailing comment would be buried by an append, so the patch must decline
	// and let the caller rebuild instead.
	if err := zw.SetComment("not patchable"); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	f.Close()

	err = patchZipEntry(path, comicInfoName, []byte("<ComicInfo/>"))
	if !errors.Is(err, errNotPatchable) {
		t.Fatalf("err = %v, want errNotPatchable", err)
	}
}

func TestPatchZipEntryRejectsTruncatedCentralDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "book.cbz")
	writeZip(t, path, map[string]string{"001.jpg": "page one"}, false)

	// Point the EOCD's central-directory offset past the end of the file.
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	binary.LittleEndian.PutUint32(raw[len(raw)-eocdLen+16:], uint32(len(raw)))
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}

	err = patchZipEntry(path, comicInfoName, []byte("<ComicInfo/>"))
	if !errors.Is(err, errNotPatchable) {
		t.Fatalf("err = %v, want errNotPatchable", err)
	}
}

// TestApplyCBZUsesPatchPath checks the whole Apply path stays correct now that
// it patches instead of rebuilding, and that it leaves the original bytes alone.
func TestApplyCBZUsesPatchPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "book.cbz")
	writeZip(t, path, map[string]string{"001.jpg": "page one"}, false)

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := Apply(path, Fields{Series: ptr("Bakemono Recchan"), Genres: []string{"Drama"}}); err != nil {
		t.Fatalf("apply: %v", err)
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(after, before) {
		t.Error("Apply rebuilt the archive; it should have patched in place")
	}

	got := readEntry(t, path, comicInfoName)
	if !bytes.Contains([]byte(got), []byte("<Series>Bakemono Recchan</Series>")) {
		t.Errorf("ComicInfo missing Series: %s", got)
	}
	if !bytes.Contains([]byte(got), []byte("<Genre>Drama</Genre>")) {
		t.Errorf("ComicInfo missing Genre: %s", got)
	}
}
