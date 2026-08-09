//go:build windows

package dbtest

import "errors"

// diskFreeAt is unavailable on Windows.
func diskFreeAt(string) (uint64, error) {
	return 0, errors.New("database test harness disk accounting is not supported on windows")
}
