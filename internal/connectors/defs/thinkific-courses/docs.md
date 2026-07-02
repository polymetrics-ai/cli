# Overview

Thinkific Courses is a wave2 fan-out declarative-HTTP migration. It reads courses, chapters,
lessons, and enrollments from the Thinkific public API (`GET https://api.thinkific.com/api/public/v1/...`).
This bundle targets capability parity with `internal/connectors/thinkific-courses` (the hand-written
connector it migrates); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Thinkific API key via the `api_key` secret (sent as the `X-Auth-API-Key` header) and the
account subdomain via the required `subdomain` config key (sent as the `X-Auth-Subdomain` header),
matching legacy's `X-Auth-API-Key`/`X-Auth-Subdomain` header pair (`thinkific_courses.go:121-129`).
Neither is ever logged. `base_url` defaults to `https://api.thinkific.com/api/public/v1`
(legacy's `defaultBaseURL`).

## Streams notes

All four streams share the identical Thinkific envelope: `GET /<resource>` returns
`{"items":[...]}`; records live at the top-level `items` key for every stream (legacy's
`recordsPath: "items"`, uniform across all four `streamSpecs`). Pagination is `page_number`
(`page`/`limit`, `start_page: 1`, static `page_size: 100` matching legacy's `defaultPageSize`).
None of the four streams declare an `incremental` block: legacy's `Read` never applies a
cursor-based filter parameter of any kind (no `updated_at`-style query param is ever sent), so
every read is a full paginated sweep — this bundle matches that exactly rather than inventing an
incremental capability legacy never had.

## Write actions & risks

None. Thinkific Courses is read-only (`capabilities.write` is `false`); this bundle ships no
`writes.json`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`page_size`/`max_pages` config overrides are not modeled.** Legacy exposes `page_size`
  (default 100, must be a positive integer) and `max_pages` (0/absent = unbounded, or a
  non-negative integer cap) as config-driven overrides (`pageSize`/`maxPages` helpers,
  `thinkific_courses.go:208-230`). The engine's `page_number` paginator has no config-driven
  page-size or request-count-cap knob (mirrors the aha/adobe-commerce-magento precedent from this
  same wave); `page_size`/`max_pages` are therefore not declared in `spec.json`, and this bundle
  sends Thinkific's own default (`limit=100`) as a static pagination-block value with no page cap.
- **Legacy's `firstConfig` subdomain-key aliasing is narrowed to a single key name.** Legacy accepts
  any of `X-Auth-Subdomain`, `x_auth_subdomain`, or `subdomain` as the config key naming the account
  subdomain (`firstConfig(cfg, "X-Auth-Subdomain", "x_auth_subdomain", "subdomain")`,
  `thinkific_courses.go:117`). This bundle declares only `subdomain`; an operator previously using
  one of the other two key names must rename it. This is a config-surface naming narrowing only —
  the resolved effective subdomain value once configured is identical.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path stamps a
  `fixture: true` marker field with no live-path equivalent (`thinkific_courses.go:175`). This
  bundle's schemas and fixtures target the live record shape only; the engine's own
  `internal/connectors/conformance` fixture-replay harness provides the credential-free test
  affordance this bundle needs.
