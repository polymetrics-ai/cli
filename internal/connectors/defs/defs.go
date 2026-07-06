// Package defs embeds the runtime connector definition bundle files.
//
// The production CLI needs identity, specs, read/write declarations, schemas,
// and public docs. It deliberately does not embed conformance-only artifacts
// such as api_surface.json and fixtures/**; those remain on disk for
// connectorgen/conformance tests. Keeping replay fixtures out of cmd/pm avoids
// compiling tens of megabytes of inert JSON into every shipped binary.
package defs

import "embed"

//go:embed */metadata.json */spec.json */streams.json */writes.json */schemas/* */docs.md
var FS embed.FS
