package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JeremiahM37/librarr/internal/config"
)

// postScan drives the handler with a raw JSON body.
func postScan(t *testing.T, s *Server, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/files/kavita/scan",
		bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	s.handleFilesKavitaScan(rr, req)
	return rr
}

func TestFilesKavitaScanRequiresKavita(t *testing.T) {
	cfg, _, _ := newRootsFixture(t)
	rr := postScan(t, &Server{cfg: cfg}, `{"root":"manga"}`)
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("HTTP %d, want 503", rr.Code)
	}
}

func TestFilesKavitaScanRejectsUnknownRoot(t *testing.T) {
	cfg, _, _ := newRootsFixture(t)
	cfg.KavitaURL, cfg.KavitaUser, cfg.KavitaPass = "http://kavita.invalid", "u", "p"
	s := &Server{cfg: cfg}

	if rr := postScan(t, s, `{"root":""}`); rr.Code != http.StatusBadRequest {
		t.Errorf("empty root: HTTP %d, want 400", rr.Code)
	}
	if rr := postScan(t, s, `not json`); rr.Code != http.StatusBadRequest {
		t.Errorf("bad body: HTTP %d, want 400", rr.Code)
	}
	// audiobooks is configured but its directory does not exist, so it is not a
	// browsable library and must be refused like any unknown id.
	if rr := postScan(t, s, `{"root":"audiobooks"}`); rr.Code != http.StatusNotFound {
		t.Errorf("missing root dir: HTTP %d, want 404", rr.Code)
	}
	if rr := postScan(t, s, `{"root":"comics"}`); rr.Code != http.StatusNotFound {
		t.Errorf("unknown root: HTTP %d, want 404", rr.Code)
	}
}

func TestFilesKavitaScanTriggersScan(t *testing.T) {
	cfg, _, _ := newRootsFixture(t)
	var scanned int
	kav := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/api/Account/login"):
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"token":"t"}`))
		case strings.HasSuffix(r.URL.Path, "/api/Library/scan"):
			scanned++
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer kav.Close()
	cfg.KavitaURL, cfg.KavitaUser, cfg.KavitaPass = kav.URL, "u", "p"

	rr := postScan(t, &Server{cfg: cfg}, `{"root":"manga"}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("HTTP %d, want 200: %s", rr.Code, rr.Body.String())
	}
	if scanned != 1 {
		t.Errorf("kavita scan called %d times, want 1", scanned)
	}
	if !strings.Contains(rr.Body.String(), `"scope":"all"`) {
		t.Errorf("response should state the scan scope, got %s", rr.Body.String())
	}
}

func TestFilesKavitaScanReportsKavitaFailure(t *testing.T) {
	cfg, _, _ := newRootsFixture(t)
	kav := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer kav.Close()
	cfg.KavitaURL, cfg.KavitaUser, cfg.KavitaPass = kav.URL, "u", "p"

	rr := postScan(t, &Server{cfg: cfg}, `{"root":"manga"}`)
	if rr.Code != http.StatusBadGateway {
		t.Fatalf("HTTP %d, want 502", rr.Code)
	}
}

func TestPathsOverlap(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"/books/manga", "/books/manga", true},
		{"/books/manga/", "/books/manga", true},
		{"/books/manga", "/books", true},
		{"/books", "/books/manga", true},
		{"/books/manga", "/books/manga-other", false},
		{"/books/manga", "/media/manga", false},
		{"", "/books", false},
	}
	for _, c := range cases {
		if got := pathsOverlap(c.a, c.b); got != c.want {
			t.Errorf("pathsOverlap(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

// WarnLibraryRootDivergence only logs, so the assertions are on the decision it
// takes rather than on the output: the point is that it must not fire when
// there is no second view to diverge from.
func TestWarnLibraryRootDivergenceStaysQuietWithoutKavita(t *testing.T) {
	cfg, _, _ := newRootsFixture(t)
	if divergenceDetected(cfg) {
		t.Error("no Kavita configured: nothing to warn about")
	}
}

func TestWarnLibraryRootDivergenceDetectsDisjointRoots(t *testing.T) {
	cfg, _, _ := newRootsFixture(t)
	cfg.KavitaURL, cfg.KavitaUser, cfg.KavitaPass = "http://kavita.invalid", "u", "p"

	cfg.KavitaLibraryPath = "/somewhere/else"
	if !divergenceDetected(cfg) {
		t.Error("disjoint roots should be reported")
	}

	cfg.KavitaLibraryPath = cfg.MangaDir
	if divergenceDetected(cfg) {
		t.Error("overlapping roots should not be reported")
	}
}

func divergenceDetected(cfg *config.Config) bool {
	_, _, diverge := libraryRootsDiverge(cfg)
	return diverge
}
