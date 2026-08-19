// Package defs embeds the runtime connector definition bundle files.
//
// The production CLI embeds connector identity, specs, read/write declarations,
// schemas, public docs, and optional command-surface, certification, rate-limit,
// and native-database policy declarations. It deliberately does not embed
// API-surface coverage manifests or fixtures/**. Those remain on disk for
// connectorgen/conformance checks. The
// generated operation endpoint ledger retains only direct-read method, path,
// operation kind, and response-cap bindings needed for runtime preflight. In
// shipped builds, direct-write endpoint validation is derived only from the
// embedded rest_write operation declarations; it checks internal declaration
// consistency, not provider documented-surface provenance (#3773 owns that).
// Keeping replay
// fixtures out of cmd/pm avoids compiling tens of megabytes of inert JSON into
// every shipped binary. A connector bundle whose spec publishes a fixture-replay
// mode as a supported config value embeds its own fixtures from its own
// subpackage instead of widening this pattern.
package defs

import "embed"

//go:embed operation_endpoint_ledger.json */metadata.json */changefeed.json */polling_watermark.json */sync_transport.json */spec.json */streams.json */writes.json */schemas/* */sources/* */docs.md */operations.json */cli_surface.json */certification.json */rate_limits.json */database.json
var FS embed.FS
