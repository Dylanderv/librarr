package rename

import (
	"path/filepath"
	"testing"
)

func TestVolumeNumber(t *testing.T) {
	tests := []struct {
		name string
		want int
		ok   bool
	}{
		{"Boruto Vol.5.cbz", 5, true},
		{"Boruto Volume 12.cbz", 12, true},
		{"Choujin X v03.cbz", 3, true},
		{"Bonne Nuit Punpun T7.cbz", 7, true},
		{"Bonne Nuit Punpun Tome 8.cbz", 8, true},
		{"Choujin X - c045.cbz", 45, true},
		{"Mujina Into The Deep 02.cbz", 2, true},
		// A trailing year must not read as a volume number.
		{"Dead Dead Demons DeDeDeDe Destruction 2001.cbz", 0, false},
		// Release metadata in brackets is not volume numbering.
		{"Inio.Asano.Anthology.OS.Asano.FR.[CBZ]-TONER-Paprika+.cbz", 0, false},
		{"Some Series [2024] (Digital).cbz", 0, false},
		{"no digits at all.cbz", 0, false},
	}

	for _, tc := range tests {
		got, ok := VolumeNumber(tc.name)
		if ok != tc.ok || got != tc.want {
			t.Errorf("VolumeNumber(%q) = (%d, %v), want (%d, %v)", tc.name, got, ok, tc.want, tc.ok)
		}
	}
}

func TestFileName(t *testing.T) {
	tests := []struct {
		series, original string
		single           bool
		want             string
		ok               bool
	}{
		{"Boruto", "Boruto Vol.5.cbz", false, "Boruto Vol.05.cbz", true},
		{"Boruto", "boruto v120.cbz", false, "Boruto Vol.120.cbz", true},
		// A single unnumbered file drops the volume rather than inventing one.
		{"Bakemono Recchan", "Inio.Asano.Anthology.[CBZ]-TONER.cbz", true, "Bakemono Recchan.cbz", true},
		// The same file in a multi-file series cannot be named safely.
		{"Bakemono Recchan", "Inio.Asano.Anthology.[CBZ]-TONER.cbz", false, "", false},
		// Illegal characters in the series name are stripped.
		{"Boruto: Two Blue Vortex", "vol 1.cbz", false, "Boruto Two Blue Vortex Vol.01.cbz", true},
	}

	for _, tc := range tests {
		got, ok := FileName(tc.series, tc.original, tc.single)
		if ok != tc.ok || got != tc.want {
			t.Errorf("FileName(%q, %q, %v) = (%q, %v), want (%q, %v)",
				tc.series, tc.original, tc.single, got, ok, tc.want, tc.ok)
		}
	}
}

func TestSanitizeSegment(t *testing.T) {
	tests := map[string]string{
		"Boruto: Naruto Next Generations": "Boruto Naruto Next Generations",
		"  spaced  out  ":                 "spaced out",
		"../../etc/passwd":                "etcpasswd",
		"trailing dots...":                "trailing dots",
		"":                                "Unknown",
		"///":                             "Unknown",
	}
	for in, want := range tests {
		if got := SanitizeSegment(in); got != want {
			t.Errorf("SanitizeSegment(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBuildPlanFolderOnly(t *testing.T) {
	files := []string{
		"/books/manga/Inio.Asano.Anthology-TONER/vol 1.cbz",
		"/books/manga/Inio.Asano.Anthology-TONER/vol 2.cbz",
	}
	p := BuildPlan("Bakemono Recchan", files, ModeFolder)

	if p.FolderFrom != "/books/manga/Inio.Asano.Anthology-TONER" {
		t.Errorf("FolderFrom = %q", p.FolderFrom)
	}
	if p.FolderTo != "/books/manga/Bakemono Recchan" {
		t.Errorf("FolderTo = %q", p.FolderTo)
	}
	if len(p.Files) != 0 {
		t.Errorf("folder-only mode planned %d file renames", len(p.Files))
	}
}

func TestBuildPlanFolderAndFiles(t *testing.T) {
	files := []string{
		"/books/manga/Old.Name/Old.Name v1.cbz",
		"/books/manga/Old.Name/Old.Name v2.cbz",
	}
	p := BuildPlan("Bakemono Recchan", files, ModeFolderAndFiles)

	if p.FolderTo != "/books/manga/Bakemono Recchan" {
		t.Fatalf("FolderTo = %q", p.FolderTo)
	}
	if len(p.Files) != 2 {
		t.Fatalf("planned %d file renames, want 2: %+v", len(p.Files), p.Files)
	}
	// Destinations must sit in the renamed folder, not the old one.
	for _, f := range p.Files {
		if filepath.Dir(f.To) != p.FolderTo {
			t.Errorf("%q lands outside the renamed folder", f.To)
		}
	}
	if p.Files[0].To != "/books/manga/Bakemono Recchan/Bakemono Recchan Vol.01.cbz" {
		t.Errorf("Files[0].To = %q", p.Files[0].To)
	}
}

func TestBuildPlanSkipsUnnumberedInMultiFileSeries(t *testing.T) {
	files := []string{
		"/books/manga/Old/Old v1.cbz",
		"/books/manga/Old/bonus chapter.cbz",
	}
	p := BuildPlan("Series", files, ModeFolderAndFiles)

	if len(p.Files) != 1 {
		t.Errorf("planned %d renames, want 1", len(p.Files))
	}
	if len(p.Skipped) != 1 {
		t.Fatalf("Skipped = %v, want one entry explaining the untouched file", p.Skipped)
	}
}

func TestBuildPlanCollision(t *testing.T) {
	// Two files that reduce to the same volume number must not both be renamed.
	files := []string{
		"/books/manga/Old/Old v1.cbz",
		"/books/manga/Old/Old vol 1 (repack).cbz",
	}
	p := BuildPlan("Series", files, ModeFolderAndFiles)

	if len(p.Files) != 1 {
		t.Errorf("planned %d renames, want 1", len(p.Files))
	}
	if len(p.Skipped) != 1 {
		t.Errorf("Skipped = %v, want the colliding file reported", p.Skipped)
	}
}

func TestBuildPlanMultipleFoldersLeavesFolderAlone(t *testing.T) {
	files := []string{
		"/books/manga/A/vol 1.cbz",
		"/books/manga/B/vol 2.cbz",
	}
	p := BuildPlan("Series", files, ModeFolderAndFiles)

	if p.FolderFrom != "" || p.FolderTo != "" {
		t.Errorf("planned a folder rename for a series spanning folders: %q -> %q", p.FolderFrom, p.FolderTo)
	}
	if len(p.Warnings) == 0 {
		t.Error("want a warning explaining why the folder was left alone")
	}
	// Files still get renamed, in place.
	if len(p.Files) != 2 {
		t.Errorf("planned %d file renames, want 2", len(p.Files))
	}
}

func TestBuildPlanNoopWhenAlreadyNamed(t *testing.T) {
	files := []string{"/books/manga/Series/Series Vol.01.cbz"}
	p := BuildPlan("Series", files, ModeFolderAndFiles)
	if !p.IsEmpty() {
		t.Errorf("plan should be empty, got %+v", p)
	}
}

func TestParseMode(t *testing.T) {
	for _, in := range []string{"none", "folder", "folder_and_files", ""} {
		if _, err := ParseMode(in); err != nil {
			t.Errorf("ParseMode(%q) errored: %v", in, err)
		}
	}
	if _, err := ParseMode("everything"); err == nil {
		t.Error("ParseMode should reject unknown modes")
	}
}
