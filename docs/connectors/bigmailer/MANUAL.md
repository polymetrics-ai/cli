# pm connectors inspect bigmailer

```text
NAME
  pm connectors inspect bigmailer - BigMailer connector manual

SYNOPSIS
  pm connectors inspect bigmailer
  pm connectors inspect bigmailer --json
  pm credentials add <name> --connector bigmailer [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes BigMailer brands, account users, and brand-scoped contacts, lists, custom fields, message types, segments, senders, templates, suppression lists, and campaigns through the BigMailer REST API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  api_key (secret) (required)

ETL STREAMS
  brands:
    primary key: id
    fields: connection_id(string), contact_limit(integer), created(integer), from_email(string), from_name(string), id(string), name(string), num_contacts(integer)
  users:
    primary key: id
    fields: created(integer), email(string), id(string), name(string), role(string)
  contacts:
    primary key: id
    fields: brand_id(string), created(integer), email(string), id(string), num_complaints(integer), num_hard_bounces(integer), num_soft_bounces(integer), unsubscribe_all(boolean)
  lists:
    primary key: id
    fields: brand_id(string), created(integer), id(string), name(string), num_contacts(integer)
  fields:
    primary key: id
    fields: brand_id(string), created(integer), id(string), name(string), tag(string), type(string)
  connections:
    primary key: id
    fields: created(integer), id(string), name(string), type(string)
  message_types:
    primary key: id
    fields: brand_id(string), created(integer), id(string), name(string), type(string)
  segments:
    primary key: id
    fields: brand_id(string), conditions(array), created(integer), id(string), name(string), operator(string)
  senders:
    primary key: id
    fields: bounce_dns_records(object), bounce_domain(string), bounce_verified(boolean), brand_id(string), created(integer), dns_records(object), id(string), identity(string), identity_type(string), share_type(string), verified(boolean)
  templates:
    primary key: id
    fields: brand_id(string), created(integer), id(string), name(string), shared_with_account(boolean), type(string)
  suppression_lists:
    primary key: id
    fields: brand_id(string), created(integer), file_name(string), file_size(integer), id(string)
  bulk_campaigns:
    primary key: id
    cursor: created
    fields: brand_id(string), created(integer), excluded_list_ids(array), from(object), id(string), list_ids(array), message_type_id(string), name(string), num_clicks(integer), num_opens(integer), num_rejected(integer), num_sent(integer), num_total_clicks(integer), reply_to(object), scheduled_for(integer), segment_id(string), status(string), subject(string), suppression_list_id(string)
  rss_campaigns:
    primary key: id
    cursor: created
    fields: brand_id(string), created(integer), excluded_list_ids(array), feed_url(string), frequency(object), from(object), hour(integer), id(string), list_ids(array), message_type_id(string), minutes(integer), name(string), reply_to(object), segment_id(string), status(string), subject(string), suppression_list_id(string)
  transactional_campaigns:
    primary key: id
    cursor: created
    fields: brand_id(string), created(integer), from(object), id(string), list_id(string), message_type_id(string), name(string), num_clicks(integer), num_complaints(integer), num_hard_bounces(integer), num_opens(integer), num_rejected(integer), num_sent(integer), num_soft_bounces(integer), num_total_clicks(integer), num_total_opens(integer), num_unsubscribes(integer), reply_to(object), status(string), subject(string)
  test_campaigns:
    primary key: id
    cursor: created
    fields: brand_id(string), created(integer), feed_url(string), from(object), id(string), name(string), num_sent(integer), recipients(array), reply_to(object), sent(integer), started(integer), status(string), subject(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_brand:
    endpoint: POST /brands
    required fields: name, from_name, from_email, connection_id
    risk: external mutation; creates a new BigMailer brand (sending identity); approval required
  update_brand:
    endpoint: POST /brands/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  create_contact:
    endpoint: POST /brands/{{ record.brand_id }}/contacts
    required fields: brand_id, email
    risk: external mutation; creates a contact in a BigMailer brand; approval required
  update_contact:
    endpoint: POST /brands/{{ record.brand_id }}/contacts/{{ record.id }}
    required fields: brand_id, id
    risk: external mutation; approval required
  upsert_contact:
    endpoint: POST /brands/{{ record.brand_id }}/contacts/upsert
    required fields: brand_id, email
    risk: external mutation; creates the contact if the email is new, otherwise updates the existing contact; approval required
  delete_contact:
    endpoint: DELETE /brands/{{ record.brand_id }}/contacts/{{ record.id }}
    required fields: brand_id, id
    risk: permanently removes a contact from a brand; irreversible; approval required
  create_list:
    endpoint: POST /brands/{{ record.brand_id }}/lists
    required fields: brand_id, name
    risk: external mutation; creates a contact list in a BigMailer brand; approval required
  update_list:
    endpoint: POST /brands/{{ record.brand_id }}/lists/{{ record.id }}
    required fields: brand_id, id
    risk: external mutation; approval required
  delete_list:
    endpoint: DELETE /brands/{{ record.brand_id }}/lists/{{ record.id }}
    required fields: brand_id, id
    risk: permanently removes a list from a brand (contacts in the list are NOT deleted); irreversible; approval required
  create_field:
    endpoint: POST /brands/{{ record.brand_id }}/fields
    required fields: brand_id, name, type
    risk: external mutation; creates a custom contact field in a BigMailer brand; approval required
  update_field:
    endpoint: POST /brands/{{ record.brand_id }}/fields/{{ record.id }}
    required fields: brand_id, id
    risk: external mutation; approval required
  delete_field:
    endpoint: DELETE /brands/{{ record.brand_id }}/fields/{{ record.id }}
    required fields: brand_id, id
    risk: permanently removes a custom contact field from a brand; irreversible; approval required
  create_message_type:
    endpoint: POST /brands/{{ record.brand_id }}/message-types
    required fields: brand_id, name
    risk: external mutation; creates a message type (unsubscribe category) in a BigMailer brand; approval required
  update_message_type:
    endpoint: POST /brands/{{ record.brand_id }}/message-types/{{ record.id }}
    required fields: brand_id, id
    risk: external mutation; approval required
  delete_message_type:
    endpoint: DELETE /brands/{{ record.brand_id }}/message-types/{{ record.id }}
    required fields: brand_id, id
    risk: permanently removes a message type from a brand; irreversible; approval required
  create_segment:
    endpoint: POST /brands/{{ record.brand_id }}/segments
    required fields: brand_id, name, operator, conditions
    risk: external mutation; creates a contact segment in a BigMailer brand; approval required
  update_segment:
    endpoint: POST /brands/{{ record.brand_id }}/segments/{{ record.id }}
    required fields: brand_id, id
    risk: external mutation; approval required
  delete_segment:
    endpoint: DELETE /brands/{{ record.brand_id }}/segments/{{ record.id }}
    required fields: brand_id, id
    risk: permanently removes a segment from a brand; irreversible; approval required
  create_sender:
    endpoint: POST /brands/{{ record.brand_id }}/senders
    required fields: brand_id, identity
    risk: external mutation; adds a sender domain/email identity to a BigMailer brand; approval required
  update_sender:
    endpoint: POST /brands/{{ record.brand_id }}/senders/{{ record.id }}
    required fields: brand_id, id
    risk: external mutation; approval required
  delete_sender:
    endpoint: DELETE /brands/{{ record.brand_id }}/senders/{{ record.id }}
    required fields: brand_id, id
    risk: permanently removes a sender identity from a brand; irreversible; approval required
  create_template:
    endpoint: POST /brands/{{ record.brand_id }}/templates
    required fields: brand_id, name, type, html
    risk: external mutation; creates a campaign template in a BigMailer brand; approval required
  update_template:
    endpoint: POST /brands/{{ record.brand_id }}/templates/{{ record.id }}
    required fields: brand_id, id
    risk: external mutation; approval required
  delete_template:
    endpoint: DELETE /brands/{{ record.brand_id }}/templates/{{ record.id }}
    required fields: brand_id, id
    risk: permanently removes a template from a brand; irreversible; approval required
  create_user:
    endpoint: POST /users
    required fields: email, role
    risk: external mutation; invites a new user into the BigMailer account; approval required
  update_user:
    endpoint: POST /users/{{ record.id }}
    required fields: id
    risk: external mutation; approval required
  delete_user:
    endpoint: DELETE /users/{{ record.id }}
    required fields: id
    risk: permanently removes a user from the BigMailer account; irreversible; approval required

SECURITY
  read risk: external BigMailer API read of brands, account users, and brand-scoped marketing/CRM resources
  write risk: external mutation of BigMailer brands, contacts, lists, custom fields, message types, segments, senders, templates, and account users; can send real emails indirectly (e.g. via a sender/template referenced by a later campaign) but issues no send action itself
  approval: required for every write action; see writes.json risk field per action
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect bigmailer

  # Inspect as structured JSON
  pm connectors inspect bigmailer --json

AGENT WORKFLOW
  - Run pm connectors inspect bigmailer before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
