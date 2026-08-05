package connectors

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

// binaryDownloadFlagsJSON is the single declaration of the destination flags a
// binary_download command accepts on top of its own surface flags.
//
// It lives in JSON rather than in Go source because three renderers have to
// agree on it and one of them is not Go: runtime help (internal/cli), the
// generated MANUAL.md/SKILL.md (RenderConnectorManual/Skill), and the website
// bundle generator (website/scripts/lib/cli-surface.mjs). Documenting
// --dest-root in only one of them is how an agent following SKILL.md ends up
// invoking a download command that refuses to run for want of a required flag
// it was never told about.
//
//go:embed binary_download_flags.json
var binaryDownloadFlagsJSON []byte

var binaryDownloadFlags = mustParseBinaryDownloadFlags()

func mustParseBinaryDownloadFlags() []CommandSurfaceFlag {
	var flags []CommandSurfaceFlag
	if err := json.Unmarshal(binaryDownloadFlagsJSON, &flags); err != nil {
		panic(fmt.Sprintf("parse binary_download_flags.json: %v", err))
	}
	if len(flags) == 0 {
		panic("binary_download_flags.json declares no flags")
	}
	for _, flag := range flags {
		if flag.Name == "" {
			panic("binary_download_flags.json declares a flag with no name")
		}
	}
	return flags
}

// BinaryDownloadFlags returns the destination flags every binary_download
// command accepts. The returned slice is a copy, so a renderer cannot mutate
// the shared declaration.
func BinaryDownloadFlags() []CommandSurfaceFlag {
	return append([]CommandSurfaceFlag(nil), binaryDownloadFlags...)
}
