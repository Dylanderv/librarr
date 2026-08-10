package api

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/JeremiahM37/librarr/internal/config"
)

// newRootsFixture builds a config with an existing manga root, an existing
// ebook root, a decoy sibling directory whose name shares the manga prefix, and
// a configured-but-missing audiobook root.
func newRootsFixture(t *testing.T) (*config.Config, string, string) {
	t.Helper()
	resetLibraryRootCache()
	t.Cleanup(resetLibraryRootCache)

	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("EvalSymlinks: %v", err)
	}
	manga := filepath.Join(base, "manga")
	ebooks := filepath.Join(base, "ebooks")
	decoy := filepath.Join(base, "manga-other")
	for _, d := range []string{manga, ebooks, decoy} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("MkdirAll %s: %v", d, err)
		}
	}
	cfg := &config.Config{
		MangaDir:     manga,
		EbookDir:     ebooks,
		AudiobookDir: filepath.Join(base, "missing-audiobooks"),
	}
	return cfg, manga, base
}

func TestLibraryRootsKeepsOnlyExistingDirs(t *testing.T) {
	cfg, manga, _ := newRootsFixture(t)

	roots := libraryRoots(cfg)
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots, got %d (%+v)", len(roots), roots)
	}
	if roots[0].ID != "manga" || roots[0].Path != manga {
		t.Errorf("unexpected first root: %+v", roots[0])
	}
	if roots[1].ID != "ebooks" {
		t.Errorf("unexpected second root: %+v", roots[1])
	}
	for _, r := range roots {
		if r.Label == "" {
			t.Errorf("root %s has no label", r.ID)
		}
	}
}

func TestLibraryRootsIgnoresFileAndEmptyConfig(t *testing.T) {
	resetLibraryRootCache()
	t.Cleanup(resetLibraryRootCache)

	base := t.TempDir()
	file := filepath.Join(base, "not-a-dir")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	cfg := &config.Config{MangaDir: file, EbookDir: "  "}
	if roots := libraryRoots(cfg); len(roots) != 0 {
		t.Fatalf("expected no roots, got %+v", roots)
	}
}

func TestResolveLibraryPathAcceptsPathUnderRoot(t *testing.T) {
	cfg, manga, _ := newRootsFixture(t)

	got, isRoot, err := resolveLibraryPath(cfg, filepath.Join(manga, "Berserk", "vol1.cbz"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isRoot {
		t.Error("expected isRoot=false for a nested path")
	}
	if want := filepath.Join(manga, "Berserk", "vol1.cbz"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolveLibraryPathRootIsReadableButNotWritable(t *testing.T) {
	cfg, manga, _ := newRootsFixture(t)

	got, isRoot, err := resolveLibraryPath(cfg, manga)
	if err != nil {
		t.Fatalf("reading a bare root must be allowed, got %v", err)
	}
	if !isRoot {
		t.Error("expected isRoot=true for a bare root")
	}
	if got != manga {
		t.Errorf("got %q, want %q", got, manga)
	}

	if _, err := resolveLibraryPathForWrite(cfg, manga); !errors.Is(err, errLibraryRootReadOnly) {
		t.Errorf("writing a bare root must be refused, got %v", err)
	}
	// Trailing separator must not sneak past the root check.
	if _, err := resolveLibraryPathForWrite(cfg, manga+string(filepath.Separator)); !errors.Is(err, errLibraryRootReadOnly) {
		t.Errorf("trailing separator: expected root refusal, got %v", err)
	}
}

func TestResolveLibraryPathRejectsTraversal(t *testing.T) {
	cfg, manga, base := newRootsFixture(t)

	cases := []string{
		filepath.Join(manga, "..", "..", "etc", "passwd"),
		manga + "/../manga-other/x.cbz",
		filepath.Join(base, "manga-other", "x.cbz"), // misleading prefix
		filepath.Join(base, "elsewhere"),
		"/etc/passwd",
	}
	for _, p := range cases {
		if _, _, err := resolveLibraryPath(cfg, p); !errors.Is(err, errLibraryPathOutside) {
			t.Errorf("resolveLibraryPath(%q): expected outside-roots error, got %v", p, err)
		}
	}
}

func TestResolveLibraryPathRejectsEmptyAndRelative(t *testing.T) {
	cfg, _, _ := newRootsFixture(t)

	for _, p := range []string{"", "   "} {
		if _, _, err := resolveLibraryPath(cfg, p); !errors.Is(err, errLibraryPathRequired) {
			t.Errorf("resolveLibraryPath(%q): expected required error, got %v", p, err)
		}
	}
	for _, p := range []string{"manga/Berserk", "./x", "../../etc"} {
		if _, _, err := resolveLibraryPath(cfg, p); !errors.Is(err, errLibraryPathRelative) {
			t.Errorf("resolveLibraryPath(%q): expected relative error, got %v", p, err)
		}
	}
}

func TestResolveLibraryPathWithoutConfiguredRoots(t *testing.T) {
	resetLibraryRootCache()
	t.Cleanup(resetLibraryRootCache)

	if _, _, err := resolveLibraryPath(&config.Config{}, "/books/manga"); !errors.Is(err, errNoLibraryRoots) {
		t.Errorf("expected no-roots error, got %v", err)
	}
	if _, _, err := resolveLibraryPath(nil, "/books/manga"); !errors.Is(err, errNoLibraryRoots) {
		t.Errorf("nil config: expected no-roots error, got %v", err)
	}
}

// A symlink inside a library pointing outside it is accepted for reads (the
// lexical check does not follow links, by design — see files_roots.go) but must
// be refused for writes.
func TestSymlinkEscapingRootIsRefusedForWrites(t *testing.T) {
	cfg, manga, base := newRootsFixture(t)

	outside := filepath.Join(base, "outside")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	link := filepath.Join(manga, "escape")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	if _, _, err := resolveLibraryPath(cfg, filepath.Join(link, "x.cbz")); err != nil {
		t.Errorf("read through symlink: expected lexical acceptance, got %v", err)
	}
	// Existing target through the link.
	target := filepath.Join(outside, "x.cbz")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if _, err := resolveLibraryPathForWrite(cfg, filepath.Join(link, "x.cbz")); !errors.Is(err, errLibraryPathOutside) {
		t.Errorf("write through symlink: expected outside-roots error, got %v", err)
	}
	// Target that does not exist yet: the closest existing ancestor is the link.
	if _, err := resolveLibraryPathForWrite(cfg, filepath.Join(link, "new.cbz")); !errors.Is(err, errLibraryPathOutside) {
		t.Errorf("write of a new file through symlink: expected outside-roots error, got %v", err)
	}
}

func TestResolveLibraryPathForWriteAllowsMissingTargetUnderRoot(t *testing.T) {
	cfg, manga, _ := newRootsFixture(t)

	dir := filepath.Join(manga, "Berserk")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	// Rename destination: neither the file nor its parent directory exists.
	p := filepath.Join(manga, "Berserk (2024)", "Berserk Vol.01.cbz")
	got, err := resolveLibraryPathForWrite(cfg, p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != p {
		t.Errorf("got %q, want %q", got, p)
	}
}
