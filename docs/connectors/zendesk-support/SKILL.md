---
name: pm-zendesk-support
description: Zendesk Support connector knowledge and safe action guide.
---

# pm-zendesk-support

## Purpose

Reads and writes allow-listed Zendesk Support resources through the Zendesk Support REST API v2; Pass B expands top-level read streams from the public Airbyte/Zendesk surface.

## Icon

- asset: icons/zendesk-support.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developer.zendesk.com/api-reference/ticketing/introduction/

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- access_token (secret)
- api_token (secret)
- email (secret)

## ETL Streams

- tickets:
  - primary key: id
  - cursor: updated_at
  - fields: assignee_id(), brand_id(), created_at(), description(), group_id(), id(), organization_id(), priority(), requester_id(), status(), subject(), type(), updated_at()
- users:
  - primary key: id
  - cursor: updated_at
  - fields: active(), created_at(), email(), id(), name(), organization_id(), phone(), role(), time_zone(), updated_at(), verified()
- organizations:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), details(), group_id(), id(), name(), notes(), shared_comments(), shared_tickets(), updated_at()
- groups:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), default(), deleted(), description(), id(), name(), updated_at()
- satisfaction_ratings:
  - primary key: id
  - cursor: updated_at
  - fields: assignee_id(), comment(), created_at(), group_id(), id(), reason(), requester_id(), score(), ticket_id(), updated_at()
- deleted_tickets:
  - primary key: id
  - fields: actor(), deleted_at(), description(), id(), previous_state(), subject()
- account_attributes:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), id(), name(), updated_at(), url()
- attribute_definitions:
  - fields: condition(), confition(), group(), metadata(), nullable(), operators(), repeatable(), subject(), title(), type(), values()
- brands:
  - primary key: id
  - cursor: updated_at
  - fields: active(), brand_url(), created_at(), default(), has_help_center(), help_center_state(), host_mapping(), id(), is_deleted(), logo(), name(), signature_template(), subdomain(), ticket_form_ids(), updated_at(), url()
- custom_roles:
  - primary key: id
  - cursor: updated_at
  - fields: configuration(), created_at(), description(), id(), manage_macro_content_suggestions(), name(), read_macro_content_suggestions(), role_type(), team_member_count(), updated_at()
- schedules:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), id(), intervals(), name(), time_zone(), updated_at()
- sla_policies:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), description(), filter(), id(), policy_metrics(), position(), title(), updated_at(), url()
- tags:
  - primary key: name
  - fields: count(), name()
- ticket_fields:
  - primary key: id
  - cursor: updated_at
  - fields: active(), agent_description(), collapsed_for_agents(), created_at(), custom_field_options(), custom_statuses(), description(), editable_in_portal(), id(), key(), position(), raw_description(), raw_title(), raw_title_in_portal(), regexp_for_validation(), removable(), required(), required_in_portal(), sub_type_id(), system_field_options(), tag(), title(), title_in_portal(), type(), updated_at(), url(), visible_in_portal()
- ticket_forms:
  - primary key: id
  - cursor: updated_at
  - fields: active(), agent_conditions(), created_at(), default(), display_name(), end_user_conditions(), end_user_visible(), id(), in_all_brands(), name(), position(), raw_display_name(), raw_name(), restricted_brand_ids(), ticket_field_ids(), updated_at(), url()
- topics:
  - primary key: id
  - cursor: updated_at
  - fields: community_id(), created_at(), description(), follower_count(), html_url(), id(), manageable_by(), name(), position(), updated_at(), url(), user_segment_id()
- user_fields:
  - primary key: id
  - cursor: updated_at
  - fields: active(), created_at(), description(), id(), key(), position(), raw_description(), raw_title(), regexp_for_validation(), system(), tag(), title(), type(), updated_at(), url()
- automations:
  - primary key: id
  - cursor: updated_at
  - fields: actions(), active(), conditions(), created_at(), default(), id(), position(), raw_title(), title(), updated_at()
- categories:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), description(), html_url(), id(), locale(), name(), outdated(), position(), source_locale(), updated_at(), url()
- sections:
  - primary key: id
  - cursor: updated_at
  - fields: category_id(), created_at(), description(), html_url(), id(), locale(), name(), outdated(), parent_section_id(), position(), source_locale(), theme_template(), updated_at(), url()
