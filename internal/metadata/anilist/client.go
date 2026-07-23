// Package anilist fetches manga metadata from the AniList GraphQL API.
//
// AniList is a free, no-auth-required public API (POST GraphQL at
// https://graphql.anilist.co). It is a good fit for manga metadata, mirroring
// what Kavita+ uses. This client mirrors the style of
// internal/metadata/openlibrary.go: one *http.Client, an in-memory cache with a
// 30-minute TTL, a shared User-Agent, and a single retry on HTTP 429.
package anilist

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	apiURL    = "https://graphql.anilist.co"
	userAgent = "Librarr/2.0 (book download manager; github.com/JeremiahM37/librarr)"
	cacheTTL  = 30 * time.Minute
)

// MangaMatch is a single search candidate mapped from AniList into the shape
// the bulk metadata editor consumes.
type MangaMatch struct {
	AniListID int      `json:"anilist_id"`
	Title     string   `json:"title"`
	Summary   string   `json:"summary"`
	Authors   []string `json:"authors"`
	Genres    []string `json:"genres"`
	Tags      []string `json:"tags"`
	Year      string   `json:"year"`
	CoverURL  string   `json:"cover_url"`
	SiteURL   string   `json:"site_url"`
	Chapters  int      `json:"chapters"`
	Volumes   int      `json:"volumes"`
	Format    string   `json:"format"`
}

type cacheEntry struct {
	data      []MangaMatch
	fetchedAt time.Time
}

// Client fetches manga metadata from AniList with in-memory caching.
type Client struct {
	httpClient *http.Client
	mu         sync.RWMutex
	cache      map[string]cacheEntry
}

// NewClient creates an AniList client.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{
		httpClient: httpClient,
		cache:      make(map[string]cacheEntry),
	}
}

const searchQuery = `query ($search: String) {
  Page(perPage: 10) {
    media(search: $search, type: MANGA) {
      id
      title { romaji english native }
      description(asHtml: false)
      genres
      tags { name rank isGeneralSpoiler isAdult }
      staff { edges { role node { name { full } } } }
      startDate { year }
      status
      format
      countryOfOrigin
      chapters
      volumes
      synonyms
      isAdult
      coverImage { large }
      siteUrl
    }
  }
}`

// SearchManga returns up to ~10 candidate matches for the given title.
// It returns (nil, nil) when nothing is found.
func (c *Client) SearchManga(ctx context.Context, title string) ([]MangaMatch, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, nil
	}
	cacheKey := strings.ToLower(title)

	c.mu.RLock()
	if entry, ok := c.cache[cacheKey]; ok && time.Since(entry.fetchedAt) < cacheTTL {
		c.mu.RUnlock()
		return entry.data, nil
	}
	c.mu.RUnlock()

	resp, err := c.query(ctx, title)
	if err != nil {
		return nil, err
	}

	matches := make([]MangaMatch, 0, len(resp.Data.Page.Media))
	for _, m := range resp.Data.Page.Media {
		matches = append(matches, m.toMatch())
	}

	c.mu.Lock()
	c.cache[cacheKey] = cacheEntry{data: matches, fetchedAt: time.Now()}
	c.mu.Unlock()

	if len(matches) == 0 {
		return nil, nil
	}
	return matches, nil
}

// query POSTs the GraphQL request, retrying once on HTTP 429 after honoring
// the Retry-After header.
func (c *Client) query(ctx context.Context, title string) (*graphQLResponse, error) {
	body, err := json.Marshal(map[string]any{
		"query":     searchQuery,
		"variables": map[string]any{"search": title},
	})
	if err != nil {
		return nil, err
	}

	for attempt := 0; attempt < 2; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", userAgent)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == http.StatusTooManyRequests && attempt == 0 {
			wait := retryAfter(resp.Header.Get("Retry-After"))
			resp.Body.Close()
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(wait):
			}
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("anilist HTTP %d", resp.StatusCode)
		}

		var out graphQLResponse
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()
		if len(out.Errors) > 0 {
			return nil, fmt.Errorf("anilist: %s", out.Errors[0].Message)
		}
		return &out, nil
	}
	return nil, fmt.Errorf("anilist: rate limited")
}

