// Package kavita is a small client for the Kavita REST API. It centralizes
// authentication (JWT login with in-client token caching) and the handful of
// endpoints librarr talks to, so callers don't re-implement the login dance.
//
// Before this package the login flow was copy-pasted in three places
// (api.library_external, organize.targets, and the scan helper). Construct one
// Client per logical operation; a bulk metadata edit that touches many series
// reuses a single token instead of logging in per request.
package kavita

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Client talks to a single Kavita server. It is safe for concurrent use.
type Client struct {
	baseURL string
	user    string
	pass    string
	http    *http.Client

	mu    sync.Mutex
	token string
}

// New builds a client for the given Kavita base URL and credentials. baseURL
// should be the internal URL librarr uses to reach Kavita (e.g. the value of
// KAVITA_URL), without a trailing slash requirement.
func New(baseURL, user, pass string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		user:    user,
		pass:    pass,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// Library is a Kavita library: its id (the only thing a scan can be targeted
// on), its name, its type, and the folders it watches.
//
// Folders holds paths as *Kavita* sees them — inside its own container, which
// is generally not where librarr sees the same files. Config.ResolveKavitaPath
// (KAVITA_PATH_MAP) is what translates them to local paths.
type Library struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	// Type is Kavita's LibraryType enum: 0 Manga, 1 Comic, 2 Book, 3 Image,
	// 4 LightNovel, 5 ComicVine. Carried for display and diagnostics only —
	// librarr never matches a library on it.
	Type    int      `json:"type"`
	Folders []string `json:"folders"`
}

// Series is the trimmed-down view of a Kavita series librarr renders and edits.
type Series struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Pages       int    `json:"pages"`
	LibraryName string `json:"libraryName"`
	LibraryID   int    `json:"libraryId"`
}

// SeriesMetadata mirrors the subset of Kavita's SeriesMetadataDto librarr uses
// to pre-fill the bulk-edit form. Kavita exposes far more; we read what maps to
// fields we can write back into files.
type SeriesMetadata struct {
	SeriesID    int      `json:"seriesId"`
	Summary     string   `json:"summary"`
	Publishers  []Person `json:"publishers"`
	Writers     []Person `json:"writers"`
	Genres      []Tag    `json:"genres"`
	Tags        []Tag    `json:"tags"`
	Language    string   `json:"language"`
	ReleaseYear int      `json:"releaseYear"`
}

// Person is a Kavita people entry (writer, publisher, etc.).
type Person struct {
	Name string `json:"name"`
}

// Tag is a Kavita genre/tag entry.
type Tag struct {
	Title string `json:"title"`
}

// MangaFile is one file backing a chapter, with its on-disk path as Kavita sees
// it. FilePath is absolute and normalized by Kavita.
type MangaFile struct {
	FilePath  string `json:"filePath"`
	Extension string `json:"extension"`
	// Format is Kavita's MangaFormat enum (3 = Epub, 1 = Archive/cbz…). We
	// prefer the file extension for routing, but keep this for diagnostics.
	Format int `json:"format"`
}

type volumeDTO struct {
	Chapters []struct {
		Files []MangaFile `json:"files"`
	} `json:"chapters"`
}

// APIError is a non-2xx answer from Kavita, carrying the status and the body
// Kavita sent with it. It exists so callers can tell "Kavita said no, and here
// is why" apart from "Kavita could not be reached at all" — the two need very
// different fixes, and collapsing them into one opaque message is what made a
// missing libraryId parameter hard to diagnose in production.
type APIError struct {
	Method string
	Path   string
	Status int
	Body   string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("kavita %s %s: HTTP %d: %s", e.Method, e.Path, e.Status, e.Body)
}

// PublicErrorDetail renders err as a short explanation safe to return over
// librarr's own HTTP API.
//
// It deliberately never includes the Kavita base URL, the credentials or the
// Authorization header: transport errors from net/http embed the full request
// URL, so anything that is not an APIError is reduced to a generic
// unreachable message rather than being echoed. Kavita's own response body is
// passed through (truncated), since that is the part that actually tells the
// user what Kavita rejected.
func PublicErrorDetail(err error) string {
	if err == nil {
		return ""
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return "Kavita could not be reached"
	}
	detail := fmt.Sprintf("Kavita responded HTTP %d", apiErr.Status)
	reason := strings.TrimSpace(apiErr.Body)
	if reason == "" {
		return detail
	}
	// One line, bounded: Kavita can answer with a full HTML error page.
	if i := strings.IndexAny(reason, "\r\n"); i >= 0 {
		reason = strings.TrimSpace(reason[:i])
	}
	if len(reason) > 200 {
		reason = reason[:200] + "…"
	}
	if reason == "" {
		return detail
	}
	return detail + ": " + reason
}

