# pm connectors inspect intercom

```text
NAME
  pm connectors inspect intercom - Intercom connector manual

SYNOPSIS
  pm connectors inspect intercom
  pm connectors inspect intercom --json
  pm credentials add <name> --connector intercom [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Intercom contacts, companies, conversations, admins, and tags; declares typed reverse-ETL write actions for official Intercom API mutations.

ICON
  asset: icons/intercom.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.intercom.com/docs/build-an-integration/learn-more/rest-apis/unversioned-changes#unversioned-changes

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  api_version
  base_url
  page_size
  access_token (secret)

ETL STREAMS
  contacts:
    primary key: id
    cursor: updated_at
    fields: created_at(), email(), external_id(), id(), last_seen_at(), name(), owner_id(), phone(), role(), signed_up_at(), type(), unsubscribed_from_emails(), updated_at()
  companies:
    primary key: id
    cursor: updated_at
    fields: company_id(), created_at(), id(), industry(), last_request_at(), monthly_spend(), name(), session_count(), size(), type(), updated_at(), user_count(), website()
  conversations:
    primary key: id
    cursor: updated_at
    fields: admin_assignee_id(), created_at(), id(), open(), priority(), read(), snoozed_until(), state(), title(), type(), updated_at(), waiting_since()
  admins:
    primary key: id
    fields: away_mode_enabled(), away_mode_reassign(), email(), has_inbox_seat(), id(), job_title(), name(), type()
  tags:
    primary key: id
    fields: id(), name(), type()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  set_away_admin:
    endpoint: PUT /admins/{{ record.admin_id }}/away
    required fields: admin_id, away_mode_enabled, away_mode_reassign
    optional fields: away_status_reason_id
    risk: Set an admin to away: live Intercom mutation against /admins/{admin_id}/away; reverse ETL requires plan, preview, explicit approval, execute
  create_content_import_source:
    endpoint: POST /ai/content_import_sources
    required fields: sync_behavior, url
    optional fields: status, audience_ids
    risk: Create a content import source: live Intercom mutation against /ai/content_import_sources; reverse ETL requires plan, preview, explicit approval, execute
  delete_content_import_source:
    endpoint: DELETE /ai/content_import_sources/{{ record.source_id }}
    required fields: source_id
    risk: Delete a content import source: live Intercom mutation against /ai/content_import_sources/{source_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  update_content_import_source:
    endpoint: PUT /ai/content_import_sources/{{ record.source_id }}
    required fields: source_id, sync_behavior, url
    optional fields: status, audience_ids, apply_audience_to_existing_content
    risk: Update a content import source: live Intercom mutation against /ai/content_import_sources/{source_id}; reverse ETL requires plan, preview, explicit approval, execute
  create_external_page:
    endpoint: POST /ai/external_pages
    required fields: title, html, locale, source_id, external_id
    optional fields: url, ai_agent_availability, ai_copilot_availability
    risk: Create an external page (or update an external page by external ID): live Intercom mutation against /ai/external_pages; reverse ETL requires plan, preview, explicit approval, execute
  delete_external_page:
    endpoint: DELETE /ai/external_pages/{{ record.page_id }}
    required fields: page_id
    risk: Delete an external page: live Intercom mutation against /ai/external_pages/{page_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  update_external_page:
    endpoint: PUT /ai/external_pages/{{ record.page_id }}
    required fields: page_id, title, html, url, locale, source_id
    optional fields: fin_availability, external_id
    risk: Update an external page: live Intercom mutation against /ai/external_pages/{page_id}; reverse ETL requires plan, preview, explicit approval, execute
  create_article:
    endpoint: POST /articles
    required fields: title, author_id
    optional fields: description, body, body_markdown, state, parent_id, parent_type, translated_content, audience_ids, ai_chatbot_availability, ai_copilot_availability, ai_sales_agent_availability, scheduled_publish_at, scheduled_unpublish_at
    risk: Create an article: live Intercom mutation against /articles; reverse ETL requires plan, preview, explicit approval, execute
  delete_article:
    endpoint: DELETE /articles/{{ record.article_id }}
    required fields: article_id
    risk: Delete an article: live Intercom mutation against /articles/{article_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  update_article:
    endpoint: PUT /articles/{{ record.article_id }}
    required fields: article_id
    optional fields: title, description, body, body_markdown, author_id, state, parent_id, parent_type, translated_content, audience_ids, ai_chatbot_availability, ai_copilot_availability, ai_sales_agent_availability, scheduled_publish_at, scheduled_unpublish_at
    risk: Update an article: live Intercom mutation against /articles/{article_id}; reverse ETL requires plan, preview, explicit approval, execute
  attach_tag_to_article:
    endpoint: POST /articles/{{ record.article_id }}/tags
    required fields: article_id, id
    optional fields: admin_id
    risk: Add a tag to an article: live Intercom mutation against /articles/{article_id}/tags; reverse ETL requires plan, preview, explicit approval, execute
  detach_tag_from_article:
    endpoint: DELETE /articles/{{ record.article_id }}/tags/{{ record.id }}
    required fields: article_id, id
    risk: Remove a tag from an article: live Intercom mutation against /articles/{article_id}/tags/{id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  stage_article_draft:
    endpoint: PUT /articles/{{ record.id }}/draft
    required fields: id
    optional fields: title, description, body, body_markdown, author_id, state, parent_id, parent_type, translated_content, audience_ids, ai_chatbot_availability, ai_copilot_availability, ai_sales_agent_availability, scheduled_publish_at, scheduled_unpublish_at
    risk: Stage an article draft: live Intercom mutation against /articles/{id}/draft; reverse ETL requires plan, preview, explicit approval, execute
  publish_article_draft:
    endpoint: POST /articles/{{ record.id }}/draft/publish
    required fields: id
    optional fields: locales
    risk: Publish an article draft: live Intercom mutation against /articles/{id}/draft/publish; reverse ETL requires plan, preview, explicit approval, execute
  create_audience:
    endpoint: POST /audiences
    required fields: name
    optional fields: predicates, role_predicates
    risk: Create an audience: live Intercom mutation against /audiences; reverse ETL requires plan, preview, explicit approval, execute
  delete_audience:
    endpoint: DELETE /audiences/{{ record.id }}
    required fields: id
    risk: Delete an audience: live Intercom mutation against /audiences/{id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  update_audience:
    endpoint: PUT /audiences/{{ record.id }}
    required fields: id
    optional fields: name, predicates, role_predicates
    risk: Update an audience: live Intercom mutation against /audiences/{id}; reverse ETL requires plan, preview, explicit approval, execute
  create_or_update_company:
    endpoint: POST /companies
    optional fields: name, company_id, plan, size, website, industry, custom_attributes, remote_created_at, update_last_request_at, monthly_spend
    risk: Create or Update a company: live Intercom mutation against /companies; reverse ETL requires plan, preview, explicit approval, execute
  delete_company:
    endpoint: DELETE /companies/{{ record.company_id }}
    required fields: company_id
    risk: Delete a company: live Intercom mutation against /companies/{company_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  update_company:
    endpoint: PUT /companies/{{ record.company_id }}
    required fields: company_id
    optional fields: name, plan, size, website, industry, custom_attributes, monthly_spend
    risk: Update a company: live Intercom mutation against /companies/{company_id}; reverse ETL requires plan, preview, explicit approval, execute
  create_company_note:
    endpoint: POST /companies/{{ record.company_id }}/notes
    required fields: company_id, body
    optional fields: admin_id
    risk: Create a company note: live Intercom mutation against /companies/{company_id}/notes; reverse ETL requires plan, preview, explicit approval, execute
  create_contact:
    endpoint: POST /contacts
    optional fields: role, external_id, email_verified, email, phone, name, avatar, signed_up_at, last_seen_at, owner_id, unsubscribed_from_emails, custom_attributes
    risk: Create contact: live Intercom mutation against /contacts; reverse ETL requires plan, preview, explicit approval, execute
  merge_contact:
    endpoint: POST /contacts/merge
    required fields: from, into
    optional fields: skip_duplicate_validation
    risk: Merge a lead and a user: live Intercom mutation against /contacts/merge; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  delete_contact:
    endpoint: DELETE /contacts/{{ record.contact_id }}
    required fields: contact_id
    risk: Delete a contact: live Intercom mutation against /contacts/{contact_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  update_contact:
    endpoint: PUT /contacts/{{ record.contact_id }}
    required fields: contact_id
    optional fields: role, external_id, email, email_verified, phone, name, avatar, signed_up_at, last_seen_at, owner_id, unsubscribed_from_emails, custom_attributes
    risk: Update a contact: live Intercom mutation against /contacts/{contact_id}; reverse ETL requires plan, preview, explicit approval, execute
  archive_contact:
    endpoint: POST /contacts/{{ record.contact_id }}/archive
    required fields: contact_id
    risk: Archive contact: live Intercom mutation against /contacts/{contact_id}/archive; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  block_contact:
    endpoint: POST /contacts/{{ record.contact_id }}/block
    required fields: contact_id
    risk: Block contact: live Intercom mutation against /contacts/{contact_id}/block; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  attach_contact_to_acompany:
    endpoint: POST /contacts/{{ record.contact_id }}/companies
    required fields: contact_id, id
    risk: Attach a Contact to a Company: live Intercom mutation against /contacts/{contact_id}/companies; reverse ETL requires plan, preview, explicit approval, execute
  detach_contact_from_acompany:
    endpoint: DELETE /contacts/{{ record.contact_id }}/companies/{{ record.company_id }}
    required fields: contact_id, company_id
    risk: Detach a contact from a company: live Intercom mutation against /contacts/{contact_id}/companies/{company_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  create_note:
    endpoint: POST /contacts/{{ record.contact_id }}/notes
    required fields: contact_id, body
    optional fields: admin_id
    risk: Create a note: live Intercom mutation against /contacts/{contact_id}/notes; reverse ETL requires plan, preview, explicit approval, execute
  attach_subscription_type_to_contact:
    endpoint: POST /contacts/{{ record.contact_id }}/subscriptions
    required fields: contact_id, id, consent_type
    risk: Add subscription to a contact: live Intercom mutation against /contacts/{contact_id}/subscriptions; reverse ETL requires plan, preview, explicit approval, execute
  detach_subscription_type_to_contact:
    endpoint: DELETE /contacts/{{ record.contact_id }}/subscriptions/{{ record.subscription_id }}
    required fields: contact_id, subscription_id
    risk: Remove subscription from a contact: live Intercom mutation against /contacts/{contact_id}/subscriptions/{subscription_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  attach_tag_to_contact:
    endpoint: POST /contacts/{{ record.contact_id }}/tags
    required fields: contact_id, id
    risk: Add tag to a contact: live Intercom mutation against /contacts/{contact_id}/tags; reverse ETL requires plan, preview, explicit approval, execute
  detach_tag_from_contact:
    endpoint: DELETE /contacts/{{ record.contact_id }}/tags/{{ record.tag_id }}
    required fields: contact_id, tag_id
    risk: Remove tag from a contact: live Intercom mutation against /contacts/{contact_id}/tags/{tag_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  unarchive_contact:
    endpoint: POST /contacts/{{ record.contact_id }}/unarchive
    required fields: contact_id
    risk: Unarchive contact: live Intercom mutation against /contacts/{contact_id}/unarchive; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  dismiss_contact_banner:
    endpoint: POST /contacts/{{ record.id }}/banners/{{ record.view_id }}/dismiss
    required fields: id, view_id
    risk: Dismiss a banner for a contact: live Intercom mutation against /contacts/{id}/banners/{view_id}/dismiss; reverse ETL requires plan, preview, explicit approval, execute
  bulk_content_actions:
    endpoint: POST /content/bulk_actions
    required fields: action, content_ids
    optional fields: availability, audience, tags
    risk: Run a bulk action on Knowledge Hub content: live Intercom mutation against /content/bulk_actions; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  create_content_snippet:
    endpoint: POST /content_snippets
    required fields: title
    optional fields: json_blocks, body_markdown, locale, audience_ids, ai_chatbot_availability, ai_copilot_availability, ai_sales_agent_availability
    risk: Create a content snippet: live Intercom mutation against /content_snippets; reverse ETL requires plan, preview, explicit approval, execute
  attach_tag_to_content_snippet:
    endpoint: POST /content_snippets/{{ record.content_snippet_id }}/tags
    required fields: content_snippet_id, id
    optional fields: admin_id
    risk: Add a tag to a content snippet: live Intercom mutation against /content_snippets/{content_snippet_id}/tags; reverse ETL requires plan, preview, explicit approval, execute
  detach_tag_from_content_snippet:
    endpoint: DELETE /content_snippets/{{ record.content_snippet_id }}/tags/{{ record.id }}
    required fields: content_snippet_id, id
    risk: Remove a tag from a content snippet: live Intercom mutation against /content_snippets/{content_snippet_id}/tags/{id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  delete_content_snippet:
    endpoint: DELETE /content_snippets/{{ record.id }}
    required fields: id
    risk: Delete a content snippet: live Intercom mutation against /content_snippets/{id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  update_content_snippet:
    endpoint: PUT /content_snippets/{{ record.id }}
    required fields: id
    optional fields: title, json_blocks, body_markdown, locale, audience_ids, ai_chatbot_availability, ai_copilot_availability, ai_sales_agent_availability
    risk: Update a content snippet: live Intercom mutation against /content_snippets/{id}; reverse ETL requires plan, preview, explicit approval, execute
  create_conversation:
    endpoint: POST /conversations
    required fields: from, body
    optional fields: subject, attachment_urls, created_at, brand_id
    risk: Creates a conversation: live Intercom mutation against /conversations; reverse ETL requires plan, preview, explicit approval, execute
  create_conversation_attribute:
    endpoint: POST /conversations/attributes
    optional fields: name, description, data_type, required, visible_to_team_ids, multiline, options, reference
    risk: Create a conversation attribute: live Intercom mutation against /conversations/attributes; reverse ETL requires plan, preview, explicit approval, execute
  delete_conversation_attribute:
    endpoint: DELETE /conversations/attributes/{{ record.id }}
    required fields: id
    risk: Delete (archive) a conversation attribute: live Intercom mutation against /conversations/attributes/{id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  update_conversation_attribute:
    endpoint: PUT /conversations/attributes/{{ record.id }}
    required fields: id
    optional fields: name, description, multiline, required, visible_to_team_ids, reference
    risk: Update a conversation attribute: live Intercom mutation against /conversations/attributes/{id}; reverse ETL requires plan, preview, explicit approval, execute
  create_conversation_attribute_option:
    endpoint: POST /conversations/attributes/{{ record.id }}/options
    required fields: id, label
    risk: Add an option to a list conversation attribute: live Intercom mutation against /conversations/attributes/{id}/options; reverse ETL requires plan, preview, explicit approval, execute
  delete_conversation_attribute_option:
    endpoint: DELETE /conversations/attributes/{{ record.id }}/options/{{ record.option_id }}
    required fields: id, option_id
    risk: Archive an option on a list conversation attribute: live Intercom mutation against /conversations/attributes/{id}/options/{option_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  update_conversation_attribute_option:
    endpoint: PUT /conversations/attributes/{{ record.id }}/options/{{ record.option_id }}
    required fields: id, option_id, label
    risk: Update an option on a list conversation attribute: live Intercom mutation against /conversations/attributes/{id}/options/{option_id}; reverse ETL requires plan, preview, explicit approval, execute
  redact_conversation:
    endpoint: POST /conversations/redact
    optional fields: type, conversation_id, conversation_part_id, source_id
    risk: Redact a conversation part: live Intercom mutation against /conversations/redact; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  delete_conversation:
    endpoint: DELETE /conversations/{{ record.conversation_id }}
    required fields: conversation_id
    risk: Delete a conversation: live Intercom mutation against /conversations/{conversation_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  update_conversation:
    endpoint: PUT /conversations/{{ record.conversation_id }}
    required fields: conversation_id
    optional fields: read, title, custom_attributes, company_id
    risk: Update a conversation: live Intercom mutation against /conversations/{conversation_id}; reverse ETL requires plan, preview, explicit approval, execute
  convert_conversation_to_ticket:
    endpoint: POST /conversations/{{ record.conversation_id }}/convert
    required fields: conversation_id, ticket_type_id
    optional fields: ticket_state_id, attributes
    risk: Convert a conversation to a ticket: live Intercom mutation against /conversations/{conversation_id}/convert; reverse ETL requires plan, preview, explicit approval, execute
  attach_contact_to_conversation:
    endpoint: POST /conversations/{{ record.conversation_id }}/customers
    required fields: conversation_id
    optional fields: admin_id, customer
    risk: Attach a contact to a conversation: live Intercom mutation against /conversations/{conversation_id}/customers; reverse ETL requires plan, preview, explicit approval, execute
  detach_contact_from_conversation:
    endpoint: DELETE /conversations/{{ record.conversation_id }}/customers/{{ record.contact_id }}
    required fields: conversation_id, contact_id, admin_id
    risk: Detach a contact from a group conversation: live Intercom mutation against /conversations/{conversation_id}/customers/{contact_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  manage_conversation:
    endpoint: POST /conversations/{{ record.conversation_id }}/parts
    required fields: conversation_id
    optional fields: message_type, type, admin_id, body, snoozed_until, assignee_id
    risk: Manage a conversation: live Intercom mutation against /conversations/{conversation_id}/parts; reverse ETL requires plan, preview, explicit approval, execute
  reply_conversation:
    endpoint: POST /conversations/{{ record.conversation_id }}/reply
    required fields: conversation_id
    optional fields: message_type, type, body, created_at, attachment_urls, reply_options, intercom_user_id, attachment_files, email, user_id, admin_id, skip_notifications
    risk: Reply to a conversation: live Intercom mutation against /conversations/{conversation_id}/reply; reverse ETL requires plan, preview, explicit approval, execute
  attach_tag_to_conversation:
    endpoint: POST /conversations/{{ record.conversation_id }}/tags
    required fields: conversation_id, id, admin_id
    risk: Add tag to a conversation: live Intercom mutation against /conversations/{conversation_id}/tags; reverse ETL requires plan, preview, explicit approval, execute
  detach_tag_from_conversation:
    endpoint: DELETE /conversations/{{ record.conversation_id }}/tags/{{ record.tag_id }}
    required fields: conversation_id, tag_id, admin_id
    risk: Remove tag from a conversation: live Intercom mutation against /conversations/{conversation_id}/tags/{tag_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  merge_conversation:
    endpoint: POST /conversations/{{ record.id }}/merge
    required fields: id, merge_into_conversation_id
    risk: Merge a conversation: live Intercom mutation against /conversations/{id}/merge; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  delete_custom_object_instances_by_id:
    endpoint: DELETE /custom_object_instances/{{ record.custom_object_type_identifier }}
    required fields: custom_object_type_identifier
    risk: Delete a Custom Object Instance by External ID: live Intercom mutation against /custom_object_instances/{custom_object_type_identifier}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  create_custom_object_instances:
    endpoint: POST /custom_object_instances/{{ record.custom_object_type_identifier }}
    required fields: custom_object_type_identifier
    optional fields: external_id, external_created_at, external_updated_at, custom_attributes
    risk: Create or Update a Custom Object Instance: live Intercom mutation against /custom_object_instances/{custom_object_type_identifier}; reverse ETL requires plan, preview, explicit approval, execute
  delete_custom_object_instances_by_external_id:
    endpoint: DELETE /custom_object_instances/{{ record.custom_object_type_identifier }}/{{ record.custom_object_instance_id }}
    required fields: custom_object_type_identifier, custom_object_instance_id
    risk: Delete a Custom Object Instance by ID: live Intercom mutation against /custom_object_instances/{custom_object_type_identifier}/{custom_object_instance_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  create_data_attribute:
    endpoint: POST /data_attributes
    optional fields: data_type, options
    risk: Create a data attribute: live Intercom mutation against /data_attributes; reverse ETL requires plan, preview, explicit approval, execute
  update_data_attribute:
    endpoint: PUT /data_attributes/{{ record.data_attribute_id }}
    required fields: data_attribute_id
    optional fields: options
    risk: Update a data attribute: live Intercom mutation against /data_attributes/{data_attribute_id}; reverse ETL requires plan, preview, explicit approval, execute
  create_data_connector:
    endpoint: POST /data_connectors
    required fields: name
    optional fields: description, http_method, url, body, direct_fin_usage, audiences, headers, data_inputs, customer_authentication, bypass_authentication, validate_missing_attributes, mock_response, token_ids
    risk: Create a data connector: live Intercom mutation against /data_connectors; reverse ETL requires plan, preview, explicit approval, execute
  delete_data_connector:
    endpoint: DELETE /data_connectors/{{ record.id }}
    required fields: id
    risk: Delete a data connector: live Intercom mutation against /data_connectors/{id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  update_data_connector:
    endpoint: PATCH /data_connectors/{{ record.id }}
    required fields: id
    optional fields: name, description, state, http_method, url, body, direct_fin_usage, audiences, headers, data_inputs, customer_authentication, bypass_authentication, validate_missing_attributes, mock_response, token_ids
    risk: Update a data connector: live Intercom mutation against /data_connectors/{id}; reverse ETL requires plan, preview, explicit approval, execute
  create_data_event:
    endpoint: POST /events
    required fields: event_name, created_at
    optional fields: user_id, id, email, metadata
    risk: Submit a data event: live Intercom mutation against /events; reverse ETL requires plan, preview, explicit approval, execute
  cancel_data_export:
    endpoint: POST /export/cancel/{{ record.job_identifier }}
    required fields: job_identifier
    risk: Cancel content data export: live Intercom mutation against /export/cancel/{job_identifier}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  submit_fin_csat:
    endpoint: POST /fin/csat
    required fields: conversation_id, rating
    optional fields: remark
    risk: Submit a CSAT rating: live Intercom mutation against /fin/csat; reverse ETL requires plan, preview, explicit approval, execute
  reply_to_fin:
    endpoint: POST /fin/reply
    required fields: conversation_id, message, user
    optional fields: attachments
    risk: Reply to Fin: live Intercom mutation against /fin/reply; reverse ETL requires plan, preview, explicit approval, execute
  start_fin_conversation:
    endpoint: POST /fin/start
    required fields: conversation_id, message, user
    optional fields: attachments, conversation_metadata
    risk: Start a conversation with Fin: live Intercom mutation against /fin/start; reverse ETL requires plan, preview, explicit approval, execute
  register_fin_voice_call:
    endpoint: POST /fin_voice/register
    required fields: phone_number, call_id
    optional fields: source, data
    risk: Register a Fin Voice call: live Intercom mutation against /fin_voice/register; reverse ETL requires plan, preview, explicit approval, execute
  create_collection:
    endpoint: POST /help_center/collections
    required fields: name
    optional fields: description, translated_content, parent_id, help_center_id
    risk: Create a collection: live Intercom mutation against /help_center/collections; reverse ETL requires plan, preview, explicit approval, execute
  delete_collection:
    endpoint: DELETE /help_center/collections/{{ record.collection_id }}
    required fields: collection_id
    risk: Delete a collection: live Intercom mutation against /help_center/collections/{collection_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  update_collection:
    endpoint: PUT /help_center/collections/{{ record.collection_id }}
    required fields: collection_id
    optional fields: name, description, translated_content, parent_id
    risk: Update a collection: live Intercom mutation against /help_center/collections/{collection_id}; reverse ETL requires plan, preview, explicit approval, execute
  create_help_center_redirect:
    endpoint: POST /help_center/help_centers/{{ record.help_center_id }}/redirects
    required fields: help_center_id, from_url, locale, target_type, target_id
    risk: Create a redirect: live Intercom mutation against /help_center/help_centers/{help_center_id}/redirects; reverse ETL requires plan, preview, explicit approval, execute
  delete_help_center_redirect:
    endpoint: DELETE /help_center/help_centers/{{ record.help_center_id }}/redirects/{{ record.id }}
    required fields: help_center_id, id
    risk: Delete a redirect: live Intercom mutation against /help_center/help_centers/{help_center_id}/redirects/{id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  create_internal_article:
    endpoint: POST /internal_articles
    required fields: title, owner_id, author_id
    optional fields: body, body_markdown, audience_ids, ai_chatbot_availability, ai_copilot_availability, ai_sales_agent_availability
    risk: Create an internal article: live Intercom mutation against /internal_articles; reverse ETL requires plan, preview, explicit approval, execute
  delete_internal_article:
    endpoint: DELETE /internal_articles/{{ record.internal_article_id }}
    required fields: internal_article_id
    risk: Delete an internal article: live Intercom mutation against /internal_articles/{internal_article_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  update_internal_article:
    endpoint: PUT /internal_articles/{{ record.internal_article_id }}
    required fields: internal_article_id
    optional fields: title, body, body_markdown, author_id, owner_id, audience_ids, ai_chatbot_availability, ai_copilot_availability, ai_sales_agent_availability
    risk: Update an internal article: live Intercom mutation against /internal_articles/{internal_article_id}; reverse ETL requires plan, preview, explicit approval, execute
  attach_tag_to_internal_article:
    endpoint: POST /internal_articles/{{ record.internal_article_id }}/tags
    required fields: internal_article_id, id
    optional fields: admin_id
    risk: Add a tag to an internal article: live Intercom mutation against /internal_articles/{internal_article_id}/tags; reverse ETL requires plan, preview, explicit approval, execute
  detach_tag_from_internal_article:
    endpoint: DELETE /internal_articles/{{ record.internal_article_id }}/tags/{{ record.id }}
    required fields: internal_article_id, id
    risk: Remove a tag from an internal article: live Intercom mutation against /internal_articles/{internal_article_id}/tags/{id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  update_ip_allowlist:
    endpoint: PUT /ip_allowlist
    optional fields: type, enabled, ip_allowlist
    risk: Update IP allowlist settings: live Intercom mutation against /ip_allowlist; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  create_message:
    endpoint: POST /messages
    required fields: message_type, from, to
    optional fields: subject, body, template, cc, bcc, components, created_at, create_conversation_without_contact_reply
    risk: Create a message: live Intercom mutation against /messages; reverse ETL requires plan, preview, explicit approval, execute
  create_news_item:
    endpoint: POST /news/news_items
    required fields: title, sender_id
    optional fields: body, state, deliver_silently, labels, reactions, newsfeed_assignments
    risk: Create a news item: live Intercom mutation against /news/news_items; reverse ETL requires plan, preview, explicit approval, execute
  delete_news_item:
    endpoint: DELETE /news/news_items/{{ record.news_item_id }}
    required fields: news_item_id
    risk: Delete a news item: live Intercom mutation against /news/news_items/{news_item_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  update_news_item:
    endpoint: PUT /news/news_items/{{ record.news_item_id }}
    required fields: news_item_id, title, sender_id
    optional fields: body, state, deliver_silently, labels, reactions, newsfeed_assignments
    risk: Update a news item: live Intercom mutation against /news/news_items/{news_item_id}; reverse ETL requires plan, preview, explicit approval, execute
  create_office_hours_schedule:
    endpoint: POST /office_hours_schedules
    required fields: name, time_zone_name, time_intervals
    risk: Create an office hours schedule: live Intercom mutation against /office_hours_schedules; reverse ETL requires plan, preview, explicit approval, execute
  delete_office_hours_schedule:
    endpoint: DELETE /office_hours_schedules/{{ record.id }}
    required fields: id
    risk: Delete an office hours schedule: live Intercom mutation against /office_hours_schedules/{id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  update_office_hours_schedule:
    endpoint: PUT /office_hours_schedules/{{ record.id }}
    required fields: id
    optional fields: name, time_zone_name, time_intervals
    risk: Update an office hours schedule: live Intercom mutation against /office_hours_schedules/{id}; reverse ETL requires plan, preview, explicit approval, execute
  create_office_hours_exception:
    endpoint: POST /office_hours_schedules/{{ record.office_hours_schedule_id }}/office_hours_exceptions
    required fields: office_hours_schedule_id, exception_date, exception_type
    optional fields: name, time_intervals, recurring_annually
    risk: Create an office hours exception: live Intercom mutation against /office_hours_schedules/{office_hours_schedule_id}/office_hours_exceptions; reverse ETL requires plan, preview, explicit approval, execute
  delete_office_hours_exception:
    endpoint: DELETE /office_hours_schedules/{{ record.office_hours_schedule_id }}/office_hours_exceptions/{{ record.id }}
    required fields: office_hours_schedule_id, id
    risk: Delete an office hours exception: live Intercom mutation against /office_hours_schedules/{office_hours_schedule_id}/office_hours_exceptions/{id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  update_office_hours_exception:
    endpoint: PUT /office_hours_schedules/{{ record.office_hours_schedule_id }}/office_hours_exceptions/{{ record.id }}
    required fields: office_hours_schedule_id, id
    optional fields: exception_date, exception_type, name, time_intervals, recurring_annually
    risk: Update an office hours exception: live Intercom mutation against /office_hours_schedules/{office_hours_schedule_id}/office_hours_exceptions/{id}; reverse ETL requires plan, preview, explicit approval, execute
  create_phone_switch:
    endpoint: POST /phone_call_redirects
    required fields: phone
    optional fields: custom_attributes
    risk: Create a phone Switch: live Intercom mutation against /phone_call_redirects; reverse ETL requires plan, preview, explicit approval, execute
  create_tag:
    endpoint: POST /tags
    optional fields: name, id, companies, users
    risk: Create or update a tag, Tag or untag companies, Tag contacts: live Intercom mutation against /tags; reverse ETL requires plan, preview, explicit approval, execute
  delete_tag:
    endpoint: DELETE /tags/{{ record.tag_id }}
    required fields: tag_id
    risk: Delete tag: live Intercom mutation against /tags/{tag_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  create_ticket_type:
    endpoint: POST /ticket_types
    required fields: name
    optional fields: description, category, icon, is_internal
    risk: Create a ticket type: live Intercom mutation against /ticket_types; reverse ETL requires plan, preview, explicit approval, execute
  update_ticket_type:
    endpoint: PUT /ticket_types/{{ record.ticket_type_id }}
    required fields: ticket_type_id
    optional fields: name, description, category, icon, archived, is_internal
    risk: Update a ticket type: live Intercom mutation against /ticket_types/{ticket_type_id}; reverse ETL requires plan, preview, explicit approval, execute
  create_ticket_type_attribute:
    endpoint: POST /ticket_types/{{ record.ticket_type_id }}/attributes
    required fields: ticket_type_id, name, description, data_type
    optional fields: required_to_create, required_to_create_for_contacts, visible_on_create, visible_to_contacts, multiline, list_items, allow_multiple_values
    risk: Create a new attribute for a ticket type: live Intercom mutation against /ticket_types/{ticket_type_id}/attributes; reverse ETL requires plan, preview, explicit approval, execute
  update_ticket_type_attribute:
    endpoint: PUT /ticket_types/{{ record.ticket_type_id }}/attributes/{{ record.attribute_id }}
    required fields: ticket_type_id, attribute_id
    optional fields: name, description, required_to_create, required_to_create_for_contacts, visible_on_create, visible_to_contacts, multiline, list_items, allow_multiple_values, archived
    risk: Update an existing attribute for a ticket type: live Intercom mutation against /ticket_types/{ticket_type_id}/attributes/{attribute_id}; reverse ETL requires plan, preview, explicit approval, execute
  create_ticket:
    endpoint: POST /tickets
    required fields: contacts, ticket_type_id
    optional fields: conversation_to_link_id, company_id, created_at, ticket_attributes, assignment, skip_notifications
    risk: Create a ticket: live Intercom mutation against /tickets; reverse ETL requires plan, preview, explicit approval, execute
  enqueue_create_ticket:
    endpoint: POST /tickets/enqueue
    required fields: contacts, ticket_type_id
    optional fields: conversation_to_link_id, company_id, created_at, ticket_attributes, assignment, skip_notifications
    risk: Enqueue create ticket: live Intercom mutation against /tickets/enqueue; reverse ETL requires plan, preview, explicit approval, execute
  delete_ticket:
    endpoint: DELETE /tickets/{{ record.ticket_id }}
    required fields: ticket_id
    risk: Delete a ticket: live Intercom mutation against /tickets/{ticket_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  update_ticket:
    endpoint: PUT /tickets/{{ record.ticket_id }}
    required fields: ticket_id
    optional fields: ticket_attributes, ticket_state_id, company_id, open, is_shared, snoozed_until, admin_id, assignee_id, skip_notifications
    risk: Update a ticket: live Intercom mutation against /tickets/{ticket_id}; reverse ETL requires plan, preview, explicit approval, execute
  change_ticket_type:
    endpoint: POST /tickets/{{ record.ticket_id }}/change_type
    required fields: ticket_id, ticket_type_id, ticket_state_id
    optional fields: ticket_attributes
    risk: Change ticket type: live Intercom mutation against /tickets/{ticket_id}/change_type; reverse ETL requires plan, preview, explicit approval, execute
  link_conversation_to_ticket:
    endpoint: POST /tickets/{{ record.ticket_id }}/linked_conversations
    required fields: ticket_id, conversation_id
    risk: Link a conversation to a ticket: live Intercom mutation against /tickets/{ticket_id}/linked_conversations; reverse ETL requires plan, preview, explicit approval, execute
  unlink_conversation_from_ticket:
    endpoint: DELETE /tickets/{{ record.ticket_id }}/linked_conversations/{{ record.id }}
    required fields: ticket_id, id
    risk: Unlink a conversation from a ticket: live Intercom mutation against /tickets/{ticket_id}/linked_conversations/{id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  reply_ticket:
    endpoint: POST /tickets/{{ record.ticket_id }}/reply
    required fields: ticket_id
    optional fields: message_type, type, body, created_at, attachment_urls, reply_options, intercom_user_id, user_id, email, admin_id, attachment_files, cross_post
    risk: Reply to a ticket: live Intercom mutation against /tickets/{ticket_id}/reply; reverse ETL requires plan, preview, explicit approval, execute
  attach_tag_to_ticket:
    endpoint: POST /tickets/{{ record.ticket_id }}/tags
    required fields: ticket_id, id, admin_id
    risk: Add tag to a ticket: live Intercom mutation against /tickets/{ticket_id}/tags; reverse ETL requires plan, preview, explicit approval, execute
  detach_tag_from_ticket:
    endpoint: DELETE /tickets/{{ record.ticket_id }}/tags/{{ record.tag_id }}
    required fields: ticket_id, tag_id, admin_id
    risk: Remove tag from a ticket: live Intercom mutation against /tickets/{ticket_id}/tags/{tag_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation
  update_visitor:
    endpoint: PUT /visitors
    optional fields: id, user_id, name, custom_attributes
    risk: Update a visitor: live Intercom mutation against /visitors; reverse ETL requires plan, preview, explicit approval, execute
  convert_visitor:
    endpoint: POST /visitors/convert
    required fields: type, user, visitor
    risk: Convert a visitor: live Intercom mutation against /visitors/convert; reverse ETL requires plan, preview, explicit approval, execute

SECURITY
  read risk: external Intercom API read of contact, company, conversation, admin, tag, and operational data
  write risk: Intercom mutations can create, update, archive, block, merge, redact, detach, or delete provider data; reverse ETL requires plan, preview, explicit approval, execute, and destructive confirmation where declared
  approval: reverse ETL writes require plan -> preview -> explicit approval -> execute; destructive actions require typed confirmation value `destructive`
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Intercom connector operations are definition-owned, typed, and bounded; live writes require reverse-ETL approval gates.
  Usage: pm intercom <resource> <operation> [--json]
  Source CLI: Intercom REST API (OpenAPI 2.16 official shared JSON)
  Global flags:
    --json (boolean): Emit machine-readable JSON when supported.
    --limit (integer): Bound local output rows for ETL/direct read commands.
  Implemented ETL streams
  Planned typed operation candidates
  Other Commands
    admins list - List Intercom admins as ETL records. [intent=etl availability=implemented stream=admins]
    admins list all activity log event types - List all activity log event types [intent=direct_read availability=planned]; approval: Blocked pending CDC truthfulness/state foundations #2986/#2988.; risk: Reads operational/audit/changefeed-like provider data.; notes: Official Intercom operation ledger row; lane=cdc_changefeed. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    admins list all activity logs - List all activity logs [intent=direct_read availability=planned]; approval: Blocked pending CDC truthfulness/state foundations #2986/#2988.; risk: Reads operational/audit/changefeed-like provider data.; notes: Official Intercom operation ledger row; lane=cdc_changefeed. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    admins search activity logs - Search activity logs [intent=direct_read availability=planned]; approval: Blocked pending CDC truthfulness/state foundations #2986/#2988.; risk: Reads operational/audit/changefeed-like provider data.; notes: Official Intercom operation ledger row; lane=cdc_changefeed. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    admins retrieve an admin - Retrieve an admin [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    admins set an admin to away - Set an admin to away [intent=reverse_etl availability=planned write=set_away_admin]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Set an admin to away: live Intercom mutation against /admins/{admin_id}/away; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    ai content list content import sources - List content import sources [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    ai content create a content import source - Create a content import source [intent=reverse_etl availability=planned write=create_content_import_source]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Create a content import source: live Intercom mutation against /ai/content_import_sources; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    ai content delete a content import source - Delete a content import source [intent=reverse_etl availability=planned write=delete_content_import_source]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Delete a content import source: live Intercom mutation against /ai/content_import_sources/{source_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    ai content retrieve a content import source - Retrieve a content import source [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    ai content update a content import source - Update a content import source [intent=reverse_etl availability=planned write=update_content_import_source]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Update a content import source: live Intercom mutation against /ai/content_import_sources/{source_id}; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    ai content list external pages - List external pages [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    ai content create an external page or update an external page by external id - Create an external page (or update an external page by external ID) [intent=reverse_etl availability=planned write=create_external_page]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Create an external page (or update an external page by external ID): live Intercom mutation against /ai/external_pages; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    ai content delete an external page - Delete an external page [intent=reverse_etl availability=planned write=delete_external_page]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Delete an external page: live Intercom mutation against /ai/external_pages/{page_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    ai content retrieve an external page - Retrieve an external page [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    ai content update an external page - Update an external page [intent=reverse_etl availability=planned write=update_external_page]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Update an external page: live Intercom mutation against /ai/external_pages/{page_id}; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    articles list all articles - List all articles [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    articles create an article - Create an article [intent=reverse_etl availability=planned write=create_article]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Create an article: live Intercom mutation against /articles; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    articles search for articles - Search for articles [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    articles delete an article - Delete an article [intent=reverse_etl availability=planned write=delete_article]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Delete an article: live Intercom mutation against /articles/{article_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    articles retrieve an article - Retrieve an article [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    articles update an article - Update an article [intent=reverse_etl availability=planned write=update_article]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Update an article: live Intercom mutation against /articles/{article_id}; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    articles add a tag to an article - Add a tag to an article [intent=reverse_etl availability=planned write=attach_tag_to_article]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Add a tag to an article: live Intercom mutation against /articles/{article_id}/tags; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    articles remove a tag from an article - Remove a tag from an article [intent=reverse_etl availability=planned write=detach_tag_from_article]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Remove a tag from an article: live Intercom mutation against /articles/{article_id}/tags/{id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    articles list article versions - List article versions [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    articles retrieve an article version - Retrieve an article version [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    articles retrieve an article draft - Retrieve an article draft [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    articles stage an article draft - Stage an article draft [intent=reverse_etl availability=planned write=stage_article_draft]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Stage an article draft: live Intercom mutation against /articles/{id}/draft; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    articles publish an article draft - Publish an article draft [intent=reverse_etl availability=planned write=publish_article_draft]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Publish an article draft: live Intercom mutation against /articles/{id}/draft/publish; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    audiences list all audiences - List all audiences [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    audiences create an audience - Create an audience [intent=reverse_etl availability=planned write=create_audience]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Create an audience: live Intercom mutation against /audiences; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    audiences delete an audience - Delete an audience [intent=reverse_etl availability=planned write=delete_audience]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Delete an audience: live Intercom mutation against /audiences/{id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    audiences retrieve an audience - Retrieve an audience [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    audiences update an audience - Update an audience [intent=reverse_etl availability=planned write=update_audience]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Update an audience: live Intercom mutation against /audiences/{id}; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    away status reasons list all away status reasons - List all away status reasons [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    brands list all brands - List all brands [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    brands retrieve a brand - Retrieve a brand [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    calls list all calls - List all calls [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    calls list calls with transcripts - List calls with transcripts [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    calls get a call - Get a call [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    calls get call recording by call id - Get call recording by call id [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    calls get call transcript by call id - Get call transcript by call id [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    companies list - List Intercom companies as ETL records. [intent=etl availability=implemented stream=companies]
    companies create or update a company - Create or Update a company [intent=reverse_etl availability=planned write=create_or_update_company]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Create or Update a company: live Intercom mutation against /companies; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    companies list all companies - List all companies [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    companies scroll over all companies - Scroll over all companies [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    companies delete a company - Delete a company [intent=reverse_etl availability=planned write=delete_company]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Delete a company: live Intercom mutation against /companies/{company_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    companies retrieve a company by id - Retrieve a company by ID [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    companies update a company - Update a company [intent=reverse_etl availability=planned write=update_company]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Update a company: live Intercom mutation against /companies/{company_id}; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    companies list attached contacts - List attached contacts [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    notes list all company notes - List all company notes [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    notes create a company note - Create a company note [intent=reverse_etl availability=planned write=create_company_note]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Create a company note: live Intercom mutation against /companies/{company_id}/notes; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    companies list attached segments for companies - List attached segments for companies [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    contacts list - List Intercom contacts as ETL records. [intent=etl availability=implemented stream=contacts]
    contacts create contact - Create contact [intent=reverse_etl availability=planned write=create_contact]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Create contact: live Intercom mutation against /contacts; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    contacts get a contact by external id - Get a contact by External ID [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    contacts merge a lead and a user - Merge a lead and a user [intent=reverse_etl availability=planned write=merge_contact]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Merge a lead and a user: live Intercom mutation against /contacts/merge; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    contacts search contacts - Search contacts [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    contacts delete a contact - Delete a contact [intent=reverse_etl availability=planned write=delete_contact]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Delete a contact: live Intercom mutation against /contacts/{contact_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    contacts get a contact - Get a contact [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    contacts update a contact - Update a contact [intent=reverse_etl availability=planned write=update_contact]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Update a contact: live Intercom mutation against /contacts/{contact_id}; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    contacts archive contact - Archive contact [intent=reverse_etl availability=planned write=archive_contact]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Archive contact: live Intercom mutation against /contacts/{contact_id}/archive; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    contacts block contact - Block contact [intent=reverse_etl availability=planned write=block_contact]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Block contact: live Intercom mutation against /contacts/{contact_id}/block; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    contacts list attached companies for contact - List attached companies for contact [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    companies attach a contact to a company - Attach a Contact to a Company [intent=reverse_etl availability=planned write=attach_contact_to_acompany]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Attach a Contact to a Company: live Intercom mutation against /contacts/{contact_id}/companies; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    companies detach a contact from a company - Detach a contact from a company [intent=reverse_etl availability=planned write=detach_contact_from_acompany]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Detach a contact from a company: live Intercom mutation against /contacts/{contact_id}/companies/{company_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    notes list all notes - List all notes [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    notes create a note - Create a note [intent=reverse_etl availability=planned write=create_note]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Create a note: live Intercom mutation against /contacts/{contact_id}/notes; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    contacts list attached segments for contact - List attached segments for contact [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    contacts list subscriptions for a contact - List subscriptions for a contact [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    subscription types add subscription to a contact - Add subscription to a contact [intent=reverse_etl availability=planned write=attach_subscription_type_to_contact]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Add subscription to a contact: live Intercom mutation against /contacts/{contact_id}/subscriptions; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    subscription types remove subscription from a contact - Remove subscription from a contact [intent=reverse_etl availability=planned write=detach_subscription_type_to_contact]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Remove subscription from a contact: live Intercom mutation against /contacts/{contact_id}/subscriptions/{subscription_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    contacts list tags attached to a contact - List tags attached to a contact [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    tags add tag to a contact - Add tag to a contact [intent=reverse_etl availability=planned write=attach_tag_to_contact]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Add tag to a contact: live Intercom mutation against /contacts/{contact_id}/tags; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    tags remove tag from a contact - Remove tag from a contact [intent=reverse_etl availability=planned write=detach_tag_from_contact]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Remove tag from a contact: live Intercom mutation against /contacts/{contact_id}/tags/{tag_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    contacts unarchive contact - Unarchive contact [intent=reverse_etl availability=planned write=unarchive_contact]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Unarchive contact: live Intercom mutation against /contacts/{contact_id}/unarchive; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    banners list banners for a contact - List banners for a contact [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    banners dismiss a banner for a contact - Dismiss a banner for a contact [intent=reverse_etl availability=planned write=dismiss_contact_banner]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Dismiss a banner for a contact: live Intercom mutation against /contacts/{id}/banners/{view_id}/dismiss; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    contacts get contact merge history - Get contact merge history [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    content run a bulk action on knowledge hub content - Run a bulk action on Knowledge Hub content [intent=reverse_etl availability=planned write=bulk_content_actions]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Run a bulk action on Knowledge Hub content: live Intercom mutation against /content/bulk_actions; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    content search knowledge base contents - Search knowledge base contents [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    content snippets list all content snippets - List all content snippets [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    content snippets create a content snippet - Create a content snippet [intent=reverse_etl availability=planned write=create_content_snippet]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Create a content snippet: live Intercom mutation against /content_snippets; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    content snippets add a tag to a content snippet - Add a tag to a content snippet [intent=reverse_etl availability=planned write=attach_tag_to_content_snippet]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Add a tag to a content snippet: live Intercom mutation against /content_snippets/{content_snippet_id}/tags; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    content snippets remove a tag from a content snippet - Remove a tag from a content snippet [intent=reverse_etl availability=planned write=detach_tag_from_content_snippet]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Remove a tag from a content snippet: live Intercom mutation against /content_snippets/{content_snippet_id}/tags/{id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    content snippets delete a content snippet - Delete a content snippet [intent=reverse_etl availability=planned write=delete_content_snippet]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Delete a content snippet: live Intercom mutation against /content_snippets/{id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    content snippets retrieve a content snippet - Retrieve a content snippet [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    content snippets update a content snippet - Update a content snippet [intent=reverse_etl availability=planned write=update_content_snippet]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Update a content snippet: live Intercom mutation against /content_snippets/{id}; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    conversations list - List Intercom conversations as ETL records. [intent=etl availability=implemented stream=conversations]
    conversations creates a conversation - Creates a conversation [intent=reverse_etl availability=planned write=create_conversation]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Creates a conversation: live Intercom mutation against /conversations; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    conversations attributes list all conversation attributes - List all conversation attributes [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    conversations attributes create a conversation attribute - Create a conversation attribute [intent=reverse_etl availability=planned write=create_conversation_attribute]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Create a conversation attribute: live Intercom mutation against /conversations/attributes; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    conversations attributes delete archive a conversation attribute - Delete (archive) a conversation attribute [intent=reverse_etl availability=planned write=delete_conversation_attribute]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Delete (archive) a conversation attribute: live Intercom mutation against /conversations/attributes/{id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    conversations attributes get a conversation attribute - Get a conversation attribute [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    conversations attributes update a conversation attribute - Update a conversation attribute [intent=reverse_etl availability=planned write=update_conversation_attribute]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Update a conversation attribute: live Intercom mutation against /conversations/attributes/{id}; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    conversations attributes add an option to a list conversation attribute - Add an option to a list conversation attribute [intent=reverse_etl availability=planned write=create_conversation_attribute_option]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Add an option to a list conversation attribute: live Intercom mutation against /conversations/attributes/{id}/options; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    conversations attributes archive an option on a list conversation attribute - Archive an option on a list conversation attribute [intent=reverse_etl availability=planned write=delete_conversation_attribute_option]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Archive an option on a list conversation attribute: live Intercom mutation against /conversations/attributes/{id}/options/{option_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    conversations attributes update an option on a list conversation attribute - Update an option on a list conversation attribute [intent=reverse_etl availability=planned write=update_conversation_attribute_option]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Update an option on a list conversation attribute: live Intercom mutation against /conversations/attributes/{id}/options/{option_id}; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    conversations list all deleted conversation ids - List all deleted conversation IDs [intent=direct_read availability=planned]; approval: Blocked pending CDC truthfulness/state foundations #2986/#2988.; risk: Reads operational/audit/changefeed-like provider data.; notes: Official Intercom operation ledger row; lane=cdc_changefeed. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    conversations redact a conversation part - Redact a conversation part [intent=reverse_etl availability=planned write=redact_conversation]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Redact a conversation part: live Intercom mutation against /conversations/redact; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    conversations search conversations - Search conversations [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    conversations delete a conversation - Delete a conversation [intent=reverse_etl availability=planned write=delete_conversation]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Delete a conversation: live Intercom mutation against /conversations/{conversation_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    conversations retrieve a conversation - Retrieve a conversation [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    conversations update a conversation - Update a conversation [intent=reverse_etl availability=planned write=update_conversation]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Update a conversation: live Intercom mutation against /conversations/{conversation_id}; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    conversations convert a conversation to a ticket - Convert a conversation to a ticket [intent=reverse_etl availability=planned write=convert_conversation_to_ticket]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Convert a conversation to a ticket: live Intercom mutation against /conversations/{conversation_id}/convert; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    conversations attach a contact to a conversation - Attach a contact to a conversation [intent=reverse_etl availability=planned write=attach_contact_to_conversation]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Attach a contact to a conversation: live Intercom mutation against /conversations/{conversation_id}/customers; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    conversations detach a contact from a group conversation - Detach a contact from a group conversation [intent=reverse_etl availability=planned write=detach_contact_from_conversation]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Detach a contact from a group conversation: live Intercom mutation against /conversations/{conversation_id}/customers/{contact_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    conversations manage a conversation - Manage a conversation [intent=reverse_etl availability=planned write=manage_conversation]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Manage a conversation: live Intercom mutation against /conversations/{conversation_id}/parts; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    conversations reply to a conversation - Reply to a conversation [intent=reverse_etl availability=planned write=reply_conversation]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Reply to a conversation: live Intercom mutation against /conversations/{conversation_id}/reply; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    tags add tag to a conversation - Add tag to a conversation [intent=reverse_etl availability=planned write=attach_tag_to_conversation]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Add tag to a conversation: live Intercom mutation against /conversations/{conversation_id}/tags; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    tags remove tag from a conversation - Remove tag from a conversation [intent=reverse_etl availability=planned write=detach_tag_from_conversation]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Remove tag from a conversation: live Intercom mutation against /conversations/{conversation_id}/tags/{tag_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    conversations list handling events - List handling events [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    conversations merge a conversation - Merge a conversation [intent=reverse_etl availability=planned write=merge_conversation]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Merge a conversation: live Intercom mutation against /conversations/{id}/merge; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    conversations list side conversations - List side conversations [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    custom object instances delete a custom object instance by external id - Delete a Custom Object Instance by External ID [intent=reverse_etl availability=planned write=delete_custom_object_instances_by_id]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Delete a Custom Object Instance by External ID: live Intercom mutation against /custom_object_instances/{custom_object_type_identifier}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    custom object instances list custom object instances - List Custom Object Instances [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    custom object instances create or update a custom object instance - Create or Update a Custom Object Instance [intent=reverse_etl availability=planned write=create_custom_object_instances]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Create or Update a Custom Object Instance: live Intercom mutation against /custom_object_instances/{custom_object_type_identifier}; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    custom object instances delete a custom object instance by id - Delete a Custom Object Instance by ID [intent=reverse_etl availability=planned write=delete_custom_object_instances_by_external_id]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Delete a Custom Object Instance by ID: live Intercom mutation against /custom_object_instances/{custom_object_type_identifier}/{custom_object_instance_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    custom object instances get custom object instance by id - Get Custom Object Instance by ID [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    data attributes list all data attributes - List all data attributes [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    data attributes create a data attribute - Create a data attribute [intent=reverse_etl availability=planned write=create_data_attribute]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Create a data attribute: live Intercom mutation against /data_attributes; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    data attributes update a data attribute - Update a data attribute [intent=reverse_etl availability=planned write=update_data_attribute]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Update a data attribute: live Intercom mutation against /data_attributes/{data_attribute_id}; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    data connectors list all data connectors - List all data connectors [intent=direct_read availability=planned]; approval: Blocked pending CDC truthfulness/state foundations #2986/#2988.; risk: Reads operational/audit/changefeed-like provider data.; notes: Official Intercom operation ledger row; lane=cdc_changefeed. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    data connectors create a data connector - Create a data connector [intent=reverse_etl availability=planned write=create_data_connector]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Create a data connector: live Intercom mutation against /data_connectors; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    data connectors list execution results for a data connector - List execution results for a data connector [intent=direct_read availability=planned]; approval: Blocked pending CDC truthfulness/state foundations #2986/#2988.; risk: Reads operational/audit/changefeed-like provider data.; notes: Official Intercom operation ledger row; lane=cdc_changefeed. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    data connectors retrieve an execution result - Retrieve an execution result [intent=direct_read availability=planned]; approval: Blocked pending CDC truthfulness/state foundations #2986/#2988.; risk: Reads operational/audit/changefeed-like provider data.; notes: Official Intercom operation ledger row; lane=cdc_changefeed. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    data connectors delete a data connector - Delete a data connector [intent=reverse_etl availability=planned write=delete_data_connector]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Delete a data connector: live Intercom mutation against /data_connectors/{id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    data connectors retrieve a data connector - Retrieve a data connector [intent=direct_read availability=planned]; approval: Blocked pending CDC truthfulness/state foundations #2986/#2988.; risk: Reads operational/audit/changefeed-like provider data.; notes: Official Intercom operation ledger row; lane=cdc_changefeed. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    data connectors update a data connector - Update a data connector [intent=reverse_etl availability=planned write=update_data_connector]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Update a data connector: live Intercom mutation against /data_connectors/{id}; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    data export download content data export - Download content data export [intent=direct_read availability=planned]; approval: No live binary execution in this connector-local slice.; risk: Reads or stages provider export/download data; output must stay bounded and redacted.; notes: Official Intercom operation ledger row; lane=binary_file. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    reporting data export download completed export job data - Download completed export job data [intent=direct_read availability=planned]; approval: No live binary execution in this connector-local slice.; risk: Reads or stages provider export/download data; output must stay bounded and redacted.; notes: Official Intercom operation ledger row; lane=binary_file. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    emails list all email settings - List all email settings [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    emails retrieve an email setting - Retrieve an email setting [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    data events list all data events - List all data events [intent=direct_read availability=planned]; approval: Blocked pending CDC truthfulness/state foundations #2986/#2988.; risk: Reads operational/audit/changefeed-like provider data.; notes: Official Intercom operation ledger row; lane=cdc_changefeed. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    data events submit a data event - Submit a data event [intent=reverse_etl availability=planned write=create_data_event]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Submit a data event: live Intercom mutation against /events; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    data events create event summaries - Create event summaries [intent=direct_read availability=planned]; approval: Blocked pending CDC truthfulness/state foundations #2986/#2988.; risk: Reads operational/audit/changefeed-like provider data.; notes: Official Intercom operation ledger row; lane=cdc_changefeed. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    data export cancel content data export - Cancel content data export [intent=reverse_etl availability=planned write=cancel_data_export]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Cancel content data export: live Intercom mutation against /export/cancel/{job_identifier}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    data export create content data export - Create content data export [intent=direct_read availability=planned]; approval: No live binary execution in this connector-local slice.; risk: Reads or stages provider export/download data; output must stay bounded and redacted.; notes: Official Intercom operation ledger row; lane=binary_file. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    data export show content data export - Show content data export [intent=direct_read availability=planned]; approval: No live binary execution in this connector-local slice.; risk: Reads or stages provider export/download data; output must stay bounded and redacted.; notes: Official Intercom operation ledger row; lane=binary_file. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    reporting data export enqueue a new reporting data export job - Enqueue a new reporting data export job [intent=direct_read availability=planned]; approval: No live binary execution in this connector-local slice.; risk: Reads or stages provider export/download data; output must stay bounded and redacted.; notes: Official Intercom operation ledger row; lane=binary_file. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    reporting data export list available datasets and attributes - List available datasets and attributes [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    reporting data export get export job status - Get export job status [intent=direct_read availability=planned]; approval: No live binary execution in this connector-local slice.; risk: Reads or stages provider export/download data; output must stay bounded and redacted.; notes: Official Intercom operation ledger row; lane=binary_file. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    workflows export a workflow - Export a workflow [intent=direct_read availability=planned]; approval: No live binary execution in this connector-local slice.; risk: Reads or stages provider export/download data; output must stay bounded and redacted.; notes: Official Intercom operation ledger row; lane=binary_file. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    fin agent submit a CSAT rating - Submit a CSAT rating [intent=reverse_etl availability=planned write=submit_fin_csat]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Submit a CSAT rating: live Intercom mutation against /fin/csat; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    fin agent reply to fin - Reply to Fin [intent=reverse_etl availability=planned write=reply_to_fin]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Reply to Fin: live Intercom mutation against /fin/reply; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    fin agent start a conversation with fin - Start a conversation with Fin [intent=reverse_etl availability=planned write=start_fin_conversation]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Start a conversation with Fin: live Intercom mutation against /fin/start; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    calls collect fin voice call by id - Collect Fin Voice call by ID [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    calls collect fin voice calls by conversation id - Collect Fin Voice calls by conversation ID [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    calls collect fin voice call by external id - Collect Fin Voice call by external ID [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    calls collect fin voice call by phone number - Collect Fin Voice call by phone number [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    calls register a fin voice call - Register a Fin Voice call [intent=reverse_etl availability=planned write=register_fin_voice_call]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Register a Fin Voice call: live Intercom mutation against /fin_voice/register; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    help center list all collections - List all collections [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    help center create a collection - Create a collection [intent=reverse_etl availability=planned write=create_collection]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Create a collection: live Intercom mutation against /help_center/collections; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    help center delete a collection - Delete a collection [intent=reverse_etl availability=planned write=delete_collection]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Delete a collection: live Intercom mutation against /help_center/collections/{collection_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    help center retrieve a collection - Retrieve a collection [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    help center update a collection - Update a collection [intent=reverse_etl availability=planned write=update_collection]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Update a collection: live Intercom mutation against /help_center/collections/{collection_id}; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    help center list all help centers - List all Help Centers [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    help center retrieve a help center - Retrieve a Help Center [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    help center list all redirects for a help center - List all redirects for a help center [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    help center create a redirect - Create a redirect [intent=reverse_etl availability=planned write=create_help_center_redirect]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Create a redirect: live Intercom mutation against /help_center/help_centers/{help_center_id}/redirects; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    help center delete a redirect - Delete a redirect [intent=reverse_etl availability=planned write=delete_help_center_redirect]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Delete a redirect: live Intercom mutation against /help_center/help_centers/{help_center_id}/redirects/{id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    help center retrieve a redirect - Retrieve a redirect [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    internal articles list all articles - List all articles [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    internal articles create an internal article - Create an internal article [intent=reverse_etl availability=planned write=create_internal_article]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Create an internal article: live Intercom mutation against /internal_articles; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    internal articles search for internal articles - Search for internal articles [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    internal articles delete an internal article - Delete an internal article [intent=reverse_etl availability=planned write=delete_internal_article]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Delete an internal article: live Intercom mutation against /internal_articles/{internal_article_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    internal articles retrieve an internal article - Retrieve an internal article [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    internal articles update an internal article - Update an internal article [intent=reverse_etl availability=planned write=update_internal_article]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Update an internal article: live Intercom mutation against /internal_articles/{internal_article_id}; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    internal articles add a tag to an internal article - Add a tag to an internal article [intent=reverse_etl availability=planned write=attach_tag_to_internal_article]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Add a tag to an internal article: live Intercom mutation against /internal_articles/{internal_article_id}/tags; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    internal articles remove a tag from an internal article - Remove a tag from an internal article [intent=reverse_etl availability=planned write=detach_tag_from_internal_article]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Remove a tag from an internal article: live Intercom mutation against /internal_articles/{internal_article_id}/tags/{id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    ip allowlist get ip allowlist settings - Get IP allowlist settings [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    ip allowlist update ip allowlist settings - Update IP allowlist settings [intent=reverse_etl availability=planned write=update_ip_allowlist]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Update IP allowlist settings: live Intercom mutation against /ip_allowlist; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    jobs retrieve job status - Retrieve job status [intent=direct_read availability=planned]; approval: Blocked pending CDC truthfulness/state foundations #2986/#2988.; risk: Reads operational/audit/changefeed-like provider data.; notes: Official Intercom operation ledger row; lane=cdc_changefeed. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    macros list all macros - List all macros [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    macros retrieve a macro - Retrieve a macro [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    admins identify an admin - Identify an admin [intent=direct_read availability=excluded]; notes: Official Intercom operation ledger row; lane=excluded_not_applicable. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    messages create a message - Create a message [intent=reverse_etl availability=planned write=create_message]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Create a message: live Intercom mutation against /messages; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    messages get statuses of all messages sent based on the specified ruleset id - Get statuses of all messages sent based on the specified ruleset_id [intent=direct_read availability=planned]; approval: Blocked pending CDC truthfulness/state foundations #2986/#2988.; risk: Reads operational/audit/changefeed-like provider data.; notes: Official Intercom operation ledger row; lane=cdc_changefeed. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    messages retrieve whats app message delivery status - Retrieve WhatsApp message delivery status [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    news list all news items - List all news items [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    news create a news item - Create a news item [intent=reverse_etl availability=planned write=create_news_item]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Create a news item: live Intercom mutation against /news/news_items; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    news delete a news item - Delete a news item [intent=reverse_etl availability=planned write=delete_news_item]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Delete a news item: live Intercom mutation against /news/news_items/{news_item_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    news retrieve a news item - Retrieve a news item [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    news update a news item - Update a news item [intent=reverse_etl availability=planned write=update_news_item]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Update a news item: live Intercom mutation against /news/news_items/{news_item_id}; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    news list all newsfeeds - List all newsfeeds [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    news retrieve a newsfeed - Retrieve a newsfeed [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    news list all live newsfeed items - List all live newsfeed items [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    notes retrieve a note - Retrieve a note [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    office hours list all office hours schedules - List all office hours schedules [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    office hours create an office hours schedule - Create an office hours schedule [intent=reverse_etl availability=planned write=create_office_hours_schedule]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Create an office hours schedule: live Intercom mutation against /office_hours_schedules; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    office hours delete an office hours schedule - Delete an office hours schedule [intent=reverse_etl availability=planned write=delete_office_hours_schedule]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Delete an office hours schedule: live Intercom mutation against /office_hours_schedules/{id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    office hours retrieve an office hours schedule - Retrieve an office hours schedule [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    office hours update an office hours schedule - Update an office hours schedule [intent=reverse_etl availability=planned write=update_office_hours_schedule]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Update an office hours schedule: live Intercom mutation against /office_hours_schedules/{id}; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    office hours list all office hours exceptions - List all office hours exceptions [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    office hours create an office hours exception - Create an office hours exception [intent=reverse_etl availability=planned write=create_office_hours_exception]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Create an office hours exception: live Intercom mutation against /office_hours_schedules/{office_hours_schedule_id}/office_hours_exceptions; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    office hours delete an office hours exception - Delete an office hours exception [intent=reverse_etl availability=planned write=delete_office_hours_exception]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Delete an office hours exception: live Intercom mutation against /office_hours_schedules/{office_hours_schedule_id}/office_hours_exceptions/{id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    office hours retrieve an office hours exception - Retrieve an office hours exception [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    office hours update an office hours exception - Update an office hours exception [intent=reverse_etl availability=planned write=update_office_hours_exception]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Update an office hours exception: live Intercom mutation against /office_hours_schedules/{office_hours_schedule_id}/office_hours_exceptions/{id}; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    switch create a phone switch - Create a phone Switch [intent=reverse_etl availability=planned write=create_phone_switch]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Create a phone Switch: live Intercom mutation against /phone_call_redirects; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    segments list all segments - List all segments [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    segments retrieve a segment - Retrieve a segment [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    subscription types list subscription types - List subscription types [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    tags list - List Intercom tags as ETL records. [intent=etl availability=implemented stream=tags]
    tags create or update a tag tag or untag companies tag contacts - Create or update a tag, Tag or untag companies, Tag contacts [intent=reverse_etl availability=planned write=create_tag]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Create or update a tag, Tag or untag companies, Tag contacts: live Intercom mutation against /tags; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    tags delete tag - Delete tag [intent=reverse_etl availability=planned write=delete_tag]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Delete tag: live Intercom mutation against /tags/{tag_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    tags find a specific tag - Find a specific tag [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    teams list all teams - List all teams [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    teams retrieve a team - Retrieve a team [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    teams retrieve team metrics - Retrieve team metrics [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    ticket states list all ticket states - List all ticket states [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    ticket types list all ticket types - List all ticket types [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    ticket types create a ticket type - Create a ticket type [intent=reverse_etl availability=planned write=create_ticket_type]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Create a ticket type: live Intercom mutation against /ticket_types; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    ticket types retrieve a ticket type - Retrieve a ticket type [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    ticket types update a ticket type - Update a ticket type [intent=reverse_etl availability=planned write=update_ticket_type]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Update a ticket type: live Intercom mutation against /ticket_types/{ticket_type_id}; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    ticket type attributes create a new attribute for a ticket type - Create a new attribute for a ticket type [intent=reverse_etl availability=planned write=create_ticket_type_attribute]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Create a new attribute for a ticket type: live Intercom mutation against /ticket_types/{ticket_type_id}/attributes; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    ticket type attributes update an existing attribute for a ticket type - Update an existing attribute for a ticket type [intent=reverse_etl availability=planned write=update_ticket_type_attribute]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Update an existing attribute for a ticket type: live Intercom mutation against /ticket_types/{ticket_type_id}/attributes/{attribute_id}; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    tickets create a ticket - Create a ticket [intent=reverse_etl availability=planned write=create_ticket]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Create a ticket: live Intercom mutation against /tickets; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    tickets enqueue create ticket - Enqueue create ticket [intent=reverse_etl availability=planned write=enqueue_create_ticket]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Enqueue create ticket: live Intercom mutation against /tickets/enqueue; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    tickets search tickets - Search tickets [intent=docs_only availability=planned]; notes: Official Intercom operation ledger row; lane=etl_read. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    tickets delete a ticket - Delete a ticket [intent=reverse_etl availability=planned write=delete_ticket]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Delete a ticket: live Intercom mutation against /tickets/{ticket_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    tickets retrieve a ticket - Retrieve a ticket [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    tickets update a ticket - Update a ticket [intent=reverse_etl availability=planned write=update_ticket]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Update a ticket: live Intercom mutation against /tickets/{ticket_id}; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    tickets change ticket type - Change ticket type [intent=reverse_etl availability=planned write=change_ticket_type]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Change ticket type: live Intercom mutation against /tickets/{ticket_id}/change_type; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    tickets link a conversation to a ticket - Link a conversation to a ticket [intent=reverse_etl availability=planned write=link_conversation_to_ticket]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Link a conversation to a ticket: live Intercom mutation against /tickets/{ticket_id}/linked_conversations; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    tickets unlink a conversation from a ticket - Unlink a conversation from a ticket [intent=reverse_etl availability=planned write=unlink_conversation_from_ticket]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Unlink a conversation from a ticket: live Intercom mutation against /tickets/{ticket_id}/linked_conversations/{id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    tickets reply to a ticket - Reply to a ticket [intent=reverse_etl availability=planned write=reply_ticket]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Reply to a ticket: live Intercom mutation against /tickets/{ticket_id}/reply; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    tags add tag to a ticket - Add tag to a ticket [intent=reverse_etl availability=planned write=attach_tag_to_ticket]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Add tag to a ticket: live Intercom mutation against /tickets/{ticket_id}/tags; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    tags remove tag from a ticket - Remove tag from a ticket [intent=reverse_etl availability=planned write=detach_tag_from_ticket]; approval: reverse ETL writes require plan, preview, explicit approval, execute and typed destructive confirmation.; risk: Remove tag from a ticket: live Intercom mutation against /tickets/{ticket_id}/tags/{tag_id}; reverse ETL requires plan, preview, explicit approval, execute and typed destructive confirmation; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    visitors retrieve a visitor with user id - Retrieve a visitor with User ID [intent=direct_read availability=planned]; notes: Official Intercom operation ledger row; lane=direct_read_query_search. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    visitors update a visitor - Update a visitor [intent=reverse_etl availability=planned write=update_visitor]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Update a visitor: live Intercom mutation against /visitors; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
    visitors convert a visitor - Convert a visitor [intent=reverse_etl availability=planned write=convert_visitor]; approval: reverse ETL writes require plan, preview, explicit approval, execute.; risk: Convert a visitor: live Intercom mutation against /visitors/convert; reverse ETL requires plan, preview, explicit approval, execute; notes: Official Intercom operation ledger row; lane=reverse_etl_write. This command metadata is definition-owned and does not expose a generic raw API escape hatch.
  Help topics:
    intercom-safety - Intercom writes require plan, preview, explicit approval, execute; destructive operations require typed destructive confirmation.
    intercom-coverage - Official Intercom OpenAPI 2.16 inventory is ledgered in api_surface.json with implemented vs blocked/planned coverage.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect intercom

  # Inspect as structured JSON
  pm connectors inspect intercom --json

AGENT WORKFLOW
  - Run pm connectors inspect intercom before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
