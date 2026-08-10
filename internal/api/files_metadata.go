package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/JeremiahM37/librarr/internal/metadata/filewriter"
)

// Metadata endpoints for the /api/files/* surface.
//
// Three handlers, three jobs:
//
//   - GET  /api/files/metadata          — what one file currently holds.
//   - POST /api/files/metadata/preview  — what an edit would change, file by file.
//   - POST /api/files/metadata/apply    — the same edit, actually written.
//
// The invariants of the rename endpoints hold here too, for the same reasons:
// the plan is recomputed server-side on apply (a plan sent by the client is a
// list of arbitrary writes signed by nobody), and every directory whose files
// change is passed to invalidateDirListing — a metadata write changes a file's
// size and mtime, both of which the cached listing carries.
//
// The reason a preview exists at all is decision 4: propagating a field over a
// folder overwrites what is already in the files. That is the behaviour we
// want, but never silently — the preview is where the overwrite becomes
// visible, which is also why it reports the *current* value of every field it
// would touch, not just the target.

// metadataMaxFiles bounds how many files one preview or apply covers.
//
// Reading metadata means opening an archive, i.e. at least one round-trip to a
// slow NFS mount per file — a recursive edit over a whole library would hang
// the request for minutes. 500 comfortably covers a series (a long-running
// manga is ~100 volumes) while keeping the worst case in the seconds. Hitting
// it sets Truncated in the response: a preview that silently stopped short
// would read as full coverage, and the user would apply it believing every file
// was accounted for.
const metadataMaxFiles = 500

// Status of one preview row. Keys, not sentences: the front end translates them.
const (
	// metaRowEdit: at least one field changes.
	metaRowEdit = "edit"
	// metaRowUnchanged: the file already holds the target values.
	metaRowUnchanged = "unchanged"
	// metaRowUnsupported: the format carries no metadata librarr can write.
	metaRowUnsupported = "unsupported"
	// metaRowError: the file could not be read (broken archive, I/O failure).
	metaRowError = "error"
)

// Why a file is not editable. Three distinct reasons, deliberately not merged
// into a generic one: they call for different actions from the user. A PDF is
// editable in principle but pointless (Kavita ignores PDF metadata, so only the
// file name matters); a CBR is a RAR archive the stdlib cannot rewrite, and
// converting it to CBZ would fix that; audio is simply outside what filewriter
// handles.
const (
	metaReasonPDF    = "pdf_ignored_by_kavita"
	metaReasonCBR    = "cbr_rar_archive"
	metaReasonAudio  = "audio_unsupported"
	metaReasonFormat = "format_unsupported"
)

// Informational notices, separated from errors on purpose. The Kavita path
// (applyMetadataToSeries) folds its skipped-file notes into Errors prefixed
// "not renamed: ", which leaves the UI unable to tell an FYI from a failure.
// Here a notice is a key plus a count, and Errors only ever holds real failures.
const (
	metaNoticeTruncated   = "truncated"
	metaNoticeUnsupported = "unsupported_skipped"
	metaNoticeUnchanged   = "already_up_to_date"
)

// unsupportedReason maps a listing kind to the reason its metadata is off limits.
func unsupportedReason(kind string) string {
	switch kind {
	case "pdf":
		return metaReasonPDF
	case "cbr":
		return metaReasonCBR
	case "audio":
		return metaReasonAudio
	default:
		return metaReasonFormat
	}
}

// metadataNotice is an informational outcome — never a failure.
type metadataNotice struct {
	Key   string `json:"key"`
	Count int    `json:"count,omitempty"`
}

// metadataValues is a file's metadata rendered for display. Unlike
// filewriter.Fields it has no pointers: this is "what is in the file right
// now", where absent and empty are the same thing to a reader.
type metadataValues struct {
	Title       string   `json:"title"`
	Series      string   `json:"series"`
	SeriesIndex string   `json:"series_index"`
	Summary     string   `json:"summary"`
	Publisher   string   `json:"publisher"`
	Language    string   `json:"language"`
	ISBN        string   `json:"isbn"`
	Authors     []string `json:"authors"`
	Genres      []string `json:"genres"`
	Tags        []string `json:"tags"`
}

