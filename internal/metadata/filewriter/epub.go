package filewriter

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
)

// applyEPUB edits the OPF package document embedded in an EPUB. It locates the
// OPF via META-INF/container.xml, performs surgical edits on the <metadata>
// block (preserving the rest of the OPF verbatim — manifest, spine, unknown
// elements), then rewrites the archive atomically.
func applyEPUB(path string, f Fields) error {
	container, err := readZipEntry(path, "META-INF/container.xml")
	if err != nil {
		return err
	}
	if container == nil {
		return fmt.Errorf("epub missing META-INF/container.xml")
	}

	opfPath, err := opfPathFromContainer(container)
	if err != nil {
		return err
	}

	opf, err := readZipEntry(path, opfPath)
	if err != nil {
		return err
	}
	if opf == nil {
		return fmt.Errorf("epub OPF not found at %q", opfPath)
	}

	edited, err := editOPF(opf, f)
	if err != nil {
		return err
	}

	return rewriteZip(path, map[string][]byte{opfPath: edited}, nil)
}

// opfPathFromContainer extracts the OPF rootfile path from container.xml.
func opfPathFromContainer(container []byte) (string, error) {
	var c struct {
		Rootfiles []struct {
			FullPath string `xml:"full-path,attr"`
		} `xml:"rootfiles>rootfile"`
	}
	if err := xml.Unmarshal(container, &c); err != nil {
		return "", fmt.Errorf("parse container.xml: %w", err)
	}
	if len(c.Rootfiles) == 0 || c.Rootfiles[0].FullPath == "" {
		return "", fmt.Errorf("container.xml has no rootfile")
	}
	return c.Rootfiles[0].FullPath, nil
}

var metadataOpenRe = regexp.MustCompile(`(?is)<([a-z0-9]+:)?metadata\b[^>]*>`)

// editOPF applies the requested field changes to an OPF document and returns
// the modified bytes. It only rewrites the inner content of the <metadata>
// element; everything else is preserved byte-for-byte.
func editOPF(opf []byte, f Fields) ([]byte, error) {
	s := string(opf)

	loc := metadataOpenRe.FindStringSubmatchIndex(s)
	if loc == nil {
		return nil, fmt.Errorf("OPF has no <metadata> element")
	}
	prefix := ""
	if loc[2] >= 0 {
		prefix = s[loc[2]:loc[3]]
	}
	openEnd := loc[1]
	closeTag := "</" + prefix + "metadata>"
	closeIdx := strings.Index(strings.ToLower(s[openEnd:]), strings.ToLower(closeTag))
	if closeIdx < 0 {
		return nil, fmt.Errorf("OPF has no closing %s", closeTag)
	}
	closeIdx += openEnd

	inner := s[openEnd:closeIdx]

	if f.Title != nil {
		inner = replaceDC(inner, "title", []string{*f.Title})
	}
	if f.Summary != nil {
		inner = replaceDC(inner, "description", []string{*f.Summary})
	}
	if f.Publisher != nil {
		inner = replaceDC(inner, "publisher", []string{*f.Publisher})
	}
	if f.Language != nil {
		inner = replaceDC(inner, "language", []string{*f.Language})
	}
	if f.Authors != nil {
		merged := mergeList(extractDC(inner, "creator"), f.Authors, f.AppendLists)
		inner = replaceDC(inner, "creator", merged)
	}
	if f.Genres != nil {
		merged := mergeList(extractDC(inner, "subject"), f.Genres, f.AppendLists)
		inner = replaceDC(inner, "subject", merged)
	}
	if f.Series != nil {
		inner = setCalibreMeta(inner, "calibre:series", *f.Series)
	}
	if f.SeriesIndex != nil {
		inner = setCalibreMeta(inner, "calibre:series_index", *f.SeriesIndex)
	}
	if f.ISBN != nil {
		inner = setISBN(inner, *f.ISBN)
	}

	return []byte(s[:openEnd] + inner + s[closeIdx:]), nil
}

// dcElementRe builds a regex matching a full <dc:local ...>...</dc:local>
// element (dot-all, case-insensitive), including leading whitespace so removals
// don't leave blank lines behind.
func dcElementRe(local string) *regexp.Regexp {
	return regexp.MustCompile(`(?is)[ \t]*<dc:` + regexp.QuoteMeta(local) + `\b[^>]*>.*?</dc:` + regexp.QuoteMeta(local) + `>\n?`)
}

// extractDC returns the (unescaped) text content of every <dc:local> element.
func extractDC(inner, local string) []string {
	re := regexp.MustCompile(`(?is)<dc:` + regexp.QuoteMeta(local) + `\b[^>]*>(.*?)</dc:` + regexp.QuoteMeta(local) + `>`)
	var out []string
	for _, m := range re.FindAllStringSubmatch(inner, -1) {
		out = append(out, unescapeXML(strings.TrimSpace(m[1])))
	}
	return out
}

// replaceDC removes all existing <dc:local> elements and appends one per value.
func replaceDC(inner, local string, values []string) string {
	inner = dcElementRe(local).ReplaceAllString(inner, "")
	var b strings.Builder
	b.WriteString(inner)
	if !strings.HasSuffix(inner, "\n") {
		b.WriteString("\n")
	}
	for _, v := range values {
		fmt.Fprintf(&b, "    <dc:%s>%s</dc:%s>\n", local, escapeXML(v), local)
	}
	return b.String()
}

// setCalibreMeta sets a <meta name="..." content="..."/> pair, replacing any
// existing one with the same name.
func setCalibreMeta(inner, name, content string) string {
	re := regexp.MustCompile(`(?is)[ \t]*<meta\b[^>]*\bname\s*=\s*["']` + regexp.QuoteMeta(name) + `["'][^>]*/?>\n?`)
	inner = re.ReplaceAllString(inner, "")
	if !strings.HasSuffix(inner, "\n") {
		inner += "\n"
	}
	return inner + fmt.Sprintf("    <meta name=%q content=%q/>\n", name, attrEscape(content))
}

// setISBN replaces any ISBN identifier with the given value.
func setISBN(inner, isbn string) string {
	re := regexp.MustCompile(`(?is)[ \t]*<dc:identifier\b[^>]*isbn[^>]*>.*?</dc:identifier>\n?`)
	inner = re.ReplaceAllString(inner, "")
	if !strings.HasSuffix(inner, "\n") {
		inner += "\n"
	}
	return inner + fmt.Sprintf("    <dc:identifier opf:scheme=\"ISBN\">%s</dc:identifier>\n", escapeXML(isbn))
}

func escapeXML(s string) string {
	var b strings.Builder
	xml.EscapeText(&b, []byte(s))
	return b.String()
}

// attrEscape escapes a value for use inside a double-quoted XML attribute.
func attrEscape(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return r.Replace(s)
}

func unescapeXML(s string) string {
	r := strings.NewReplacer("&amp;", "&", "&lt;", "<", "&gt;", ">", "&quot;", `"`, "&apos;", "'", "&#39;", "'")
	return r.Replace(s)
}
