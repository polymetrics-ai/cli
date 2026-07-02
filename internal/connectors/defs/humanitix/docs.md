# Overview

Humanitix is a wave2 fan-out declarative-HTTP migration. It reads Humanitix events, tags, orders,
and tickets through the Humanitix public REST API (`GET https://api.humanitix.com/v1/...`). This
bundle is a capability-parity port of the hand-written connector at
`internal/connectors/humanitix` (`humanitix.go`/`streams.go`), which stays registered and
unchanged until wave6's registry flip.

## Auth setup

Provide a Humanitix API key via the `api_key` secret; it is sent as the `x-api-key` header
(`streams.json`'s `base.auth`, `mode: api_key_header`), matching legacy's
`connsdk.APIKeyHeader("x-api-key", secret, "")` (`humanitix.go:242`). `base_url` defaults to
`https://api.humanitix.com/v1` and may be overridden for tests/proxies (legacy's own
`humanitixBaseURL` validates scheme+host the same way; the engine's base-URL resolution has no
equivalent runtime validation, but every parity/conformance fixture only ever points at an
httptest server, so this is not exercised differently on either side).

## Streams notes

`events` and `tags` are account-scoped list endpoints (`GET /events`, `GET /tags`); records live
at the top-level key matching the stream name. Both use `page_number` pagination
(`page`/`pageSize` query params, 1-based `start_page`, `page_size: 100`), matching legacy's
`connsdk.PageNumberPaginator{PageParam:"page", SizeParam:"pageSize", StartPage:1, PageSize:
pageSize}` (`humanitix.go:156-161`) and its short-page stop rule.

`orders` and `tickets` are event-scoped sub-resources: the path templates
`/events/{{ config.event_id }}/orders` and `/events/{{ config.event_id }}/tickets` substitute the
required `event_id` config value (urlencoded by `InterpolatePath`'s per-segment default, matching
legacy's own `url.PathEscape(eventID)` in `Read`, `humanitix.go:146`); an absent `event_id`
hard-errors on both sides (legacy: `"humanitix stream %q requires config event_id"`; engine: an
unresolved `config.event_id` path-template key — same failure classification, different literal
text, per conventions.md §5's precedent for config-validation parity).

`events` carries legacy's `since` incremental filter (`humanitixEventFields`'s cursor field
`updatedAt`, legacy's `incrementalLowerBound` helper at `humanitix.go:250-255`): the stream
declares `incremental.cursor_field: updatedAt`, `request_param: since`, `start_config_key: since`,
and the `since` query param is wired through the opt-in optional-query dialect
(`{{ incremental.lower_bound }}` with `omit_when_absent: true`) so it is sent only once a lower
bound resolves (a state cursor from a prior sync, or the `since` config on a first run) — omitted
entirely on a from-scratch full sync, exactly matching legacy's `if since := ...; since != ""`
gate. `tags`, `orders`, and `tickets` expose no incremental filter in the Humanitix API (legacy
never sends `since` for them), so no `incremental` block is declared for those three streams.

## Write actions & risks

None. The Humanitix public API exposes no safe reverse-ETL writes (legacy's own package doc: "no
safe reverse-ETL writes, so Capabilities.Write is false"); `capabilities.write` is `false` and this
bundle ships no `writes.json`, matching legacy's `Write` returning
`connectors.ErrUnsupportedOperation`.

## Known limits

- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`, a credential-free conformance-harness affordance) stamps an
  extra `previous_cursor` field (echoing a prior cursor when set) onto fixture-mode records
  (`humanitix.go:216-219`). This bundle's schemas and fixtures target the live record shape only;
  the engine's own conformance/fixture-replay harness (`internal/connectors/conformance`) provides
  the credential-free test affordance legacy's fixture mode existed for, so no fixture-mode
  equivalent is needed here.
- **`max_pages` is not modeled as a bundle-level config knob.** Legacy exposes `max_pages` as a
  config override (`humanitixMaxPages`, `humanitix.go:300-313`, accepting an integer, `all`, or
  `unlimited`). The engine's `page_number` paginator has no config-driven `max_pages` override
  wired to a spec property — pagination is bounded only by the short-page stop signal, matching
  Humanitix's own real termination behavior (the same "unbounded by default" outcome as legacy's
  `max_pages` unset/`0`/`all`/`unlimited` case). A future capability-expansion pass could wire a
  `spec.json` `max_pages` property into `streams.json`'s pagination block once/if the engine grows
  a config-driven `MaxPages` template reference; not modeled here to avoid declaring dead config
  (F6, REVIEW.md).
