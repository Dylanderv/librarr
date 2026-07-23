package anilist

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

// roundTripFunc lets a test stub the transport without any live network call.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func stubClient(status int, body string, header http.Header) *Client {
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: status,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     header,
		}, nil
	})
	return NewClient(&http.Client{Transport: rt})
}

const sampleJSON = `{
  "data": {
    "Page": {
      "media": [
        {
          "id": 30013,
          "title": {"romaji": "ONE PIECE", "english": "One Piece", "native": "ワンピース"},
          "description": "Gold Roger was known as the <b>Pirate King</b>.<br><br>Luffy sets sail.",
          "genres": ["Action", "Adventure"],
          "tags": [
            {"name": "Pirates", "rank": 90, "isGeneralSpoiler": false, "isAdult": false},
            {"name": "Shounen", "rank": 80, "isGeneralSpoiler": false, "isAdult": false},
            {"name": "Minor Noise", "rank": 20, "isGeneralSpoiler": false, "isAdult": false},
            {"name": "Big Twist", "rank": 95, "isGeneralSpoiler": true, "isAdult": false}
          ],
          "staff": {"edges": [
            {"role": "Story & Art", "node": {"name": {"full": "Eiichiro Oda"}}},
            {"role": "Assistant", "node": {"name": {"full": "Someone Else"}}}
          ]},
          "startDate": {"year": 1997},
          "status": "RELEASING",
          "format": "MANGA",
          "countryOfOrigin": "JP",
          "chapters": null,
          "volumes": 100,
          "synonyms": [],
          "isAdult": false,
          "coverImage": {"large": "https://example.com/cover.jpg"},
          "siteUrl": "https://anilist.co/manga/30013"
        }
      ]
    }
  }
}`

func TestSearchManga(t *testing.T) {
	c := stubClient(200, sampleJSON, nil)
	matches, err := c.SearchManga(context.Background(), "one piece")
	if err != nil {
		t.Fatalf("SearchManga: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	m := matches[0]

	if m.AniListID != 30013 {
		t.Errorf("AniListID = %d, want 30013", m.AniListID)
	}
	if m.Title != "One Piece" {
		t.Errorf("Title = %q, want English title", m.Title)
	}
	if want := []string{"Eiichiro Oda"}; !equal(m.Authors, want) {
		t.Errorf("Authors = %v, want %v", m.Authors, want)
	}
	if want := []string{"Pirates", "Shounen"}; !equal(m.Tags, want) {
		t.Errorf("Tags = %v, want %v (spoiler and low-rank tags should be dropped)", m.Tags, want)
	}
	if m.Year != "1997" {
		t.Errorf("Year = %q, want 1997", m.Year)
	}
	if strings.Contains(m.Summary, "<") {
		t.Errorf("Summary still contains HTML: %q", m.Summary)
	}
	if !strings.Contains(m.Summary, "Pirate King") {
		t.Errorf("Summary lost content: %q", m.Summary)
	}
	if !strings.Contains(m.Summary, "\n") {
		t.Errorf("Summary should keep a newline from <br>: %q", m.Summary)
	}
}

func TestSearchMangaEmpty(t *testing.T) {
	c := stubClient(200, `{"data":{"Page":{"media":[]}}}`, nil)
	matches, err := c.SearchManga(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("SearchManga: %v", err)
	}
	if matches != nil {
		t.Errorf("expected nil for no results, got %v", matches)
	}
}

func TestSearchMangaBlankTitle(t *testing.T) {
	c := stubClient(200, sampleJSON, nil)
	matches, err := c.SearchManga(context.Background(), "   ")
	if err != nil {
		t.Fatalf("SearchManga: %v", err)
	}
	if matches != nil {
		t.Errorf("expected nil for blank title, got %v", matches)
	}
}

func TestSearchMangaCaches(t *testing.T) {
	calls := 0
	rt := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(sampleJSON))}, nil
	})
	c := NewClient(&http.Client{Transport: rt})

	if _, err := c.SearchManga(context.Background(), "One Piece"); err != nil {
		t.Fatal(err)
	}
	// Different casing must hit the same cache key.
	if _, err := c.SearchManga(context.Background(), "one piece"); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Errorf("expected 1 network call (cached), got %d", calls)
	}
}

func TestStripHTML(t *testing.T) {
	got := stripHTML("A <i>tale</i>.<br><br>The end &amp; more.")
	want := "A tale.\n\nThe end & more."
	if got != want {
		t.Errorf("stripHTML = %q, want %q", got, want)
	}
}

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
