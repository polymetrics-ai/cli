# Overview

SolarWinds Service Desk is a wave2 fan-out declarative-HTTP migration. It reads SolarWinds Service
Desk incidents, users, departments, categories, problems, and changes through the SolarWinds
Service Desk (Samanage) REST API (`GET https://api.samanage.com/<resource>.json`). This bundle is
capability-parity migrated from `internal/connectors/solarwinds-service-desk` (the hand-written
connector it migrates); the legacy package stays registered and unchanged until wave6's registry
flip.

## Auth setup

Provide a SolarWinds Service Desk API key via either the `api_key_2` secret or the `api_key`
secret; both are sent as a Bearer token. `base.auth` declares two `when`-gated bearer candidates in
declared order (conventions.md §3's dual-auth-ordering pattern): `api_key_2` first (gated `when:
{{ secrets.api_key_2 }}`), `api_key` second (gated `when: {{ secrets.api_key }}`) — reproducing
legacy's own `firstSecret(cfg, "api_key_2", "api_key")` precedence exactly
(`solarwinds_service_desk.go:141`: `api_key_2` wins when both are configured). Every request also
sends a fixed `Accept: application/vnd.samanage.v1.1+json` header, matching legacy's
`acceptHeader` constant (`solarwinds_service_desk.go:18`, `connsdk.Requester.Accept`). Neither
secret is ever logged.

## Streams notes

All 6 streams (`incidents`, `users`, `departments`, `categories`, `problems`, `changes`) share the
same shape: `GET /<resource>.json`, records at the response body's ROOT array (`records.path: ""`,
matching legacy's `streamEndpoints[...].recordsPath == ""`), and `pagination.type: "none"` — legacy
itself performs **no automatic pagination loop** (`readRecords`, `solarwinds_service_desk.go:104`,
issues exactly one request and returns; `page`/`per_page` are pass-through query params a caller
may set manually, never auto-incremented), so a single unconditional request per `Read` call is
exact parity, not a narrowing.

`incidents` additionally forwards `config.start_date` as the `updated_after` query param via the
opt-in optional-query dialect (`omit_when_absent: true`) — present only when `start_date` is
configured, omitted entirely otherwise, matching legacy's own `copyConfig(q, cfg, "start_date",
"updated_after")` (`solarwinds_service_desk.go:150`, only wired for the `incidents` stream). All 6
streams forward `config.page`/`config.per_page` verbatim as `page`/`per_page` query params
(likewise `omit_when_absent`), matching legacy's own unconditional `copyConfig(q, cfg, "page",
"page")`/`copyConfig(q, cfg, "per_page", "per_page")` (applied to every stream, unlike
`updated_after`). No stream declares an `incremental` block: legacy's own catalog declares no
`CursorFields` for any of the 6 streams, and `updated_after` is a raw pass-through param, not an
engine-computed incremental lower bound — matching that, no schema declares `x-cursor-field`.

All 6 streams declare `"projection": "passthrough"`. Legacy's `Read` emits the raw API record
verbatim (`emit(connectors.Record(rec))`, `solarwinds_service_desk.go:117`, inside `readRecords`)
with no field-building/filtering step — `streams()`'s four-field `Fields` list
(`solarwinds_service_desk.go:99`) is consumed only by `Catalog`, never by `Read`. Every real
Samanage field beyond each schema's narrow `id`/`name`/`created_at`/`updated_at` properties (e.g.
`priority`, `state`, `requester`, `assignee`, custom fields) survives to the emitted record exactly
as legacy would emit it. Declaring the default `"schema"` projection mode here would silently
narrow every emitted record to the schema's declared properties — an undocumented parity
deviation from legacy's verbatim passthrough — so `passthrough` is required, matching
conventions.md §8 rule 1 (legacy's raw `emit(record)` with no `mapRecord` field-building is the
mechanical signal to use `passthrough`).

## Write actions & risks

None. SolarWinds Service Desk is read-only (`capabilities.write: false`, no `writes.json`),
matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **No automatic pagination is a legacy limitation this bundle reproduces exactly, not a bundle
  narrowing.** Legacy's own `Read` never loops or reads a next-page signal; a result set larger
  than one API page is genuinely truncated by legacy today. This bundle's `pagination.type: "none"`
  is exact behavioral parity, not a workaround.
- **`page`/`per_page`/`start_date` are forwarded with no local validation**, matching legacy's own
  `copyConfig` (a plain pass-through with no parsing/bounds-check). An invalid value now surfaces as
  the API's own error response instead of a local validation error — same eventual behavior legacy
  itself exhibits (legacy does not validate these either).
- Legacy's own hard-error message when neither secret is configured names only `api_key_2`
  (`"solarwinds-service-desk connector requires secret api_key_2"`,
  `solarwinds_service_desk.go:143`) even though either secret satisfies the requirement — a
  pre-existing legacy wording quirk. This bundle's engine-native "no auth spec matched" error text
  differs but the same input (neither secret set) fails identically; this is a parity-neutral
  error-message difference, not an accepted-input-behavior change (§5 meta-rule).