// toMetadataValues flattens what filewriter read out of a file.
func toMetadataValues(f filewriter.Fields) metadataValues {
	return metadataValues{
		Title:       derefString(f.Title),
		Series:      derefString(f.Series),
		SeriesIndex: derefString(f.SeriesIndex),
		Summary:     derefString(f.Summary),
		Publisher:   derefString(f.Publisher),
		Language:    derefString(f.Language),
		ISBN:        derefString(f.ISBN),
		Authors:     nonNil(f.Authors),
		Genres:      nonNil(f.Genres),
		Tags:        nonNil(f.Tags),
	}
}

func derefString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// metadataChange is one field of one file: what is there, what would replace
// it. Both are carried even when they are equal, so the panel can show the
// whole edit rather than only its diff.
type metadataChange struct {
	Field   string `json:"field"`
	Current string `json:"current"`
	Target  string `json:"target"`
	Changed bool   `json:"changed"`
}

// metadataFileRow is one line of the preview.
type metadataFileRow struct {
	Name     string           `json:"name"`
	Path     string           `json:"path"`
	Kind     string           `json:"kind"`
	Editable bool             `json:"editable"`
	Status   string           `json:"status"`
	Reason   string           `json:"reason,omitempty"`
	Error    string           `json:"error,omitempty"`
	Changes  []metadataChange `json:"changes,omitempty"`
}

// metadataRequest is the payload of both preview and apply — identical by
// design, so applying is "preview again, then act".
//
// Path is a folder (the edit propagates to its files) or a single file. Fields
// keeps the pointer semantics of filewriter.Fields all the way from JSON: a key
// absent from the object leaves that field alone, a key present with an empty
// value clears it.
type metadataRequest struct {
	Path      string         `json:"path"`
	Recursive bool           `json:"recursive"`
	Mode      string         `json:"mode"` // "set" (default, replaces lists) or "append"
	Fields    bulkMetaFields `json:"fields"`
}

// metadataPreviewResponse is what the preview table renders.
type metadataPreviewResponse struct {
	Path      string            `json:"path"`
	IsDir     bool              `json:"is_dir"`
	Recursive bool              `json:"recursive"`
	Files     []metadataFileRow `json:"files"`
	Inspected int               `json:"inspected"`
	Limit     int               `json:"limit"`
	Truncated bool              `json:"truncated"`
	Notices   []metadataNotice  `json:"notices,omitempty"`
	Errors    []string          `json:"errors,omitempty"`
}

// metadataApplyResponse reports what was written. Notices and Errors are
// distinct fields: "12 files were already up to date" is not a problem, and
// mixing the two into one list is exactly the defect the Kavita path has.
type metadataApplyResponse struct {
	Path        string           `json:"path"`
	Written     int              `json:"written"`
	Unchanged   int              `json:"unchanged"`
	Unsupported int              `json:"unsupported"`
	Inspected   int              `json:"inspected"`
	Limit       int              `json:"limit"`
	Truncated   bool             `json:"truncated"`
	Notices     []metadataNotice `json:"notices,omitempty"`
	Errors      []string         `json:"errors,omitempty"`
}

// perFileFields are the fields a folder-wide edit refuses.
//
// Propagating a Title, a volume number or an ISBN to every file of a folder
// produces a series whose volumes are all called the same thing and all claim
// to be volume 3 — a destructive no-op nobody asks for on purpose. The UI hides
// them on the folder panel, and the server refuses them too: a guard that only
// exists in the front end is not a guard.
//
// The other fields (Series, Summary, Publisher, Language, Authors, Genres,
// Tags) describe the work rather than the single volume, so they propagate.
var perFileFields = []struct {
	name    string
	present func(bulkMetaFields) bool
}{
	{"title", func(f bulkMetaFields) bool { return f.Title != nil }},
	{"series_index", func(f bulkMetaFields) bool { return f.SeriesIndex != nil }},
	{"isbn", func(f bulkMetaFields) bool { return f.ISBN != nil }},
}

// metadataTarget is one file a preview or an apply covers.
type metadataTarget struct {
	dir      string
	name     string
	kind     string
	editable bool
}

