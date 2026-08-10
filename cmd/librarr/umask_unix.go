//go:build !windows && !plan9 && !js

package main

import (
	"log/slog"
	"os"
	"strconv"
	"syscall"
)

// applyUmask honours the UMASK environment variable, matching the convention of
// the linuxserver.io images that most *arr deployments are built around.
//
// Librarr creates files and directories with permissive modes (0666/0777) and
// lets the umask subtract from them, so operators can widen or narrow group and
// other access without a rebuild. This matters on shared storage: a hardcoded
// mode caps the permissions inherited from the destination's ACL, which is what
// prevents companion services (Kavita, Komga, Calibre) from writing to the
// files librarr imports.
//
// Must be called once, before anything creates a file: the umask is
// process-wide and syscall.Umask is not safe to race against other goroutines.
func applyUmask() {
	raw := os.Getenv("UMASK")
	if raw == "" {
		return
	}

	mask, err := strconv.ParseInt(raw, 8, 32)
	if err != nil || mask < 0 || mask > 0o777 {
		slog.Warn("ignoring invalid UMASK, expected an octal mask such as 002", "value", raw)
		return
	}

	syscall.Umask(int(mask))
	slog.Info("umask applied", "umask", raw)
}
