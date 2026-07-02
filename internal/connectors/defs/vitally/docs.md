# Overview

Vitally is a customer-success platform. This bundle reads customer-success **accounts** from the
Vitally REST API (`GET {base_url}/resources/accounts`, default `https://rest.vitally.io`). Read-only;
it migrates `internal/connectors/vitally` (188 loc), which stays registered and unchanged until
wave6's registry flip. Engine-vs-legacy parity is tested in
`internal/connectors/paritytest/vitally/parity_test.go` per SPEC.md §6's per-connector
parity-test-package decision.

## Auth setup

Vitally authenticates via a single secret, `basic_auth_header`, which must contain the **entire,
pre-built `Authorization` header value** (e.g. `Basic <base64-encoded apiKey:>`) — not a bare API
key or token. This mirrors legacy exactly: `vitally.go:100-104` builds the requester's
authenticator as `connsdk.APIKeyHeader("Authorization", auth, "")`, i.e. an arbitrary-header
authenticator with an **empty prefix**, so whatever string is configured in `basic_auth_header` is
sent as the `Authorization` header value completely unmodified (no re-encoding, no `Basic ` prefix
added by the connector itself — the caller is expected to have already produced the full header
value, typically by base64-encoding an `apiKey:` pair themselves before configuring the secret).

This bundle reproduces that exact behavior using the engine's `api_key_header` auth mode:
`{"mode":"api_key_header","header":"Authorization","value":"{{ secrets.basic_auth_header }}","prefix":""}`.
The engine's `basic` auth mode was deliberately **not** used here, even though Vitally's own API
uses HTTP Basic auth conceptually — `basic` mode base64-encodes a `username:password` pair at
request time (`connsdk.Basic`), which would require decomposing `basic_auth_header` back into its
constituent parts (information the connector never receives; the secret already IS the encoded
header). Using `api_key_header` with an empty prefix is the byte-exact reproduction, asserted by
`TestParityVitally_AuthorizationHeaderByteExact`.

## Streams notes

One stream: `accounts`, `GET /resources/accounts`, records extracted from the top-level `results`
array (`vitally.go:79`), each record mapped field-for-field to `{id, name, traits}`
(`vitally.go:84`) — `traits` is an arbitrary, unprojected object bag, passed through unchanged on
both sides. No pagination (legacy issues exactly one request per read — no page/cursor param is
ever sent or consumed, `vitally.go:56-89`); no incremental cursor (Vitally's accounts endpoint
exposes no last-modified/cursor field in legacy's mapped shape). `x-primary-key: ["id"]`; no
`x-cursor-field` (matching legacy's full-refresh-only behavior).

## Write actions & risks

None. Vitally's legacy connector is read-only: `Write` always returns
`connectors.ErrUnsupportedOperation` (`vitally.go:91-93`); `capabilities.write` is `false` and this
bundle ships no `writes.json`.

## Known limits

- **The optional `status` query filter is not modeled.** Legacy accepts an optional `status`
  config value and, only when it is non-empty, appends `?status=<value>` to the accounts request
  (`vitally.go:72-74`) — an absent/empty `status` sends no query param at all. The engine
  dialect's `stream.Query` templating has **no absent-key-falsy tolerance** (unlike `auth`'s
  `when` grammar): every `{{ }}` reference in a `query` map value is resolved unconditionally via
  `Interpolate`, so declaring `"status": "{{ config.status }}"` would hard-error on every read
  where `status` is unset — the common case. This is the identical limitation documented for
  searxng's optional passthrough filters (`docs/migration/conventions.md` §3's `stream.Query`
  note, and searxng's own `docs.md` "Known limits"). `status` is therefore **not declared** in
  `spec.json` at all (a declared-but-unwireable key is worse than an absent one — F6,
  `docs/migration/conventions.md`). This is a deliberate, out-of-scope simplification, not a
  defect: no bundle in this dialect can express an optional query passthrough today without a
  conditional-query-templating primitive the engine does not have. If a future pilot/wave hits
  this same shape a third time, it becomes an `ENGINE_GAP` candidate for a wave-level engine
  increment (conventions.md §6's recurrence rule), not a per-connector workaround.
- **Legacy's `fixture` mode is not part of the bundle.** Legacy's `mode=fixture` config value
  short-circuits network access and emits one synthetic record stamped with a `fixture: true`
  field (`vitally.go:64-65,107-112`) — this is a legacy testing affordance, not part of the real
  API's record shape, and parity is asserted against legacy's LIVE read path via `httptest`
  (SPEC.md §5.1's xkcd note documents the same principle for fixture-mode connectors generally).
