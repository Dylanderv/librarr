package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/JeremiahM37/librarr/internal/config"
	"github.com/JeremiahM37/librarr/internal/rename"
)

// renameTestServer wires the minimum a rename needs: a config whose manga
// library root is a temp dir.
func renameTestServer(t *testing.T, root string) *Server {
	t.Helper()
	return &Server{cfg: &config.Config{KavitaMangaLibraryPath: root}}
}

func mustWrite(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o666); err != nil {
		t.Fatal(err)
	}
}

func TestApplyRenamePlanFolderAndFiles(t *testing.T) {
	root := t.TempDir()
	s := renameTestServer(t, root)

	old := filepath.Join(root, "Old.Torrent.Name-GROUP")
	mustWrite(t, filepath.Join(old, "Old.Torrent.Name v1.cbz"))
	mustWrite(t, filepath.Join(old, "Old.Torrent.Name v2.cbz"))

	plan := seriesRenamePlan{SeriesID: 1, SeriesName: "Bakemono Recchan"}
	plan.Plan = rename.BuildPlan("Bakemono Recchan", []string{
		filepath.Join(old, "Old.Torrent.Name v1.cbz"),
		filepath.Join(old, "Old.Torrent.Name v2.cbz"),
	}, rename.ModeFolderAndFiles)

	n := s.applyRenamePlan(&plan)
	if len(plan.Errors) != 0 {
		t.Fatalf("errors: %v", plan.Errors)
	}
	// One folder plus two files.
	if n != 3 {
		t.Errorf("renamed = %d, want 3", n)
	}

	newDir := filepath.Join(root, "Bakemono Recchan")
	for _, want := range []string{"Bakemono Recchan Vol.01.cbz", "Bakemono Recchan Vol.02.cbz"} {
		if _, err := os.Stat(filepath.Join(newDir, want)); err != nil {
			t.Errorf("missing %s: %v", want, err)
		}
	}
	if _, err := os.Stat(old); !os.IsNotExist(err) {
		t.Error("old folder still exists")
	}
}

func TestApplyRenamePlanRefusesLibraryRoot(t *testing.T) {
	root := t.TempDir()
	s := renameTestServer(t, root)
	mustWrite(t, filepath.Join(root, "loose.cbz"))

	// A file sitting directly in the library root would make the root itself the
	// series folder. Renaming it would move the entire library.
	plan := seriesRenamePlan{SeriesID: 1}
	plan.Plan = rename.BuildPlan("Series", []string{filepath.Join(root, "loose.cbz")}, rename.ModeFolder)

	s.applyRenamePlan(&plan)
	if len(plan.Errors) == 0 {
		t.Fatal("renaming the library root must be refused")
	}
	if _, err := os.Stat(root); err != nil {
		t.Errorf("library root was touched: %v", err)
	}
}

func TestApplyRenamePlanRefusesExistingTarget(t *testing.T) {
	root := t.TempDir()
	s := renameTestServer(t, root)

	old := filepath.Join(root, "Old.Name")
	mustWrite(t, filepath.Join(old, "Old.Name v1.cbz"))
	// The destination already holds another series.
	mustWrite(t, filepath.Join(root, "Series", "keep.cbz"))

	plan := seriesRenamePlan{SeriesID: 1}
	plan.Plan = rename.BuildPlan("Series", []string{filepath.Join(old, "Old.Name v1.cbz")}, rename.ModeFolderAndFiles)

	s.applyRenamePlan(&plan)
	if len(plan.Errors) == 0 {
		t.Fatal("merging into an existing folder must be refused")
	}
	if _, err := os.Stat(filepath.Join(root, "Series", "keep.cbz")); err != nil {
		t.Errorf("existing folder's contents were disturbed: %v", err)
	}
	// The folder move failed, so the file must have been renamed where it is.
	if _, err := os.Stat(filepath.Join(old, "Series Vol.01.cbz")); err != nil {
		t.Errorf("file was not renamed in place after the folder move failed: %v", err)
	}
}

func TestApplyRenamePlanFolderOnlyKeepsFileNames(t *testing.T) {
	root := t.TempDir()
	s := renameTestServer(t, root)

	old := filepath.Join(root, "Old.Name")
	mustWrite(t, filepath.Join(old, "weird name v1.cbz"))

	plan := seriesRenamePlan{SeriesID: 1}
	plan.Plan = rename.BuildPlan("Series", []string{filepath.Join(old, "weird name v1.cbz")}, rename.ModeFolder)

	if n := s.applyRenamePlan(&plan); n != 1 {
		t.Errorf("renamed = %d, want 1 (the folder only)", n)
	}
	if _, err := os.Stat(filepath.Join(root, "Series", "weird name v1.cbz")); err != nil {
		t.Errorf("file should have moved with the folder, unrenamed: %v", err)
	}
}

func TestCheckRenamableDirRejectsOutsideRoots(t *testing.T) {
	root := t.TempDir()
	s := renameTestServer(t, root)

	if err := s.checkRenamableDir("/etc"); err == nil {
		t.Error("a path outside every library root must be rejected")
	}
	if err := s.checkRenamableDir(filepath.Join(root, "Series")); err != nil {
		t.Errorf("a path inside the root must be accepted: %v", err)
	}
}
