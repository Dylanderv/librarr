//go:build windows || plan9 || js

package main

import (
	"log/slog"
	"os"
)

// applyUmask is a no-op on platforms without a settable process umask.
func applyUmask() {
	if _, ok := os.LookupEnv("UMASK"); ok {
		slog.Warn("UMASK is not supported on this platform, ignoring")
	}
}
