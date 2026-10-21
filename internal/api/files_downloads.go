package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/JeremiahM37/librarr/internal/config"
)

// Active telegram downloads for the Browse tab.
//
// Files land in the libraries straight from the telegram-downloader sidecar,
// so a file can sit in a listing while still downloading — and every Browse
// write (rename, move, metadata, convert) would happily chew on the partial
// file. This endpoint lets the tab badge those rows with a progress bar.
//
// It proxies the sidecar's GET /downloads, keeping only queued/downloading
// jobs with a known filename, matched by the front end on the bare file
// name — both services mount the same /books/... volumes, so names line up
// without any path translation.
//
// The sidecar is optional: unreachable means an empty list, never an error,
// and a 30 s backoff keeps a dead sidecar from slowing every Browse poll.
// The 4 s cache keeps five browsers polling from hammering it.

// activeDownload is one in-flight file.
type activeDownload struct {
	Name     string  `json:"name"`
	Progress float64 `json:"progress"`
	Status   string  `json:"status"`
}

// sidecarJob mirrors just what the badge needs from the sidecar's JobOut.
type sidecarJob struct {
	Filename *string `json:"filename"`
	Progress float64 `json:"progress"`
	Status   string  `json:"status"`
}

const (
	dlCacheTTL       = 4 * time.Second
	dlUnreachableFor = 30 * time.Second
	dlProxyTimeout   = 3 * time.Second
)

var (
	dlClient = &http.Client{Timeout: dlProxyTimeout}

	dlMu          sync.Mutex
	dlUntil       time.Time
	dlItems       []activeDownload
	dlUnreachable time.Time
)

// resetDownloadsCache empties the proxy cache. Tests only.
func resetDownloadsCache() {
	dlMu.Lock()
	dlUntil, dlUnreachable = time.Time{}, time.Time{}
	dlItems = nil
	dlMu.Unlock()
}

// handleFilesDownloads lists the sidecar's in-flight downloads.
//
// GET /api/files/downloads
func (s *Server) handleFilesDownloads(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.activeDownloads())
}

// activeDownloads returns the cached in-flight downloads, refreshing from the
// sidecar when stale.
func (s *Server) activeDownloads() []activeDownload {
	now := time.Now()

	dlMu.Lock()
	if now.Before(dlUntil) {
		out := dlItems
		dlMu.Unlock()
		return out
	}
	if now.Before(dlUnreachable) {
		dlMu.Unlock()
		return nil
	}
	dlMu.Unlock()

	items := fetchActiveDownloads(s.cfg)

	dlMu.Lock()
	if items == nil {
		// Sidecar down: serve emptiness for a while instead of retrying
		// every Browse poll.
		dlItems = nil
		dlUnreachable = now.Add(dlUnreachableFor)
	} else {
		dlItems = items
		dlUntil = now.Add(dlCacheTTL)
	}
	out := dlItems
	dlMu.Unlock()
	return out
}

// fetchActiveDownloads queries the sidecar, returning nil when it cannot be
// reached — nil (unknown) and [] (nothing in flight) both serve as empty to
// the front end, but only nil triggers the backoff.
func fetchActiveDownloads(cfg *config.Config) []activeDownload {
	if cfg == nil {
		return nil
	}
	base := strings.TrimSpace(cfg.TelegramSidecarURL)
	if base == "" {
		return nil
	}
	req, err := http.NewRequest(http.MethodGet, strings.TrimSuffix(base, "/")+"/downloads", nil)
	if err != nil {
		return nil
	}
	if key := strings.TrimSpace(cfg.TelegramSidecarAPIKey); key != "" {
		req.Header.Set("X-API-Key", key)
	}
	resp, err := dlClient.Do(req)
	if err != nil {
		slog.Warn("telegram sidecar unreachable", "url", base, "err", err)
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		slog.Warn("telegram sidecar refused downloads", "status", resp.Status)
		return nil
	}
	var jobs []sidecarJob
	if err := json.NewDecoder(resp.Body).Decode(&jobs); err != nil {
		slog.Warn("telegram sidecar sent bad downloads", "err", err)
		return nil
	}
	var out []activeDownload
	for _, j := range jobs {
		if j.Status != "queued" && j.Status != "downloading" {
			continue
		}
		if j.Filename == nil || strings.TrimSpace(*j.Filename) == "" {
			continue
		}
		out = append(out, activeDownload{Name: *j.Filename, Progress: j.Progress, Status: j.Status})
	}
	if out == nil {
		out = []activeDownload{}
	}
	return out
}
