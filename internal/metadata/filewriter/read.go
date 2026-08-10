package filewriter

import (
	"encoding/xml"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// Reading is the mirror of Apply: same formats, same containers, same parsers.
// It exists so a UI can show what a file currently holds *before* overwriting
// it — a metadata edit that propagates over a folder is destructive, and the
// only honest way to offer it is to display what it would replace.
//
// It is deliberately not part of a directory listing: opening every archive of
// a folder on a network share costs one round-trip per file. Callers read on
// demand, for the files they are about to touch.

// Read returns the metadata currently stored in the file at path.
//
// The pointer semantics of Fields are preserved: a field is non-nil only when
// the file actually carries a value for it, so the result can be compared
// field-by-field with an edit without confusing "empty" and "absent".
//
// A file whose format simply has no metadata container — a CBZ without
// ComicInfo.xml, an OPF without a <metadata> element — is not an error: it
// reads as an empty Fields. Only a broken container (unreadable zip, EPUB
// without container.xml) is. Formats Apply cannot write are rejected with
// ErrUnsupported, so a caller never shows values it could not write back.
func Read(path string) (Fields, error) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".epub":
		return readEPUB(path)
	case ".cbz", ".zip":
		return readCBZ(path)
	default:
		return Fields{}, ErrUnsupported{Ext: strings.ToLower(filepath.Ext(path))}
	}
}

// MergeLists exposes the list-merging rule Apply uses, so a caller can compute
// the value a write would produce without performing the write.
func MergeLists(existing, desired []string, appendMode bool) []string {
	return mergeList(existing, desired, appendMode)
}

// readCBZ pulls the fields out of a CBZ's ComicInfo.xml.
//
// A malformed ComicInfo reads as empty rather than as an error, mirroring
// applyCBZ: it starts from a fresh document in that case, so "nothing usable in
// there" is exactly what the user is about to replace.
func readCBZ(path string) (Fields, error) {
	raw, err := readZipEntry(path, comicInfoName)
	if err != nil {
		return Fields{}, err
	}
	if len(raw) == 0 {
		return Fields{}, nil
	}
	var ci comicInfo
	if err := xml.Unmarshal(raw, &ci); err != nil {
		return Fields{}, nil
	}

	// ComicInfo has no ISBN element in the subset we manage, so ISBN stays nil.
	return Fields{
		Title:       optString(ci.Title),
		Series:      optString(ci.Series),
		SeriesIndex: optString(ci.Number),
		Summary:     optString(ci.Summary),
		Publisher:   optString(ci.Publisher),
		Language:    optString(ci.LanguageISO),
		Authors:     splitList(ci.Writer),
		Genres:      splitList(ci.Genre),
		Tags:        splitList(ci.Tags),
	}, nil
}

// readEPUB pulls the fields out of an EPUB's OPF package document.
func readEPUB(path string) (Fields, error) {
	container, err := readZipEntry(path, "META-INF/container.xml")
	if err != nil {
		return Fields{}, err
	}
	if container == nil {
		return Fields{}, errEPUBNoContainer
	}
	opfPath, err := opfPathFromContainer(container)
	if err != nil {
		return Fields{}, err
	}
	opf, err := readZipEntry(path, opfPath)
	if err != nil {
		return Fields{}, err
	}
	if opf == nil {
		return Fields{}, errEPUBNoOPF{Path: opfPath}
	}

	inner, _, _, err := opfMetadataSection(opf)
	if err != nil {
		// No <metadata> block at all: the book carries no metadata. Writing to it
		// would fail, but reading it is simply empty.
		return Fields{}, nil
	}

	return Fields{
		Title:       optString(firstOf(extractDC(inner, "title"))),
		Summary:     optString(firstOf(extractDC(inner, "description"))),
		Publisher:   optString(firstOf(extractDC(inner, "publisher"))),
		Language:    optString(firstOf(extractDC(inner, "language"))),
		Series:      optString(readOPFMeta(inner, "calibre:series")),
		SeriesIndex: optString(readOPFMeta(inner, "calibre:series_index")),
		ISBN:        optString(readISBN(inner)),
		Authors:     extractDC(inner, "creator"),
		Genres:      extractDC(inner, "subject"),
	}, nil
}

// opfMetaPatterns read a <meta name=… content=…/> pair. Attribute order is not
// fixed by the OPF schema and Calibre-produced files are not the only ones we
// get, so both orders are tried.
var opfMetaPatterns = []string{
	`(?is)<meta\b[^>]*\bname\s*=\s*["']%s["'][^>]*\bcontent\s*=\s*["']([^"']*)["']`,
	`(?is)<meta\b[^>]*\bcontent\s*=\s*["']([^"']*)["'][^>]*\bname\s*=\s*["']%s["']`,
}

// readOPFMeta returns the content of the <meta> element with the given name.
func readOPFMeta(inner, name string) string {
	quoted := regexp.QuoteMeta(name)
	for _, tpl := range opfMetaPatterns {
		re := regexp.MustCompile(fmt.Sprintf(tpl, quoted))
		if m := re.FindStringSubmatch(inner); m != nil {
			return unescapeXML(strings.TrimSpace(m[1]))
		}
	}
	return ""
}

var opfISBNRe = regexp.MustCompile(`(?is)<dc:identifier\b[^>]*isbn[^>]*>(.*?)</dc:identifier>`)

// readISBN returns the ISBN identifier, if the OPF carries one.
func readISBN(inner string) string {
	if m := opfISBNRe.FindStringSubmatch(inner); m != nil {
		return unescapeXML(strings.TrimSpace(m[1]))
	}
	return ""
}

// optString returns nil for an empty value, so "absent" stays distinguishable
// from "present and empty" on the way back out.
func optString(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	v := s
	return &v
}

func firstOf(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
