# Overview

Braze is a declarative-HTTP migration (unblocked from quarantine now that the engine's
`page_number` paginator supports an explicit 0-indexed `start_page`), expanded in Pass B to the
full documented Braze REST API surface (research cross-checked against the community-maintained
OpenAPI spec, 82 documented paths / 95 operations — see `api_surface.json`). It reads Braze
campaigns/Canvases/segments (list, per-id details, and Canvas analytics summary via sub-resource
fan-out), catalogs, content blocks, email templates, Content Cards, email hard-bounce/unsubscribe
lists, SMS invalid-number lists, KPIs, app sessions, preference centers, and upcoming scheduled
broadcasts; it writes user data (track/identify/merge/delete/alias new+update/external-id
remove+rename), subscription-group status, catalog and catalog-item mutations, content block and
email template mutations, email/SMS compliance-list mutations, preference center mutations, and
live message/campaign/Canvas sends. The legacy package (`internal/connectors/braze`) stays
registered and unchanged until wave6's registry flip, and remains authoritative for the two
streams this bundle cannot model at all (`events`, `purchases_product_list` — see Known limits).

## Auth setup

Provide a Braze REST API key via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged, matching legacy's
`connsdk.Bearer(secret)` (`braze.go:252`). Braze has no single global host — each customer is
provisioned on a regional REST endpoint (e.g. `https://rest.iad-01.braze.com`) — so `base_url` is
**required**, matching legacy's `brazeBaseURL`'s hard requirement (`braze.go:275-277`, no
built-in default).

## Streams notes

- `campaigns`/`canvases`/`segments` — `GET` against the Braze list-export endpoint, records at the
  endpoint's own top-level array field, primary key `["id"]`. `campaigns`/`canvases` additionally
  declare `last_edited` as the incremental cursor field; `segments` has none. Braze paginates its
  list-export endpoints with a genuinely 0-based `?page=` query parameter and never sends a
  page-size query param at all — expressed as `pagination.type: page_number`, `page_param: page`,
  `size_param: ""`, `start_page: 0` (base-level, inherited by every stream that doesn't override
  it). The engine's `page_number` paginator stops on a short page (fewer than the declared
  `page_size: 100` records).
- `campaign_details`/`canvas_details`/`segment_details`/`canvas_data_summary` — per-id detail/
  analytics-summary reads, expressed via the engine's `fan_out` dialect: each fans out over EVERY
  id discovered by its own parent list stream's endpoint (`fan_out.ids_from.request` re-issues the
  same paginated `/campaigns/list`, `/canvas/list`, or `/segments/list` request the sibling list
  stream itself uses), threading the resolved id in as a query param (`campaign_id`/`canvas_id`/
  `segment_id`) and stamping it onto every emitted record after projection. Because `fan_out`
  reuses the child stream's OWN effective pagination spec for both the id-listing request AND every
  per-id detail request (conventions.md §3 — same constraint breezy-hr's `candidates` stream
  documents), each per-id `/campaigns/details`, `/canvas/details`, `/canvas/data_summary`, or
  `/segments/details` request also carries a harmless `page=0` query param it wouldn't otherwise
  need; see Known limits.
- `catalogs` — `GET /catalogs`, records at `catalogs`, unpaginated (Braze's docs show no
  pagination params for this endpoint), primary key `["name"]`. Item-level records
  (`/catalogs/{catalog_name}/items`) are not modeled as a stream — see Known limits.
- `content_blocks` — `GET /content_blocks/list`, `offset`/`limit` pagination (`page_size: 100`),
  incremental via `modified_after` (RFC3339, the server-side filter Braze documents), cursor field
  `last_edited`.
- `email_templates` — `GET /templates/email/list`, `offset`/`limit` pagination, incremental via
  `modified_after`, cursor field `updated_at`; `tags` is comma-joined via `computed_fields`
  (`join:,`) since the schema models Braze's raw array as a flat string.
- `feed_cards` — `GET /feed/list` (Content Cards), records at `cards`, inherits base `page_number`
  pagination (0-based `page`, no size param, matching the same convention as `campaigns`/`canvas`).
- `email_hard_bounces`/`email_unsubscribes` — `GET /email/hard_bounces` /
  `GET /email/unsubscribes`, `offset`/`limit` pagination, incremental via the `start_date`
  server-side filter (`param_format: date`), cursor fields `hard_bounced_at`/`unsubscribed_at`.
- `sms_invalid_phone_numbers` — `GET /sms/invalid_phone_numbers`; the response's top-level array
  key is `sms` (NOT `invalid_phone_numbers` — confirmed against the live docs), each record's
  fields are `phone`/`invalid_detected_at`/`reason`; `offset`/`limit` pagination, incremental via
  `start_date`, cursor field `invalid_detected_at`.
- `kpi_dau`/`kpi_mau`/`kpi_new_users`/`kpi_uninstalls`/`sessions` — the 5 `.../data_series` KPI/
  session analytics endpoints, records at `data`, unpaginated (a single `length`-bounded time
  window per request; `length: "14"` requests the last 14 days), cursor field `time` (not wired as
  an `incremental` block — Braze's own `length`/`ending_at` window params are not the same shape as
  a lower-bound cursor filter, so these streams are full-refresh-only, matching the analogous
  export-style endpoints elsewhere in this bundle).
- `preference_centers` — `GET /preference_center/v1/list`, records at `preference_centers`
  (confirmed against the live docs — NOT a bare array), unpaginated, incremental cursor field
  `updated_at` (no server-side filter; full re-read each sync).
