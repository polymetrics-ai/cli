package warehouse

import (
	"errors"
	"testing"
)

func TestRequireMinimumFastPathFreeSpaceAdmitsSafeCapacity(t *testing.T) {
	previous := availableDiskBytes
	availableDiskBytes = func(string) (int64, error) { return MinimumFastPathFreeBytes, nil }
	t.Cleanup(func() { availableDiskBytes = previous })
	if err := RequireMinimumFastPathFreeSpace("unused"); err != nil {
		t.Fatalf("RequireMinimumFastPathFreeSpace() = %v, want exact safe lower bound admitted", err)
	}
}

func TestRequireMinimumFastPathFreeSpaceRefusesBeforeSegmentWrite(t *testing.T) {
	previous := availableDiskBytes
	availableDiskBytes = func(string) (int64, error) { return MinimumFastPathFreeBytes - 1, nil }
	t.Cleanup(func() { availableDiskBytes = previous })
	err := RequireMinimumFastPathFreeSpace("unused")
	var disk *InsufficientFastPathDiskError
	if !errors.As(err, &disk) || !errors.Is(err, ErrInsufficientFastPathDisk) || disk.Available != MinimumFastPathFreeBytes-1 || disk.Required != MinimumFastPathFreeBytes {
		t.Fatalf("RequireMinimumFastPathFreeSpace() error = %T %v, want typed 3 GiB refusal", err, err)
	}
}

func TestRequireMinimumFastPathFreeSpacePropagatesFilesystemProbeFailure(t *testing.T) {
	previous := availableDiskBytes
	probeErr := errors.New("probe failed")
	availableDiskBytes = func(string) (int64, error) { return 0, probeErr }
	t.Cleanup(func() { availableDiskBytes = previous })
	if err := RequireMinimumFastPathFreeSpace("unused"); !errors.Is(err, probeErr) {
		t.Fatalf("RequireMinimumFastPathFreeSpace() error = %T %v, want filesystem probe failure", err, err)
	}
}
