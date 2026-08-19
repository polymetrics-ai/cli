---
name: pm-braze
description: Braze connector knowledge and safe action guide.
---

# pm-braze

## Purpose

Reads Braze campaigns, Canvases, segments (list + per-id details/analytics-summary), catalogs, content blocks, email templates, Content Cards, email bounce/unsubscribe lists, SMS invalid-number lists, KPIs, sessions, preference centers, and scheduled broadcasts; writes user data (track/identify/merge/alias/delete), subscription-group status, catalog and catalog-item mutations, content block/email template mutations, email/SMS compliance-list mutations, preference center mutations, and live message/campaign/Canvas sends through the Braze REST API. The events (custom event names) and purchases/product_list streams are not modeled by this bundle; see docs.md Known limits.

## Icon

- id: braze
- asset: icons/braze.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://www.braze.com/docs/api/home

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url (required)
- mode
- api_key (secret) (required)

## ETL Streams

- campaigns:
  - primary key: id
  - cursor: last_edited
  - fields: id(string), is_api_campaign(boolean), last_edited(string), name(string), tags(array)
- canvases:
  - primary key: id
  - cursor: last_edited
  - fields: id(string), last_edited(string), name(string), tags(array)
- segments:
  - primary key: id
  - fields: analytics_tracking_enabled(boolean), id(string), name(string), tags(array)
- campaign_details:
  - primary key: campaign_id
  - fields: archived(boolean), campaign_id(string), channels(string), conversion_behaviors(array), created_at(string), description(string), draft(boolean), enabled(boolean), first_sent(string), has_post_launch_draft(boolean), last_sent(string), message(string), messages(object), name(string), schedule_type(string), tags(string), teams(string), updated_at(string)
- canvas_details:
  - primary key: canvas_id
  - fields: archived(boolean), canvas_id(string), channels(array), created_at(string), description(string), draft(boolean), enabled(boolean), first_entry(string), last_entry(string), message(string), name(string), schedule_type(string), steps(array), tags(string), teams(string), updated_at(string), variants(array)
- canvas_data_summary:
  - primary key: canvas_id
  - fields: canvas_id(string), data(array), message(string), name(string)
- segment_details:
  - primary key: segment_id
  - fields: created_at(string), description(string), message(string), name(string), segment_id(string), tags(array), text_description(string), updated_at(string)
- catalogs:
  - primary key: name
  - fields: description(string), fields(array), name(string), num_items(integer), updated_at(string)
- content_blocks:
  - primary key: content_block_id
  - cursor: last_edited
  - fields: content_block_id(string), content_type(string), created_at(string), inclusion_count(integer), last_edited(string), liquid_tag(string), name(string), tags(array)
- email_templates:
  - primary key: email_template_id
  - cursor: updated_at
  - fields: created_at(string), email_template_id(string), tags(string), template_name(string), updated_at(string)
- feed_cards:
  - primary key: id
  - fields: id(string), tags(array), title(string), type(string)
- email_hard_bounces:
  - primary key: email, hard_bounced_at
  - cursor: hard_bounced_at
  - fields: email(string), hard_bounced_at(string)
- email_unsubscribes:
  - primary key: email, unsubscribed_at
  - cursor: unsubscribed_at
  - fields: email(string), unsubscribed_at(string)
- sms_invalid_phone_numbers:
  - primary key: phone, invalid_detected_at
  - cursor: invalid_detected_at
  - fields: invalid_detected_at(string), phone(string), reason(string)
- kpi_dau:
  - primary key: time
  - cursor: time
  - fields: dau(integer), time(string)
- kpi_mau:
  - primary key: time
  - cursor: time
  - fields: mau(integer), time(string)
- kpi_new_users:
  - primary key: time
  - cursor: time
  - fields: new_users(integer), time(string)
- kpi_uninstalls:
  - primary key: time
  - cursor: time
  - fields: time(string), uninstalls(integer)
- sessions:
  - primary key: time
  - cursor: time
  - fields: sessions(integer), time(string)
- preference_centers:
  - primary key: preference_center_api_id
  - cursor: updated_at
  - fields: created_at(string), name(string), preference_center_api_id(string), updated_at(string)
