// Package amazonsqs implements the Tier-3 native Amazon SQS source connector
// (architecture v2 design §B.7 Tier 3 — see internal/connectors/native/postgres
// for the golden migration reference). It migrates to Tier 3, not a
// declarative Tier-1/Tier-2 bundle, because it signs AWS SQS Query API
// requests with hand-rolled AWS SigV4 (canonical request construction, HMAC
// derivation) and decodes an XML (not JSON) ReceiveMessageResponse body —
// both are named Tier-2-ineligible/protocol-native triggers in
// docs/migration/conventions.md (§1's "signature auth (SigV4, HMAC)" hook
// trigger list, and a 3rd distinct shape — XML decoding — on top of it,
// which alone would still exceed a single AuthHook's scope). The catalog
// label ("native/destination") on this connector is WRONG on both counts:
// see docs.md's Overview for why.
//
// Component split (mirrors postgres's connector.go/connection.go/reader.go,
// minus cataloger.go/cdc.go — this connector's single-stream catalog is a
// two-line literal not worth a dedicated file, and legacy implements no
// CDC path at all, unlike postgres's documented stub):
//   - connector.go (this file) — entry/wiring, Metadata, Catalog, Write stub.
//   - connection.go — SigV4 signing, the signed HTTP request helper, config
//     resolution/validation.
//   - reader.go — Read: the bounded ReceiveMessage poll loop, XML decoding,
//     record mapping.
//
// Still ships a defs bundle (internal/connectors/defs/amazon-sqs/
// {metadata.json,spec.json,api_surface.json,docs.md}) so identity/spec/docs
// stay uniform with every other connector; metadata.json sets
// capabilities.dynamic_schema: true and the bundle ships no streams.json
// (bundle.go's loadStreams only tolerates a missing streams.json when
// dynamic_schema is true) — a structural requirement of the Tier-3 loader,
// not a claim that this connector's fixed single-stream catalog is actually
// schema-discovered at runtime from anywhere (see docs.md's Known limits,
// matching native/faker's identical precedent).
//
// This connector has no incremental cursor state (legacy implements no
// InitialState/StatefulReader either — SQS's ReceiveMessage has no
// timestamp/offset filter), and no CDC path (legacy implements no ReadCDC
// either), so neither connectors.StatefulReader nor connectors.CDCReader is
// implemented here, unlike postgres.
//
// NO init()/RegisterFactory call exists in this package
// (enforced by a grep-guard test, amazon_sqs_test.go TestNoInitRegistration)
// — the registration flip that wires native/amazon-sqs into the production
// registry is a wave6 change; this wave only builds and tests the package
// standalone.
package amazonsqs

import (
	"context"
	"net/http"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

// Connector is the Tier-3 native Amazon SQS source connector. It embeds
// engine.Base for Name()/Metadata()/Definition(), synthesized from the
// embedded defs/amazon-sqs bundle loaded once at construction (New), and
// implements Check/Catalog/Read itself since SigV4 signing and XML
// decoding have no declarative equivalent.
//
// Client and Now are optional test seams, mirroring legacy's exported
// fields of the same name/shape exactly: a nil Client means "build a fresh
// *http.Client with a 60s timeout per request" (connection.go's do); a nil
// Now means "use time.Now" (connection.go's sign). Both are exported so
// tests can inject a fixed clock/http.Client precisely like legacy's own
// test suite does, without a hand-rolled interface neither legacy nor this
// package otherwise needs.
type Connector struct {
	engine.Base

	Client *http.Client
	Now    func() time.Time
}

// New returns the Amazon SQS connector as a connectors.Connector, loading
// its Definition()/Metadata() from the embedded defs/amazon-sqs bundle. New
// panics if the bundle fails to load — a broken build, not a runtime error
// (mirrors native/postgres.New's and native/faker.New's identical
// contract).
func New() Connector {
	b, err := engine.Load(defs.FS, "amazon-sqs")
	if err != nil {
		panic("native/amazon-sqs: failed to load defs/amazon-sqs bundle: " + err.Error())
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
	m.Description = "Reads messages from Amazon SQS via signed ReceiveMessage calls. Read-only; messages are not deleted."
	return m
}

// Catalog returns the fixed single messages stream, matching legacy's
// hand-written Catalog exactly (amazon_sqs.go).
func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{
		Connector: c.Name(),
		Streams: []connectors.Stream{
			{
				Name:        "messages",
				Description: "Messages received from the configured SQS queue. The connector does not delete messages.",
				PrimaryKey:  []string{"message_id"},
				Fields: []connectors.Field{
					{Name: "message_id", Type: "string"},
					{Name: "md5_of_body", Type: "string"},
					{Name: "body", Type: "object"},
					{Name: "sent_timestamp", Type: "string"},
				},
			},
		},
	}, nil
}

// Write is unsupported: this is a read-only source connector, matching
// legacy's Write returning ErrUnsupportedOperation exactly. Capabilities.Write
// is false.
func (c Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}
