# pm connectors inspect help-scout

```text
NAME
  pm connectors inspect help-scout - Help Scout connector manual

SYNOPSIS
  pm connectors inspect help-scout
  pm connectors inspect help-scout --json
  pm credentials add <name> --connector help-scout [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads, directly inspects, and plans typed Help Scout Inbox API operations through the Mailbox API using OAuth2 client-credentials authentication.

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
  start_date
  token_url
  client_id (secret)
  client_secret (secret)

ETL STREAMS
  conversations:
    primary key: id
    cursor: userUpdatedAt
    fields: assigneeId(), closedAt(), createdAt(), folderId(), id(), mailboxId(), number(), preview(), state(), status(), subject(), threads(), type(), userUpdatedAt()
  customers:
    primary key: id
    cursor: updatedAt
    fields: age(), createdAt(), firstName(), gender(), id(), jobTitle(), lastName(), organization(), photoUrl(), updatedAt()
  mailboxes:
    primary key: id
    cursor: updatedAt
    fields: createdAt(), email(), id(), name(), slug(), updatedAt()
  users:
    primary key: id
    cursor: updatedAt
    fields: createdAt(), email(), firstName(), id(), jobTitle(), lastName(), role(), timezone(), type(), updatedAt()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  delete_attachment:
    endpoint: DELETE /v2/conversations/{{ record.conversation_id }}/attachments/{{ record.attachment_id }}
    required fields: conversation_id, attachment_id
    risk: Deletes Help Scout data or configuration.
  upload_attachment:
    endpoint: POST /v2/conversations/{{ record.conversation_id }}/threads/{{ record.thread_id }}/attachments
    required fields: conversation_id, thread_id
    optional fields: fileName, mimeType, data
    risk: Mutates Help Scout Inbox API data.
  update_custom_fields:
    endpoint: PUT /v2/conversations/{{ record.conversation_id }}/fields
    required fields: conversation_id
    optional fields: fields
    risk: Mutates Help Scout Inbox API data.
  delete_snooze:
    endpoint: DELETE /v2/conversations/{{ record.conversation_id }}/snooze
    required fields: conversation_id
    risk: Deletes Help Scout data or configuration.
  update_snooze:
    endpoint: PUT /v2/conversations/{{ record.conversation_id }}/snooze
    required fields: conversation_id
    optional fields: snoozedUntil, unsnoozeOnCustomerReply
    risk: Mutates Help Scout Inbox API data.
  update_tags:
    endpoint: PUT /v2/conversations/{{ record.conversation_id }}/tags
    required fields: conversation_id
    optional fields: tags
    risk: Mutates Help Scout Inbox API data.
  delete_thread_schedule:
    endpoint: DELETE /v2/conversations/{{ record.conversation_id }}/threads/{{ record.thread_id }}/schedule
    required fields: conversation_id, thread_id
    risk: Deletes Help Scout data or configuration.
  publish_thread_schedule:
    endpoint: PATCH /v2/conversations/{{ record.conversation_id }}/threads/{{ record.thread_id }}/schedule
    required fields: conversation_id, thread_id
    optional fields: op, path, value
    risk: Mutates Help Scout Inbox API data.
  update_thread_schedule:
    endpoint: PUT /v2/conversations/{{ record.conversation_id }}/threads/{{ record.thread_id }}/schedule
    required fields: conversation_id, thread_id
    optional fields: scheduledFor, unscheduleOnCustomerReply, sendAsCreator
    risk: Mutates Help Scout Inbox API data.
  create_chat_thread:
    endpoint: POST /v2/conversations/{{ record.conversation_id }}/chats
    required fields: conversation_id
    optional fields: customer, text, attachments
    risk: Mutates Help Scout Inbox API data.
  create_customer_thread:
    endpoint: POST /v2/conversations/{{ record.conversation_id }}/customer
    required fields: conversation_id
    optional fields: customer, text, attachments
    risk: Mutates Help Scout Inbox API data.
  create_note:
    endpoint: POST /v2/conversations/{{ record.conversation_id }}/notes
    required fields: conversation_id
    optional fields: text, attachments
    risk: Mutates Help Scout Inbox API data.
  create_phone_thread:
    endpoint: POST /v2/conversations/{{ record.conversation_id }}/phones
    required fields: conversation_id
    optional fields: customer, text, attachments
    risk: Mutates Help Scout Inbox API data.
  create_reply_thread:
    endpoint: POST /v2/conversations/{{ record.conversation_id }}/reply
    required fields: conversation_id
    optional fields: customer, text, attachments
    risk: Mutates Help Scout Inbox API data.
  update_thread:
    endpoint: PATCH /v2/conversations/{{ record.conversation_id }}/threads/{{ record.thread_id }}
    required fields: conversation_id, thread_id
    optional fields: op, path, value
    risk: Mutates Help Scout Inbox API data.
  create_conversation:
    endpoint: POST /v2/conversations
    optional fields: subject, customer, mailboxId, type, status, createdAt, threads, tags, fields
    risk: Mutates Help Scout Inbox API data.
  delete_conversation:
    endpoint: DELETE /v2/conversations/{{ record.conversation_id }}
    required fields: conversation_id
    risk: Deletes Help Scout data or configuration.
  update_conversation:
    endpoint: PATCH /v2/conversations/{{ record.conversation_id }}
    required fields: conversation_id
    optional fields: op, path, value
    risk: Mutates Help Scout Inbox API data.
  create_address:
    endpoint: POST /v2/customers/{{ record.customer_id }}/address
    required fields: customer_id
    optional fields: city, state, postalCode, country, lines
    risk: Mutates Help Scout Inbox API data.
  delete_address:
    endpoint: DELETE /v2/customers/{{ record.customer_id }}/address
    required fields: customer_id
    risk: Deletes Help Scout data or configuration.
  update_address:
    endpoint: PUT /v2/customers/{{ record.customer_id }}/address
    required fields: customer_id
    optional fields: city, state, postalCode, country, lines
    risk: Mutates Help Scout Inbox API data.
  create_chat_handle:
    endpoint: POST /v2/customers/{{ record.customer_id }}/chats
    required fields: customer_id
    optional fields: type, value
    risk: Mutates Help Scout Inbox API data.
  delete_chat_handle:
    endpoint: DELETE /v2/customers/{{ record.customer_id }}/chats/{{ record.chat_id }}
    required fields: customer_id, chat_id
    risk: Deletes Help Scout data or configuration.
  update_chat_handles:
    endpoint: PUT /v2/customers/{{ record.customer_id }}/chats/{{ record.chat_id }}
    required fields: customer_id, chat_id
    optional fields: type, value
    risk: Mutates Help Scout Inbox API data.
  create_email:
    endpoint: POST /v2/customers/{{ record.customer_id }}/emails
    required fields: customer_id
    optional fields: type, value
    risk: Mutates Help Scout Inbox API data.
  delete_email:
    endpoint: DELETE /v2/customers/{{ record.customer_id }}/emails/{{ record.email_id }}
    required fields: customer_id, email_id
    risk: Deletes Help Scout data or configuration.
  update_email:
    endpoint: PUT /v2/customers/{{ record.customer_id }}/emails/{{ record.email_id }}
    required fields: customer_id, email_id
    optional fields: type, value
    risk: Mutates Help Scout Inbox API data.
  create_phone:
    endpoint: POST /v2/customers/{{ record.customer_id }}/phones
    required fields: customer_id
    optional fields: type, value
    risk: Mutates Help Scout Inbox API data.
  delete_phone:
    endpoint: DELETE /v2/customers/{{ record.customer_id }}/phones/{{ record.phone_id }}
    required fields: customer_id, phone_id
    risk: Deletes Help Scout data or configuration.
  update_phone:
    endpoint: PUT /v2/customers/{{ record.customer_id }}/phones/{{ record.phone_id }}
    required fields: customer_id, phone_id
    optional fields: type, value
    risk: Mutates Help Scout Inbox API data.
  create_social_profile:
    endpoint: POST /v2/customers/{{ record.customer_id }}/social-profiles
    required fields: customer_id
    optional fields: type, value
    risk: Mutates Help Scout Inbox API data.
  delete_social_profile:
    endpoint: DELETE /v2/customers/{{ record.customer_id }}/social-profiles/{{ record.social_profile_id }}
    required fields: customer_id, social_profile_id
    risk: Deletes Help Scout data or configuration.
  update_social_profile:
    endpoint: PUT /v2/customers/{{ record.customer_id }}/social-profiles/{{ record.social_profile_id }}
    required fields: customer_id, social_profile_id
    optional fields: type, value
    risk: Mutates Help Scout Inbox API data.
  create_website:
    endpoint: POST /v2/customers/{{ record.customer_id }}/websites
    required fields: customer_id
    optional fields: value
    risk: Mutates Help Scout Inbox API data.
  delete_website:
    endpoint: DELETE /v2/customers/{{ record.customer_id }}/websites/{{ record.website_id }}
    required fields: customer_id, website_id
    risk: Deletes Help Scout data or configuration.
  update_website:
    endpoint: PUT /v2/customers/{{ record.customer_id }}/websites/{{ record.website_id }}
    required fields: customer_id, website_id
    optional fields: value
    risk: Mutates Help Scout Inbox API data.
  create_customer:
    endpoint: POST /v2/customers
    optional fields: firstName, lastName, photoUrl, photoType, jobTitle, location, background, age, gender, organization, organizationId, emails
    risk: Mutates Help Scout Inbox API data.
  delete_customer:
    endpoint: DELETE /v2/customers/{{ record.customer_id }}
    required fields: customer_id
    risk: Deletes Help Scout data or configuration.
  delete_customer_asynchronously:
    endpoint: DELETE /v2/customers/{{ record.customer_id }}?async=true
    required fields: customer_id
    risk: Deletes Help Scout data or configuration.
  overwrite_customer:
    endpoint: PUT /v2/customers/{{ record.customer_id }}
    required fields: customer_id
    optional fields: firstName, lastName, photoUrl, photoType, jobTitle, location, background, age, gender, organization, organizationId
    risk: Mutates Help Scout Inbox API data.
  update_customer:
    endpoint: PATCH /v2/customers/{{ record.customer_id }}
    required fields: customer_id
    optional fields: op, path, value
    risk: Mutates Help Scout Inbox API data.
  update_routing_configuration:
    endpoint: PUT /v2/mailboxes/{{ record.mailbox_id }}/routing
    required fields: mailbox_id
    optional fields: state, assignmentLimit, assignmentMethod, userIds
    risk: Mutates Help Scout Inbox API data.
  create_saved_reply:
    endpoint: POST /v2/mailboxes/{{ record.mailbox_id }}/saved-replies
    required fields: mailbox_id
    optional fields: name, text, chatText
    risk: Mutates Help Scout Inbox API data.
  delete_saved_reply:
    endpoint: DELETE /v2/mailboxes/{{ record.mailbox_id }}/saved-replies/{{ record.saved_reply_id }}
    required fields: mailbox_id, saved_reply_id
    risk: Deletes Help Scout data or configuration.
  update_saved_reply:
    endpoint: PUT /v2/mailboxes/{{ record.mailbox_id }}/saved-replies/{{ record.saved_reply_id }}
    required fields: mailbox_id, saved_reply_id
    optional fields: name, text, chatText
    risk: Mutates Help Scout Inbox API data.
  create_organization_property_definition:
    endpoint: POST /v2/organizations/properties
    optional fields: type, slug, name, options
    risk: Mutates Help Scout Inbox API data.
  delete_organization_property_definition:
    endpoint: DELETE /v2/organizations/properties/{{ record.property_slug }}
    required fields: property_slug
    risk: Deletes Help Scout data or configuration.
  remove_organization_property_value:
    endpoint: DELETE /v2/organizations/{{ record.organization_id }}/properties/{{ record.property_slug }}
    required fields: organization_id, property_slug
    risk: Deletes Help Scout data or configuration.
  set_organization_property_value:
    endpoint: PUT /v2/organizations/{{ record.organization_id }}/properties/{{ record.property_slug }}
    required fields: organization_id, property_slug
    optional fields: value
    risk: Mutates Help Scout Inbox API data.
  update_organization_property_definition:
    endpoint: PUT /v2/organizations/properties/{{ record.property_slug }}
    required fields: property_slug
    optional fields: name, options
    risk: Mutates Help Scout Inbox API data.
  create_organization:
    endpoint: POST /v2/organizations
    optional fields: name, domain, website, description, location, note, phones, brandColor
    risk: Mutates Help Scout Inbox API data.
  delete_organization_by_id:
    endpoint: DELETE /v2/organizations/{{ record.organization_id }}
    required fields: organization_id
    risk: Deletes Help Scout data or configuration.
  update_organization:
    endpoint: PUT /v2/organizations/{{ record.organization_id }}
    required fields: organization_id
    optional fields: name, website, description, location, brandColor, note, phones, domains
    risk: Mutates Help Scout Inbox API data.
  create_customer_property_definition:
    endpoint: POST /v2/customer-properties
    optional fields: type, slug, name, options
    risk: Mutates Help Scout Inbox API data.
  delete_customer_property_definition:
    endpoint: DELETE /v2/customer-properties/{{ record.property_slug }}
    required fields: property_slug
    risk: Deletes Help Scout data or configuration.
  update_customer_properties:
    endpoint: PATCH /v2/customers/{{ record.customer_id }}/properties
    required fields: customer_id
    optional fields: op, value, path
    risk: Mutates Help Scout Inbox API data.
  update_team_members:
    endpoint: PUT /v2/teams/{{ record.team_id }}/members
    required fields: team_id
    optional fields: addedUserIds, deletedUserIds
    risk: Mutates Help Scout Inbox API data.
  update_conversation_reassignment_configuration:
    endpoint: PUT /v2/users/{{ record.user_id }}/conversation-reassignment
    required fields: user_id
    optional fields: enabled, awayDuration
    risk: Mutates Help Scout Inbox API data.
  set_user_status:
    endpoint: PUT /v2/users/{{ record.user_id }}/status
    required fields: user_id
    optional fields: status, customStatus
    risk: Mutates Help Scout Inbox API data.
  create_user:
    endpoint: POST /v2/users
    optional fields: firstName, lastName, email, role, timezone, jobTitle, phone, sendInvite, mailboxes
    risk: Mutates Help Scout Inbox API data.
  delete_user:
    endpoint: DELETE /v2/users/{{ record.user_id }}
    required fields: user_id
    risk: Deletes Help Scout data or configuration.
  create_webhook:
    endpoint: POST /v2/webhooks
    optional fields: url, events, secret, payloadVersion, label
    risk: Mutates Help Scout Inbox API data.
  delete_webhook:
    endpoint: DELETE /v2/webhooks/{{ record.webhook_id }}
    required fields: webhook_id
    risk: Deletes Help Scout data or configuration.
  update_webhook:
    endpoint: PUT /v2/webhooks/{{ record.webhook_id }}
    required fields: webhook_id
    optional fields: url, events, secret, payloadVersion, label
    risk: Mutates Help Scout Inbox API data.
  run_manual_workflows:
    endpoint: POST /v2/workflows/{{ record.workflow_id }}/run
    required fields: workflow_id
    optional fields: conversationIds
    risk: Mutates Help Scout Inbox API data.
  update_workflow_status:
    endpoint: PATCH /v2/workflows/{{ record.workflow_id }}
    required fields: workflow_id
    optional fields: value, op, path
    risk: Mutates Help Scout Inbox API data.

SECURITY
  read risk: external Help Scout API read of conversation, customer, mailbox, and user data
  write risk: typed Help Scout mutations are available only through reverse ETL plan, preview, approval token, and typed destructive confirmation gates
  approval: reverse ETL writes require plan, preview, approval token, and --confirm destructive before execution
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Work with Help Scout Inbox API operations safely.
  Usage: pm help-scout <resource> <action> [flags]
  Source CLI: Help Scout Inbox API (https://developer.helpscout.com/mailbox-api/)
  Global flags:
    --json (boolean): Write machine-readable JSON output.
    --connection (string): Use a saved Help Scout connector credential.: maps_to=connection
    --preview (boolean): Preview a reverse-ETL write plan without executing it.
    --confirm (string): Typed confirmation required for destructive/sensitive Help Scout write execution.
  Stream-backed reads
  Bounded JSON direct reads
  Plan/preview/approval reverse ETL writes
  Bounded binary/raw payload operations
  Other Commands
    conversation list - List Help Scout conversations as an ETL stream. [intent=etl availability=implemented stream=conversations]; flags: --modified-since
    customer list - List Help Scout customers as an ETL stream. [intent=etl availability=implemented stream=customers]; flags: --modified-since
    mailbox list - List Help Scout mailboxes as an ETL stream. [intent=etl availability=implemented stream=mailboxes]
    user list - List Help Scout users as an ETL stream. [intent=etl availability=implemented stream=users]
    conversations attachments file download - Download Attachment File [intent=direct_read availability=implemented]; risk: Binary/raw message retrieval is bounded by operation metadata and remains feature-gated until binary execution policy is enabled.; notes: Operation-backed binary command; no raw unbounded download is exposed.
    conversations attachments data get - Get Attachment Data [intent=direct_read availability=implemented]; flags: --conversation-id, --attachment-id
    conversations threads original source get - Get Thread Original Source / Get Thread Original Source (RFC 822) [intent=direct_read availability=implemented]; risk: Binary/raw message retrieval is bounded by operation metadata and remains feature-gated until binary execution policy is enabled.; notes: Operation-backed binary command; no raw unbounded download is exposed.
    conversations threads list - List Threads [intent=direct_read availability=implemented]; flags: --conversation-id
    v3 conversations threads list - List Threads (v3) [intent=direct_read availability=implemented]; flags: --conversation-id
    conversations get - Get Conversation [intent=direct_read availability=implemented]; flags: --conversation-id
    v3 conversations get - Get Conversation (v3) [intent=direct_read availability=implemented]; flags: --conversation-id
    customers address get - Get Address [intent=direct_read availability=implemented]; flags: --customer-id
    customers chats list - List Chats Handles [intent=direct_read availability=implemented]; flags: --customer-id
    customers emails list - List Emails [intent=direct_read availability=implemented]; flags: --customer-id
    customers phones list - List Phones [intent=direct_read availability=implemented]; flags: --customer-id
    customers social profiles list - List Social Profiles [intent=direct_read availability=implemented]; flags: --customer-id
    customers websites list - List Websites [intent=direct_read availability=implemented]; flags: --customer-id
    customers get - Get Customer [intent=direct_read availability=implemented]; flags: --customer-id
    v3 customers list - List Customers (v3) [intent=direct_read availability=implemented]
    mailboxes routing get - Get Routing configuration [intent=direct_read availability=implemented]; flags: --mailbox-id
    mailboxes saved replies get - Get Saved Reply [intent=direct_read availability=implemented]; flags: --mailbox-id, --saved-reply-id
    mailboxes saved replies list - List Saved Replies [intent=direct_read availability=implemented]; flags: --mailbox-id
    mailboxes get - Get Inbox [intent=direct_read availability=implemented]; flags: --mailbox-id
    mailboxes fields list - List Inbox Custom Fields [intent=direct_read availability=implemented]; flags: --mailbox-id
    mailboxes folders list - List Inbox Folders [intent=direct_read availability=implemented]; flags: --mailbox-id
    organizations get - Get Organization by ID [intent=direct_read availability=implemented]; flags: --organization-id
    organizations properties get - Get Organization Property Definition [intent=direct_read availability=implemented]; flags: --property-slug
    organizations properties list - List Organization Property Definitions [intent=direct_read availability=implemented]
    organizations conversations get - Get Organization Conversations [intent=direct_read availability=implemented]; flags: --organization-id
    organizations customers get - Get Organization Customers [intent=direct_read availability=implemented]; flags: --organization-id
    organizations list - List Organizations [intent=direct_read availability=implemented]
    customer properties list - List Customer Property Definitions [intent=direct_read availability=implemented]
    ratings get - Get Satisfaction Rating [intent=direct_read availability=implemented]; flags: --rating-id
    reports company get - Company Overall Report [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports company customers helped get - Company Customers Helped [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports company drilldown get - Company Drilldown [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports conversations get - Conversations - Overall Report [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports conversations volume by channel get - All Channels - Volumes by Channel [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports conversations busy times get - Conversations - Busiest Time of Day [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports conversations drilldown get - Conversations - Drilldown [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports conversations fields drilldown get - Conversations - Drilldown by Field [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports conversations new get - Conversations - New Conversations [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports conversations new drilldown get - Conversations - New Conversations Drilldown [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports conversations received messages get - Conversations - Received Messages Statistics [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports docs get - Docs Overall Report [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports happiness get - Happiness Overall Report [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports happiness ratings get - Happiness Ratings Report [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports productivity get - Productivity Overall Report [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports productivity first response time get - Productivity - First Response Time [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports productivity replies sent get - Productivity - Replies Sent [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports productivity resolution time get - Productivity - Resolution Time [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports productivity resolved get - Productivity - Resolved [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports productivity response time get - Productivity - Response Time [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports user get - User/Team Overall Report [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports user conversation history get - User Conversation History [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports user customers helped get - User Customers Helped [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports user drilldown get - User Drill-down [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports user happiness get - User Happiness [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports user ratings get - User Happiness drilldown [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports user replies get - User Replies [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports user resolutions get - User Resolutions [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports user chat get - User/Team Chat Report [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports chat get - Chat Report [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports email get - Email Report [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    reports phone get - Phone Report [intent=direct_read availability=implemented]; flags: --start, --end, --mailboxes, --users, --tags
    v3 system users get - Get System User [intent=direct_read availability=implemented]; flags: --system-user-id
    v3 system users list - List System Users [intent=direct_read availability=implemented]
    tags get - Get Tag by ID [intent=direct_read availability=implemented]; flags: --tag-id
    tags list - List Tags [intent=direct_read availability=implemented]
    teams members list - List Team Members [intent=direct_read availability=implemented]; flags: --team-id
    teams list - List Teams [intent=direct_read availability=implemented]
    users conversation reassignment get - Get Conversation Reassignment configuration [intent=direct_read availability=implemented]; flags: --user-id
    users status get - Get user status [intent=direct_read availability=implemented]; flags: --user-id
    users status list - List users statuses [intent=direct_read availability=implemented]
    users me get - Get Resource Owner [intent=direct_read availability=implemented]
    users get - Get User [intent=direct_read availability=implemented]; flags: --user-id
    webhooks get - Get Webhook [intent=direct_read availability=implemented]; flags: --webhook-id
    webhooks list - List Webhooks [intent=direct_read availability=implemented]
    workflows list - List Workflows [intent=direct_read availability=implemented]
    conversations attachments delete - Delete Attachment [intent=reverse_etl availability=implemented write=delete_attachment]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Deletes Help Scout data or configuration.; flags: --conversation-id, --attachment-id
    conversations threads attachments upload - Upload Attachment [intent=reverse_etl availability=implemented write=upload_attachment]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --conversation-id, --thread-id, --fileName, --mimeType, --data
    conversations fields update - Update Custom Fields [intent=reverse_etl availability=implemented write=update_custom_fields]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --conversation-id
    conversations snooze delete - Delete Snooze [intent=reverse_etl availability=implemented write=delete_snooze]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Deletes Help Scout data or configuration.; flags: --conversation-id
    conversations snooze update - Update Snooze [intent=reverse_etl availability=implemented write=update_snooze]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --conversation-id, --snoozedUntil, --unsnoozeOnCustomerReply
    conversations tags update - Update Tags [intent=reverse_etl availability=implemented write=update_tags]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --conversation-id, --tags
    conversations threads schedule delete - Delete Thread Schedule [intent=reverse_etl availability=implemented write=delete_thread_schedule]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Deletes Help Scout data or configuration.; flags: --conversation-id, --thread-id
    conversations threads schedule publish - Publish Thread Schedule [intent=reverse_etl availability=implemented write=publish_thread_schedule]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --conversation-id, --thread-id, --op, --path, --value
    conversations threads schedule update - Update Thread Schedule [intent=reverse_etl availability=implemented write=update_thread_schedule]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --conversation-id, --thread-id, --scheduledFor, --unscheduleOnCustomerReply, --sendAsCreator
    conversations chats create - Create Chat Thread [intent=reverse_etl availability=implemented write=create_chat_thread]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --conversation-id, --text
    conversations customer create - Create Customer Thread [intent=reverse_etl availability=implemented write=create_customer_thread]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --conversation-id, --text
    conversations notes create - Create Note [intent=reverse_etl availability=implemented write=create_note]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --conversation-id, --text
    conversations phones create - Create Phone Thread [intent=reverse_etl availability=implemented write=create_phone_thread]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --conversation-id, --text
    conversations reply create - Create Reply Thread [intent=reverse_etl availability=implemented write=create_reply_thread]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --conversation-id, --text
    conversations threads update - Update Thread [intent=reverse_etl availability=implemented write=update_thread]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --conversation-id, --thread-id, --op, --path, --value
    conversations create - Create Conversation [intent=reverse_etl availability=implemented write=create_conversation]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --subject, --mailboxId, --type, --status, --createdAt, --tags
    conversations delete - Delete Conversation [intent=reverse_etl availability=implemented write=delete_conversation]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Deletes Help Scout data or configuration.; flags: --conversation-id
    conversations update - Update Conversation [intent=reverse_etl availability=implemented write=update_conversation]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --conversation-id, --op, --path, --value
    customers address create - Create Address [intent=reverse_etl availability=implemented write=create_address]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --customer-id, --city, --state, --postalCode, --country, --lines
    customers address delete - Delete Address [intent=reverse_etl availability=implemented write=delete_address]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Deletes Help Scout data or configuration.; flags: --customer-id
    customers address update - Update Address [intent=reverse_etl availability=implemented write=update_address]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --customer-id, --city, --state, --postalCode, --country, --lines
    customers chats create - Create Chat Handle [intent=reverse_etl availability=implemented write=create_chat_handle]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --customer-id, --type, --value
    customers chats delete - Delete Chat Handle [intent=reverse_etl availability=implemented write=delete_chat_handle]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Deletes Help Scout data or configuration.; flags: --customer-id, --chat-id
    customers chats update - Update Chat Handles [intent=reverse_etl availability=implemented write=update_chat_handles]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --customer-id, --chat-id, --type, --value
    customers emails create - Create Email [intent=reverse_etl availability=implemented write=create_email]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --customer-id, --type, --value
    customers emails delete - Delete Email [intent=reverse_etl availability=implemented write=delete_email]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Deletes Help Scout data or configuration.; flags: --customer-id, --email-id
    customers emails update - Update Email [intent=reverse_etl availability=implemented write=update_email]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --customer-id, --email-id, --type, --value
    customers phones create - Create Phone [intent=reverse_etl availability=implemented write=create_phone]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --customer-id, --type, --value
    customers phones delete - Delete Phone [intent=reverse_etl availability=implemented write=delete_phone]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Deletes Help Scout data or configuration.; flags: --customer-id, --phone-id
    customers phones update - Update Phone [intent=reverse_etl availability=implemented write=update_phone]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --customer-id, --phone-id, --type, --value
    customers social profiles create - Create Social Profile [intent=reverse_etl availability=implemented write=create_social_profile]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --customer-id, --type, --value
    customers social profiles delete - Delete Social Profile [intent=reverse_etl availability=implemented write=delete_social_profile]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Deletes Help Scout data or configuration.; flags: --customer-id, --social-profile-id
    customers social profiles update - Update Social Profile [intent=reverse_etl availability=implemented write=update_social_profile]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --customer-id, --social-profile-id, --type, --value
    customers websites create - Create Website [intent=reverse_etl availability=implemented write=create_website]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --customer-id, --value
    customers websites delete - Delete Website [intent=reverse_etl availability=implemented write=delete_website]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Deletes Help Scout data or configuration.; flags: --customer-id, --website-id
    customers websites update - Update Website [intent=reverse_etl availability=implemented write=update_website]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --customer-id, --website-id, --value
    customers create - Create Customer [intent=reverse_etl availability=implemented write=create_customer]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --firstName, --lastName, --photoUrl, --photoType, --jobTitle, --location, --background, --age, --gender, --organization, --organizationId
    customers delete - Delete Customer [intent=reverse_etl availability=implemented write=delete_customer]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Deletes Help Scout data or configuration.; flags: --customer-id
    customers delete async - Delete Customer Asynchronously [intent=reverse_etl availability=implemented write=delete_customer_asynchronously]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Deletes Help Scout data or configuration.; flags: --customer-id
    customers overwrite - Overwrite Customer [intent=reverse_etl availability=implemented write=overwrite_customer]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --customer-id, --firstName, --lastName, --photoUrl, --photoType, --jobTitle, --location, --background, --age, --gender, --organization, --organizationId
    customers update - Update Customer [intent=reverse_etl availability=implemented write=update_customer]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --customer-id, --op, --path, --value
    mailboxes routing update - Update Routing configuration [intent=reverse_etl availability=implemented write=update_routing_configuration]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --mailbox-id, --state, --assignmentLimit, --assignmentMethod
    mailboxes saved replies create - Create Saved Reply [intent=reverse_etl availability=implemented write=create_saved_reply]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --mailbox-id, --name, --text, --chatText
    mailboxes saved replies delete - Delete Saved Reply [intent=reverse_etl availability=implemented write=delete_saved_reply]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Deletes Help Scout data or configuration.; flags: --mailbox-id, --saved-reply-id
    mailboxes saved replies update - Update Saved Reply [intent=reverse_etl availability=implemented write=update_saved_reply]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --mailbox-id, --saved-reply-id, --name, --text, --chatText
    organizations properties create - Create Organization Property Definition [intent=reverse_etl availability=implemented write=create_organization_property_definition]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --type, --slug, --name
    organizations properties delete - Delete Organization Property Definition [intent=reverse_etl availability=implemented write=delete_organization_property_definition]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Deletes Help Scout data or configuration.; flags: --property-slug
    organizations properties remove - Remove Organization Property Value [intent=reverse_etl availability=implemented write=remove_organization_property_value]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Deletes Help Scout data or configuration.; flags: --organization-id, --property-slug
    organizations properties set - Set Organization Property Value [intent=reverse_etl availability=implemented write=set_organization_property_value]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --organization-id, --property-slug, --value
    organizations properties update - Update Organization Property Definition [intent=reverse_etl availability=implemented write=update_organization_property_definition]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --property-slug, --name
    organizations create - Create Organization [intent=reverse_etl availability=implemented write=create_organization]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --name, --domain, --website, --description, --location, --note, --phones, --brandColor
    organizations delete - Delete Organization by ID [intent=reverse_etl availability=implemented write=delete_organization_by_id]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Deletes Help Scout data or configuration.; flags: --organization-id
    organizations update - Update Organization [intent=reverse_etl availability=implemented write=update_organization]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --organization-id, --name, --website, --description, --location, --brandColor, --note, --phones, --domains
    customer properties create - Create Customer Property Definition [intent=reverse_etl availability=implemented write=create_customer_property_definition]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --type, --slug, --name
    customer properties delete - Delete Customer Property Definition [intent=reverse_etl availability=implemented write=delete_customer_property_definition]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Deletes Help Scout data or configuration.; flags: --property-slug
    customers properties update - Update Customer Properties [intent=reverse_etl availability=implemented write=update_customer_properties]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --customer-id, --op, --value, --path
    teams members update - Update Team Members [intent=reverse_etl availability=implemented write=update_team_members]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --team-id
    users conversation reassignment update - Update Conversation Reassignment configuration [intent=reverse_etl availability=implemented write=update_conversation_reassignment_configuration]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --user-id, --enabled
    users status set - Set user status [intent=reverse_etl availability=implemented write=set_user_status]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --user-id, --status
    users create - Create User [intent=reverse_etl availability=implemented write=create_user]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --firstName, --lastName, --email, --role, --timezone, --jobTitle, --phone, --sendInvite
    users delete - Delete User [intent=reverse_etl availability=implemented write=delete_user]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Deletes Help Scout data or configuration.; flags: --user-id
    webhooks create - Create Webhook [intent=reverse_etl availability=implemented write=create_webhook]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --url, --events, --secret, --payloadVersion, --label
    webhooks delete - Delete Webhook [intent=reverse_etl availability=implemented write=delete_webhook]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Deletes Help Scout data or configuration.; flags: --webhook-id
    webhooks update - Update Webhook [intent=reverse_etl availability=implemented write=update_webhook]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --webhook-id, --url, --events, --secret, --payloadVersion, --label
    workflows run - Run Manual Workflows [intent=reverse_etl availability=implemented write=run_manual_workflows]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --workflow-id
    workflows update - Update workflow status [intent=reverse_etl availability=implemented write=update_workflow_status]; approval: Requires connector command plan, preview, approval token, and --confirm destructive before execution.; risk: Mutates Help Scout customer, conversation, or mailbox data.; flags: --workflow-id, --value, --op, --path
  Help topics:
    safety - Help Scout writes require plan, preview, approval token, and typed confirmation before execution.
    binary - Binary/raw payload endpoints are bounded operation metadata and are not exposed as unbounded downloads.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect help-scout

  # Inspect as structured JSON
  pm connectors inspect help-scout --json

AGENT WORKFLOW
  - Run pm connectors inspect help-scout before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
