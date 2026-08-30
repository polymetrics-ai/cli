package main

import (
	"io/fs"
	"path"
)

// withoutCertificationOverlayFS gives authoring projections the connector
// definition bundle without its optional live-proof overlay. Runtime loading
// and certification commands continue to use the complete filesystem.
type withoutCertificationOverlayFS struct {
	fs.FS
	connector string
}

func (f withoutCertificationOverlayFS) Open(name string) (fs.File, error) {
	if name == path.Join(f.connector, "certification.json") {
		return nil, fs.ErrNotExist
	}
	return f.FS.Open(name)
}
