package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/JeremiahM37/librarr/internal/rename"
)

// Rename endpoints for the /api/files/* surface.
//
// These sit on the filesystem roots (MANGA_DIR/EBOOK_DIR/AUDIOBOOK_DIR), not on
// Kavita: unlike /api/library/metadata/rename-preview, nothing here needs a
// series id, so a folder Kavita has not scanned yet can still be renamed.
//
// Two invariants:
//
//   - The plan is always recomputed server-side from {path, series_name, mode,
//     volumes}. A plan received from the client is never applied — it would be a
//     list of arbitrary rename operations signed by nobody.
//   - Every directory whose contents change is passed to invalidateDirListing,
//     otherwise the browser keeps showing the pre-rename listing for up to
//     dirListingTTL.

// Status of one preview row. The front end maps these to labels; they are keys,
// not sentences, precisely so the reason is translatable.
const (
	// renameRowRename: the file will be renamed.
	renameRowRename = "rename"
	// renameRowUnchanged: already correctly named.
	renameRowUnchanged = "unchanged"
	// renameRowNoVolume: no volume supplied and none detected — never guessed.
	renameRowNoVolume = "no_volume"
	// renameRowCollision: its target is already claimed by another file.
	renameRowCollision = "collision"
	// renameRowFolderOnly: file renaming is off in this mode.
	renameRowFolderOnly = "folder_only"
)

// Where a row's volume number comes from.
const (
	volumeSourceManual   = "manual"
	volumeSourceDetected = "detected"
	volumeSourceNone     = "none"
)

// renameRequest is the payload of both preview and apply — identical by design,
// so applying is "preview again, then act" rather than a second code path.
//
// Volumes is keyed by the file's current name (not its full path): the folder is
// already given by Path, and a base name is what the UI has in hand. A number
// here overrides detection; a file absent from the map falls back to detection.
type renameRequest struct {
	Path       string         `json:"path"`
	SeriesName string         `json:"series_name"`
	Mode       string         `json:"mode"`
	Volumes    map[string]int `json:"volumes"`
}

// renameFileRow is one line of the preview: every file in the folder appears,
// including those that will not move. A plan alone only lists what changes,
// which would leave the user guessing about the rest.
type renameFileRow struct {
	Name         string `json:"name"`
	Kind         string `json:"kind"`
	Editable     bool   `json:"editable"`
	Volume       int    `json:"volume,omitempty"`
	VolumeSource string `json:"volume_source"`
	Target       string `json:"target,omitempty"`
	Status       string `json:"status"`
}

// renamePreviewResponse is what the preview panel renders.
type renamePreviewResponse struct {
	Path       string          `json:"path"`
	SeriesName string          `json:"series_name"`
	Mode       string          `json:"mode"`
	FolderFrom string          `json:"folder_from,omitempty"`
	FolderTo   string          `json:"folder_to,omitempty"`
	Files      []renameFileRow `json:"files"`
	Warnings   []string        `json:"warnings,omitempty"`
}

