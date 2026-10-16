package api

import (
	"encoding/binary"
	"encoding/json"
	"hash/crc32"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// Minimal RAR 4 builder, self-contained so this package does not depend on
// internal/convert's test helpers: one stored file entry is all the handler
// tests need (byte-level fidelity is covered in internal/convert).
func writeTestCBR(t *testing.T, path, name string, data []byte) {
	t.Helper()
	block := func(typ byte, flags uint16, payload []byte) []byte {
		size := 7 + len(payload)
		h := crc32.NewIEEE()
		var hdr [5]byte
		hdr[0] = typ
		binary.LittleEndian.PutUint16(hdr[1:3], flags)
		binary.LittleEndian.PutUint16(hdr[3:5], uint16(size))
		h.Write(hdr[:])
		h.Write(payload)
		var out []byte
		var tmp [2]byte
		binary.LittleEndian.PutUint16(tmp[:], uint16(h.Sum32()))
		out = append(out, tmp[:]...)
		out = append(out, typ)
		binary.LittleEndian.PutUint16(tmp[:], flags)
		out = append(out, tmp[:]...)
		binary.LittleEndian.PutUint16(tmp[:], uint16(size))
		out = append(out, tmp[:]...)
		return append(out, payload...)
	}
	head := func() []byte {
		var p []byte
		var tmp [4]byte
		put := func(v uint32) { binary.LittleEndian.PutUint32(tmp[:], v); p = append(p, tmp[:]...) }
		put(uint32(len(data)))
		put(uint32(len(data)))
		p = append(p, 0x00)
		put(crc32.ChecksumIEEE(data))
		p = append(p, 0x00, 0x00, 0x00, 0x00, 0x1D, 0x30)
		var ns [2]byte
		binary.LittleEndian.PutUint16(ns[:], uint16(len(name)))
		p = append(p, ns[:]...)
		put(0x20)
		return append(p, name...)
	}
	var out []byte
	out = append(out, 'R', 'a', 'r', '!', 0x1A, 0x07, 0x00)
	out = append(out, block(0x73, 0x0000, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00})...)
	out = append(out, block(0x74, 0x8000, head())...)
	out = append(out, data...)
	out = append(out, block(0x7B, 0x0000, nil)...)
	if err := os.WriteFile(path, out, 0o666); err != nil {
		t.Fatal(err)
	}
}

func doConvert(t *testing.T, s *Server, body any) *httptest.ResponseRecorder {
	t.Helper()
	return postJSON(t, s.handleFilesConvert, "/api/files/convert", body)
}

func TestConvertCBRToCBZ(t *testing.T) {
	s, manga := newBrowseFixture(t)
	src := filepath.Join(manga, "Series.cbr")
	writeTestCBR(t, src, "page01.jpg", []byte{0xFF, 0xD8, 0xFF, 0x01})

	// Prime the cache so the test proves invalidation, not just conversion.
	decodeList(t, doList(t, s, manga))

	rec := doConvert(t, s, convertRequest{Path: src})
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(manga, "Series.cbz")); err != nil {
		t.Fatalf("cbz missing on disk: %v", err)
	}
	// The source goes only once the replacement exists.
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("source .cbr must be removed after conversion")
	}
	// The listing shows the swap without waiting for the TTL.
	listed := decodeList(t, doList(t, s, manga))
	names := map[string]bool{}
	for _, e := range listed.Entries {
		names[e.Name] = true
	}
	if names["Series.cbr"] || !names["Series.cbz"] {
		t.Errorf("listing must show the swap, got %+v", listed.Entries)
	}
}

func TestConvertRejectsBadRequests(t *testing.T) {
	s, manga := newBrowseFixture(t)
	mustMkdir(t, filepath.Join(manga, "dir"))
	mustWriteN(t, filepath.Join(manga, "a.cbz"), 3)
	writeTestCBR(t, filepath.Join(manga, "taken.cbr"), "p.jpg", []byte{1})
	mustWriteN(t, filepath.Join(manga, "taken.cbz"), 3)

	cases := []struct {
		name string
		body convertRequest
		want int
	}{
		{"missing path", convertRequest{}, http.StatusBadRequest},
		{"not a cbr", convertRequest{Path: filepath.Join(manga, "a.cbz")}, http.StatusBadRequest},
		{"directory", convertRequest{Path: filepath.Join(manga, "dir")}, http.StatusBadRequest},
		{"not found", convertRequest{Path: filepath.Join(manga, "nope.cbr")}, http.StatusNotFound},
		{"traversal", convertRequest{Path: filepath.Join(manga, "..", "etc", "x.cbr")}, http.StatusForbidden},
		{"outside", convertRequest{Path: "/etc/passwd.cbr"}, http.StatusForbidden},
		{"target taken", convertRequest{Path: filepath.Join(manga, "taken.cbr")}, http.StatusConflict},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := doConvert(t, s, tc.body).Code; got != tc.want {
				t.Errorf("status = %d, want %d", got, tc.want)
			}
		})
	}
}

