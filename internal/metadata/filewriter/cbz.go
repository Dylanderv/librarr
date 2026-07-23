package filewriter

import (
	"encoding/xml"
	"strings"
)

const comicInfoName = "ComicInfo.xml"

// comicInfo mirrors the common subset of the Anansi Project ComicInfo schema
// that Kavita reads. Fields we don't manage are preserved across a read/modify/
// write cycle because we unmarshal the existing document into this struct
// first. Order follows the schema's typical layout.
type comicInfo struct {
	XMLName xml.Name `xml:"ComicInfo"`

	Title           string `xml:"Title,omitempty"`
	Series          string `xml:"Series,omitempty"`
	Number          string `xml:"Number,omitempty"`
	Count           string `xml:"Count,omitempty"`
	Volume          string `xml:"Volume,omitempty"`
	AlternateSeries string `xml:"AlternateSeries,omitempty"`
	Summary         string `xml:"Summary,omitempty"`
	Notes           string `xml:"Notes,omitempty"`
	Year            string `xml:"Year,omitempty"`
	Month           string `xml:"Month,omitempty"`
	Day             string `xml:"Day,omitempty"`
	Writer          string `xml:"Writer,omitempty"`
	Penciller       string `xml:"Penciller,omitempty"`
	Inker           string `xml:"Inker,omitempty"`
	Colorist        string `xml:"Colorist,omitempty"`
	Letterer        string `xml:"Letterer,omitempty"`
	CoverArtist     string `xml:"CoverArtist,omitempty"`
	Editor          string `xml:"Editor,omitempty"`
	Translator      string `xml:"Translator,omitempty"`
	Publisher       string `xml:"Publisher,omitempty"`
	Imprint         string `xml:"Imprint,omitempty"`
	Genre           string `xml:"Genre,omitempty"`
	Tags            string `xml:"Tags,omitempty"`
	Web             string `xml:"Web,omitempty"`
	PageCount       string `xml:"PageCount,omitempty"`
	LanguageISO     string `xml:"LanguageISO,omitempty"`
	Format          string `xml:"Format,omitempty"`
	AgeRating       string `xml:"AgeRating,omitempty"`
	Manga           string `xml:"Manga,omitempty"`
	Characters      string `xml:"Characters,omitempty"`
	Teams           string `xml:"Teams,omitempty"`
	Locations       string `xml:"Locations,omitempty"`
}

// applyCBZ writes metadata into a CBZ's ComicInfo.xml, creating it if absent.
func applyCBZ(path string, f Fields) error {
	raw, err := readZipEntry(path, comicInfoName)
	if err != nil {
		return err
	}

	var ci comicInfo
	if len(raw) > 0 {
		// Best-effort parse; a malformed existing file just means we start
		// from an empty document rather than aborting the edit.
		_ = xml.Unmarshal(raw, &ci)
	}

	if f.Title != nil {
		ci.Title = *f.Title
	}
	if f.Series != nil {
		ci.Series = *f.Series
	}
	if f.SeriesIndex != nil {
		ci.Number = *f.SeriesIndex
	}
	if f.Summary != nil {
		ci.Summary = *f.Summary
	}
	if f.Publisher != nil {
		ci.Publisher = *f.Publisher
	}
	if f.Language != nil {
		ci.LanguageISO = *f.Language
	}
	if f.Authors != nil {
		ci.Writer = joinList(mergeList(splitList(ci.Writer), f.Authors, f.AppendLists))
	}
	if f.Genres != nil {
		ci.Genre = joinList(mergeList(splitList(ci.Genre), f.Genres, f.AppendLists))
	}
	if f.Tags != nil {
		ci.Tags = joinList(mergeList(splitList(ci.Tags), f.Tags, f.AppendLists))
	}

	out, err := xml.MarshalIndent(&ci, "", "  ")
	if err != nil {
		return err
	}
	doc := append([]byte(xml.Header), out...)
	doc = append(doc, '\n')

	return rewriteZip(path, map[string][]byte{comicInfoName: doc}, map[string][]byte{comicInfoName: doc})
}

// splitList parses a comma-separated ComicInfo list field.
func splitList(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// joinList renders a ComicInfo list field.
func joinList(items []string) string {
	return strings.Join(items, ", ")
}
