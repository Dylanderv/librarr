package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func doMergePreview(t *testing.T, s *Server, body any) *httptest.ResponseRecorder {
	t.Helper()
	return postJSON(t, s.handleFilesMergePreview, "/api/files/merge/preview", body)
}

func doMergeApply(t *testing.T, s *Server, body any) *httptest.ResponseRecorder {
	t.Helper()
	return postJSON(t, s.handleFilesMergeApply, "/api/files/merge/apply", body)
}

func decodeMergePreview(t *testing.T, rec *httptest.ResponseRecorder) mergePreviewResponse {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var out mergePreviewResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v (body %s)", err, rec.Body.String())
	}
	return out
}

func decodeMergeApply(t *testing.T, rec *httptest.ResponseRecorder) mergeApplyResponse {
	t.Helper()
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var out mergeApplyResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode: %v (body %s)", err, rec.Body.String())
	}
	return out
}

func mergeRowByKey(t *testing.T, rows []mergeFileRow, key string) mergeFileRow {
	t.Helper()
	for _, r := range rows {
		if r.Key == key {
			return r
		}
	}
	t.Fatalf("no row for %q in %+v", key, rows)
	return mergeFileRow{}
}

// beachVolumes seeds the exact layout the feature was written for: one torrent
// folder per volume, each holding a single file named after the folder.
func beachVolumes(t *testing.T, manga string) []string {
	t.Helper()
	names := []string{
		"La.Fille.de.la.plage.T01.FRENCH.CBZ.eBook-ebdz",
		"La.Fille.de.la.plage.T02.FRENCH.CBZ.eBook-ebdz",
	}
	dirs := make([]string, 0, len(names))
	for _, n := range names {
		dir := filepath.Join(manga, n)
		mustWriteN(t, filepath.Join(dir, n+".cbz"), 1)
		dirs = append(dirs, dir)
	}
	return dirs
}

func TestMergePreviewShowsEveryFileAndItsTarget(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dirs := beachVolumes(t, manga)

	out := decodeMergePreview(t, doMergePreview(t, s, mergeRequest{
		Sources: dirs, SeriesName: "La Fille de la plage",
	}))

	if out.Target != filepath.Join(manga, "La Fille de la plage") {
		t.Errorf("Target = %q", out.Target)
	}
	if out.TargetExists {
		t.Error("the target folder does not exist yet")
	}
	if len(out.Folders) != 2 || out.Folders[0].Files != 1 {
		t.Fatalf("unexpected folder summary: %+v", out.Folders)
	}
	if len(out.Files) != 2 {
		t.Fatalf("got %d rows, want 2: %+v", len(out.Files), out.Files)
	}
	for i, want := range []string{"La Fille de la plage Vol.01.cbz", "La Fille de la plage Vol.02.cbz"} {
		r := out.Files[i]
		if r.Status != mergeRowMove || r.Target != want || r.Volume != i+1 ||
			r.VolumeSource != volumeSourceDetected {
			t.Errorf("row %d: %+v, want target %q", i, r, want)
		}
	}
}

func TestMergeApplyProducesOneFolderPerSeries(t *testing.T) {
	s, manga := newBrowseFixture(t)
	dirs := beachVolumes(t, manga)

	out := decodeMergeApply(t, doMergeApply(t, s, mergeRequest{
		Sources: dirs, SeriesName: "La Fille de la plage",
	}))
	if len(out.Errors) != 0 {
		t.Fatalf("errors: %v", out.Errors)
	}
	if out.Moved != 2 {
		t.Errorf("moved = %d, want 2", out.Moved)
	}
	if len(out.RemovedFolders) != 2 || len(out.KeptFolders) != 0 {
		t.Errorf("emptied source folders must be removed: %+v / %+v", out.RemovedFolders, out.KeptFolders)
	}

	target := filepath.Join(manga, "La Fille de la plage")
	if out.Path != target {
		t.Errorf("Path = %q, want %q", out.Path, target)
	}
	for _, want := range []string{"La Fille de la plage Vol.01.cbz", "La Fille de la plage Vol.02.cbz"} {
		if _, err := os.Stat(filepath.Join(target, want)); err != nil {
			t.Errorf("missing %s: %v", want, err)
		}
	}
	for _, dir := range dirs {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			t.Errorf("%s survived the merge", dir)
		}
	}
}

