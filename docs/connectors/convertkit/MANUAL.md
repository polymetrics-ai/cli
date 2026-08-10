# pm connectors inspect convertkit

```text
NAME
  pm connectors inspect convertkit - ConvertKit connector manual

SYNOPSIS
  pm connectors inspect convertkit
  pm connectors inspect convertkit --json
  pm credentials add <name> --connector convertkit [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads ConvertKit (Kit) subscribers, forms, sequences, tags, broadcasts, custom fields, and purchases, and writes subscriber/tag/form/sequence/broadcast/custom-field/purchase/webhook mutations, through the ConvertKit v3 REST API.

ICON
  id: convertkit
  asset: icons/convertkit.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.convertkit.com/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  access_token (secret)
  api_key (secret)
  api_secret (secret)

ETL STREAMS
  subscribers:
    primary key: id
    cursor: created_at
    fields: created_at(string), email_address(string), first_name(string), id(integer), state(string)
  forms:
    primary key: id
    cursor: created_at
    fields: archived(boolean), created_at(string), format(string), id(integer), name(string), type(string)
  sequences:
    primary key: id
    cursor: created_at
    fields: created_at(string), hold(boolean), id(integer), name(string), repeat(boolean)
  tags:
    primary key: id
    cursor: created_at
    fields: created_at(string), id(integer), name(string)
  broadcasts:
    primary key: id
    cursor: created_at
    fields: created_at(string), description(string), id(integer), public(boolean), published_at(string), subject(string)
  custom_fields:
    primary key: id
    fields: id(integer), key(string), label(string), name(string)
  purchases:
    primary key: id
    cursor: transaction_time
    fields: currency(string), discount(string), email_address(string), id(integer), shipping(string), status(string), subtotal(string), tax(string), total(string), transaction_id(string), transaction_time(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  update_subscriber:
    endpoint: PUT /subscribers/{{ record.id }}
    required fields: id
    risk: mutates an existing subscriber's name/email/custom-field values; external mutation, no approval required
  create_tag:
    endpoint: POST /tags
    required fields: name
    risk: creates a new tag on the account; low-risk external mutation, no approval required
  tag_subscriber:
    endpoint: POST /tags/{{ record.tag_id }}/subscribe
    required fields: tag_id, email
    risk: applies a tag to a subscriber (creating the subscriber if the email is new); external mutation, no approval required
  remove_tag_from_subscriber:
    endpoint: DELETE /subscribers/{{ record.subscriber_id }}/tags/{{ record.tag_id }}
    required fields: subscriber_id, tag_id
    risk: removes a tag from a subscriber; external mutation, no approval required
  subscribe_to_form:
    endpoint: POST /forms/{{ record.form_id }}/subscribe
    required fields: form_id, email
    risk: subscribes an email address to a form (creating the subscriber if the email is new); external mutation, no approval required
  subscribe_to_sequence:
    endpoint: POST /sequences/{{ record.sequence_id }}/subscribe
    required fields: sequence_id, email
    risk: subscribes an email address to a sequence (creating the subscriber if the email is new); external mutation, no approval required
  create_broadcast:
    endpoint: POST /broadcasts
    required fields: subject, content
    risk: creates a draft or scheduled email broadcast; a scheduled broadcast (send_at/published_at set) will send to the account's live subscriber list, external mutation, approval required
  update_broadcast:
    endpoint: PUT /broadcasts/{{ record.id }}
    required fields: id
    risk: mutates a draft or scheduled broadcast's content/send time; external mutation, approval required
  delete_broadcast:
    endpoint: DELETE /broadcasts/{{ record.id }}
    required fields: id
    risk: permanently deletes a draft or scheduled broadcast record; irreversible, approval required
  create_custom_field:
    endpoint: POST /custom_fields
    required fields: label
    risk: creates a new custom subscriber field on the account (up to 140 total); low-risk external mutation, no approval required
  update_custom_field:
    endpoint: PUT /custom_fields/{{ record.id }}
    required fields: id, label
    risk: renames a custom field's label (the underlying key is unchanged per Kit's own docs); external mutation, no approval required
  create_purchase:
    endpoint: POST /purchases
    required fields: purchase
    risk: records a new purchase-tracking transaction for a subscriber; external mutation, no approval required
  create_webhook:
    endpoint: POST /automations/hooks
    required fields: target_url, event
    risk: creates a webhook that POSTs subscriber-event payloads to an external URL the caller controls; external mutation, approval required
  delete_webhook:
    endpoint: DELETE /automations/hooks/{{ record.rule_id }}
    required fields: rule_id
    risk: permanently deletes a webhook automation; irreversible, approval required

SECURITY
  read risk: external ConvertKit API read of subscriber and campaign data
  write risk: external mutation: creates/updates subscribers, tags, forms/sequences subscriptions, broadcasts (including scheduling live sends), custom fields, purchase records, and webhooks; deletes are limited to broadcasts/webhooks/tag-removal (no subscriber/custom-field/global-unsubscribe deletes)
  approval: required for all write actions; read-only otherwise
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run ConvertKit's declared streams and reverse-ETL actions.
  Usage: pm convertkit <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    api delete v3 custom-fields custom-field-id - Documented DELETE /v3/custom_fields/{custom_field_id} (not implemented) [intent=direct_write availability=not_implemented operation=convertkit.delete.v3-custom-fields-custom-field-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get v3 account - Documented GET /v3/account (not implemented) [intent=direct_read availability=not_implemented operation=convertkit.get.v3-account]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 broadcasts broadcast-id - Documented GET /v3/broadcasts/{broadcast_id} (not implemented) [intent=direct_read availability=not_implemented operation=convertkit.get.v3-broadcasts-broadcast-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 broadcasts broadcast-id stats - Documented GET /v3/broadcasts/{broadcast_id}/stats (not implemented) [intent=direct_read availability=not_implemented operation=convertkit.get.v3-broadcasts-broadcast-id-stats]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 forms form-id subscriptions - Documented GET /v3/forms/{form_id}/subscriptions (not implemented) [intent=direct_read availability=not_implemented operation=convertkit.get.v3-forms-form-id-subscriptions]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 purchases purchase-id - Documented GET /v3/purchases/{purchase_id} (not implemented) [intent=direct_read availability=not_implemented operation=convertkit.get.v3-purchases-purchase-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 sequences sequence-id subscriptions - Documented GET /v3/sequences/{sequence_id}/subscriptions (not implemented) [intent=direct_read availability=not_implemented operation=convertkit.get.v3-sequences-sequence-id-subscriptions]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 subscribers subscriber-id - Documented GET /v3/subscribers/{subscriber_id} (not implemented) [intent=direct_read availability=not_implemented operation=convertkit.get.v3-subscribers-subscriber-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 subscribers subscriber-id tags - Documented GET /v3/subscribers/{subscriber_id}/tags (not implemented) [intent=direct_read availability=not_implemented operation=convertkit.get.v3-subscribers-subscriber-id-tags]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 tags tag-id subscriptions - Documented GET /v3/tags/{tag_id}/subscriptions (not implemented) [intent=direct_read availability=not_implemented operation=convertkit.get.v3-tags-tag-id-subscriptions]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api post v3 tags tag-id unsubscribe - Documented POST /v3/tags/{tag_id}/unsubscribe (not implemented) [intent=direct_write availability=not_implemented operation=convertkit.post.v3-tags-tag-id-unsubscribe]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put v3 unsubscribe - Documented PUT /v3/unsubscribe (not implemented) [intent=direct_write availability=not_implemented operation=convertkit.put.v3-unsubscribe]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    broadcasts list - Run the broadcasts ETL stream [intent=etl availability=implemented stream=broadcasts]
    create broadcast apply - Plan and execute the create broadcast reverse-ETL action [intent=reverse_etl availability=implemented write=create_broadcast]; approval: requires plan, preview, approval, and execute; risk: creates a draft or scheduled email broadcast; a scheduled broadcast (send_at/published_at set) will send to the account's live subscriber list, external mutation, approval required; flags: --content (required), --subject (required)
    create custom field apply - Plan and execute the create custom field reverse-ETL action [intent=reverse_etl availability=implemented write=create_custom_field]; approval: requires plan, preview, approval, and execute; risk: creates a new custom subscriber field on the account (up to 140 total); low-risk external mutation, no approval required; flags: --label (required)
    create purchase apply - Plan and execute the create purchase reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_purchase]; approval: requires plan, preview, approval, and execute; risk: records a new purchase-tracking transaction for a subscriber; external mutation, no approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create tag apply - Plan and execute the create tag reverse-ETL action [intent=reverse_etl availability=implemented write=create_tag]; approval: requires plan, preview, approval, and execute; risk: creates a new tag on the account; low-risk external mutation, no approval required; flags: --name (required)
    create webhook apply - Plan and execute the create webhook reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_webhook]; approval: requires plan, preview, approval, and execute; risk: creates a webhook that POSTs subscriber-event payloads to an external URL the caller controls; external mutation, approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    custom fields list - Run the custom fields ETL stream [intent=etl availability=implemented stream=custom_fields]
    delete broadcast apply - Plan and execute the delete broadcast reverse-ETL action [intent=reverse_etl availability=implemented write=delete_broadcast]; approval: requires plan, preview, approval, and execute; risk: permanently deletes a draft or scheduled broadcast record; irreversible, approval required; flags: --id (required)
    delete webhook apply - Plan and execute the delete webhook reverse-ETL action [intent=reverse_etl availability=implemented write=delete_webhook]; approval: requires plan, preview, approval, and execute; risk: permanently deletes a webhook automation; irreversible, approval required; flags: --rule_id (required)
    forms list - Run the forms ETL stream [intent=etl availability=implemented stream=forms]
    purchases list - Run the purchases ETL stream [intent=etl availability=implemented stream=purchases]
    remove tag from subscriber apply - Plan and execute the remove tag from subscriber reverse-ETL action [intent=reverse_etl availability=implemented write=remove_tag_from_subscriber]; approval: requires plan, preview, approval, and execute; risk: removes a tag from a subscriber; external mutation, no approval required; flags: --subscriber_id (required), --tag_id (required)
    sequences list - Run the sequences ETL stream [intent=etl availability=implemented stream=sequences]
    subscribe to form apply - Plan and execute the subscribe to form reverse-ETL action [intent=reverse_etl availability=implemented write=subscribe_to_form]; approval: requires plan, preview, approval, and execute; risk: subscribes an email address to a form (creating the subscriber if the email is new); external mutation, no approval required; flags: --email (required), --form_id (required)
    subscribe to sequence apply - Plan and execute the subscribe to sequence reverse-ETL action [intent=reverse_etl availability=implemented write=subscribe_to_sequence]; approval: requires plan, preview, approval, and execute; risk: subscribes an email address to a sequence (creating the subscriber if the email is new); external mutation, no approval required; flags: --email (required), --sequence_id (required)
    subscribers list - Run the subscribers ETL stream [intent=etl availability=implemented stream=subscribers]
    tag subscriber apply - Plan and execute the tag subscriber reverse-ETL action [intent=reverse_etl availability=implemented write=tag_subscriber]; approval: requires plan, preview, approval, and execute; risk: applies a tag to a subscriber (creating the subscriber if the email is new); external mutation, no approval required; flags: --email (required), --tag_id (required)
    tags list - Run the tags ETL stream [intent=etl availability=implemented stream=tags]
    update broadcast apply - Plan and execute the update broadcast reverse-ETL action [intent=reverse_etl availability=implemented write=update_broadcast]; approval: requires plan, preview, approval, and execute; risk: mutates a draft or scheduled broadcast's content/send time; external mutation, approval required; flags: --id (required)
    update custom field apply - Plan and execute the update custom field reverse-ETL action [intent=reverse_etl availability=implemented write=update_custom_field]; approval: requires plan, preview, approval, and execute; risk: renames a custom field's label (the underlying key is unchanged per Kit's own docs); external mutation, no approval required; flags: --id (required), --label (required)
    update subscriber apply - Plan and execute the update subscriber reverse-ETL action [intent=reverse_etl availability=implemented write=update_subscriber]; approval: requires plan, preview, approval, and execute; risk: mutates an existing subscriber's name/email/custom-field values; external mutation, no approval required; flags: --id (required)

EXAMPLES
  # Inspect as a manual
  pm connectors inspect convertkit

  # Inspect as structured JSON
  pm connectors inspect convertkit --json

AGENT WORKFLOW
  - Run pm connectors inspect convertkit before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
