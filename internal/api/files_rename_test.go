package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// postJSON drives one of the rename handlers with a JSON body.
func postJSON(t *testing.T, h http.HandlerFunc, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(string(raw)))
	rec := httptest.NewRecorder()
	h(rec, req)
	return rec
}

func doRenamePreview(t *testing.T, s *Server, body any) *httptest.ResponseRecorder {
	t.Helper()
	return postJSON(t, s.handleFilesRenamePreview, "/api/files/rename/preview", body)
}

func doRenameApply(t *testing.T, s *Server, body any) *httptest.ResponseRecorder {
	t.Helper()
	return postJSON(t, s.handleFilesRenameApply, "/api/files/rename/apply", body)
}

func decodePreview(t *testing.T, rec *httptest.ResponseRecorder) renamePreviewResponse {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var out renamePreviewResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v (body %s)", err, rec.Body.String())
	}
	return out
}

func rowByName(t *testing.T, rows []renameFileRow, name string) renameFileRow {
	t.Helper()
	for _, r := range rows {
		if r.Name == name {
			return r
		}
	}
	t.Fatalf("no preview row for %q in %+v", name, rows)
	return renameFileRow{}
}

func TestRenamePreviewListsEveryFile(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Old.Torrent.Name-GROUP")
	for _, n := range []string{"Old.Name v1.cbz", "bonus chapter.cbz", "Berserk Vol.02.pdf"} {
		mustWriteN(t, filepath.Join(dir, n), 1)
	}

	out := decodePreview(t, doRenamePreview(t, s, renameRequest{
		Path: dir, SeriesName: "Berserk", Mode: "folder_and_files",
	}))

	if out.FolderTo != filepath.Join(manga, "Berserk") {
		t.Errorf("FolderTo = %q", out.FolderTo)
	}
	// Every file gets a row, including the one that cannot be renamed.
	if len(out.Files) != 3 {
		t.Fatalf("got %d rows, want 3: %+v", len(out.Files), out.Files)
	}

	if r := rowByName(t, out.Files, "Old.Name v1.cbz"); r.Status != renameRowRename ||
		r.Target != "Berserk Vol.01.cbz" || r.Volume != 1 || r.VolumeSource != volumeSourceDetected {
		t.Errorf("unexpected row: %+v", r)
	}
	if r := rowByName(t, out.Files, "bonus chapter.cbz"); r.Status != renameRowNoVolume ||
		r.Target != "" || r.VolumeSource != volumeSourceNone {
		t.Errorf("an undetectable file must be reported, not guessed: %+v", r)
	}
	if r := rowByName(t, out.Files, "Berserk Vol.02.pdf"); r.Status != renameRowUnchanged {
		t.Errorf("an already-correct file must read as unchanged: %+v", r)
	}
}

func TestRenamePreviewManualVolumeWins(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Old")
	// The trailing-number fallback reads this as volume 42.
	mustWriteN(t, filepath.Join(dir, "Serie Truc 42.cbz"), 1)
	mustWriteN(t, filepath.Join(dir, "bonus.cbz"), 1)

	out := decodePreview(t, doRenamePreview(t, s, renameRequest{
		Path: dir, SeriesName: "Series", Mode: "folder_and_files",
		Volumes: map[string]int{"Serie Truc 42.cbz": 3, "bonus.cbz": 4},
	}))

	if r := rowByName(t, out.Files, "Serie Truc 42.cbz"); r.Volume != 3 ||
		r.VolumeSource != volumeSourceManual || r.Target != "Series Vol.03.cbz" {
		t.Errorf("manual volume must override detection: %+v", r)
	}
	if r := rowByName(t, out.Files, "bonus.cbz"); r.Volume != 4 || r.Target != "Series Vol.04.cbz" {
		t.Errorf("manual volume must rescue an undetectable file: %+v", r)
	}
}

func TestRenamePreviewFolderModeLeavesFilesAlone(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Old")
	mustWriteN(t, filepath.Join(dir, "Old v1.cbz"), 1)

	out := decodePreview(t, doRenamePreview(t, s, renameRequest{
		Path: dir, SeriesName: "Series", Mode: "folder",
	}))
	if out.FolderTo == "" {
		t.Error("folder mode must still plan the folder rename")
	}
	// The volume is still shown, so switching mode needs no second thought.
	if r := rowByName(t, out.Files, "Old v1.cbz"); r.Status != renameRowFolderOnly ||
		r.Volume != 1 || r.Target != "" {
		t.Errorf("unexpected row: %+v", r)
	}
}

