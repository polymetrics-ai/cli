# Overview

Reads Braze campaigns, Canvases, segments (list + per-id details/analytics-summary), catalogs,
content blocks, email templates, Content Cards, email bounce/unsubscribe lists, SMS invalid-number
lists, KPIs, sessions, preference centers, and scheduled broadcasts; writes user data
(track/identify/merge/alias/delete), subscription-group status, catalog and catalog-item mutations,
content block/email template mutations, email/SMS compliance-list mutations, preference center
mutations, and live message/campaign/Canvas sends through the Braze REST API.

Readable streams: `campaigns`, `canvases`, `segments`, `campaign_details`, `canvas_details`,
`canvas_data_summary`, `segment_details`, `catalogs`, `content_blocks`, `email_templates`,
`feed_cards`, `email_hard_bounces`, `email_unsubscribes`, `sms_invalid_phone_numbers`, `kpi_dau`,
`kpi_mau`, `kpi_new_users`, `kpi_uninstalls`, `sessions`, `preference_centers`,
`scheduled_broadcasts`.

Write actions: `track_users`, `identify_users`, `delete_users`, `merge_users`, `create_user_alias`,
`update_user_alias`, `remove_user_external_ids`, `rename_user_external_ids`,
`set_subscription_status_v2`, `create_catalog`, `delete_catalog`, `create_catalog_items`,
`update_catalog_items`, `update_catalog_item`, `delete_catalog_item`, `create_content_block`,
`update_content_block`, `create_email_template`, `update_email_template`, `create_email_blocklist`,
`remove_email_hard_bounce`, `remove_email_spam`, `set_email_subscription_status`,
`remove_sms_invalid_phone_numbers`, `create_preference_center`, `update_preference_center`,
`send_message`, `trigger_campaign_send`, `trigger_canvas_send`.

Service API documentation: https://www.braze.com/docs/api/basics/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Braze REST API key. Sent as Authorization: Bearer <api_key>.
  Never logged.
- `base_url` (required, string); format `uri`; Your regional Braze REST endpoint, e.g.
  https://rest.iad-01.braze.com. Braze has no single global host, so this is required.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/campaigns/list` with query `page`=`0`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; no page-size parameter; starts at
0; page size 100.

Pagination by stream: none: `catalogs`, `kpi_dau`, `kpi_mau`, `kpi_new_users`, `kpi_uninstalls`,
`sessions`, `preference_centers`, `scheduled_broadcasts`; offset_limit: `content_blocks`,
`email_templates`, `email_hard_bounces`, `email_unsubscribes`, `sms_invalid_phone_numbers`;
page_number: `campaigns`, `canvases`, `segments`, `campaign_details`, `canvas_details`,
`canvas_data_summary`, `segment_details`, `feed_cards`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `campaigns`: GET `/campaigns/list` - records path `campaigns`; page-number pagination; page
  parameter `page`; no page-size parameter; starts at 0; page size 100; incremental cursor
  `last_edited`; formatted as `rfc3339`.
- `canvases`: GET `/canvas/list` - records path `canvases`; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 0; page size 100; incremental cursor `last_edited`;
  formatted as `rfc3339`.
- `segments`: GET `/segments/list` - records path `segments`; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 0; page size 100.
- `campaign_details`: GET `/campaigns/details` - records path `.`; page-number pagination; page
  parameter `page`; no page-size parameter; starts at 0; page size 100; computed output fields
  `channels`, `tags`, `teams`; fan-out; ids from request `/campaigns/list`; id-list records path
  `campaigns`; id field `id`; id sent as query parameter `campaign_id`; stamps `campaign_id`.
- `canvas_details`: GET `/canvas/details` - records path `.`; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 0; page size 100; computed output fields `tags`,
  `teams`; fan-out; ids from request `/canvas/list`; id-list records path `canvases`; id field `id`;
  id sent as query parameter `canvas_id`; stamps `canvas_id`.
- `canvas_data_summary`: GET `/canvas/data_summary` - records path `.`; page-number pagination; page
  parameter `page`; no page-size parameter; starts at 0; page size 100; fan-out; ids from request
  `/canvas/list`; id-list records path `canvases`; id field `id`; id sent as query parameter
  `canvas_id`; stamps `canvas_id`.
- `segment_details`: GET `/segments/details` - records path `.`; page-number pagination; page
  parameter `page`; no page-size parameter; starts at 0; page size 100; fan-out; ids from request
  `/segments/list`; id-list records path `segments`; id field `id`; id sent as query parameter
  `segment_id`; stamps `segment_id`.
