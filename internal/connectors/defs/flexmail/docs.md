# Overview

Flexmail is a wave2 fan-out declarative-HTTP migration. It reads Flexmail contacts, custom fields,
interests, segments, and sources through the Flexmail REST API
(`GET https://api.flexmail.eu/...`). This bundle migrates `internal/connectors/flexmail` (the
hand-written connector); the legacy package stays registered and unchanged until wave6's registry
flip.

## Auth setup

Provide the Flexmail `account_id` config value and a `personal_access_token` secret; both are sent
as HTTP Basic auth (username = `account_id`, password = `personal_access_token`), matching legacy's
`connsdk.Basic(accountID, secret)` (`flexmail.go:242`). The token is never logged. `base_url`
defaults to `https://api.flexmail.eu` and may be overridden for tests/proxies.

## Streams notes

All 5 streams read Flexmail's HAL-style collection envelope: records live at `_embedded.item`
(`records.path`), matching legacy's `recordsPath` for every endpoint. `contacts` and `sources` are
legacy's two paginated endpoints (`endpoint.paginated == true`); pagination is declared as
`pagination.type: offset_limit` with `limit_param: limit`/`offset_param: offset` and a
`page_size: 500` stop threshold — the same offset-advance-until-short-page behavior as legacy's
`harvest` loop. `custom_fields`, `interests`, and `segments` are legacy's non-paginated endpoints:
they return their full collection in a single response, so no `pagination` block is declared (a
stream with no pagination block defaults to `type: none`, a single request).

Every Flexmail object exposes an `id` primary key (`x-primary-key: ["id"]`); none of the streams
support incremental sync (Flexmail is full-refresh only, matching legacy's empty `CursorFields`),
so no schema declares `x-cursor-field` and no stream declares an `incremental` block.

`offset_limit`'s only stop signal is a short page (`recordCount < page_size`); it has no
Stripe-style `stop_path`/token equivalent. The required 2-page fixtures for `contacts`/`sources`
(conventions.md §4) therefore each ship a full `page_size`-sized (500-record) page 1 followed by a
1-record page 2 — a smaller page 1 would trigger the short-page stop after a single request and
never demonstrate real pagination, and lowering the bundle's *declared* `page_size` purely to make
a smaller fixture easier to author would change production request cadence (far more roundtrips
than legacy's real default), which this bundle avoids.

## Write actions & risks

None. Flexmail is read-only in legacy (`capabilities.write: false`, `Write` returns
`connectors.ErrUnsupportedOperation`); this bundle ships no `writes.json`.

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes `page_size` (1-500,
  default 500) and `max_pages` (default unbounded) as config-driven overrides on the two paginated
  endpoints (`flexmailPageSize`/`flexmailMaxPages`, `flexmail.go:275-303`). The engine's
  `offset_limit` paginator reads `PaginationSpec.PageSize` as a static bundle-level integer, not a
  `{{ config.* }}` template, and `PaginationSpec.MaxPages` is likewise a static int with no
  per-request config wiring mechanism — matching bitly's identical documented gap
  (`docs/migration/conventions.md`). This bundle declares `page_size: 500` (legacy's own default)
  directly in `streams.json`'s pagination blocks and leaves `max_pages` unset (unbounded, legacy's
  own default when unconfigured); `spec.json` still documents `page_size` for operator awareness
  even though it is not wired to any template (informational parity with legacy's config surface),
  but does not declare `max_pages` at all (F6, REVIEW.md: a declared-but-unwireable key is worse
  than an absent one).
- **Legacy's fixture-mode-only synthetic records are not modeled.** Legacy's `readFixture` path
  (only reached when `config.mode == "fixture"`) emits deterministic in-code records with fields
  not present on the live wire shape (`type`, `label`, `description`, etc. stamped onto every
  fixture regardless of stream). This bundle's fixtures instead follow the conventions.md fixture
  rules (recorded-real-shape per stream, sanitized), which the engine's own
  `internal/connectors/conformance` fixture-replay harness consumes for credential-free testing —
  no fixture-mode equivalent is needed in the bundle itself.
