package kavita

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakeKavita stands in for a Kavita server. It always answers the login so
// tests only have to describe the endpoint under test; handler receives every
// other request and reports whether it handled it.
func fakeKavita(t *testing.T, handler func(w http.ResponseWriter, r *http.Request) bool) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/api/Account/login") {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"token":"jwt"}`))
			return
		}
		if !handler(w, r) {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return New(srv.URL, "user", "pass")
}

func TestLibrariesReturnsIDNameTypeAndFolders(t *testing.T) {
	c := fakeKavita(t, func(w http.ResponseWriter, r *http.Request) bool {
		if r.URL.Path != "/api/Library/libraries" || r.Method != http.MethodGet {
			return false
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[
			{"id":7,"name":"Manga","type":0,"folders":["/manga","/manga-extra"]},
			{"id":9,"name":"Books","type":2,"folders":["/books"]}
		]`))
		return true
	})

	libs, err := c.Libraries(context.Background())
	if err != nil {
		t.Fatalf("Libraries: %v", err)
	}
	if len(libs) != 2 {
		t.Fatalf("got %d libraries, want 2", len(libs))
	}
	if libs[0].ID != 7 || libs[0].Name != "Manga" || libs[0].Type != 0 {
		t.Errorf("first library = %+v", libs[0])
	}
	if len(libs[0].Folders) != 2 || libs[0].Folders[0] != "/manga" {
		t.Errorf("folders = %v, want the two Kavita-side paths", libs[0].Folders)
	}
}

// The whole bug: libraryId must travel as a query parameter. A body-only or
// parameterless POST is what Kavita rejected in production.
func TestScanLibrarySendsLibraryIDAndForceAsQueryParams(t *testing.T) {
	var gotPath, gotID, gotForce, gotMethod string
	c := fakeKavita(t, func(w http.ResponseWriter, r *http.Request) bool {
		if r.URL.Path != "/api/Library/scan" {
			return false
		}
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotID = r.URL.Query().Get("libraryId")
		gotForce = r.URL.Query().Get("force")
		w.WriteHeader(http.StatusOK)
		return true
	})

	if err := c.ScanLibrary(context.Background(), 7, true); err != nil {
		t.Fatalf("ScanLibrary: %v", err)
	}
	if gotPath != "/api/Library/scan" || gotMethod != http.MethodPost {
		t.Errorf("called %s %s", gotMethod, gotPath)
	}
	if gotID != "7" {
		t.Errorf("libraryId = %q, want \"7\"", gotID)
	}
	if gotForce != "true" {
		t.Errorf("force = %q, want \"true\"", gotForce)
	}

	// force must be explicit in the other direction too, not just omitted.
	if err := c.ScanLibrary(context.Background(), 3, false); err != nil {
		t.Fatalf("ScanLibrary(force=false): %v", err)
	}
	if gotID != "3" || gotForce != "false" {
		t.Errorf("libraryId=%q force=%q, want 3/false", gotID, gotForce)
	}
}

func TestScanAllLibrariesScansEachListedLibrary(t *testing.T) {
	var scanned []string
	c := fakeKavita(t, func(w http.ResponseWriter, r *http.Request) bool {
		switch r.URL.Path {
		case "/api/Library/libraries":
			w.Write([]byte(`[{"id":1,"name":"A"},{"id":4,"name":"B"}]`))
		case "/api/Library/scan":
			scanned = append(scanned, r.URL.Query().Get("libraryId"))
			w.WriteHeader(http.StatusOK)
		default:
			return false
		}
		return true
	})

	if err := c.ScanAllLibraries(context.Background(), false); err != nil {
		t.Fatalf("ScanAllLibraries: %v", err)
	}
	if strings.Join(scanned, ",") != "1,4" {
		t.Errorf("scanned %v, want every listed library id", scanned)
	}
}

func TestScanLibrarySurfacesKavitaStatusAndBody(t *testing.T) {
	c := fakeKavita(t, func(w http.ResponseWriter, r *http.Request) bool {
		if r.URL.Path != "/api/Library/scan" {
			return false
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Library does not exist"))
		return true
	})

	err := c.ScanLibrary(context.Background(), 42, false)
	if err == nil {
		t.Fatal("expected an error on HTTP 400")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error %v is not an *APIError", err)
	}
	if apiErr.Status != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", apiErr.Status)
	}
	if apiErr.Body != "Library does not exist" {
		t.Errorf("body = %q, want Kavita's reason preserved", apiErr.Body)
	}

	detail := PublicErrorDetail(err)
	if !strings.Contains(detail, "400") || !strings.Contains(detail, "Library does not exist") {
		t.Errorf("detail = %q, want the status and the reason", detail)
	}
}

// PublicErrorDetail is echoed to HTTP clients, so a transport error — whose
// message embeds the full Kavita URL — must never be passed through.
func TestPublicErrorDetailHidesInternalURLs(t *testing.T) {
	c := New("http://kavita.internal:5000", "user", "s3cret")
	err := c.ScanLibrary(context.Background(), 1, false)
	if err == nil {
		t.Fatal("expected a transport error against an unreachable host")
	}
	detail := PublicErrorDetail(err)
	for _, leak := range []string{"kavita.internal", "5000", "s3cret", "http://"} {
		if strings.Contains(detail, leak) {
			t.Errorf("detail %q leaks %q", detail, leak)
		}
	}
	if detail == "" {
		t.Error("detail should still say something useful")
	}
}

func TestPublicErrorDetailTruncatesToOneLine(t *testing.T) {
	err := &APIError{Status: 500, Body: "first line\n<html>a very long page</html>"}
	detail := PublicErrorDetail(err)
	if strings.Contains(detail, "\n") || strings.Contains(detail, "html") {
		t.Errorf("detail = %q, want a single trimmed line", detail)
	}
	if !strings.Contains(detail, "first line") {
		t.Errorf("detail = %q, want the first line kept", detail)
	}
}
