package api

import (
	"encoding/json"
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

func decodeBatchMove(t *testing.T, rec *httptest.ResponseRecorder) batchMoveResponse {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var out batchMoveResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v (body %s)", err, rec.Body.String())
	}
	return out
}

func TestMoveBatchIntoSubfolder(t *testing.T) {
	s, manga := newBrowseFixture(t)
	sub := filepath.Join(manga, "Series")
	mustMkdir(t, sub)
	mustWriteN(t, filepath.Join(manga, "a.cbz"), 1)
	mustWriteN(t, filepath.Join(manga, "b.cbz"), 2)

	decodeList(t, doList(t, s, manga))
	decodeList(t, doList(t, s, sub))

	out := decodeBatchMove(t, doMove(t, s, moveRequest{
		Sources: []string{filepath.Join(manga, "a.cbz"), filepath.Join(manga, "b.cbz")},
		Dest:    sub,
	}))
	if len(out.Moved) != 2 || len(out.Skipped) != 0 {
		t.Fatalf("unexpected response: %+v", out)
	}
	for _, name := range []string{"a.cbz", "b.cbz"} {
		if _, err := os.Stat(filepath.Join(sub, name)); err != nil {
			t.Errorf("%s missing at destination: %v", name, err)
		}
		if _, err := os.Stat(filepath.Join(manga, name)); !os.IsNotExist(err) {
			t.Errorf("%s source must be gone", name)
		}
	}
	// Both listings reflect the swap without waiting for the TTL.
	for _, e := range decodeList(t, doList(t, s, manga)).Entries {
		if e.Name == "a.cbz" || e.Name == "b.cbz" {
			t.Errorf("source listing still shows %s", e.Name)
		}
	}
	if n := len(decodeList(t, doList(t, s, sub)).Entries); n != 2 {
		t.Errorf("destination listing has %d entries, want 2", n)
	}
}

func TestMoveBatchReportsSkips(t *testing.T) {
	s, manga := newBrowseFixture(t)
	sub := filepath.Join(manga, "Series")
	mustMkdir(t, sub)
	mustWriteN(t, filepath.Join(manga, "ok.cbz"), 1)
	mustWriteN(t, filepath.Join(manga, "taken.cbz"), 1)
	mustWriteN(t, filepath.Join(sub, "taken.cbz"), 1)

	out := decodeBatchMove(t, doMove(t, s, moveRequest{
		Sources: []string{
			filepath.Join(manga, "ok.cbz"),
			filepath.Join(manga, "taken.cbz"),
			filepath.Join(manga, "gone.cbz"),
			sub,
		},
		Dest: sub,
	}))
	if len(out.Moved) != 1 || out.Moved[0].Name != "ok.cbz" {
		t.Fatalf("moved = %+v, want [ok.cbz]", out.Moved)
	}
	skipped := map[string]string{}
	for _, sk := range out.Skipped {
		skipped[sk.Name] = sk.Reason
	}
	if skipped["taken.cbz"] != moveSkipExists ||
		skipped["gone.cbz"] != moveSkipMissing ||
		skipped["Series"] != moveSkipNotFile {
		t.Errorf("skipped = %+v", out.Skipped)
	}
	// The collision loser stays where it was.
	if _, err := os.Stat(filepath.Join(manga, "taken.cbz")); err != nil {
		t.Error("colliding source must survive the batch")
	}
}

func TestMoveBatchRejectsBadRequests(t *testing.T) {
	s, manga := newBrowseFixture(t)
	sub := filepath.Join(manga, "Series")
	mustMkdir(t, sub)
	mustWriteN(t, filepath.Join(manga, "a.cbz"), 1)

	cases := []struct {
		name string
		body moveRequest
		want int
	}{
		{"no sources", moveRequest{Dest: sub}, http.StatusBadRequest},
		{"no dest", moveRequest{Sources: []string{filepath.Join(manga, "a.cbz")}}, http.StatusBadRequest},
		{"dest missing", moveRequest{Sources: []string{filepath.Join(manga, "a.cbz")}, Dest: filepath.Join(manga, "nope")}, http.StatusNotFound},
		{"source traversal aborts all", moveRequest{Sources: []string{filepath.Join(manga, "a.cbz"), filepath.Join(manga, "..", "x.cbz")}, Dest: sub}, http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := doMove(t, s, tc.body).Code; got != tc.want {
				t.Errorf("status = %d, want %d", got, tc.want)
			}
		})
	}
	// Nothing moved when the batch aborts.
	if _, err := os.Stat(filepath.Join(manga, "a.cbz")); err != nil {
		t.Error("aborted batch must move nothing")
	}
}
