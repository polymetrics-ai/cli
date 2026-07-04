# Overview

Zapier Supported Storage is Zapier's simple key/value store for Zaps. This bundle reads its stored
records through `GET {base_url}/api/records` and writes through `PATCH`/`DELETE` on the same single
resource. It migrates `internal/connectors/zapier-supported-storage` (the hand-written connector);
the legacy package stays registered and unchanged until wave6's registry flip.

**Pass B full-surface expansion**: the real API was re-verified directly by live probing
(`curl` against `https://store.zapier.com/api/records` with a fresh, never-shared UUID4 secret;
every test key created during probing was deleted immediately afterward) since Zapier publishes no
OpenAPI/swagger document for this API — the help-center article alone does not enumerate every
method/action. `capabilities.write` is now `true`; see `writes.json` and `api_surface.json`.

## Auth setup

Provide `secret` (secret), sent as the `secret` query parameter on every request via `auth:
[{"mode": "api_key_query", "param": "secret", "value": "{{ secrets.secret }}"}]`, matching legacy's
`connsdk.APIKeyQuery("secret", secret)`. Never logged. The same `secret` scopes both reads and
writes to one caller's storage bucket; there is no separate write-scoped credential.

## Streams notes

The single `records` stream reads `GET /api/records`, extracting records from a top-level
`records` array via schema projection. No pagination is declared — the real API returns the
entire bucket in one response with no paging support at all, matching this bundle's omitted
`pagination` block (defaulting to `none`).

**No `check` block is declared** (deliberate, not an oversight): legacy's `Check` performs
config/secret presence validation only (`base_url` well-formedness, `secret` non-empty) and issues
**no HTTP request at all** (`zapier_supported_storage.go:33-47`). The engine's `Check` dispatch,
when `streams.json`'s `base.check` is unset, still resolves auth (validating `secret` is
configured) via `newRuntime` but performs no network call and returns `nil` — the exact, honest
parity representation of legacy's no-network-call Check. `conformance`'s `check_fixture` dynamic
check structurally Skips (no fixture needed) when `HTTP.Check == nil`.

## Write actions & risks

- **`set_record`** (`PATCH /api/records`, `action: set_value_if`) — creates or overwrites a single
  key/value pair; `data.only_if_value` optionally makes the write conditional on the key's current
  value (live-verified: `null` matches an absent/never-set key). External mutation, no approval
  required.
- **`increment_record`** (`PATCH /api/records`, `action: increment_by`) — atomically increments a
  numeric-valued key by `data.amount`, creating the key at `amount` if it does not yet exist
  (live-verified). External mutation, no approval required.
- **`delete_record`** (`DELETE /api/records?key={{ record.key }}`) — deletes exactly one key;
  idempotent, a repeat delete or a delete of an absent key both live-verify as `200 {}` (Zapier's
  delete is unconditionally idempotent, so `missing_ok_status` is declared defensively even though
  a live 404 was never observed for this endpoint).
- **`delete_all_records`** (`DELETE /api/records`, no query param) — irreversibly wipes every key
  in the caller's bucket in one call (live-verified: a bare `DELETE` with no `key` clears the whole
  store). Destructive; requires explicit confirmation (`confirm: destructive`).

## Known limits

- **`GET /api/records`'s real wire shape is a bare flat JSON object (`{key: value, ...}`), NOT
  `{"records": [...]}`.** Live probing confirms the actual response body has no `records` wrapper
  at all — every stored key is a top-level property of the response object itself. This bundle's
  pre-existing `streams.json` (`records.path: "records"`) and its legacy Go connector
  (`zapier_supported_storage.go`'s `connsdk.RecordsAt(resp.Body, "records")`) both inherited this
  incorrect assumption before this Pass B review; per the parity-deviation meta-rule
  (`docs/migration/conventions.md` §5), this bundle's read behavior is left UNCHANGED rather than
  silently "fixed" without an ENGINE_GAP mechanism to correctly express the fix. The reason a
  straightforward fix is not applied: the real shape is a keyed object whose VALUES are typically
  bare scalars (`"foo": "bar"`), not nested objects. The engine's `records.keyed_object` dialect
  (`docs/migration/conventions.md` §3) explodes each value of a keyed object into a record only
  when that value is itself a JSON object — a scalar-valued entry is silently SKIPPED, which would
  make every plain string/number-valued key vanish from the stream, a worse outcome than the
  current (also wrong, but at least non-lossy against its own tests) `"records"`-path assumption.
  Modeling "one record per key of a flat scalar-valued object" is an **ENGINE_GAP**: the dialect has
  no declarative way to flatten a scalar-keyed object into `{key, value}` pairs. Until the engine
  gains that primitive, `records` is read-parity-locked to its pre-existing (legacy-inherited, real-
  API-inaccurate) shape rather than changed to something equally imperfect in the opposite
  direction. A caller pointing this bundle's `records` stream at the real live API today will see
  either zero records (an empty/whole-object body has no `records` key to select) or, if the raw
  API response happens to nest an object-valued key some caller stored, that one key's own fields
  misread as if they were a single flat "record" — this is the honest, documented gap, not a hidden
  behavior change.
- Full API surface is otherwise covered: the real API is exactly one resource path (`/api/records`)
  with 4 HTTP methods, one of which (`PATCH`) multiplexes 7 sub-operations via a required `action`
  field; see `api_surface.json` for the complete enumeration and exclusion reasoning (2 actions
  covered as writes, 1 GET variant and 1 PATCH action excluded as `duplicate_of` an already-covered
  shape, 4 narrower PATCH actions and the multi-key POST excluded as `out_of_scope` for this pass).
- `set_record`/`increment_record`/`delete_record`/`delete_all_records` write one call per record,
  matching the engine's one-request-per-record write model; there is no bulk/batch write endpoint
  in the real API to consolidate multiple keys into a single request (the closest, `POST`'s
  multi-key-merge body, cannot be expressed at all per the `api_surface.json` `POST` exclusion
  above).
- **`delete_record`'s fixture cannot assert its `?key=` query parameter.** The action's `path`
  (`/api/records?key={{ record.key }}`) is genuinely sent with the query string (verified by
  tracing `engine.InterpolatePath`/`joinURL`, which treat `path` as an opaque, already-`{{ }}`-
  resolved string concatenated onto the base URL — a literal `?key=...` suffix survives untouched),
  but `conformance`'s `write_request_shape` fixture harness (`internal/connectors/conformance/
  dynamic.go`'s `writeExpectation`) only exposes `method`/`path`/`body` for assertion, and compares
  `path` against `capturedRequest.Path`, which is `r.URL.Path` — the query string is captured
  separately (`capturedRequest.Query`) but never compared. `fixtures/writes/delete_record.json`'s
  `expect.path` is therefore `/api/records` (matching what the harness actually checks), not the
  full `/api/records?key=fixture_key` the real request line carries. This is a fixture-harness
  expressiveness gap, not a defect in the write action itself.
