# pm connectors inspect help-scout

```text
NAME
  pm connectors inspect help-scout - Help Scout connector manual

SYNOPSIS
  pm connectors inspect help-scout
  pm connectors inspect help-scout --json
  pm credentials add <name> --connector help-scout [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes the documented Help Scout Mailbox API v2 surface: conversations and threads, customers and their email/phone/chat/social/website contact records, organizations, mailboxes and mailbox configuration, users, teams, tags, webhooks, and workflows, through OAuth2 client-credentials authentication.

ICON
  id: simple-icons-helpscout
  asset: icons/simple-icons/helpscout.svg
  title: Help Scout
  simple_icon_slug: helpscout
  simple_icon_hex: 1292EE
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Help%20Scout
  match: exact-name-or-slug
  matched_by: help-scout

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  conversationid
  customerid
  mailboxid
  organizationid
  start_date
  teamid
  token_url
  client_id (secret)
  client_secret (secret)

ETL STREAMS
  conversations:
    primary key: id
    cursor: userUpdatedAt
    fields: assigneeId(integer), closedAt(string), createdAt(string), folderId(integer), id(integer), mailboxId(integer), number(integer), preview(string), state(string), status(string), subject(string), threads(integer), type(string), userUpdatedAt(string)
  conversations_threads:
    primary key: id
    fields: action(string), assignedTo(string), bcc(string), body(string), cc(string), conversationId(string), createdAt(string), createdBy(string), customer(string), id(integer), openedAt(string), rating(string), savedReplyId(string), scheduled(string), source(string), state(string), status(string), to(string), type(string)
  customer_properties:
    primary key: id
    fields: id(integer)
  customers:
    primary key: id
    cursor: updatedAt
    fields: age(string), createdAt(string), firstName(string), gender(string), id(integer), jobTitle(string), lastName(string), organization(string), photoUrl(string), updatedAt(string)
  customers_chats:
    primary key: id
    fields: customerId(string), id(integer), type(string), value(string)
  customers_emails:
    primary key: id
    fields: customerId(string), id(integer), type(string), value(string)
  customers_phones:
    primary key: id
    fields: customerId(string), id(integer), type(string), value(string)
  customers_social_profiles:
    primary key: id
    fields: customerId(string), id(integer)
  customers_websites:
    primary key: id
    fields: customerId(string), id(integer), value(string)
  mailboxes:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), email(string), id(integer), name(string), slug(string), updatedAt(string)
  mailboxes_fields:
    primary key: id
    fields: id(integer), mailboxId(string), name(string), options(string), order(string), required(string), type(string)
  mailboxes_folders:
    primary key: id
    fields: activeCount(string), id(integer), mailboxId(string), name(string), totalCount(string), type(string), updatedAt(string), userId(string)
  mailboxes_saved_replies:
    primary key: id
    fields: id(integer), mailboxId(string)
  organizations:
    primary key: id
    fields: brandColor(string), description(string), domains(string), id(integer), location(string), logoUrl(string), name(string), note(string), phones(string), website(string)
  organizations_properties:
    primary key: id
    fields: id(integer)
  organizations_conversations:
    primary key: id
    fields: assignee(string), bcc(string), cc(string), closedAt(string), closedBy(string), closedByUser(string), createdAt(string), createdBy(string), customFields(string), folderId(string), id(integer), mailboxId(string), number(string), organizationId(string), preview(string), primaryCustomer(string), source(string), state(string), status(string), subject(string), tags(string), threads(string), type(string), userUpdatedAt(string)
  organizations_customers:
    primary key: id
    fields: age(string), background(string), createdAt(string), firstName(string), gender(string), id(integer), jobTitle(string), lastName(string), location(string), organization(string), organizationId(string), photoType(string), photoUrl(string), updatedAt(string)
  tags:
    primary key: id
    fields: color(string), createdAt(string), id(integer), name(string), slug(string), ticketCount(string)
  teams:
    primary key: id
    fields: createdAt(string), id(integer), initials(string), mention(string), name(string), photoUrl(string), timezone(string), updatedAt(string)
  teams_members:
    primary key: id
    fields: alternateEmails(string), createdAt(string), email(string), firstName(string), id(integer), initials(string), jobTitle(string), lastName(string), lastVisit(string), mention(string), phone(string), photoUrl(string), role(string), teamId(string), timezone(string), type(string), updatedAt(string)
  users:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), email(string), firstName(string), id(integer), jobTitle(string), lastName(string), role(string), timezone(string), type(string), updatedAt(string)
  users_status:
    primary key: userId
    fields: chat(string), email(string), userId(integer)
  webhooks:
    primary key: id
    fields: events(string), id(integer), label(string), mailboxIds(string), notification(string), payloadVersion(string), state(string), url(string)
  workflows:
    primary key: id
    fields: createdAt(string), id(integer), mailboxId(string), modifiedAt(string), name(string), order(string), status(string), type(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_conversation:
    endpoint: POST /conversations
    risk: Help Scout mutation: Create Conversation.
  delete_conversation:
    endpoint: DELETE /conversations/{{ record.conversationId }}
    required fields: conversationId
    risk: Destructive Help Scout mutation: Delete Conversation.
  update_conversation:
    endpoint: PATCH /conversations/{{ record.conversationId }}
    required fields: conversationId
    risk: Help Scout mutation: Update Conversation.
  delete_attachment:
    endpoint: DELETE /conversations/{{ record.conversationId }}/attachments/{{ record.attachmentId }}
    required fields: conversationId, attachmentId
    risk: Destructive Help Scout mutation: Delete Attachment.
  create_chat_thread:
    endpoint: POST /conversations/{{ record.conversationId }}/chats
    required fields: conversationId
    risk: Help Scout mutation: Create Chat Thread.
  create_customer_thread:
    endpoint: POST /conversations/{{ record.conversationId }}/customer
    required fields: conversationId
    risk: Help Scout mutation: Create Customer Thread.
  update_custom_fields:
    endpoint: PUT /conversations/{{ record.conversationId }}/fields
    required fields: conversationId
    risk: Help Scout mutation: Update Custom Fields.
  create_note:
    endpoint: POST /conversations/{{ record.conversationId }}/notes
    required fields: conversationId
    risk: Help Scout mutation: Create Note.
  create_phone_thread:
    endpoint: POST /conversations/{{ record.conversationId }}/phones
    required fields: conversationId
    risk: Help Scout mutation: Create Phone Thread.
  create_reply_thread:
    endpoint: POST /conversations/{{ record.conversationId }}/reply
    required fields: conversationId
    risk: Help Scout mutation: Create Reply Thread.
  delete_snooze:
    endpoint: DELETE /conversations/{{ record.conversationId }}/snooze
    required fields: conversationId
    risk: Destructive Help Scout mutation: Delete Snooze.
  update_snooze:
    endpoint: PUT /conversations/{{ record.conversationId }}/snooze
    required fields: conversationId
    risk: Help Scout mutation: Update Snooze.
  update_tags:
    endpoint: PUT /conversations/{{ record.conversationId }}/tags
    required fields: conversationId
    risk: Help Scout mutation: Update Tags.
  update_thread:
    endpoint: PATCH /conversations/{{ record.conversationId }}/threads/{{ record.threadId }}
    required fields: conversationId, threadId
    risk: Help Scout mutation: Update Thread.
  upload_attachment:
    endpoint: POST /conversations/{{ record.conversationId }}/threads/{{ record.threadId }}/attachments
    required fields: conversationId, threadId
    risk: Help Scout mutation: Upload Attachment.
  delete_thread_schedule:
    endpoint: DELETE /conversations/{{ record.conversationId }}/threads/{{ record.threadId }}/schedule
    required fields: conversationId, threadId
    risk: Destructive Help Scout mutation: Delete Thread Schedule.
  publish_thread_schedule:
    endpoint: PATCH /conversations/{{ record.conversationId }}/threads/{{ record.threadId }}/schedule
    required fields: conversationId, threadId
    risk: Help Scout mutation: Publish Thread Schedule.
  update_thread_schedule:
    endpoint: PUT /conversations/{{ record.conversationId }}/threads/{{ record.threadId }}/schedule
    required fields: conversationId, threadId
    risk: Help Scout mutation: Update Thread Schedule.
  create_customer_property_definition:
    endpoint: POST /customer-properties
    risk: Help Scout mutation: Create Customer Property Definition.
  delete_customer_property_definition:
    endpoint: DELETE /customer-properties/{{ record.slug }}
    required fields: slug
    risk: Destructive Help Scout mutation: Delete Customer Property Definition.
  create_customer:
    endpoint: POST /customers
    risk: Help Scout mutation: Create Customer.
  delete_customer:
    endpoint: DELETE /customers/{{ record.customerId }}
    required fields: customerId
    risk: Destructive Help Scout mutation: Delete Customer. Hard-deletes the customer, their microsurvey responses and every conversation they participated in, in compliance with GDPR erasure. There is no undo. Pass async=true to run it asynchronously (202 Accepted instead of 204 No Content), which is the only form that works for a customer with more than 100 conversations.
  update_customer:
    endpoint: PATCH /customers/{{ record.customerId }}
    required fields: customerId
    risk: Help Scout mutation: Update Customer.
  overwrite_customer:
    endpoint: PUT /customers/{{ record.customerId }}
    required fields: customerId
    risk: Help Scout mutation: Overwrite Customer.
  delete_address:
    endpoint: DELETE /customers/{{ record.customerId }}/address
    required fields: customerId
    risk: Destructive Help Scout mutation: Delete Address.
  create_address:
    endpoint: POST /customers/{{ record.customerId }}/address
    required fields: customerId
    risk: Help Scout mutation: Create Address.
  update_address:
    endpoint: PUT /customers/{{ record.customerId }}/address
    required fields: customerId
    risk: Help Scout mutation: Update Address.
  create_chat_handle:
    endpoint: POST /customers/{{ record.customerId }}/chats
    required fields: customerId
    risk: Help Scout mutation: Create Chat Handle.
  delete_chat_handle:
    endpoint: DELETE /customers/{{ record.customerId }}/chats/{{ record.chatId }}
    required fields: customerId, chatId
    risk: Destructive Help Scout mutation: Delete Chat Handle.
  update_chat_handles:
    endpoint: PUT /customers/{{ record.customerId }}/chats/{{ record.chatId }}
    required fields: customerId, chatId
    risk: Help Scout mutation: Update Chat Handles.
  create_email:
    endpoint: POST /customers/{{ record.customerId }}/emails
    required fields: customerId
    risk: Help Scout mutation: Create Email.
  delete_email:
    endpoint: DELETE /customers/{{ record.customerId }}/emails/{{ record.emailId }}
    required fields: customerId, emailId
    risk: Destructive Help Scout mutation: Delete Email.
  update_email:
    endpoint: PUT /customers/{{ record.customerId }}/emails/{{ record.emailId }}
    required fields: customerId, emailId
    risk: Help Scout mutation: Update Email.
  create_phone:
    endpoint: POST /customers/{{ record.customerId }}/phones
    required fields: customerId
    risk: Help Scout mutation: Create Phone.
  delete_phone:
    endpoint: DELETE /customers/{{ record.customerId }}/phones/{{ record.phoneId }}
    required fields: customerId, phoneId
    risk: Destructive Help Scout mutation: Delete Phone.
  update_phone:
    endpoint: PUT /customers/{{ record.customerId }}/phones/{{ record.phoneId }}
    required fields: customerId, phoneId
    risk: Help Scout mutation: Update Phone.
  update_customer_properties:
    endpoint: PATCH /customers/{{ record.customerId }}/properties
    required fields: customerId
    risk: Help Scout mutation: Update Customer Properties.
  create_social_profile:
    endpoint: POST /customers/{{ record.customerId }}/social-profiles
    required fields: customerId
    risk: Help Scout mutation: Create Social Profile.
  delete_social_profile:
    endpoint: DELETE /customers/{{ record.customerId }}/social-profiles/{{ record.socialProfileId }}
    required fields: customerId, socialProfileId
    risk: Destructive Help Scout mutation: Delete Social Profile.
  update_social_profile:
    endpoint: PUT /customers/{{ record.customerId }}/social-profiles/{{ record.socialProfileId }}
    required fields: customerId, socialProfileId
    risk: Help Scout mutation: Update Social Profile.
  create_website:
    endpoint: POST /customers/{{ record.customerId }}/websites
    required fields: customerId
    risk: Help Scout mutation: Create Website.
  delete_website:
    endpoint: DELETE /customers/{{ record.customerId }}/websites/{{ record.websiteId }}
    required fields: customerId, websiteId
    risk: Destructive Help Scout mutation: Delete Website.
  update_website:
    endpoint: PUT /customers/{{ record.customerId }}/websites/{{ record.websiteId }}
    required fields: customerId, websiteId
    risk: Help Scout mutation: Update Website.
  update_routing_configuration:
    endpoint: PUT /mailboxes/{{ record.mailboxId }}/routing
    required fields: mailboxId
    risk: Help Scout mutation: Update Routing configuration.
  create_saved_reply:
    endpoint: POST /mailboxes/{{ record.mailboxId }}/saved-replies
    required fields: mailboxId
    risk: Help Scout mutation: Create Saved Reply.
  delete_saved_reply:
    endpoint: DELETE /mailboxes/{{ record.mailboxId }}/saved-replies/{{ record.savedReplyId }}
    required fields: mailboxId, savedReplyId
    risk: Destructive Help Scout mutation: Delete Saved Reply.
  update_saved_reply:
    endpoint: PUT /mailboxes/{{ record.mailboxId }}/saved-replies/{{ record.savedReplyId }}
    required fields: mailboxId, savedReplyId
    risk: Help Scout mutation: Update Saved Reply.
  create_organization:
    endpoint: POST /organizations
    risk: Help Scout mutation: Create Organization.
  create_organization_property_definition:
    endpoint: POST /organizations/properties
    risk: Help Scout mutation: Create Organization Property Definition.
  delete_organization_property_definition:
    endpoint: DELETE /organizations/properties/{{ record.slug }}
    required fields: slug
    risk: Destructive Help Scout mutation: Delete Organization Property Definition.
  update_organization_property_definition:
    endpoint: PUT /organizations/properties/{{ record.slug }}
    required fields: slug
    risk: Help Scout mutation: Update Organization Property Definition.
  delete_organization_by_id:
    endpoint: DELETE /organizations/{{ record.organizationId }}
    required fields: organizationId
    risk: Destructive Help Scout mutation: Delete Organization by ID.
  update_organization:
    endpoint: PUT /organizations/{{ record.organizationId }}
    required fields: organizationId
    risk: Help Scout mutation: Update Organization.
  remove_organization_property_value:
    endpoint: DELETE /organizations/{{ record.organizationId }}/properties/{{ record.slug }}
    required fields: organizationId, slug
    risk: Destructive Help Scout mutation: Remove Organization Property Value.
  set_organization_property_value:
    endpoint: PUT /organizations/{{ record.organizationId }}/properties/{{ record.slug }}
    required fields: organizationId, slug
    risk: Help Scout mutation: Set Organization Property Value.
  update_team_members:
    endpoint: PUT /teams/{{ record.teamId }}/members
    required fields: teamId
    risk: Help Scout mutation: Update Team Members.
  create_user:
    endpoint: POST /users
    risk: Help Scout mutation: Create User.
  delete_user:
    endpoint: DELETE /users/{{ record.userId }}
    required fields: userId
    risk: Destructive Help Scout mutation: Delete User.
  update_conversation_reassignment_configuration:
    endpoint: PUT /users/{{ record.userId }}/conversation-reassignment
    required fields: userId
    risk: Help Scout mutation: Update Conversation Reassignment configuration.
  set_user_status:
    endpoint: PUT /users/{{ record.userId }}/status
    required fields: userId
    risk: Help Scout mutation: Set user status.
  create_webhook:
    endpoint: POST /webhooks
    risk: Help Scout mutation: Create Webhook.
  delete_webhook:
    endpoint: DELETE /webhooks/{{ record.webhookId }}
    required fields: webhookId
    risk: Destructive Help Scout mutation: Delete Webhook.
  update_webhook:
    endpoint: PUT /webhooks/{{ record.webhookId }}
    required fields: webhookId
    risk: Help Scout mutation: Update Webhook.
  update_workflow_status:
    endpoint: PATCH /workflows/{{ record.workflowId }}
    required fields: workflowId
    risk: Help Scout mutation: Update workflow status.
  run_manual_workflows:
    endpoint: POST /workflows/{{ record.workflowId }}/run
    required fields: workflowId
    risk: Help Scout mutation: Run Manual Workflows.

SECURITY
  read risk: external Help Scout API read of conversation, thread, customer, organization, mailbox, team, tag, webhook, workflow, and user data
  write risk: external Help Scout Mailbox API mutation of conversations, threads, customers and their contact records, organizations, users, teams, tags, webhooks, and workflows; 18 of the 65 write actions are permanent DELETEs that remove Help Scout records outright, including delete_customer, delete_conversation, delete_attachment, delete_email, and delete_phone
  approval: required for every write action; reverse ETL plan, preview, explicit approval, then execute, and the 18 destructive delete actions additionally require a typed confirmation
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Help Scout's declared streams and reverse-ETL actions.
  Usage: pm help-scout <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    api get v3 conversations conversationid - Documented GET /v3/conversations/{conversationId} (not implemented) [intent=direct_read availability=not_implemented operation=help-scout.get.v3-conversations-conversationid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 conversations conversationid threads - Documented GET /v3/conversations/{conversationId}/threads (not implemented) [intent=direct_read availability=not_implemented operation=help-scout.get.v3-conversations-conversationid-threads]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 customers - Documented GET /v3/customers (not implemented) [intent=direct_read availability=not_implemented operation=help-scout.get.v3-customers]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 system-users - Documented GET /v3/system-users (not implemented) [intent=direct_read availability=not_implemented operation=help-scout.get.v3-system-users]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v3 system-users systemuserid - Documented GET /v3/system-users/{systemUserId} (not implemented) [intent=direct_read availability=not_implemented operation=help-scout.get.v3-system-users-systemuserid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    attachments download - Download a conversation attachment to a bounded destination. [intent=binary_download availability=implemented operation=help-scout.attachment_download]; approval: filesystem writes require an explicit destination policy; risk: Bounded binary download; size-capped and written only to an explicit destination.; flags: --conversationId, --attachmentId, --dest-root (required), --file-name, --max-bytes
    conversations list - Read Help Scout List Conversations as ETL records. [intent=etl availability=implemented stream=conversations]
    conversations-threads list - Read Help Scout List Threads as ETL records. [intent=etl availability=implemented stream=conversations_threads]; flags: --conversation-id (required)
    create-address plan - Create Address. [intent=reverse_etl availability=not_implemented write=create_address]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Create Address.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create-chat-handle plan - Create Chat Handle. [intent=reverse_etl availability=not_implemented write=create_chat_handle]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Create Chat Handle.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create-chat-thread plan - Create Chat Thread. [intent=reverse_etl availability=not_implemented write=create_chat_thread]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Create Chat Thread.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create-conversation plan - Create Conversation. [intent=reverse_etl availability=implemented write=create_conversation]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Create Conversation.
    create-customer plan - Create Customer. [intent=reverse_etl availability=implemented write=create_customer]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Create Customer.
    create-customer-property-definition plan - Create Customer Property Definition. [intent=reverse_etl availability=implemented write=create_customer_property_definition]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Create Customer Property Definition.
    create-customer-thread plan - Create Customer Thread. [intent=reverse_etl availability=not_implemented write=create_customer_thread]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Create Customer Thread.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create-email plan - Create Email. [intent=reverse_etl availability=not_implemented write=create_email]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Create Email.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create-note plan - Create Note. [intent=reverse_etl availability=not_implemented write=create_note]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Create Note.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create-organization plan - Create Organization. [intent=reverse_etl availability=implemented write=create_organization]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Create Organization.
    create-organization-property-definition plan - Create Organization Property Definition. [intent=reverse_etl availability=implemented write=create_organization_property_definition]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Create Organization Property Definition.
    create-phone plan - Create Phone. [intent=reverse_etl availability=not_implemented write=create_phone]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Create Phone.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create-phone-thread plan - Create Phone Thread. [intent=reverse_etl availability=not_implemented write=create_phone_thread]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Create Phone Thread.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create-reply-thread plan - Create Reply Thread. [intent=reverse_etl availability=not_implemented write=create_reply_thread]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Create Reply Thread.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create-saved-reply plan - Create Saved Reply. [intent=reverse_etl availability=not_implemented write=create_saved_reply]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Create Saved Reply.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create-social-profile plan - Create Social Profile. [intent=reverse_etl availability=not_implemented write=create_social_profile]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Create Social Profile.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create-user plan - Create User. [intent=reverse_etl availability=implemented write=create_user]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Create User.
    create-webhook plan - Create Webhook. [intent=reverse_etl availability=implemented write=create_webhook]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Create Webhook.
    create-website plan - Create Website. [intent=reverse_etl availability=not_implemented write=create_website]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Create Website.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    customer-properties list - Read Help Scout List Customer Property Definitions as ETL records. [intent=etl availability=implemented stream=customer_properties]
    customers list - Read Help Scout List Customers as ETL records. [intent=etl availability=implemented stream=customers]
    customers-chats list - Read Help Scout List Chats Handles as ETL records. [intent=etl availability=implemented stream=customers_chats]; flags: --customer-id (required)
    customers-emails list - Read Help Scout List Emails as ETL records. [intent=etl availability=implemented stream=customers_emails]; flags: --customer-id (required)
    customers-phones list - Read Help Scout List Phones as ETL records. [intent=etl availability=implemented stream=customers_phones]; flags: --customer-id (required)
    customers-social-profiles list - Read Help Scout List Social Profiles as ETL records. [intent=etl availability=implemented stream=customers_social_profiles]; flags: --customer-id (required)
    customers-websites list - Read Help Scout List Websites as ETL records. [intent=etl availability=implemented stream=customers_websites]; flags: --customer-id (required)
    delete-address plan - Delete Address. [intent=reverse_etl availability=not_implemented write=delete_address]; approval: requires plan, preview, approval, and execute; risk: Destructive Help Scout mutation: Delete Address.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete-attachment plan - Delete Attachment. [intent=reverse_etl availability=not_implemented write=delete_attachment]; approval: requires plan, preview, approval, and execute; risk: Destructive Help Scout mutation: Delete Attachment.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete-chat-handle plan - Delete Chat Handle. [intent=reverse_etl availability=not_implemented write=delete_chat_handle]; approval: requires plan, preview, approval, and execute; risk: Destructive Help Scout mutation: Delete Chat Handle.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete-conversation plan - Delete Conversation. [intent=reverse_etl availability=not_implemented write=delete_conversation]; approval: requires plan, preview, approval, and execute; risk: Destructive Help Scout mutation: Delete Conversation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete-customer plan - Delete Customer. Hard-deletes the customer, their microsurvey responses and every conversation they participated in, in compliance with GDPR erasure. There is no undo. Pass async=true to run it asynchronously (202 Accepted instead of 204 No Content), which is the only form that works for a customer with more than 100 conversations. [intent=reverse_etl availability=not_implemented write=delete_customer]; approval: requires plan, preview, approval, and execute; risk: Destructive Help Scout mutation: Delete Customer. Hard-deletes the customer, their microsurvey responses and every conversation they participated in, in compliance with GDPR erasure. There is no undo. Pass async=true to run it asynchronously (202 Accepted instead of 204 No Content), which is the only form that works for a customer with more than 100 conversations.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete-customer-property-definition plan - Delete Customer Property Definition. [intent=reverse_etl availability=not_implemented write=delete_customer_property_definition]; approval: requires plan, preview, approval, and execute; risk: Destructive Help Scout mutation: Delete Customer Property Definition.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete-email plan - Delete Email. [intent=reverse_etl availability=not_implemented write=delete_email]; approval: requires plan, preview, approval, and execute; risk: Destructive Help Scout mutation: Delete Email.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete-organization-by-id plan - Delete Organization by ID. [intent=reverse_etl availability=not_implemented write=delete_organization_by_id]; approval: requires plan, preview, approval, and execute; risk: Destructive Help Scout mutation: Delete Organization by ID.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete-organization-property-definition plan - Delete Organization Property Definition. [intent=reverse_etl availability=not_implemented write=delete_organization_property_definition]; approval: requires plan, preview, approval, and execute; risk: Destructive Help Scout mutation: Delete Organization Property Definition.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete-phone plan - Delete Phone. [intent=reverse_etl availability=not_implemented write=delete_phone]; approval: requires plan, preview, approval, and execute; risk: Destructive Help Scout mutation: Delete Phone.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete-saved-reply plan - Delete Saved Reply. [intent=reverse_etl availability=not_implemented write=delete_saved_reply]; approval: requires plan, preview, approval, and execute; risk: Destructive Help Scout mutation: Delete Saved Reply.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete-snooze plan - Delete Snooze. [intent=reverse_etl availability=not_implemented write=delete_snooze]; approval: requires plan, preview, approval, and execute; risk: Destructive Help Scout mutation: Delete Snooze.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete-social-profile plan - Delete Social Profile. [intent=reverse_etl availability=not_implemented write=delete_social_profile]; approval: requires plan, preview, approval, and execute; risk: Destructive Help Scout mutation: Delete Social Profile.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete-thread-schedule plan - Delete Thread Schedule. [intent=reverse_etl availability=not_implemented write=delete_thread_schedule]; approval: requires plan, preview, approval, and execute; risk: Destructive Help Scout mutation: Delete Thread Schedule.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete-user plan - Delete User. [intent=reverse_etl availability=not_implemented write=delete_user]; approval: requires plan, preview, approval, and execute; risk: Destructive Help Scout mutation: Delete User.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete-webhook plan - Delete Webhook. [intent=reverse_etl availability=not_implemented write=delete_webhook]; approval: requires plan, preview, approval, and execute; risk: Destructive Help Scout mutation: Delete Webhook.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    delete-website plan - Delete Website. [intent=reverse_etl availability=not_implemented write=delete_website]; approval: requires plan, preview, approval, and execute; risk: Destructive Help Scout mutation: Delete Website.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    direct all-channels-volumes-by-channel - All Channels - Volumes by Channel. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct chat-report - Chat Report. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct company-customers-helped - Company Customers Helped. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct company-drilldown - Company Drilldown. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct company-overall-report - Company Overall Report. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct conversations-busiest-time-of-day - Conversations - Busiest Time of Day. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct conversations-drilldown - Conversations - Drilldown. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct conversations-drilldown-by-field - Conversations - Drilldown by Field. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct conversations-new-conversations - Conversations - New Conversations. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct conversations-new-conversations-drilldown - Conversations - New Conversations Drilldown. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct conversations-overall-report - Conversations - Overall Report. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct conversations-received-messages-statistics - Conversations - Received Messages Statistics. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct docs-overall-report - Docs Overall Report. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct email-report - Email Report. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct get-address - Get Address. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --customerId, --page, --page-cursor
    direct get-attachment-data - Get Attachment Data. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --conversationId, --attachmentId, --page, --page-cursor
    direct get-conversation - Get Conversation. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --conversationId, --page, --page-cursor
    direct get-conversation-reassignment-configuration - Get Conversation Reassignment configuration. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --userId, --page, --page-cursor
    direct get-customer - Get Customer. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --customerId, --page, --page-cursor
    direct get-inbox - Get Inbox. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --mailboxId, --page, --page-cursor
    direct get-organization-by-id - Get Organization by ID. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --organizationId, --page, --page-cursor
    direct get-organization-property-definition - Get Organization Property Definition. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --slug, --page, --page-cursor
    direct get-resource-owner - Get Resource Owner. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct get-routing-configuration - Get Routing configuration. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --mailboxId, --page, --page-cursor
    direct get-satisfaction-rating - Get Satisfaction Rating. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --ratingId, --page, --page-cursor
    direct get-saved-reply - Get Saved Reply. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --mailboxId, --savedReplyId, --page, --page-cursor
    direct get-tag-by-id - Get Tag by ID. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --tagId, --page, --page-cursor
    direct get-thread-original-source - Get Thread Original Source. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --conversationId, --threadId, --page, --page-cursor
    direct get-user - Get User. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --userId, --page, --page-cursor
    direct get-user-status - Get user status. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --userId, --page, --page-cursor
    direct get-webhook - Get Webhook. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --webhookId, --page, --page-cursor
    direct happiness-overall-report - Happiness Overall Report. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct happiness-ratings-report - Happiness Ratings Report. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct phone-report - Phone Report. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct productivity-first-response-time - Productivity - First Response Time. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct productivity-overall-report - Productivity Overall Report. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct productivity-replies-sent - Productivity - Replies Sent. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct productivity-resolution-time - Productivity - Resolution Time. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct productivity-resolved - Productivity - Resolved. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct productivity-response-time - Productivity - Response Time. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct user-conversation-history - User Conversation History. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct user-customers-helped - User Customers Helped. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct user-drill-down - User Drill-down. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct user-happiness - User Happiness. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct user-happiness-drilldown - User Happiness drilldown. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct user-replies - User Replies. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct user-resolutions - User Resolutions. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct user-team-chat-report - User/Team Chat Report. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    direct user-team-overall-report - User/Team Overall Report. [intent=direct_read availability=implemented]; risk: Bounded Help Scout JSON read; the response is size-limited and secret-shaped fields are redacted.; flags: --page, --page-cursor
    mailboxes list - Read Help Scout List Inboxes as ETL records. [intent=etl availability=implemented stream=mailboxes]
    mailboxes-fields list - Read Help Scout List Inbox Custom Fields as ETL records. [intent=etl availability=implemented stream=mailboxes_fields]; flags: --mailbox-id (required)
    mailboxes-folders list - Read Help Scout List Inbox Folders as ETL records. [intent=etl availability=implemented stream=mailboxes_folders]; flags: --mailbox-id (required)
    mailboxes-saved-replies list - Read Help Scout List Saved Replies as ETL records. [intent=etl availability=implemented stream=mailboxes_saved_replies]; flags: --mailbox-id (required)
    organizations list - Read Help Scout List Organizations as ETL records. [intent=etl availability=implemented stream=organizations]
    organizations-conversations list - Read Help Scout Get Organization Conversations as ETL records. [intent=etl availability=implemented stream=organizations_conversations]; flags: --organization-id (required)
    organizations-customers list - Read Help Scout Get Organization Customers as ETL records. [intent=etl availability=implemented stream=organizations_customers]; flags: --organization-id (required)
    organizations-properties list - Read Help Scout List Organization Property Definitions as ETL records. [intent=etl availability=implemented stream=organizations_properties]
    overwrite-customer plan - Overwrite Customer. [intent=reverse_etl availability=not_implemented write=overwrite_customer]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Overwrite Customer.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    publish-thread-schedule plan - Publish Thread Schedule. [intent=reverse_etl availability=not_implemented write=publish_thread_schedule]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Publish Thread Schedule.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    remove-organization-property-value plan - Remove Organization Property Value. [intent=reverse_etl availability=not_implemented write=remove_organization_property_value]; approval: requires plan, preview, approval, and execute; risk: Destructive Help Scout mutation: Remove Organization Property Value.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    run-manual-workflows plan - Run Manual Workflows. [intent=reverse_etl availability=not_implemented write=run_manual_workflows]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Run Manual Workflows.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    set-organization-property-value plan - Set Organization Property Value. [intent=reverse_etl availability=not_implemented write=set_organization_property_value]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Set Organization Property Value.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    set-user-status plan - Set user status. [intent=reverse_etl availability=not_implemented write=set_user_status]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Set user status.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    tags list - Read Help Scout List Tags as ETL records. [intent=etl availability=implemented stream=tags]
    teams list - Read Help Scout List Teams as ETL records. [intent=etl availability=implemented stream=teams]
    teams-members list - Read Help Scout List Team Members as ETL records. [intent=etl availability=implemented stream=teams_members]; flags: --team-id (required)
    update-address plan - Update Address. [intent=reverse_etl availability=not_implemented write=update_address]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Update Address.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update-chat-handles plan - Update Chat Handles. [intent=reverse_etl availability=not_implemented write=update_chat_handles]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Update Chat Handles.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update-conversation plan - Update Conversation. [intent=reverse_etl availability=not_implemented write=update_conversation]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Update Conversation.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update-conversation-reassignment-configuration plan - Update Conversation Reassignment configuration. [intent=reverse_etl availability=not_implemented write=update_conversation_reassignment_configuration]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Update Conversation Reassignment configuration.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update-custom-fields plan - Update Custom Fields. [intent=reverse_etl availability=not_implemented write=update_custom_fields]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Update Custom Fields.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update-customer plan - Update Customer. [intent=reverse_etl availability=not_implemented write=update_customer]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Update Customer.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update-customer-properties plan - Update Customer Properties. [intent=reverse_etl availability=not_implemented write=update_customer_properties]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Update Customer Properties.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update-email plan - Update Email. [intent=reverse_etl availability=not_implemented write=update_email]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Update Email.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update-organization plan - Update Organization. [intent=reverse_etl availability=not_implemented write=update_organization]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Update Organization.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update-organization-property-definition plan - Update Organization Property Definition. [intent=reverse_etl availability=not_implemented write=update_organization_property_definition]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Update Organization Property Definition.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update-phone plan - Update Phone. [intent=reverse_etl availability=not_implemented write=update_phone]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Update Phone.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update-routing-configuration plan - Update Routing configuration. [intent=reverse_etl availability=not_implemented write=update_routing_configuration]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Update Routing configuration.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update-saved-reply plan - Update Saved Reply. [intent=reverse_etl availability=not_implemented write=update_saved_reply]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Update Saved Reply.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update-snooze plan - Update Snooze. [intent=reverse_etl availability=not_implemented write=update_snooze]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Update Snooze.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update-social-profile plan - Update Social Profile. [intent=reverse_etl availability=not_implemented write=update_social_profile]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Update Social Profile.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update-tags plan - Update Tags. [intent=reverse_etl availability=not_implemented write=update_tags]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Update Tags.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update-team-members plan - Update Team Members. [intent=reverse_etl availability=not_implemented write=update_team_members]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Update Team Members.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update-thread plan - Update Thread. [intent=reverse_etl availability=not_implemented write=update_thread]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Update Thread.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update-thread-schedule plan - Update Thread Schedule. [intent=reverse_etl availability=not_implemented write=update_thread_schedule]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Update Thread Schedule.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update-webhook plan - Update Webhook. [intent=reverse_etl availability=not_implemented write=update_webhook]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Update Webhook.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update-website plan - Update Website. [intent=reverse_etl availability=not_implemented write=update_website]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Update Website.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update-workflow-status plan - Update workflow status. [intent=reverse_etl availability=not_implemented write=update_workflow_status]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Update workflow status.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    upload-attachment plan - Upload Attachment. [intent=reverse_etl availability=not_implemented write=upload_attachment]; approval: requires plan, preview, approval, and execute; risk: Help Scout mutation: Upload Attachment.; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    users list - Read Help Scout List Users as ETL records. [intent=etl availability=implemented stream=users]
    users-status list - Read Help Scout List users statuses as ETL records. [intent=etl availability=implemented stream=users_status]
    webhooks list - Read Help Scout List Webhooks as ETL records. [intent=etl availability=implemented stream=webhooks]
    workflows list - Read Help Scout List Workflows as ETL records. [intent=etl availability=implemented stream=workflows]

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