- articles:
  - primary key: id
  - cursor: updated_at
  - fields: author_id(), body(), comments_disabled(), content_tag_ids(), created_at(), draft(), edited_at(), html_url(), id(), label_names(), locale(), name(), outdated(), outdated_locales(), permission_group_id(), position(), promoted(), section_id(), source_locale(), title(), updated_at(), url(), user_segment_id(), vote_count(), vote_sum()
- group_memberships:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), default(), group_id(), id(), updated_at(), url(), user_id()
- macros:
  - primary key: id
  - cursor: updated_at
  - fields: actions(), active(), created_at(), default(), description(), id(), position(), raw_title(), restriction(), title(), updated_at(), url()
- organization_fields:
  - primary key: id
  - cursor: updated_at
  - fields: active(), created_at(), custom_field_options(), description(), id(), key(), position(), raw_description(), raw_title(), regexp_for_validation(), relationship_filter(), relationship_target_type(), system(), tag(), title(), type(), updated_at(), url()
- organization_memberships:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), default(), id(), organization_id(), organization_name(), updated_at(), url(), user_id(), view_tickets()
- posts:
  - primary key: id
  - cursor: updated_at
  - fields: author_id(), closed(), comment_count(), content_tag_ids(), created_at(), description(), details(), featured(), follower_count(), frozen(), html_url(), id(), non_author_editor_id(), non_author_updated_at(), pinned(), status(), title(), topic_id(), updated_at(), url(), vote_count(), vote_sum()
- ticket_activities:
  - primary key: id
  - cursor: updated_at
  - fields: actor(), actor_id(), created_at(), description(), id(), object(), target(), title(), updated_at(), url(), user(), user_id(), verb()
- ticket_audits:
  - primary key: id
  - fields: attachments(), author_id(), created_at(), events(), id(), metadata(), ticket_id(), via()
- ticket_metric_events:
  - primary key: id
  - fields: deleted(), group_sla(), id(), instance_id(), metric(), sla(), status(), ticket_id(), time(), type()
- ticket_events:
  - primary key: id
  - fields: child_events(), created_at(), event_type(), id(), system(), ticket_id(), timestamp(), updater_id(), via(), via_reference_id()
- ticket_skips:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), id(), reason(), ticket(), ticket_id(), updated_at(), user_id()
- triggers:
  - primary key: id
  - cursor: updated_at
  - fields: actions(), active(), category_id(), conditions(), created_at(), default(), description(), id(), position(), raw_title(), title(), updated_at(), url()
