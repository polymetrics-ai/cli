# pm connectors inspect postmarkapp

```text
NAME
  pm connectors inspect postmarkapp - Postmark App connector manual

SYNOPSIS
  pm connectors inspect postmarkapp
  pm connectors inspect postmarkapp --json
  pm credentials add <name> --connector postmarkapp [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Postmark server-token API resources including messages, bounces, templates, message streams, stats, webhooks, suppressions, and inbound rules; exposes server-token write actions for sends and resource mutations.

ICON
  id: postmark
  asset: icons/postmark.svg
  source: official
  review_status: official_verified
  review_url: https://postmarkapp.com/developer

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  bounce_id
  bulk_request_id
  message_id
  message_stream_id
  mode
  template_id_or_alias
  webhook_id
  X-Postmark-Server-Token (secret)

ETL STREAMS
  outbound_messages:
    primary key: id
    cursor: received_at
    fields: from(string), id(string), received_at(string), status(string), subject(string), to(array)
  inbound_messages:
    primary key: id
    fields: from(string), id(string), status(string), subject(string), to(string)
  current_server:
    primary key: ID
    fields: ID(string)
  bulk_email_status:
    primary key: Id
    fields: Id(string)
  delivery_stats:
  bounces:
    primary key: ID
    fields: ID(string)
  bounce:
    primary key: ID
    fields: ID(string)
  bounce_dump:
  templates:
    primary key: TemplateId
    fields: TemplateId(string)
  template:
    primary key: TemplateId
    fields: TemplateId(string)
  message_streams:
    primary key: ID
    fields: ID(string)
  message_stream:
    primary key: ID
    fields: ID(string)
  outbound_message_details:
    primary key: MessageID
    fields: MessageID(string)
  outbound_message_dump:
  inbound_message_details:
    primary key: MessageID
    fields: MessageID(string)
  outbound_message_opens:
  outbound_message_opens_by_message:
  outbound_message_clicks:
  outbound_message_clicks_by_message:
  stats_outbound:
  stats_outbound_sends:
  stats_outbound_bounces:
  stats_outbound_spam:
  stats_outbound_tracked:
  stats_outbound_opens:
  stats_outbound_open_platforms:
  stats_outbound_email_clients:
  stats_outbound_clicks:
  stats_outbound_click_browser_families:
  stats_outbound_click_platforms:
  stats_outbound_click_location:
  inbound_rule_triggers:
    primary key: ID
    fields: ID(string)
  webhooks:
    primary key: ID
    fields: ID(string)
  webhook:
    primary key: ID
    fields: ID(string)
  suppressions:

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  send_email:
    endpoint: POST /email
    required fields: From, To, Subject
    risk: sends a live Postmark email; approval required
  send_bulk_email:
    endpoint: POST /email/bulk
    required fields: From, Subject, Messages
    risk: submits a live Postmark bulk email request; approval required
  send_email_with_template:
    endpoint: POST /email/withTemplate
    required fields: From, To
    risk: sends a live Postmark template email; approval required
  edit_current_server:
    endpoint: PUT /server
    risk: mutates the current Postmark server settings; approval required
  activate_bounce:
    endpoint: PUT /bounces/{{ record.bounce_id }}/activate
    required fields: bounce_id
    risk: reactivates a bounced email address in Postmark; approval required
  create_template:
    endpoint: POST /templates
    required fields: Name
    risk: creates a Postmark template; approval required
  edit_template:
    endpoint: PUT /templates/{{ record.template_id_or_alias }}
    required fields: template_id_or_alias
    risk: updates a Postmark template; approval required
  delete_template:
    endpoint: DELETE /templates/{{ record.template_id_or_alias }}
    required fields: template_id_or_alias
    risk: deletes a Postmark template; destructive external mutation
  validate_template:
    endpoint: POST /templates/validate
    risk: validates Postmark template content; no persistent mutation expected but still invokes the external API
  create_message_stream:
    endpoint: POST /message-streams
    required fields: ID, Name, MessageStreamType
    risk: creates a Postmark message stream; approval required
  edit_message_stream:
    endpoint: PATCH /message-streams/{{ record.message_stream_id }}
    required fields: message_stream_id
    risk: updates a Postmark message stream; approval required
  archive_message_stream:
    endpoint: POST /message-streams/{{ record.message_stream_id }}/archive
    required fields: message_stream_id
    risk: archives a Postmark message stream; approval required
  unarchive_message_stream:
    endpoint: POST /message-streams/{{ record.message_stream_id }}/unarchive
    required fields: message_stream_id
    risk: unarchives a Postmark message stream; approval required
  bypass_inbound_message:
    endpoint: PUT /messages/inbound/{{ record.message_id }}/bypass
    required fields: message_id
    risk: bypasses inbound blocking for one Postmark message; approval required
  retry_inbound_message:
    endpoint: PUT /messages/inbound/{{ record.message_id }}/retry
    required fields: message_id
    risk: retries processing for one inbound Postmark message; approval required
  create_inbound_rule_trigger:
    endpoint: POST /triggers/inboundrules
    required fields: Rule
    risk: creates an inbound rule trigger; approval required
  delete_inbound_rule_trigger:
    endpoint: DELETE /triggers/inboundrules/{{ record.trigger_id }}
    required fields: trigger_id
    risk: deletes an inbound rule trigger; destructive external mutation
  create_webhook:
    endpoint: POST /webhooks
    required fields: Url
    risk: creates a Postmark webhook endpoint; approval required
  edit_webhook:
    endpoint: PUT /webhooks/{{ record.webhook_id }}
    required fields: webhook_id
    risk: updates a Postmark webhook endpoint; approval required
  delete_webhook:
    endpoint: DELETE /webhooks/{{ record.webhook_id }}
    required fields: webhook_id
    risk: deletes a Postmark webhook endpoint; destructive external mutation
  create_suppression:
    endpoint: POST /message-streams/{{ record.message_stream_id }}/suppressions
    required fields: message_stream_id, Suppressions
    risk: adds one or more suppressions to a Postmark message stream; approval required
  delete_suppression:
    endpoint: POST /message-streams/{{ record.message_stream_id }}/suppressions/delete
    required fields: message_stream_id, Suppressions
    risk: removes one or more suppressions from a Postmark message stream; approval required

SECURITY
  read risk: external Postmark API read of message, bounce, template, stream, stats, webhook, suppression, and inbound-rule data
  write risk: sends emails and mutates Postmark templates, message streams, server settings, webhooks, inbound rules, suppressions, and inbound message processing state
  approval: reverse ETL writes require plan preview and approval token
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Postmark App's declared streams and reverse-ETL actions.
  Usage: pm postmarkapp <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    activate bounce apply - Plan and execute the activate bounce reverse-ETL action [intent=reverse_etl availability=not_implemented write=activate_bounce]; approval: requires plan, preview, approval, and execute; risk: reactivates a bounced email address in Postmark; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    api delete domains domainid - Documented DELETE /domains/{domainid} (not implemented) [intent=direct_write availability=not_implemented operation=postmarkapp.delete.domains-domainid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete senders signatureid - Documented DELETE /senders/{signatureid} (not implemented) [intent=direct_write availability=not_implemented operation=postmarkapp.delete.senders-signatureid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete servers serverid - Documented DELETE /servers/{serverid} (not implemented) [intent=direct_write availability=not_implemented operation=postmarkapp.delete.servers-serverid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get data-removals id - Documented GET /data-removals/{id} (not implemented) [intent=direct_read availability=not_implemented operation=postmarkapp.get.data-removals-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get domains - Documented GET /domains (not implemented) [intent=direct_read availability=not_implemented operation=postmarkapp.get.domains]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get domains domainid - Documented GET /domains/{domainid} (not implemented) [intent=direct_read availability=not_implemented operation=postmarkapp.get.domains-domainid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get senders - Documented GET /senders (not implemented) [intent=direct_read availability=not_implemented operation=postmarkapp.get.senders]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get senders signatureid - Documented GET /senders/{signatureid} (not implemented) [intent=direct_read availability=not_implemented operation=postmarkapp.get.senders-signatureid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get servers - Documented GET /servers (not implemented) [intent=direct_read availability=not_implemented operation=postmarkapp.get.servers]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get servers serverid - Documented GET /servers/{serverid} (not implemented) [intent=direct_read availability=not_implemented operation=postmarkapp.get.servers-serverid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api post data-removals - Documented POST /data-removals (not implemented) [intent=direct_write availability=not_implemented operation=postmarkapp.post.data-removals]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post domains - Documented POST /domains (not implemented) [intent=direct_write availability=not_implemented operation=postmarkapp.post.domains]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post domains domainid rotatedkim - Documented POST /domains/{domainid}/rotatedkim (not implemented) [intent=direct_write availability=not_implemented operation=postmarkapp.post.domains-domainid-rotatedkim]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post domains domainid verifyspf - Documented POST /domains/{domainid}/verifyspf (not implemented) [intent=direct_write availability=not_implemented operation=postmarkapp.post.domains-domainid-verifyspf]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post email batch - Documented POST /email/batch (not implemented) [intent=direct_write availability=not_implemented operation=postmarkapp.post.email-batch]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post email batchwithtemplates - Documented POST /email/batchWithTemplates (not implemented) [intent=direct_write availability=not_implemented operation=postmarkapp.post.email-batchwithtemplates]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post senders - Documented POST /senders (not implemented) [intent=direct_write availability=not_implemented operation=postmarkapp.post.senders]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post senders signatureid requestnewdkim - Documented POST /senders/{signatureid}/requestnewdkim (not implemented) [intent=direct_write availability=not_implemented operation=postmarkapp.post.senders-signatureid-requestnewdkim]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post senders signatureid resend - Documented POST /senders/{signatureid}/resend (not implemented) [intent=direct_write availability=not_implemented operation=postmarkapp.post.senders-signatureid-resend]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post senders signatureid verifyspf - Documented POST /senders/{signatureid}/verifyspf (not implemented) [intent=direct_write availability=not_implemented operation=postmarkapp.post.senders-signatureid-verifyspf]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post servers - Documented POST /servers (not implemented) [intent=direct_write availability=not_implemented operation=postmarkapp.post.servers]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put domains domainid - Documented PUT /domains/{domainid} (not implemented) [intent=direct_write availability=not_implemented operation=postmarkapp.put.domains-domainid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put domains domainid verifydkim - Documented PUT /domains/{domainid}/verifyDkim (not implemented) [intent=direct_write availability=not_implemented operation=postmarkapp.put.domains-domainid-verifydkim]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put domains domainid verifyreturnpath - Documented PUT /domains/{domainid}/verifyReturnPath (not implemented) [intent=direct_write availability=not_implemented operation=postmarkapp.put.domains-domainid-verifyreturnpath]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put senders signatureid - Documented PUT /senders/{signatureid} (not implemented) [intent=direct_write availability=not_implemented operation=postmarkapp.put.senders-signatureid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put servers serverid - Documented PUT /servers/{serverid} (not implemented) [intent=direct_write availability=not_implemented operation=postmarkapp.put.servers-serverid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put templates push - Documented PUT /templates/push (not implemented) [intent=direct_write availability=not_implemented operation=postmarkapp.put.templates-push]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    archive message stream apply - Plan and execute the archive message stream reverse-ETL action [intent=reverse_etl availability=not_implemented write=archive_message_stream]; approval: requires plan, preview, approval, and execute; risk: archives a Postmark message stream; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    bounce dump list - Run the bounce dump ETL stream [intent=etl availability=implemented stream=bounce_dump]
    bounce list - Run the bounce ETL stream [intent=etl availability=implemented stream=bounce]
    bounces list - Run the bounces ETL stream [intent=etl availability=implemented stream=bounces]
    bulk email status list - Run the bulk email status ETL stream [intent=etl availability=implemented stream=bulk_email_status]
    bypass inbound message apply - Plan and execute the bypass inbound message reverse-ETL action [intent=reverse_etl availability=not_implemented write=bypass_inbound_message]; approval: requires plan, preview, approval, and execute; risk: bypasses inbound blocking for one Postmark message; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create inbound rule trigger apply - Plan and execute the create inbound rule trigger reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_inbound_rule_trigger]; approval: requires plan, preview, approval, and execute; risk: creates an inbound rule trigger; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create message stream apply - Plan and execute the create message stream reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_message_stream]; approval: requires plan, preview, approval, and execute; risk: creates a Postmark message stream; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create suppression apply - Plan and execute the create suppression reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_suppression]; approval: requires plan, preview, approval, and execute; risk: adds one or more suppressions to a Postmark message stream; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create template apply - Plan and execute the create template reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_template]; approval: requires plan, preview, approval, and execute; risk: creates a Postmark template; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create webhook apply - Plan and execute the create webhook reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_webhook]; approval: requires plan, preview, approval, and execute; risk: creates a Postmark webhook endpoint; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    current server list - Run the current server ETL stream [intent=etl availability=implemented stream=current_server]
    delete inbound rule trigger apply - Plan and execute the delete inbound rule trigger reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_inbound_rule_trigger]; approval: requires plan, preview, approval, and execute; risk: deletes an inbound rule trigger; destructive external mutation; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete suppression apply - Plan and execute the delete suppression reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_suppression]; approval: requires plan, preview, approval, and execute; risk: removes one or more suppressions from a Postmark message stream; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete template apply - Plan and execute the delete template reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_template]; approval: requires plan, preview, approval, and execute; risk: deletes a Postmark template; destructive external mutation; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete webhook apply - Plan and execute the delete webhook reverse-ETL action [intent=reverse_etl availability=not_implemented write=delete_webhook]; approval: requires plan, preview, approval, and execute; risk: deletes a Postmark webhook endpoint; destructive external mutation; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delivery stats list - Run the delivery stats ETL stream [intent=etl availability=implemented stream=delivery_stats]
    edit current server apply - Plan and execute the edit current server reverse-ETL action [intent=reverse_etl availability=implemented write=edit_current_server]; approval: requires plan, preview, approval, and execute; risk: mutates the current Postmark server settings; approval required
    edit message stream apply - Plan and execute the edit message stream reverse-ETL action [intent=reverse_etl availability=not_implemented write=edit_message_stream]; approval: requires plan, preview, approval, and execute; risk: updates a Postmark message stream; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    edit template apply - Plan and execute the edit template reverse-ETL action [intent=reverse_etl availability=not_implemented write=edit_template]; approval: requires plan, preview, approval, and execute; risk: updates a Postmark template; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    edit webhook apply - Plan and execute the edit webhook reverse-ETL action [intent=reverse_etl availability=not_implemented write=edit_webhook]; approval: requires plan, preview, approval, and execute; risk: updates a Postmark webhook endpoint; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    inbound message details list - Run the inbound message details ETL stream [intent=etl availability=implemented stream=inbound_message_details]
    inbound messages list - Run the inbound messages ETL stream [intent=etl availability=implemented stream=inbound_messages]
    inbound rule triggers list - Run the inbound rule triggers ETL stream [intent=etl availability=implemented stream=inbound_rule_triggers]
    message stream list - Run the message stream ETL stream [intent=etl availability=implemented stream=message_stream]
    message streams list - Run the message streams ETL stream [intent=etl availability=implemented stream=message_streams]
    outbound message clicks by message list - Run the outbound message clicks by message ETL stream [intent=etl availability=implemented stream=outbound_message_clicks_by_message]
    outbound message clicks list - Run the outbound message clicks ETL stream [intent=etl availability=implemented stream=outbound_message_clicks]
    outbound message details list - Run the outbound message details ETL stream [intent=etl availability=implemented stream=outbound_message_details]
    outbound message dump list - Run the outbound message dump ETL stream [intent=etl availability=implemented stream=outbound_message_dump]
    outbound message opens by message list - Run the outbound message opens by message ETL stream [intent=etl availability=implemented stream=outbound_message_opens_by_message]
    outbound message opens list - Run the outbound message opens ETL stream [intent=etl availability=implemented stream=outbound_message_opens]
    outbound messages list - Run the outbound messages ETL stream [intent=etl availability=implemented stream=outbound_messages]
    retry inbound message apply - Plan and execute the retry inbound message reverse-ETL action [intent=reverse_etl availability=not_implemented write=retry_inbound_message]; approval: requires plan, preview, approval, and execute; risk: retries processing for one inbound Postmark message; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    send bulk email apply - Plan and execute the send bulk email reverse-ETL action [intent=reverse_etl availability=not_implemented write=send_bulk_email]; approval: requires plan, preview, approval, and execute; risk: submits a live Postmark bulk email request; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    send email apply - Plan and execute the send email reverse-ETL action [intent=reverse_etl availability=not_implemented write=send_email]; approval: requires plan, preview, approval, and execute; risk: sends a live Postmark email; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    send email with template apply - Plan and execute the send email with template reverse-ETL action [intent=reverse_etl availability=not_implemented write=send_email_with_template]; approval: requires plan, preview, approval, and execute; risk: sends a live Postmark template email; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    stats outbound bounces list - Run the stats outbound bounces ETL stream [intent=etl availability=implemented stream=stats_outbound_bounces]
    stats outbound click browser families list - Run the stats outbound click browser families ETL stream [intent=etl availability=implemented stream=stats_outbound_click_browser_families]
    stats outbound click location list - Run the stats outbound click location ETL stream [intent=etl availability=implemented stream=stats_outbound_click_location]
    stats outbound click platforms list - Run the stats outbound click platforms ETL stream [intent=etl availability=implemented stream=stats_outbound_click_platforms]
    stats outbound clicks list - Run the stats outbound clicks ETL stream [intent=etl availability=implemented stream=stats_outbound_clicks]
    stats outbound email clients list - Run the stats outbound email clients ETL stream [intent=etl availability=implemented stream=stats_outbound_email_clients]
    stats outbound list - Run the stats outbound ETL stream [intent=etl availability=implemented stream=stats_outbound]
    stats outbound open platforms list - Run the stats outbound open platforms ETL stream [intent=etl availability=implemented stream=stats_outbound_open_platforms]
    stats outbound opens list - Run the stats outbound opens ETL stream [intent=etl availability=implemented stream=stats_outbound_opens]
    stats outbound sends list - Run the stats outbound sends ETL stream [intent=etl availability=implemented stream=stats_outbound_sends]
    stats outbound spam list - Run the stats outbound spam ETL stream [intent=etl availability=implemented stream=stats_outbound_spam]
    stats outbound tracked list - Run the stats outbound tracked ETL stream [intent=etl availability=implemented stream=stats_outbound_tracked]
    suppressions list - Run the suppressions ETL stream [intent=etl availability=implemented stream=suppressions]
    template list - Run the template ETL stream [intent=etl availability=implemented stream=template]
    templates list - Run the templates ETL stream [intent=etl availability=implemented stream=templates]
    unarchive message stream apply - Plan and execute the unarchive message stream reverse-ETL action [intent=reverse_etl availability=not_implemented write=unarchive_message_stream]; approval: requires plan, preview, approval, and execute; risk: unarchives a Postmark message stream; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    validate template apply - Plan and execute the validate template reverse-ETL action [intent=reverse_etl availability=implemented write=validate_template]; approval: requires plan, preview, approval, and execute; risk: validates Postmark template content; no persistent mutation expected but still invokes the external API
    webhook list - Run the webhook ETL stream [intent=etl availability=implemented stream=webhook]
    webhooks list - Run the webhooks ETL stream [intent=etl availability=implemented stream=webhooks]

EXAMPLES
  # Inspect as a manual
  pm connectors inspect postmarkapp

  # Inspect as structured JSON
  pm connectors inspect postmarkapp --json

AGENT WORKFLOW
  - Run pm connectors inspect postmarkapp before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
