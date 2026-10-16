// Package convert migrates comic archives librarr cannot edit into ones it
// can.
//
// A CBR is a RAR archive, which the stdlib cannot rewrite — so metadata edits
// stop at "rename only". The way out is the manual workflow users already do
// (unpack the RAR, repack it as ZIP, rename to .cbz), automated here with a
// pure-Go RAR reader: no external unrar binary, nothing to install in the
// Docker image, and it works on the NAS exactly like everywhere else.
//
// Only stored and compressed entries are carried over; directories, links and
// encrypted entries are refused or skipped with a reason, never guessed at.
// The source file is left alone — deleting it is the caller's decision.
package convert

import (
	"archive/zip"
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	rardecode "github.com/nwaples/rardecode/v2"
)

// Conversion failures the caller maps to user-facing answers.
var (
	// ErrExists: a file already occupies the .cbz name — never overwritten.
	ErrExists = errors.New("a file already has the converted name")
	// ErrEmpty: the archive holds no file entry worth carrying over.
	ErrEmpty = errors.New("archive holds no files to convert")
	// ErrUnsafe: an entry would land outside the archive's own folder.
	ErrUnsafe = errors.New("archive holds an entry that escapes its folder")
	// ErrEncrypted: an entry needs a password librarr never asks for.
	ErrEncrypted = errors.New("archive holds an encrypted entry")
)

// Guards against archive bombs: a hostile or corrupt CBR must fail the
// conversion, not the server.
const (
	// maxConvertBytes caps the total unpacked output of one conversion.
	maxConvertBytes = 8 << 30
	// maxConvertEntries caps the entry count (and the CBZ central directory).
	maxConvertEntries = 10000
)

// Result describes a finished conversion.
type Result struct {
	// Path is the new .cbz file, next to the source.
	Path string
	// Files is the number of entries carried over.
	Files int
	// Skipped names the entries left behind (directories, links).
	Skipped []string
	// Bytes is the total unpacked payload written.
	Bytes int64
}

// CBRToCBZ repacks the RAR archive at src as a ZIP archive with a .cbz
// extension, next to the source. Entry names, contents and modification times
// are preserved; compression is uniform Deflate. Atomic: the output appears
// under its final name only once fully written, and a failure leaves no
// half-written file behind.
func CBRToCBZ(src string) (Result, error) {
	var res Result

	info, err := os.Stat(src)
	if err != nil {
		return res, err
	}
	if info.IsDir() {
		return res, fmt.Errorf("not a file: %s", src)
	}

	dst := strings.TrimSuffix(src, filepath.Ext(src)) + ".cbz"
	if _, err := os.Stat(dst); err == nil {
		return res, ErrExists
	}

	raw, err := os.Open(src)
	if err != nil {
		return res, err
	}
	defer raw.Close()

	r, err := rardecode.NewReader(raw)
	if err != nil {
		return res, fmt.Errorf("not a RAR archive: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(src), ".librarr-convert-*.cbz")
	if err != nil {
		return res, fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	success := false
	defer func() {
		tmp.Close()
		if !success {
			os.Remove(tmpName)
		}
	}()

	// Buffered like filewriter.rewriteZip: 32 KB io.Copy chunks are a poor
	// fit for network storage negotiating much larger transfer sizes.
	bw := bufio.NewWriterSize(tmp, 4<<20)
	zw := zip.NewWriter(bw)

	seen := map[string]bool{}
	var total int64
	for {
		fh, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return res, fmt.Errorf("read archive: %w", err)
		}
		// No password UI exists, so an encrypted entry fails the whole
		// conversion rather than silently dropping the file.
		if fh.Encrypted || fh.HeaderEncrypted {
			return res, fmt.Errorf("%w: %s", ErrEncrypted, fh.Name)
		}
		if fh.LinkType != 0 {
			res.Skipped = append(res.Skipped, fh.Name+" (link)")
			continue
		}
		if fh.IsDir {
			res.Skipped = append(res.Skipped, fh.Name)
			continue
		}
		name, err := safeEntryName(fh.Name)
		if err != nil {
			return res, fmt.Errorf("%w: %s", ErrUnsafe, fh.Name)
		}
		if seen[name] {
			return res, fmt.Errorf("archive holds %q twice", name)
		}
		seen[name] = true
		if len(seen) > maxConvertEntries {
			return res, fmt.Errorf("too many entries in one archive (max %d)", maxConvertEntries)
		}
		if !fh.UnKnownSize && fh.UnPackedSize > maxConvertBytes {
			return res, fmt.Errorf("entry %q exceeds the %d-byte limit", name, maxConvertBytes)
		}

		w, err := zw.CreateHeader(&zip.FileHeader{
			Name:     name,
			Method:   zip.Deflate,
			Modified: fh.ModificationTime,
		})
		if err != nil {
			return res, fmt.Errorf("write entry %s: %w", name, err)
		}
		n, err := io.Copy(w, io.LimitReader(r, maxConvertBytes-total+1))
		total += n
		if err != nil {
			return res, fmt.Errorf("unpack entry %s: %w", name, err)
		}
		if total > maxConvertBytes {
			return res, fmt.Errorf("unpacked output exceeds the %d-byte limit", maxConvertBytes)
		}
		res.Files++
	}

	if res.Files == 0 {
		return res, ErrEmpty
	}

	if err := zw.Close(); err != nil {
		return res, fmt.Errorf("finish archive: %w", err)
	}
	if err := bw.Flush(); err != nil {
		return res, fmt.Errorf("flush archive: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return res, fmt.Errorf("close archive: %w", err)
	}
	if err := os.Rename(tmpName, dst); err != nil {
		return res, fmt.Errorf("publish archive: %w", err)
	}
	success = true

	res.Path = dst
	res.Bytes = total
	return res, nil
}

// safeEntryName keeps an archive entry inside its own folder. The reader
// reports '/' separators already; anything absolute or climbing out with ".."
// fails the conversion instead of writing where it should not.
func safeEntryName(name string) (string, error) {
	cleaned := path.Clean(strings.ReplaceAll(name, "\\", "/"))
	if cleaned == "" || cleaned == "." ||
		path.IsAbs(cleaned) || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", ErrUnsafe
	}
	return cleaned, nil
}
