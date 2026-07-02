# Overview

EmailOctopus is a wave2 fan-out declarative-HTTP migration. It reads lists, campaigns, and
per-list contacts through the EmailOctopus v1.6 REST API (`GET
{{ config.base_url }}/...`). This bundle is migrated from `internal/connectors/emailoctopus` (the
hand-written connector it replaces); the legacy package stays registered and unchanged until
wave6's registry flip. Read-only (`capabilities.write` is `false`, matching legacy's `Write`
returning `connectors.ErrUnsupportedOperation`).

## Auth setup

Provide an EmailOctopus API key via the `api_key` secret; it is sent as the `api_key` query
parameter on every request (`mode: api_key_query`), matching legacy's
`connsdk.APIKeyQuery("api_key", secret)`. `base_url` defaults to
`https://emailoctopus.com/api/1.6` and may be overridden for test proxies.

## Streams notes

All 3 streams share the identical EmailOctopus envelope (`{data:[...], paging:{next, previous}}`,
records at `data`) and `next_url` pagination (`paging.next`, an absolute URL the API itself
returns — same-host by default, matching legacy's `harvest` loop, which follows `paging.next`
directly and stops when it is `null`). `lists` (`GET /lists`) and `campaigns` (`GET /campaigns`)
need no config beyond `api_key`/`base_url`. `list_contacts` (`GET
/lists/{{ config.list_id }}/contacts`) requires the `list_id` config value to resolve the path —
absent when `list_id` is unset, matching legacy's `resolveResource`, which errors with
`"emailoctopus stream \"list_contacts\" requires config list_id"` for the identical case (this
bundle's path-templating error is a more generic engine message, not legacy's specific wording —
see Known limits).

`lists`' nested `counts.pending`/`counts.subscribed`/`counts.unsubscribed` are flattened into
`pending_count`/`subscribed_count`/`unsubscribed_count` via `computed_fields`, and `campaigns`'
nested `from.name`/`from.email_address` are flattened into `from_name`/`from_email_address`,
matching legacy's `listRecord`/`campaignRecord` mappers exactly. `list_contacts`' fields
(`id`/`email_address`/`status`/`tags`/`fields`/`created_at`/`last_updated_at`) pass straight
through with no renaming, matching legacy's `contactRecord`. Primary key is `id` for every stream.
None declare an incremental cursor — legacy exposes none (EmailOctopus v1.6 has no time-range
filter parameter any stream's request ever sends).

## Write actions & risks

None. Legacy `emailoctopus.go`'s `Write` returns `connectors.ErrUnsupportedOperation`
unconditionally; `capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **Every stream in this bundle uses `next_url` pagination, so the sanctioned "2-page fixture,
  except a single-page `next_url` exception proven live by a `paritytest/<name>` suite"
  (`docs/migration/conventions.md` §4) does not fully apply here**: the exception's intended
  shape pairs a single-page `next_url` fixture with a DIFFERENT non-paginated stream in the same
  bundle for `pagination_terminates`, and a live `httptest.Server`-backed parity test proving real
  2-page `next_url` correctness. This bundle has no non-paginated stream to substitute, and this
  wave's mandate is JSON/`docs.md` only (no `paritytest` Go package). Each stream therefore ships
  a single-page fixture (satisfying `fixtures_present`/`read_fixture_nonempty`); `lists` (the
  first declared stream) is `pagination_terminates`' candidate and passes trivially (one fixture
  page, one request, `paging.next: null` stops immediately) — this proves the paginator does not
  loop on a terminal page, but does NOT exercise the actual next-page-follow behavior end to end.
  A genuine two-hop `next_url` follow (a real absolute second-page URL, re-authenticated,
  query-preserving) for this connector is unverified by this wave's fixtures and should be closed
  by a follow-up wave's `paritytest/emailoctopus` suite (the same pattern already used by
  `paritytest/bitly`/`paritytest/calendly`), not silently assumed correct.
- **`page_size`/`max_pages` config overrides are not modeled.** Legacy accepts optional
  `page_size` (1-100, default 100) and `max_pages` (default unlimited, `all`/`unlimited`/`0`
  synonyms) config keys read at request time (`emailOctopusPageSize`/`emailOctopusMaxPages`). The
  engine's `next_url` paginator does not consult `PaginationSpec.PageSize` at all past the first
  request (subsequent pages are whatever URL/query the API itself returns); this bundle sends a
  static `limit: "100"` on the first request only (legacy's own default), and neither key is
  declared in `spec.json` (F6, `docs/migration/conventions.md`: dead, unwireable config is worse
  than absent config).
- **`base_url` scheme/host validation is enforced by legacy in Go** with dedicated error messages
  (`emailOctopusBaseURL`); the engine has no equivalent declarative URL-shape validator, so a
  malformed `base_url` surfaces as a generic request-construction/connection error rather than
  legacy's specific messages. The engine's own same-origin SSRF guard on `next_url` pagination
  (`allow_cross_host` defaults `false`) independently bounds cross-host redirection risk for the
  `paging.next` follow itself.
- The full EmailOctopus API surface (list/campaign/contact mutation, campaign delivery reports)
  is out of scope for this wave; see `api_surface.json`'s `excluded` entries.
