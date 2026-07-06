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
  api_token (secret)

ETL STREAMS
  deals:
    primary key: id
    cursor: update_time
    fields: add_time(), currency(), id(), org_id(), person_id(), stage_id(), status(), title(), update_time(), value()
  persons:
    primary key: id
    cursor: update_time
    fields: add_time(), email(), id(), name(), org_id(), owner_id(), phone(), update_time()
  organizations:
    primary key: id
    cursor: update_time
    fields: add_time(), id(), name(), owner_id(), people_count(), update_time()
  activities:
    primary key: id
    cursor: update_time
    fields: add_time(), deal_id(), done(), due_date(), id(), org_id(), person_id(), subject(), type(), update_time()
  products:
    primary key: id
    fields: active_flag(), add_time(), code(), id(), name(), owner_id(), unit(), update_time()
  users:
    primary key: id
    fields: active_flag(), created(), email(), id(), is_admin(), modified(), name()
  notes:
    primary key: id
    cursor: update_time
    fields: active_flag(), add_time(), content(), deal(), deal_id(), id(), last_update_user_id(), lead_id(), org_id(), organization(), person(), person_id(), pinned_to_deal_flag(), pinned_to_organization_flag(), pinned_to_person_flag(), pinned_to_project_flag(), pinned_to_task_flag(), project(), project_id(), task(), task_id(), update_time(), user(), user_id()
  leads:
    primary key: id
    cursor: update_time
    fields: add_time(), cc_email(), channel(), channel_id(), creator_id(), expected_close_date(), id(), is_archived(), label_ids(), next_activity_id(), organization_id(), origin(), origin_id(), owner_id(), person_id(), source_deal_id(), source_name(), title(), update_time(), value(), visible_to(), was_seen()
  deal_fields:
    primary key: id
    fields: active_flag(), add_time(), add_visible_flag(), bulk_edit_allowed(), created_by_user_id(), details_visible_flag(), edit_flag(), field_type(), filtering_allowed(), id(), important_flag(), index_visible_flag(), is_subfield(), key(), last_updated_by_user_id(), mandatory_flag(), name(), options(), options_deleted(), order_nr(), searchable_flag(), sortable_flag(), subfields(), update_time()
  person_fields:
    primary key: id
    fields: active_flag(), add_time(), add_visible_flag(), bulk_edit_allowed(), created_by_user_id(), details_visible_flag(), edit_flag(), field_type(), filtering_allowed(), id(), important_flag(), index_visible_flag(), is_subfield(), key(), last_updated_by_user_id(), mandatory_flag(), name(), options(), options_deleted(), order_nr(), searchable_flag(), sortable_flag(), subfields(), update_time()
  organization_fields:
    primary key: id
    fields: active_flag(), add_time(), add_visible_flag(), bulk_edit_allowed(), created_by_user_id(), details_visible_flag(), edit_flag(), field_type(), filtering_allowed(), id(), important_flag(), index_visible_flag(), is_subfield(), key(), last_updated_by_user_id(), mandatory_flag(), name(), options(), options_deleted(), order_nr(), searchable_flag(), sortable_flag(), subfields(), update_time()
  product_fields:
    primary key: id
    fields: active_flag(), add_time(), add_visible_flag(), bulk_edit_allowed(), created_by_user_id(), details_visible_flag(), edit_flag(), field_type(), filtering_allowed(), id(), important_flag(), index_visible_flag(), is_subfield(), key(), last_updated_by_user_id(), mandatory_flag(), name(), options(), options_deleted(), order_nr(), searchable_flag(), sortable_flag(), subfields(), update_time()
  lead_fields:
    primary key: id
    fields: active_flag(), add_time(), add_visible_flag(), bulk_edit_allowed(), created_by_user_id(), details_visible_flag(), edit_flag(), field_type(), filtering_allowed(), id(), important_flag(), index_visible_flag(), is_subfield(), key(), last_updated_by_user_id(), mandatory_flag(), name(), options(), options_deleted(), order_nr(), searchable_flag(), sortable_flag(), subfields(), update_time()
  roles:
    primary key: id
    fields: active_flag(), assignment_count(), id(), level(), name(), parent_role_id(), sub_role_count()
  filters:
    primary key: id
    fields: active_flag(), add_time(), custom_view_id(), filter_code(), id(), is_editable(), last_used_time(), name(), temporary_flag(), type(), update_time(), user_id(), visible_to()
  activity_types:
    primary key: id
    fields: active_flag(), add_time(), color(), icon_key(), id(), is_custom_flag(), key_string(), name(), order_nr(), update_time()
  legacy_teams:
    primary key: id
    fields: active_flag(), add_time(), created_by_user_id(), deleted_flag(), description(), id(), manager_id(), name(), users()
  webhooks:
    primary key: id
    fields: add_time(), admin_id(), company_id(), event_action(), event_object(), http_auth_user(), id(), is_active(), last_delivery_time(), last_http_status(), name(), owner_id(), remove_reason(), remove_time(), subscription_url(), type(), user_id(), version()
  lead_labels:
    primary key: id
    fields: add_time(), color(), id(), name(), update_time()
  lead_sources:
    primary key: name
    fields: name()
  currencies:
    primary key: id
    fields: active_flag(), code(), decimal_points(), id(), is_custom_flag(), name(), symbol()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_lead:
    endpoint: POST /leads
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
    risk: creates a new note attached to a deal/person/organization/lead; low-risk external mutation, no approval required
  update_note:
    endpoint: PUT /notes/{{ record.id }}
    required fields: id
    risk: updates an existing note's content; external mutation, approval required
  delete_note:
    endpoint: DELETE /notes/{{ record.id }}
    required fields: id
    risk: permanently deletes a note; destructive external mutation, approval required
  create_filter:
    endpoint: POST /filters
    risk: creates a new saved filter; low-risk external mutation, no approval required
  update_filter:
    endpoint: PUT /filters/{{ record.id }}
    required fields: id
    risk: updates an existing saved filter's name/conditions; external mutation, approval required
  delete_filter:
    endpoint: DELETE /filters/{{ record.id }}
    required fields: id
    risk: permanently deletes a saved filter; destructive external mutation, approval required
  create_activity_type:
    endpoint: POST /activityTypes
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
