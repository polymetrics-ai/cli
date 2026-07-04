# Overview

OPUSWatch is a read-only declarative migration of `internal/connectors/opuswatch` (legacy Go
connector). It reads monitors, incidents, and checks from the OPUSWatch HTTP API. This bundle is
capability-parity with legacy; legacy stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide an OPUSWatch API key via the `api_key` secret; it is sent as the `X-API-Key` header
(`auth: [{"mode": "api_key_header", "header": "X-API-Key", "value": "{{ secrets.api_key }}"}]`) and
is never logged. There is no fallback unauthenticated mode — legacy hard-errors when `api_key` is
unset (`opuswatch connector requires secret api_key`), matching this bundle's `required: ["api_key"]`.

## Streams notes

All 3 streams (`monitors`, `incidents`, `checks`) share the identical shape: `GET` against the
OPUSWatch list endpoint, records at `data`, primary key `["id"]`, cursor field `updated_at`
(declared on the schema for catalog/candidate-cursor purposes; legacy's `Read` never applies a
server-side or client-side incremental filter, so no `incremental` block is declared on any stream
here — matching legacy's real behavior of always emitting every record on every sync run).

Pagination follows legacy's own `next_page`-in-body convention: the response body's `next_page`
field carries the literal value of the NEXT `page` query parameter to send (not an offset or a
separate opaque cursor token) — modeled as `pagination.type: cursor` with `token_path: next_page`
and `cursor_param: page`. Pagination stops when `next_page` is absent or empty, identical to
legacy's `strings.TrimSpace(next) == ""` check; no `stop_path` is declared since legacy has no
separate boolean stop signal beyond the token itself. `per_page` is sent on every request from the
`page_size` config value (default `100`, matching legacy's `defaultPageSize`) via each stream's
`query.per_page` object-form entry (`default: "100"`). Legacy's `page_size` bounds (1-500) are
narrowed to the engine's own generic string-config handling.

## Write actions & risks

None. Legacy `Write` always returns `connectors.ErrUnsupportedOperation`; `metadata.json` declares
`capabilities.write: false` and no `writes.json` file exists, matching legacy exactly.

## Known limits

- `page_size` config validation (legacy's numeric range) is not reproduced at the bundle-config
  level; the engine treats `page_size` as an opaque string substituted directly into the `per_page`
  query param. A caller-supplied malformed value (e.g. `page_size: "abc"`) is sent to OPUSWatch
  as-is rather than rejected client-side the way legacy's `strconv.Atoi` validation would. This
  never changes emitted record DATA for any legacy-valid input; it only narrows client-side input
  validation, which is out of scope for wave2 fan-out (Pass B).
- Legacy also accepts a runtime `max_pages` cap, but the declarative engine only supports fixed
  bundle-authored `pagination.max_pages` integers. This bundle intentionally does not declare an
  ignored `max_pages` `spec.json` property.
- No `incremental` block is declared on any stream: legacy itself performs no incremental
  filtering (every `Read` call re-fetches every page from the start), so this is exact parity, not
  a narrowing.