- `catalogs`: GET `/catalogs` - records path `catalogs`; computed output fields `name`.
- `content_blocks`: GET `/content_blocks/list` - records path `content_blocks`; query
  `modified_after` from template `{{ incremental.lower_bound }}`, omitted when absent; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100; incremental cursor
  `last_edited`; formatted as `rfc3339`.
- `email_templates`: GET `/templates/email/list` - records path `templates`; query `modified_after`
  from template `{{ incremental.lower_bound }}`, omitted when absent; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100; incremental cursor
  `updated_at`; formatted as `rfc3339`; computed output fields `tags`.
- `feed_cards`: GET `/feed/list` - records path `cards`; page-number pagination; page parameter
  `page`; no page-size parameter; starts at 0; page size 100.
- `email_hard_bounces`: GET `/email/hard_bounces` - records path `emails`; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100; incremental cursor
  `hard_bounced_at`; sent as `start_date`; formatted as YYYY-MM-DD date.
- `email_unsubscribes`: GET `/email/unsubscribes` - records path `emails`; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100; incremental cursor
  `unsubscribed_at`; sent as `start_date`; formatted as YYYY-MM-DD date.
- `sms_invalid_phone_numbers`: GET `/sms/invalid_phone_numbers` - records path `sms`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100; incremental cursor
  `invalid_detected_at`; sent as `start_date`; formatted as YYYY-MM-DD date.
- `kpi_dau`: GET `/kpi/dau/data_series` - records path `data`; query `length`=`14`.
- `kpi_mau`: GET `/kpi/mau/data_series` - records path `data`; query `length`=`14`.
- `kpi_new_users`: GET `/kpi/new_users/data_series` - records path `data`; query `length`=`14`.
- `kpi_uninstalls`: GET `/kpi/uninstalls/data_series` - records path `data`; query `length`=`14`.
- `sessions`: GET `/sessions/data_series` - records path `data`; query `length`=`14`.
- `preference_centers`: GET `/preference_center/v1/list` - records path `preference_centers`;
  incremental cursor `updated_at`; formatted as `rfc3339`.
- `scheduled_broadcasts`: GET `/messages/scheduled_broadcasts` - records path `.`; computed output
  fields `tags`.

## Write actions & risks

