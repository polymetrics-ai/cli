// Package dynamodb implements the Tier-3 native Amazon DynamoDB connector.
// DynamoDB is protocol-native rather than REST: every request is a signed
// JSON-RPC POST to "/" with an X-Amz-Target operation header and AWS SigV4
// authentication. Reads and writes also use DynamoDB AttributeValue envelopes,
// so this package keeps connector-local Go builders instead of exposing raw
// HTTP, raw PartiQL, arbitrary expression, endpoint, or body passthrough.
//
// Capabilities:
//   - Check validates config/credential presence and fixture mode without live
//     provider calls.
//   - Catalog is fixed from defs/dynamodb/streams.json; dynamic schema remains
//     false because table item shapes are caller data.
//   - Read covers bounded Scan/Query, metadata/list streams, and DynamoDB
//     Streams helper reads with documented pagination tokens.
//   - OperationDirectRead exposes only three keyed direct reads
//     (GetItem/BatchGetItem/TransactGetItems) through closed command fields.
//   - Write executes typed reverse-ETL actions from writes.json after the app's
//     plan -> preview -> approval gate; destructive/admin actions declare
//     destructive confirmation where applicable.
//   - CDC reads DynamoDB Streams through typed shard iterator/GetRecords calls.
//
// A mode=fixture config (cfg.Config["mode"]=="fixture") short-circuits all
// network access so conformance and unit tests run without live AWS credentials.
// Registration is owned by the central connector registry; this package does
// not use init-time side effects.
package dynamodb

import (
	"net/http"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

// itemsStreamName is the single static stream this connector publishes,
// matching legacy dynamodb.go's Catalog/Read hardcoded "items" name.
const itemsStreamName = "items"

// Connector is the Tier-3 native pm DynamoDB source connector. It embeds
// engine.Base for Name()/Metadata()/Definition(), synthesized from the
// defs/dynamodb bundle loaded once at construction (New), and implements
// Check/Catalog/Read/Write itself (connection.go/cataloger.go/reader.go)
// since DynamoDB's SigV4 signing and typed AttributeValue wire shape have
// no declarative equivalent.
type Connector struct {
	engine.Base
	bundle engine.Bundle

	// Client overrides the HTTP client used for signed JSON-RPC requests. Left
	// nil in production; injectable for tests (mirrors legacy's identical
	// field).
	Client *http.Client
	// Now overrides the clock used to derive SigV4's date/x-amz-date
	// headers. Left nil in production (time.Now is used); injectable for
	// deterministic signature assertions in tests (mirrors legacy's
	// identical field).
	Now func() time.Time
}

// New returns the DynamoDB connector as a connectors.Connector, loading its
// Definition()/Metadata() from the embedded defs/dynamodb bundle. New
// panics if the bundle fails to load — the same "build-time guaranteed by
// connectorgen validate + tests" invariant native/postgres's New documents,
// since a bundle that fails to load here indicates a broken build, not a
// runtime/user error.
func New() Connector {
	b, err := engine.Load(defs.FS, "dynamodb")
	if err != nil {
		panic("native/dynamodb: failed to load defs/dynamodb bundle: " + err.Error())
	}
	return Connector{Base: engine.NewBase(b), bundle: b}
}

func (c Connector) definitionBundle() engine.Bundle {
	if c.bundle.Name != "" {
		return c.bundle
	}
	b, err := engine.Load(defs.FS, "dynamodb")
	if err != nil {
		panic("native/dynamodb: failed to load defs/dynamodb bundle: " + err.Error())
	}
	return b
}

// Metadata overrides engine.Base's bundle-synthesized Metadata with the
// legacy-shaped description text, matching the pre-migration
// connectors.Metadata field-for-field (parity target); Capabilities are
// still whatever the bundle's metadata.json declares (single source of
// truth for capability flags), so this override only refines
// Description/DisplayName wording, never capability semantics.
func (c Connector) Metadata() connectors.Metadata {
	m := c.Base.Metadata()
	m.Description = "Reads, changes, and writes Amazon DynamoDB tables through the AWS JSON API with connector-local SigV4 signing, typed operation schemas, bounded reads, DynamoDB Streams changefeed support, and reverse-ETL write gates."
	return m
}

func (c Connector) Manifest() connectors.Manifest {
	return engine.New(c.definitionBundle(), nil).Manifest()
}
