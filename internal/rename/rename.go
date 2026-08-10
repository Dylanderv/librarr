// Package rename computes how a series' folder and files should be named once
// its metadata is known, so the same plan can be previewed and applied.
//
// Kavita derives a series' name from the folder and file names when the
// embedded metadata does not win, so a torrent-derived name like
// "Inio.Asano.Anthology.OS.Asano.FR.-TONER-Paprika+" keeps showing up in the
// library however good the ComicInfo is. Renaming to the metadata's series name
// fixes it at the source.
//
// Nothing here touches the filesystem: callers get a Plan they can show to the
// user and then apply.
package rename

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Mode selects how much of the layout to rename.
type Mode string

const (
	// ModeNone leaves the filesystem alone.
	ModeNone Mode = "none"
	// ModeFolder renames only the series folder.
	ModeFolder Mode = "folder"
	// ModeFolderAndFiles renames the folder and every file inside it.
	ModeFolderAndFiles Mode = "folder_and_files"
)

// ParseMode validates a mode coming from an API payload.
func ParseMode(s string) (Mode, error) {
	switch Mode(s) {
	case ModeNone, ModeFolder, ModeFolderAndFiles:
		return Mode(s), nil
	case "":
		return ModeNone, nil
	default:
		return "", fmt.Errorf("unknown rename mode %q", s)
	}
}

// FileChange is one file to rename, with absolute paths.
type FileChange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// Plan describes every change for one series. Empty From/To on the folder means
// the folder is not being renamed. Skipped explains, per file, why it was left
// alone — surfacing that is the point: a silent partial rename is worse than a
// visible one.
type Plan struct {
	FolderFrom string       `json:"folder_from,omitempty"`
	FolderTo   string       `json:"folder_to,omitempty"`
	Files      []FileChange `json:"files"`
	Skipped    []string     `json:"skipped,omitempty"`
	Warnings   []string     `json:"warnings,omitempty"`
}

// IsEmpty reports whether the plan would change nothing on disk.
func (p Plan) IsEmpty() bool {
	return p.FolderFrom == "" && len(p.Files) == 0
}

// Volume numbering as it actually appears in scene releases, most explicit
// first. A bare trailing number is the last resort and the most error-prone,
// which is why it only matches at the very end of the name.
var volumePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bvol(?:ume)?\.?\s*(\d{1,4})\b`),
	regexp.MustCompile(`(?i)\bt(?:ome)?\.?\s*(\d{1,4})\b`),
	regexp.MustCompile(`(?i)\bv\.?\s*(\d{1,4})\b`),
	regexp.MustCompile(`(?i)\b(?:c|ch|chapter)\.?\s*(\d{1,4})\b`),
	// Capped at three digits so a trailing year ("… Destruction 2001") is not
	// mistaken for volume 2001.
	regexp.MustCompile(`\b(\d{1,3})\s*$`),
}

// VolumeNumber extracts a volume number from a file name, without its
// extension. It reports false when no number can be read, which callers must
// treat as "do not rename this file" rather than guessing.
func VolumeNumber(name string) (int, bool) {
	stem := strings.TrimSuffix(name, filepath.Ext(name))
	// Bracketed groups hold release metadata ([2024], [TONER]) whose digits are
	// not volume numbers.
	stem = bracketRe.ReplaceAllString(stem, " ")
	// Torrent-derived names separate words with "_" or ".", which defeats the
	// word boundaries below: "_" is a word character, so "\bt" never matches in
	// "Boruto_Two_Blue_Vortex_T4". Normalising to spaces is safe because the
	// extension is already gone and every pattern tolerates a space where it
	// used to see a dot ("Vol.03" becomes "Vol 03").
	stem = separatorRe.ReplaceAllString(stem, " ")

	for _, re := range volumePatterns {
		if m := re.FindStringSubmatch(stem); m != nil {
			if n, err := strconv.Atoi(m[1]); err == nil && n > 0 {
				return n, true
			}
		}
	}
	return 0, false
}

var (
	bracketRe    = regexp.MustCompile(`[\[\(][^\]\)]*[\]\)]`)
	separatorRe  = regexp.MustCompile(`[._]+`)
	unsafeRe     = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	whitespaceRe = regexp.MustCompile(`\s+`)
)

// SanitizeSegment makes a single path segment safe on every platform librarr
// runs on, and trims what would confuse a scanner (leading/trailing dots and
// spaces). It never returns an empty string.
func SanitizeSegment(name string) string {
	name = unsafeRe.ReplaceAllString(name, "")
	name = whitespaceRe.ReplaceAllString(name, " ")
	name = strings.TrimSpace(name)
	name = strings.Trim(name, ".")
	name = strings.TrimSpace(name)
	if len(name) > 120 {
		name = strings.TrimSpace(name[:120])
	}
	if name == "" {
		return "Unknown"
	}
	return name
}

// FileName builds the target file name for a series file. Multi-file series get
// "<Series> Vol.NN.ext" — the layout Kavita's scanner reads most reliably —
// while a single-file series drops the volume entirely, since "Vol.01" on a
// standalone anthology is noise.
func FileName(series, original string, singleFile bool) (string, bool) {
	ext := filepath.Ext(original)
	safe := SanitizeSegment(series)

	if singleFile {
		if _, ok := VolumeNumber(original); !ok {
			return safe + ext, true
		}
	}

	n, ok := VolumeNumber(original)
	if !ok {
		return "", false
	}
	width := 2
	if n > 99 {
		width = 3
	}
	return fmt.Sprintf("%s Vol.%0*d%s", safe, width, n, ext), true
}

// BuildPlan works out the changes for a series whose files live at the given
// absolute paths. It renames the folder only when every file shares one parent:
// a series spread over several directories has no single folder to rename, and
// picking one would move files around.
func BuildPlan(seriesName string, files []string, mode Mode) Plan {
	var p Plan
	if mode == ModeNone || len(files) == 0 || strings.TrimSpace(seriesName) == "" {
		return p
	}

	safe := SanitizeSegment(seriesName)

	parents := make(map[string]bool)
	for _, f := range files {
		parents[filepath.Dir(f)] = true
	}

	folder := ""
	if len(parents) == 1 {
		for d := range parents {
			folder = d
		}
	} else {
		p.Warnings = append(p.Warnings,
			fmt.Sprintf("files span %d folders, folder left unchanged", len(parents)))
	}

	// Where the files will live once the folder move is accounted for, so file
	// renames are expressed against their final directory.
	destDir := folder
	if folder != "" && filepath.Base(folder) != safe {
		p.FolderFrom = folder
		p.FolderTo = filepath.Join(filepath.Dir(folder), safe)
		destDir = p.FolderTo
	}

	if mode != ModeFolderAndFiles {
		return p
	}

	singleFile := len(files) == 1
	taken := make(map[string]string, len(files))

	for _, f := range files {
		base := filepath.Base(f)
		target, ok := FileName(seriesName, base, singleFile)
		if !ok {
			p.Skipped = append(p.Skipped,
				fmt.Sprintf("%s: no volume number found, left as is", base))
			continue
		}
		if target == base {
			continue
		}
		if prev, clash := taken[target]; clash {
			p.Skipped = append(p.Skipped,
				fmt.Sprintf("%s: would collide with %s as %q", base, prev, target))
			continue
		}
		taken[target] = base

		// The source is the current path; the folder rename, if any, happens
		// first, so express the destination under destDir.
		p.Files = append(p.Files, FileChange{
			From: f,
			To:   filepath.Join(destDir, target),
		})
	}

	return p
}
