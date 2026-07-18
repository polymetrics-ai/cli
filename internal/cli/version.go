package cli

import (
	"fmt"
	"io"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func runVersion(stdout io.Writer, jsonOut bool) error {
	if jsonOut {
		return writeJSON(stdout, envelope{
			"kind":    "Version",
			"version": version,
			"commit":  commit,
			"date":    buildDate,
		})
	}
	fmt.Fprintf(stdout, "pm %s\ncommit: %s\nbuilt: %s\n", version, commit, buildDate)
	return nil
}
