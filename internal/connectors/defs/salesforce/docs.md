# Overview

Salesforce is migrated as a Tier-1 declarative bundle. Legacy
(`internal/connectors/salesforce/salesforce.go`) is a pure `connsdk`-HTTP connector — it builds a
`connsdk.Requester` with Bearer auth and issues plain `GET` requests against the Salesforce REST
API (object metadata plus allow-listed Account/Contact/Lead SOQL queries); there is no SQL/queue/
SDK protocol, no non-declarative auth flow, and no write path (`Write` always returns
`connectors.ErrUnsupportedOperation`). Every behavior legacy has is expressible in
`streams.json`/`spec.json`/schemas alone, so this bundle needs no Go hooks at all (Tier 2) and no
native component split (Tier 3) — despite the wave2 catalog inventory mislabeling this connector's
`runtime_kind` as `destination_go`, legacy is read-only and was never a destination.

## Auth setup

Provide a Salesforce OAuth access token via the `access_token` secret; it is used only for Bearer
auth (`Authorization: Bearer <access_token>`) and is never logged. `instance_url` (required) is the
org/instance base URL (e.g. `https://yourinstance.my.salesforce.com`), matching legacy's
`instance_url`/`base_url` config fallback (this bundle keeps only `instance_url`, the primary key
legacy tries first — see Known limits). `api_version` defaults to `v60.0` and is interpolated
directly into every request path.

## Streams notes

- **`sobjects`** — `GET /services/data/{api_version}/sobjects` (Describe Global), records at
  `sobjects`, no pagination (matches legacy: a single unpaginated request). `computed_fields` maps
  the real Salesforce wire fields `name`/`label` to `qualified_api_name`/`label`.
- **`accounts`** / **`contacts`** / **`leads`** — `GET /services/data/{api_version}/query` with a
  fixed allow-listed SOQL `q` query string (`SELECT Id, Name[, Email], LastModifiedDate FROM
  <Object> ORDER BY LastModifiedDate ASC`, exactly legacy's `salesforceStreamEndpoints` SOQL text),
  records at `records`. Pagination follows Salesforce's own `nextRecordsUrl` convention
  (`pagination.type: next_url`, `next_url_path: nextRecordsUrl`) — the response body's
  `nextRecordsUrl` is an absolute-ish next-page URL, consumed identically to legacy's own
  `path = next; query = nil` loop. `computed_fields` renames each raw SOQL record's PascalCase
  fields (`Id`/`Name`/`Email`/`LastModifiedDate`) to the schema's snake_case names, matching
  `salesforceNamedObjectRecord` field-for-field. No `incremental` block is declared: legacy's
  catalog never sets `CursorFields` for these streams and the read path has no incremental filter
  parameter — every read is a full SOQL sweep ordered by `LastModifiedDate ASC`, exactly matching
  legacy (§8 rule 2 of `conventions.md`).

## Write actions & risks

None. Legacy's `Write` always returns `connectors.ErrUnsupportedOperation`; this bundle declares
`capabilities.write: false` and ships no `writes.json`, matching legacy exactly (the wave2 catalog
inventory's `destination-salesforce` slug is a mislabel — see Overview).

## Known limits

- **`api_version` normalization is not modeled.** Legacy accepts either a bare version
  (`"60.0"`) or an already-prefixed version (`"v60.0"`), trimming slashes and prepending `v` if
  missing (`salesforceVersion`). The engine's path interpolation has no string-transform/prefix
  primitive, so this bundle requires the config value to already carry the `v` prefix (the
  `"v60.0"` default matches legacy's own default exactly). A caller who previously configured a
  bare `"60.0"` must add the `v` prefix when migrating. Strictly a config-input-shape narrowing,
  never a data/behavior change for any already-`v`-prefixed input.
- **`instance_url`-only, no `base_url` fallback.** Legacy tries `instance_url` first, falling back
  to a `base_url` config key. This bundle declares only `instance_url` (the primary, first-tried
  key) since a single spec property cannot express an either-or fallback declaratively; a caller
  relying solely on the undocumented `base_url` fallback must switch to `instance_url`.
- **`sobjects`' `qualified_api_name`/`label` fields use the real Salesforce Describe Global wire
  shape (`name`/`label`), not legacy's defensive 3-way `first(qualifiedApiName, QualifiedApiName,
  name)` / `first(label, Label)` fallback chase.** `computed_fields` resolves exactly one bare
  reference per output field with no fallback-chain primitive. Salesforce's real, documented
  Describe Global response uses `name`/`label` (lowercase-camelCase) — the third candidate in
  legacy's own fallback list, and the one any live Salesforce org actually sends — so this bundle's
  single-reference mapping matches real traffic identically; only legacy's synthetic
  `readFixture`/unit-test PascalCase (`QualifiedApiName`/`Label`) shape, which no live Salesforce
  response ever sends, is not reproduced. See `docs/migration/conventions.md`'s parity-deviation
  ledger meta-rule (never changes emitted data for any input legacy's LIVE code path would accept).
- **`max_pages` is not modeled.** Legacy exposes a `max_pages` config override
  (0/all/unlimited or a positive integer request-count cap) for the SOQL query streams. The
  engine's `next_url` paginator has no config-driven page-size or request-count-cap knob at all
  (the same limitation ledgered for aircall/bitly's `next_url` streams) — `max_pages` is therefore
  not declared in `spec.json` as genuinely dead config (F6, `conventions.md`); every SOQL stream
  reads to full exhaustion (`nextRecordsUrl` absent/empty), which is legacy's own default
  (`max_pages` unset) behavior.
- **`next_url` fixtures are single-page, per the sanctioned exception (`conventions.md` §4).** A
  `next_url` stream's next-page URL is the replay server's own runtime address, unknown until the
  harness picks a port — a static fixture file cannot embed the correct absolute URL for a second
  page. Every stream in this bundle ships a single-page fixture (satisfies `fixtures_present`/
  `read_fixture_nonempty`); `pagination_terminates` passes on the first stream (`sobjects`, which is
  genuinely unpaginated) rather than proving 2-page `next_url` pagination specifically. Real 2-page
  `nextRecordsUrl`-following correctness for this exact request shape is proven by legacy's own
  existing test (`internal/connectors/salesforce/salesforce_test.go`'s
  `TestReadPaginatesAndAuthenticates`, which drives a real 2-page `httptest.Server` and asserts the
  second page is requested via the served `nextRecordsUrl`), plus the engine's own generic
  `next_url` paginator and read-path integration tests
  (`internal/connectors/engine/paginate_test.go`, `internal/connectors/engine/read_test.go`).
