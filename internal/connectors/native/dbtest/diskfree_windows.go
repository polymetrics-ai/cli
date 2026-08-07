//go:build windows

package dbtest

import "errors"

// diskFreeBytes is unavailable on Windows. This package is test support for a
// Podman-hosted Linux database container, so the platform is never exercised
// here; the split exists so the package still compiles into the Windows
// release build rather than breaking it.
func diskFreeBytes() (uint64, error) {
	return 0, errors.New("database test harness disk accounting is not supported on windows")
}
