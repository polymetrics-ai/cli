// Package defs embeds every connector definition bundle
// (internal/connectors/engine.Bundle source trees). Wave0 ships no bundles
// yet (they land in Wave F: stripe, searxng, postgres); the embed pattern
// below must therefore tolerate an empty tree.
//
// //go:embed all:* on a directory containing only this Go file still
// compiles (it simply embeds defs.go itself, which engine.LoadAll ignores
// since it only descends into directories). This avoids adding a
// placeholder file purely to satisfy the embed directive.
package defs

import "embed"

//go:embed all:*
var FS embed.FS
