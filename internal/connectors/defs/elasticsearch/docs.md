# Overview

Reads Elasticsearch index metadata and documents through the REST API. This bundle migrates
`internal/connectors/elasticsearch/elasticsearch.go` (the legacy hand-written connector, which stays
registered and unchanged until wave6's registry flip). Despite the catalog labeling this connector
`destination-elasticsearch`/`source-elasticsearch` (`docs/migration/inventory.json`'s
`runtime_kind: "destination_go"`), the legacy Go source is the ground truth and it is **read-only**:
`Capabilities.Write` is `false` and `Write()` unconditionally returns
`connectors.ErrUnsupportedOperation` — there is no write path to migrate. The catalog label is
misleading here; this bundle declares `capabilities.write: false` and ships no `writes.json`,
matching legacy exactly.

**Tier justification**: legacy is a pure `connsdk.Requester`-based HTTP connector (no SQL/queue/SDK
protocol, no direct connection-lifecycle management) — the default Tier-1 declarative bundle is the
right target. Two narrow gaps in the declarative dialect (composite-secret auth header encoding, and
`_source` object flattening) are not expressible in `streams.json` alone, so this bundle escalates to
**Tier 2**: `internal/connectors/hooks/elasticsearch/hooks.go` implements `AuthHook` +
`RecordHook` (2 interfaces, well under the ~300-line soft cap) — see "Auth setup" and "Streams notes"
below for exactly what each hook covers and why the gap is real, not a convenience shortcut.

## Auth setup

Legacy's `auth()` (`elasticsearch.go:186-198`) tries three credential shapes in order:

1. **API key** — `apiKeyId` (config) + `apiKeySecret` (secret) both set: sends
   `Authorization: ApiKey base64(apiKeyId:apiKeySecret)`.
2. **Basic auth** — `username` (config) set (API key absent/incomplete): sends
   `Authorization: Basic base64(username:password)` via `connsdk.Basic`.
3. **None** — neither set: no `Authorization` header at all.

This bundle reproduces the identical three-way precedence via `streams.json`'s `base.auth` candidate
list (first-match-wins, per conventions.md's dual-auth ordering rule): a `when`-gated `mode: custom`
candidate (config.api_key_id truthy) first, a `when`-gated `mode: basic` candidate (config.username
truthy) second, and an unconditional `mode: none` candidate last.

The API-key candidate requires a hook because `base64(id:secret)` under an `"ApiKey "` prefix has no
declarative expression: the engine's `base64` filter (`interpolate.go`) only ever encodes ONE
resolved `{{ }}` reference at a time (`resolveExpr` processes each `{{ }}` occurrence in a template
independently) — there is no way to concatenate `config.api_key_id` + `":"` + `secrets.api_key_secret`
into one string and base64-encode the JOINED result via templating alone. This is the exact mechanics
`connsdk.Basic` already performs (`base64(user:pass)` under a hardcoded `"Basic "` prefix); no
`AuthSpec` field exposes the equivalent under an arbitrary prefix. `hooks/elasticsearch/hooks.go`'s
`Authenticator` reads `config.api_key_id`/`secrets.api_key_secret` directly and returns
`connsdk.APIKeyHeader("Authorization", base64(id:secret), "ApiKey ")` — byte-for-byte what legacy's
`auth()` builds.

Legacy also accepted `base_url` as a config alias for `endpoint` (`elasticsearch.go:200-204`, tried
`endpoint` first, falling back to `base_url`). This bundle standardizes on `endpoint` only (dead-alias
narrowing, see "Known limits") since `spec.json`'s declared-config-must-be-consumed rule (F6,
conventions.md) forbids declaring a second property no template would ever reference; a caller still
using `base_url` must switch to `endpoint`.

## Streams notes

- **`indices`** — `GET /_cat/indices?format=json`, records at the response body's array root
  (`records.path: ""`), no pagination (matches legacy's single unpaginated request). Primary key
  `index`. The real API's row objects use a literal dotted string key `"docs.count"` (not a nested
  `docs: {count: ...}` object), so schema projection copies it by exact top-level key match with no
  rename needed.
- **`documents`** — `GET /{config.index}/_search`, `offset_limit` pagination (`from`/`size` query
  params, `page_size` default 100 matching legacy's `defaultPageSize`, short-page stop). Records are
  extracted at `hits.hits`; each raw hit is `{_index, _id, _score, _source: {...arbitrary document
  fields...}}`. Legacy's `mapHit` (`elasticsearch.go:142-153`) flattens every key of `_source` onto
  the TOP LEVEL of the emitted record, then stamps `id` from `_id` — dropping `_index`/`_score`/`_id`
  themselves. No dialect primitive flattens a nested object's keys onto the record's top level
  (`RecordsSpec.KeyedObject` explodes a keyed OBJECT of records, a different shape entirely), so
  `hooks/elasticsearch/hooks.go`'s `MapRecord` ports the flatten + `id`-stamp verbatim for the
  `documents` stream only (the `indices` stream passes through this hook untouched). Because
  `_source`'s field set is genuinely arbitrary per index (unknowable at bundle-authoring time,
  exactly like legacy's own schema-free `mapHit`), `schemas/documents.json` declares only the
  guaranteed `id` field with `additionalProperties: true` — matching legacy's own catalog, which
  likewise declares only `{Name: "id", Type: "string"}`.
- Neither stream is incremental in legacy (no cursor field, no `created`/`updated_at`-style
  server-side filter) — neither schema declares `x-cursor-field`.

## Write actions & risks

None. Legacy `elasticsearch.go` is read-only: `Capabilities.Write: false`,
`Write()` always returns `connectors.ErrUnsupportedOperation`. `capabilities.write` is `false` and
this bundle ships no `writes.json`, despite the catalog's `destination-elasticsearch` slug label
(see "Overview" — the label is not corroborated by the legacy Go source, which is this migration's
ground truth per `docs/migration/conventions.md`).

## Known limits

- **Conformance dynamic checks are skipped bundle-wide** (`metadata.json`'s
  `conformance.skip_dynamic`). The API-key `mode: custom` candidate is evaluated FIRST in
  `base.auth`, and conformance's synthetic non-secret config populates every declared property
  (including `api_key_id`) with a truthy `"synthetic-conformance-value"` — so its `when` gate always
  matches during dynamic replay, and `buildCustomAuth` requires `hooks/elasticsearch` to be
  registered in `internal/connectors/hooks/hookset` to resolve at all. Hook registration is
  orchestrator-owned (wave6 scope per `docs/migration/conventions.md` §7's forbidden-files list;
  migration agents never edit `hookset`), so conformance's dynamic replay currently runs with
  `Hooks=nil` for this bundle and cannot exercise `mode: custom` auth or the `documents` stream's
  `RecordHook` flatten. Both are covered instead by `internal/connectors/hooks/elasticsearch/
  hooks_test.go` (unit tests for `Authenticator`'s 2-secret composite encoding and `MapRecord`'s
  flatten+id-stamp, ported directly from legacy's `auth()`/`mapHit`). Static checks
  (`connectorgen validate`, `fixtures_present`, `docs_present`, schema/pk/cursor structural checks)
  still run and pass normally — only the fixture-replay dynamic checks are skipped.
- **`base_url` config alias is not modeled.** Legacy accepted `base_url` as a fallback when
  `endpoint` was unset. This bundle declares only `endpoint` (a second alias property that no
  template consumes is worse than omitting it, per F6/REVIEW.md's "declared config must be consumed"
  rule) — a caller must configure `endpoint` directly. Documented scope narrowing, not a silent
  behavior change for any config shape `endpoint`-based callers already use.
- **Legacy's fixture mode** (`elasticsearch.go:155-176`, `mode=fixture`) emitted 2 synthetic records
  per stream with a `fixture: true` marker field and stream-specific fields
  (`index`/`docs.count` or `id`/`order_number`). This bundle's `mode=fixture`/conformance fixtures
  follow the standard engine dialect fixture-replay convention instead (recorded-real-wire-shape,
  sanitized fixtures under `fixtures/streams/<stream>/page_N.json`) rather than reproducing legacy's
  bespoke in-code fixture generator field-for-field; both are credential-free test paths, not a
  production behavior difference for any real Elasticsearch cluster read.
