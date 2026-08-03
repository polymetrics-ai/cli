// Package ashby embeds the Ashby connector's replay fixtures.
//
// The shared runtime bundle embed in the parent defs package deliberately
// excludes every connector's fixtures/** tree. Ashby exposes mode=fixture as a
// documented connection-spec property, so its own fixtures travel with the
// binary through this connector-local embed instead.
package ashby

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed fixtures
var fixturesFS embed.FS

// Fixtures returns the Ashby replay fixture tree rooted at fixtures/, so
// callers address entries as "streams/<stream>/<page>.json".
func Fixtures() (fs.FS, error) {
	sub, err := fs.Sub(fixturesFS, "fixtures")
	if err != nil {
		return nil, fmt.Errorf("open ashby fixtures: %w", err)
	}
	return sub, nil
}
