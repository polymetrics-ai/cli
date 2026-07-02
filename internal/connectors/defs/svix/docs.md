# Overview

Svix is a wave2 fan-out declarative-HTTP migration of `internal/connectors/svix` (the
hand-written legacy connector this bundle migrates; the legacy package stays registered and
unchanged until wave6's registry flip). It reads Svix applications through the Svix REST API
(`GET https://api.svix.com/api/v1/app`). Read-only.

## Auth setup

Provide a Svix API key via the `api_key` secret; it is sent as a Bearer token (`Authorization:
Bearer <api_key>`) and is never logged, matching legacy's `connsdk.Bearer(secret)`
(`svix.go:115`). `base_url` defaults to `https://api.svix.com/api/v1` and may be overridden for
tests/proxies.

## Streams notes

`applications` is the only stream: `GET /app`, records at `data`, primary key `["id"]`. Svix's
real wire shape returns the creation timestamp as camelCase `createdAt` (legacy's own `first()`
helper prefers `created_at` and falls back to `createdAt`, but the live API only ever emits
`createdAt` — `created_at` never appears on the wire); a `computed_fields` rename
(`"created_at": "{{ record.createdAt }}"`) reproduces legacy's effective output field name exactly.

Pagination follows Svix's `iterator` cursor convention (`pagination.type: cursor`,
`cursor_param: iterator`, `token_path: iterator`): the next page's `iterator` query value is read
from the response body's own `iterator` field, and pagination stops when that field is empty —
identical to legacy's `connsdk.CursorPaginator{CursorParam: "iterator", TokenPath: "iterator"}`.
Every request sends `limit=50` (matches legacy's `defaultPageSize`) via the stream's static
`query: {"limit": "50"}` — NOT via a `{{ config.page_size }}` template, since `stream.query` is
plain unconditional interpolation with no absent-key-falsy tolerance and conformance's synthetic
per-property config generator would otherwise make the resolved `limit` value unpredictable
against any static fixture (see Known limits).

## Write actions & risks

None. Legacy `svix` is read-only (`Write` returns `connectors.ErrUnsupportedOperation`);
`metadata.json` declares `capabilities.write: false` and this bundle ships no `writes.json`.

## Known limits

- Full Svix API surface (endpoints, event types, messages, message attempts, background tasks) is
  out of scope for wave2; see `api_surface.json`'s `excluded: {category: out_of_scope, reason:
  "Pass B capability expansion"}` entries. Only the single legacy-parity `applications` stream is
  implemented.
- **`max_pages` is not modeled.** Legacy exposes a config-driven `max_pages` override (`0`/`all`/
  `unlimited` meaning unbounded, or a positive integer hard cap; `svix.go:195-205`). The engine's
  `PaginationSpec.MaxPages` is a static bundle-level integer, not a config-templated field, so
  there is no mechanism to make it runtime-configurable from `config.max_pages` without inventing
  Go. This bundle omits `max_pages` entirely, which is unbounded — legacy's own default when the
  config value is unset, `0`, `all`, or `unlimited` — so every input legacy itself defaults to
  (the common case) behaves identically; only an operator who explicitly set a positive
  `max_pages` override to cap requests loses that cap here. Documented scope narrowing, not silent
  divergence.
- **`page_size` is not runtime-configurable.** Legacy exposes a config-driven `page_size` override
  (1-250, default 50; `svix.go:183-193`). `stream.query`'s plain-string dialect has no
  absent-key-falsy tolerance, so a `{{ config.page_size }}` template would hard-error under any
  caller (including conformance) that never sets `page_size` explicitly, and would make the
  resolved `limit` value fixture-unmatchable under conformance's synthetic-per-property config
  generator regardless. This bundle hardcodes `limit: 50`, legacy's own default, matching every
  input that does not explicitly override the page size (the common case); an operator who
  previously set a smaller/larger `page_size` config value loses that override here. `page_size` is
  not declared in `spec.json` at all (F6, REVIEW.md: a declared-but-unwireable key is worse than an
  absent one).
- All fixtures (`fixtures/streams/applications/**`, `fixtures/check.json`) represent Svix's real
  wire shape, including the camelCase `createdAt` field and the `iterator` cursor token.