func TestMergeInvalidatesListings(t *testing.T) {
	s, manga := newBrowseFixture(t)
	// doList builds a raw request line, so the paths here stay space-free.
	var dirs []string
	for _, n := range []string{"Serie.T01.GROUP", "Serie.T02.GROUP"} {
		dir := filepath.Join(manga, n)
		mustWriteN(t, filepath.Join(dir, n+".cbz"), 1)
		dirs = append(dirs, dir)
	}

	// Prime the cache for the parent and for both source folders.
	decodeList(t, doList(t, s, manga))
	for _, dir := range dirs {
		decodeList(t, doList(t, s, dir))
	}

	decodeMergeApply(t, doMergeApply(t, s, mergeRequest{Sources: dirs, SeriesName: "Serie"}))

	// A stale parent listing would keep showing the two torrent folders.
	parent := decodeList(t, doList(t, s, manga))
	if len(parent.Entries) != 1 || parent.Entries[0].Name != "Serie" {
		t.Fatalf("parent listing was not invalidated: %+v", parent.Entries)
	}
	inner := decodeList(t, doList(t, s, filepath.Join(manga, "Serie")))
	if len(inner.Entries) != 2 {
		t.Errorf("target listing was not invalidated: %+v", inner.Entries)
	}
}

func TestMergeTargetIsOneOfTheSources(t *testing.T) {
	s, manga := newBrowseFixture(t)
	// The user names the series after one of the selected folders — the common
	// case, and the one that must not lose the files already in it.
	keeper := filepath.Join(manga, "Berserk")
	mustWriteN(t, filepath.Join(keeper, "Berserk Vol.01.cbz"), 1)
	other := filepath.Join(manga, "Berserk.T02.FRENCH-GROUP")
	mustWriteN(t, filepath.Join(other, "Berserk.T02.FRENCH-GROUP.cbz"), 1)

	req := mergeRequest{Sources: []string{keeper, other}, SeriesName: "Berserk"}

	preview := decodeMergePreview(t, doMergePreview(t, s, req))
	if !preview.TargetExists || preview.Target != keeper {
		t.Fatalf("target should be the existing folder: %+v", preview)
	}
	if r := mergeRowByKey(t, preview.Files, "Berserk/Berserk Vol.01.cbz"); r.Status != mergeRowInPlace {
		t.Errorf("a file already at its final name must read as in_place: %+v", r)
	}

	out := decodeMergeApply(t, doMergeApply(t, s, req))
	if len(out.Errors) != 0 {
		t.Fatalf("errors: %v", out.Errors)
	}
	if out.Moved != 1 {
		t.Errorf("moved = %d, want 1 (the other file was already in place)", out.Moved)
	}
	for _, want := range []string{"Berserk Vol.01.cbz", "Berserk Vol.02.cbz"} {
		if _, err := os.Stat(filepath.Join(keeper, want)); err != nil {
			t.Errorf("missing %s: %v", want, err)
		}
	}
	if _, err := os.Stat(other); !os.IsNotExist(err) {
		t.Error("the emptied source folder survived")
	}
	// The target is a source, so it is never a removal candidate.
	if len(out.RemovedFolders) != 1 || out.RemovedFolders[0] != filepath.Base(other) {
		t.Errorf("removed = %+v", out.RemovedFolders)
	}
}