func (t metadataTarget) path() string { return filepath.Join(t.dir, t.name) }

// collectMetadataTargets lists the files under root, one directory at a time.
//
// fresh selects the listing source, as in renamePlanFor: a preview may serve
// the cached listing (it is what the user is looking at), an apply reads the
// directory again so it acts on the current state of the disk.
//
// The traversal is breadth-first and bounded by metadataMaxFiles; the second
// return value says whether the bound cut it short.
func collectMetadataTargets(root string, recursive, fresh bool) ([]metadataTarget, bool, error) {
	read := listDir
	if fresh {
		read = readDirEntries
	}

	var out []metadataTarget
	queue := []string{root}
	for len(queue) > 0 {
		dir := queue[0]
		queue = queue[1:]

		entries, err := read(dir)
		if err != nil {
			if dir == root {
				return nil, false, err
			}
			// A subfolder that vanished or is unreadable must not sink the whole
			// preview; the files we did collect are still valid.
			continue
		}
		for _, e := range entries {
			if e.IsDir {
				if recursive {
					queue = append(queue, filepath.Join(dir, e.Name))
				}
				continue
			}
			if len(out) >= metadataMaxFiles {
				return out, true, nil
			}
			out = append(out, metadataTarget{dir: dir, name: e.Name, kind: e.Kind, editable: e.Editable})
		}
	}
	return out, false, nil
}

// metadataRows builds one row per file: what it holds, what the edit would make
// of it. Unsupported formats are listed without being opened — classifying them
// costs nothing, and opening them would buy nothing.
func metadataRows(targets []metadataTarget, want filewriter.Fields) []metadataFileRow {
	rows := make([]metadataFileRow, 0, len(targets))
	for _, tgt := range targets {
		row := metadataFileRow{
			Name:     tgt.name,
			Path:     tgt.path(),
			Kind:     tgt.kind,
			Editable: tgt.editable,
		}
		if !tgt.editable {
			row.Status = metaRowUnsupported
			row.Reason = unsupportedReason(tgt.kind)
			rows = append(rows, row)
			continue
		}

		current, err := filewriter.Read(row.Path)
		if err != nil {
			var unsup filewriter.ErrUnsupported
			if errors.As(err, &unsup) {
				// classifyFile and filewriter disagree on this extension; trust
				// filewriter, it is the one that would do the writing.
				row.Editable = false
				row.Status = metaRowUnsupported
				row.Reason = metaReasonFormat
			} else {
				row.Status = metaRowError
				row.Error = err.Error()
			}
			rows = append(rows, row)
			continue
		}

		row.Changes = metadataChanges(current, want)
		row.Status = metaRowUnchanged
		for _, c := range row.Changes {
			if c.Changed {
				row.Status = metaRowEdit
				break
			}
		}
		rows = append(rows, row)
	}
	return rows
}

// metadataChanges compares a file's current metadata with the edit, one row per
// field the edit actually mentions.
func metadataChanges(current, want filewriter.Fields) []metadataChange {
	var out []metadataChange

	scalars := []struct {
		field   string
		current *string
		want    *string
	}{
		{"title", current.Title, want.Title},
		{"series", current.Series, want.Series},
		{"series_index", current.SeriesIndex, want.SeriesIndex},
		{"summary", current.Summary, want.Summary},
		{"publisher", current.Publisher, want.Publisher},
		{"language", current.Language, want.Language},
		{"isbn", current.ISBN, want.ISBN},
	}
	for _, s := range scalars {
		if s.want == nil {
			continue
		}
		now := derefString(s.current)
		out = append(out, metadataChange{
			Field: s.field, Current: now, Target: *s.want, Changed: now != *s.want,
		})
	}

	lists := []struct {
		field   string
		current []string
		want    []string
	}{
		{"authors", current.Authors, want.Authors},
		{"genres", current.Genres, want.Genres},
		{"tags", current.Tags, want.Tags},
	}
	for _, l := range lists {
		if l.want == nil {
			continue
		}
		// The target is what filewriter would actually store, append mode
		// included — otherwise an "append" preview would show the new values
		// alone and hide the ones being kept.
		merged := filewriter.MergeLists(l.current, l.want, want.AppendLists)
		now := strings.Join(l.current, ", ")
		target := strings.Join(merged, ", ")
		out = append(out, metadataChange{
			Field: l.field, Current: now, Target: target, Changed: now != target,
		})
	}

	return out
}

