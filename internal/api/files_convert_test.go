package api

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

type convertJobResponse struct {
	ID       string          `json:"id"`
	Batch    bool            `json:"batch"`
	Status   string          `json:"status"`
	Total    int             `json:"total"`
	Finished int             `json:"finished"`
	Current  string          `json:"current"`
	Result   json.RawMessage `json:"result"`
	Error    string          `json:"error"`
}

// decodeConvertJob accepts the 202 answer.
func decodeConvertJob(t *testing.T, rec *httptest.ResponseRecorder) convertJobResponse {
	t.Helper()
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var out convertJobResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v (body %s)", err, rec.Body.String())
	}
	if out.ID == "" || out.Status != convertJobQueued {
		t.Fatalf("unexpected job: %+v", out)
	}
	return out
}

// waitConvertJob polls the manager until the job terminates.
func waitConvertJob(t *testing.T, s *Server, id string) convertJobResponse {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for {
		job, ok := s.convertQueue().get(id)
		if !ok {
			t.Fatalf("job %s vanished", id)
		}
		if job.Status == convertJobDone || job.Status == convertJobError {
			raw, _ := json.Marshal(job.snapshot())
			var out convertJobResponse
			if err := json.Unmarshal(raw, &out); err != nil {
				t.Fatal(err)
			}
			return out
		}
		if time.Now().After(deadline) {
			t.Fatalf("job %s still %s after 30s", id, job.Status)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// doConvertJob enqueues and waits: the async equivalent of the old sync call.
func doConvertJob(t *testing.T, s *Server, body any) convertJobResponse {
	t.Helper()
	return waitConvertJob(t, s, decodeConvertJob(t, doConvert(t, s, body)).ID)
}

func decodeConvertResult(t *testing.T, job convertJobResponse) convertResponse {
	t.Helper()
	if job.Status != convertJobDone {
		t.Fatalf("job status = %s, error = %s", job.Status, job.Error)
	}
	var out convertResponse
	if err := json.Unmarshal(job.Result, &out); err != nil {
		t.Fatalf("decode result: %v (body %s)", err, string(job.Result))
	}
	return out
}

func TestConvertCBRToCBZ(t *testing.T) {
	s, manga := newBrowseFixture(t)
	src := filepath.Join(manga, "Series.cbr")
	writeTestCBR(t, src, "page01.jpg", []byte{0xFF, 0xD8, 0xFF, 0x01})

	// Prime the cache so the test proves invalidation, not just conversion.
	decodeList(t, doList(t, s, manga))

	out := decodeConvertResult(t, doConvertJob(t, s, convertRequest{Path: src}))
	if out.Name != "Series.cbz" {
		t.Fatalf("unexpected result: %+v", out)
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
		{"not convertible", convertRequest{Path: filepath.Join(manga, "a.cbz")}, http.StatusBadRequest},
		{"directory", convertRequest{Path: filepath.Join(manga, "dir")}, http.StatusBadRequest},
		{"not found", convertRequest{Path: filepath.Join(manga, "nope.cbr")}, http.StatusNotFound},
		{"traversal", convertRequest{Path: filepath.Join(manga, "..", "etc", "x.cbr")}, http.StatusForbidden},
		{"outside", convertRequest{Path: "/etc/passwd.cbr"}, http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := doConvert(t, s, tc.body).Code; got != tc.want {
				t.Errorf("status = %d, want %d", got, tc.want)
			}
		})
	}
}

func decodeBatchConvert(t *testing.T, job convertJobResponse) batchConvertResponse {
	t.Helper()
	if job.Status != convertJobDone {
		t.Fatalf("job status = %s, error = %s", job.Status, job.Error)
	}
	var out batchConvertResponse
	if err := json.Unmarshal(job.Result, &out); err != nil {
		t.Fatalf("decode: %v (body %s)", err, string(job.Result))
	}
	return out
}

func TestConvertBatchSuccess(t *testing.T) {
	s, manga := newBrowseFixture(t)
	writeTestCBR(t, filepath.Join(manga, "a.cbr"), "p1.jpg", []byte{1})
	writeTestCBR(t, filepath.Join(manga, "b.cbr"), "p2.jpg", []byte{2})

	decodeList(t, doList(t, s, manga))

	out := decodeBatchConvert(t, doConvertJob(t, s, convertRequest{
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

	out := decodeBatchConvert(t, doConvertJob(t, s, convertRequest{
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
		failed["plain.cbz"] != convertFailNotSupported ||
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

// Minimal one-page vector PDF for handler tests (byte-level fidelity is
// covered in internal/convert; here the handler only needs a real PDF).
func TestConvertSingleCollisionIsAJobError(t *testing.T) {
	s, manga := newBrowseFixture(t)
	writeTestCBR(t, filepath.Join(manga, "taken.cbr"), "p.jpg", []byte{1})
	mustWriteN(t, filepath.Join(manga, "taken.cbz"), 3)

	job := doConvertJob(t, s, convertRequest{Path: filepath.Join(manga, "taken.cbr")})
	if job.Status != convertJobError || job.Error == "" {
		t.Fatalf("expected a failed job, got %+v", job)
	}
}

func TestConvertJobNotFound(t *testing.T) {
	s, _ := newBrowseFixture(t)
	req := httptest.NewRequest(http.MethodGet, "/api/files/convert/jobs/nope", nil)
	req.SetPathValue("id", "nope")
	rec := httptest.NewRecorder()
	s.handleFilesConvertJob(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func writeTestPDF(t *testing.T, path string) {
	t.Helper()
	var sb strings.Builder
	sb.WriteString("%PDF-1.4\n")
	offsets := []int{}
	addObj := func(num int, body string) {
		offsets = append(offsets, sb.Len())
		fmt.Fprintf(&sb, "%d 0 obj\n%s\nendobj\n", num, body)
	}
	stream := "0.1 0.2 0.8 rg\n20 20 160 160 re f\n"
	addObj(1, "<< /Type /Catalog /Pages 2 0 R >>")
	addObj(2, "<< /Type /Pages /Kids [3 0 R] /Count 1 >>")
	addObj(3, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Contents 4 0 R /Resources << >> >>")
	addObj(4, fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", len(stream), stream))
	xrefPos := sb.Len()
	fmt.Fprintf(&sb, "xref\n0 5\n0000000000 65535 f \n")
	for _, off := range offsets {
		fmt.Fprintf(&sb, "%010d 00000 n \n", off)
	}
	fmt.Fprintf(&sb, "trailer\n<< /Size 5 /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", xrefPos)
	if err := os.WriteFile(path, []byte(sb.String()), 0o666); err != nil {
		t.Fatal(err)
	}
}

func TestConvertSinglePDF(t *testing.T) {
	if _, err := exec.LookPath("pdfinfo"); err != nil {
		t.Skip("poppler not installed")
	}
	s, manga := newBrowseFixture(t)
	src := filepath.Join(manga, "doc.pdf")
	writeTestPDF(t, src)

	decodeConvertResult(t, doConvertJob(t, s, convertRequest{Path: src}))
	if _, err := os.Stat(filepath.Join(manga, "doc.cbz")); err != nil {
		t.Fatalf("cbz missing on disk: %v", err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("source .pdf must be gone after conversion")
	}
}

func TestConvertBatchMixedFormats(t *testing.T) {
	if _, err := exec.LookPath("pdfinfo"); err != nil {
		t.Skip("poppler not installed")
	}
	s, manga := newBrowseFixture(t)
	writeTestCBR(t, filepath.Join(manga, "a.cbr"), "p.jpg", []byte{1})
	writeTestPDF(t, filepath.Join(manga, "doc.pdf"))
	mustWriteN(t, filepath.Join(manga, "notes.txt"), 3)

	out := decodeBatchConvert(t, doConvertJob(t, s, convertRequest{
		Sources: []string{
			filepath.Join(manga, "a.cbr"),
			filepath.Join(manga, "doc.pdf"),
			filepath.Join(manga, "notes.txt"),
		},
	}))
	if len(out.Converted) != 2 {
		t.Fatalf("converted = %+v, want 2", out.Converted)
	}
	if len(out.Failed) != 1 || out.Failed[0].Name != "notes.txt" ||
		out.Failed[0].Reason != convertFailNotSupported {
		t.Errorf("failed = %+v", out.Failed)
	}
}