func TestMergeCollisionIsShownAndNotApplied(t *testing.T) {
	s, manga := newBrowseFixture(t)
	// Two folders that both claim volume 1: a torrent re-release next to the
	// original. Nothing may overwrite anything.
	a := filepath.Join(manga, "Serie.T01.GROUPA")
	mustWriteN(t, filepath.Join(a, "Serie.T01.GROUPA.cbz"), 1)
	b := filepath.Join(manga, "Serie.T01.GROUPB")
	mustWriteN(t, filepath.Join(b, "Serie.T01.GROUPB.cbz"), 1)

	req := mergeRequest{Sources: []string{a, b}, SeriesName: "Serie"}

	preview := decodeMergePreview(t, doMergePreview(t, s, req))
	for _, r := range preview.Files {
		if r.Status != mergeRowCollision {
			t.Errorf("both rows must be flagged as colliding: %+v", r)
		}
	}

	out := decodeMergeApply(t, doMergeApply(t, s, req))
	if out.Moved != 0 {
		t.Errorf("moved = %d, want 0", out.Moved)
	}
	if len(out.Skipped) != 2 {
		t.Errorf("the report must name what was skipped: %+v", out.Skipped)
	}
	// Both files are still where they were, and both folders survive.
	for _, dir := range []string{a, b} {
		if _, err := os.Stat(filepath.Join(dir, filepath.Base(dir)+".cbz")); err != nil {
			t.Errorf("a colliding file was moved: %v", err)
		}
	}
	if len(out.KeptFolders) != 2 {
		t.Errorf("kept = %+v", out.KeptFolders)
	}
}

func TestMergeExistingTargetFileIsNotOverwritten(t *testing.T) {
	s, manga := newBrowseFixture(t)
	target := filepath.Join(manga, "Serie")
	mustWriteN(t, filepath.Join(target, "Serie Vol.01.cbz"), 99)
	a := filepath.Join(manga, "Serie.T01.GROUP")
	mustWriteN(t, filepath.Join(a, "Serie.T01.GROUP.cbz"), 1)
	b := filepath.Join(manga, "Serie.T02.GROUP")
	mustWriteN(t, filepath.Join(b, "Serie.T02.GROUP.cbz"), 1)

	// The target folder exists but is not selected, so its file is untouchable.
	req := mergeRequest{Sources: []string{a, b}, SeriesName: "Serie"}

	preview := decodeMergePreview(t, doMergePreview(t, s, req))
	if r := mergeRowByKey(t, preview.Files, "Serie.T01.GROUP/Serie.T01.GROUP.cbz"); r.Status != mergeRowExists {
		t.Errorf("a taken target name must be reported: %+v", r)
	}

	out := decodeMergeApply(t, doMergeApply(t, s, req))
	if out.Moved != 1 {
		t.Errorf("moved = %d, want 1", out.Moved)
	}
	if info, err := os.Stat(filepath.Join(target, "Serie Vol.01.cbz")); err != nil || info.Size() != 99 {
		t.Errorf("the pre-existing file was overwritten: %v", err)
	}
	if _, err := os.Stat(filepath.Join(a, "Serie.T01.GROUP.cbz")); err != nil {
		t.Errorf("the refused file must stay where it is: %v", err)
	}
}

func TestMergeKeepsFoldersThatAreNotEmpty(t *testing.T) {
	s, manga := newBrowseFixture(t)
	a := filepath.Join(manga, "Serie.T01")
	mustWriteN(t, filepath.Join(a, "Serie.T01.cbz"), 1)
	// A subfolder is not a volume: it is not moved, so its folder survives.
	mustWriteN(t, filepath.Join(a, "extras", "art.jpg"), 1)
	b := filepath.Join(manga, "Serie.T02")
	mustWriteN(t, filepath.Join(b, "Serie.T02.cbz"), 1)
	// @eaDir is hidden from every listing, so this folder looks empty on screen
	// once its file has moved — and os.Remove will still refuse it.
	mustWriteN(t, filepath.Join(b, "@eaDir", "thumb.jpg"), 1)

	req := mergeRequest{Sources: []string{a, b}, SeriesName: "Serie"}

	preview := decodeMergePreview(t, doMergePreview(t, s, req))
	if len(preview.Warnings) == 0 {
		t.Error("the subfolder must be announced before the merge, not after")
	}
	if len(preview.Folders[0].Subfolders) != 1 || preview.Folders[0].Subfolders[0] != "extras" {
		t.Errorf("subfolders must be visible in the preview: %+v", preview.Folders[0])
	}

	out := decodeMergeApply(t, doMergeApply(t, s, req))
	if out.Moved != 2 {
		t.Fatalf("moved = %d, want 2", out.Moved)
	}
	if len(out.RemovedFolders) != 0 || len(out.KeptFolders) != 2 {
		t.Fatalf("removed = %+v, kept = %+v", out.RemovedFolders, out.KeptFolders)
	}
	byName := map[string]mergeKeptFolder{}
	for _, k := range out.KeptFolders {
		byName[k.Name] = k
	}
	if k := byName["Serie.T01"]; k.Reason != mergeKeptNotEmpty || len(k.Remaining) != 1 || k.Remaining[0] != "extras" {
		t.Errorf("unexpected report for the subfolder case: %+v", k)
	}
	// The whole point: the report names the sidecar the listing never showed.
	if k := byName["Serie.T02"]; k.Reason != mergeKeptNotEmpty || len(k.Remaining) != 1 || k.Remaining[0] != "@eaDir" {
		t.Errorf("a hidden entry must be named, not swallowed: %+v", k)
	}
	// Nothing was deleted recursively.
	for _, p := range []string{filepath.Join(a, "extras", "art.jpg"), filepath.Join(b, "@eaDir", "thumb.jpg")} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("content the user never saw was deleted: %v", err)
		}
	}
}

