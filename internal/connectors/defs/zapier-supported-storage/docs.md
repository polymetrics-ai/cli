# Overview

Zapier Supported Storage is Zapier's simple key/value store for Zaps. This bundle reads its stored
records through `GET {base_url}/api/records`. It migrates
`internal/connectors/zapier-supported-storage` (the hand-written connector); the legacy package
stays registered and unchanged until wave6's registry flip. Read-only: `capabilities.write` is
`false` and this bundle ships no `writes.json`.

## Auth setup

Provide `secret` (secret), sent as the `secret` query parameter on every request via `auth:
[{"mode": "api_key_query", "param": "secret", "value": "{{ secrets.secret }}"}]`, matching legacy's
`connsdk.APIKeyQuery("secret", secret)`. Never logged.

## Streams notes

The single `records` stream reads `GET /api/records`, extracting records from the top-level
`records` array. Fields (`id`, `key`, `value`, `updated_at`) pass straight through via schema
projection — legacy's `mapRecord`-equivalent (`zapier_supported_storage.go:79-83`) copies these
four fields verbatim with no renaming, so no `computed_fields` are needed. No pagination is
declared — legacy issues a single unpaginated request and emits every record in the response,
matching this bundle's omitted `pagination` block (defaulting to `none`).

**No `check` block is declared** (deliberate, not an oversight): legacy's `Check` performs
config/secret presence validation only (`base_url` well-formedness, `secret` non-empty) and issues
**no HTTP request at all** (`zapier_supported_storage.go:33-47`). The engine's `Check` dispatch,
when `streams.json`'s `base.check` is unset, still resolves auth (validating `secret` is
configured) via `newRuntime` but performs no network call and returns `nil` — the exact, honest
parity representation of legacy's no-network-call Check. `conformance`'s `check_fixture` dynamic
check structurally Skips (no fixture needed) when `HTTP.Check == nil`.

## Write actions & risks

None. Zapier Supported Storage is modeled read-only in legacy (`capabilities.Write: false`); this
bundle matches that exactly and ships no `writes.json`.

## Known limits

- Only the single `records` read stream is modeled, matching legacy's sole stream. Write/delete
  endpoints for Zapier Storage are out of scope for wave2; see `api_surface.json`'s `excluded`
  entries (`destructive_admin` for bulk delete, `out_of_scope` for record creation).