Overall write risk: external mutation of Braze user profiles (track/identify/merge/delete/alias),
subscription-group membership, catalogs and catalog items, content blocks, email templates,
email/SMS compliance lists, preference centers, and live message/campaign/Canvas sends;
send_message/trigger_campaign_send/trigger_canvas_send dispatch real, irreversible communications to
end users and delete_users/remove_user_external_ids/create_email_blocklist are destructive - every
write ships with an explicit per-action risk string.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `track_users`: POST `/users/track` - kind `custom`; body type `json`; accepted fields
  `attributes`, `events`, `purchases`; risk: records user attribute/event/purchase data in bulk (up
  to 75 events/attributes/purchases per request per Braze's documented limit); an external user_id
  or user_alias in the record is the ONLY way Braze correlates the update to a user, so a
  mistargeted identifier silently attaches data to the wrong (or a newly-created) user profile.
- `identify_users`: POST `/users/identify` - kind `custom`; body type `json`; accepted fields
  `aliases_to_identify`, `emails_to_identify`, `merge_behavior`, `phone_numbers_to_identify`; risk:
  converts an anonymous/aliased user profile into an identified one and can merge attribute/behavior
  history onto the target identified profile; merge_behavior: merge combines the two profiles' full
  history irreversibly.
- `delete_users`: POST `/users/delete` - kind `delete`; body type `json`; accepted fields
  `braze_ids`, `external_ids`, `user_aliases`; confirmation `destructive`; risk: permanently and
  irreversibly deletes user profiles and all their associated data (attributes, event/purchase
  history, message engagement); Braze does not offer an undelete path.
- `merge_users`: POST `/users/merge` - kind `update`; body type `json`; required record fields
  `merge_updates`; accepted fields `merge_updates`; risk: irreversibly merges one user profile's
  full history into another and deletes the source profile identifier; up to 50 merge pairs per
  request per Braze's documented limit.
- `create_user_alias`: POST `/users/alias/new` - kind `create`; body type `json`; required record
  fields `user_aliases`; accepted fields `user_aliases`; risk: creates new alias identifiers for
  existing (or new anonymous) user profiles; low-risk additive mutation, no approval required.
- `update_user_alias`: POST `/users/alias/update` - kind `update`; body type `json`; required record
  fields `alias_updates`; accepted fields `alias_updates`; risk: renames an existing alias
  identifier on a user profile; any external system correlating users by the old alias_name stops
  matching after this runs.
- `remove_user_external_ids`: POST `/users/external_ids/remove` - kind `delete`; body type `json`;
  required record fields `external_ids`; accepted fields `external_ids`; confirmation `destructive`;
  risk: detaches an external_id from its user profile, converting that profile to anonymous; the
  profile itself is not deleted but becomes unreachable by the removed identifier.
- `rename_user_external_ids`: POST `/users/external_ids/rename` - kind `update`; body type `json`;
  required record fields `external_id_renames`; accepted fields `external_id_renames`; risk: renames
  a user's external_id; any external system correlating users by the old id stops matching after
  this runs.
- `set_subscription_status_v2`: POST `/v2/subscription/status/set` - kind `update`; body type
  `json`; required record fields `subscription_groups`; accepted fields `subscription_groups`; risk:
  opts users into or out of an email/SMS subscription group in bulk (up to 50 groups x 50
  identifiers per Braze's documented limit); setting subscription_state to unsubscribed on a
  transactional-adjacent group can stop legally-required or expected communications reaching those
  users.
- `create_catalog`: POST `/catalogs` - kind `create`; body type `json`; required record fields
  `catalogs`; accepted fields `catalogs`; risk: creates a new catalog container with a fixed field
  schema; low-risk additive mutation, no approval required.
- `delete_catalog`: DELETE `/catalogs/{{ record.catalog_name }}` - kind `delete`; body type `none`;
  path fields `catalog_name`; required record fields `catalog_name`; accepted fields `catalog_name`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: permanently
  deletes a catalog and every item it contains; any campaign/Canvas/Connected Content template
  referencing this catalog by name starts failing to resolve.
- `create_catalog_items`: POST `/catalogs/{{ record.catalog_name }}/items` - kind `create`; body
  type `json`; path fields `catalog_name`; body fields `items`; required record fields
  `catalog_name`, `items`; accepted fields `catalog_name`, `items`; risk: adds new rows (up to 50
  per request per Braze's documented limit) to an existing catalog; low-risk additive mutation, no
  approval required.
- `update_catalog_items`: PATCH `/catalogs/{{ record.catalog_name }}/items` - kind `update`; body
  type `json`; path fields `catalog_name`; body fields `items`; required record fields
  `catalog_name`, `items`; accepted fields `catalog_name`, `items`; risk: partially updates existing
  catalog rows in bulk by their id field; any Connected Content template or campaign personalization
  reading this catalog reflects the new values on its next fetch.
- `update_catalog_item`: PATCH `/catalogs/{{ record.catalog_name }}/items/{{ record.item_id }}` -
  kind `update`; body type `json`; path fields `catalog_name`, `item_id`; body fields `items`;
  required record fields `catalog_name`, `item_id`, `items`; accepted fields `catalog_name`,
  `item_id`, `items`; risk: partially updates a single existing catalog row; any Connected Content
  template or campaign personalization reading this catalog reflects the new value on its next
  fetch.
- `delete_catalog_item`: DELETE `/catalogs/{{ record.catalog_name }}/items/{{ record.item_id }}` -
  kind `delete`; body type `none`; path fields `catalog_name`, `item_id`; required record fields
  `catalog_name`, `item_id`; accepted fields `catalog_name`, `item_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: permanently removes a single row from
  a catalog; any Connected Content template or campaign personalization referencing this item_id
  starts returning no match.
- `create_content_block`: POST `/content_blocks/create` - kind `create`; body type `json`; required
  record fields `name`, `content`; accepted fields `content`, `description`, `name`, `state`,
  `tags`; risk: creates a new reusable email Content Block; low-risk additive mutation, no approval
  required.
- `update_content_block`: POST `/content_blocks/update` - kind `update`; body type `json`; required
  record fields `content_block_id`; accepted fields `content`, `content_block_id`, `description`,
  `name`, `state`, `tags`; risk: mutates an existing Content Block's markup/text; changes are
  reflected in EVERY campaign/Canvas/template that includes this block on their next send, including
  already-scheduled sends.
- `create_email_template`: POST `/templates/email/create` - kind `create`; body type `json`;
  required record fields `template_name`, `subject`, `body`; accepted fields `body`,
  `plaintext_body`, `preheader`, `should_inline_css`, `subject`, `tags`, `template_name`; risk:
  creates a new reusable email template; low-risk additive mutation, no approval required.
- `update_email_template`: POST `/templates/email/update` - kind `update`; body type `json`;
  required record fields `email_template_id`; accepted fields `body`, `email_template_id`,
  `plaintext_body`, `preheader`, `should_inline_css`, `subject`, `tags`, `template_name`; risk:
  mutates an existing email template's subject/body; changes are reflected in EVERY campaign using
  this template on its next send, including already-scheduled sends.
- `create_email_blocklist`: POST `/email/blocklist` - kind `create`; body type `json`; required
  record fields `email`; accepted fields `email`; confirmation `destructive`; risk: permanently
  blocklists email addresses from ever receiving Braze email again for this workspace; Braze's own
  docs note blocklisting cannot be undone via the API (requires a support request to reverse).
- `remove_email_hard_bounce`: POST `/email/bounce/remove` - kind `delete`; body type `json`;
  required record fields `email`; accepted fields `email`; risk: clears an email address's
  hard-bounced status, allowing future sends to resume; use only after confirming the underlying
  delivery issue is actually resolved, or the address will likely hard-bounce again and harm sender
  reputation.
- `remove_email_spam`: POST `/email/spam/remove` - kind `delete`; body type `json`; required record
  fields `email`; accepted fields `email`; risk: clears an email address's spam-complaint status,
  allowing future sends to resume; reversing a genuine spam complaint risks another complaint and
  further sender-reputation damage.
- `set_email_subscription_status`: POST `/email/status` - kind `update`; body type `json`; required
  record fields `email`, `subscription_state`; accepted fields `email`, `subscription_state`; risk:
  changes a single email address's global subscription state (subscribed/unsubscribed/opted_in);
  setting unsubscribed stops all future non-transactional email to that address.
- `remove_sms_invalid_phone_numbers`: POST `/sms/invalid_phone_numbers/remove` - kind `delete`; body
  type `json`; required record fields `phone_numbers`; accepted fields `phone_numbers`; risk: clears
  the invalid-number flag for phone numbers, allowing future SMS/MMS sends to resume; use only after
  confirming the number can actually receive messages again, or it will likely be re-flagged and
  waste sending budget.
- `create_preference_center`: POST `/preference_center/v1` - kind `create`; body type `json`;
  required record fields `name`, `preference_center_title`, `preference_center_page_html`,
  `confirmation_page_html`; accepted fields `confirmation_page_html`, `name`, `options`,
  `preference_center_page_html`, `preference_center_title`, `state`; risk: publishes a new
  customer-facing preference center page (a live, externally-reachable URL once active); low-risk
  additive mutation but review the submitted HTML before use since it is served to end users
  verbatim.
- `update_preference_center`: PUT `/preference_center/v1/{{ record.preference_center_external_id }}`
  - kind `update`; body type `json`; path fields `preference_center_external_id`; required record
  fields `preference_center_external_id`, `name`, `preference_center_title`,
  `preference_center_page_html`, `confirmation_page_html`; accepted fields `confirmation_page_html`,
  `name`, `options`, `preference_center_external_id`, `preference_center_page_html`,
  `preference_center_title`; risk: overwrites an already-live, externally-reachable preference
  center page's HTML/title; visible to any end user who visits the page immediately after this runs.
- `send_message`: POST `/messages/send` - kind `custom`; body type `json`; required record fields
  `messages`; accepted fields `audience`, `broadcast`, `campaign_id`, `external_user_ids`,
  `messages`, `override_frequency_capping`, `recipient_subscription_state`, `segment_id`, `send_id`,
  `user_aliases`; confirmation `destructive`; risk: immediately sends a live message
  (push/email/SMS/webhook/Content Card) to the specified users, segment, or broadcast audience;
  irreversible once dispatched and the single riskiest write this connector exposes - always confirm
  the audience scope (segment_id/audience filter vs. an explicit small external_user_ids list)
  before use.
- `trigger_campaign_send`: POST `/campaigns/trigger/send` - kind `custom`; body type `json`;
  required record fields `campaign_id`; accepted fields `audience`, `broadcast`, `campaign_id`,
  `recipients`, `send_id`, `trigger_properties`; confirmation `destructive`; risk: immediately
  dispatches an existing API-triggered campaign to the specified recipients/audience; irreversible
  once dispatched, always confirm the recipients/audience scope before use.
- `trigger_canvas_send`: POST `/canvas/trigger/send` - kind `custom`; body type `json`; required
  record fields `canvas_id`; accepted fields `audience`, `broadcast`, `canvas_entry_properties`,
  `canvas_id`, `recipients`; confirmation `destructive`; risk: immediately enters the specified
  recipients/audience into an existing API-triggered Canvas; irreversible once dispatched, always
  confirm the recipients/audience scope before use.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 21 stream-backed endpoint group(s), 29 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=17, non_data_endpoint=2, out_of_scope=17, requires_elevated_scope=9.
