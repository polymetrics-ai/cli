package connectors

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

// directReadPageFlagsJSON is the single declaration of the page-navigation
// flags every direct_read command accepts on top of its own surface flags.
//
// A direct read returns ONE page. Without these documented, a caller who is
// told "more remain" has no way to ask for the rest — which is the same silent
// incompleteness in a different costume.
//
// It lives in JSON rather than in Go source for the reason
// binary_download_flags.json does: three renderers have to agree on it and one
// of them is not Go — runtime help (internal/cli), the generated
// MANUAL.md/SKILL.md (internal/connectors/guide.go), and the website bundle
// generator (website/scripts/lib/cli-surface.mjs).
//
// These flags are DERIVED behaviour, not per-command metadata: the engine
// answers them from each connector's already-declared pagination spec, so no
// bundle declares them and none ever needs to.
//
//go:embed direct_read_page_flags.json
var directReadPageFlagsJSON []byte

var directReadPageFlags = mustParseDirectReadPageFlags()

func mustParseDirectReadPageFlags() []CommandSurfaceFlag {
	var flags []CommandSurfaceFlag
	if err := json.Unmarshal(directReadPageFlagsJSON, &flags); err != nil {
		panic(fmt.Sprintf("parse direct_read_page_flags.json: %v", err))
	}
	if len(flags) == 0 {
		panic("direct_read_page_flags.json declares no flags")
	}
	for _, flag := range flags {
		if flag.Name == "" {
			panic("direct_read_page_flags.json declares a flag with no name")
		}
	}
	return flags
}

// DirectReadPageFlags returns the page-navigation flags every direct_read
// command accepts. The returned slice is a copy, so a renderer cannot mutate
// the shared declaration.
func DirectReadPageFlags() []CommandSurfaceFlag {
	return append([]CommandSurfaceFlag(nil), directReadPageFlags...)
}