// login performs the JWT login and caches the token on the client.
func (c *Client) login(ctx context.Context) (string, error) {
	payload, _ := json.Marshal(map[string]string{
		"username": c.user,
		"password": c.pass,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/api/Account/login", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("kavita login request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("kavita login HTTP %d", resp.StatusCode)
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("kavita login decode: %w", err)
	}
	if result.Token == "" {
		return "", fmt.Errorf("kavita login returned empty token")
	}

	c.mu.Lock()
	c.token = result.Token
	c.mu.Unlock()
	return result.Token, nil
}

// Token returns a cached JWT, logging in if necessary. Exposed so callers that
// still build raw requests (existing series/search handlers) can share the
// token cache.
func (c *Client) Token(ctx context.Context) (string, error) {
	c.mu.Lock()
	tok := c.token
	c.mu.Unlock()
	if tok != "" {
		return tok, nil
	}
	return c.login(ctx)
}

// do issues an authenticated request against path (e.g. "/api/Series/scan"),
// re-authenticating once on a 401. When out is non-nil the JSON body is decoded
// into it. It returns the raw response for callers that need headers.
func (c *Client) do(ctx context.Context, method, path string, body any, out any) (*http.Response, error) {
	send := func(token string) (*http.Response, error) {
		var r io.Reader
		if body != nil {
			b, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			r = bytes.NewReader(b)
		}
		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, r)
		if err != nil {
			return nil, err
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		return c.http.Do(req)
	}

	token, err := c.Token(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := send(token)
	if err != nil {
		return nil, err
	}

	// Token may have expired: force a fresh login and retry once.
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		c.mu.Lock()
		c.token = ""
		c.mu.Unlock()
		if token, err = c.login(ctx); err != nil {
			return nil, err
		}
		if resp, err = send(token); err != nil {
			return nil, err
		}
	}

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		resp.Body.Close()
		return nil, &APIError{
			Method: method,
			Path:   path,
			Status: resp.StatusCode,
			Body:   strings.TrimSpace(string(b)),
		}
	}

	if out != nil {
		defer resp.Body.Close()
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return nil, fmt.Errorf("kavita %s %s decode: %w", method, path, err)
		}
		return resp, nil
	}

	// No body to decode (scan triggers and friends). Drain and close anyway:
	// returning an unread body leaks the connection out of the pool, and these
	// are exactly the calls made repeatedly by the import pipeline.
	io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	resp.Body.Close()
	return resp, nil
}

// SeriesMetadata fetches a series' metadata for pre-filling the edit form.
func (c *Client) SeriesMetadata(ctx context.Context, seriesID int) (*SeriesMetadata, error) {
	var md SeriesMetadata
	_, err := c.do(ctx, http.MethodGet, fmt.Sprintf("/api/Series/metadata?seriesId=%d", seriesID), nil, &md)
	if err != nil {
		return nil, err
	}
	return &md, nil
}

// FilesForSeries returns every backing file across all volumes/chapters of a
// series, so callers know which files on disk to edit.
func (c *Client) FilesForSeries(ctx context.Context, seriesID int) ([]MangaFile, error) {
	var volumes []volumeDTO
	_, err := c.do(ctx, http.MethodGet, fmt.Sprintf("/api/Series/volumes?seriesId=%d", seriesID), nil, &volumes)
	if err != nil {
		return nil, err
	}
	var files []MangaFile
	for _, v := range volumes {
		for _, ch := range v.Chapters {
			files = append(files, ch.Files...)
		}
	}
	return files, nil
}

// ScanSeries triggers a targeted rescan of one series so Kavita re-reads the
// metadata we just wrote into its files.
func (c *Client) ScanSeries(ctx context.Context, libraryID, seriesID int) error {
	body := map[string]any{
		"libraryId":   libraryID,
		"seriesId":    seriesID,
		"forceUpdate": true,
	}
	_, err := c.do(ctx, http.MethodPost, "/api/Series/scan", body, nil)
	return err
}