func TestRenameApplyRenamesOnDisk(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Old.Torrent.Name-GROUP")
	mustWriteN(t, filepath.Join(dir, "Old.Name v1.cbz"), 1)
	mustWriteN(t, filepath.Join(dir, "bonus.cbz"), 1)

	rec := doRenameApply(t, s, renameRequest{
		Path: dir, SeriesName: "Berserk", Mode: "folder_and_files",
		Volumes: map[string]int{"bonus.cbz": 2},
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var out renameApplyResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Errors) != 0 {
		t.Fatalf("errors: %v", out.Errors)
	}
	// One folder plus two files.
	if out.Renamed != 3 {
		t.Errorf("renamed = %d, want 3", out.Renamed)
	}

	newDir := filepath.Join(manga, "Berserk")
	if out.Path != newDir {
		t.Errorf("Path = %q, want the folder's new location %q", out.Path, newDir)
	}
	for _, want := range []string{"Berserk Vol.01.cbz", "Berserk Vol.02.cbz"} {
		if _, err := os.Stat(filepath.Join(newDir, want)); err != nil {
			t.Errorf("missing %s: %v", want, err)
		}
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("the old folder still exists")
	}
}

func TestRenameApplyInvalidatesListings(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Old")
	mustWriteN(t, filepath.Join(dir, "Old v1.cbz"), 1)

	// Prime the cache for both the library root and the series folder.
	decodeList(t, doList(t, s, manga))
	decodeList(t, doList(t, s, dir))

	if rec := doRenameApply(t, s, renameRequest{
		Path: dir, SeriesName: "Series", Mode: "folder_and_files",
	}); rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	// A stale listing here would show the pre-rename folder name for 30s.
	root := decodeList(t, doList(t, s, manga))
	if len(root.Entries) != 1 || root.Entries[0].Name != "Series" {
		t.Fatalf("root listing was not invalidated: %+v", root.Entries)
	}
	inner := decodeList(t, doList(t, s, filepath.Join(manga, "Series")))
	if len(inner.Entries) != 1 || inner.Entries[0].Name != "Series Vol.01.cbz" {
		t.Errorf("folder listing was not invalidated: %+v", inner.Entries)
	}
}

func TestRenameApplyRefusesExistingTarget(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Old")
	mustWriteN(t, filepath.Join(dir, "Old v1.cbz"), 1)
	mustWriteN(t, filepath.Join(manga, "Series", "keep.cbz"), 1)

	rec := doRenameApply(t, s, renameRequest{
		Path: dir, SeriesName: "Series", Mode: "folder_and_files",
	})
	var out renameApplyResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Errors) == 0 {
		t.Fatal("merging into an existing folder must be refused")
	}
	if _, err := os.Stat(filepath.Join(manga, "Series", "keep.cbz")); err != nil {
		t.Errorf("the existing folder was disturbed: %v", err)
	}
	// The folder move failed, so the file must have been renamed where it is.
	if _, err := os.Stat(filepath.Join(dir, "Series Vol.01.cbz")); err != nil {
		t.Errorf("file not renamed in place: %v", err)
	}
	if out.Path != dir {
		t.Errorf("Path = %q, want the unmoved folder %q", out.Path, dir)
	}
}

func TestRenameRejectsInvalidTargets(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Old")
	mustWriteN(t, filepath.Join(dir, "Old v1.cbz"), 1)

	cases := []struct {
		name string
		req  renameRequest
		want int
	}{
		// Renaming a library root would move the whole library.
		{"library root", renameRequest{Path: manga, SeriesName: "X", Mode: "folder"}, http.StatusForbidden},
		{"traversal", renameRequest{Path: filepath.Join(manga, "..", "..", "etc"), SeriesName: "X", Mode: "folder"}, http.StatusForbidden},
		{"outside", renameRequest{Path: "/etc", SeriesName: "X", Mode: "folder"}, http.StatusForbidden},
		{"relative", renameRequest{Path: "Old", SeriesName: "X", Mode: "folder"}, http.StatusBadRequest},
		{"missing series", renameRequest{Path: dir, Mode: "folder"}, http.StatusBadRequest},
		{"unknown mode", renameRequest{Path: dir, SeriesName: "X", Mode: "everything"}, http.StatusBadRequest},
		{"mode none", renameRequest{Path: dir, SeriesName: "X", Mode: "none"}, http.StatusBadRequest},
		{"not a directory", renameRequest{Path: filepath.Join(dir, "Old v1.cbz"), SeriesName: "X", Mode: "folder"}, http.StatusBadRequest},
		{"not found", renameRequest{Path: filepath.Join(manga, "nope"), SeriesName: "X", Mode: "folder"}, http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := doRenamePreview(t, s, tc.req).Code; got != tc.want {
				t.Errorf("preview status = %d, want %d", got, tc.want)
			}
			if got := doRenameApply(t, s, tc.req).Code; got != tc.want {
				t.Errorf("apply status = %d, want %d", got, tc.want)
			}
		})
	}
	// Nothing was touched by the refusals.
	if _, err := os.Stat(filepath.Join(dir, "Old v1.cbz")); err != nil {
		t.Errorf("a refused request modified the disk: %v", err)
	}
}