// countRows tallies a preview by status.
func countRows(rows []metadataFileRow) (edit, unchanged, unsupported int) {
	for _, r := range rows {
		switch r.Status {
		case metaRowEdit:
			edit++
		case metaRowUnchanged:
			unchanged++
		case metaRowUnsupported:
			unsupported++
		}
	}
	return
}

// buildMetadataNotices turns the tally into informational notices. Nothing here
// is a failure — that is the whole point of keeping it out of Errors.
func buildMetadataNotices(truncated bool, unchanged, unsupported int) []metadataNotice {
	var out []metadataNotice
	if truncated {
		out = append(out, metadataNotice{Key: metaNoticeTruncated, Count: metadataMaxFiles})
	}
	if unsupported > 0 {
		out = append(out, metadataNotice{Key: metaNoticeUnsupported, Count: unsupported})
	}
	if unchanged > 0 {
		out = append(out, metadataNotice{Key: metaNoticeUnchanged, Count: unchanged})
	}
	return out
}

// decodeMetadataRequest validates the shared payload and resolves the target.
func (s *Server) decodeMetadataRequest(w http.ResponseWriter, r *http.Request) (metadataRequest, string, bool, filewriter.Fields, bool) {
	var req metadataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return req, "", false, filewriter.Fields{}, false
	}

	// ForWrite even for the preview: a bare library root is never an edit
	// target, and previewing an edit that could never be applied is a trap.
	target, err := resolveLibraryPathForWrite(s.cfg, req.Path)
	if err != nil {
		writeLibraryPathError(w, err)
		return req, "", false, filewriter.Fields{}, false
	}
	info, err := os.Stat(target)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "path not found"})
			return req, "", false, filewriter.Fields{}, false
		}
		writeError(w, http.StatusInternalServerError, "failed to stat path", err)
		return req, "", false, filewriter.Fields{}, false
	}
	isDir := info.IsDir()

	fields := toWriterFields(req.Fields, req.Mode)
	if fields.IsEmpty() {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "no fields to update"})
		return req, "", false, filewriter.Fields{}, false
	}
	if isDir {
		var refused []string
		for _, f := range perFileFields {
			if f.present(req.Fields) {
				refused = append(refused, f.name)
			}
		}
		if len(refused) > 0 {
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"error": strings.Join(refused, ", ") + " can only be set on a single file",
			})
			return req, "", false, filewriter.Fields{}, false
		}
	}

	return req, target, isDir, fields, true
}

// metadataTargetsFor resolves the request to the list of files it covers: the
// one file it points at, or the files of the folder it points at.
func metadataTargetsFor(target string, isDir, recursive, fresh bool) ([]metadataTarget, bool, error) {
	if !isDir {
		kind, editable := classifyFile(filepath.Base(target))
		return []metadataTarget{{
			dir: filepath.Dir(target), name: filepath.Base(target),
			kind: kind, editable: editable,
		}}, false, nil
	}
	return collectMetadataTargets(target, recursive, fresh)
}

