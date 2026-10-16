package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func doMkdir(t *testing.T, s *Server, body any) *httptest.ResponseRecorder {
	t.Helper()
	return postJSON(t, s.handleFilesMkdir, "/api/files/mkdir", body)
}

func decodeMkdir(t *testing.T, rec *httptest.ResponseRecorder) mkdirResponse {
	t.Helper()
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var out mkdirResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v (body %s)", err, rec.Body.String())
	}
	return out
}

func TestMkdirCreatesFolderInSubdirectory(t *testing.T) {
	s, manga := newBrowseFixture(t)
	parent := filepath.Join(manga, "Existing")
	mustMkdir(t, parent)

	// Prime the cache so the test proves invalidation, not just creation.
	decodeList(t, doList(t, s, parent))

	out := decodeMkdir(t, doMkdir(t, s, mkdirRequest{Path: parent, Name: "New Series"}))
	if out.Name != "New Series" || out.Path != filepath.Join(parent, "New Series") {
		t.Fatalf("unexpected response: %+v", out)
	}
	if info, err := os.Stat(filepath.Join(parent, "New Series")); err != nil || !info.IsDir() {
		t.Fatalf("folder was not created on disk: %v", err)
	}
	// The parent listing must show the new folder without waiting for the TTL.
	listed := decodeList(t, doList(t, s, parent))
	found := false
	for _, e := range listed.Entries {
		if e.Name == "New Series" && e.IsDir {
			found = true
		}
	}
	if !found {
		t.Errorf("new folder missing from parent listing: %+v", listed.Entries)
	}
}

func TestMkdirCreatesFolderDirectlyUnderRoot(t *testing.T) {
	s, manga := newBrowseFixture(t)

	out := decodeMkdir(t, doMkdir(t, s, mkdirRequest{Path: manga, Name: "Fresh"}))
	if out.Path != filepath.Join(manga, "Fresh") {
		t.Fatalf("unexpected path: %+v", out)
	}
	if info, err := os.Stat(filepath.Join(manga, "Fresh")); err != nil || !info.IsDir() {
		t.Fatalf("folder was not created under the root: %v", err)
	}
}

func TestMkdirSanitizesName(t *testing.T) {
	s, manga := newBrowseFixture(t)

	out := decodeMkdir(t, doMkdir(t, s, mkdirRequest{Path: manga, Name: "  Series: One  "}))
	if out.Name != "Series One" {
		t.Fatalf("name was not sanitized, got %q", out.Name)
	}
	if _, err := os.Stat(filepath.Join(manga, "Series One")); err != nil {
		t.Fatalf("sanitized folder missing on disk: %v", err)
	}
}

func TestMkdirRejectsInvalidRequests(t *testing.T) {
	s, manga := newBrowseFixture(t)
	parent := filepath.Join(manga, "Existing")
	mustMkdir(t, parent)
	mustWriteN(t, filepath.Join(parent, "file.cbz"), 1)

	cases := []struct {
		name string
		body mkdirRequest
		want int
	}{
		{"empty name", mkdirRequest{Path: parent, Name: "   "}, http.StatusBadRequest},
		{"only dots", mkdirRequest{Path: parent, Name: "..."}, http.StatusBadRequest},
		{"only separators", mkdirRequest{Path: parent, Name: "///"}, http.StatusBadRequest},
		{"missing parent", mkdirRequest{Path: "", Name: "New"}, http.StatusBadRequest},
		{"relative parent", mkdirRequest{Path: "manga/Existing", Name: "New"}, http.StatusBadRequest},
		{"traversal parent", mkdirRequest{Path: filepath.Join(manga, "..", "..", "etc"), Name: "New"}, http.StatusForbidden},
		{"outside libraries", mkdirRequest{Path: "/etc", Name: "New"}, http.StatusForbidden},
		{"parent not found", mkdirRequest{Path: filepath.Join(manga, "nope"), Name: "New"}, http.StatusNotFound},
		{"parent is file", mkdirRequest{Path: filepath.Join(parent, "file.cbz"), Name: "New"}, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := doMkdir(t, s, tc.body).Code; got != tc.want {
				t.Errorf("status = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestMkdirConflictWhenFolderExists(t *testing.T) {
	s, manga := newBrowseFixture(t)
	mustMkdir(t, filepath.Join(manga, "Taken"))

	rec := doMkdir(t, s, mkdirRequest{Path: manga, Name: "Taken"})
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409 (body %s)", rec.Code, rec.Body.String())
	}

	// A name that sanitizes onto the same segment conflicts too.
	rec = doMkdir(t, s, mkdirRequest{Path: manga, Name: "  Taken  "})
	if rec.Code != http.StatusConflict {
		t.Fatalf("sanitized conflict: status = %d, want 409 (body %s)", rec.Code, rec.Body.String())
	}
}
