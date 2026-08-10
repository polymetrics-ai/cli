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
  api_key (secret)

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

COMMAND SURFACE
  Run BigMailer's declared streams and reverse-ETL actions.
  Usage: pm bigmailer <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    api delete v1 brands brand-id properties brand-property-id - Documented DELETE /v1/brands/{brand_id}/properties/{brand_property_id} (not implemented) [intent=direct_write availability=not_implemented operation=bigmailer.delete.v1-brands-brand-id-properties-brand-property-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get v1 brands brand-id - Documented GET /v1/brands/{brand_id} (not implemented) [intent=direct_read availability=not_implemented operation=bigmailer.get.v1-brands-brand-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 brands brand-id bulk-campaigns campaign-id - Documented GET /v1/brands/{brand_id}/bulk-campaigns/{campaign_id} (not implemented) [intent=direct_read availability=not_implemented operation=bigmailer.get.v1-brands-brand-id-bulk-campaigns-campaign-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 brands brand-id contacts batches batch-id - Documented GET /v1/brands/{brand_id}/contacts/batches/{batch_id} (not implemented) [intent=direct_read availability=not_implemented operation=bigmailer.get.v1-brands-brand-id-contacts-batches-batch-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 brands brand-id contacts contact-id - Documented GET /v1/brands/{brand_id}/contacts/{contact_id} (not implemented) [intent=direct_read availability=not_implemented operation=bigmailer.get.v1-brands-brand-id-contacts-contact-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 brands brand-id fields field-id - Documented GET /v1/brands/{brand_id}/fields/{field_id} (not implemented) [intent=direct_read availability=not_implemented operation=bigmailer.get.v1-brands-brand-id-fields-field-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 brands brand-id lists list-id - Documented GET /v1/brands/{brand_id}/lists/{list_id} (not implemented) [intent=direct_read availability=not_implemented operation=bigmailer.get.v1-brands-brand-id-lists-list-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 brands brand-id message-types message-type-id - Documented GET /v1/brands/{brand_id}/message-types/{message_type_id} (not implemented) [intent=direct_read availability=not_implemented operation=bigmailer.get.v1-brands-brand-id-message-types-message-type-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 brands brand-id properties - Documented GET /v1/brands/{brand_id}/properties (not implemented) [intent=direct_read availability=not_implemented operation=bigmailer.get.v1-brands-brand-id-properties]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 brands brand-id properties brand-property-id - Documented GET /v1/brands/{brand_id}/properties/{brand_property_id} (not implemented) [intent=direct_read availability=not_implemented operation=bigmailer.get.v1-brands-brand-id-properties-brand-property-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 brands brand-id rss-campaigns campaign-id - Documented GET /v1/brands/{brand_id}/rss-campaigns/{campaign_id} (not implemented) [intent=direct_read availability=not_implemented operation=bigmailer.get.v1-brands-brand-id-rss-campaigns-campaign-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 brands brand-id rss-campaigns parent-id updates - Documented GET /v1/brands/{brand_id}/rss-campaigns/{parent_id}/updates (not implemented) [intent=direct_read availability=not_implemented operation=bigmailer.get.v1-brands-brand-id-rss-campaigns-parent-id-updates]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 brands brand-id rss-campaigns parent-id updates campaign-id - Documented GET /v1/brands/{brand_id}/rss-campaigns/{parent_id}/updates/{campaign_id} (not implemented) [intent=direct_read availability=not_implemented operation=bigmailer.get.v1-brands-brand-id-rss-campaigns-parent-id-updates-campaign-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 brands brand-id segments segment-id - Documented GET /v1/brands/{brand_id}/segments/{segment_id} (not implemented) [intent=direct_read availability=not_implemented operation=bigmailer.get.v1-brands-brand-id-segments-segment-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 brands brand-id senders sender-id - Documented GET /v1/brands/{brand_id}/senders/{sender_id} (not implemented) [intent=direct_read availability=not_implemented operation=bigmailer.get.v1-brands-brand-id-senders-sender-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 brands brand-id suppression-lists suppression-list-id - Documented GET /v1/brands/{brand_id}/suppression-lists/{suppression_list_id} (not implemented) [intent=direct_read availability=not_implemented operation=bigmailer.get.v1-brands-brand-id-suppression-lists-suppression-list-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 brands brand-id templates template-id - Documented GET /v1/brands/{brand_id}/templates/{template_id} (not implemented) [intent=direct_read availability=not_implemented operation=bigmailer.get.v1-brands-brand-id-templates-template-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 brands brand-id test-campaigns campaign-id - Documented GET /v1/brands/{brand_id}/test-campaigns/{campaign_id} (not implemented) [intent=direct_read availability=not_implemented operation=bigmailer.get.v1-brands-brand-id-test-campaigns-campaign-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 brands brand-id transactional-campaigns campaign-id - Documented GET /v1/brands/{brand_id}/transactional-campaigns/{campaign_id} (not implemented) [intent=direct_read availability=not_implemented operation=bigmailer.get.v1-brands-brand-id-transactional-campaigns-campaign-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 connections connection-id - Documented GET /v1/connections/{connection_id} (not implemented) [intent=direct_read availability=not_implemented operation=bigmailer.get.v1-connections-connection-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 users user-id - Documented GET /v1/users/{user_id} (not implemented) [intent=direct_read availability=not_implemented operation=bigmailer.get.v1-users-user-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api post v1 brands brand-id bulk-campaigns - Documented POST /v1/brands/{brand_id}/bulk-campaigns (not implemented) [intent=direct_write availability=not_implemented operation=bigmailer.post.v1-brands-brand-id-bulk-campaigns]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 brands brand-id bulk-campaigns campaign-id - Documented POST /v1/brands/{brand_id}/bulk-campaigns/{campaign_id} (not implemented) [intent=direct_write availability=not_implemented operation=bigmailer.post.v1-brands-brand-id-bulk-campaigns-campaign-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 brands brand-id contacts batches - Documented POST /v1/brands/{brand_id}/contacts/batches (not implemented) [intent=direct_write availability=not_implemented operation=bigmailer.post.v1-brands-brand-id-contacts-batches]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 brands brand-id properties - Documented POST /v1/brands/{brand_id}/properties (not implemented) [intent=direct_write availability=not_implemented operation=bigmailer.post.v1-brands-brand-id-properties]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 brands brand-id properties brand-property-id - Documented POST /v1/brands/{brand_id}/properties/{brand_property_id} (not implemented) [intent=direct_write availability=not_implemented operation=bigmailer.post.v1-brands-brand-id-properties-brand-property-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 brands brand-id rss-campaigns - Documented POST /v1/brands/{brand_id}/rss-campaigns (not implemented) [intent=direct_write availability=not_implemented operation=bigmailer.post.v1-brands-brand-id-rss-campaigns]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 brands brand-id rss-campaigns campaign-id - Documented POST /v1/brands/{brand_id}/rss-campaigns/{campaign_id} (not implemented) [intent=direct_write availability=not_implemented operation=bigmailer.post.v1-brands-brand-id-rss-campaigns-campaign-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 brands brand-id rss-campaigns campaign-id pause - Documented POST /v1/brands/{brand_id}/rss-campaigns/{campaign_id}/pause (not implemented) [intent=direct_write availability=not_implemented operation=bigmailer.post.v1-brands-brand-id-rss-campaigns-campaign-id-pause]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 brands brand-id rss-campaigns campaign-id unpause - Documented POST /v1/brands/{brand_id}/rss-campaigns/{campaign_id}/unpause (not implemented) [intent=direct_write availability=not_implemented operation=bigmailer.post.v1-brands-brand-id-rss-campaigns-campaign-id-unpause]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 brands brand-id senders sender-id bounce-domains - Documented POST /v1/brands/{brand_id}/senders/{sender_id}/bounce-domains (not implemented) [intent=direct_write availability=not_implemented operation=bigmailer.post.v1-brands-brand-id-senders-sender-id-bounce-domains]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 brands brand-id senders sender-id bounce-domains verify - Documented POST /v1/brands/{brand_id}/senders/{sender_id}/bounce-domains/verify (not implemented) [intent=direct_write availability=not_implemented operation=bigmailer.post.v1-brands-brand-id-senders-sender-id-bounce-domains-verify]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 brands brand-id senders sender-id verify - Documented POST /v1/brands/{brand_id}/senders/{sender_id}/verify (not implemented) [intent=direct_write availability=not_implemented operation=bigmailer.post.v1-brands-brand-id-senders-sender-id-verify]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 brands brand-id suppression-lists - Documented POST /v1/brands/{brand_id}/suppression-lists (not implemented) [intent=direct_write availability=not_implemented operation=bigmailer.post.v1-brands-brand-id-suppression-lists]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 brands brand-id test-campaigns - Documented POST /v1/brands/{brand_id}/test-campaigns (not implemented) [intent=direct_write availability=not_implemented operation=bigmailer.post.v1-brands-brand-id-test-campaigns]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 brands brand-id transactional-campaigns - Documented POST /v1/brands/{brand_id}/transactional-campaigns (not implemented) [intent=direct_write availability=not_implemented operation=bigmailer.post.v1-brands-brand-id-transactional-campaigns]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 brands brand-id transactional-campaigns campaign-id - Documented POST /v1/brands/{brand_id}/transactional-campaigns/{campaign_id} (not implemented) [intent=direct_write availability=not_implemented operation=bigmailer.post.v1-brands-brand-id-transactional-campaigns-campaign-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 brands brand-id transactional-campaigns campaign-id send - Documented POST /v1/brands/{brand_id}/transactional-campaigns/{campaign_id}/send (not implemented) [intent=direct_write availability=not_implemented operation=bigmailer.post.v1-brands-brand-id-transactional-campaigns-campaign-id-send]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    brands list - Run the brands ETL stream [intent=etl availability=implemented stream=brands]
    bulk campaigns list - Run the bulk campaigns ETL stream [intent=etl availability=implemented stream=bulk_campaigns]
    connections list - Run the connections ETL stream [intent=etl availability=implemented stream=connections]
    contacts list - Run the contacts ETL stream [intent=etl availability=implemented stream=contacts]
    create brand apply - Plan and execute the create brand reverse-ETL action [intent=reverse_etl availability=implemented write=create_brand]; approval: requires plan, preview, approval, and execute; risk: external mutation; creates a new BigMailer brand (sending identity); approval required; flags: --connection_id (required), --from_email (required), --from_name (required), --name (required)
    create contact apply - Plan and execute the create contact reverse-ETL action [intent=reverse_etl availability=implemented write=create_contact]; approval: requires plan, preview, approval, and execute; risk: external mutation; creates a contact in a BigMailer brand; approval required; flags: --brand_id (required), --email (required)
    create field apply - Plan and execute the create field reverse-ETL action [intent=reverse_etl availability=implemented write=create_field]; approval: requires plan, preview, approval, and execute; risk: external mutation; creates a custom contact field in a BigMailer brand; approval required; flags: --brand_id (required), --name (required), --type (required)
    create list apply - Plan and execute the create list reverse-ETL action [intent=reverse_etl availability=implemented write=create_list]; approval: requires plan, preview, approval, and execute; risk: external mutation; creates a contact list in a BigMailer brand; approval required; flags: --brand_id (required), --name (required)
    create message type apply - Plan and execute the create message type reverse-ETL action [intent=reverse_etl availability=implemented write=create_message_type]; approval: requires plan, preview, approval, and execute; risk: external mutation; creates a message type (unsubscribe category) in a BigMailer brand; approval required; flags: --brand_id (required), --name (required)
    create segment apply - Plan and execute the create segment reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_segment]; approval: requires plan, preview, approval, and execute; risk: external mutation; creates a contact segment in a BigMailer brand; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create sender apply - Plan and execute the create sender reverse-ETL action [intent=reverse_etl availability=implemented write=create_sender]; approval: requires plan, preview, approval, and execute; risk: external mutation; adds a sender domain/email identity to a BigMailer brand; approval required; flags: --brand_id (required), --identity (required)
    create template apply - Plan and execute the create template reverse-ETL action [intent=reverse_etl availability=implemented write=create_template]; approval: requires plan, preview, approval, and execute; risk: external mutation; creates a campaign template in a BigMailer brand; approval required; flags: --brand_id (required), --html (required), --name (required), --type (required)
    create user apply - Plan and execute the create user reverse-ETL action [intent=reverse_etl availability=implemented write=create_user]; approval: requires plan, preview, approval, and execute; risk: external mutation; invites a new user into the BigMailer account; approval required; flags: --email (required), --role (required)
    delete contact apply - Plan and execute the delete contact reverse-ETL action [intent=reverse_etl availability=implemented write=delete_contact]; approval: requires plan, preview, approval, and execute; risk: permanently removes a contact from a brand; irreversible; approval required; flags: --brand_id (required), --id (required)
    delete field apply - Plan and execute the delete field reverse-ETL action [intent=reverse_etl availability=implemented write=delete_field]; approval: requires plan, preview, approval, and execute; risk: permanently removes a custom contact field from a brand; irreversible; approval required; flags: --brand_id (required), --id (required)
    delete list apply - Plan and execute the delete list reverse-ETL action [intent=reverse_etl availability=implemented write=delete_list]; approval: requires plan, preview, approval, and execute; risk: permanently removes a list from a brand (contacts in the list are NOT deleted); irreversible; approval required; flags: --brand_id (required), --id (required)
    delete message type apply - Plan and execute the delete message type reverse-ETL action [intent=reverse_etl availability=implemented write=delete_message_type]; approval: requires plan, preview, approval, and execute; risk: permanently removes a message type from a brand; irreversible; approval required; flags: --brand_id (required), --id (required)
    delete segment apply - Plan and execute the delete segment reverse-ETL action [intent=reverse_etl availability=implemented write=delete_segment]; approval: requires plan, preview, approval, and execute; risk: permanently removes a segment from a brand; irreversible; approval required; flags: --brand_id (required), --id (required)
    delete sender apply - Plan and execute the delete sender reverse-ETL action [intent=reverse_etl availability=implemented write=delete_sender]; approval: requires plan, preview, approval, and execute; risk: permanently removes a sender identity from a brand; irreversible; approval required; flags: --brand_id (required), --id (required)
    delete template apply - Plan and execute the delete template reverse-ETL action [intent=reverse_etl availability=implemented write=delete_template]; approval: requires plan, preview, approval, and execute; risk: permanently removes a template from a brand; irreversible; approval required; flags: --brand_id (required), --id (required)
    delete user apply - Plan and execute the delete user reverse-ETL action [intent=reverse_etl availability=implemented write=delete_user]; approval: requires plan, preview, approval, and execute; risk: permanently removes a user from the BigMailer account; irreversible; approval required; flags: --id (required)
    fields list - Run the fields ETL stream [intent=etl availability=implemented stream=fields]
    lists list - Run the lists ETL stream [intent=etl availability=implemented stream=lists]
    message types list - Run the message types ETL stream [intent=etl availability=implemented stream=message_types]
    rss campaigns list - Run the rss campaigns ETL stream [intent=etl availability=implemented stream=rss_campaigns]
    segments list - Run the segments ETL stream [intent=etl availability=implemented stream=segments]
    senders list - Run the senders ETL stream [intent=etl availability=implemented stream=senders]
    suppression lists list - Run the suppression lists ETL stream [intent=etl availability=implemented stream=suppression_lists]
    templates list - Run the templates ETL stream [intent=etl availability=implemented stream=templates]
    test campaigns list - Run the test campaigns ETL stream [intent=etl availability=implemented stream=test_campaigns]
    transactional campaigns list - Run the transactional campaigns ETL stream [intent=etl availability=implemented stream=transactional_campaigns]
    update brand apply - Plan and execute the update brand reverse-ETL action [intent=reverse_etl availability=implemented write=update_brand]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --id (required)
    update contact apply - Plan and execute the update contact reverse-ETL action [intent=reverse_etl availability=implemented write=update_contact]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --brand_id (required), --id (required)
    update field apply - Plan and execute the update field reverse-ETL action [intent=reverse_etl availability=implemented write=update_field]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --brand_id (required), --id (required)
    update list apply - Plan and execute the update list reverse-ETL action [intent=reverse_etl availability=implemented write=update_list]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --brand_id (required), --id (required)
    update message type apply - Plan and execute the update message type reverse-ETL action [intent=reverse_etl availability=implemented write=update_message_type]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --brand_id (required), --id (required)
    update segment apply - Plan and execute the update segment reverse-ETL action [intent=reverse_etl availability=implemented write=update_segment]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --brand_id (required), --id (required)
    update sender apply - Plan and execute the update sender reverse-ETL action [intent=reverse_etl availability=implemented write=update_sender]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --brand_id (required), --id (required)
    update template apply - Plan and execute the update template reverse-ETL action [intent=reverse_etl availability=implemented write=update_template]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --brand_id (required), --id (required)
    update user apply - Plan and execute the update user reverse-ETL action [intent=reverse_etl availability=implemented write=update_user]; approval: requires plan, preview, approval, and execute; risk: external mutation; approval required; flags: --id (required)
    upsert contact apply - Plan and execute the upsert contact reverse-ETL action [intent=reverse_etl availability=implemented write=upsert_contact]; approval: requires plan, preview, approval, and execute; risk: external mutation; creates the contact if the email is new, otherwise updates the existing contact; approval required; flags: --brand_id (required), --email (required)
    users list - Run the users ETL stream [intent=etl availability=implemented stream=users]

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