- scheduled_broadcasts:
  - primary key: id, next_send_time
  - fields: id(string), name(string), next_send_time(string), schedule_type(string), tags(string), type(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- track_users:
  - endpoint: POST /users/track
  - risk: records user attribute/event/purchase data in bulk (up to 75 events/attributes/purchases per request per Braze's documented limit); an external user_id or user_alias in the record is the ONLY way Braze correlates the update to a user, so a mistargeted identifier silently attaches data to the wrong (or a newly-created) user profile
- identify_users:
  - endpoint: POST /users/identify
  - risk: converts an anonymous/aliased user profile into an identified one and can merge attribute/behavior history onto the target identified profile; merge_behavior: merge combines the two profiles' full history irreversibly
- delete_users:
  - endpoint: POST /users/delete
  - risk: permanently and irreversibly deletes user profiles and all their associated data (attributes, event/purchase history, message engagement); Braze does not offer an undelete path
- merge_users:
  - endpoint: POST /users/merge
  - required fields: merge_updates
  - risk: irreversibly merges one user profile's full history into another and deletes the source profile identifier; up to 50 merge pairs per request per Braze's documented limit
- create_user_alias:
  - endpoint: POST /users/alias/new
  - required fields: user_aliases
  - risk: creates new alias identifiers for existing (or new anonymous) user profiles; low-risk additive mutation, no approval required
- update_user_alias:
  - endpoint: POST /users/alias/update
  - required fields: alias_updates
  - risk: renames an existing alias identifier on a user profile; any external system correlating users by the old alias_name stops matching after this runs
- remove_user_external_ids:
  - endpoint: POST /users/external_ids/remove
  - required fields: external_ids
  - risk: detaches an external_id from its user profile, converting that profile to anonymous; the profile itself is not deleted but becomes unreachable by the removed identifier
- rename_user_external_ids:
  - endpoint: POST /users/external_ids/rename
  - required fields: external_id_renames
  - risk: renames a user's external_id; any external system correlating users by the old id stops matching after this runs
- set_subscription_status_v2:
  - endpoint: POST /v2/subscription/status/set
  - required fields: subscription_groups
  - risk: opts users into or out of an email/SMS subscription group in bulk (up to 50 groups x 50 identifiers per Braze's documented limit); setting subscription_state to unsubscribed on a transactional-adjacent group can stop legally-required or expected communications reaching those users
- create_catalog:
  - endpoint: POST /catalogs
  - required fields: catalogs
  - risk: creates a new catalog container with a fixed field schema; low-risk additive mutation, no approval required
- delete_catalog:
  - endpoint: DELETE /catalogs/{{ record.catalog_name }}
  - required fields: catalog_name
  - risk: permanently deletes a catalog and every item it contains; any campaign/Canvas/Connected Content template referencing this catalog by name starts failing to resolve
- create_catalog_items:
  - endpoint: POST /catalogs/{{ record.catalog_name }}/items
  - required fields: catalog_name, items
  - risk: adds new rows (up to 50 per request per Braze's documented limit) to an existing catalog; low-risk additive mutation, no approval required
- update_catalog_items:
  - endpoint: PATCH /catalogs/{{ record.catalog_name }}/items
  - required fields: catalog_name, items
  - risk: partially updates existing catalog rows in bulk by their id field; any Connected Content template or campaign personalization reading this catalog reflects the new values on its next fetch
- update_catalog_item:
  - endpoint: PATCH /catalogs/{{ record.catalog_name }}/items/{{ record.item_id }}
  - required fields: catalog_name, item_id, items
  - risk: partially updates a single existing catalog row; any Connected Content template or campaign personalization reading this catalog reflects the new value on its next fetch
- delete_catalog_item:
  - endpoint: DELETE /catalogs/{{ record.catalog_name }}/items/{{ record.item_id }}
  - required fields: catalog_name, item_id
  - risk: permanently removes a single row from a catalog; any Connected Content template or campaign personalization referencing this item_id starts returning no match
- create_content_block:
  - endpoint: POST /content_blocks/create
  - required fields: name, content
  - risk: creates a new reusable email Content Block; low-risk additive mutation, no approval required
- update_content_block:
  - endpoint: POST /content_blocks/update
  - required fields: content_block_id
  - risk: mutates an existing Content Block's markup/text; changes are reflected in EVERY campaign/Canvas/template that includes this block on their next send, including already-scheduled sends
- create_email_template:
  - endpoint: POST /templates/email/create
  - required fields: template_name, subject, body
  - risk: creates a new reusable email template; low-risk additive mutation, no approval required
- update_email_template:
  - endpoint: POST /templates/email/update
  - required fields: email_template_id
  - risk: mutates an existing email template's subject/body; changes are reflected in EVERY campaign using this template on its next send, including already-scheduled sends
- create_email_blocklist:
  - endpoint: POST /email/blocklist
  - required fields: email
  - risk: permanently blocklists email addresses from ever receiving Braze email again for this workspace; Braze's own docs note blocklisting cannot be undone via the API (requires a support request to reverse)
- remove_email_hard_bounce:
  - endpoint: POST /email/bounce/remove
  - required fields: email
  - risk: clears an email address's hard-bounced status, allowing future sends to resume; use only after confirming the underlying delivery issue is actually resolved, or the address will likely hard-bounce again and harm sender reputation
- remove_email_spam:
  - endpoint: POST /email/spam/remove
  - required fields: email
  - risk: clears an email address's spam-complaint status, allowing future sends to resume; reversing a genuine spam complaint risks another complaint and further sender-reputation damage
- set_email_subscription_status:
  - endpoint: POST /email/status
  - required fields: email, subscription_state
  - risk: changes a single email address's global subscription state (subscribed/unsubscribed/opted_in); setting unsubscribed stops all future non-transactional email to that address
- remove_sms_invalid_phone_numbers:
  - endpoint: POST /sms/invalid_phone_numbers/remove
  - required fields: phone_numbers
  - risk: clears the invalid-number flag for phone numbers, allowing future SMS/MMS sends to resume; use only after confirming the number can actually receive messages again, or it will likely be re-flagged and waste sending budget
- create_preference_center:
  - endpoint: POST /preference_center/v1
  - required fields: name, preference_center_title, preference_center_page_html, confirmation_page_html
  - risk: publishes a new customer-facing preference center page (a live, externally-reachable URL once active); low-risk additive mutation but review the submitted HTML before use since it is served to end users verbatim
- update_preference_center:
  - endpoint: PUT /preference_center/v1/{{ record.preference_center_external_id }}
  - required fields: preference_center_external_id, name, preference_center_title, preference_center_page_html, confirmation_page_html
  - risk: overwrites an already-live, externally-reachable preference center page's HTML/title; visible to any end user who visits the page immediately after this runs
- send_message:
  - endpoint: POST /messages/send
  - required fields: messages
  - risk: immediately sends a live message (push/email/SMS/webhook/Content Card) to the specified users, segment, or broadcast audience; irreversible once dispatched and the single riskiest write this connector exposes — always confirm the audience scope (segment_id/audience filter vs. an explicit small external_user_ids list) before use
- trigger_campaign_send:
  - endpoint: POST /campaigns/trigger/send
  - required fields: campaign_id
  - risk: immediately dispatches an existing API-triggered campaign to the specified recipients/audience; irreversible once dispatched, always confirm the recipients/audience scope before use
- trigger_canvas_send:
  - endpoint: POST /canvas/trigger/send
  - required fields: canvas_id
  - risk: immediately enters the specified recipients/audience into an existing API-triggered Canvas; irreversible once dispatched, always confirm the recipients/audience scope before use

## Security

- read risk: external Braze API read of campaign, Canvas, segment, catalog, content-block, template, Content-Card, email-compliance, SMS-compliance, KPI/session, and preference-center metadata
- write risk: external mutation of Braze user profiles (track/identify/merge/delete/alias), subscription-group membership, catalogs and catalog items, content blocks, email templates, email/SMS compliance lists, preference centers, and live message/campaign/Canvas sends; send_message/trigger_campaign_send/trigger_canvas_send dispatch real, irreversible communications to end users and delete_users/remove_user_external_ids/create_email_blocklist are destructive — every write ships with an explicit per-action risk string
- approval: required for destructive actions (delete_users, remove_user_external_ids, delete_catalog, delete_catalog_item, create_email_blocklist) and for any live send (send_message, trigger_campaign_send, trigger_canvas_send); catalog/content-block/template/preference-center create-update and compliance-list removals are lower-risk but still externally visible
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect braze
```

### Inspect as structured JSON

```bash
pm connectors inspect braze --json
```

## Agent Rules

- Run pm connectors inspect braze before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
