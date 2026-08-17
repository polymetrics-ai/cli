# pm connectors inspect braze

```text
NAME
  pm connectors inspect braze - Braze connector manual

SYNOPSIS
  pm connectors inspect braze
  pm connectors inspect braze --json
  pm credentials add <name> --connector braze [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Braze campaigns, Canvases, segments (list + per-id details/analytics-summary), catalogs, content blocks, email templates, Content Cards, email bounce/unsubscribe lists, SMS invalid-number lists, KPIs, sessions, preference centers, and scheduled broadcasts; writes user data (track/identify/merge/alias/delete), subscription-group status, catalog and catalog-item mutations, content block/email template mutations, email/SMS compliance-list mutations, preference center mutations, and live message/campaign/Canvas sends through the Braze REST API. The events (custom event names) and purchases/product_list streams are not modeled by this bundle; see docs.md Known limits.

ICON
  asset: icons/braze.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://www.braze.com/docs/api/home

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  api_key (secret)

ETL STREAMS
  campaigns:
    primary key: id
    cursor: last_edited
    fields: id(), is_api_campaign(), last_edited(), name(), tags()
  canvases:
    primary key: id
    cursor: last_edited
    fields: id(), last_edited(), name(), tags()
  segments:
    primary key: id
    fields: analytics_tracking_enabled(), id(), name(), tags()
  campaign_details:
    primary key: campaign_id
    fields: archived(), campaign_id(), channels(), conversion_behaviors(), created_at(), description(), draft(), enabled(), first_sent(), has_post_launch_draft(), last_sent(), message(), messages(), name(), schedule_type(), tags(), teams(), updated_at()
  canvas_details:
    primary key: canvas_id
    fields: archived(), canvas_id(), channels(), created_at(), description(), draft(), enabled(), first_entry(), last_entry(), message(), name(), schedule_type(), steps(), tags(), teams(), updated_at(), variants()
  canvas_data_summary:
    primary key: canvas_id
    fields: canvas_id(), data(), message(), name()
  segment_details:
    primary key: segment_id
    fields: created_at(), description(), message(), name(), segment_id(), tags(), text_description(), updated_at()
  catalogs:
    primary key: name
    fields: description(), fields(), name(), num_items(), updated_at()
  content_blocks:
    primary key: content_block_id
    cursor: last_edited
    fields: content_block_id(), content_type(), created_at(), inclusion_count(), last_edited(), liquid_tag(), name(), tags()
  email_templates:
    primary key: email_template_id
    cursor: updated_at
    fields: created_at(), email_template_id(), tags(), template_name(), updated_at()
  feed_cards:
    primary key: id
    fields: id(), tags(), title(), type()
  email_hard_bounces:
    primary key: email, hard_bounced_at
    cursor: hard_bounced_at
    fields: email(), hard_bounced_at()
  email_unsubscribes:
    primary key: email, unsubscribed_at
    cursor: unsubscribed_at
    fields: email(), unsubscribed_at()
  sms_invalid_phone_numbers:
    primary key: phone, invalid_detected_at
    cursor: invalid_detected_at
    fields: invalid_detected_at(), phone(), reason()
  kpi_dau:
    primary key: time
    cursor: time
    fields: dau(), time()
  kpi_mau:
    primary key: time
    cursor: time
    fields: mau(), time()
  kpi_new_users:
    primary key: time
    cursor: time
    fields: new_users(), time()
  kpi_uninstalls:
    primary key: time
    cursor: time
    fields: time(), uninstalls()
  sessions:
    primary key: time
    cursor: time
    fields: sessions(), time()
  preference_centers:
    primary key: preference_center_api_id
    cursor: updated_at
    fields: created_at(), name(), preference_center_api_id(), updated_at()
  scheduled_broadcasts:
    primary key: id, next_send_time
    fields: id(), name(), next_send_time(), schedule_type(), tags(), type()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  track_users:
    endpoint: POST /users/track
    risk: records user attribute/event/purchase data in bulk (up to 75 events/attributes/purchases per request per Braze's documented limit); an external user_id or user_alias in the record is the ONLY way Braze correlates the update to a user, so a mistargeted identifier silently attaches data to the wrong (or a newly-created) user profile
  identify_users:
    endpoint: POST /users/identify
    risk: converts an anonymous/aliased user profile into an identified one and can merge attribute/behavior history onto the target identified profile; merge_behavior: merge combines the two profiles' full history irreversibly
  delete_users:
    endpoint: POST /users/delete
    risk: permanently and irreversibly deletes user profiles and all their associated data (attributes, event/purchase history, message engagement); Braze does not offer an undelete path
  merge_users:
    endpoint: POST /users/merge
    risk: irreversibly merges one user profile's full history into another and deletes the source profile identifier; up to 50 merge pairs per request per Braze's documented limit
  create_user_alias:
    endpoint: POST /users/alias/new
    risk: creates new alias identifiers for existing (or new anonymous) user profiles; low-risk additive mutation, no approval required
  update_user_alias:
    endpoint: POST /users/alias/update
    risk: renames an existing alias identifier on a user profile; any external system correlating users by the old alias_name stops matching after this runs
  remove_user_external_ids:
    endpoint: POST /users/external_ids/remove
    risk: detaches an external_id from its user profile, converting that profile to anonymous; the profile itself is not deleted but becomes unreachable by the removed identifier
  rename_user_external_ids:
    endpoint: POST /users/external_ids/rename
    risk: renames a user's external_id; any external system correlating users by the old id stops matching after this runs
  set_subscription_status_v2:
    endpoint: POST /v2/subscription/status/set
    risk: opts users into or out of an email/SMS subscription group in bulk (up to 50 groups x 50 identifiers per Braze's documented limit); setting subscription_state to unsubscribed on a transactional-adjacent group can stop legally-required or expected communications reaching those users
  create_catalog:
    endpoint: POST /catalogs
    risk: creates a new catalog container with a fixed field schema; low-risk additive mutation, no approval required
  delete_catalog:
    endpoint: DELETE /catalogs/{{ record.catalog_name }}
    required fields: catalog_name
    risk: permanently deletes a catalog and every item it contains; any campaign/Canvas/Connected Content template referencing this catalog by name starts failing to resolve
  create_catalog_items:
    endpoint: POST /catalogs/{{ record.catalog_name }}/items
    required fields: catalog_name
    optional fields: items
    risk: adds new rows (up to 50 per request per Braze's documented limit) to an existing catalog; low-risk additive mutation, no approval required
  update_catalog_items:
    endpoint: PATCH /catalogs/{{ record.catalog_name }}/items
    required fields: catalog_name
    optional fields: items
    risk: partially updates existing catalog rows in bulk by their id field; any Connected Content template or campaign personalization reading this catalog reflects the new values on its next fetch
  update_catalog_item:
    endpoint: PATCH /catalogs/{{ record.catalog_name }}/items/{{ record.item_id }}
    required fields: catalog_name, item_id
    optional fields: items
    risk: partially updates a single existing catalog row; any Connected Content template or campaign personalization reading this catalog reflects the new value on its next fetch
  delete_catalog_item:
    endpoint: DELETE /catalogs/{{ record.catalog_name }}/items/{{ record.item_id }}
    required fields: catalog_name, item_id
    risk: permanently removes a single row from a catalog; any Connected Content template or campaign personalization referencing this item_id starts returning no match
  create_content_block:
    endpoint: POST /content_blocks/create
    risk: creates a new reusable email Content Block; low-risk additive mutation, no approval required
  update_content_block:
    endpoint: POST /content_blocks/update
    risk: mutates an existing Content Block's markup/text; changes are reflected in EVERY campaign/Canvas/template that includes this block on their next send, including already-scheduled sends
  create_email_template:
    endpoint: POST /templates/email/create
    risk: creates a new reusable email template; low-risk additive mutation, no approval required
  update_email_template:
    endpoint: POST /templates/email/update
    risk: mutates an existing email template's subject/body; changes are reflected in EVERY campaign using this template on its next send, including already-scheduled sends
  create_email_blocklist:
    endpoint: POST /email/blocklist
    risk: permanently blocklists email addresses from ever receiving Braze email again for this workspace; Braze's own docs note blocklisting cannot be undone via the API (requires a support request to reverse)
  remove_email_hard_bounce:
    endpoint: POST /email/bounce/remove
    risk: clears an email address's hard-bounced status, allowing future sends to resume; use only after confirming the underlying delivery issue is actually resolved, or the address will likely hard-bounce again and harm sender reputation
  remove_email_spam:
    endpoint: POST /email/spam/remove
    risk: clears an email address's spam-complaint status, allowing future sends to resume; reversing a genuine spam complaint risks another complaint and further sender-reputation damage
  set_email_subscription_status:
    endpoint: POST /email/status
    risk: changes a single email address's global subscription state (subscribed/unsubscribed/opted_in); setting unsubscribed stops all future non-transactional email to that address
  remove_sms_invalid_phone_numbers:
    endpoint: POST /sms/invalid_phone_numbers/remove
    risk: clears the invalid-number flag for phone numbers, allowing future SMS/MMS sends to resume; use only after confirming the number can actually receive messages again, or it will likely be re-flagged and waste sending budget
  create_preference_center:
    endpoint: POST /preference_center/v1
    risk: publishes a new customer-facing preference center page (a live, externally-reachable URL once active); low-risk additive mutation but review the submitted HTML before use since it is served to end users verbatim
  update_preference_center:
    endpoint: PUT /preference_center/v1/{{ record.preference_center_external_id }}
    required fields: preference_center_external_id
    risk: overwrites an already-live, externally-reachable preference center page's HTML/title; visible to any end user who visits the page immediately after this runs
  send_message:
    endpoint: POST /messages/send
    risk: immediately sends a live message (push/email/SMS/webhook/Content Card) to the specified users, segment, or broadcast audience; irreversible once dispatched and the single riskiest write this connector exposes — always confirm the audience scope (segment_id/audience filter vs. an explicit small external_user_ids list) before use
  trigger_campaign_send:
    endpoint: POST /campaigns/trigger/send
    risk: immediately dispatches an existing API-triggered campaign to the specified recipients/audience; irreversible once dispatched, always confirm the recipients/audience scope before use
  trigger_canvas_send:
    endpoint: POST /canvas/trigger/send
    risk: immediately enters the specified recipients/audience into an existing API-triggered Canvas; irreversible once dispatched, always confirm the recipients/audience scope before use

SECURITY
  read risk: external Braze API read of campaign, Canvas, segment, catalog, content-block, template, Content-Card, email-compliance, SMS-compliance, KPI/session, and preference-center metadata
  write risk: external mutation of Braze user profiles (track/identify/merge/delete/alias), subscription-group membership, catalogs and catalog items, content blocks, email templates, email/SMS compliance lists, preference centers, and live message/campaign/Canvas sends; send_message/trigger_campaign_send/trigger_canvas_send dispatch real, irreversible communications to end users and delete_users/remove_user_external_ids/create_email_blocklist are destructive — every write ships with an explicit per-action risk string
  approval: required for destructive actions (delete_users, remove_user_external_ids, delete_catalog, delete_catalog_item, create_email_blocklist) and for any live send (send_message, trigger_campaign_send, trigger_canvas_send); catalog/content-block/template/preference-center create-update and compliance-list removals are lower-risk but still externally visible
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect braze

  # Inspect as structured JSON
  pm connectors inspect braze --json

AGENT WORKFLOW
  - Run pm connectors inspect braze before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
