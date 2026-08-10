package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/JeremiahM37/librarr/internal/rename"
)

// errMergeTooManyFiles is returned by the planner rather than by the decoder:
// the count is only known once every selected folder has been listed.
var errMergeTooManyFiles = errors.New("too many files in one merge")

// Folder merge for the /api/files/* surface.
//
// Torrents hand out one folder per volume; Kavita wants one folder per series.
// Merging is therefore its own operation, not a relaxation of the rename: the
// "target folder already exists" refusal in applyPlan stays exactly as strict as
// it is, because a rename that silently lands in a populated folder is a
// different intent from the one the user expressed here.
//
// Scope, deliberately narrow:
//
//   - Sources must be siblings — same parent. Gathering folders scattered across
//     the tree is another operation with other risks (which parent wins? what
//     does the breadcrumb become?), and it is refused rather than guessed at.
//   - Only the direct files of each source move. Subfolders are left where they
//     are: a subfolder is not a volume, moving it as-is could collide with a
//     subfolder of the target, and recursing would move files the user never saw
//     listed. A source that still holds one is simply not removed.
//   - A source folder is removed only when it is empty afterwards, and only with
//     os.Remove — never a recursive delete. What is left is named in the report,
//     hidden entries included: the listing hides dotfiles and @eaDir (the
//     Synology sidecar), so a folder that looked empty on screen can very well
//     refuse to go, and saying which file holds it is the only useful answer.
//
// The two invariants of files_rename.go hold here too: the plan is always
// recomputed server-side from the payload, and every directory whose contents
// change is passed to invalidateDirListing.

// Status of one preview row. Keys, not sentences, so the front end can
// translate them.
const (
	// mergeRowMove: the file moves into the target folder under its new name.
	mergeRowMove = "move"
	// mergeRowInPlace: the file already sits in the target folder under exactly
	// that name — it is shown so the row set matches what the user selected.
	mergeRowInPlace = "in_place"
	// mergeRowNoVolume: no volume supplied and none detected — never guessed, so
	// the file stays where it is.
	mergeRowNoVolume = "no_volume"
	// mergeRowCollision: several selected files claim the same target name.
	// None of them is applied; overwriting one with another is never silent.
	mergeRowCollision = "collision"
	// mergeRowExists: a file already occupies that name in the target folder and
	// is not part of the move.
	mergeRowExists = "exists"
)

// Why a source folder survived the merge.
const (
	// mergeKeptNotEmpty: something is still inside it.
	mergeKeptNotEmpty = "not_empty"
	// mergeKeptRemoveFailed: it looks empty but os.Remove refused (permissions,
	// a stale NFS handle).
	mergeKeptRemoveFailed = "remove_failed"
)

// mergeMaxSources bounds one request. Merging is a per-series gesture; a
// three-digit selection is a mis-click or a script, and either way the preview
// would be unreadable.
const mergeMaxSources = 100

// mergeMaxFiles bounds the preview the same way, since every file is one row.
const mergeMaxFiles = 2000

// mergeMaxRemaining caps how many leftover entries a kept folder reports. The
// point is to name what blocks the removal, not to mirror the directory.
const mergeMaxRemaining = 20

// mergeRequest is the payload of both preview and apply — identical by design,
// so applying is "preview again, then act".
//
// Volumes is keyed by "<source folder name>/<file name>", not by the bare file
// name: files come from several folders here, and two volume folders very often
// hold identically named files. The key is exactly the path each row displays,
// relative to the shared parent.
type mergeRequest struct {
	Sources    []string       `json:"sources"`
	SeriesName string         `json:"series_name"`
	Volumes    map[string]int `json:"volumes"`
}

// mergeKey builds the Volumes key of one file.
func mergeKey(folder, name string) string { return folder + "/" + name }

