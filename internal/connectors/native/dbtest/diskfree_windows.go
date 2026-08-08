//go:build windows

package dbtest

import "errors"

// diskFreeBytes is unavailable on Windows. This package is test support for a
// Podman-hosted Linux database container, so the platform is never exercised
// here. Windows is not a release target either (see AGENTS.md); the split
// exists only so a GOOS=windows build or vet of this tree stays clean.
func diskFreeBytes() (uint64, error) {
	return 0, errors.New("database test harness disk accounting is not supported on windows")
}
