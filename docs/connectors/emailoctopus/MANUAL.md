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
  id: emailoctopus
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
    fields: created_at(string), double_opt_in(boolean), id(string), name(string), pending_count(integer), subscribed_count(integer), unsubscribed_count(integer)
  campaigns:
    primary key: id
    fields: created_at(string), from_email_address(string), from_name(string), id(string), name(string), sent_at(string), status(string), subject(string)
  list_contacts:
    primary key: id
    fields: created_at(string), email_address(string), fields(object), id(string), last_updated_at(string), status(string), tags(array)
  list_tags:
    primary key: tag
    fields: tag(string)
  campaign_summary_reports:
    primary key: campaign_id
    fields: bounced_hard(integer), bounced_soft(integer), campaign_id(string), clicked_total(integer), clicked_unique(integer), complained(integer), id(string), opened_total(integer), opened_unique(integer), sent(integer), unsubscribed(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_list:
    endpoint: POST /lists
    required fields: name
    risk: creates a new contact list; low-risk external mutation, no approval required
  update_list:
    endpoint: PUT /lists/{{ record.id }}
    required fields: id, name
    risk: renames an existing list; the id used by campaigns/API integrations to reference it is unchanged
  delete_list:
    endpoint: DELETE /lists/{{ record.id }}
    required fields: id
    risk: permanently removes a list and all of its contacts/tags/custom fields
  create_list_contact:
    endpoint: POST /lists/{{ record.list_id }}/contacts
    required fields: list_id, email_address
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
    required fields: list_id, tag
    risk: creates a new tag on a list, up to that list's tag-count limit; low-risk external mutation, no approval required
  update_list_tag:
    endpoint: PUT /lists/{{ record.list_id }}/tags/{{ record.tag }}
    required fields: list_id, tag, new_tag
    risk: renames an existing tag on a list; any external automation/segment referencing the old tag name stops matching contacts by that name
  delete_list_tag:
    endpoint: DELETE /lists/{{ record.list_id }}/tags/{{ record.tag }}
    required fields: list_id, tag
    risk: permanently removes a tag from a list and from every contact currently carrying it
  create_list_field:
    endpoint: POST /lists/{{ record.list_id }}/fields
    required fields: list_id, label, tag, type
    risk: creates a new custom field on a list; the field's type (NUMBER/TEXT/DATE) cannot be changed after creation
  update_list_field:
    endpoint: PUT /lists/{{ record.list_id }}/fields/{{ record.tag }}
    required fields: list_id, tag, label, new_tag
    optional fields: fallback
    risk: renames a custom field's label/tag or changes its fallback default; any email template referencing the old field tag stops resolving a value
  delete_list_field:
    endpoint: DELETE /lists/{{ record.list_id }}/fields/{{ record.tag }}
    required fields: list_id, tag
    risk: permanently removes a custom field and its stored values from every contact on the list
  start_automation:
    endpoint: POST /automations/{{ record.automation_id }}/queue
    required fields: automation_id, list_member_id
    risk: enrolls a contact into a live automation sequence, triggering its configured emails/delays; the automation must already have the 'Started via API' trigger enabled in the EmailOctopus dashboard

SECURITY
  read risk: external EmailOctopus API read of list, campaign, campaign-report, contact, tag, and custom-field data
  write risk: external EmailOctopus API mutations covering list/contact/tag/custom-field lifecycle management, plus start_automation, which enrolls a contact into a live automation sequence and triggers its configured email sends
  approval: standard; no destructive-admin or elevated-scope actions are exposed
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run EmailOctopus's declared streams and reverse-ETL actions.
  Usage: pm emailoctopus <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    api get campaigns campaignid reports bounced - Documented GET /campaigns/{campaignId}/reports/bounced (not implemented) [intent=direct_read availability=not_implemented operation=emailoctopus.get.campaigns-campaignid-reports-bounced]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get campaigns campaignid reports clicked - Documented GET /campaigns/{campaignId}/reports/clicked (not implemented) [intent=direct_read availability=not_implemented operation=emailoctopus.get.campaigns-campaignid-reports-clicked]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get campaigns campaignid reports complained - Documented GET /campaigns/{campaignId}/reports/complained (not implemented) [intent=direct_read availability=not_implemented operation=emailoctopus.get.campaigns-campaignid-reports-complained]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get campaigns campaignid reports links - Documented GET /campaigns/{campaignId}/reports/links (not implemented) [intent=direct_read availability=not_implemented operation=emailoctopus.get.campaigns-campaignid-reports-links]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get campaigns campaignid reports not-clicked - Documented GET /campaigns/{campaignId}/reports/not-clicked (not implemented) [intent=direct_read availability=not_implemented operation=emailoctopus.get.campaigns-campaignid-reports-not-clicked]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get campaigns campaignid reports not-opened - Documented GET /campaigns/{campaignId}/reports/not-opened (not implemented) [intent=direct_read availability=not_implemented operation=emailoctopus.get.campaigns-campaignid-reports-not-opened]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get campaigns campaignid reports opened - Documented GET /campaigns/{campaignId}/reports/opened (not implemented) [intent=direct_read availability=not_implemented operation=emailoctopus.get.campaigns-campaignid-reports-opened]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get campaigns campaignid reports sent - Documented GET /campaigns/{campaignId}/reports/sent (not implemented) [intent=direct_read availability=not_implemented operation=emailoctopus.get.campaigns-campaignid-reports-sent]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get campaigns campaignid reports unsubscribed - Documented GET /campaigns/{campaignId}/reports/unsubscribed (not implemented) [intent=direct_read availability=not_implemented operation=emailoctopus.get.campaigns-campaignid-reports-unsubscribed]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get campaigns id - Documented GET /campaigns/{id} (not implemented) [intent=direct_read availability=not_implemented operation=emailoctopus.get.campaigns-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get lists id - Documented GET /lists/{id} (not implemented) [intent=direct_read availability=not_implemented operation=emailoctopus.get.lists-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get lists id contacts email - Documented GET /lists/{id}/contacts/{email} (not implemented) [intent=direct_read availability=not_implemented operation=emailoctopus.get.lists-id-contacts-email]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get lists id contacts subscribed - Documented GET /lists/{id}/contacts/subscribed (not implemented) [intent=direct_read availability=not_implemented operation=emailoctopus.get.lists-id-contacts-subscribed]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get lists id contacts tagged tag - Documented GET /lists/{id}/contacts/tagged/{tag} (not implemented) [intent=direct_read availability=not_implemented operation=emailoctopus.get.lists-id-contacts-tagged-tag]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get lists id contacts unsubscribed - Documented GET /lists/{id}/contacts/unsubscribed (not implemented) [intent=direct_read availability=not_implemented operation=emailoctopus.get.lists-id-contacts-unsubscribed]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api put lists id contacts bulk - Documented PUT /lists/{id}/contacts/bulk (not implemented) [intent=direct_write availability=not_implemented operation=emailoctopus.put.lists-id-contacts-bulk]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    campaign summary reports list - Run the campaign summary reports ETL stream [intent=etl availability=implemented stream=campaign_summary_reports]
    campaigns list - Run the campaigns ETL stream [intent=etl availability=implemented stream=campaigns]
    create list apply - Plan and execute the create list reverse-ETL action [intent=reverse_etl availability=implemented write=create_list]; approval: requires plan, preview, approval, and execute; risk: creates a new contact list; low-risk external mutation, no approval required; flags: --name (required)
    create list contact apply - Plan and execute the create list contact reverse-ETL action [intent=reverse_etl availability=implemented write=create_list_contact]; approval: requires plan, preview, approval, and execute; risk: adds a new contact to a list, immediately eligible to receive future campaigns targeting it (unless status is PENDING on a double opt-in list); flags: --email_address (required), --list_id (required)
    create list field apply - Plan and execute the create list field reverse-ETL action [intent=reverse_etl availability=implemented write=create_list_field]; approval: requires plan, preview, approval, and execute; risk: creates a new custom field on a list; the field's type (NUMBER/TEXT/DATE) cannot be changed after creation; flags: --label (required), --list_id (required), --tag (required), --type (required)
    create list tag apply - Plan and execute the create list tag reverse-ETL action [intent=reverse_etl availability=implemented write=create_list_tag]; approval: requires plan, preview, approval, and execute; risk: creates a new tag on a list, up to that list's tag-count limit; low-risk external mutation, no approval required; flags: --list_id (required), --tag (required)
    delete list apply - Plan and execute the delete list reverse-ETL action [intent=reverse_etl availability=implemented write=delete_list]; approval: requires plan, preview, approval, and execute; risk: permanently removes a list and all of its contacts/tags/custom fields; flags: --id (required)
    delete list contact apply - Plan and execute the delete list contact reverse-ETL action [intent=reverse_etl availability=implemented write=delete_list_contact]; approval: requires plan, preview, approval, and execute; risk: permanently removes a contact from a list and its subscription/consent history; flags: --list_id (required), --member_id (required)
    delete list field apply - Plan and execute the delete list field reverse-ETL action [intent=reverse_etl availability=implemented write=delete_list_field]; approval: requires plan, preview, approval, and execute; risk: permanently removes a custom field and its stored values from every contact on the list; flags: --list_id (required), --tag (required)
    delete list tag apply - Plan and execute the delete list tag reverse-ETL action [intent=reverse_etl availability=implemented write=delete_list_tag]; approval: requires plan, preview, approval, and execute; risk: permanently removes a tag from a list and from every contact currently carrying it; flags: --list_id (required), --tag (required)
    list contacts list - Run the list contacts ETL stream [intent=etl availability=implemented stream=list_contacts]
    list tags list - Run the list tags ETL stream [intent=etl availability=implemented stream=list_tags]
    lists list - Run the lists ETL stream [intent=etl availability=implemented stream=lists]
    start automation apply - Plan and execute the start automation reverse-ETL action [intent=reverse_etl availability=implemented write=start_automation]; approval: requires plan, preview, approval, and execute; risk: enrolls a contact into a live automation sequence, triggering its configured emails/delays; the automation must already have the 'Started via API' trigger enabled in the EmailOctopus dashboard; flags: --automation_id (required), --list_member_id (required)
    update list apply - Plan and execute the update list reverse-ETL action [intent=reverse_etl availability=implemented write=update_list]; approval: requires plan, preview, approval, and execute; risk: renames an existing list; the id used by campaigns/API integrations to reference it is unchanged; flags: --id (required), --name (required)
    update list contact apply - Plan and execute the update list contact reverse-ETL action [intent=reverse_etl availability=implemented write=update_list_contact]; approval: requires plan, preview, approval, and execute; risk: mutates an existing contact's email/fields/tags/status; a status change to UNSUBSCRIBED or SUBSCRIBED changes future campaign eligibility for this recipient; flags: --list_id (required), --member_id (required)
    update list field apply - Plan and execute the update list field reverse-ETL action [intent=reverse_etl availability=implemented write=update_list_field]; approval: requires plan, preview, approval, and execute; risk: renames a custom field's label/tag or changes its fallback default; any email template referencing the old field tag stops resolving a value; flags: --label (required), --list_id (required), --new_tag (required), --tag (required)
    update list tag apply - Plan and execute the update list tag reverse-ETL action [intent=reverse_etl availability=implemented write=update_list_tag]; approval: requires plan, preview, approval, and execute; risk: renames an existing tag on a list; any external automation/segment referencing the old tag name stops matching contacts by that name; flags: --list_id (required), --new_tag (required), --tag (required)

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
