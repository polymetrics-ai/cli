# Overview

Postmark App is a **partial** read-only declarative-HTTP bundle migrated from
`internal/connectors/postmarkapp` (the hand-written legacy connector, which stays registered and
unchanged until wave6's registry flip). It reads Postmark outbound and inbound message activity
through the Postmark HTTP API. Legacy's third stream, `servers`, is NOT migrated in this bundle —
see Known limits — and legacy stays authoritative for that stream until a follow-up wave closes
the underlying engine gap.

## Auth setup

Provide a Postmark **server token** via the `X-Postmark-Server-Token` secret. It is sent as the
`X-Postmark-Server-Token` request header (`api_key_header` auth mode), matching legacy's
`connsdk.APIKeyHeader("X-Postmark-Server-Token", token, "")` construction exactly for both migrated
streams (`outbound_messages`, `inbound_messages` — legacy's `streamSpecs` table routes both to
`tokenName: "X-Postmark-Server-Token"`).

Legacy's `servers` stream instead requires a Postmark **account token**
(`X-Postmark-Account-Token`), a *different* header name selected by stream — this bundle does not
declare or wire that secret at all, since nothing in this bundle's `streams.json` would consume it
(a declared-but-unwireable secret is worse than an absent one, per conventions.md F6's "dead
config" reasoning applied to secrets). See Known limits for why.

## Streams notes

Both migrated streams (`outbound_messages`, `inbound_messages`) share the same shape: `GET`
against a Postmark message-search endpoint, `offset_limit` pagination (`limit_param: count`,
`offset_param: offset`, `page_size: 100` matching legacy's `defaultPageSize` constant), and a
`computed_fields` rename from Postmark's PascalCase wire fields (`MessageID`, `Subject`, `From`,
`To`, `Status`, `ReceivedAt`) to this bundle's snake_case schema properties — plain schema
projection only matches by exact key name, so every field needed an explicit rename (unlike a
bundle whose raw API already returns snake_case/lowercase keys).

- `outbound_messages` (`GET /messages/outbound`, records at `Messages`): keeps `id`, `subject`,
  `from`, `to` (the raw `To` array of `{Email,Name}` objects, copied through as-is via a bare
  `{{ record.To }}` reference — typed extraction preserves the array, no stringification), `status`,
  and `received_at` (`x-cursor-field`, matching legacy's `CursorFields: ["received_at"]`).
- `inbound_messages` (`GET /messages/inbound`, records at `InboundMessages`): keeps `id`,
  `subject`, `from`, `to` (here the raw `To` field is a STRING, not an array — Postmark's inbound
  response shape differs from outbound's; the schema types `to` accordingly), and `status`. Legacy
  also declares `CursorFields: ["received_at"]` on this stream, but the real Postmark inbound
  message response has no `ReceivedAt` field at all (confirmed against Postmark's own API
  reference) — legacy's shared `mapRecord` function reads `item["ReceivedAt"]` unconditionally and
  gets `nil` for every real inbound record, so this bundle omits `received_at` entirely from the
  `inbound_messages` schema (and therefore has no `x-cursor-field` on that stream) rather than
  declaring a property that would always be null in practice; this is schema-as-projection applied
  honestly, not a narrowing of real legacy-emitted data.

No `incremental` block is declared on either stream: legacy's `Read` never sends a server-side
date-range filter param for either endpoint (no `fromdate`/`todate` request param appears anywhere
in `postmarkapp.go`), so declaring one here would be new, legacy-diverging behavior.

## Write actions & risks

None. Postmark App is a read-only source in both legacy and this bundle (`capabilities.write:
false`, no `writes.json`).

## Known limits

- **`servers` stream is NOT migrated (ENGINE_GAP, partial status).** Legacy's `servers` stream
  (`GET /servers`, records at `Servers`) authenticates with a *different* header
  (`X-Postmark-Account-Token`) than the two migrated message streams (`X-Postmark-Server-Token`).
  The engine resolves `base.auth`/`base.headers` exactly ONCE per bundle, in `newRuntime`
  (`internal/connectors/engine/read.go`), with no per-stream override mechanism at all —
  `StreamSpec` has no `Auth`/`Headers` field, so a single declarative bundle structurally cannot
  select a different credential header depending on which stream is being read. Sending BOTH
  tokens as two simultaneous static headers on every request was considered and rejected: legacy
  never sends the account token when reading messages (or vice versa), and assuming Postmark's live
  API tolerates an extra, stream-irrelevant auth header without verifying it against the real API
  would be an unverified accepted-input-behavior change, which the parity meta-rule forbids. This
  is reported as a typed `ENGINE_GAP` blocker; `servers` keeps its current legacy implementation
  until a follow-up wave adds per-stream auth/header override to the engine dialect (or a Tier-2
  hook, out of scope for this wave's JSON-only mandate).
- `page_size`/`max_pages` are not exposed as config properties: `PaginationSpec.PageSize`/
  `MaxPages` are static ints in `streams.json`'s `pagination` block with no runtime config-driven
  override mechanism (same F6 lesson as searxng) — `page_size` is fixed at `100` (matching legacy's
  `defaultPageSize`) and `max_pages` is unbounded (matching legacy's own default of `0`/unlimited
  when unset).
- Only 2 of 3 legacy streams are implemented (`servers` excluded per above); the full Postmark
  surface (message details, delivery stats, bounces, templates, sending, suppressions) is out of
  scope for this wave regardless — see `api_surface.json`'s `excluded` entries.