func doRenameFile(t *testing.T, s *Server, body any) *httptest.ResponseRecorder {
	t.Helper()
	return postJSON(t, s.handleFilesRenameFile, "/api/files/rename/file", body)
}

func TestRenameFileKeepsExtensionAndInvalidates(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Berserk")
	mustWriteN(t, filepath.Join(dir, "raw name.cbz"), 1)
	decodeList(t, doList(t, s, dir)) // prime the cache

	rec := doRenameFile(t, s, fileRenameRequest{
		Path: filepath.Join(dir, "raw name.cbz"), NewName: "Berserk Vol.01",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "Berserk Vol.01.cbz")); err != nil {
		t.Fatalf("file not renamed: %v", err)
	}
	out := decodeList(t, doList(t, s, dir))
	if len(out.Entries) != 1 || out.Entries[0].Name != "Berserk Vol.01.cbz" {
		t.Errorf("listing was not invalidated: %+v", out.Entries)
	}
}

func TestRenameFileNameHandling(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Berserk")

	cases := []struct {
		name    string
		newName string
		want    string
	}{
		// Retyping the original extension is accepted, not doubled.
		{"same extension", "Berserk Vol.02.cbz", "Berserk Vol.02.cbz"},
		{"other extension", "Berserk.exe", "Berserk.exe.cbz"},
		// Path separators and illegal characters are stripped, so a rename can
		// never escape its folder.
		{"traversal", "../../evil", "evil.cbz"},
		{"illegal characters", `a:b"c|d`, "abcd.cbz"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := filepath.Join(dir, "src.cbz")
			mustWriteN(t, src, 1)
			rec := doRenameFile(t, s, fileRenameRequest{Path: src, NewName: tc.newName})
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
			}
			var out struct {
				Name string `json:"name"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
				t.Fatal(err)
			}
			if out.Name != tc.want {
				t.Errorf("name = %q, want %q", out.Name, tc.want)
			}
			if _, err := os.Stat(filepath.Join(dir, tc.want)); err != nil {
				t.Errorf("missing %s: %v", tc.want, err)
			}
			if err := os.Remove(filepath.Join(dir, tc.want)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRenameFileRejections(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dir := filepath.Join(manga, "Berserk")
	mustWriteN(t, filepath.Join(dir, "a.cbz"), 1)
	mustWriteN(t, filepath.Join(dir, "b.cbz"), 1)

	cases := []struct {
		name string
		req  fileRenameRequest
		want int
	}{
		{"existing target", fileRenameRequest{Path: filepath.Join(dir, "a.cbz"), NewName: "b"}, http.StatusConflict},
		{"unchanged", fileRenameRequest{Path: filepath.Join(dir, "a.cbz"), NewName: "a.cbz"}, http.StatusBadRequest},
		{"empty name", fileRenameRequest{Path: filepath.Join(dir, "a.cbz"), NewName: "  "}, http.StatusBadRequest},
		{"a directory", fileRenameRequest{Path: dir, NewName: "other"}, http.StatusBadRequest},
		{"library root", fileRenameRequest{Path: manga, NewName: "other"}, http.StatusForbidden},
		{"outside", fileRenameRequest{Path: "/etc/passwd", NewName: "other"}, http.StatusForbidden},
		{"traversal", fileRenameRequest{Path: filepath.Join(dir, "..", "..", "etc", "passwd"), NewName: "other"}, http.StatusForbidden},
		{"not found", fileRenameRequest{Path: filepath.Join(dir, "nope.cbz"), NewName: "other"}, http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := doRenameFile(t, s, tc.req).Code; got != tc.want {
				t.Errorf("status = %d, want %d", got, tc.want)
			}
		})
	}
	for _, n := range []string{"a.cbz", "b.cbz"} {
		if _, err := os.Stat(filepath.Join(dir, n)); err != nil {
			t.Errorf("a refused request modified the disk: %v", err)
		}
	}
}
