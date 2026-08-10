// Package filewriter writes book metadata directly into the media files
// (EPUB OPF and CBZ ComicInfo.xml) rather than into Kavita's database, so the
// metadata travels with the file and is the single source of truth. After
// writing, callers trigger a Kavita rescan to pick the changes up.
//
// A failure never leaves a corrupted book behind. CBZs are patched in place by
// appending to the archive, which leaves every existing byte untouched (see
// zippatch.go); everything else is rebuilt into a temp file in the same
// directory and renamed over the original.
package filewriter

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Fields describes a metadata edit. Pointer fields distinguish "set to this
// value" (non-nil) from "leave unchanged" (nil). Slice fields are applied only
// when non-nil; AppendLists controls whether they union with or replace the
// file's existing values.
type Fields struct {
	Title       *string
	Series      *string
	SeriesIndex *string
	Summary     *string
	Publisher   *string
	Language    *string
	ISBN        *string

	Authors []string
	Genres  []string
	Tags    []string

	// AppendLists, when true, merges Authors/Genres/Tags with what's already
	// in the file instead of replacing them.
	AppendLists bool
}

// IsEmpty reports whether the edit would change nothing.
func (f Fields) IsEmpty() bool {
	return f.Title == nil && f.Series == nil && f.SeriesIndex == nil &&
		f.Summary == nil && f.Publisher == nil && f.Language == nil && f.ISBN == nil &&
		f.Authors == nil && f.Genres == nil && f.Tags == nil
}

// ErrUnsupported is returned for file types we can't safely edit (e.g. CBR,
// which is a RAR archive and cannot be rewritten with the stdlib zip writer).
type ErrUnsupported struct{ Ext string }

func (e ErrUnsupported) Error() string {
	return fmt.Sprintf("unsupported file type for metadata write: %q", e.Ext)
}

// Apply writes the given fields into the file at path, dispatching on its
// extension. It is atomic and leaves the original untouched on error.
func Apply(path string, f Fields) error {
	if f.IsEmpty() {
		return nil
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".epub":
		return applyEPUB(path, f)
	case ".cbz", ".zip":
		return applyCBZ(path, f)
	default:
		return ErrUnsupported{Ext: strings.ToLower(filepath.Ext(path))}
	}
}

// rewriteZip rebuilds the zip at path, replacing the contents of entries named
// in replace and appending any add entries whose names aren't already present.
// Entry order and per-entry compression are preserved (important for EPUB's
// leading stored "mimetype" entry). The write is atomic via temp file + rename.
func rewriteZip(path string, replace map[string][]byte, add map[string][]byte) error {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer zr.Close()

	tmp, err := os.CreateTemp(filepath.Dir(path), ".librarr-meta-*"+filepath.Ext(path))
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpName := tmp.Name()
	// Best-effort cleanup if we bail before the rename succeeds.
	success := false
	defer func() {
		tmp.Close()
		if !success {
			os.Remove(tmpName)
		}
	}()

	// Buffer the temp file: io.Copy's 32 KB chunks turn into 32 KB writes on a
	// bare *os.File, which is a poor fit for network storage negotiating much
	// larger transfer sizes.
	bw := bufio.NewWriterSize(tmp, 4<<20)
	zw := zip.NewWriter(bw)
	present := make(map[string]bool, len(zr.File))

	for _, f := range zr.File {
		present[f.Name] = true
		// Preserve the original header (name, method, modtime) so structural
		// requirements like EPUB's stored mimetype entry survive the rewrite.
		hdr := f.FileHeader

		if data, ok := replace[f.Name]; ok {
			w, err := zw.CreateHeader(&hdr)
			if err != nil {
				zw.Close()
				return fmt.Errorf("write header %s: %w", f.Name, err)
			}
			if _, err := w.Write(data); err != nil {
				zw.Close()
				return fmt.Errorf("write %s: %w", f.Name, err)
			}
			continue
		}

		// Entries we're not changing are copied compressed-as-is: decoding and
		// re-encoding them would burn CPU to produce the same bytes.
		w, err := zw.CreateRaw(&hdr)
		if err != nil {
			zw.Close()
			return fmt.Errorf("write header %s: %w", f.Name, err)
		}
		// OpenRaw yields a plain io.Reader — there is nothing to close.
		rc, err := f.OpenRaw()
		if err != nil {
			zw.Close()
			return fmt.Errorf("open entry %s: %w", f.Name, err)
		}
		if _, err := io.Copy(w, rc); err != nil {
			zw.Close()
			return fmt.Errorf("copy %s: %w", f.Name, err)
		}
	}

	for name, data := range add {
		if present[name] {
			continue
		}
		w, err := zw.Create(name)
		if err != nil {
			zw.Close()
			return fmt.Errorf("create %s: %w", name, err)
		}
		if _, err := w.Write(data); err != nil {
			zw.Close()
			return fmt.Errorf("write %s: %w", name, err)
		}
	}

	if err := zw.Close(); err != nil {
		return fmt.Errorf("finalize zip: %w", err)
	}
	if err := bw.Flush(); err != nil {
		return fmt.Errorf("flush temp: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("sync temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}

	// os.CreateTemp opens at 0600, which would leave the rewritten file
	// unreadable by the media server. Match the original file mode; if it cannot
	// be read, fall back to something readable rather than keeping 0600.
	if info, err := os.Stat(path); err == nil {
		os.Chmod(tmpName, info.Mode())
	} else {
		os.Chmod(tmpName, 0o644)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("rename temp over original: %w", err)
	}
	success = true
	return nil
}

// readZipEntry returns the bytes of a single named entry, or (nil, nil) if the
// entry does not exist.
func readZipEntry(path, name string) ([]byte, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	for _, f := range zr.File {
		if f.Name == name {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}
	return nil, nil
}

// mergeList applies a desired list against existing values honoring append
// mode, de-duplicating case-insensitively while preserving order.
func mergeList(existing, desired []string, appendMode bool) []string {
	var base []string
	if appendMode {
		base = append(base, existing...)
	}
	base = append(base, desired...)

	seen := make(map[string]bool)
	var out []string
	for _, v := range base {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		key := strings.ToLower(v)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, v)
	}
	return out
}
