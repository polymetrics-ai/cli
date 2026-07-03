# Overview

Beamer is a read-only feedback/changelog source connector. It reads NPS survey responses,
announcement posts, feature requests, and comments through the Beamer REST API
(`https://api.getbeamer.com/v0`). This bundle migrates `internal/connectors/beamer` (the
hand-written legacy connector); the legacy package stays registered and unchanged until wave6's
registry flip.

## Auth setup

Provide a Beamer API key via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged, matching legacy's
`connsdk.Bearer(secret)` wiring exactly.

## Streams notes

All 4 streams (`nps`, `posts`, `feature_requests`, `comments`) share the same shape: `GET` against
a Beamer list endpoint that returns a bare JSON array at the response root (`records.path: "."`),
primary key `["id"]`. Pagination is genuinely 0-indexed (`pagination.type: page_number`,
`start_page: 0`, `page_param: page`, `size_param: maxResults`, `page_size: 100`) — the first
request sends `page=0`, matching legacy's `beamer_test.go:27-35` assertion and `harvest`'s
`for page := 0; ...` loop exactly; pagination stops on a short/empty page (fewer than `page_size`
records), identical to legacy's `len(records) < pageSize` check. Only `nps` is incremental: its
`date` field is filtered server-side via the `dateFrom` query param
(`incremental.request_param: dateFrom`, `param_format: rfc3339`, `start_config_key: start_date`),
matching legacy's `incrementalLowerBound` (persisted cursor, else `start_date` config) — Beamer's
API expects RFC3339 timestamps verbatim, so no reformatting is applied (legacy's own comment notes
the same: "no reformatting is required"). `posts`, `feature_requests`, and `comments` declare no
`incremental` block (full-refresh only) even though legacy exposes `date` as a `CursorFields`
candidate on every stream — only `nps` actually wires a request-time filter (`cursorParam`) in
legacy's per-stream routing table, so per §8 rule 2 the other three streams keep `x-cursor-field`
in their schemas only, with no `incremental` block.

`check` issues a single bounded `GET /nps?page=0&maxResults=1`, mirroring legacy's `Check`
implementation exactly (a 1-record probe of the `nps` endpoint confirms auth and connectivity
without mutating anything).

## Write actions & risks

None. Beamer is read-only (`capabilities.write: false`); legacy's own `Write` always returns
`connectors.ErrUnsupportedOperation` and there is no reverse-ETL write target for this API.

## Known limits

- Only the 4 legacy-parity read streams are implemented; see `api_surface.json`. Beamer's
  documented API surface (translations, help center, etc.) is out of scope until Pass B.
- `posts`/`feature_requests`/`comments` are full-refresh only (no server-side incremental filter
  wired in legacy), even though `date` is a schema-declared cursor candidate — matches legacy's own
  per-stream `cursorParam` routing table, where only `nps` sets a non-empty value.