// mergeFileRow is one line of the preview: every file of every selected folder
// appears, including those that will not move.
type mergeFileRow struct {
	Key          string `json:"key"`
	Folder       string `json:"folder"`
	Name         string `json:"name"`
	Kind         string `json:"kind"`
	Volume       int    `json:"volume,omitempty"`
	VolumeSource string `json:"volume_source"`
	Target       string `json:"target,omitempty"`
	Status       string `json:"status"`
}

// mergeFolderRow summarises one selected folder, so the preview can show why a
// folder will survive before the merge is applied rather than after.
type mergeFolderRow struct {
	Name       string   `json:"name"`
	Path       string   `json:"path"`
	IsTarget   bool     `json:"is_target"`
	Files      int      `json:"files"`
	Subfolders []string `json:"subfolders,omitempty"`
}

// mergePreviewResponse is what the merge panel renders.
type mergePreviewResponse struct {
	Parent       string           `json:"parent"`
	SeriesName   string           `json:"series_name"`
	Target       string           `json:"target"`
	TargetExists bool             `json:"target_exists"`
	Folders      []mergeFolderRow `json:"folders"`
	Files        []mergeFileRow   `json:"files"`
	Warnings     []string         `json:"warnings,omitempty"`
}

// mergeKeptFolder is one source folder that could not be removed.
type mergeKeptFolder struct {
	Name      string   `json:"name"`
	Reason    string   `json:"reason"`
	Remaining []string `json:"remaining,omitempty"`
}

// mergeApplyResponse reports what actually happened. Skipped carries the rows
// that were not applied, with the same status keys as the preview, so the report
// explains itself instead of announcing a count that does not add up.
type mergeApplyResponse struct {
	Path           string            `json:"path"`
	Moved          int               `json:"moved"`
	Skipped        []mergeFileRow    `json:"skipped,omitempty"`
	RemovedFolders []string          `json:"removed_folders,omitempty"`
	KeptFolders    []mergeKeptFolder `json:"kept_folders,omitempty"`
	Errors         []string          `json:"errors,omitempty"`
	Warnings       []string          `json:"warnings,omitempty"`
}

// mergeTarget holds the resolved geometry of a merge: where the folders live,
// and where their files are going.
type mergeTarget struct {
	parent  string
	sources []string
	target  string
	// isSource is true when the target folder is itself one of the selected
	// folders — the common case when the user names the series after one of
	// them. Its files are then renamed where they are, and it is never removed.
	isSource bool
}

// decodeMergeRequest validates the shared payload and resolves every path it
// names. Sources go through resolveLibraryPathForWrite, so a library root and a
// traversal are refused before anything is read.
func (s *Server) decodeMergeRequest(w http.ResponseWriter, r *http.Request) (mergeRequest, mergeTarget, bool) {
	var req mergeRequest
	var mt mergeTarget

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return req, mt, false
	}
	if strings.TrimSpace(req.SeriesName) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "series name is required"})
		return req, mt, false
	}
	if len(req.Sources) < 2 {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "at least two folders are required to merge"})
		return req, mt, false
	}
	if len(req.Sources) > mergeMaxSources {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": fmt.Sprintf("too many folders in one merge (max %d)", mergeMaxSources)})
		return req, mt, false
	}

	seen := make(map[string]bool, len(req.Sources))
	for _, raw := range req.Sources {
		dir, err := resolveLibraryPathForWrite(s.cfg, raw)
		if err != nil {
			writeLibraryPathError(w, err)
			return req, mt, false
		}
		if seen[dir] {
			// The same folder twice would double-count its files and make the
			// removal pass trip over itself.
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "the same folder was selected twice"})
			return req, mt, false
		}
		info, err := os.Stat(dir)
		if err != nil {
			if os.IsNotExist(err) {
				writeJSON(w, http.StatusNotFound, map[string]any{"error": "directory not found"})
				return req, mt, false
			}
			writeError(w, http.StatusInternalServerError, "failed to stat directory", err)
			return req, mt, false
		}
		if !info.IsDir() {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "only folders can be merged"})
			return req, mt, false
		}
		parent := filepath.Dir(dir)
		if mt.parent == "" {
			mt.parent = parent
		} else if parent != mt.parent {
			// Refused on purpose: see the file header.
			writeJSON(w, http.StatusBadRequest, map[string]any{
				"error": "only folders sharing the same parent can be merged"})
			return req, mt, false
		}
		seen[dir] = true
		mt.sources = append(mt.sources, dir)
	}

	// The target is a plain segment of the shared parent — SanitizeSegment strips
	// separators, so it cannot climb out — but it is resolved anyway rather than
	// trusted to stay that way.
	target, err := resolveLibraryPathForWrite(s.cfg, filepath.Join(mt.parent, rename.SanitizeSegment(req.SeriesName)))
	if err != nil {
		writeLibraryPathError(w, err)
		return req, mt, false
	}
	mt.target = target
	for _, dir := range mt.sources {
		if dir == target {
			mt.isSource = true
		}
	}
	return req, mt, true
}