func decodeBatchConvert(t *testing.T, rec *httptest.ResponseRecorder) batchConvertResponse {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var out batchConvertResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v (body %s)", err, rec.Body.String())
	}
	return out
}

func TestConvertBatchSuccess(t *testing.T) {
	s, manga := newBrowseFixture(t)
	writeTestCBR(t, filepath.Join(manga, "a.cbr"), "p1.jpg", []byte{1})
	writeTestCBR(t, filepath.Join(manga, "b.cbr"), "p2.jpg", []byte{2})

	decodeList(t, doList(t, s, manga))

	out := decodeBatchConvert(t, doConvert(t, s, convertRequest{
		Sources: []string{filepath.Join(manga, "a.cbr"), filepath.Join(manga, "b.cbr")},
	}))
	if len(out.Converted) != 2 || len(out.Failed) != 0 {
		t.Fatalf("unexpected response: %+v", out)
	}
	for _, name := range []string{"a.cbz", "b.cbz"} {
		if _, err := os.Stat(filepath.Join(manga, name)); err != nil {
			t.Errorf("%s missing on disk: %v", name, err)
		}
	}
	for _, name := range []string{"a.cbr", "b.cbr"} {
		if _, err := os.Stat(filepath.Join(manga, name)); !os.IsNotExist(err) {
			t.Errorf("source %s must be gone", name)
		}
	}
	if n := len(decodeList(t, doList(t, s, manga)).Entries); n != 2 {
		t.Errorf("listing has %d entries, want 2", n)
	}
}

func TestConvertBatchReportsFailures(t *testing.T) {
	s, manga := newBrowseFixture(t)
	writeTestCBR(t, filepath.Join(manga, "ok.cbr"), "p.jpg", []byte{1})
	writeTestCBR(t, filepath.Join(manga, "taken.cbr"), "p.jpg", []byte{1})
	mustWriteN(t, filepath.Join(manga, "taken.cbz"), 3)
	mustWriteN(t, filepath.Join(manga, "plain.cbz"), 3)
	mustMkdir(t, filepath.Join(manga, "somedir"))

	out := decodeBatchConvert(t, doConvert(t, s, convertRequest{
		Sources: []string{
			filepath.Join(manga, "ok.cbr"),
			filepath.Join(manga, "taken.cbr"),
			filepath.Join(manga, "gone.cbr"),
			filepath.Join(manga, "plain.cbz"),
			filepath.Join(manga, "somedir"),
		},
	}))
	if len(out.Converted) != 1 || out.Converted[0].Name != "ok.cbz" {
		t.Fatalf("converted = %+v, want [ok.cbz]", out.Converted)
	}
	failed := map[string]string{}
	for _, f := range out.Failed {
		failed[f.Name] = f.Reason
	}
	if failed["taken.cbr"] != convertFailExists ||
		failed["gone.cbr"] != convertFailMissing ||
		failed["plain.cbz"] != convertFailNotCBR ||
		failed["somedir"] != convertFailNotFile {
		t.Errorf("failed = %+v", out.Failed)
	}
}

func TestConvertBatchRejectsBadRequests(t *testing.T) {
	s, manga := newBrowseFixture(t)
	writeTestCBR(t, filepath.Join(manga, "a.cbr"), "p.jpg", []byte{1})

	cases := []struct {
		name string
		body convertRequest
		want int
	}{
		{"no sources", convertRequest{}, http.StatusBadRequest},
		{"source traversal aborts all", convertRequest{
			Sources: []string{filepath.Join(manga, "a.cbr"), filepath.Join(manga, "..", "x.cbr")},
		}, http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := doConvert(t, s, tc.body).Code; got != tc.want {
				t.Errorf("status = %d, want %d", got, tc.want)
			}
		})
	}
	if _, err := os.Stat(filepath.Join(manga, "a.cbr")); err != nil {
		t.Error("aborted batch must convert nothing")
	}
	if _, err := os.Stat(filepath.Join(manga, "a.cbz")); !os.IsNotExist(err) {
		t.Error("aborted batch must leave no output")
	}
}
