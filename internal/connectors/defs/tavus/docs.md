# Overview

Tavus is a wave2 fan-out declarative-HTTP migration of `internal/connectors/tavus` (the
hand-written legacy connector this bundle migrates; the legacy package stays registered and
unchanged until wave6's registry flip). It reads Tavus replicas through the Tavus REST API
(`GET https://tavusapi.com/v2/replicas`). Read-only.

## Auth setup

Provide a Tavus API key via the `api_key` secret; it is sent as the `x-api-key` header
(`api_key_header` auth mode, no prefix), matching legacy's
`connsdk.APIKeyHeader("x-api-key", key, "")` (`tavus.go:125`). Never logged. `base_url` defaults
to `https://tavusapi.com/v2` and may be overridden for tests/proxies.

## Streams notes

`replicas` is the only stream: `GET /replicas`, records at `data`, primary key `["id"]`. Tavus's
real wire shape names the id/name fields `replica_id`/`replica_name`; `computed_fields` renames
both (`"id": "{{ record.replica_id }}"`, `"name": "{{ record.replica_name }}"`) to match legacy's
effective output field names exactly. `created_at` is copied as-is (no rename needed).

Pagination follows a 1-based page-number convention (`pagination.type: page_number`,
`page_param: page`, `size_param: page_size`, `page_size: 100`) — matches legacy's hand-rolled loop
(`tavus.go:88-109`: `page`/`page_size` query params, stopping when a page returns fewer records
than the configured size).

## Write actions & risks

None. Legacy `tavus` is read-only (`Write` returns `connectors.ErrUnsupportedOperation`);
`metadata.json` declares `capabilities.write: false` and this bundle ships no `writes.json`.

## Known limits

- Full Tavus API surface (videos, conversations, personas) is out of scope for wave2; see
  `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}`
  entries. Only the single legacy-parity `replicas` stream is implemented.
- **`page_size` is not runtime-configurable.** Legacy exposes a config-driven `page_size` override
  (`tavus.go:145-151`, `pageSize(cfg, 100)`, any positive integer, defaulting to 100 when unset or
  invalid). The engine's `page_number` paginator constructor reads `PaginationSpec.PageSize` as a
  static bundle-level integer from `streams.json`, not a config-templated field, so there is no
  mechanism to make it runtime-configurable from `config.page_size` without inventing Go. This
  bundle hardcodes `page_size: 100`, legacy's own default, matching every input that does not
  explicitly override the page size (the common case); an operator who previously set a
  smaller/larger `page_size` config value loses that override here. `page_size` is not declared in
  `spec.json` at all (F6, REVIEW.md: a declared-but-unwireable key is worse than an absent one).
- All fixtures (`fixtures/streams/replicas/**`, `fixtures/check.json`) represent Tavus's real wire
  shape, including the `replica_id`/`replica_name` field names before the `computed_fields` rename.
  `fixtures/streams/replicas/page_1.json` carries exactly 100 synthetic records (matching the
  static `page_size: 100`) so the real page-number paginator's short-page-stop rule genuinely
  exercises a second request, proving `pagination_terminates` rather than short-circuiting after
  page 1.