// mergePlanFor builds the folder summaries and the per-file rows.
//
// fresh selects the listing source, exactly as renamePlanFor does: the preview
// may serve the cached listing the user is already looking at, an apply reads
// the directories again so it acts on the current state of the disk.
func (s *Server) mergePlanFor(mt mergeTarget, req mergeRequest, fresh bool) ([]mergeFolderRow, []mergeFileRow, []string, error) {
	list := listDir
	if fresh {
		list = readDirEntries
	}

	var (
		folders  []mergeFolderRow
		inputs   []rename.FileInput
		rows     []mergeFileRow
		warnings []string
		subdirs  int
	)

	for _, dir := range mt.sources {
		entries, err := list(dir)
		if err != nil {
			return nil, nil, nil, err
		}
		folder := mergeFolderRow{Name: filepath.Base(dir), Path: dir, IsTarget: dir == mt.target}
		for _, e := range entries {
			if e.IsDir {
				subdirs++
				folder.Subfolders = append(folder.Subfolders, e.Name)
				continue
			}
			folder.Files++
			in := rename.FileInput{Path: filepath.Join(dir, e.Name)}
			if n, ok := req.Volumes[mergeKey(folder.Name, e.Name)]; ok && n > 0 {
				v := n
				in.Volume = &v
			}
			inputs = append(inputs, in)
			rows = append(rows, mergeFileRow{
				Key:    mergeKey(folder.Name, e.Name),
				Folder: folder.Name,
				Name:   e.Name,
				Kind:   e.Kind,
			})
		}
		folders = append(folders, folder)
	}
	if len(rows) > mergeMaxFiles {
		return nil, nil, nil, fmt.Errorf("%w (%d files, max %d)", errMergeTooManyFiles, len(rows), mergeMaxFiles)
	}
	if subdirs > 0 {
		warnings = append(warnings, "subfolders are not moved; a folder that still holds one is kept")
	}

	// BuildPlanFor supplies the volume detection and the naming convention. Its
	// folder decision is ignored on purpose: it renames a folder only when every
	// file shares one parent, which is precisely what a merge is not, so its
	// "files span N folders" warning describes the intent rather than a problem.
	// Its per-folder collision arbitration is not enough either — it settles
	// clashes inside one folder, and the merged namespace is wider — so the pass
	// below re-arbitrates over the target folder as a whole.
	plan := rename.BuildPlanFor(req.SeriesName, inputs, rename.ModeFolderAndFiles)
	planned := make(map[string]string, len(plan.Files))
	for _, fc := range plan.Files {
		planned[fc.From] = filepath.Base(fc.To)
	}
	singleFile := len(inputs) == 1

	for i := range rows {
		if n, ok := inputs[i].ForcedVolume(); ok {
			rows[i].Volume, rows[i].VolumeSource = n, volumeSourceManual
		} else if n, ok := rename.VolumeNumber(rows[i].Name); ok {
			rows[i].Volume, rows[i].VolumeSource = n, volumeSourceDetected
		} else {
			rows[i].VolumeSource = volumeSourceNone
		}

		if target, ok := planned[inputs[i].Path]; ok {
			rows[i].Target = target
			continue
		}
		// Absent from the plan means either that the file cannot be named at all,
		// or that it already bears its target name (a same-folder rename would be
		// a no-op — but a move is not), or that a sibling claimed the name first.
		// The last two still have a target here; the arbitration decides.
		if want, ok := rename.TargetFileName(req.SeriesName, rows[i].Name, singleFile, inputs[i].Volume); ok {
			rows[i].Target = want
		} else {
			rows[i].Status = mergeRowNoVolume
		}
	}

	occupied, err := s.mergeOccupiedNames(mt, rows, list)
	if err != nil {
		return nil, nil, nil, err
	}

	claims := make(map[string][]int, len(rows))
	for i := range rows {
		if rows[i].Status != "" {
			continue
		}
		claims[rows[i].Target] = append(claims[rows[i].Target], i)
	}
	for name, idx := range claims {
		if len(idx) > 1 {
			for _, i := range idx {
				rows[i].Status = mergeRowCollision
			}
			continue
		}
		i := idx[0]
		switch {
		case filepath.Join(mt.parent, rows[i].Folder, rows[i].Name) == filepath.Join(mt.target, name):
			rows[i].Status = mergeRowInPlace
		case occupied[name]:
			rows[i].Status = mergeRowExists
		default:
			rows[i].Status = mergeRowMove
		}
	}

	return folders, rows, warnings, nil
}

