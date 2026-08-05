// Package defs embeds the runtime connector definition bundle files.
//
// The production CLI needs identity, specs, read/write declarations, schemas,
// public docs, optional command-surface metadata, and optional certification
// contracts. It deliberately does not embed API-surface coverage manifests or
// fixtures/**. Those remain on disk for connectorgen/conformance checks, and a
// disk-backed direct-write preflight can also use api_surface.json. Embedded
// rest_write definitions retain the limited endpoint shape needed by
// direct-write preflight. Keeping replay
// fixtures out of cmd/pm avoids compiling tens of megabytes of inert JSON into
// every shipped binary. A connector bundle whose spec publishes a fixture-replay
// mode as a supported config value embeds its own fixtures from its own
// subpackage instead of widening this pattern.
package defs

import "embed"

//go:embed */metadata.json */changefeed.json */spec.json */streams.json */writes.json */schemas/* */docs.md */operations.json */cli_surface.json */certification.json
var FS embed.FS
