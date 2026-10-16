package convert

import (
	"archive/zip"
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// PDF to CBZ conversion, via poppler's pdftoppm.
//
// Unlike the CBR repack (pure Go, lossless), a PDF must be rasterized: pages
// become JPEG images and the text layer is dropped. It is the right trade for
// scans and comics, and a destructive one for text books — the UI says so
// before applying. Rendering needs the poppler-utils package (pdftoppm and
// pdfinfo): apk/brew/apt install it; without it every conversion fails with
// ErrNoPoppler instead of a cryptic exec error.

// ErrNoPoppler reports a system without pdftoppm/pdfinfo on PATH.
var ErrNoPoppler = errors.New("poppler is not installed (pdftoppm/pdfinfo missing)")

// Guards: rendering is CPU- and disk-hungry, so a hostile or absurd PDF must
// fail fast instead of filling the disk.
const (
	// maxPDFPages caps the page count of one conversion.
	maxPDFPages = 2000
	// pdfMinDPI/pdfMaxDPI bound the configured rendering resolution.
	pdfMinDPI = 72
	pdfMaxDPI = 600
	// pdfMinQuality/pdfMaxQuality bound the configured JPEG quality.
	pdfMinQuality = 1
	pdfMaxQuality = 100
)

// PDFOptions tunes the rendering. Zero values select the defaults (200 DPI,
// JPEG quality 90) — the same defaults the server configures.
type PDFOptions struct {
	DPI         int
	JPEGQuality int
}

// normalized clamps the options into the supported ranges.
func (o PDFOptions) normalized() (dpi, quality int) {
	dpi, quality = o.DPI, o.JPEGQuality
	if dpi == 0 {
		dpi = 200
	}
	if quality == 0 {
		quality = 90
	}
	dpi = min(max(dpi, pdfMinDPI), pdfMaxDPI)
	quality = min(max(quality, pdfMinQuality), pdfMaxQuality)
	return dpi, quality
}

// HavePoppler reports whether both poppler tools are on PATH.
func HavePoppler() bool { return havePoppler() }

// havePoppler reports whether both poppler tools are on PATH.
func havePoppler() bool {
	_, err1 := exec.LookPath("pdftoppm")
	_, err2 := exec.LookPath("pdfinfo")
	return err1 == nil && err2 == nil
}

// pdfPageCount reads the page count from pdfinfo, refusing encrypted files:
// there is no password UI, so they fail with a reason instead of hanging on
// a prompt that can never be answered.
func pdfPageCount(src string) (int, error) {
	out, err := exec.Command("pdfinfo", src).Output()
	if err != nil {
		return 0, fmt.Errorf("pdfinfo: %w", err)
	}
	pages, encrypted := 0, ""
	for line := range strings.Lines(string(out)) {
		if v, ok := strings.CutPrefix(line, "Pages:"); ok {
			pages, _ = strconv.Atoi(strings.TrimSpace(v))
		}
		if v, ok := strings.CutPrefix(line, "Encrypted:"); ok {
			encrypted = strings.TrimSpace(v)
		}
	}
	if encrypted != "" && encrypted != "no" {
		return 0, ErrEncrypted
	}
	if pages <= 0 {
		return 0, fmt.Errorf("could not read the page count")
	}
	return pages, nil
}

// PDFToCBZ renders every page of the PDF at src to JPEG and packs them as a
// .cbz next to the source, pages in order (page-0001.jpg, …). Atomic like the
// CBR conversion: the output appears under its final name only once fully
// written, and a failure leaves no half-written file behind.
func PDFToCBZ(src string, opts PDFOptions) (Result, error) {
	var res Result

	info, err := os.Stat(src)
	if err != nil {
		return res, err
	}
	if info.IsDir() {
		return res, fmt.Errorf("not a file: %s", src)
	}
	if !havePoppler() {
		return res, ErrNoPoppler
	}

	dst := strings.TrimSuffix(src, filepath.Ext(src)) + ".cbz"
	if _, err := os.Stat(dst); err == nil {
		return res, ErrExists
	}

	pages, err := pdfPageCount(src)
	if err != nil {
		return res, err
	}
	if pages > maxPDFPages {
		return res, fmt.Errorf("too many pages (%d, max %d)", pages, maxPDFPages)
	}

	dpi, quality := opts.normalized()

	work, err := os.MkdirTemp("", ".librarr-pdf-*")
	if err != nil {
		return res, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(work)

	prefix := filepath.Join(work, "page")
	cmd := exec.Command("pdftoppm",
		"-jpeg", "-r", strconv.Itoa(dpi),
		"-jpegopt", "quality="+strconv.Itoa(quality),
		"-f", "1", "-l", strconv.Itoa(pages),
		src, prefix)
	if out, err := cmd.CombinedOutput(); err != nil {
		return res, fmt.Errorf("pdftoppm: %v: %s", err, firstLine(out))
	}

	rendered, err := collectRenderedPages(work, "page")
	if err != nil {
		return res, err
	}
	if len(rendered) == 0 {
		return res, ErrEmpty
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

	bw := bufio.NewWriterSize(tmp, 4<<20)
	zw := zip.NewWriter(bw)
	mod := info.ModTime()
	if mod.IsZero() {
		mod = time.Now()
	}
	var total int64
	for i, rp := range rendered {
		data, err := os.ReadFile(rp.path)
		if err != nil {
			return res, fmt.Errorf("read rendered page: %w", err)
		}
		total += int64(len(data))
		if total > maxConvertBytes {
			return res, fmt.Errorf("unpacked output exceeds the %d-byte limit", maxConvertBytes)
		}
		w, err := zw.CreateHeader(&zip.FileHeader{
			Name:     fmt.Sprintf("page-%04d.jpg", i+1),
			Method:   zip.Deflate,
			Modified: mod,
		})
		if err != nil {
			return res, fmt.Errorf("write entry: %w", err)
		}
		if _, err := w.Write(data); err != nil {
			return res, fmt.Errorf("write entry: %w", err)
		}
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
	res.Files = len(rendered)
	res.Bytes = total
	return res, nil
}

// renderedPage is one pdftoppm output with its page number for ordering:
// pdftoppm names files prefix-1.jpg … prefix-10.jpg, which sort wrong
// lexically, so the number is parsed and the zip is written in numeric order.
type renderedPage struct {
	path string
	num  int
}

func collectRenderedPages(dir, prefix string) ([]renderedPage, error) {
	raw, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("list rendered pages: %w", err)
	}
	var out []renderedPage
	for _, de := range raw {
		name := de.Name()
		rest, ok := strings.CutPrefix(name, prefix+"-")
		if !ok || !strings.HasSuffix(rest, ".jpg") {
			continue
		}
		n, err := strconv.Atoi(strings.TrimSuffix(rest, ".jpg"))
		if err != nil || n < 1 {
			continue
		}
		out = append(out, renderedPage{path: filepath.Join(dir, name), num: n})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].num < out[j].num })
	return out, nil
}

func firstLine(out []byte) string {
	s, _, _ := strings.Cut(string(out), "\n")
	if len(s) > 300 {
		s = s[:300]
	}
	return strings.TrimSpace(s)
}
