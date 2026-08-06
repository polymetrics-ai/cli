//go:build windows

package durable

import (
	"os"
	"reflect"
	"testing"
)

func TestEnsureDirectoryTreeSkipsWindowsVolumeRoot(t *testing.T) {
	path := `C:\Users\alice\repo\.polymetrics\state`
	want := []string{
		`C:\Users`,
		`C:\Users\alice`,
		`C:\Users\alice\repo`,
		`C:\Users\alice\repo\.polymetrics`,
		path,
	}
	created := []string{}
	synced := []string{}
	if err := ensureDirectoryTree(path, `C:\Users\alice\repo`, 0o700, func(path string, _ os.FileMode) error {
		created = append(created, path)
		return nil
	}, func(path string) error {
		synced = append(synced, path)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(created, want) {
		t.Fatalf("created directories = %q, want %q", created, want)
	}
	if !reflect.DeepEqual(synced, want) {
		t.Fatalf("synced directories = %q, want %q", synced, want)
	}
}