// mergeOccupiedNames lists the names already taken in the target folder by
// something the merge is not going to move. A file that stays put — no volume,
// or a subfolder — keeps its name, so it must be counted.
func (s *Server) mergeOccupiedNames(mt mergeTarget, rows []mergeFileRow, list func(string) ([]fileEntry, error)) (map[string]bool, error) {
	occupied := map[string]bool{}
	entries, err := list(mt.target)
	if err != nil {
		if os.IsNotExist(err) {
			// The target is created at apply time; nothing occupies it yet.
			return occupied, nil
		}
		return nil, err
	}

	moving := map[string]bool{}
	if mt.isSource {
		for _, r := range rows {
			// Only a row still eligible to move frees its name; one already ruled
			// out (no volume) stays exactly where it is.
			if r.Status == "" && filepath.Join(mt.parent, r.Folder) == mt.target {
				moving[r.Name] = true
			}
		}
	}
	for _, e := range entries {
		if !moving[e.Name] {
			occupied[e.Name] = true
		}
	}
	return occupied, nil
}

// writeMergePlanError maps a planning failure to an HTTP status. A folder that
// vanished between the listing and the request is a 404, not a server fault.
func writeMergePlanError(w http.ResponseWriter, err error) {
	if errors.Is(err, errMergeTooManyFiles) {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	if os.IsNotExist(err) {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "directory not found"})
		return
	}
	writeLibraryPathError(w, err)
}

// handleFilesMergePreview shows what a merge would do, without touching
// anything.
//
// POST /api/files/merge/preview
func (s *Server) handleFilesMergePreview(w http.ResponseWriter, r *http.Request) {
	req, mt, ok := s.decodeMergeRequest(w, r)
	if !ok {
		return
	}

	folders, rows, warnings, err := s.mergePlanFor(mt, req, false)
	if err != nil {
		writeMergePlanError(w, err)
		return
	}

	targetExists := false
	if info, err := os.Stat(mt.target); err == nil && info.IsDir() {
		targetExists = true
	}

	writeJSON(w, http.StatusOK, mergePreviewResponse{
		Parent:       mt.parent,
		SeriesName:   req.SeriesName,
		Target:       mt.target,
		TargetExists: targetExists,
		Folders:      folders,
		Files:        rows,
		Warnings:     warnings,
	})
}

