# Overview

Campaign Monitor reads and writes Campaign Monitor clients, campaigns, subscriber lists,
subscribers, segments, and templates through the createsend.com v3.3 REST API
(`https://api.createsend.com/api/v3.3/...`). This bundle targets capability parity with
`internal/connectors/campaign-monitor` (the hand-written connector it migrates, Go package
`campaignmonitor`); the legacy package stays registered and unchanged until wave6's registry flip.

**Pass B full-surface expansion** (`api_surface.json`, reviewed 2026-07-04 against the live
campaignmonitor.com/api/ docs across its Account/Clients/Campaigns/Lists/Subscribers/Segments/
Templates/Transactional sections): beyond the 4 legacy streams (`clients`, `campaigns`, `lists`,
`suppressionlist`), this bundle adds 9 more: `segments` and `templates` (fanned out per client),
`list_custom_fields` and `list_webhooks` (fanned out per list), and 5 subscriber-state streams
(`active_subscribers`, `unconfirmed_subscribers`, `unsubscribed_subscribers`, `bounced_subscribers`,
`deleted_subscribers`, likewise fanned out per list). `capabilities.write` is now `true`, with 13
write actions covering list/subscriber/segment/campaign lifecycle. Transactional (a genuinely
separate Campaign Monitor product for single-recipient triggered email, not bulk marketing
campaigns) and every agency-administrative surface (admins, billing, client provisioning, sending
domains) are out of scope — see `api_surface.json` for the full accounting.

## Auth setup

Campaign Monitor authenticates with HTTP Basic auth: the account/client API key is the username,
and the password may be blank or a dummy value. Provide the API key via the `username` config value
(required). An optional `password` secret is sent as the Basic auth password when set (`auth`
candidate 1: `when: "{{ secrets.password }}"`); when `password` is unset, the second candidate
(`mode: basic`, no `when`, always matches) sends the literal dummy password `"x"` instead — matching
legacy's own `cmDefaultDummyPasswrd` fallback (`campaign_monitor.go:258-262`) exactly via the
dual-auth-candidate first-match-wins pattern (`docs/migration/conventions.md` §3's zendesk-support
precedent). `base_url` defaults to `https://api.createsend.com/api/v3.3` and may be overridden for
tests/proxies.

## Streams notes

`clients` (`GET /clients.json`) and `lists` (`GET /clients/{{ config.client_id }}/lists.json`) are
bare top-level JSON arrays with no pagination (`pagination.type: none`), matching legacy's
`endpoint.paged == false` / `readArray` branch. `campaigns` and `suppressionlist` use Campaign
Monitor's page/`NumberOfPages` envelope (`{Results:[...], PageNumber, NumberOfPages}`) —
`pagination.type: page_number` with `page_param: page`, `size_param: pagesize`, `page_size: 100`;
records live at `Results`. A page shorter than 100 records stops pagination, functionally equivalent
to legacy's own `current >= numberOfPages` stop check for every real response (the last page is
never longer than the requested `pagesize`); an exact-multiple-of-100 result set costs one extra,
empty-page request on the engine side that legacy's page-count check would have avoided, never a
data difference.

`lists` and `campaigns` (as well as `suppressionlist`) are scoped under `/clients/{{ config.client_id
}}/...`; `client_id` is urlencoded into the path by `InterpolatePath`'s per-segment default, matching
legacy's own `url.PathEscape(clientID)` in `resolveResource`. An absent `client_id` hard-errors on
both sides (legacy: `"campaign-monitor config client_id is required for this stream"`; engine: an
unresolved `config.client_id` path-template key). Cursor fields match legacy's own declarations:
`campaigns` -> `SentDate`, `suppressionlist` -> `Date`; `clients`/`lists` have none (full refresh),
matching legacy's `CursorFields: nil` for both.