- views:
  - primary key: id
  - cursor: updated_at
  - fields: active(), created_at(), id(), position(), title(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_ticket:
  - endpoint: POST /api/v2/tickets
  - optional fields: ticket
  - risk: creates a Zendesk ticket record; external mutation requiring approval
- update_ticket:
  - endpoint: PUT /api/v2/tickets/{{ record.id }}
  - required fields: id
  - optional fields: ticket
  - risk: updates a Zendesk ticket record; external mutation requiring approval
- delete_ticket:
  - endpoint: DELETE /api/v2/tickets/{{ record.id }}
  - required fields: id
  - risk: deletes a Zendesk ticket record; destructive external mutation requiring approval
- create_user:
  - endpoint: POST /api/v2/users
  - optional fields: user
  - risk: creates a Zendesk user record; external mutation requiring approval
- update_user:
  - endpoint: PUT /api/v2/users/{{ record.id }}
  - required fields: id
  - optional fields: user
  - risk: updates a Zendesk user record; external mutation requiring approval
- delete_user:
  - endpoint: DELETE /api/v2/users/{{ record.id }}
  - required fields: id
  - risk: deletes a Zendesk user record; destructive external mutation requiring approval
- create_organization:
  - endpoint: POST /api/v2/organizations
  - optional fields: organization
  - risk: creates a Zendesk organization record; external mutation requiring approval
- update_organization:
  - endpoint: PUT /api/v2/organizations/{{ record.id }}
  - required fields: id
  - optional fields: organization
  - risk: updates a Zendesk organization record; external mutation requiring approval
- delete_organization:
  - endpoint: DELETE /api/v2/organizations/{{ record.id }}
  - required fields: id
  - risk: deletes a Zendesk organization record; destructive external mutation requiring approval
- create_group:
  - endpoint: POST /api/v2/groups
  - optional fields: group
  - risk: creates a Zendesk group record; external mutation requiring approval
- update_group:
  - endpoint: PUT /api/v2/groups/{{ record.id }}
  - required fields: id
  - optional fields: group
  - risk: updates a Zendesk group record; external mutation requiring approval
- delete_group:
  - endpoint: DELETE /api/v2/groups/{{ record.id }}
  - required fields: id
  - risk: deletes a Zendesk group record; destructive external mutation requiring approval
- create_macro:
  - endpoint: POST /api/v2/macros
  - optional fields: macro
  - risk: creates a Zendesk macro record; external mutation requiring approval
- update_macro:
  - endpoint: PUT /api/v2/macros/{{ record.id }}
  - required fields: id
  - optional fields: macro
  - risk: updates a Zendesk macro record; external mutation requiring approval
- delete_macro:
  - endpoint: DELETE /api/v2/macros/{{ record.id }}
  - required fields: id
  - risk: deletes a Zendesk macro record; destructive external mutation requiring approval
- create_trigger:
  - endpoint: POST /api/v2/triggers
  - optional fields: trigger
  - risk: creates a Zendesk trigger record; external mutation requiring approval
- update_trigger:
  - endpoint: PUT /api/v2/triggers/{{ record.id }}
  - required fields: id
  - optional fields: trigger
  - risk: updates a Zendesk trigger record; external mutation requiring approval
- delete_trigger:
  - endpoint: DELETE /api/v2/triggers/{{ record.id }}
  - required fields: id
  - risk: deletes a Zendesk trigger record; destructive external mutation requiring approval
- create_automation:
  - endpoint: POST /api/v2/automations
  - optional fields: automation
  - risk: creates a Zendesk automation record; external mutation requiring approval
- update_automation:
  - endpoint: PUT /api/v2/automations/{{ record.id }}
  - required fields: id
  - optional fields: automation
  - risk: updates a Zendesk automation record; external mutation requiring approval
- delete_automation:
  - endpoint: DELETE /api/v2/automations/{{ record.id }}
  - required fields: id
  - risk: deletes a Zendesk automation record; destructive external mutation requiring approval
- create_view:
  - endpoint: POST /api/v2/views
  - optional fields: view
  - risk: creates a Zendesk view record; external mutation requiring approval
- update_view:
  - endpoint: PUT /api/v2/views/{{ record.id }}
  - required fields: id
  - optional fields: view
  - risk: updates a Zendesk view record; external mutation requiring approval
- delete_view:
  - endpoint: DELETE /api/v2/views/{{ record.id }}
  - required fields: id
  - risk: deletes a Zendesk view record; destructive external mutation requiring approval
- create_ticket_field:
  - endpoint: POST /api/v2/ticket_fields
  - optional fields: ticket_field
  - risk: creates a Zendesk ticket field record; external mutation requiring approval
- update_ticket_field:
  - endpoint: PUT /api/v2/ticket_fields/{{ record.id }}
  - required fields: id
  - optional fields: ticket_field
  - risk: updates a Zendesk ticket field record; external mutation requiring approval
- delete_ticket_field:
  - endpoint: DELETE /api/v2/ticket_fields/{{ record.id }}
  - required fields: id
  - risk: deletes a Zendesk ticket field record; destructive external mutation requiring approval

## Security

- read risk: external Zendesk Support API read across tickets, users, organizations, groups, satisfaction ratings, admin metadata, business rules, and Guide/community top-level resources
- write risk: allow-listed Zendesk Support mutations for tickets, users, organizations, groups, macros, triggers, automations, views, and ticket fields; destructive deletes require approval
- approval: writes require reverse-ETL approval; reads require normal connector access
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect zendesk-support
```

### Inspect as structured JSON

```bash
pm connectors inspect zendesk-support --json
```

## Agent Rules

- Run pm connectors inspect zendesk-support before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
