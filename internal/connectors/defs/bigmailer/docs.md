# Overview

BigMailer is a wave2 fan-out declarative-HTTP migration, expanded to full API-surface coverage
(reads and writes) in Pass B. It reads BigMailer brands, account users, connections, and
brand-scoped contacts, lists, custom fields, message types, segments, senders, templates,
suppression lists, and campaign metadata (bulk/RSS/transactional/test) through the BigMailer REST
API (`GET https://api.bigmailer.io/v1/<resource>`), and writes brands, contacts, lists, fields,
message types, segments, senders, templates, and account users. This bundle is migrated from
`internal/connectors/bigmailer` (the hand-written connector); the legacy package stays registered
and unchanged until wave6's registry flip. The original 5 legacy streams (`brands`, `users`,
`contacts`, `lists`, `fields`) use the engine's `fan_out` dialect (S4 engine mini-wave item 2) for
the 3 brand-scoped substreams — the `ENGINE_GAP` that previously blocked them is closed. Pass B
adds 10 more streams (`connections` top-level; `message_types`/`segments`/`senders`/`templates`/
`suppression_lists`/`bulk_campaigns`/`rss_campaigns`/`transactional_campaigns`/`test_campaigns`
brand-scoped, same `fan_out` shape) and a full `writes.json` (`capabilities.write` flips to
`true`).

## Auth setup

Provide a BigMailer API key via the `api_key` secret; it is sent as the `X-API-Key` header
(`streams.json` `base.auth`'s `api_key_header` mode), matching legacy's
`connsdk.APIKeyHeader(bigmailerAuthHeader, secret, "")` (`bigmailer.go:311`). Never logged.
`base_url` defaults to `https://api.bigmailer.io/v1` and may be overridden for tests/proxies.

## Streams notes

`brands` (`GET /brands`) and `users` (`GET /users`) are top-level collections read directly;
records live at the `data` key, matching legacy's `connsdk.RecordsAt(resp.Body, "data")`. Neither
stream is incremental — legacy declares no `CursorFields` for either (BigMailer's list API supports
only cursor pagination, not a time-based incremental filter), and neither schema declares
`x-cursor-field`.

Pagination is `cursor` (`token_path: cursor`, `stop_path: has_more`): the next page is requested
with `cursor=<value>` from the response body's `cursor` field, and pagination stops when
`has_more` is not literally `"true"` OR the returned `cursor` is empty — reproducing legacy's exact
`hasMore != "true" || strings.TrimSpace(next) == ""` stop rule (`bigmailer.go:187`) via the
engine's `stop_path`-on-`tokenPathCursor` mechanism (conventions.md §3).

`connections` (`GET /connections`) is a new top-level collection, identical shape to `brands`/
`users` (records at `data`, `cursor` pagination, no incremental filter — BigMailer's list API
supports only cursor pagination).

`message_types`, `segments`, `senders`, `templates`, `suppression_lists`, `bulk_campaigns`,
`rss_campaigns`, `transactional_campaigns`, and `test_campaigns` are new brand-scoped substreams,
following the identical `fan_out` shape documented below for `contacts`/`lists`/`fields`: list
every brand id via `GET /brands`, then paginate `GET /brands/{brand_id}/<resource>` once per
brand, stamping `brand_id` onto every emitted record. The 4 campaign streams
(`bulk_campaigns`/`rss_campaigns`/`transactional_campaigns`/`test_campaigns`) declare
`x-cursor-field: created` (a genuine Unix-seconds creation timestamp on every campaign object) but
no `incremental` block — BigMailer's campaign list endpoints support only cursor pagination, no
server-side time filter, matching every other stream in this bundle. RSS campaigns' auto-generated
per-feed-update sub-resource (`GET .../rss-campaigns/{id}/updates`) is not modeled as a further
fan-out level (`api_surface.json`: `non_data_endpoint` — those records are derived campaign
instances, not independently authored data).

`contacts`, `lists`, and `fields` are brand-scoped substreams: legacy's `harvestSubstream`
(`bigmailer.go:195-214`) first lists every brand id (`listBrandIDs`, bounded defensively by
`bigmailerMaxBrands = 1000`), then paginates `GET /brands/{brand_id}/<resource>` once per brand,
stamping `brand_id` onto every emitted record. This bundle expresses the identical sequence via
`streams.json`'s `fan_out` block: `ids_from.request` issues a preliminary `GET /brands` (paginated
to exhaustion using the SAME base `cursor` pagination spec every other stream uses — the
id-listing request declares no pagination block of its own, conventions.md §3), extracts `id` off
every returned brand record, then `into.path_var: "brand_id"` threads each resolved id into
`/brands/{{ fanout.id }}/<resource>`'s path, and `stamp_field: "brand_id"` writes it onto every
emitted record after projection — matching legacy's stamped `brand_id` field exactly. The one
documented, non-blocking divergence: legacy caps the brand-id fan-out at `bigmailerMaxBrands =
1000` as a defensive bound against a runaway crawl; the engine's `fan_out.ids_from.request` has no
equivalent cap (only `PaginationSpec.MaxPages`, applied per sub-sequence, not to the id-listing
request) — an account with more than 1000 brands would fan out over all of them here versus being
capped at 1000 in legacy. This is accepted as a parity deviation (§5): it never changes emitted
record DATA for any account legacy itself would fully sync (mid-cap accounts are identical;
over-cap accounts get MORE data here, never less or wrong), and no such account exists in the
fixture/conformance surface.