- `scheduled_broadcasts` — `GET /messages/scheduled_broadcasts`, records at the response ROOT (a
  bare top-level array, unlike every other stream in this bundle), unpaginated; `tags` is
  comma-joined via `computed_fields`.

## Write actions & risks

29 write actions, grouped by risk:

- **Destructive (`confirm: destructive`)**: `delete_users` (irreversible profile deletion),
  `remove_user_external_ids` (detaches an id, orphaning the profile), `delete_catalog`/
  `delete_catalog_item` (irreversible catalog/row removal), `create_email_blocklist` (Braze
  documents this as effectively irreversible via the API), `send_message`/`trigger_campaign_send`/
  `trigger_canvas_send` (dispatch a real, immediate, irreversible communication to end users — the
  single riskiest action class this connector exposes).
- **User data mutations**: `track_users`, `identify_users`, `merge_users`, `create_user_alias`/
  `update_user_alias`, `rename_user_external_ids`, `set_subscription_status_v2` (the v2 batch
  endpoint supersedes v1's single-group shape for any input v1 accepts, per `api_surface.json`).
- **Catalog mutations**: `create_catalog`, `create_catalog_items`/`update_catalog_items` (batch,
  `body_fields: ["items"]` restricts the body to the item array — `catalog_name` is path-only),
  `update_catalog_item` (single-item variant, same `items`-array body shape Braze's own API
  requires even for a single-id PATCH).
- **Content mutations**: `create_content_block`/`update_content_block`, `create_email_template`/
  `update_email_template`, `create_preference_center`/`update_preference_center` — all low-risk
  additive/update mutations but each risk string flags that live campaigns/Canvases referencing the
  mutated object pick up the change on their NEXT send, including already-scheduled ones.
- **Compliance-list mutations**: `remove_email_hard_bounce`, `remove_email_spam`,
  `set_email_subscription_status`, `remove_sms_invalid_phone_numbers` — each risk string warns
  against reversing a bounce/spam/invalid-number flag without confirming the underlying delivery
  issue is actually fixed.

## Known limits

- **Two streams are not modeled by this bundle (both ENGINE_GAP, the same underlying limitation):
  `events` (unchanged from wave2) and `purchases_product_list` (newly identified this pass).**
  Braze's `/events/list` and `/purchases/product_list` endpoints both return a JSON array of bare
  STRINGS (event names / product names respectively), not objects — legacy's `decodeRecords`
  special-cases this shape by wrapping every string element into a one-field object before mapping.
  The engine's declarative record extraction (`connsdk.RecordsAt`,
  `internal/connectors/connsdk/extract.go:33-56`) only ever yields a `Record` for an array element
  that decodes as a JSON object (`map[string]any`); a bare string element is silently skipped, not
  wrapped — there is no declarative primitive to turn a scalar array element into a one-field
  record. This is a genuine ENGINE_GAP (an omission would silently under-report data, not a
  defensible parity approximation), not a Tier-1/2-fixable shape. Legacy stays authoritative for
  both streams until the engine gains an array-of-scalars wrapping primitive; see
  `api_surface.json`'s `excluded` entries for `/events/list` and `/purchases/product_list`.
- **The 4 fan-out detail/summary streams (`campaign_details`, `canvas_details`,
  `canvas_data_summary`, `segment_details`) re-paginate their own id-listing request AND their
  per-id detail request with the identical `page_number` spec** (conventions.md §3: `fan_out`
  reuses the child stream's own effective pagination for both purposes, with no way to declare one
  spec for the id-listing sub-request and a different one for the per-id sub-sequence). Since
  Braze's actual `/campaigns/details`, `/canvas/details`, `/canvas/data_summary`, and
  `/segments/details` endpoints return a single object (not a list), each per-id request carries a
  harmless, unused `page=0` query parameter — Braze's API ignores unrecognized query params on
  these endpoints, so this never changes the emitted detail record's data for any campaign/
  Canvas/segment id a legacy-equivalent single-object read would return. Documented parity
  deviation, ACCEPTABLE (same class as breezy-hr's identical `candidates` deviation).
- **Catalog ITEM-level records are not modeled as their own stream.** `GET
  /catalogs/{catalog_name}/items` requires an already-known `catalog_name`; wiring it as a
  `fan_out` child of the `catalogs` stream (analogous to the campaign/canvas/segment detail
  streams) was scoped out of this pass for breadth-vs-cost triage — the `catalogs` stream itself
  already syncs every catalog's name/description/field-schema/item-count metadata. Tracked as
  future Pass B follow-up, not an ENGINE_GAP (the dialect can express it).
- Full Braze API surface still excluded as genuinely out of scope this pass: async broadcast/
  trigger SCHEDULING (a job-lifecycle shape this dialect's synchronous fire-once write model
  doesn't fit), Live Activity updates, SCIM user provisioning, Transactional Messaging, the
  ~15-endpoint analytics `data_series`/breakdown matrix beyond the single-object summaries already
  covered, and various duplicate/non-data/destructive-admin endpoints — see `api_surface.json` for
  the complete, closed-vocabulary per-endpoint accounting.
- `page_size`/`max_pages` config overrides legacy exposes (`brazePageSize`/`brazeMaxPages`, clamped
  1-100 / `all`/`unlimited`) are not runtime-configurable here: the engine's `page_number`
  paginator's `PageSize` is a static int set once in `streams.json`, not template-resolvable, and
  `PaginationSpec` has no config-driven `MaxPages` override wired to any spec key. `spec.json`
  intentionally does not declare `page_size`/`max_pages` (a declared-but-unwireable key is worse
  than an absent one, per conventions.md F6).
