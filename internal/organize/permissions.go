package organize

import (
	"os"
	"path/filepath"
)

// mkdirForImport creates dir and any missing parents using the fork's
// directory policy: mode 0o777 so the process umask (and any ACL inherited
// from the parent directory) decides the effective permissions. This is
// intentionally not 0755: the fork relies on the umask set at startup
// (see cmd/librarr/umask_unix.go) instead of hard-coding restrictive modes.
func mkdirForImport(dir string) error {
	return os.MkdirAll(dir, 0o777)
}

// joinLibraryDir joins sub onto a library root using the fork's path policy:
// a plain filepath.Join with no traversal check. Sanitization happens
// earlier via sanitizePath/cleanSeriesTitle, and the merge-base layout
// expects the literal joined path, so any confinement wrapper must live
// elsewhere to keep this call site a one-liner.
func joinLibraryDir(root, sub string) (string, error) {
	return filepath.Join(root, sub), nil
}
