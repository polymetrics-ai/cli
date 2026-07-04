# Overview

EmailOctopus is a wave2 fan-out declarative-HTTP migration, expanded to the full documented v1.6
API surface in Pass B (2026-07-04). It reads lists, campaigns, per-list contacts, per-list tags,
and per-campaign summary reports through the EmailOctopus v1.6 REST API (`{{ config.base_url
}}/...`), and writes list/contact/tag/custom-field lifecycle mutations plus automation enrollment.
This bundle originated as a migration from `internal/connectors/emailoctopus` (the hand-written
connector it replaces, which stays registered and unchanged until wave6's registry flip);
`capabilities.write` is now `true` following the Pass B write-surface expansion — this is a
capability WIDENING beyond legacy's read-only `Write` stub, not a parity port, since EmailOctopus's
real v1.6 API genuinely exposes these mutation endpoints and legacy never called them.

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

**Pass B new streams**: `list_tags` (`GET /lists/{{ config.list_id }}/tags`, same envelope/
pagination/`list_id` config requirement as `list_contacts`; primary key `tag` — the API's own
identifier, there is no separate numeric tag id) and `campaign_summary_reports` (`GET
/campaigns/{{ fanout.id }}/reports/summary`, a `fan_out` stream — `ids_from.request` lists every
campaign id off the `campaigns` endpoint's own `data[].id`, then issues one summary-report request
per campaign id, stamping `campaign_id` onto every emitted record). The real report endpoint
returns a single object, not a paginated array (`pagination: none`, `records.path: ""`, matching
the `seller_funds_summary`-style single-object convention used elsewhere in this codebase); nested
`bounced.{hard,soft}`/`opened.{total,unique}`/`clicked.{total,unique}` are flattened via
`computed_fields` into `bounced_hard`/`bounced_soft`/`opened_total`/`opened_unique`/
`clicked_total`/`clicked_unique`. The real API errors on a campaign whose `status` is not `SENT`
("The campaign does not have a status of 'SENT'"); this bundle does not filter fanned-out campaign
ids by status first, so a per-id request failure for a not-yet-sent campaign propagates as an
ordinary read error for that id — an accepted, documented limitation (see Known limits), not a
silent skip.

## Write actions & risks

Pass B added `capabilities.write: true` and 13 actions — see `writes.json` for the full
`record_schema`/`risk` text per action. Summary:

- **lists**: `create_list` (`POST /lists`, only `name` is a real accepted field), `update_list`
  (`PUT /lists/{id}`, body restricted to `name` — the real API accepts no other update field),
  `delete_list`.
- **list contacts**: `create_list_contact` (`POST /lists/{list_id}/contacts`), `update_list_contact`
  (`PUT /lists/{list_id}/contacts/{member_id}` — `member_id` is the real API's own contact id OR an
  MD5 hash of the lowercased email address; this bundle passes whatever value the caller supplies
  verbatim, it does not compute the MD5 hash itself), `delete_list_contact`.
- **list tags**: `create_list_tag`, `update_list_tag` (body field is `new_tag`, the real API's own
  rename-target field name — distinct from the path-keyed current `tag`), `delete_list_tag`.
- **list custom fields**: `create_list_field` (`type` is immutable after creation per the real
  API — `NUMBER`/`TEXT`/`DATE`), `update_list_field` (body fields `label`/`new_tag`/`fallback` —
  `type` is deliberately NOT in the update body, matching the real API's own immutability rule),
  `delete_list_field`.
- **automations**: `start_automation` (`POST /automations/{automation_id}/queue`) — flagged
  **higher-risk**: enrolls a real contact into a live automation sequence with its own configured
  email sends/delays; the target automation must already have EmailOctopus's "Started via API"
  trigger type enabled (an out-of-band dashboard prerequisite this connector cannot verify).

Every write action's `path_fields` names the real API's own path-parameter identifier — natural
resource ids (`id` for lists, `member_id` for contacts, `tag` for list tags/fields, `automation_id`)
rather than any synthetic key this connector invents.

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
- **Pass B full-surface review (2026-07-04).** `api_surface.json` now enumerates the entire
  documented EmailOctopus v1.6 REST surface (29 endpoints). This connector deliberately stays on
  v1.6 (the version its existing `base_url`/spec already target) rather than migrating to v2, which
  the vendor's docs recommend for new integrations but which is a materially different API surface
  outside this pass's scope. Deliberately excluded: the 8 per-recipient campaign engagement-detail
  reports (bounced/clicked/complained/opened/sent/unsubscribed/not-clicked/not-opened) — large
  per-recipient analytics breakdowns distinct from the aggregate `campaign_summary_reports` stream
  already covered — and the bulk multi-contact update endpoint (`PUT
  /lists/{id}/contacts/bulk`, up to 100 contacts per call; `update_list_contact` already covers the
  per-record shape this dialect expresses). See `api_surface.json`'s per-endpoint `excluded.reason`
  for every other omission (mostly `duplicate_of` single-object/status-filtered views already
  covered by their list stream).
- **`campaign_summary_reports` does not pre-filter fanned-out ids by `status: SENT`.** The real
  report endpoint 400s for any campaign not yet fully sent; since the `campaigns` stream's own
  fixture/fan_out id-source does not filter on status, a live sync against an account with
  in-progress/draft campaigns will surface a per-id read failure for those ids rather than silently
  skipping them. This mirrors the fan_out dialect's own fail-fast design (§3,
  `docs/migration/conventions.md`) — a per-id request failure is a real error condition to
  surface, not a case the declarative dialect should paper over with an invented status filter.
- **`update_list_contact`'s `member_id` MD5-hash convention is caller-supplied, not
  computed.** The real API accepts either the contact's own id or an MD5 hash of the lowercased
  email address as `member_id`; this connector has no `md5` interpolation filter, so a caller
  wanting to address a contact by email must compute that hash themselves before calling
  `update_list_contact`/`delete_list_contact` — passing a raw email address directly will not
  resolve to the expected contact.
