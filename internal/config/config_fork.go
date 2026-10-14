package config

import (
	"errors"
	"path/filepath"
	"strings"
)

// Errors returned by ResolveKavitaPath.
var (
	errEmptyPath        = errors.New("empty path")
	errNoRoots          = errors.New("no Kavita library roots configured; set KAVITA_LIBRARY_PATH, KAVITA_MANGA_LIBRARY_PATH, or KAVITA_PATH_MAP")
	errPathOutsideRoots = errors.New("resolved path is outside allowed library roots")
)

// loadForkEnv applies fork-specific environment overrides to c.
// It is called by Load before settings-file overrides are applied,
// so values saved from the UI keep precedence over the environment,
// exactly as with the base loader.
func loadForkEnv(c *Config) {
	c.KavitaPathMap = getEnv("KAVITA_PATH_MAP", "")
}

// forkSettingsKeys returns the fork-specific settings-file keys pointing
// at fields on c. The caller merges them into its own key map.
func forkSettingsKeys(c *Config) map[string]*string {
	return map[string]*string{
		"kavita_path_map": &c.KavitaPathMap,
	}
}

// kavitaPathMappings parses KavitaPathMap into (from, to) prefix pairs.
func (c *Config) kavitaPathMappings() [][2]string {
	var pairs [][2]string
	for _, entry := range strings.Split(c.KavitaPathMap, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		// Split on the last colon so Windows-style "C:\..." targets survive;
		// in practice these are container POSIX paths.
		idx := strings.LastIndex(entry, ":")
		if idx <= 0 || idx == len(entry)-1 {
			continue
		}
		from := strings.TrimRight(strings.TrimSpace(entry[:idx]), "/")
		to := strings.TrimRight(strings.TrimSpace(entry[idx+1:]), "/")
		if from != "" && to != "" {
			pairs = append(pairs, [2]string{from, to})
		}
	}
	return pairs
}

// KavitaAllowedRoots returns the set of directories under which librarr is
// permitted to write book files. A resolved path outside every root is
// rejected as a safety measure against path traversal.
func (c *Config) KavitaAllowedRoots() []string {
	var roots []string
	add := func(p string) {
		if p = strings.TrimRight(strings.TrimSpace(p), "/"); p != "" {
			roots = append(roots, p)
		}
	}
	add(c.KavitaLibraryPath)
	add(c.KavitaMangaLibraryPath)
	for _, m := range c.kavitaPathMappings() {
		add(m[1])
	}
	return roots
}

// ResolveKavitaPath translates an absolute path reported by Kavita into a path
// librarr can write, applying KavitaPathMap and validating the result stays
// within an allowed root. It returns an error when no roots are configured
// (nothing to validate against) or when the path escapes them.
func (c *Config) ResolveKavitaPath(kavitaPath string) (string, error) {
	if kavitaPath == "" {
		return "", errEmptyPath
	}

	resolved := kavitaPath
	for _, m := range c.kavitaPathMappings() {
		if kavitaPath == m[0] || strings.HasPrefix(kavitaPath, m[0]+"/") {
			resolved = m[1] + kavitaPath[len(m[0]):]
			break
		}
	}
	resolved = filepath.Clean(resolved)

	roots := c.KavitaAllowedRoots()
	if len(roots) == 0 {
		return "", errNoRoots
	}
	for _, root := range roots {
		root = filepath.Clean(root)
		if resolved == root || strings.HasPrefix(resolved, root+string(filepath.Separator)) {
			return resolved, nil
		}
	}
	return "", errPathOutsideRoots
}