**New Pass B streams**: `segments` and `templates` fan out per CLIENT (`fan_out.ids_from.request`
against `/clients.json`, `into.path_var: "client_id"`, `stamp_field: "OwningClientID"` — the docs'
own client-scoped GETs return no client-identifying field of their own, so the fan-out id is stamped
onto every record so the partition is recoverable downstream); both are unpaginated bare arrays
(`pagination.type: none`), matching their documented shape. `list_custom_fields`, `list_webhooks`,
and the 5 subscriber-state streams (`active_subscribers`/`unconfirmed_subscribers`/
`unsubscribed_subscribers`/`bounced_subscribers`/`deleted_subscribers`) fan out per LIST instead
(`ids_from.request` against `/clients/{{ config.client_id }}/lists.json`, `into.path_var:
"list_id"`, `stamp_field: "ListID"`) — the 5 subscriber-state streams additionally paginate
(`page_number`, `page_param: page`, `size_param: pagesize`, `page_size: 1000` matching the docs'
own default `PageSize`) and declare `incremental.cursor_field: Date`, all sharing one
`schemas/subscribers.json` (identical `Subscriber` record shape across every state, per the docs);
`list_custom_fields`/`list_webhooks` are unpaginated bare arrays like segments/templates.

## Write actions & risks

`capabilities.write` is now `true` (Pass B). 14 actions:

- **Lists**: `create_list`, `update_list` (low risk), `delete_list` (irreversible — removes the
  list and everything under it: subscribers, segments; approval required).
- **Subscribers**: `add_subscriber`, `update_subscriber` (identified by `CurrentEmailAddress`, sent
  as a query parameter rather than a body field per the docs — declared as a `path_fields` entry so
  it is excluded from the JSON body and instead appended to the path template as `?email=...`),
  `unsubscribe_subscriber` (low risk), `delete_subscriber` (permanently removes the subscriber
  record, distinct from unsubscribing; approval recommended).
- **Segments**: `create_segment`, `update_segment` (replaces the full `RuleGroups` rule set — not
  a partial patch), `delete_segment` (all low risk).
- **Campaigns**: `create_campaign` (creates a DRAFT only — no delivery side effect on its own),
  `send_campaign` (delivers real email to every targeted recipient; irreversible once sent, approval
  required), `unschedule_campaign` (reverts a scheduled campaign back to draft; low risk),
  `delete_campaign` (approval recommended).

## Known limits

- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config overrides
  (`cmPageSize`/`cmMaxPages`, `campaign_monitor.go:318-346`). The engine's `page_number` paginator's
  `PageSize` is a static bundle-authored int (not templated), and there is no `MaxPages`-equivalent
  config-driven knob either; `page_size` is fixed at legacy's own default (100) in `streams.json`'s
  base pagination block, and `max_pages` is unbounded (matching legacy's own
  `max_pages=0`/`all`/`unlimited` default), following bitly's identical documented scope-narrowing
  precedent (`docs/migration/conventions.md`).
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`) synthesizes deterministic records with no network access
  (`campaign_monitor.go:216-244`); this bundle's schemas and fixtures target the live path only.
- **`update_subscriber`'s identifying email is a query parameter, not a path segment.** The docs'
  `PUT /subscribers/{listid}.json?email={email}` shape identifies the subscriber to update via a
  query string, not a `{subscriber_id}`-style path segment like every other update action in this
  bundle. `CurrentEmailAddress` is declared as a `path_fields` entry purely to exclude it from the
  JSON body (matching the docs, which never show the current email echoed back in the body); the
  actual value is embedded directly in the `path` template as a literal `?email={{ record.
  CurrentEmailAddress | urlencode }}` suffix (conformance's `write_request_shape` check only asserts
  `r.URL.Path`, which excludes the query string, so fixtures assert the bare path only — this is a
  fixture/test-harness scoping note, not a behavioral gap in the actual outgoing request).
- Every remaining known Campaign Monitor endpoint is either covered or excluded with a specific
  reason; see `api_surface.json`'s `excluded` entries for the full accounting (agency-administrative
  surfaces, the separate Transactional product, bulk/import variants of already-covered per-record
  writes, and analytics/event-log endpoints judged out of scope for this pass).
