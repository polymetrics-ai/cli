// Package defs embeds the runtime connector definition bundle files.
//
// The production CLI embeds execution JSON only: connector identity and config,
// protocol declarations, schemas, command bindings, and optional execution
// policies. Source locks and provider evidence remain repository-only authoring
// inputs and cannot affect runtime.
package defs

import "embed"

//go:embed */metadata.json */changefeed.json */polling_watermark.json */sync_transport.json */spec.json */streams.json */writes.json */schemas/* */operations.json */cli_surface.json */rate_limits.json */database.json
var FS embed.FS