func TestMergeManualVolumeWins(t *testing.T) {
	s, manga := newBrowseFixture(t)
	a := filepath.Join(manga, "Serie.T01")
	mustWriteN(t, filepath.Join(a, "Serie.T01.cbz"), 1)
	b := filepath.Join(manga, "Serie.bonus")
	mustWriteN(t, filepath.Join(b, "bonus.cbz"), 1)

	req := mergeRequest{
		Sources: []string{a, b}, SeriesName: "Serie",
		// Undetectable on its own; the user says it is volume 2.
		Volumes: map[string]int{"Serie.bonus/bonus.cbz": 2},
	}

	preview := decodeMergePreview(t, doMergePreview(t, s, req))
	r := mergeRowByKey(t, preview.Files, "Serie.bonus/bonus.cbz")
	if r.VolumeSource != volumeSourceManual || r.Volume != 2 || r.Target != "Serie Vol.02.cbz" {
		t.Fatalf("a hand-entered volume must win: %+v", r)
	}

	out := decodeMergeApply(t, doMergeApply(t, s, req))
	if out.Moved != 2 {
		t.Fatalf("moved = %d, want 2 (errors %v)", out.Moved, out.Errors)
	}
	if _, err := os.Stat(filepath.Join(manga, "Serie", "Serie Vol.02.cbz")); err != nil {
		t.Errorf("missing the manually numbered volume: %v", err)
	}
}

func TestMergeLeavesUnnumberedFilesAlone(t *testing.T) {
	s, manga := newBrowseFixture(t)
	a := filepath.Join(manga, "Serie.T01")
	mustWriteN(t, filepath.Join(a, "Serie.T01.cbz"), 1)
	b := filepath.Join(manga, "Serie.artbook")
	mustWriteN(t, filepath.Join(b, "artbook.cbz"), 1)

	req := mergeRequest{Sources: []string{a, b}, SeriesName: "Serie"}

	preview := decodeMergePreview(t, doMergePreview(t, s, req))
	if r := mergeRowByKey(t, preview.Files, "Serie.artbook/artbook.cbz"); r.Status != mergeRowNoVolume {
		t.Errorf("no volume must never be guessed: %+v", r)
	}

	out := decodeMergeApply(t, doMergeApply(t, s, req))
	if out.Moved != 1 {
		t.Errorf("moved = %d, want 1", out.Moved)
	}
	if _, err := os.Stat(filepath.Join(b, "artbook.cbz")); err != nil {
		t.Errorf("an unnumbered file must stay put: %v", err)
	}
	if len(out.KeptFolders) != 1 || out.KeptFolders[0].Name != "Serie.artbook" {
		t.Errorf("kept = %+v", out.KeptFolders)
	}
}

