# pm connectors inspect systeme

```text
NAME
  pm connectors inspect systeme - Systeme connector manual

SYNOPSIS
  pm connectors inspect systeme
  pm connectors inspect systeme --json
  pm credentials add <name> --connector systeme [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Systeme.io contacts, tags, contact fields, funnels, and funnel steps, and writes contact/tag/contact-field/funnel lifecycle mutations and contact-tag assignment, through the Systeme.io public API.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret)

ETL STREAMS
  contacts:
    primary key: id
    cursor: created_at
    fields: created_at(), email(), id()
  tags:
    primary key: id
    fields: id(), name()
  contact_fields:
    primary key: id
    fields: id(), slug(), type()
  funnels:
    primary key: id
    fields: id(), name()
  funnel_steps:
    primary key: funnel_id, id
    fields: funnel_id(), id(), name()
  webhooks:
    primary key: id
    fields: event(), id(), url()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_contact:
    endpoint: POST /contacts
    risk: creates a new contact; low-risk external mutation, no approval required
  update_contact:
    endpoint: PATCH /contacts/{{ record.id }}
    required fields: id
    risk: updates an existing contact's email/locale/custom-field values; external mutation, no approval required
  delete_contact:
    endpoint: DELETE /contacts/{{ record.id }}
    required fields: id
    risk: irreversibly deletes a contact; approval required
  create_tag:
    endpoint: POST /tags
    risk: creates a new tag; low-risk external mutation, no approval required
  update_tag:
    endpoint: PUT /tags/{{ record.id }}
    required fields: id
    risk: renames an existing tag; external mutation, no approval required
  delete_tag:
    endpoint: DELETE /tags/{{ record.id }}
    required fields: id
    risk: irreversibly deletes a tag, removing it from every contact it is assigned to; approval required
  add_contact_tag:
    endpoint: POST /contacts/{{ record.contact_id }}/tags
    required fields: contact_id
    risk: assigns a tag to a contact; assigning certain tags can trigger Systeme.io automations (enrollment in a course/campaign); external mutation, no approval required
  remove_contact_tag:
    endpoint: DELETE /contacts/{{ record.contact_id }}/tags/{{ record.tag_id }}
    required fields: contact_id, tag_id
    risk: removes a tag from a contact; removing certain tags can trigger Systeme.io automations; external mutation, no approval required
  create_contact_field:
    endpoint: POST /contact_fields
    risk: creates a new custom contact field definition; low-risk external mutation, no approval required
  update_contact_field:
    endpoint: PATCH /contact_fields/{{ record.id }}
    required fields: id
    risk: updates an existing custom contact field definition; external mutation, no approval required
  delete_contact_field:
    endpoint: DELETE /contact_fields/{{ record.id }}
    required fields: id
    risk: irreversibly deletes a custom contact field definition and its stored values on every contact; approval required
  create_funnel:
    endpoint: POST /funnels
    risk: creates a new sales funnel; low-risk external mutation, no approval required
  create_funnel_step:
    endpoint: POST /funnels/{{ record.funnel_id }}/steps
    required fields: funnel_id
    risk: creates a new step within an existing funnel; low-risk external mutation, no approval required
  create_webhook:
    endpoint: POST /webhooks
    risk: creates a new outgoing webhook subscription; low-risk external mutation, no approval required

SECURITY
  read risk: external Systeme.io API read of contact, tag, funnel, and funnel-step data
  write risk: external Systeme.io API mutation (contact/tag/contact-field/funnel lifecycle, contact-tag assignment)
  approval: reverse ETL plan approval required before writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect systeme

  # Inspect as structured JSON
  pm connectors inspect systeme --json

AGENT WORKFLOW
  - Run pm connectors inspect systeme before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