// retryAfter parses the Retry-After header (delay in seconds), clamped to a
// sane range so a hostile/broken header can't stall us for minutes.
func retryAfter(h string) time.Duration {
	secs, err := strconv.Atoi(strings.TrimSpace(h))
	if err != nil || secs <= 0 {
		return time.Second
	}
	if secs > 10 {
		secs = 10
	}
	return time.Duration(secs) * time.Second
}

// GraphQL response types.

type graphQLResponse struct {
	Data struct {
		Page struct {
			Media []media `json:"media"`
		} `json:"Page"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type media struct {
	ID    int `json:"id"`
	Title struct {
		Romaji  string `json:"romaji"`
		English string `json:"english"`
		Native  string `json:"native"`
	} `json:"title"`
	Description string   `json:"description"`
	Genres      []string `json:"genres"`
	Tags        []struct {
		Name             string `json:"name"`
		Rank             int    `json:"rank"`
		IsGeneralSpoiler bool   `json:"isGeneralSpoiler"`
		IsAdult          bool   `json:"isAdult"`
	} `json:"tags"`
	Staff struct {
		Edges []struct {
			Role string `json:"role"`
			Node struct {
				Name struct {
					Full string `json:"full"`
				} `json:"name"`
			} `json:"node"`
		} `json:"edges"`
	} `json:"staff"`
	StartDate struct {
		Year int `json:"year"`
	} `json:"startDate"`
	Status          string   `json:"status"`
	Format          string   `json:"format"`
	CountryOfOrigin string   `json:"countryOfOrigin"`
	Chapters        int      `json:"chapters"`
	Volumes         int      `json:"volumes"`
	Synonyms        []string `json:"synonyms"`
	IsAdult         bool     `json:"isAdult"`
	CoverImage      struct {
		Large string `json:"large"`
	} `json:"coverImage"`
	SiteURL string `json:"siteUrl"`
}

func (m *media) toMatch() MangaMatch {
	out := MangaMatch{
		AniListID: m.ID,
		Title:     m.displayTitle(),
		Summary:   stripHTML(m.Description),
		Authors:   m.authors(),
		Genres:    m.Genres,
		Tags:      m.tagNames(),
		CoverURL:  m.CoverImage.Large,
		SiteURL:   m.SiteURL,
		Chapters:  m.Chapters,
		Volumes:   m.Volumes,
		Format:    m.Format,
	}
	if m.StartDate.Year > 0 {
		out.Year = strconv.Itoa(m.StartDate.Year)
	}
	if out.Genres == nil {
		out.Genres = []string{}
	}
	return out
}

// displayTitle prefers the English title, then romaji, then native.
func (m *media) displayTitle() string {
	switch {
	case m.Title.English != "":
		return m.Title.English
	case m.Title.Romaji != "":
		return m.Title.Romaji
	default:
		return m.Title.Native
	}
}

// authors returns staff credited with "Story" (incl. "Story & Art"), falling
// back to "Art" credits when there is no story credit.
func (m *media) authors() []string {
	var story, art []string
	for _, e := range m.Staff.Edges {
		name := strings.TrimSpace(e.Node.Name.Full)
		if name == "" {
			continue
		}
		role := strings.ToLower(e.Role)
		if strings.Contains(role, "story") {
			story = append(story, name)
		} else if strings.Contains(role, "art") {
			art = append(art, name)
		}
	}
	authors := story
	if len(authors) == 0 {
		authors = art
	}
	return dedupe(authors)
}

// tagNames returns tag names, dropping spoilers, adult tags, and low-rank
// (< 50) noise.
func (m *media) tagNames() []string {
	out := make([]string, 0, len(m.Tags))
	for _, t := range m.Tags {
		if t.IsGeneralSpoiler || t.IsAdult || t.Rank < 50 {
			continue
		}
		if name := strings.TrimSpace(t.Name); name != "" {
			out = append(out, name)
		}
	}
	return out
}

func dedupe(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

var (
	brTag     = regexp.MustCompile(`(?i)<br\s*/?>`)
	otherTags = regexp.MustCompile(`<[^>]+>`)
	manyLines = regexp.MustCompile(`\n{3,}`)
)

// stripHTML converts AniList's lightweight HTML description into plain text:
// <br> → newline, all other tags removed, HTML entities unescaped.
func stripHTML(s string) string {
	if s == "" {
		return ""
	}
	s = brTag.ReplaceAllString(s, "\n")
	s = otherTags.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	s = manyLines.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}
