package convert

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func requirePoppler(t *testing.T) {
	t.Helper()
	if !havePoppler() {
		t.Skip("poppler not installed (pdftoppm/pdfinfo missing)")
	}
}

func TestPDFToCBZRoundTrip(t *testing.T) {
	requirePoppler(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "scan.pdf")
	writeVectorPDF(t, src, 3)

	res, err := PDFToCBZ(src, PDFOptions{DPI: 72, JPEGQuality: 70})
	if err != nil {
		t.Fatalf("PDFToCBZ: %v", err)
	}
	if res.Path != filepath.Join(dir, "scan.cbz") {
		t.Errorf("path = %q", res.Path)
	}
	if res.Files != 3 {
		t.Errorf("files = %d, want 3", res.Files)
	}
	if _, err := os.Stat(src); err != nil {
		t.Error("source must survive conversion")
	}

	zr, err := zip.OpenReader(res.Path)
	if err != nil {
		t.Fatalf("output is not a valid zip: %v", err)
	}
	defer zr.Close()
	if len(zr.File) != 3 {
		t.Fatalf("entries = %d, want 3", len(zr.File))
	}
	seen := map[string][]byte{}
	for i, f := range zr.File {
		want := map[int]string{0: "page-0001.jpg", 1: "page-0002.jpg", 2: "page-0003.jpg"}[i]
		if f.Name != want {
			t.Errorf("entry %d = %q, want %q", i, f.Name, want)
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		var buf []byte
		tmp := make([]byte, 4096)
		for {
			n, err := rc.Read(tmp)
			buf = append(buf, tmp[:n]...)
			if err != nil {
				break
			}
		}
		rc.Close()
		if len(buf) == 0 {
			t.Errorf("%s is empty", f.Name)
		}
		if len(buf) < 2 || buf[0] != 0xFF || buf[1] != 0xD8 {
			t.Errorf("%s is not a JPEG", f.Name)
		}
		seen[f.Name] = buf
	}
	if string(seen["page-0001.jpg"]) == string(seen["page-0002.jpg"]) {
		t.Error("pages must render distinctly (order check)")
	}
}

func TestPDFToCBZRefusesMissingPoppler(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a.pdf")
	writeVectorPDF(t, src, 1)

	t.Setenv("PATH", dir) // nothing executable here, certainly no pdftoppm
	if _, err := PDFToCBZ(src, PDFOptions{}); err != ErrNoPoppler {
		t.Fatalf("err = %v, want ErrNoPoppler", err)
	}
}

func TestPDFToCBZRefusesGarbage(t *testing.T) {
	requirePoppler(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "fake.pdf")
	if err := os.WriteFile(src, []byte("not a pdf at all"), 0o666); err != nil {
		t.Fatal(err)
	}
	if _, err := PDFToCBZ(src, PDFOptions{DPI: 72}); err == nil {
		t.Fatal("expected an error for a non-PDF file")
	}
	if _, err := os.Stat(filepath.Join(dir, "fake.cbz")); !os.IsNotExist(err) {
		t.Error("no output must remain after a failed conversion")
	}
}

func TestPDFOptionsNormalized(t *testing.T) {
	dpi, q := PDFOptions{}.normalized()
	if dpi != 200 || q != 90 {
		t.Errorf("defaults = %d/%d, want 200/90", dpi, q)
	}
	dpi, q = PDFOptions{DPI: 10000, JPEGQuality: -5}.normalized()
	if dpi != 600 || q != 1 {
		t.Errorf("clamped = %d/%d, want 600/1", dpi, q)
	}
	dpi, q = PDFOptions{DPI: 50, JPEGQuality: 50}.normalized()
	if dpi != 72 || q != 50 {
		t.Errorf("clamped = %d/%d, want 72/50", dpi, q)
	}
}
