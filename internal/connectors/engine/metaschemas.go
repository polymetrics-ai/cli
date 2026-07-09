package engine

import _ "embed"

// The bundle-file meta-schemas live as data files under schema/ so they can
// be read/reviewed independently (and reused by cmd/connectorgen), but are
// embedded here as the single compile-time source the loader validates
// against.

//go:embed schema/metadata.schema.json
var metadataSchemaJSON string

//go:embed schema/spec.schema.json
var specSchemaJSON string

//go:embed schema/streams.schema.json
var streamsSchemaJSON string

//go:embed schema/writes.schema.json
var writesSchemaJSON string

//go:embed schema/api_surface.schema.json
var apiSurfaceSchemaJSON string

//go:embed schema/operations.schema.json
var operationsSchemaJSON string

//go:embed schema/cli_surface.schema.json
var cliSurfaceSchemaJSON string
