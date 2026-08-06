//go:build windows

package durable

import (
	"reflect"
	"testing"
)

func TestDirectoryAncestorsExcludeWindowsVolumeRoot(t *testing.T) {
	path := `C:\Users\alice\repo\.polymetrics\state`
	want := []string{
		`C:\Users`,
		`C:\Users\alice`,
		`C:\Users\alice\repo`,
		`C:\Users\alice\repo\.polymetrics`,
		path,
	}
	if got := directoryAncestors(path); !reflect.DeepEqual(got, want) {
		t.Fatalf("directoryAncestors() = %q, want %q", got, want)
	}
}