// renameApplyResponse reports what happened. Path is where the folder lives now,
// so the browser can re-open it after a folder rename.
type renameApplyResponse struct {
	Path     string   `json:"path"`
	Renamed  int      `json:"renamed"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// libraryDirValidator is the dirValidator for the filesystem roots. It rejects
// anything outside them and, via resolveLibraryPathForWrite, the bare roots.
func (s *Server) libraryDirValidator() dirValidator {
	return func(dir string) error {
		_, err := resolveLibraryPathForWrite(s.cfg, dir)
		return err
	}
}

// decodeRenameRequest validates the shared payload and resolves the folder.
func (s *Server) decodeRenameRequest(w http.ResponseWriter, r *http.Request) (renameRequest, string, rename.Mode, bool) {
	var req renameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return req, "", "", false
	}
	if strings.TrimSpace(req.SeriesName) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "series name is required"})
		return req, "", "", false
	}
	mode, err := rename.ParseMode(req.Mode)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return req, "", "", false
	}
	if mode == rename.ModeNone {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "a rename mode is required"})
		return req, "", "", false
	}

	// ForWrite: a bare library root is never renamable.
	dir, err := resolveLibraryPathForWrite(s.cfg, req.Path)
	if err != nil {
		writeLibraryPathError(w, err)
		return req, "", "", false
	}
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "directory not found"})
			return req, "", "", false
		}
		writeError(w, http.StatusInternalServerError, "failed to stat directory", err)
		return req, "", "", false
	}
	if !info.IsDir() {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "path is not a directory"})
		return req, "", "", false
	}
	return req, dir, mode, true
}

// renamePlanFor builds the plan and the per-file rows for one folder.
//
// fresh selects the listing source: the preview may serve the cached listing
// (it is what the user is looking at anyway), while an apply reads the directory
// again so it acts on the current state of the disk.
func (s *Server) renamePlanFor(dir string, req renameRequest, mode rename.Mode, fresh bool) (rename.Plan, []renameFileRow, []string, error) {
	var entries []fileEntry
	var err error
	if fresh {
		entries, err = readDirEntries(dir)
	} else {
		entries, err = listDir(dir)
	}
	if err != nil {
		return rename.Plan{}, nil, nil, err
	}

	var warnings []string
	files := make([]fileEntry, 0, len(entries))
	subdirs := 0
	for _, e := range entries {
		if e.IsDir {
			subdirs++
			continue
		}
		files = append(files, e)
	}
	if subdirs > 0 {
		warnings = append(warnings,
			"subfolders move with the folder but their files are not renamed")
	}

	inputs := make([]rename.FileInput, len(files))
	for i, e := range files {
		inputs[i] = rename.FileInput{Path: filepath.Join(dir, e.Name)}
		if n, ok := req.Volumes[e.Name]; ok && n > 0 {
			v := n
			inputs[i].Volume = &v
		}
	}

	plan := rename.BuildPlanFor(req.SeriesName, inputs, mode)
	if len(inputs) == 0 {
		// BuildPlanFor derives the folder from its files, so an empty folder (or
		// one holding only subfolders) yields no plan at all. The user still asked
		// to rename this folder, and its identity does not depend on its contents.
		if safe := rename.SanitizeSegment(req.SeriesName); filepath.Base(dir) != safe {
			plan.FolderFrom = dir
			plan.FolderTo = filepath.Join(filepath.Dir(dir), safe)
		}
	}
	warnings = append(warnings, plan.Warnings...)

	// The destination is a sibling of the source, so it is confined by
	// construction — check anyway rather than rely on that staying true.
	if plan.FolderTo != "" {
		if _, err := resolveLibraryPathForWrite(s.cfg, plan.FolderTo); err != nil {
			return rename.Plan{}, nil, nil, err
		}
	}

	return plan, renameRows(req.SeriesName, dir, files, inputs, plan, mode), warnings, nil
}

// renameRows turns the plan into one row per file. Rows are derived from the
// plan rather than recomputed beside it: whatever the plan decided is what the
// user is shown, and what the apply will do.
func renameRows(series, dir string, files []fileEntry, inputs []rename.FileInput, plan rename.Plan, mode rename.Mode) []renameFileRow {
	planned := make(map[string]string, len(plan.Files))
	for _, fc := range plan.Files {
		planned[fc.From] = filepath.Base(fc.To)
	}
	singleFile := len(files) == 1

	rows := make([]renameFileRow, 0, len(files))
	for i, e := range files {
		row := renameFileRow{Name: e.Name, Kind: e.Kind, Editable: e.Editable}

		if n, ok := inputs[i].ForcedVolume(); ok {
			row.Volume, row.VolumeSource = n, volumeSourceManual
		} else if n, ok := rename.VolumeNumber(e.Name); ok {
			row.Volume, row.VolumeSource = n, volumeSourceDetected
		} else {
			row.VolumeSource = volumeSourceNone
		}

		switch target, isPlanned := planned[filepath.Join(dir, e.Name)]; {
		case mode != rename.ModeFolderAndFiles:
			row.Status = renameRowFolderOnly
		case isPlanned:
			row.Status, row.Target = renameRowRename, target
		default:
			// Not planned: either it cannot be named, or it already bears its
			// target name, or another file claimed that name first.
			want, ok := rename.TargetFileName(series, e.Name, singleFile, inputs[i].Volume)
			switch {
			case !ok:
				row.Status = renameRowNoVolume
			case want == e.Name:
				row.Status, row.Target = renameRowUnchanged, want
			default:
				row.Status, row.Target = renameRowCollision, want
			}
		}
		rows = append(rows, row)
	}
	return rows
}

// writeRenamePlanError maps a planning failure to an HTTP status. A folder that
// vanished between the listing and the request is a 404, not a server fault.
func writeRenamePlanError(w http.ResponseWriter, err error) {
	if os.IsNotExist(err) {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "directory not found"})
		return
	}
	writeLibraryPathError(w, err)
}

// handleFilesRenamePreview shows what a folder rename would do, without touching
// anything.
//
// POST /api/files/rename/preview
func (s *Server) handleFilesRenamePreview(w http.ResponseWriter, r *http.Request) {
	req, dir, mode, ok := s.decodeRenameRequest(w, r)
	if !ok {
		return
	}

	plan, rows, warnings, err := s.renamePlanFor(dir, req, mode, false)
	if err != nil {
		writeRenamePlanError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, renamePreviewResponse{
		Path:       dir,
		SeriesName: req.SeriesName,
		Mode:       string(mode),
		FolderFrom: plan.FolderFrom,
		FolderTo:   plan.FolderTo,
		Files:      rows,
		Warnings:   warnings,
	})
}

// handleFilesRenameApply recomputes the plan from the payload and runs it.
//
// POST /api/files/rename/apply
func (s *Server) handleFilesRenameApply(w http.ResponseWriter, r *http.Request) {
	req, dir, mode, ok := s.decodeRenameRequest(w, r)
	if !ok {
		return
	}

	plan, _, warnings, err := s.renamePlanFor(dir, req, mode, true)
	if err != nil {
		writeRenamePlanError(w, err)
		return
	}

	// Collected before applying: applyPlan rewrites destinations as it goes.
	touched := touchedDirs(dir, plan)

	renamed, errs := applyPlan(&plan, s.libraryDirValidator())
	for _, d := range touched {
		invalidateDirListing(d)
	}

	// Where the folder actually ended up. Trusting plan.FolderTo is not enough:
	// applyPlan may have failed to move it.
	current := dir
	if plan.FolderTo != "" {
		if info, err := os.Stat(plan.FolderTo); err == nil && info.IsDir() {
			if _, err := os.Stat(plan.FolderFrom); os.IsNotExist(err) {
				current = plan.FolderTo
			}
		}
	}

	writeJSON(w, http.StatusOK, renameApplyResponse{
		Path:     current,
		Renamed:  renamed,
		Errors:   errs,
		Warnings: warnings,
	})
}

// touchedDirs lists every directory a plan can change, source and destination
// alike, so their cached listings can be dropped.
func touchedDirs(dir string, plan rename.Plan) []string {
	seen := map[string]bool{dir: true}
	if plan.FolderFrom != "" {
		// The folder changes name inside its parent, so the parent's listing is
		// stale too.
		seen[filepath.Dir(plan.FolderFrom)] = true
		seen[plan.FolderFrom] = true
	}
	if plan.FolderTo != "" {
		seen[plan.FolderTo] = true
	}
	for _, fc := range plan.Files {
		seen[filepath.Dir(fc.From)] = true
		seen[filepath.Dir(fc.To)] = true
	}
	dirs := make([]string, 0, len(seen))
	for d := range seen {
		dirs = append(dirs, d)
	}
	return dirs
}

// fileRenameRequest renames one file, free-form.
type fileRenameRequest struct {
	Path    string `json:"path"`
	NewName string `json:"new_name"`
}

// handleFilesRenameFile renames a single file to a name the user typed.
//
// POST /api/files/rename/file
//
// The original extension is always the one kept. A name whose tail already
// matches it ("Berserk Vol.01.cbz" on a .cbz) is accepted as-is; any other
// extension the user types is treated as part of the name, so "Berserk.exe"
// becomes "Berserk.exe.cbz" rather than a file Kavita can no longer read. That
// is deliberately blunt: changing a file's type is not a rename, and silently
// dropping the typed suffix would hide what happened.
func (s *Server) handleFilesRenameFile(w http.ResponseWriter, r *http.Request) {
	var req fileRenameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON body"})
		return
	}
	if strings.TrimSpace(req.NewName) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "new name is required"})
		return
	}

	src, err := resolveLibraryPathForWrite(s.cfg, req.Path)
	if err != nil {
		writeLibraryPathError(w, err)
		return
	}
	info, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "file not found"})
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to stat file", err)
		return
	}
	if info.IsDir() {
		// Folders are renamed through the plan endpoints, which know how to carry
		// their files along.
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "path is not a file"})
		return
	}

	ext := filepath.Ext(src)
	stem := strings.TrimSpace(req.NewName)
	if ext != "" && strings.EqualFold(filepath.Ext(stem), ext) {
		stem = strings.TrimSuffix(stem, filepath.Ext(stem))
	}
	name := rename.SanitizeSegment(stem) + ext

	if name == filepath.Base(src) {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "the name is unchanged"})
		return
	}

	dst, err := resolveLibraryPathForWrite(s.cfg, filepath.Join(filepath.Dir(src), name))
	if err != nil {
		writeLibraryPathError(w, err)
		return
	}
	if _, err := os.Stat(dst); err == nil {
		writeJSON(w, http.StatusConflict, map[string]any{"error": "target already exists"})
		return
	}
	if err := os.Rename(src, dst); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to rename file", err)
		return
	}
	invalidateDirListing(filepath.Dir(src))

	writeJSON(w, http.StatusOK, map[string]any{"name": name, "path": dst})
}
