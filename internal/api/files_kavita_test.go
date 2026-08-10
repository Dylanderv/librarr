package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JeremiahM37/librarr/internal/config"
	"github.com/JeremiahM37/librarr/internal/kavita"
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

// fakeKavitaServer stands in for Kavita: it answers the login, serves the
// library list from libsJSON, and records every scan it is asked for. scanCode,
// when non-zero, is the status the scan endpoint answers with instead of 200.
type fakeKavitaServer struct {
	scannedIDs []string
	scanForce  []string
	listCalls  int
}

func startFakeKavita(t *testing.T, libsJSON string, scanCode int, scanBody string) (*fakeKavitaServer, string) {
	t.Helper()
	fake := &fakeKavitaServer{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/api/Account/login"):
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"token":"t"}`))
		case strings.HasSuffix(r.URL.Path, "/api/Library/libraries"):
			fake.listCalls++
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(libsJSON))
		case strings.HasSuffix(r.URL.Path, "/api/Library/scan"):
			fake.scannedIDs = append(fake.scannedIDs, r.URL.Query().Get("libraryId"))
			fake.scanForce = append(fake.scanForce, r.URL.Query().Get("force"))
			if scanCode != 0 {
				w.WriteHeader(scanCode)
				w.Write([]byte(scanBody))
				return
			}
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return fake, srv.URL
}

// The nominal path: Kavita reports its own container-side folder, KAVITA_PATH_MAP
// translates it to the local MANGA_DIR, and only that library is scanned — with
// its id in the query string, which is the parameter whose absence broke this.
func TestFilesKavitaScanTargetsTheMatchingLibrary(t *testing.T) {
	cfg, manga, _ := newRootsFixture(t)
	libs := `[
		{"id":7,"name":"Manga","type":0,"folders":["/kavita/manga"]},
		{"id":9,"name":"Books","type":2,"folders":["/kavita/books"]}
	]`
	fake, url := startFakeKavita(t, libs, 0, "")
	cfg.KavitaURL, cfg.KavitaUser, cfg.KavitaPass = url, "u", "p"
	cfg.KavitaPathMap = "/kavita/manga:" + manga

	rr := postScan(t, &Server{cfg: cfg}, `{"root":"manga"}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("HTTP %d, want 200: %s", rr.Code, rr.Body.String())
	}
	if len(fake.scannedIDs) != 1 || fake.scannedIDs[0] != "7" {
		t.Fatalf("scanned %v, want only library 7", fake.scannedIDs)
	}
	if fake.scanForce[0] != "false" {
		t.Errorf("force = %q, want false for a Browse scan", fake.scanForce[0])
	}
	body := rr.Body.String()
	if !strings.Contains(body, `"name":"Manga"`) || !strings.Contains(body, `"id":7`) {
		t.Errorf("response should name the library scanned, got %s", body)
	}
	if strings.Contains(body, `"scope":"all"`) {
		t.Errorf("scope=all is gone: the scan is targeted now, got %s", body)
	}
}

// No mapping lines up: the request must fail loudly and, above all, must not
// fall back to scanning something.
func TestFilesKavitaScanFailsWhenNoLibraryMatches(t *testing.T) {
	cfg, _, _ := newRootsFixture(t)
	libs := `[{"id":9,"name":"Books","type":2,"folders":["/kavita/books"]}]`
	fake, url := startFakeKavita(t, libs, 0, "")
	cfg.KavitaURL, cfg.KavitaUser, cfg.KavitaPass = url, "u", "p"

	rr := postScan(t, &Server{cfg: cfg}, `{"root":"manga"}`)
	if rr.Code != http.StatusConflict {
		t.Fatalf("HTTP %d, want 409: %s", rr.Code, rr.Body.String())
	}
	if len(fake.scannedIDs) != 0 {
		t.Fatalf("nothing should have been scanned, got %v", fake.scannedIDs)
	}
	body := rr.Body.String()
	if !strings.Contains(body, cfg.MangaDir) {
		t.Errorf("detail should name the local root, got %s", body)
	}
	if !strings.Contains(body, "/kavita/books") {
		t.Errorf("detail should list the Kavita folders it compared, got %s", body)
	}
}

// A library called "Manga" whose folder is somewhere else entirely must not be
// picked: matching on the name would silently rescan the wrong library.
func TestFilesKavitaScanNeverMatchesByLibraryName(t *testing.T) {
	cfg, _, _ := newRootsFixture(t)
	libs := `[{"id":3,"name":"Manga","type":0,"folders":["/somewhere/else"]}]`
	fake, url := startFakeKavita(t, libs, 0, "")
	cfg.KavitaURL, cfg.KavitaUser, cfg.KavitaPass = url, "u", "p"

	rr := postScan(t, &Server{cfg: cfg}, `{"root":"manga"}`)
	if rr.Code != http.StatusConflict {
		t.Fatalf("HTTP %d, want 409 — the name must not create a match", rr.Code)
	}
	if len(fake.scannedIDs) != 0 {
		t.Fatalf("nothing should have been scanned, got %v", fake.scannedIDs)
	}
}

