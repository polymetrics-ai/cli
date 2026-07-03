// Package dynamodb implements the Tier-3 native Amazon DynamoDB source
// connector (docs/migration/conventions.md §1 Tier 3, following the
// mandated golden component split copied from internal/connectors/native/
// postgres/): connector.go (entry, wiring), connection.go (config
// resolution, AWS SigV4 request signing), reader.go (Read/InitialState/Scan
// pagination), cataloger.go (Catalog + fixture generation, attribute-value
// flattening). Each file is well under the design's <400-line cap.
//
// DynamoDB is genuinely protocol-native, not a REST/JSON HTTP API a
// declarative Tier-1 bundle can express: every request is a POST to the
// service root ("/") carrying an X-Amz-Target header naming the RPC
// operation (DynamoDB_20120810.Scan) and is authenticated with AWS
// Signature Version 4 (SigV4) — a canonicalized-request HMAC-SHA256 signing
// scheme the engine's declarative auth dialect (bearer/basic/api_key/
// oauth2_client_credentials) has no shape for, and which docs/migration/
// conventions.md §6 names explicitly as a Tier-3 trigger ("does it need
// direct control over connection lifecycle... SQL, message queues,
// filesystems, CDC" / SigV4 is listed alongside HMAC as the AUTH_COMPLEX
// example the escape-hatch decision tree gives for "an auth scheme the
// dialect/hooks genuinely cannot express"). Records also arrive as
// DynamoDB's own typed AttributeValue envelope ({"S":...}/{"N":...}/
// {"BOOL":...}/{"M":...}/{"L":...}/{"NULL":...}), which needs recursive
// unwrapping no declarative schema-projection mechanism provides.
//
// Unlike a Tier-1/Tier-2 declarative bundle, this package implements
// connectors.Connector directly: Check/Catalog/Read are hand-written Go, not
// derived from streams.json (there is none — this bundle ships no
// streams.json since a DynamoDB table's item shape is only known at
// Scan-time and legacy itself only ever published a single generic "items"
// stream. Catalog is a static single-stream constant, though — NOT a
// discovered-at-runtime shape like postgres's information_schema
// enumeration — so capabilities.dynamic_schema stays false, matching
// legacy's Catalog() which returns the identical fixed one-stream
// definition every call, no live discovery request involved).
//
// Capabilities:
//   - Check:   validates config/credential presence and, outside fixture
//     mode, that a usable endpoint resolves (connection.go); legacy never
//     issued a live network call from Check either.
//   - Catalog: a static one-stream ("items") catalog (connector.go), ported
//     verbatim from legacy's Catalog.
//   - Read:    a signed DynamoDB_20120810.Scan JSON-RPC POST per page,
//     paginating on LastEvaluatedKey/ExclusiveStartKey, flattening each
//     item's AttributeValue envelope into a plain connectors.Record
//     (reader.go/cataloger.go).
//   - Write:   not implemented; this is a read-only source for parity with
//     the legacy package (legacy's Write always returns
//     ErrUnsupportedOperation, and its Metadata already declared
//     Capabilities.Write: false — the catalog_data.json "destination-
//     dynamodb"/"source-dynamodb" slug labels are stale Airbyte-slug
//     residue, not a signal that legacy ever implemented writes; see
//     defs/dynamodb/docs.md's "Known limits").
//
// There is no cdc.go: unlike postgres (which documents a genuine
// pglogrepl-gated CDC stub already recorded as future scope), legacy
// dynamodb.go has no CDC concept at all (DynamoDB Streams is a distinct,
// separate AWS API this connector never touched), so a CDC file is simply
// omitted rather than padded with an inapplicable stub (mirrors
// native/bing-ads's identical "there is no cdc.go" precedent).
//
// A mode=fixture config (cfg.Config["mode"]=="fixture") short-circuits all
// network access so the conformance harness and unit tests can run with no
// live AWS credentials: in fixture mode Check succeeds, Catalog returns the
// static stream, and Read emits two canned rows (ported verbatim from
// legacy's readFixture).
//
// NO init()/RegisterFactory/RegisterNativeLive call exists in this package
// (enforced by a grep-guard test, dynamodb_test.go TestNoInitRegistration,
// mirroring native/postgres's and native/bing-ads's identical guard) — the
// registration flip that wires native/dynamodb into the production registry
// is a later-wave change; this package only builds and tests standalone.
package dynamodb

import (
	"context"
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

	// Client overrides the HTTP client used for signed Scan requests. Left
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
	return Connector{Base: engine.NewBase(b)}
}

// Metadata overrides engine.Base's bundle-synthesized Metadata with the
// legacy-shaped description text, matching the pre-migration
// connectors.Metadata field-for-field (parity target); Capabilities are
// still whatever the bundle's metadata.json declares (single source of
// truth for capability flags), so this override only refines
// Description/DisplayName wording, never capability semantics.
func (c Connector) Metadata() connectors.Metadata {
	m := c.Base.Metadata()
	m.Description = "Reads DynamoDB table items through the AWS JSON HTTP API (DynamoDB_20120810.Scan), authenticated with hand-rolled AWS Signature Version 4 request signing. Read-only source; no write support."
	return m
}

// Write is unsupported: this is a read-only source connector, matching
// legacy dynamodb.go's Write (capabilities.write is false in
// defs/dynamodb/metadata.json — the catalog_data.json "destination-
// dynamodb"/"source-dynamodb" slugs are stale Airbyte-slug residue and do
// not reflect any write capability legacy ever implemented).
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
