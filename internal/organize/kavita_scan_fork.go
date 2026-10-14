package organize

import (
	"context"
	"log/slog"

	"github.com/JeremiahM37/librarr/internal/kavita"
)

// scanKavita triggers a Kavita library scan after an import.
//
// SCOPE: every library is scanned, not just the one the imported file landed
// in. Narrowing it would mean matching the destination path against each
// library's folders, and at this point in the pipeline the caller no longer
// carries that path — copyToKavita and friends are fire-and-forget. Scanning
// one library too many costs Kavita some I/O; scanning the wrong one, or none,
// costs the user a book that never shows up. If this becomes expensive, the
// fix is to thread the destination path down to here and reuse the same
// folder-matching rule as the Browse scan button (internal/api/files_kavita.go).
//
// FORK POLICY: this implementation lives in its own file so the fork's
// scan-all behavior (via the kavita client) stays out of targets.go and does
// not collide textually with alternative per-library scan implementations.
// Callers keep calling lt.scanKavita() with no arguments.
func (lt *LibraryTargets) scanKavita() {
	if !lt.cfg.HasKavita() {
		return
	}

	client := kavita.New(lt.cfg.KavitaURL, lt.cfg.KavitaUser, lt.cfg.KavitaPass)
	if err := client.ScanAllLibraries(context.Background(), false); err != nil {
		slog.Error("kavita scan failed", "error", err)
		return
	}
	slog.Info("kavita library scan triggered")
}
