# pm connectors inspect pipedrive

```text
NAME
  pm connectors inspect pipedrive - Pipedrive connector manual

SYNOPSIS
  pm connectors inspect pipedrive
  pm connectors inspect pipedrive --json
  pm credentials add <name> --connector pipedrive [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Pipedrive deals, persons, organizations, activities, products, users, notes, leads, saved filters, activity types, roles, webhooks, and field/reference metadata, and writes lead/note/filter/activity-type/lead-label/webhook mutations through REST API v1.

ICON
  id: pipedrive
  asset: icons/pipedrive.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.pipedrive.com/docs/api/v1

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  replication_start_date
  api_token (secret) (required)

ETL STREAMS
  deals:
    primary key: id
    cursor: update_time
    fields: add_time(string), currency(string), id(integer), org_id(integer), person_id(integer), stage_id(integer), status(string), title(string), update_time(string), value(number)
  persons:
    primary key: id
    cursor: update_time
    fields: add_time(string), email(array), id(integer), name(string), org_id(integer), owner_id(integer), phone(array), update_time(string)
  organizations:
    primary key: id
    cursor: update_time
    fields: add_time(string), id(integer), name(string), owner_id(integer), people_count(integer), update_time(string)
  activities:
    primary key: id
    cursor: update_time
    fields: add_time(string), deal_id(integer), done(boolean), due_date(string), id(integer), org_id(integer), person_id(integer), subject(string), type(string), update_time(string)
  products:
    primary key: id
    fields: active_flag(boolean), add_time(string), code(string), id(integer), name(string), owner_id(integer), unit(string), update_time(string)
  users:
    primary key: id
    fields: active_flag(boolean), created(string), email(string), id(integer), is_admin(integer), modified(string), name(string)
  notes:
    primary key: id
    cursor: update_time
    fields: active_flag(boolean), add_time(string), content(string), deal(object), deal_id(integer), id(integer), last_update_user_id(integer), lead_id(string), org_id(integer), organization(object), person(object), person_id(integer), pinned_to_deal_flag(boolean), pinned_to_organization_flag(boolean), pinned_to_person_flag(boolean), pinned_to_project_flag(boolean), pinned_to_task_flag(boolean), project(object), project_id(integer), task(object), task_id(integer), update_time(string), user(object), user_id(integer)
  leads:
    primary key: id
    cursor: update_time
    fields: add_time(string), cc_email(string), channel(integer), channel_id(string), creator_id(integer), expected_close_date(string), id(string), is_archived(boolean), label_ids(array), next_activity_id(integer), organization_id(integer), origin(string), origin_id(string), owner_id(integer), person_id(integer), source_deal_id(integer), source_name(string), title(string), update_time(string), value(object), visible_to(string), was_seen(boolean)
  deal_fields:
    primary key: id
    fields: active_flag(boolean), add_time(string), add_visible_flag(boolean), bulk_edit_allowed(boolean), created_by_user_id(integer), details_visible_flag(boolean), edit_flag(boolean), field_type(string), filtering_allowed(boolean), id(integer), important_flag(boolean), index_visible_flag(boolean), is_subfield(boolean), key(string), last_updated_by_user_id(integer), mandatory_flag(boolean), name(string), options(array), options_deleted(object), order_nr(integer), searchable_flag(boolean), sortable_flag(boolean), subfields(array), update_time(string)
  person_fields:
    primary key: id
    fields: active_flag(boolean), add_time(string), add_visible_flag(boolean), bulk_edit_allowed(boolean), created_by_user_id(integer), details_visible_flag(boolean), edit_flag(boolean), field_type(string), filtering_allowed(boolean), id(integer), important_flag(boolean), index_visible_flag(boolean), is_subfield(boolean), key(string), last_updated_by_user_id(integer), mandatory_flag(boolean), name(string), options(array), options_deleted(object), order_nr(integer), searchable_flag(boolean), sortable_flag(boolean), subfields(array), update_time(string)
  organization_fields:
    primary key: id
    fields: active_flag(boolean), add_time(string), add_visible_flag(boolean), bulk_edit_allowed(boolean), created_by_user_id(integer), details_visible_flag(boolean), edit_flag(boolean), field_type(string), filtering_allowed(boolean), id(integer), important_flag(boolean), index_visible_flag(boolean), is_subfield(boolean), key(string), last_updated_by_user_id(integer), mandatory_flag(boolean), name(string), options(array), options_deleted(object), order_nr(integer), searchable_flag(boolean), sortable_flag(boolean), subfields(array), update_time(string)
  product_fields:
    primary key: id
    fields: active_flag(boolean), add_time(string), add_visible_flag(boolean), bulk_edit_allowed(boolean), created_by_user_id(integer), details_visible_flag(boolean), edit_flag(boolean), field_type(string), filtering_allowed(boolean), id(integer), important_flag(boolean), index_visible_flag(boolean), is_subfield(boolean), key(string), last_updated_by_user_id(integer), mandatory_flag(boolean), name(string), options(array), options_deleted(object), order_nr(integer), searchable_flag(boolean), sortable_flag(boolean), subfields(array), update_time(string)
  lead_fields:
    primary key: id
    fields: active_flag(boolean), add_time(string), add_visible_flag(boolean), bulk_edit_allowed(boolean), created_by_user_id(integer), details_visible_flag(boolean), edit_flag(boolean), field_type(string), filtering_allowed(boolean), id(integer), important_flag(boolean), index_visible_flag(boolean), is_subfield(boolean), key(string), last_updated_by_user_id(integer), mandatory_flag(boolean), name(string), options(array), options_deleted(object), order_nr(integer), searchable_flag(boolean), sortable_flag(boolean), subfields(array), update_time(string)
  roles:
    primary key: id
    fields: active_flag(boolean), assignment_count(string), id(integer), level(integer), name(string), parent_role_id(integer), sub_role_count(string)
  filters:
    primary key: id
    fields: active_flag(boolean), add_time(string), custom_view_id(integer), filter_code(string), id(integer), is_editable(boolean), last_used_time(string), name(string), temporary_flag(boolean), type(string), update_time(string), user_id(integer), visible_to(string)
  activity_types:
    primary key: id
    fields: active_flag(boolean), add_time(string), color(string), icon_key(string), id(integer), is_custom_flag(boolean), key_string(string), name(string), order_nr(integer), update_time(string)
  legacy_teams:
    primary key: id
    fields: active_flag(boolean), add_time(string), created_by_user_id(integer), deleted_flag(boolean), description(string), id(integer), manager_id(integer), name(string), users(array)
  webhooks:
    primary key: id
    fields: add_time(string), admin_id(integer), company_id(integer), event_action(string), event_object(string), http_auth_user(string), id(integer), is_active(boolean), last_delivery_time(string), last_http_status(integer), name(string), owner_id(integer), remove_reason(string), remove_time(string), subscription_url(string), type(string), user_id(integer), version(string)
  lead_labels:
    primary key: id
    fields: add_time(string), color(string), id(string), name(string), update_time(string)
  lead_sources:
    primary key: name
    fields: name(string)
  currencies:
    primary key: id
    fields: active_flag(boolean), code(string), decimal_points(integer), id(integer), is_custom_flag(boolean), name(string), symbol(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_lead:
    endpoint: POST /leads
    required fields: title
    risk: creates a new lead; low-risk external mutation, no approval required
  update_lead:
    endpoint: PATCH /leads/{{ record.id }}
    required fields: id
    risk: updates an existing lead's fields (partial patch); external mutation, approval required
  delete_lead:
    endpoint: DELETE /leads/{{ record.id }}
    required fields: id
    risk: permanently deletes a lead; destructive external mutation, approval required
  create_note:
    endpoint: POST /notes
    required fields: content
    risk: creates a new note attached to a deal/person/organization/lead; low-risk external mutation, no approval required
  update_note:
    endpoint: PUT /notes/{{ record.id }}
    required fields: id, content
    risk: updates an existing note's content; external mutation, approval required
  delete_note:
    endpoint: DELETE /notes/{{ record.id }}
    required fields: id
    risk: permanently deletes a note; destructive external mutation, approval required
  create_filter:
    endpoint: POST /filters
    required fields: name, conditions, type
    risk: creates a new saved filter; low-risk external mutation, no approval required
  update_filter:
    endpoint: PUT /filters/{{ record.id }}
    required fields: id, name, conditions
    risk: updates an existing saved filter's name/conditions; external mutation, approval required
  delete_filter:
    endpoint: DELETE /filters/{{ record.id }}
    required fields: id
    risk: permanently deletes a saved filter; destructive external mutation, approval required
  create_activity_type:
    endpoint: POST /activityTypes
    required fields: name, icon_key
    risk: creates a new custom activity type; low-risk external mutation, no approval required
  update_activity_type:
    endpoint: PUT /activityTypes/{{ record.id }}
    required fields: id
    risk: updates an existing activity type's name/color/order; external mutation, approval required
  delete_activity_type:
    endpoint: DELETE /activityTypes/{{ record.id }}
    required fields: id
    risk: permanently deletes a custom activity type; destructive external mutation, approval required
  create_lead_label:
    endpoint: POST /leadLabels
    required fields: name, color
    risk: creates a new lead label; low-risk external mutation, no approval required
  update_lead_label:
    endpoint: PATCH /leadLabels/{{ record.id }}
    required fields: id
    risk: updates an existing lead label's name/color; external mutation, approval required
  delete_lead_label:
    endpoint: DELETE /leadLabels/{{ record.id }}
    required fields: id
    risk: permanently deletes a lead label; destructive external mutation, approval required
  create_webhook:
    endpoint: POST /webhooks
    required fields: subscription_url, event_action, event_object, name
    risk: registers a new webhook subscription that will receive event notifications; low-risk external mutation, no approval required
  delete_webhook:
    endpoint: DELETE /webhooks/{{ record.id }}
    required fields: id
    risk: permanently deletes a webhook subscription; destructive external mutation, approval required

SECURITY
  read risk: external Pipedrive API read of CRM deal, contact, organization, lead, note, and configuration data
  write risk: creates/updates/deletes leads, notes, saved filters, custom activity types, lead labels, and webhook subscriptions
  approval: required for update_lead/update_note/update_filter/update_activity_type/update_lead_label/delete_lead/delete_note/delete_filter/delete_activity_type/delete_lead_label/delete_webhook; create_lead/create_note/create_filter/create_activity_type/create_lead_label/create_webhook require no approval (low-risk, non-destructive)
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect pipedrive

  # Inspect as structured JSON
  pm connectors inspect pipedrive --json

AGENT WORKFLOW
  - Run pm connectors inspect pipedrive before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
