# PLAN — Issue 3886: dynamic schema discovery foundation

## GSD and skill evidence

- Commands resolved: `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
  `verify-work`, and `code-review`; prompt generation and
  `agentcontractgen check` passed.
- Inline/manual GSD fallback is used because this is an unnumbered foundation
  issue and the project contract requires one canonical worker.
- Loaded skills: `golang-how-to`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
  `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`,
  `golang-cli`, and `golang-documentation`.

## Owned paths

- `internal/connectors/discovery/**` — shared driver and unit tests.
- `internal/connectors/connectors.go` plus catalog-shape tests — single
  schema-to-stream contract and discovery status.
- `internal/connectors/engine/{bundle,connector}.go` plus tests — preserve
  raw static schemas and send static catalogs through the shared path.
- `internal/connectors/native/hubspot/**` and the named nativeset wiring —
  the one proving adoption.
- `internal/app/{app,types,catalog_storage}.go` plus targeted tests —
  account-scoped persistent catalog lookup, durable writes, and refresh/stale
  state.
- `internal/cli/**`, `docs/cli/catalog.md`, `website/content/docs/etl.mdx`,
  and generated docs only if the established docs generator changes them.
- `internal/connectors/defs/hubspot/{metadata,spec,streams,docs,api_surface}.json`
  only where necessary to make the new HubSpot catalog/read claim truthful.

## Design

### One catalog shape

`connectors.Stream` gains a raw `Schema` and the catalog gains a discovery
status block. A new schema-to-stream helper validates/derives the normal
properties, primary key and cursor from raw draft-07. The static engine stores
the raw schema bytes on load and calls that helper; discovery does the same.
That makes an object from a JSON bundle and an object returned by a provider
pass through exactly one downstream representation.

### Shared discovery driver

`internal/connectors/discovery` has a small consumer-owned provider contract:
list objects and describe one object. The other input is a field converter.
The driver owns object deduplication/order, max-ten worker pool, cancellation,
five-attempt exponential retry with bounded jitter for typed rate-limit errors,
100-object progress heartbeats, raw-schema assembly, and result sorting.

It never emits the provider's raw error text. A discovery status exposes only
safe aggregate facts: complete/partial, cache state, timestamps, failure
stage/object identifier, and fallback use. A failed global list invokes the
connector-declared fallback list. Failed object descriptions retain the
successful streams but return `complete=false` plus a named partial status.

### Cache and refresh

The driver cache key is connector name plus the opaque `AuthCohortKey`
projection. This is deliberately an account cache: two connections may point
at the same endpoint/account, so they reuse discovery rather than paying for
hundreds of duplicate calls or drifting into separate schema copies. Rotation
does not reset this key. Cached entries carry their discovery time and TTL.
Fresh cache results are returned without provider calls; expired entries are
never returned by the in-process cache. `pm catalog refresh` bypasses the
fresh cache. The opaque key is never rendered.

Persistent catalog data is one schema-only file per `connector +
AuthCohortKey`, referenced by a small `state.json` record rather than embedded
in that document. Connections resolve their endpoint identity and reference
the matching account catalog; `CatalogSnapshot` remains the public
connection-scoped response shape. Narrow app accessors own catalog lookup and
durable writes so destination catalogs can reuse the same mechanism when
reverse ETL needs them. The catalog file is fsynced (including directory
metadata from actual successful mkdir operations) before its state reference
is atomically committed. A crash may leave an unreferenced catalog file, but
can never leave state referring to a non-durable file. Files contain only
catalog schemas/status/timestamps: no configuration, credential, token,
cache-key, binding, or account identifier. It does not assume the current shared
final-table behavior is safe: today
`localWarehouseTablePath` omits the connection while stream state and raw
paths include it, so two connections can collide in one final table. That
unrelated migration must make final tables connection-scoped before dynamic
schemas are safe for production. This discovery design remains correct when
that path changes because it never merges, unions, or shares schemas across
connections.

### HubSpot adapter

The native adapter embeds the engine connector for normal definition,
authentication, checks, and unsupported writes, overriding only `Catalog` and
the known `{objectType}` collection reader. Its provider declaration calls
the documented custom-schema list endpoint, combines returned custom type IDs
with a declared standard-object fallback baseline, and describes each object
through the documented properties endpoint. `hasUniqueValue` supplies a
primary key when present; otherwise the provider-described `hs_object_id` is
used only if present. A cursor is exposed only when the provider described the
matching last-modified property.

The bounded read uses the discovered object type identifier in the fixed
HubSpot collection path and only requests fields the discovery result
described. It flattens the documented `properties` envelope and uses the
existing bounded requester/pagination machinery. An arbitrary unknown custom
object type appears in fixture discovery, can be read, and is never hard-coded
as an object-type enum.

## TDD slice order

1. **RED:** catalog schema helper tests prove static/dynamic interchangeability
   and status JSON behavior before implementation.
2. **RED:** driver tests cover bounded worker count, cancellation, rate-limit
   retry+jitter, progress, global-list fallback, partial descriptions, safe
   errors, fresh/stale/refresh cache state, account-cache reuse, and durable
   file-before-pointer ordering.
3. **GREEN:** implement the helper and dependency-free driver, then route an
   existing static test bundle through it without regressions.
4. **RED/GREEN:** add HubSpot HTTP fixtures featuring a custom type never in
   source, its custom property, rate limit response, and paginated collection
   read. Implement only the thin provider/field converter/reader declaration.
5. **RED/GREEN:** persist/retrieve snapshots by opaque account identity and
   surface stale status through `pm catalog show` / `refresh`, updating docs
   and generated manual outputs if required.
6. **Regression:** run focused package, engine, app, native/registry, CLI,
   conformance, docs and boundary gates; run generated-surface checks.

## Definition of done

- A dynamically discovered HubSpot custom type and property are cataloged and
  read through a tested fixed, bounded route.
- Static and discovered schemas enter the same catalog stream conversion and
  are asserted interchangeable downstream.
- Partial/fallback/cache/stale outcomes are unmistakable in structured and
  human catalog output; expiry requires `pm catalog refresh` and no old data
  silently passes as current.
- No provider/credential value reaches logs, errors, fixtures, cache keys,
  state output, docs, or PR text. No new dependency is added.
- PR evidence says `live HubSpot run: not performed (no supplied credential)`.
