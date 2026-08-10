package api

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/JeremiahM37/librarr/internal/config"
)

// Path resolution for the /api/files/* surface.
//
// This is a deliberately separate path from validateAllowedPath in
// pathvalidate.go: the old route keeps its own behaviour untouched. Two things
// differ here.
//
//  1. Root symlink resolution is cached. pathvalidate.go calls EvalSymlinks on
//     every configured root on every request; storage is a slow NFS mount, so
//     that is a handful of avoidable network round-trips per HTTP call.
//  2. A bare root is a valid target for reads (listing a library) but never for
//     writes: renaming /books/manga would take the whole library with it. The
//     distinction is carried by two entry points, resolveLibraryPath (read) and
//     resolveLibraryPathForWrite (write), rather than a boolean the caller could
//     forget to test.

// libraryRoot is one browsable library. ID is URL-safe and stable, so the front
// end and the API can refer to a library without passing a filesystem path.
type libraryRoot struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Path  string `json:"path"`
}

var (
	errLibraryPathRequired = errors.New("path is required")
	errLibraryPathRelative = errors.New("path must be absolute")
	errLibraryPathOutside  = errors.New("path is outside the configured libraries")
	errLibraryRootReadOnly = errors.New("a library root cannot be modified")
	errNoLibraryRoots      = errors.New("no library directory is configured")
)

// rootResolveCache memoises EvalSymlinks per configured root directory. Only
// successful resolutions are cached: a root that does not exist yet must stay
// discoverable if it is created later, whereas a resolved root moving under a
// running process is not a case worth paying an NFS stat per request for.
var (
	rootResolveMu    sync.Mutex
	rootResolveCache = map[string]string{}
)

// resolveRootDir returns the real (symlink-free) absolute path of a configured
// root, and false when it is unset or is not an existing directory.
func resolveRootDir(dir string) (string, bool) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return "", false
	}

	rootResolveMu.Lock()
	cached, ok := rootResolveCache[dir]
	rootResolveMu.Unlock()
	if ok {
		return cached, true
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", false
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", false
	}
	info, err := os.Stat(real)
	if err != nil || !info.IsDir() {
		return "", false
	}

	rootResolveMu.Lock()
	rootResolveCache[dir] = real
	rootResolveMu.Unlock()
	return real, true
}

// resetLibraryRootCache drops the memoised root resolutions. Tests only.
func resetLibraryRootCache() {
	rootResolveMu.Lock()
	rootResolveCache = map[string]string{}
	rootResolveMu.Unlock()
}

// libraryRoots lists the configured libraries that actually exist on disk.
func libraryRoots(cfg *config.Config) []libraryRoot {
	if cfg == nil {
		return nil
	}
	candidates := []struct {
		id    string
		label string
		dir   string
	}{
		{"manga", "Manga", cfg.MangaDir},
		{"ebooks", "Ebooks", cfg.EbookDir},
		{"audiobooks", "Audiobooks", cfg.AudiobookDir},
	}

	var roots []libraryRoot
	for _, c := range candidates {
		real, ok := resolveRootDir(c.dir)
		if !ok {
			continue
		}
		roots = append(roots, libraryRoot{ID: c.id, Label: c.label, Path: real})
	}
	return roots
}

// resolveLibraryPath resolves a path received over HTTP into a local path safe
// to read. It returns the cleaned path and whether that path is a bare library
// root (the caller may want to forbid some operations on it — see
// resolveLibraryPathForWrite).
//
// Confinement is lexical: Clean, then a prefix test on the separator so that
// /books/manga-other is not accepted under /books/manga. The requested path is
// deliberately *not* passed through EvalSymlinks here. That costs one NFS
// round-trip per request on the hot listing path, and the only thing it buys is
// blocking a symlink planted inside a library that points outside it — which
// requires write access to the NFS share, i.e. an actor who can already read and
// write those files directly. Writes take the opposite trade-off: they are rare
// and destructive, so resolveLibraryPathForWrite does pay for EvalSymlinks.
func resolveLibraryPath(cfg *config.Config, p string) (string, bool, error) {
	if strings.TrimSpace(p) == "" {
		return "", false, errLibraryPathRequired
	}
	roots := libraryRoots(cfg)
	if len(roots) == 0 {
		return "", false, errNoLibraryRoots
	}

	clean := filepath.Clean(p)
	if !filepath.IsAbs(clean) {
		return "", false, errLibraryPathRelative
	}

	for _, root := range roots {
		if clean == root.Path {
			return clean, true, nil
		}
		if strings.HasPrefix(clean, root.Path+string(filepath.Separator)) {
			return clean, false, nil
		}
	}
	return "", false, errLibraryPathOutside
}

// resolveLibraryPathForWrite resolves a path for an operation that modifies the
// filesystem (rename, metadata write). On top of resolveLibraryPath it refuses a
// bare library root, and it verifies the real path — following symlinks — is
// still confined, so a link pointing outside the libraries cannot be used as an
// escape hatch. The target itself may not exist yet (a rename destination), in
// which case the closest existing ancestor is checked instead.
func resolveLibraryPathForWrite(cfg *config.Config, p string) (string, error) {
	clean, isRoot, err := resolveLibraryPath(cfg, p)
	if err != nil {
		return "", err
	}
	if isRoot {
		return "", errLibraryRootReadOnly
	}
	if err := checkRealPathConfined(cfg, clean); err != nil {
		return "", err
	}
	return clean, nil
}

// checkRealPathConfined walks up to the closest existing ancestor of p, resolves
// symlinks on it, and verifies the result is still under a library root.
func checkRealPathConfined(cfg *config.Config, p string) error {
	roots := libraryRoots(cfg)
	candidate := p
	for {
		real, err := filepath.EvalSymlinks(candidate)
		if err == nil {
			// Re-append the part that does not exist yet: it cannot itself be a
			// symlink, so a lexical join is enough.
			suffix := strings.TrimPrefix(p, candidate)
			real = filepath.Clean(real + suffix)
			for _, root := range roots {
				if real == root.Path || strings.HasPrefix(real, root.Path+string(filepath.Separator)) {
					return nil
				}
			}
			return errLibraryPathOutside
		}
		parent := filepath.Dir(candidate)
		if parent == candidate {
			return errLibraryPathOutside
		}
		candidate = parent
	}
}
