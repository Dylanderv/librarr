package convert

import (
	"archive/zip"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// A valid conversion carries every file entry across with identical bytes,
// reports the directory entry as skipped, and leaves the source alone
// (removal is the API layer's job, not the translator's).
func TestCBRToCBZRoundTrip(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "Series Vol.01.cbr")
	writeStoredRAR(t, src, []rarEntry{
		{name: "Series Vol.01/", dir: true},
		{name: "Series Vol.01/cover.jpg", data: []byte{0xFF, 0xD8, 0xFF, 0x00, 0x01, 0x02}},
		{name: "Series Vol.01/page01.png", data: []byte{0x89, 'P', 'N', 'G', 0x00, 0x01}},
		{name: "notes.txt", data: []byte("hello")},
	})

	res, err := CBRToCBZ(src)
	if err != nil {
		t.Fatalf("CBRToCBZ: %v", err)
	}
	if res.Path != filepath.Join(dir, "Series Vol.01.cbz") {
		t.Errorf("path = %q", res.Path)
	}
	if res.Files != 3 {
		t.Errorf("files = %d, want 3", res.Files)
	}
	if len(res.Skipped) != 1 || res.Skipped[0] != "Series Vol.01/" {
		t.Errorf("skipped = %q", res.Skipped)
	}
	// The source is untouched: deleting it is the caller's decision.
	if _, err := os.Stat(src); err != nil {
		t.Errorf("source must survive conversion: %v", err)
	}

	zr, err := zip.OpenReader(res.Path)
	if err != nil {
		t.Fatalf("output is not a valid zip: %v", err)
	}
	defer zr.Close()
	got := map[string][]byte{}
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open %s: %v", f.Name, err)
		}
		var buf []byte
		tmp := make([]byte, 512)
		for {
			n, err := rc.Read(tmp)
			buf = append(buf, tmp[:n]...)
			if err != nil {
				break
			}
		}
		rc.Close()
		got[f.Name] = buf
		if f.Method != zip.Deflate {
			t.Errorf("%s: method = %v, want Deflate", f.Name, f.Method)
		}
	}
	want := map[string]string{
		"Series Vol.01/cover.jpg":  "\xFF\xD8\xFF\x00\x01\x02",
		"Series Vol.01/page01.png": "\x89PNG\x00\x01",
		"notes.txt":                "hello",
	}
	if len(got) != len(want) {
		t.Fatalf("entries = %q, want %q", keys(got), keysStr(want))
	}
	for name, content := range want {
		if string(got[name]) != content {
			t.Errorf("%s: content mismatch", name)
		}
	}
}

func TestCBRToCBZRefusesTargetCollision(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a.cbr")
	writeStoredRAR(t, src, []rarEntry{{name: "p.jpg", data: []byte{1}}})
	if err := os.WriteFile(filepath.Join(dir, "a.cbz"), []byte("x"), 0o666); err != nil {
		t.Fatal(err)
	}
	if _, err := CBRToCBZ(src); !errors.Is(err, ErrExists) {
		t.Fatalf("err = %v, want ErrExists", err)
	}
}

func TestCBRToCBZRefusesNonArchive(t *testing.T) {
	src := filepath.Join(t.TempDir(), "fake.cbr")
	if err := os.WriteFile(src, []byte("not a rar archive at all"), 0o666); err != nil {
		t.Fatal(err)
	}
	if _, err := CBRToCBZ(src); err == nil {
		t.Fatal("expected an error for a non-RAR file")
	}
}

func TestCBRToCBZRefusesEmptyArchive(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "empty.cbr")
	writeStoredRAR(t, src, []rarEntry{{name: "onlydir/", dir: true}})
	if _, err := CBRToCBZ(src); !errors.Is(err, ErrEmpty) {
		t.Fatalf("err = %v, want ErrEmpty", err)
	}
	// Nothing usable means nothing left behind.
	if _, err := os.Stat(filepath.Join(dir, "empty.cbz")); !os.IsNotExist(err) {
		t.Error("no output must remain after a failed conversion")
	}
}

func TestCBRToCBZRefusesPathEscape(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "evil.cbr")
	writeStoredRAR(t, src, []rarEntry{{name: "../../evil.jpg", data: []byte{1}}})
	if _, err := CBRToCBZ(src); !errors.Is(err, ErrUnsafe) {
		t.Fatalf("err = %v, want ErrUnsafe", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "evil.cbz")); !os.IsNotExist(err) {
		t.Error("no output must remain after a refused conversion")
	}
}

func TestCBRToCBZRefusesEncryptedEntry(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "locked.cbr")
	writeStoredRAR(t, src, []rarEntry{{name: "p.jpg", data: []byte{1}, encrypted: true}})
	if _, err := CBRToCBZ(src); !errors.Is(err, ErrEncrypted) {
		t.Fatalf("err = %v, want ErrEncrypted", err)
	}
}

func keys(got map[string][]byte) []string {
	out := make([]string, 0, len(got))
	for k := range got {
		out = append(out, k)
	}
	return out
}

func keysStr(want map[string]string) []string {
	out := make([]string, 0, len(want))
	for k := range want {
		out = append(out, k)
	}
	return out
}
