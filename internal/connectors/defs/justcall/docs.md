# Overview

JustCall reads users (agents), call logs, and SMS/texts through the JustCall REST API
(`https://api.justcall.io/v2.1`). This bundle migrates 3 of legacy `internal/connectors/justcall`'s
5 streams to the declarative engine at capability parity; `contacts` and `phone_numbers` remain on
the legacy Go implementation for now (see Known limits — a validator/tooling coverage-bookkeeping
conflict, not a read-behavior gap). The legacy package stays registered and unchanged until wave6's
registry flip regardless. The connector is read-only.

## Auth setup

Provide a JustCall API key pair (`api_key:api_secret`) via the `api_key_2` secret; it is sent
verbatim as the `Authorization` header with no `Bearer` prefix (`auth: [{mode: api_key_header,
header: Authorization, value: "{{ secrets.api_key_2 }}"}]`), matching legacy's
`connsdk.APIKeyHeader("Authorization", secret, "")` call exactly. Never logged.

## Streams notes

Three streams are ported at Tier 1, matching legacy's `justcallStreamEndpoints` entries for the
same names exactly:

- `users` (`GET /v2.1/users`) — page-increment pagination (`page` starts at 0, `per_page` size
  param), full refresh, primary key `id`.
- `calls` (`GET /v2.1/calls`) — same pagination shape; incremental cursor field `call_date`. The
  `from_datetime` query param is sent only when the incremental lower bound resolves (state cursor,
  falling back to the `start_date` config value on a fresh sync) via the optional-query dialect:
  `"from_datetime": {"template": "{{ incremental.lower_bound }}", "omit_when_absent": true}` — a
  full sync with no `start_date` configured omits the param entirely, matching legacy's
  `incrementalLowerBound` returning `""` in that case.
- `sms` (`GET /v2.1/texts`) — identical shape to `calls`, incremental cursor field `sms_date`.

## Write actions & risks

None. JustCall is read-only in both legacy and this bundle (`capabilities.write: false`, no
`writes.json`).

## Known limits

- **`contacts` (`POST /v1/contacts/list`) and `phone_numbers` (`POST /v1/numbers/list`) are NOT
  ported in this wave — kept on the legacy Go implementation.** Both are read-only legacy streams
  (list-all endpoints that happen to use POST with an empty JSON body, per JustCall's own API
  design) and are fully expressible in the engine's declarative dialect (`method: "POST"` on a
  `StreamSpec`, `records.path: "data"`, `page_number` pagination) — the read behavior itself is not
  the blocker. The blocker is `cmd/connectorgen validate`'s `api_surface_fail_first_run` rule
  (`cmd/connectorgen/validate.go`'s `checkAPISurface`, rule `surface_fail_first_run`): it treats ANY
  non-`excluded` POST/PUT/PATCH/DELETE endpoint in `api_surface.json` as implying
  `capabilities.write` must be `true`, unconditionally — regardless of whether the endpoint's
  `covered_by` points at a `stream` (read) or a `write` action. There is no vocabulary in the
  closed `excluded.category` enum (`destructive_admin`, `requires_elevated_scope`, `binary_payload`,
  `deprecated`, `non_data_endpoint`, `duplicate_of`, `out_of_scope`) that accurately describes "a
  POST-method endpoint that is actually a read, already implemented as a stream" — every option
  either falsely implies non-implementation (`out_of_scope`) or doesn't fit semantically. Declaring
  `capabilities.write: true` to satisfy the rule was rejected as a worse alternative: it is a false
  capability claim (this connector has zero write actions and no `writes.json`) that would let
  `internal/app/app.go`'s reverse-ETL path (gated on `Capabilities.Write`) accept JustCall as a
  reverse-ETL destination and fail unpredictably at write time instead of being correctly refused
  up front. **This is a validator/dialect tooling gap, not an engine read-path gap** — a follow-up
  increment extending the `covered_by` vocabulary (e.g. a `read_via_post` marker exempted from the
  mutation-method check) would let both streams port cleanly at Tier 1 with no behavior change from
  what's already proven to work here.
- **`max_pages` is not configurable**: legacy's default (unbounded: `0`/`all`/`unlimited`) matches
  the engine's own default (`PaginationSpec.MaxPages <= 0` = unbounded) exactly, so this bundle
  declares no `max_pages` spec property or pagination field.
- Full JustCall API surface (SMS send, call actions, webhooks, other v1/v2 resources) is out of
  scope; see `api_surface.json` — only `users`/`calls`/`sms` are implemented at Tier 1 in this wave.
- `users`/`calls`/`sms` fixtures each ship a 100-record page 1 (matching the fixed
  `pagination.page_size: 100` short-page threshold) plus a 1-record page 2.
