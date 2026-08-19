# Issue 3886 — Dynamic schema discovery foundation: context

**Gathered:** 2026-08-06
**Status:** Ready for TDD implementation
**Issue:** https://github.com/polymetrics-ai/cli/issues/3886

## Phase boundary

Build the shared dynamic catalog foundation and prove it on HubSpot. This is a
foundation issue, deliberately separate from a connector migration lane: the
only connector adoption in scope is HubSpot. No existing connector's static
schema is changed.

## Locked decisions

- `connectors.Connector.Catalog(ctx, cfg)` is the discovery entry point. No
  parallel connector-facing discovery method or generic raw HTTP surface is
  introduced.
- A shared package owns object enumeration, bounded description fetches,
  cancellation, retry/backoff/jitter, progress events, partial-completion
  status, cache handling, and JSON-Schema assembly. An adopter supplies its
  provider calls and its provider-field converter only.
- Discovery produces the local draft-07 dialect: schema properties plus the
  normal key/cursor markers and stream-sync extensions (`x-stream_name`,
  `x-supported_sync_modes`, `x-default_sync_mode`,
  `x-source_defined_primary_key`, `x-source_defined_cursor`, and
  `x-default_cursor_field`). The raw schema preserves this complete contract;
  Go's existing stream view still derives its normal key/cursor fields from
  the established extensions.
- A catalog stream carries its raw draft-07 record schema. Static bundle
  cataloging and discovery both construct `connectors.Stream` through that
  one schema-to-stream path, making the field/primary-key/cursor view
  interchangeable rather than merely similar.
- The cache key is `connector name + RuntimeConfig.CoordinationIdentity`
  opaque auth-cohort projection. It never uses `Secrets`, a raw credential
  identifier, a credential revision, or a token-derived value.
- The existing `pm catalog refresh --connection …` is the explicit refresh
  path. A normal `pm catalog show` returns the persisted cache only while it
  is fresh; if it is beyond its declared TTL, its JSON status and human output
  explicitly say `stale` and direct the user to refresh. Stale data is never
  represented as current.
- Durable cache files are account-scoped (`connector + AuthCohortKey`) and
  referenced from state, while `CatalogSnapshot` remains a connection-scoped
  presentation. The durable file must precede the state pointer and contain
  only schemas/status/timestamps.
- HubSpot's live path uses the documented read-only discovery endpoints:
  custom schema listing (`/crm-object-schemas/v3/schemas`) and per-object
  properties (`/crm/v3/properties/{objectType}`). Its declared standard
  fallback objects avoid an empty catalog if global custom-schema discovery
  is unavailable; successful global discovery always includes account-defined
  custom `objectTypeId` values the source code cannot know in advance.
- No credentialed run is performed. The PR must state this plainly; fixtures
  establish behavior only, not live-provider compatibility.

## Evidence read before planning

| Claim | Evidence |
| --- | --- |
| The sole connector contract already has credential-aware `Catalog` | `internal/connectors/connectors.go` |
| Engine catalogs static bundles via `legacyStreamOf` | `internal/connectors/engine/connector.go` |
| Dynamic-schema bundles may have no static stream list | `internal/connectors/conformance/static.go:497`, `engine/bundle.go:loadStreams` |
| Existing dynamic flags are structural, hard-coded catalog examples | `native/{amazon-sqs,faker,postgres,tally-prime}` |
| Saved catalog refresh/show exists but does not label staleness | `internal/app/app.go:RefreshCatalog`, `ShowCatalog` |
| Opaque coordination identity exists exactly for account-scoped state | `internal/connectors/coordination_identity.go`, #3863 plumbing |
| Current HubSpot bundle has zero static streams and no enabled reads | `internal/connectors/defs/hubspot/*` |
| Schema listing returns tenant custom object IDs/FQNs and properties are per-object | HubSpot official schemas/properties/object API documentation, retrieved 2026-08-06 |
| Ruby prior art supplies bounded parallel description fetches, fallback, retries, and schema assembly | `ruby_connectors/.../salesforce_connector/cataloger.rb`, `core/api_cataloger.rb` |

## Explicit exclusions

- No live credential request, credentialed check, token log, fixture token, or
  token-derived cache/index value.
- No dependency; use the Go standard library plus existing `connsdk`.
- No generic HTTP/shell/SQL write interface and no reverse-ETL change.
- No HubSpot operations beyond account-discovered object catalog + the bounded
  object collection read needed to make each discovered `{objectType}` stream
  genuinely invocable.
- No broad HubSpot operation-ledger promotion and no unrelated bundle edits.

## Manual GSD fallback

`scripts/gsd doctor`, all five required command source resolutions, and
`go run ./cmd/agentcontractgen check` passed. This issue is not a numbered
roadmap phase and the canonical single-worker contract forbids role spawning,
so the generated `discuss-phase` and `plan-phase --tdd` workflows are being
executed inline. The plan, ledger, verification checklist, prompts, and later
review record preserve the same lifecycle evidence.
