package convert

// Hand-rolled minimal PDF fixture: N pages, each a solid-color rectangle drawn
// with vector operators — no fonts, no images, nothing a renderer can choke
// on differently across poppler versions. Offsets for the xref table are
// computed while writing, so the file is structurally valid by construction.

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func writeVectorPDF(t *testing.T, path string, pages int) {
	t.Helper()
	if pages < 1 {
		t.Fatal("pages must be >= 1")
	}
	var sb strings.Builder
	sb.WriteString("%PDF-1.4\n")
	offsets := []int{}
	addObj := func(num int, body string) {
		offsets = append(offsets, sb.Len())
		fmt.Fprintf(&sb, "%d 0 obj\n%s\nendobj\n", num, body)
	}
	kidRefs := []string{}
	for p := 0; p < pages; p++ {
		kidRefs = append(kidRefs, fmt.Sprintf("%d 0 R", 3+p*2))
	}
	addObj(1, "<< /Type /Catalog /Pages 2 0 R >>")
	addObj(2, fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", strings.Join(kidRefs, " "), pages))
	for p := 0; p < pages; p++ {
		r := float64((p*97+30)%200) / 255.0
		g := float64((p*57+90)%200) / 255.0
		b := float64((p*31+150)%200) / 255.0
		stream := fmt.Sprintf("%.3f %.3f %.3f rg\n20 20 160 160 re f\n", r, g, b)
		addObj(3+p*2, fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Contents %d 0 R /Resources << >> >>", 4+p*2))
		addObj(4+p*2, fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", len(stream), stream))
	}
	xrefPos := sb.Len()
	fmt.Fprintf(&sb, "xref\n0 %d\n", 2*pages+3)
	sb.WriteString("0000000000 65535 f \n")
	for _, off := range offsets {
		fmt.Fprintf(&sb, "%010d 00000 n \n", off)
	}
	fmt.Fprintf(&sb, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", 2*pages+3, xrefPos)
	if err := os.WriteFile(path, []byte(sb.String()), 0o666); err != nil {
		t.Fatal(err)
	}
}