// Libraries lists every library on the server, with the folders each one
// watches as Kavita sees them (container-side paths, which callers translate
// with Config.ResolveKavitaPath).
//
// Folders is only populated for an admin account: Kavita answers a non-admin
// with a LiteLibraryDto that carries id/name/type and nothing else. A caller
// that needs to match folders should treat an all-empty Folders as "librarr's
// Kavita user is not an admin" rather than "no library matches".
func (c *Client) Libraries(ctx context.Context) ([]Library, error) {
	var libs []Library
	if _, err := c.do(ctx, http.MethodGet, "/api/Library/libraries", nil, &libs); err != nil {
		return nil, err
	}
	return libs, nil
}

// ScanLibrary triggers a scan of one Kavita library.
//
// libraryId goes in the *query string*, not in a JSON body: Kavita's
// LibraryController binds it from the query, and a POST /api/Library/scan with
// no libraryId is rejected outright. That missing parameter is what made every
// scan librarr requested fail. force=true makes Kavita ignore its
// skip-unchanged optimisations and re-read files as if it were a first scan.
func (c *Client) ScanLibrary(ctx context.Context, libraryID int, force bool) error {
	q := url.Values{}
	q.Set("libraryId", strconv.Itoa(libraryID))
	q.Set("force", strconv.FormatBool(force))
	_, err := c.do(ctx, http.MethodPost, "/api/Library/scan?"+q.Encode(), nil, nil)
	return err
}

// ScanAllLibraries scans every library on the server, listing them first and
// issuing one targeted scan per library.
//
// Kavita does expose POST /api/Library/scan-all, but librarr goes through the
// list + per-library scan path on purpose: it is the exact request shape known
// to work against the Kavita versions librarr targets, and a failure names the
// library that failed instead of being a single opaque error. Scanning
// continues across failures so one broken library cannot mask the others; the
// first error is returned once every library has been attempted.
func (c *Client) ScanAllLibraries(ctx context.Context, force bool) error {
	libs, err := c.Libraries(ctx)
	if err != nil {
		return err
	}
	if len(libs) == 0 {
		return fmt.Errorf("kavita reported no libraries to scan")
	}
	var firstErr error
	for _, lib := range libs {
		if err := c.ScanLibrary(ctx, lib.ID, force); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("scanning library %q (id %d): %w", lib.Name, lib.ID, err)
		}
	}
	return firstErr
}

// SeriesForID looks up a single series' library id (needed for a targeted scan
// when the caller only has the series id).
func (c *Client) SeriesForID(ctx context.Context, seriesID int) (*Series, error) {
	var sr Series
	_, err := c.do(ctx, http.MethodGet, fmt.Sprintf("/api/Series/%d", seriesID), nil, &sr)
	if err != nil {
		return nil, err
	}
	return &sr, nil
}

// SearchSeries runs Kavita's search and returns matching series.
func (c *Client) SearchSeries(ctx context.Context, query string) ([]Series, error) {
	var result struct {
		Series []Series `json:"series"`
	}
	_, err := c.do(ctx, http.MethodGet,
		"/api/Search/search?queryString="+url.QueryEscape(query), nil, &result)
	if err != nil {
		return nil, err
	}
	return result.Series, nil
}

// AllSeries lists series (1-indexed page) and returns the total count parsed
// from Kavita's Pagination response header.
func (c *Client) AllSeries(ctx context.Context, page, pageSize int) ([]Series, int, error) {
	body := map[string]any{
		"statements":  []any{},
		"combination": 1,
		"limitTo":     0,
		"sortOptions": map[string]any{
			"sortField":   1,
			"isAscending": true,
		},
	}
	var series []Series
	resp, err := c.do(ctx, http.MethodPost,
		fmt.Sprintf("/api/Series/all-v2?pageNumber=%d&pageSize=%d", page-1, pageSize),
		body, &series)
	if err != nil {
		return nil, 0, err
	}
	total := 0
	if pagination := resp.Header.Get("Pagination"); pagination != "" {
		var pag struct {
			TotalItems int `json:"totalItems"`
		}
		if json.Unmarshal([]byte(pagination), &pag) == nil {
			total = pag.TotalItems
		}
	}
	return series, total, nil
}