func TestMergeRejectsInvalidSelections(t *testing.T) {
	s, manga := newBrowseFixture(t)
	a := filepath.Join(manga, "Serie.T01")
	mustWriteN(t, filepath.Join(a, "Serie.T01.cbz"), 1)
	b := filepath.Join(manga, "Serie.T02")
	mustWriteN(t, filepath.Join(b, "Serie.T02.cbz"), 1)
	// A folder one level down: a legitimate directory, but not a sibling.
	nested := filepath.Join(a, "inner")
	mustMkdir(t, nested)
	ebooks := filepath.Join(filepath.Dir(manga), "ebooks", "Serie.T03")
	mustMkdir(t, ebooks)

	cases := []struct {
		name string
		req  mergeRequest
		want int
	}{
		{"one folder", mergeRequest{Sources: []string{a}, SeriesName: "Serie"}, http.StatusBadRequest},
		{"no series name", mergeRequest{Sources: []string{a, b}}, http.StatusBadRequest},
		{"same folder twice", mergeRequest{Sources: []string{a, a}, SeriesName: "Serie"}, http.StatusBadRequest},
		// Merging across the tree is a different operation, refused explicitly.
		{"not siblings", mergeRequest{Sources: []string{a, nested}, SeriesName: "Serie"}, http.StatusBadRequest},
		{"another library", mergeRequest{Sources: []string{a, ebooks}, SeriesName: "Serie"}, http.StatusBadRequest},
		{"a file", mergeRequest{Sources: []string{a, filepath.Join(b, "Serie.T02.cbz")}, SeriesName: "Serie"}, http.StatusBadRequest},
		{"library root", mergeRequest{Sources: []string{manga, a}, SeriesName: "Serie"}, http.StatusForbidden},
		{"outside", mergeRequest{Sources: []string{a, "/etc"}, SeriesName: "Serie"}, http.StatusForbidden},
		{"traversal", mergeRequest{Sources: []string{a, filepath.Join(manga, "..", "..", "etc")}, SeriesName: "Serie"}, http.StatusForbidden},
		{"relative", mergeRequest{Sources: []string{a, "Serie.T02"}, SeriesName: "Serie"}, http.StatusBadRequest},
		{"not found", mergeRequest{Sources: []string{a, filepath.Join(manga, "nope")}, SeriesName: "Serie"}, http.StatusNotFound},
		// The series name sanitizes down to a traversal-free segment, but an
		// attempt to aim the target elsewhere must not land outside the parent.
		{"escaping series name", mergeRequest{Sources: []string{a, b}, SeriesName: "../../evil"}, http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := doMergePreview(t, s, tc.req).Code; got != tc.want {
				t.Errorf("preview status = %d, want %d", got, tc.want)
			}
		})
	}

	// The one case above that is accepted must have stayed inside the parent.
	out := decodeMergePreview(t, doMergePreview(t, s, mergeRequest{
		Sources: []string{a, b}, SeriesName: "../../evil",
	}))
	if out.Target != filepath.Join(manga, "evil") {
		t.Errorf("Target = %q, want a plain child of the parent", out.Target)
	}

	// Nothing was touched by the refusals.
	for _, p := range []string{filepath.Join(a, "Serie.T01.cbz"), filepath.Join(b, "Serie.T02.cbz")} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("a refused request modified the disk: %v", err)
		}
	}
}

func TestMergeApplyRejectsInvalidSelections(t *testing.T) {
	s, manga := newBrowseFixture(t)
	a := filepath.Join(manga, "Serie.T01")
	mustWriteN(t, filepath.Join(a, "Serie.T01.cbz"), 1)

	// The apply path shares decodeMergeRequest, so it refuses the same things —
	// asserted here rather than assumed, since it is the one that writes.
	for _, tc := range []struct {
		name string
		req  mergeRequest
		want int
	}{
		{"library root", mergeRequest{Sources: []string{manga, a}, SeriesName: "Serie"}, http.StatusForbidden},
		{"traversal", mergeRequest{Sources: []string{a, filepath.Join(manga, "..", "..", "etc")}, SeriesName: "Serie"}, http.StatusForbidden},
		{"one folder", mergeRequest{Sources: []string{a}, SeriesName: "Serie"}, http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := doMergeApply(t, s, tc.req).Code; got != tc.want {
				t.Errorf("status = %d, want %d", got, tc.want)
			}
		})
	}
	if _, err := os.Stat(filepath.Join(a, "Serie.T01.cbz")); err != nil {
		t.Errorf("a refused apply modified the disk: %v", err)
	}
}
