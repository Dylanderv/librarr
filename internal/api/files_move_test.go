package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func doMove(t *testing.T, s *Server, body any) *httptest.ResponseRecorder {
	t.Helper()
	return postJSON(t, s.handleFilesMove, "/api/files/move", body)
}

// Moving a loose file from a library root into a subfolder is the core
// gesture: a file the downloader dropped at the root joins its series folder.
func TestMoveFileIntoSubfolder(t *testing.T) {
	s, manga := newBrowseFixture(t)
	sub := filepath.Join(manga, "Series")
	mustMkdir(t, sub)
	src := filepath.Join(manga, "loose.cbz")
	mustWriteN(t, src, 7)

	// Prime both listings so the test proves invalidation, not just the move.
	decodeList(t, doList(t, s, manga))
	decodeList(t, doList(t, s, sub))

	rec := doMove(t, s, moveRequest{Path: src, Dest: sub})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(sub, "loose.cbz")); err != nil {
		t.Fatalf("file missing at destination: %v", err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("source must be gone after the move")
	}
	for _, dir := range []string{manga, sub} {
		listed := decodeList(t, doList(t, s, dir))
		for _, e := range listed.Entries {
			if dir == manga && e.Name == "loose.cbz" && !e.IsDir {
				t.Error("source listing still shows the moved file")
			}
			if dir == sub && e.Name == "loose.cbz" && e.IsDir {
				t.Error("destination listing shows the file as a directory")
			}
		}
	}
}

func TestMoveRejectsBadRequests(t *testing.T) {
	s, manga := newBrowseFixture(t)
	sub := filepath.Join(manga, "Series")
	mustMkdir(t, sub)
	mustWriteN(t, filepath.Join(manga, "loose.cbz"), 3)
	mustWriteN(t, filepath.Join(sub, "taken.cbz"), 3)
	mustWriteN(t, filepath.Join(manga, "taken.cbz"), 3)

	cases := []struct {
		name string
		body moveRequest
		want int
	}{
		{"missing path", moveRequest{Dest: sub}, http.StatusBadRequest},
		{"missing dest", moveRequest{Path: filepath.Join(manga, "loose.cbz")}, http.StatusBadRequest},
		{"source is a directory", moveRequest{Path: sub, Dest: manga}, http.StatusBadRequest},
		{"same folder", moveRequest{Path: filepath.Join(sub, "taken.cbz"), Dest: sub}, http.StatusBadRequest},
		{"dest is a file", moveRequest{Path: filepath.Join(manga, "loose.cbz"), Dest: filepath.Join(manga, "loose.cbz")}, http.StatusBadRequest},
		{"dest missing", moveRequest{Path: filepath.Join(manga, "loose.cbz"), Dest: filepath.Join(manga, "nope")}, http.StatusNotFound},
		{"source missing", moveRequest{Path: filepath.Join(manga, "nope.cbz"), Dest: sub}, http.StatusNotFound},
		{"source traversal", moveRequest{Path: filepath.Join(manga, "..", "x.cbz"), Dest: sub}, http.StatusForbidden},
		{"dest traversal", moveRequest{Path: filepath.Join(manga, "loose.cbz"), Dest: filepath.Join(manga, "..", "x")}, http.StatusForbidden},
		{"outside", moveRequest{Path: "/etc/passwd", Dest: sub}, http.StatusForbidden},
		{"target taken", moveRequest{Path: filepath.Join(manga, "taken.cbz"), Dest: sub}, http.StatusConflict},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := doMove(t, s, tc.body).Code; got != tc.want {
				t.Errorf("status = %d, want %d", got, tc.want)
			}
		})
	}
}
