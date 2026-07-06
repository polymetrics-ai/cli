# pm connectors inspect emailoctopus

```text
NAME
  pm connectors inspect emailoctopus - EmailOctopus connector manual

SYNOPSIS
  pm connectors inspect emailoctopus
  pm connectors inspect emailoctopus --json
  pm credentials add <name> --connector emailoctopus [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes EmailOctopus lists, campaigns, campaign summary reports, list contacts, list tags, and list custom fields through the EmailOctopus v1.6 REST API.

ICON
  asset: icons/emailoctopus.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://emailoctopus.com/api-documentation

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  list_id
  mode
  api_key (secret)

ETL STREAMS
  lists:
    primary key: id
    fields: created_at(), double_opt_in(), id(), name(), pending_count(), subscribed_count(), unsubscribed_count()
  campaigns:
    primary key: id
    fields: created_at(), from_email_address(), from_name(), id(), name(), sent_at(), status(), subject()
  list_contacts:
    primary key: id
    fields: created_at(), email_address(), fields(), id(), last_updated_at(), status(), tags()
  list_tags:
    primary key: tag
    fields: tag()
  campaign_summary_reports:
    primary key: campaign_id
    fields: bounced_hard(), bounced_soft(), campaign_id(), clicked_total(), clicked_unique(), complained(), id(), opened_total(), opened_unique(), sent(), unsubscribed()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_list:
    endpoint: POST /lists
    risk: creates a new contact list; low-risk external mutation, no approval required
  update_list:
    endpoint: PUT /lists/{{ record.id }}
    required fields: id
    optional fields: name
    risk: renames an existing list; the id used by campaigns/API integrations to reference it is unchanged
  delete_list:
    endpoint: DELETE /lists/{{ record.id }}
    required fields: id
    risk: permanently removes a list and all of its contacts/tags/custom fields
  create_list_contact:
    endpoint: POST /lists/{{ record.list_id }}/contacts
    required fields: list_id
    risk: adds a new contact to a list, immediately eligible to receive future campaigns targeting it (unless status is PENDING on a double opt-in list)
  update_list_contact:
    endpoint: PUT /lists/{{ record.list_id }}/contacts/{{ record.member_id }}
    required fields: list_id, member_id
    risk: mutates an existing contact's email/fields/tags/status; a status change to UNSUBSCRIBED or SUBSCRIBED changes future campaign eligibility for this recipient
  delete_list_contact:
    endpoint: DELETE /lists/{{ record.list_id }}/contacts/{{ record.member_id }}
    required fields: list_id, member_id
    risk: permanently removes a contact from a list and its subscription/consent history
  create_list_tag:
    endpoint: POST /lists/{{ record.list_id }}/tags
    required fields: list_id
    optional fields: tag
    risk: creates a new tag on a list, up to that list's tag-count limit; low-risk external mutation, no approval required
  update_list_tag:
    endpoint: PUT /lists/{{ record.list_id }}/tags/{{ record.tag }}
    required fields: list_id, tag
    optional fields: new_tag
    risk: renames an existing tag on a list; any external automation/segment referencing the old tag name stops matching contacts by that name
  delete_list_tag:
    endpoint: DELETE /lists/{{ record.list_id }}/tags/{{ record.tag }}
    required fields: list_id, tag
    risk: permanently removes a tag from a list and from every contact currently carrying it
  create_list_field:
    endpoint: POST /lists/{{ record.list_id }}/fields
    required fields: list_id
    risk: creates a new custom field on a list; the field's type (NUMBER/TEXT/DATE) cannot be changed after creation
  update_list_field:
    endpoint: PUT /lists/{{ record.list_id }}/fields/{{ record.tag }}
    required fields: list_id, tag
    optional fields: label, new_tag, fallback
    risk: renames a custom field's label/tag or changes its fallback default; any email template referencing the old field tag stops resolving a value
  delete_list_field:
    endpoint: DELETE /lists/{{ record.list_id }}/fields/{{ record.tag }}
    required fields: list_id, tag
    risk: permanently removes a custom field and its stored values from every contact on the list
  start_automation:
    endpoint: POST /automations/{{ record.automation_id }}/queue
    required fields: automation_id
    optional fields: list_member_id
    risk: enrolls a contact into a live automation sequence, triggering its configured emails/delays; the automation must already have the 'Started via API' trigger enabled in the EmailOctopus dashboard

SECURITY
  read risk: external EmailOctopus API read of list, campaign, campaign-report, contact, tag, and custom-field data
  write risk: external EmailOctopus API mutations covering list/contact/tag/custom-field lifecycle management, plus start_automation, which enrolls a contact into a live automation sequence and triggers its configured email sends
  approval: standard; no destructive-admin or elevated-scope actions are exposed
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect emailoctopus

  # Inspect as structured JSON
  pm connectors inspect emailoctopus --json

AGENT WORKFLOW
  - Run pm connectors inspect emailoctopus before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
