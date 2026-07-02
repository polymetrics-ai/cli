# Overview

Apptivo is a wave2 fan-out declarative-HTTP migration. It reads Apptivo CRM customers, contacts,
leads, and opportunities through the read-only Apptivo REST DAO API
(`GET https://app.apptivo.com/app/dao/v6/<object>`). This bundle targets capability parity with
`internal/connectors/apptivo` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Apptivo authenticates every request with two credentials sent as query parameters: `apiKey` and
`accessKey`. Provide both as the `api_key`/`access_key` secrets; `api_key` is wired via the
`api_key_query` auth mode (`apiKey={{ secrets.api_key }}`) and `access_key` is wired via a static
per-stream `query` entry (`accessKey={{ secrets.access_key }}`) — the engine's `stream.Query`
templates resolve against `secrets.*` exactly like `auth` does, so both credentials reach every
request the same way legacy's `apptivoAuth` closure (`apptivo.go:241`) sets both query params on
every call. Neither value is ever logged. `base_url` defaults to `https://app.apptivo.com` and may
be overridden for tests/proxies.

## Streams notes

All 4 streams (`customers`, `contacts`, `leads`, `opportunities`) share the same shape: `GET`
against the Apptivo DAO `getAll` action (`query: {"a": "getAll"}`), records at the `data` body key.
Pagination follows Apptivo's offset convention (`pagination.type: offset_limit`,
`limit_param: numRecords`, `offset_param: startIndex`, `page_size: 100`) — a page shorter than 100
records terminates the loop, matching legacy's `harvest` (`apptivo.go:139`) exactly.

## Write actions & risks

None. Legacy's `Write` unconditionally returns `connectors.ErrUnsupportedOperation`;
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`startIndex=0` is sent explicitly on the first page request; legacy omits it entirely.**
  Legacy's `harvest` (`apptivo.go:148`) only sets `startIndex` in the query when `offset > 0`,
  mirroring the "upstream inject_on_first_request:false behaviour" its own comment describes — the
  API is documented to treat a missing `startIndex` as `0`. The engine's `offset_limit` paginator
  (`connsdk.OffsetPaginator.Start()`) always sends `offset_param` (here `startIndex=0`) on the
  first request, with no "omit on first page" option in the dialect. Per legacy's own comment this
  is accepted-behaviorally-identical by the API (a missing `startIndex` and an explicit `0` are the
  same request from Apptivo's point of view), so this never changes emitted record data for any
  input legacy itself would accept — an ACCEPTABLE parity deviation, not a data-changing one.
- `page_size`/`max_pages` config overrides legacy exposes (`apptivoPageSize`/`apptivoMaxPages`,
  clamped 1-500 / `all`/`unlimited`) are not runtime-configurable here: the engine's
  `offset_limit` paginator's `PageSize` is a static int set once in `streams.json`, not
  template-resolvable, and no config knob feeds a `MaxPages` cap for this paginator type.
  `spec.json` intentionally does not declare `page_size`/`max_pages` (a declared-but-unwireable key
  is worse than an absent one, per conventions.md F6).