// handleFilesMergeApply recomputes the plan from the payload and runs it.
//
// POST /api/files/merge/apply
func (s *Server) handleFilesMergeApply(w http.ResponseWriter, r *http.Request) {
	req, mt, ok := s.decodeMergeRequest(w, r)
	if !ok {
		return
	}

	_, rows, warnings, err := s.mergePlanFor(mt, req, true)
	if err != nil {
		writeMergePlanError(w, err)
		return
	}

	if !mt.isSource {
		// No hard-coded mode: 0o777 minus the umask, so the share's inherited ACL
		// is what decides (see LIBRARY-REWORK-PROMPT.md).
		if err := os.MkdirAll(mt.target, 0o777); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create the target folder", err)
			return
		}
	}

	out := mergeApplyResponse{Path: mt.target, Warnings: warnings}
	for _, row := range rows {
		if row.Status != mergeRowMove {
			if row.Status != mergeRowInPlace {
				out.Skipped = append(out.Skipped, row)
			}
			continue
		}
		src := filepath.Join(mt.parent, row.Folder, row.Name)
		dst := filepath.Join(mt.target, row.Target)
		// Both ends are re-checked here even though the plan was just rebuilt: a
		// symlink can appear between the listing and the rename, and this is the
		// only place that actually writes.
		if _, err := resolveLibraryPathForWrite(s.cfg, src); err != nil {
			out.Errors = append(out.Errors, fmt.Sprintf("%s: %v", row.Name, err))
			continue
		}
		if _, err := resolveLibraryPathForWrite(s.cfg, dst); err != nil {
			out.Errors = append(out.Errors, fmt.Sprintf("%s: %v", row.Target, err))
			continue
		}
		if _, err := os.Stat(dst); err == nil {
			out.Errors = append(out.Errors, fmt.Sprintf("%s: target already exists", row.Target))
			continue
		}
		if err := os.Rename(src, dst); err != nil {
			out.Errors = append(out.Errors, fmt.Sprintf("%s: %v", row.Name, err))
			continue
		}
		out.Moved++
	}

	out.RemovedFolders, out.KeptFolders = pruneMergedFolders(mt)

	invalidateDirListing(mt.parent)
	invalidateDirListing(mt.target)
	for _, dir := range mt.sources {
		invalidateDirListing(dir)
	}

	slog.Info("folders merged", "target", mt.target, "sources", len(mt.sources),
		"moved", out.Moved, "removed", len(out.RemovedFolders), "kept", len(out.KeptFolders))

	writeJSON(w, http.StatusOK, out)
}

// pruneMergedFolders removes the source folders that are empty and reports what
// held the others back.
//
// os.Remove, never os.RemoveAll: it succeeds only on an empty directory, which
// is the whole guarantee — nothing the user did not see listed is ever deleted.
// A folder that refuses to go is listed with what remains in it, dotfiles and
// @eaDir included: those are hidden from the listing, so they are exactly the
// reason a folder that looked empty on screen survives, and the report is the
// only place that can say so.
func pruneMergedFolders(mt mergeTarget) (removed []string, kept []mergeKeptFolder) {
	for _, dir := range mt.sources {
		if dir == mt.target {
			continue
		}
		if err := os.Remove(dir); err == nil {
			removed = append(removed, filepath.Base(dir))
			continue
		}
		k := mergeKeptFolder{Name: filepath.Base(dir), Reason: mergeKeptNotEmpty}
		raw, rerr := os.ReadDir(dir)
		if rerr != nil || len(raw) == 0 {
			k.Reason = mergeKeptRemoveFailed
		}
		for _, e := range raw {
			if len(k.Remaining) >= mergeMaxRemaining {
				break
			}
			k.Remaining = append(k.Remaining, e.Name())
		}
		kept = append(kept, k)
	}
	return removed, kept
}