## Write actions & risks

BigMailer's legacy connector was read-only, but Pass B's full-surface research found BigMailer's
CRUD mutation surface for every covered resource is plain-JSON-bodied and fully dialect-expressible
— `capabilities.write` now flips to `true` and this bundle ships a full `writes.json`. Every action
requires operator approval (external mutation) per its own `risk` string:

- `create_brand` / `update_brand` — creates/updates a BigMailer brand (sending identity). Does not
  itself send email.
- `create_contact` / `update_contact` / `upsert_contact` / `delete_contact` — full contact CRUD in
  a brand. `delete_contact` is idempotent (`missing_ok_status: [404]`).
- `create_list` / `update_list` / `delete_list` — contact list CRUD. Deleting a list does not
  delete its contacts (BigMailer's own semantics).
- `create_field` / `update_field` / `delete_field` — custom contact field CRUD.
- `create_message_type` / `update_message_type` / `delete_message_type` — message-type (unsubscribe
  category) CRUD.
- `create_segment` / `update_segment` / `delete_segment` — contact segment CRUD.
- `create_sender` / `update_sender` / `delete_sender` — sender domain/email identity CRUD. Does not
  perform DNS verification (see Known limits).
- `create_template` / `update_template` / `delete_template` — campaign template CRUD.
- `create_user` / `update_user` / `delete_user` — account user CRUD (invites/removes BigMailer
  console users, not contacts).

Every path-parameterized action (all except `create_user`) requires the record to carry `brand_id`
(and, for update/delete, the resource's own `id`) as ordinary record fields — the engine's
`path_fields` mechanism excludes them from the request body automatically, matching every other
two-path-var write in this repo (e.g. webflow's `update_collection_item`).

**Not modeled — see `api_surface.json` for the full per-endpoint breakdown**: campaign
create/update/send actions (bulk/RSS/transactional/test campaigns) are `destructive_admin` —
BigMailer campaigns dispatch real outbound email to real recipients once scheduled/sent, and this
is the one write shape a data-integration connector should never expose without a human explicitly
composing and reviewing the send in BigMailer's own console. RSS campaign pause/unpause are the
same class (immediately affects a live recurring send schedule). Contact-batch upload and
suppression-list upload are `requires_elevated_scope` (multipart file upload, no async-job-polling
mechanism in the write dialect). Sender/bounce-domain DNS verification are `requires_elevated_scope`
(external DNS side effects with no data record to write).

## Known limits

- **`contacts`/`lists`/`fields` fan-out has no brand-count cap.** Legacy's `listBrandIDs` bounds
  the brand-id fan-out at `bigmailerMaxBrands = 1000` as a defensive measure. The engine's
  `fan_out.ids_from.request` fully paginates the id-listing request to exhaustion with no
  equivalent cap. Documented parity deviation (§5, ACCEPTABLE): never changes emitted data for any
  account legacy itself would fully sync; only affects the hypothetical >1000-brand account, where
  this bundle emits strictly more (never wrong or missing) data than legacy's capped crawl.
- **`page_size`/`max_pages` config overrides are not modeled.** Legacy exposes `page_size`
  (1-100, default 100) and `max_pages` (0/all/unlimited default) as config-driven overrides
  (`bigmailerPageSize`/`bigmailerMaxPages`, `bigmailer.go:344-372`). The engine's `cursor`
  paginator has no page-size-equivalent knob at all (BigMailer's `limit` query param is a static
  per-stream `query` literal here, matching stripe's `limit=100` static-query precedent), and
  `PaginationSpec.MaxPages` is a fixed bundle-time int, never `config.*`-templated
  (`docs/migration/conventions.md`'s searxng/bitly precedent). `limit=100` (legacy's own default)
  is baked into each stream's static `query`; neither `page_size` nor `max_pages` is declared in
  `spec.json` (F6, REVIEW.md).
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only
  reached when `config.mode == "fixture"`) is a synthetic, non-live-shape fixture generator, not a
  representation of BigMailer's real wire format for `brands`/`users` — this bundle's schemas and
  fixtures instead use the real `{data:[...], has_more, cursor}` envelope legacy's live `harvest`
  path actually decodes.
- **Every campaign send/lifecycle action, contact-batch upload, suppression-list upload, and
  sender/bounce-domain DNS verification is out of scope** (`api_surface.json`: `destructive_admin`
  / `requires_elevated_scope`). See "Write actions & risks" above for the full reasoning per
  action.
- **Brand `properties` (an account-configuration key/value namespace) is not modeled as a
  stream or write surface.** It is settings metadata, not a contact/CRM/campaign data resource
  meaningfully synced alongside the rest of this bundle.