// handleFilesMetadata returns the metadata one file currently holds.
//
// GET /api/files/metadata?path=/books/manga/Berserk/Berserk Vol.01.cbz
//
// Called only when a file panel is opened (decision 9): reading metadata means
// opening the archive, and doing that for every row of a listing would make
// browsing an NFS-backed library unusable.
func (s *Server) handleFilesMetadata(w http.ResponseWriter, r *http.Request) {
	clean, _, err := resolveLibraryPath(s.cfg, r.URL.Query().Get("path"))
	if err != nil {
		writeLibraryPathError(w, err)
		return
	}
	info, err := os.Stat(clean)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "file not found"})
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to stat file", err)
		return
	}
	if info.IsDir() {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "path is not a file"})
		return
	}

	name := filepath.Base(clean)
	kind, editable := classifyFile(name)
	resp := map[string]any{
		"path":     clean,
		"name":     name,
		"kind":     kind,
		"editable": editable,
		"fields":   toMetadataValues(filewriter.Fields{}),
	}
	if !editable {
		// Not an error: a PDF row opens its panel like any other, it just shows
		// why its fields are locked (decision 11).
		resp["reason"] = unsupportedReason(kind)
		writeJSON(w, http.StatusOK, resp)
		return
	}

	fields, err := filewriter.Read(clean)
	if err != nil {
		var unsup filewriter.ErrUnsupported
		if errors.As(err, &unsup) {
			resp["editable"] = false
			resp["reason"] = metaReasonFormat
			writeJSON(w, http.StatusOK, resp)
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to read metadata", err)
		return
	}
	resp["fields"] = toMetadataValues(fields)
	writeJSON(w, http.StatusOK, resp)
}

// handleFilesMetadataPreview shows what an edit would change, without touching
// anything.
//
// POST /api/files/metadata/preview
func (s *Server) handleFilesMetadataPreview(w http.ResponseWriter, r *http.Request) {
	req, target, isDir, fields, ok := s.decodeMetadataRequest(w, r)
	if !ok {
		return
	}

	targets, truncated, err := metadataTargetsFor(target, isDir, req.Recursive, false)
	if err != nil {
		writeRenamePlanError(w, err)
		return
	}
	rows := metadataRows(targets, fields)
	_, unchanged, unsupported := countRows(rows)

	writeJSON(w, http.StatusOK, metadataPreviewResponse{
		Path:      target,
		IsDir:     isDir,
		Recursive: req.Recursive,
		Files:     rows,
		Inspected: len(rows),
		Limit:     metadataMaxFiles,
		Truncated: truncated,
		Notices:   buildMetadataNotices(truncated, unchanged, unsupported),
	})
}

// handleFilesMetadataApply recomputes the plan from the payload and writes it.
//
// POST /api/files/metadata/apply
//
// Files are written one by one and a failure never aborts the batch: on a
// library of this size, one corrupt archive must not cost the other 99 their
// edit. Every failure is reported.
func (s *Server) handleFilesMetadataApply(w http.ResponseWriter, r *http.Request) {
	req, target, isDir, fields, ok := s.decodeMetadataRequest(w, r)
	if !ok {
		return
	}

	// fresh: the apply reads the directories again rather than trusting the
	// cached listing the preview was built from.
	targets, truncated, err := metadataTargetsFor(target, isDir, req.Recursive, true)
	if err != nil {
		writeRenamePlanError(w, err)
		return
	}
	rows := metadataRows(targets, fields)
	_, unchanged, unsupported := countRows(rows)

	written := 0
	var errs []string
	touched := map[string]bool{}
	for _, row := range rows {
		switch row.Status {
		case metaRowError:
			errs = append(errs, row.Name+": "+row.Error)
			continue
		case metaRowEdit:
		default:
			// Unchanged and unsupported files are accounted for in the notices;
			// neither is a failure and neither needs a write.
			continue
		}
		// Re-validate each file: the folder was confined, but a symlink inside it
		// could point anywhere.
		path, err := resolveLibraryPathForWrite(s.cfg, row.Path)
		if err != nil {
			errs = append(errs, row.Name+": "+err.Error())
			continue
		}
		if err := filewriter.Apply(path, fields); err != nil {
			errs = append(errs, row.Name+": "+err.Error())
			continue
		}
		written++
		touched[filepath.Dir(path)] = true
	}

	// A write changes the file's size and mtime, both of which the cached
	// listing carries.
	for dir := range touched {
		invalidateDirListing(dir)
	}

	writeJSON(w, http.StatusOK, metadataApplyResponse{
		Path:        target,
		Written:     written,
		Unchanged:   unchanged,
		Unsupported: unsupported,
		Inspected:   len(rows),
		Limit:       metadataMaxFiles,
		Truncated:   truncated,
		Notices:     buildMetadataNotices(truncated, unchanged, unsupported),
		Errors:      errs,
	})
}
