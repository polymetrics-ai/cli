# Overview

Height is a wave2 fan-out declarative-HTTP migration. It reads Height tasks, lists, field
templates, users, and the workspace object through the Height REST API
(`GET https://api.height.app/...`). This bundle targets capability parity with
`internal/connectors/height` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Height API key via the `api_key` secret; it is sent as an `Authorization: api-key
<api_key>` header (`api_key_header` auth mode with `prefix: "api-key "`), matching legacy's
`connsdk.APIKeyHeader("Authorization", secret, "api-key ")` (`height.go:253`). It is never logged.
`base_url` defaults to `https://api.height.app` and may be overridden for tests/proxies.

## Streams notes

`lists`, `field_templates`, `users`, and `workspace` are simple, non-paginated reads. Records for
`lists`/`field_templates`/`users` live under the `list` key of a `{"list":[...]}` envelope;
`workspace` returns a single object at the root, expressed here as `records.path: ""` (the engine's
root-object convention — a single JSON object at that path becomes a one-element record set,
matching legacy's `recordsPath: ""` + `connsdk.RecordsAt` behavior exactly).

`tasks` is Height's one paginated endpoint: legacy drives a bounded loop reading `nextPageToken`
from the response body and supplying it back as the `after` query param, only when
`usePagination=true` is also sent (`height.go:146-152`). This is exactly the engine's `cursor` +
`token_path` pagination shape (`pagination.token_path: "nextPageToken"`, `cursor_param: "after"`);
`usePagination=true` is declared as a static per-stream `query` value on the `tasks` stream only
(other streams never send it, matching legacy's `endpoint.paginated` gate). A `nextPageToken` value
of JSON `null` or an absent key both stringify to `""` via `connsdk.StringAt`, so pagination stops
identically to legacy's `next == "" || next == "null"` check; no `stop_path` is declared since
Height has no separate boolean "has more" signal (legacy's own extra `len(records) == 0` stop
condition is subsumed by the engine's own "0 records this page" empty-page defensive stop, applied
uniformly across cursor paginators).

None of the 5 streams expose a server-side incremental filter parameter in the legacy connector
(Height's `createdAt` is declared as a stream-level cursor field for state-tracking purposes only —
`heightStreams()`'s `CursorFields`, `height/streams.go:41-70` — but is never sent as a request
filter anywhere in `harvest`). This bundle therefore declares `x-cursor-field: createdAt` on every
schema (parity with legacy's advisory cursor fields, so `incremental_append`/`_deduped` sync modes
are available where a primary key + cursor field pair exists) but no `incremental` block on any
stream — full refresh only, matching legacy exactly.

## Write actions & risks

None. Height has no obviously-safe reverse-ETL writes in the legacy connector (`Capabilities:
Write: false`); this bundle ships no `writes.json`, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`.

## Known limits

- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`, a credential-free conformance-harness affordance) stamps
  deterministic synthetic values across every field regardless of stream shape and appends
  `previous_cursor` when `req.State["cursor"]` happens to be set (`height.go:188-236`). None of
  these are part of the LIVE record shape; this bundle's schemas target the live path only. The
  engine's own conformance/fixture-replay harness supplies the credential-free test affordance this
  bundle needs, so no fixture-mode equivalent is needed here.
- **`max_pages` is not modeled as connector config.** Legacy exposes a `max_pages` config override
  (`heightMaxPages`, `height.go:286-304`) that is purely a client-side safety cap (default 1000) —
  it never changes which records are emitted for any real Height dataset under 1000 pages, only how
  far a pathological/misbehaving API's loop is allowed to run. The engine's `cursor`+`token_path`
  paginator has no config-driven `MaxPages` knob wired to a spec property (unlike, say, stripe's
  static `page_size`/`max_pages` spec entries, which are also declared-but-unwireable per that
  bundle's own ledger note), so this bundle does not declare `max_pages` in `spec.json` at all (F6,
  REVIEW.md: a declared-but-unwireable config key is worse than an absent one). Pagination is
  bounded only by the empty/absent-token stop signal, matching Height's own real termination
  behavior for any well-behaved sync.
