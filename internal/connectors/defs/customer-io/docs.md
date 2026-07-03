# Overview

Customer.io is a Tier-1 declarative-HTTP source bundle (`internal/connectors/defs/customer-io/`):
Bearer auth, plain JSON list endpoints under the Customer.io App (Beta) API, and cursor pagination
where each page's response carries a `next` token echoed back as the `start` query parameter (an
empty/null `next` ends the loop). It reads campaigns, newsletters, segments, and broadcasts. This
port is a pure `streams.json`/`spec.json`/`schemas/*.json` bundle with zero Go — the legacy package
(`internal/connectors/customer-io/`) is a thin connsdk composition with no auth/stream hooks, so it
maps directly onto the engine's `cursor` + `token_path` pagination dialect (the identical shape
airtable's `records` stream uses for its `offset`/`offset` pair).

## Auth setup

Provide `app_api_key` (an App API Key from Customer.io's UI) as a secret; it is sent as
`Authorization: Bearer <app_api_key>` via `streams.json`'s `base.auth`, never logged.

`base_url` is **required** in this bundle, unlike legacy where it was derived from an optional
`region` config value (`US` -> `https://api.customer.io/v1`, `EU` -> `https://api-eu.customer.io/v1`,
with an explicit `base_url` override always taking priority). See Known limits below for why the
derivation itself could not be ported.

## Streams notes

All 4 streams (`campaigns`, `newsletters`, `segments`, `broadcasts`) share the identical shape:
`GET` against the resource path, `records.path` set to the resource's own top-level JSON key (e.g.
`{"campaigns": [...]}`), a `limit` query param sourced from `config.page_size` (default `100`,
matching legacy's `customerIODefaultPage`/`customerIOMaxPage` clamp — the engine dialect does not
enforce a 1-100 range on a config value the way legacy's `customerIOPageSize` did; see Known limits),
and `incremental.cursor_field: updated` with `client_filtered: true` (the Customer.io App API has no
server-side `updated`-since filter parameter, matching legacy's `harvest`, which fetches every page
unconditionally and relies on the caller's own downstream dedup/append semantics — `client_filtered`
is the sanctioned dialect for exactly this "API can't filter server-side" shape, per
`docs/migration/conventions.md` §3).

Every object exposes a numeric `id` and Unix-seconds `created`/`updated` timestamps (segments omit
`created`, matching legacy's `segmentFields`), so every schema declares `x-primary-key: ["id"]` and
`x-cursor-field: "updated"`.

## Write actions & risks

None. Customer.io is read-only here (`capabilities.write: false`), matching legacy's
`Connector.Write` (`connectors.ErrUnsupportedOperation`) — the legacy package never implemented any
reverse-ETL action (Customer.io does support triggering broadcasts/sending transactional messages via
its API, but the ported connector never called those endpoints; see `api_surface.json`'s excluded
`out_of_scope` entries for Pass B).

## Known limits

- **`base_url` cannot be derived from a `region` config value.** Legacy's `customerIOBaseURL`
  switches on an optional `region` config key (`US`/`EU`/unset-defaults-to-US) to choose between two
  hardcoded base URLs, only falling back to a directly-configured `base_url` override. The engine's
  `spec.json` `"default"` mechanism materializes a single **fixed literal** default value for an
  absent key (see `docs/migration/conventions.md` §3, "`spec.json` `default` values ARE now
  materialized") — it has no mechanism to derive one config value's default from ANOTHER config
  value's value (the same gap `sentry`'s hostname-derived URL and `chargebee`'s site-derived URL
  hit). Per the sanctioned resolution for this exact shape (conventions §3), `base_url` is declared
  **required** here instead of re-deriving the branch in Go (which would need a 3rd Tier-2 hook
  interface or a Tier-3 escalation neither justified by this connector's otherwise-uniform HTTP
  shape). This is a documented, accepted narrowing of the config surface, never a change to any
  emitted record's data: an operator who previously left `region` unset (or set it to `US`/`EU`) now
  supplies the resolved `https://api.customer.io/v1` or `https://api-eu.customer.io/v1` value
  directly as `base_url`.
- **`page_size`'s 1-100 range is not enforced.** Legacy's `customerIOPageSize` rejects a `page_size`
  outside `[1, 100]` with a config-validation error before the first request. The engine dialect has
  no range-validation primitive for a plain templated query parameter — an out-of-range `page_size`
  here is sent to the Customer.io API as-is and would surface as a live API error rather than a local
  config-validation error. This never changes emitted record DATA for any `page_size` legacy itself
  would have accepted (1-100); it only moves where an out-of-range value is rejected, from local
  config validation to the live API's own response.
- **`max_pages` is not configurable.** Legacy accepts a `max_pages` config value (default unlimited)
  as a client-side page-count cap. The engine's `PaginationSpec.MaxPages` is a static bundle-declared
  integer, not a per-request templated value (mirroring stripe's identical, already-accepted
  `max_pages`/`page_size` dead-config resolution recorded in the parity-deviation ledger, §5 item 3)
  — there is no config-driven override mechanism for it at all. This bundle declares no `max_pages`
  spec property (a declared-but-unwireable key is worse than an absent one, per `conventions.md` F6)
  and leaves pagination unbounded (matching legacy's default `max_pages=0`/unlimited behavior, and the
  paginator's own short-page/empty-token stop signal still terminates every real sync).