// Kavita answering a non-admin returns libraries without folders. That is a
// distinct diagnosis and the message must say so rather than blaming the paths.
func TestFilesKavitaScanExplainsMissingFolders(t *testing.T) {
	cfg, _, _ := newRootsFixture(t)
	_, url := startFakeKavita(t, `[{"id":3,"name":"Manga","type":0}]`, 0, "")
	cfg.KavitaURL, cfg.KavitaUser, cfg.KavitaPass = url, "u", "p"

	rr := postScan(t, &Server{cfg: cfg}, `{"root":"manga"}`)
	if rr.Code != http.StatusConflict {
		t.Fatalf("HTTP %d, want 409", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "administrator") {
		t.Errorf("detail should point at the admin requirement, got %s", rr.Body.String())
	}
}

// Kavita refusing the scan must reach the user as a status and a reason: an
// opaque "kavita scan failed" is exactly what made this bug hard to diagnose.
func TestFilesKavitaScanReportsKavitaFailure(t *testing.T) {
	cfg, manga, _ := newRootsFixture(t)
	libs := `[{"id":7,"name":"Manga","type":0,"folders":["/kavita/manga"]}]`
	_, url := startFakeKavita(t, libs, http.StatusBadRequest, "Library does not exist")
	cfg.KavitaURL, cfg.KavitaUser, cfg.KavitaPass = url, "u", "p"
	cfg.KavitaPathMap = "/kavita/manga:" + manga

	rr := postScan(t, &Server{cfg: cfg}, `{"root":"manga"}`)
	if rr.Code != http.StatusBadGateway {
		t.Fatalf("HTTP %d, want 502", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "400") || !strings.Contains(body, "Library does not exist") {
		t.Errorf("response should carry Kavita's status and reason, got %s", body)
	}
	if strings.Contains(body, url) {
		t.Errorf("response must not leak the Kavita URL, got %s", body)
	}
}

// Kavita unreachable at listing time is a 502 too, but must not echo the URL.
func TestFilesKavitaScanReportsUnreachableKavita(t *testing.T) {
	cfg, _, _ := newRootsFixture(t)
	cfg.KavitaURL, cfg.KavitaUser, cfg.KavitaPass = "http://kavita.invalid:5000", "u", "s3cret"

	rr := postScan(t, &Server{cfg: cfg}, `{"root":"manga"}`)
	if rr.Code != http.StatusBadGateway {
		t.Fatalf("HTTP %d, want 502", rr.Code)
	}
	for _, leak := range []string{"kavita.invalid", "s3cret"} {
		if strings.Contains(rr.Body.String(), leak) {
			t.Errorf("response leaks %q: %s", leak, rr.Body.String())
		}
	}
}

// The matching rule itself, away from HTTP: both containment directions count,
// and one root may legitimately be backed by several Kavita libraries.
func TestMatchKavitaLibraries(t *testing.T) {
	cfg, manga, base := newRootsFixture(t)
	// Kavita sees /kavita/<x>; the map turns that back into the local base dir,
	// so /kavita/manga resolves to MANGA_DIR itself.
	cfg.KavitaPathMap = "/kavita:" + base

	libs := []kavita.Library{
		{ID: 1, Name: "Exact", Folders: []string{"/kavita/manga"}},
		{ID: 2, Name: "Subfolder of the root", Folders: []string{"/kavita/manga/shonen"}},
		{ID: 3, Name: "Contains the root", Folders: []string{"/kavita"}},
		{ID: 4, Name: "Elsewhere", Folders: []string{"/kavita/ebooks"}},
		{ID: 5, Name: "Lookalike", Folders: []string{"/kavita/manga-other"}},
	}

	matched, seen := matchKavitaLibraries(cfg, libs, manga)
	var ids []int
	for _, m := range matched {
		ids = append(ids, m.ID)
	}
	if len(ids) != 3 || ids[0] != 1 || ids[1] != 2 || ids[2] != 3 {
		t.Errorf("matched %v, want libraries 1, 2 and 3 (exact, below, above)", ids)
	}
	if len(seen) != len(libs) {
		t.Errorf("seen = %v, want one resolved path per folder for the error message", seen)
	}
}

// Without KAVITA_PATH_MAP, ResolveKavitaPath has nothing to validate against.
// Kavita's folders are then already local paths (same host, same mounts) and
// must still be matched rather than dropped.
func TestMatchKavitaLibrariesWithoutPathMap(t *testing.T) {
	cfg, manga, _ := newRootsFixture(t)

	matched, _ := matchKavitaLibraries(cfg,
		[]kavita.Library{{ID: 2, Name: "Manga", Folders: []string{manga}}}, manga)
	if len(matched) != 1 || matched[0].ID != 2 {
		t.Errorf("matched %+v, want the library on the identical local path", matched)
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
