//go:build unix

package durable

import (
	"os"
	"reflect"
	"testing"
)

func TestEnsureDirectoryTreeSyncsFreshUnixProjectParent(t *testing.T) {
	path := "/project/.polymetrics/state"
	wantCreated := []string{
		"/project",
		"/project/.polymetrics",
		path,
	}
	wantSynced := append([]string{"/"}, wantCreated...)
	created := []string{}
	synced := []string{}
	if err := ensureDirectoryTree(path, "/project", 0o700, func(path string, _ os.FileMode) error {
		created = append(created, path)
		return nil
	}, func(path string) error {
		synced = append(synced, path)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(created, wantCreated) {
		t.Fatalf("created directories = %q, want %q", created, wantCreated)
	}
	if !reflect.DeepEqual(synced, wantSynced) {
		t.Fatalf("synced directories = %q, want %q", synced, wantSynced)
	}
}
