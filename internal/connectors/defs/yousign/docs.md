# Overview

Yousign is an e-signature platform. This bundle reads Yousign signature requests, contacts, and
documents through the Yousign REST API v3 (`GET {base_url}/signature_requests|contacts|documents`).
It migrates `internal/connectors/yousign` (the hand-written connector); the legacy package stays
registered and unchanged until wave6's registry flip. Read-only: `capabilities.write` is `false`
and this bundle ships no `writes.json`.

## Auth setup

Provide `api_key` (secret) for Bearer auth (`Authorization: Bearer <api_key>`), matching legacy's
`connsdk.Bearer(token)`. Never logged.

## Streams notes

All 3 streams (`signature_requests`, `contacts`, `documents`) share the same shape: `GET` against
the Yousign list endpoint, records at `data`, primary key `["id"]`, cursor field `updated_at`. No
pagination is declared — legacy issues a single unpaginated request per stream and emits every
record in the response's `data` array, so this bundle's `streams.json` omits any `pagination` block
(defaulting to `none`), matching legacy exactly.

An optional `limit` config value is sent as the `limit` query parameter on every stream's read
request when set (`stream.Query`'s `omit_when_absent: true` opt-in dialect), matching legacy's
`baseQuery` (`yousign.go:176-182`). `updated_at` is a `computed_fields` rename from the raw
`created_at` field, matching legacy's `cursorKeys: {"created_at","updated_at"}` primary preference
(`created_at` is tried first).

## Write actions & risks

None. Yousign is modeled read-only in legacy (`capabilities.Write: false`); this bundle matches
that exactly and ships no `writes.json`.

## Known limits

- **Check does not send legacy's `limit=1` query parameter.** Legacy's `Check` sends a literal
  `limit=1` on its underlying `GET /signature_requests` probe (`yousign.go:63`, distinct from
  Read's config-driven `limit`), to keep the connectivity check cheap. The engine's declarative
  `check` block (`RequestSpec`) supports only `method`+`path`, no query parameters — this bundle's
  check therefore requests `/signature_requests` with no `limit` param at all. This does not change
  accepted-input behavior (the same endpoint, same auth, same 200/401/403 outcomes); it may return
  a marginally larger response body than legacy's probe, which is immaterial since `Check` only
  inspects the status code / error, never the body. Acceptable per the parity-deviation meta-rule
  (`docs/migration/conventions.md` §5): never changes emitted record data, only Check's own request
  shape.
- **`name` multi-key fallback is approximated by the primary key only.** Legacy's `contacts` and
  `documents` streams try `name` first, falling back to `email`/`filename` respectively only when
  `name` itself is absent from the raw record (`nameKeys: {"name","email"}` /
  `{"name","filename"}`). The engine's `computed_fields` dialect has no coalesce/fallback filter (a
  single template resolves a single dotted path only), so this bundle relies on `name`'s direct
  schema-projection passthrough (the common case) and does not model the `email`/`filename`
  fallback. A contact/document response that omits `name` entirely would silently emit a null
  `name` here where legacy would have recovered `email`/`filename`. Documented, not silently worked
  around.
- Full Yousign API surface (signature request creation/activation, document upload, workflow
  steps) is out of scope for wave2; see `api_surface.json`'s `excluded: {category: out_of_scope}`
  entries.
