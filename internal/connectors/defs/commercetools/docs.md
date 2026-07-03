# Overview

commercetools reads customers, orders, and products from the commercetools Composable Commerce
HTTP API. This bundle is a full capability-parity migration of the legacy hand-written connector
(`internal/connectors/commercetools`), which stays registered and unchanged until wave6's registry
flip. Read-only: legacy implements no write path (`Write` is a stub returning
`ErrUnsupportedOperation`).

## Auth setup

Provide `client_id`/`client_secret` secrets for an OAuth2 client-credentials grant
(`mode: oauth2_client_credentials`); the engine fetches and caches a Bearer token from `token_url`,
refreshing automatically before expiry, exactly matching legacy's
`connsdk.OAuth2ClientCredentials` usage. `base_url` and `token_url` must be the fully-formed API and
OAuth2 token endpoint URLs (see Known limits for why the region/host convenience shorthand is not
supported here). `project_key` scopes every stream path (`/{project_key}/customers` etc.).

## Streams notes

All 3 streams (`customers`, `orders`, `products`) share the same shape: `GET
/{project_key}/<resource>`, records at `results` (commercetools' list-endpoint envelope), `limit`/
`offset` query pagination (`pagination.type: offset_limit`), page size configurable via
`page_size` (default 100, matching legacy's `defaultPageSize`). Pagination stops on a short page
(fewer than `page_size` records), the same signal legacy's own `readOffset` checks first;
`max_pages` is a static cap of 100 (`streams.json`'s `base.pagination.max_pages`), matching
legacy's `defaultMaxPages` constant exactly — legacy also allows a per-read config override of
`max_pages`, which has no engine-dialect equivalent (see Known limits).

Every stream uses `projection: "passthrough"`: legacy's `readOffset` emits
`connectors.Record(rec)` directly from the raw decoded JSON with no field-built mapping, so
schema-mode projection (which would silently drop any field the schema doesn't declare) would be a
parity violation here. Schemas describe commercetools' real wire shape for the fields most
consumers need (`id`/`version`/`createdAt`/etc.), but `additionalProperties` is not restricted, so
every other raw API field still survives untouched, matching legacy's verbatim passthrough.

Legacy publishes no `CursorFields` for any of these 3 streams and its `Read` path performs no
incremental filtering at all (always a fresh full offset-paginated read from 0) — no `incremental`
block is declared on any stream, matching legacy exactly. `x-cursor-field: createdAt` is
schema-only documentation of each resource's natural timestamp field, not a claim that incremental
sync is implemented.

## Write actions & risks

None. `capabilities.write` is `false`; no `writes.json` is shipped.

## Known limits

- Legacy derives `base_url` and the OAuth2 `token_url` from `region`+`host` config values when
  `base_url`/`token_url` are not explicitly set (`https://api.<region>.<host>.commercetools.com` /
  `https://auth.<region>.<host>.commercetools.com/oauth/token`). This cross-key derivation has no
  engine-dialect mechanism (`spec.json`'s `default` materialization only fills in a FIXED literal,
  never a value computed from another config key — see `docs/migration/conventions.md`'s
  `param_format`/default-materialization section and the sentry/chargebee precedent for the same
  narrowing). This bundle requires `base_url` and `token_url` as fully-formed URLs instead; a
  caller previously relying on the region/host shorthand must now precompute those two URLs itself.
- Legacy's config-overridable `max_pages` (a per-read hard page-count cap, defaulting to 100) has
  no engine-dialect equivalent for a per-call override: `PaginationSpec.MaxPages` is a static
  integer declared in `streams.json`, not templated/config-driven. This bundle declares the static
  default (100), matching legacy's own default for every caller that never overrides it; a caller
  that previously set a different numeric `max_pages` has no equivalent knob here.
- Legacy's `Catalog()` publishes a shared generic `Fields` list (`id`/`version`/`createdAt`) across
  all 3 streams as a simplified illustrative catalog. This bundle's per-stream schemas instead
  describe commercetools' real per-resource wire shape (matching the "recorded-real-shape" fixture
  rule), which is closer to what `records_match_schema` actually needs; runtime output is
  unaffected either way since projection is `passthrough`.
