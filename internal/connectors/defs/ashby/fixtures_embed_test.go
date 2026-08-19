package ashby

import (
	"io/fs"
	"testing"
)

func TestFixturesEmbedsStreamReplayPages(t *testing.T) {
	fixtures, err := Fixtures()
	if err != nil {
		t.Fatalf("Fixtures(): %v", err)
	}
	entries, err := fs.ReadDir(fixtures, "streams")
	if err != nil {
		t.Fatalf("fs.ReadDir(streams): %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("fixtures/streams embeds zero stream directories")
	}

	pages := 0
	err = fs.WalkDir(fixtures, "streams", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			pages++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("fs.WalkDir(streams): %v", err)
	}
	if pages == 0 {
		t.Fatal("fixtures/streams embeds zero replay pages")
	}
}
