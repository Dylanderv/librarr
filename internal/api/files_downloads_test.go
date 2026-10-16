package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JeremiahM37/librarr/internal/config"
)

func browseServerWithSidecar(t *testing.T, handler http.HandlerFunc) *Server {
	t.Helper()
	resetDownloadsCache()
	t.Cleanup(resetDownloadsCache)
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	cfg := &config.Config{MangaDir: t.TempDir(), TelegramSidecarURL: srv.URL}
	return &Server{cfg: cfg}
}

func doDownloads(t *testing.T, s *Server) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/files/downloads", nil)
	rec := httptest.NewRecorder()
	s.handleFilesDownloads(rec, req)
	return rec
}

func decodeDownloads(t *testing.T, rec *httptest.ResponseRecorder) []activeDownload {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var out []activeDownload
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v (body %s)", err, rec.Body.String())
	}
	if out == nil {
		out = []activeDownload{}
	}
	return out
}

func TestDownloadsListsInflight(t *testing.T) {
	s := browseServerWithSidecar(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/downloads" {
			t.Errorf("unexpected sidecar path %s", r.URL.Path)
		}
		if r.Header.Get("X-API-Key") != "" {
			t.Errorf("no key configured, none must be sent")
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"filename": "a.cbz", "progress": 0.42, "status": "downloading"},
			{"filename": "b.cbz", "progress": 0.0, "status": "queued"},
			{"filename": "c.cbz", "progress": 1.0, "status": "done"},
			{"filename": nil, "progress": 0.1, "status": "downloading"},
			{"filename": "e.cbz", "progress": 0.9, "status": "error"},
		})
	})

	out := decodeDownloads(t, doDownloads(t, s))
	if len(out) != 2 || out[0].Name != "a.cbz" || out[0].Progress != 0.42 || out[1].Name != "b.cbz" {
		t.Fatalf("unexpected active downloads: %+v", out)
	}
}

func TestDownloadsSendsAPIKey(t *testing.T) {
	resetDownloadsCache()
	t.Cleanup(resetDownloadsCache)
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("X-API-Key")
		_, _ = w.Write([]byte("[]"))
	}))
	t.Cleanup(srv.Close)
	s := &Server{cfg: &config.Config{MangaDir: t.TempDir(), TelegramSidecarURL: srv.URL, TelegramSidecarAPIKey: "s3cret"}}

	decodeDownloads(t, doDownloads(t, s))
	if got != "s3cret" {
		t.Errorf("api key sent = %q, want s3cret", got)
	}
}

func TestDownloadsSidecarDownServesEmpty(t *testing.T) {
	resetDownloadsCache()
	t.Cleanup(resetDownloadsCache)
	s := &Server{cfg: &config.Config{MangaDir: t.TempDir(), TelegramSidecarURL: "http://127.0.0.1:1"}}
	start := time.Now()
	if out := decodeDownloads(t, doDownloads(t, s)); len(out) != 0 {
		t.Fatalf("expected empty, got %+v", out)
	}
	// The backoff must make the second call instant, not another dial.
	if out := decodeDownloads(t, doDownloads(t, s)); len(out) != 0 {
		t.Fatalf("expected empty, got %+v", out)
	}
	if time.Since(start) > 10*time.Second {
		t.Error("backoff did not kick in: two dials took too long")
	}
}

func TestDownloadsWithoutConfiguredURL(t *testing.T) {
	resetDownloadsCache()
	t.Cleanup(resetDownloadsCache)
	s := &Server{cfg: &config.Config{}}
	if out := decodeDownloads(t, doDownloads(t, s)); len(out) != 0 {
		t.Fatalf("expected empty, got %+v", out)
	}
}
