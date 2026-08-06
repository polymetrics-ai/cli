package durable

import "testing"

func TestSyncDirectory(t *testing.T) {
	if err := SyncDirectory(t.TempDir()); err != nil {
		t.Fatal(err)
	}
}
