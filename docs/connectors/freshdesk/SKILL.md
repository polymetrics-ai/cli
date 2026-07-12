---
name: pm-freshdesk
description: Freshdesk connector knowledge and safe action guide.
---

# pm-freshdesk

## Purpose

Reads Freshdesk support resources and exposes fixed Freshdesk REST API v2 operations through streams, bounded JSON direct reads, and gated reverse-ETL write actions.

## Icon

- asset: icons/freshdesk.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://developers.freshdesk.com/api/#change_log

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- max_pages
- mode
- page_size
- start_date
- api_key (secret)

## ETL Streams

- tickets:
  - primary key: id
  - cursor: updated_at
  - fields: company_id(), created_at(), due_by(), group_id(), id(), priority(), requester_id(), responder_id(), source(), spam(), status(), subject(), type(), updated_at()
- contacts:
  - primary key: id
  - cursor: updated_at
  - fields: active(), company_id(), created_at(), email(), id(), mobile(), name(), phone(), updated_at()
- companies:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), description(), id(), name(), note(), updated_at()
- agents:
  - primary key: id
  - cursor: updated_at
  - fields: available(), created_at(), id(), occasional(), ticket_scope(), updated_at()
- groups:
  - primary key: id
  - cursor: updated_at
  - fields: created_at(), description(), id(), name(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- delete_admin_groups_id:
  - endpoint: DELETE /admin/groups/{{ record.id }}
  - required fields: id
  - risk: Freshdesk Delete a Group (DELETE /admin/groups/{id}); requires reverse ETL plan, preview, approval, and execute.
- delete_admin_skills_id:
  - endpoint: DELETE /admin/skills/{{ record.id }}
  - required fields: id
  - risk: Freshdesk Delete a Skill (DELETE /admin/skills/{id}); requires reverse ETL plan, preview, approval, and execute.
- delete_admin_ticket_fields_id:
  - endpoint: DELETE /admin/ticket_fields/{{ record.id }}
  - required fields: id
  - risk: Freshdesk Delete a Ticket Field (DELETE /admin/ticket_fields/{id}); requires reverse ETL plan, preview, approval, and execute.
- delete_admin_ticket_fields_id_sections_section_id:
  - endpoint: DELETE /admin/ticket_fields/{{ record.id }}/sections/{{ record.section_id }}
  - required fields: id, section_id
  - risk: Freshdesk Delete a section (DELETE /admin/ticket_fields/{id}/sections/{section_id}); requires reverse ETL plan, preview, approval, and execute.
- delete_agents_id:
  - endpoint: DELETE /agents/{{ record.id }}
  - required fields: id
  - risk: Freshdesk Delete an Agent (DELETE /agents/{id}); requires reverse ETL plan, preview, approval, and execute.
- delete_attachments_id:
  - endpoint: DELETE /attachments/{{ record.id }}
  - required fields: id
  - risk: Freshdesk Delete an attachment (DELETE /attachments/{id}); requires reverse ETL plan, preview, approval, and execute.
- delete_companies_id:
  - endpoint: DELETE /companies/{{ record.id }}
  - required fields: id
  - risk: Freshdesk Delete a Company (DELETE /companies/{id}); requires reverse ETL plan, preview, approval, and execute.
- delete_company_field_id:
  - endpoint: DELETE /company_field/{{ record.id }}
  - required fields: id
  - risk: Freshdesk Delete a company field (DELETE /company_field/{id}); requires reverse ETL plan, preview, approval, and execute.
- delete_contact_field_id:
  - endpoint: DELETE /contact_field/{{ record.id }}
  - required fields: id
  - risk: Freshdesk Delete a contact field (DELETE /contact_field/{id}); requires reverse ETL plan, preview, approval, and execute.
- delete_contacts_id:
  - endpoint: DELETE /contacts/{{ record.id }}
  - required fields: id
  - risk: Freshdesk Soft Delete a Contact (DELETE /contacts/{id}); requires reverse ETL plan, preview, approval, and execute.
- delete_contacts_id_hard_delete:
  - endpoint: DELETE /contacts/{{ record.id }}/hard_delete
  - required fields: id
  - risk: Freshdesk Permanently Delete a Contact (DELETE /contacts/{id}/hard_delete); requires reverse ETL plan, preview, approval, and execute.
- delete_conversations_id:
  - endpoint: DELETE /conversations/{{ record.id }}
  - required fields: id
  - risk: Freshdesk Delete a conversation (DELETE /conversations/{id}); requires reverse ETL plan, preview, approval, and execute.
- delete_custom_objects_schemas_schema_id_records_record_id:
  - endpoint: DELETE /custom_objects/schemas/{{ record.schema_id }}/records/{{ record.record_id }}
  - required fields: schema_id, record_id
  - risk: Freshdesk Delete A Record (DELETE /custom_objects/schemas/{schema_id}/records/{record_id}); requires reverse ETL plan, preview, approval, and execute.
- delete_discussions_categories_id:
  - endpoint: DELETE /discussions/categories/{{ record.id }}
  - required fields: id
  - risk: Freshdesk Get details of Canned Responses in a Folder (DELETE /discussions/categories/{id}); requires reverse ETL plan, preview, approval, and execute.
- delete_discussions_comments_id:
  - endpoint: DELETE /discussions/comments/{{ record.id }}
  - required fields: id
  - risk: Freshdesk Get details of Canned Responses in a Folder (DELETE /discussions/comments/{id}); requires reverse ETL plan, preview, approval, and execute.
- delete_discussions_forums_forum_id_follow:
  - endpoint: DELETE /discussions/forums/{{ record.forum_id }}/follow?user_id={{ record.user_id }}
  - required fields: forum_id, user_id
  - risk: Freshdesk Get details of Canned Responses in a Folder (DELETE /discussions/forums/{forum_id}/follow?user_id={user_id}); requires reverse ETL plan, preview, approval, and execute.
- delete_discussions_forums_id:
  - endpoint: DELETE /discussions/forums/{{ record.id }}
  - required fields: id
  - risk: Freshdesk Get details of Canned Responses in a Folder (DELETE /discussions/forums/{id}); requires reverse ETL plan, preview, approval, and execute.
- delete_discussions_topics_id:
  - endpoint: DELETE /discussions/topics/{{ record.id }}
  - required fields: id
  - risk: Freshdesk Get details of Canned Responses in a Folder (DELETE /discussions/topics/{id}); requires reverse ETL plan, preview, approval, and execute.
- delete_discussions_topics_topic_id_follow:
  - endpoint: DELETE /discussions/topics/{{ record.topic_id }}/follow?user_id={{ record.user_id }}
  - required fields: topic_id, user_id
  - risk: Freshdesk Get details of Canned Responses in a Folder (DELETE /discussions/topics/{topic_id}/follow?user_id={user_id}); requires reverse ETL plan, preview, approval, and execute.
- delete_email_mailboxes_id:
  - endpoint: DELETE /email/mailboxes/{{ record.id }}
  - required fields: id
  - risk: Freshdesk Delete an Email Mailbox (DELETE /email/mailboxes/{id}); requires reverse ETL plan, preview, approval, and execute.
- delete_groups_id:
  - endpoint: DELETE /groups/{{ record.id }}
  - required fields: id
  - risk: Freshdesk Delete a Group (DELETE /groups/{id}); requires reverse ETL plan, preview, approval, and execute.
- delete_solutions_articles_id:
  - endpoint: DELETE /solutions/articles/{{ record.id }}
  - required fields: id
  - risk: Freshdesk Get details of Canned Responses in a Folder (DELETE /solutions/articles/{id}); requires reverse ETL plan, preview, approval, and execute.
- delete_solutions_categories_id:
  - endpoint: DELETE /solutions/categories/{{ record.id }}
  - required fields: id
  - risk: Freshdesk Get details of Canned Responses in a Folder (DELETE /solutions/categories/{id}); requires reverse ETL plan, preview, approval, and execute.
- delete_solutions_folders_id:
  - endpoint: DELETE /solutions/folders/{{ record.id }}
  - required fields: id
  - risk: Freshdesk Get details of Canned Responses in a Folder (DELETE /solutions/folders/{id}); requires reverse ETL plan, preview, approval, and execute.
- delete_ticket_forms_form_id_fields_field_id:
  - endpoint: DELETE /ticket-forms/{{ record.form_id }}/fields/{{ record.field_id }}
  - required fields: form_id, field_id
  - risk: Freshdesk Delete a Ticket Form's Field (DELETE /ticket-forms/{form_id}/fields/{field_id}); requires reverse ETL plan, preview, approval, and execute.
- delete_ticket_forms_id:
  - endpoint: DELETE /ticket-forms/{{ record.id }}
  - required fields: id
  - risk: Freshdesk Delete a Ticket Form (DELETE /ticket-forms/{id}); requires reverse ETL plan, preview, approval, and execute.
- delete_tickets_archived_id:
  - endpoint: DELETE /tickets/archived/{{ record.id }}
  - required fields: id
  - risk: Freshdesk List All Satisfaction Ratings of a Ticket (DELETE /tickets/archived/{id}); requires reverse ETL plan, preview, approval, and execute.
- delete_tickets_id:
  - endpoint: DELETE /tickets/{{ record.id }}
  - required fields: id
  - risk: Freshdesk Delete a Ticket (DELETE /tickets/{id}); requires reverse ETL plan, preview, approval, and execute.
- delete_tickets_id_accesses:
  - endpoint: DELETE /tickets/{{ record.id }}/accesses
  - required fields: id
  - risk: Freshdesk List All Satisfaction Ratings of a Ticket (DELETE /tickets/{id}/accesses); requires reverse ETL plan, preview, approval, and execute.
- delete_tickets_id_summary:
  - endpoint: DELETE /tickets/{{ record.id }}/summary
  - required fields: id
  - risk: Freshdesk List All Satisfaction Ratings of a Ticket (DELETE /tickets/{id}/summary); requires reverse ETL plan, preview, approval, and execute.
- delete_time_entries_id:
  - endpoint: DELETE /time_entries/{{ record.id }}
  - required fields: id
  - risk: Freshdesk Delete a Time Entry (DELETE /time_entries/{id}); requires reverse ETL plan, preview, approval, and execute.
- post_companies:
  - endpoint: POST /companies
  - optional fields: custom_fields, description, domains, name, note, health_score, account_tier, renewal_date, industry, lookup_parameter
  - risk: Freshdesk Create a Company (POST /companies); requires reverse ETL plan, preview, approval, and execute.
- post_contacts:
  - endpoint: POST /contacts
  - optional fields: name, email, phone, mobile, twitter_id, social_handler, unique_external_id, other_emails, company_id, view_all_tickets, other_companies, address, avatar, custom_fields, description, job_title, language, tags, time_zone, lookup_parameter
  - risk: Freshdesk Create a Contact (POST /contacts); requires reverse ETL plan, preview, approval, and execute.
- post_contacts_export:
  - endpoint: POST /contacts/export
  - optional fields: fields
  - risk: Freshdesk Contact Export (POST /contacts/export); requires reverse ETL plan, preview, approval, and execute.
- post_contacts_merge:
  - endpoint: POST /contacts/merge
  - optional fields: primary_contact_id, secondary_contact_ids, contact
  - risk: Freshdesk Merge Contacts (POST /contacts/merge); requires reverse ETL plan, preview, approval, and execute.
- post_tickets:
  - endpoint: POST /tickets
  - optional fields: name, requester_id, email, facebook_id, phone, twitter_id, unique_external_id, subject, type, status, priority, description, responder_id, attachments, cc_emails, custom_fields, due_by, email_config_id, fr_due_by, group_id, parent_id, product_id, source, tags, company_id, internal_agent_id, internal_group_id, lookup_parameter
  - risk: Freshdesk Create a Ticket (POST /tickets); requires reverse ETL plan, preview, approval, and execute.
- post_tickets_bulk_delete:
  - endpoint: POST /tickets/bulk_delete
  - optional fields: bulk_action
  - risk: Freshdesk Delete Multiple Tickets (POST /tickets/bulk_delete); requires reverse ETL plan, preview, approval, and execute.
- post_tickets_bulk_update:
  - endpoint: POST /tickets/bulk_update
  - optional fields: bulk_action
  - risk: Freshdesk Update Multiple Tickets (POST /tickets/bulk_update); requires reverse ETL plan, preview, approval, and execute.
- post_tickets_outbound_email:
  - endpoint: POST /tickets/outbound_email
  - optional fields: name, email, subject, type, status, priority, description, attachments, custom_fields, due_by, email_config_id, fr_due_by, group_id, tags
  - risk: Freshdesk Create an Outbound Email (POST /tickets/outbound_email); requires reverse ETL plan, preview, approval, and execute.
- post_tickets_id_forward:
  - endpoint: POST /tickets/{{ record.id }}/forward
  - required fields: id
  - optional fields: body, to_emails
  - risk: Freshdesk Forward a ticket (POST /tickets/{id}/forward); requires reverse ETL plan, preview, approval, and execute.
- put_agents_id:
  - endpoint: PUT /agents/{{ record.id }}
  - required fields: id
  - optional fields: occasional, signature, ticket_scope, skill_ids, group_ids, contribution_group_ids, role_ids, email, language, time_zone, focus_mode
  - risk: Freshdesk Update an Agent (PUT /agents/{id}); requires reverse ETL plan, preview, approval, and execute.
- put_companies_id:
  - endpoint: PUT /companies/{{ record.id }}
  - required fields: id
  - optional fields: custom_fields, description, domains, name, note, health_score, account_tier, renewal_date, industry, lookup_parameter
  - risk: Freshdesk Update a Company (PUT /companies/{id}); requires reverse ETL plan, preview, approval, and execute.
- put_contacts_id:
  - endpoint: PUT /contacts/{{ record.id }}
  - required fields: id
  - optional fields: name, email, phone, mobile, twitter_id, social_handler, unique_external_id, other_emails, company_id, view_all_tickets, other_companies, address, avatar, custom_fields, description, job_title, language, tags, time_zone, lookup_parameter
  - risk: Freshdesk Update a Contact (PUT /contacts/{id}); requires reverse ETL plan, preview, approval, and execute.
- put_contacts_id_make_agent:
  - endpoint: PUT /contacts/{{ record.id }}/make_agent
  - required fields: id
  - optional fields: occasional, signature, ticket_scope, skill_ids, group_ids, contribution_group_ids, role_ids, type, focus_mode
  - risk: Freshdesk Make Agent (PUT /contacts/{id}/make_agent); requires reverse ETL plan, preview, approval, and execute.
- put_contacts_id_restore:
  - endpoint: PUT /contacts/{{ record.id }}/restore
  - required fields: id
  - risk: Freshdesk Restore a Contact (PUT /contacts/{id}/restore); requires reverse ETL plan, preview, approval, and execute.
- put_groups_id:
  - endpoint: PUT /groups/{{ record.id }}
  - required fields: id
  - optional fields: agent_ids, auto_ticket_assign, description, escalate_to, name, unassigned_for
  - risk: Freshdesk Update a Group (PUT /groups/{id}); requires reverse ETL plan, preview, approval, and execute.
- put_tickets_merge:
  - endpoint: PUT /tickets/merge
  - optional fields: primary_id, ticket_ids, convert_recepients_to_cc, note_in_primary
  - risk: Freshdesk Ticket Merge API (PUT /tickets/merge); requires reverse ETL plan, preview, approval, and execute.
- put_tickets_id:
  - endpoint: PUT /tickets/{{ record.id }}
  - required fields: id
  - optional fields: name, requester_id, email, facebook_id, phone, twitter_id, unique_external_id, subject, type, status, priority, description, responder_id, attachments, custom_fields, due_by, email_config_id, fr_due_by, group_id, parent_id, product_id, source, tags, company_id, internal_agent_id, internal_group_id, lookup_parameter
  - risk: Freshdesk Update a Ticket (PUT /tickets/{id}); requires reverse ETL plan, preview, approval, and execute.
- put_tickets_id_restore:
  - endpoint: PUT /tickets/{{ record.id }}/restore
  - required fields: id
  - risk: Freshdesk Restore a Ticket (PUT /tickets/{id}/restore); requires reverse ETL plan, preview, approval, and execute.
- put_time_entries_id:
  - endpoint: PUT /time_entries/{{ record.id }}
  - required fields: id
  - optional fields: agent_id, billable, executed_at, note, start_time, time_spent, timer_running
  - risk: Freshdesk Update a Time Entry (PUT /time_entries/{id}); requires reverse ETL plan, preview, approval, and execute.

## Security

- read risk: external Freshdesk API reads of support, contact, company, team, admin, collaboration, and settings records through streams and bounded direct reads
- write risk: Freshdesk create, update, merge, restore, export-job, admin, and delete operations are named reverse-ETL actions; execution requires plan, preview, approval, and destructive confirmation for high-risk deletes/admin changes
- approval: Freshdesk writes require reverse ETL plan, preview, approval, execute; destructive/high-risk writes require typed confirmation
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Work with Freshdesk support records from the command line.
- Usage: pm freshdesk <command> <subcommand> [flags]
- Global flags:
  - --json (boolean): Write machine-readable JSON output.
  - --connection (string): Use a saved Freshdesk connector credential.: maps_to=connection
  - --max-bytes (integer): Maximum direct-read response bytes; capped at 1048576.
  - --plan (string): Preview or approve an existing reverse-ETL command plan.
  - --preview (boolean): Preview a reverse-ETL command plan before approval.
  - --approve (string): Approval token for executing an existing reverse-ETL command plan.
  - --confirm (string): Typed confirmation for destructive Freshdesk write plans.
- Stream Reads
  - ticket list - List tickets [intent=etl availability=implemented stream=tickets]; flags: --updated-since
  - contact list - List contacts [intent=etl availability=implemented stream=contacts]
  - company list - List companies [intent=etl availability=implemented stream=companies]
  - agent list - List agents [intent=etl availability=implemented stream=agents]
  - group list - List groups [intent=etl availability=implemented stream=groups]
- Bounded Direct Reads
  - read account - View Account [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read admin groups - List All Groups [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read admin groups id - View a Group [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read admin groups id agents - List All Agents in a Group [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read admin skills - List all Skills [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read admin skills id - View a Skill [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read admin ticket-fields - List All Ticket Fields [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read admin ticket-fields id - View a Ticket Field [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read admin ticket-fields id sections - List All Sections for a Ticket Field [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read admin ticket-fields id sections section-id - List a specific Section details [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id, --section-id
  - read agents - List All Agents Availability [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read agents autocomplete - Search Agents [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --keyword
  - read agents id - View an Agent [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read agents id availability - View Agent Availability [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read agents me - Currently Authenticated Agent [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read automations automation-type-id rules - List All Automation Rules [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --automation-type-id
  - read automations automation-type-id rules id - View an Automation Rule [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --automation-type-id, --id
  - read business-hours - List All Business Hours [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read business-hours id - View a Business Hour [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read canned-response-folders - List All Folders [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read canned-response-folders id - List All Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read canned-response-folders id responses - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read canned-responses id - View a Canned Response [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read channels outbound-messages message-id - Retrieve Message [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --message-id
  - read collaboration messages id - Get message for thread [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read collaboration messages id generate-quote - Get quoted text for message [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read collaboration threads id - Get a thread [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read companies autocomplete - Search Companies [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --keyword
  - read companies export id - Company Export [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read companies id - View a Company [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read companies imports - Company Import [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read companies imports id - Company Import [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read company-fields - List all company fields [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read company-fields id - View a company field [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read contact-fields - List all contact fields [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read contact-fields id - View a contact field [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read contacts autocomplete - Search Contacts [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --keyword
  - read contacts export id - Contact Export [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read contacts id - View a Contact [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read contacts imports - Contact Import [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read contacts imports id - Contact Import [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read custom-objects schemas - Retrieve Existing Custom Objects [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read custom-objects schemas schema-id - Retrieve A Specific Custom Object [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --schema-id
  - read custom-objects schemas schema-id records - Retrieve All Records [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --schema-id
  - read custom-objects schemas schema-id records - Filter Records Of A Custom Object [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --field-name, --operator, --value
  - read custom-objects schemas schema-id records count - Retrieve The Count Of Records [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read custom-objects schemas schema-id records record-id - Retrieve A Record [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --schema-id, --record-id
  - read customer-satisfaction surveys - Surveys [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read customer-satisfaction surveys survey-id responses - View all Survey Responses [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --survey-id
  - read discussions categories - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read discussions categories id - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read discussions categories id forums - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read discussions forums id - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read discussions forums id follow - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id, --id
  - read discussions forums id topics - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read discussions topics followed-by - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read discussions topics id - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read discussions topics id comments - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read discussions topics id follow - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id, --id
  - read discussions topics participated-by - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read email mailboxes - List All Email Mailboxes [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read email mailboxes id - View an Email Mailbox [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read email settings - View Mailbox Settings [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read email-configs - List All Email Configs [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read email-configs id - View an Email Config [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read groups id - View a Group [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read jobs id - View a Job [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read notifications email bcc - View Automatic Bcc emails [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read products - List All Products [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read products id - View a Product [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read roles - List All Roles [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read roles id - View a Role [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read search companies - Filter CompaniesBETA [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --query
  - read search contacts - Filter ContactsBETA [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --query
  - read search solutions - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --keyword
  - read search tickets - Filter Tickets [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --query
  - read settings helpdesk - View Helpdesk Settings [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read sla-policies - List All SLA Policies [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read solutions articles id - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read solutions articles id language-code - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id, --language-code
  - read solutions categories - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read solutions categories category-id folders - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --category-id
  - read solutions categories category-id folders language-code - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --category-id, --language-code
  - read solutions categories id - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read solutions categories id language-code - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id, --language-code
  - read solutions categories language-code - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --language-code
  - read solutions folders folder-id subfolders - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --folder-id
  - read solutions folders folder-id subfolders language-code - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --folder-id, --language-code
  - read solutions folders id - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read solutions folders id articles - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read solutions folders id articles language-code - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id, --language-code
  - read solutions folders id language-code - Get details of Canned Responses in a Folder [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id, --language-code
  - read surveys - List all Surveys [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read surveys satisfaction-ratings - View all Satisfaction Ratings [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read ticket-fields - List All Ticket Fields [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read ticket-forms - List All Ticket Forms [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
  - read ticket-forms form-id fields field-id - View A Ticket Form's Field [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --form-id, --field-id
  - read ticket-forms id - View a Ticket Form [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read tickets archived id - List All Satisfaction Ratings of a Ticket [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read tickets archived id conversations - List All Satisfaction Ratings of a Ticket [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read tickets id - View a Ticket [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read tickets id accesses - List All Satisfaction Ratings of a Ticket [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read tickets id associated-tickets - View a Ticket [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read tickets id conversations - List All Conversations of a Ticket [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read tickets id prime-association - View a Ticket [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read tickets id summary - List All Satisfaction Ratings of a Ticket [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read tickets id time-entries - List All Time Entries of a Ticket [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read tickets id watchers - Ticket Merge API [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --id
  - read tickets ticket-id satisfaction-ratings - List All Satisfaction Ratings of a Ticket [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.; flags: --ticket-id
  - read time-entries - List All Time Entries [intent=direct_read availability=implemented]; notes: Bounded JSON direct read; connector-relative GET only, capped at 1 MiB.
- Reverse ETL Writes
  - write admin groups id delete - Delete a Group [intent=reverse_etl availability=implemented write=delete_admin_groups_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Delete a Group (DELETE /admin/groups/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write admin skills id delete - Delete a Skill [intent=reverse_etl availability=implemented write=delete_admin_skills_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Delete a Skill (DELETE /admin/skills/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write admin ticket-fields id delete - Delete a Ticket Field [intent=reverse_etl availability=implemented write=delete_admin_ticket_fields_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Delete a Ticket Field (DELETE /admin/ticket_fields/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write admin ticket-fields id sections section-id delete - Delete a section [intent=reverse_etl availability=implemented write=delete_admin_ticket_fields_id_sections_section_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Delete a section (DELETE /admin/ticket_fields/{id}/sections/{section_id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id, --section-id
  - write agents id delete - Delete an Agent [intent=reverse_etl availability=implemented write=delete_agents_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Delete an Agent (DELETE /agents/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write attachments id delete - Delete an attachment [intent=reverse_etl availability=implemented write=delete_attachments_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Delete an attachment (DELETE /attachments/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write companies id delete - Delete a Company [intent=reverse_etl availability=implemented write=delete_companies_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Delete a Company (DELETE /companies/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write company-field id delete - Delete a company field [intent=reverse_etl availability=implemented write=delete_company_field_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Delete a company field (DELETE /company_field/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write contact-field id delete - Delete a contact field [intent=reverse_etl availability=implemented write=delete_contact_field_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Delete a contact field (DELETE /contact_field/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write contacts id delete - Soft Delete a Contact [intent=reverse_etl availability=implemented write=delete_contacts_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Soft Delete a Contact (DELETE /contacts/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write contacts id hard-delete delete - Permanently Delete a Contact [intent=reverse_etl availability=implemented write=delete_contacts_id_hard_delete]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Permanently Delete a Contact (DELETE /contacts/{id}/hard_delete); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write conversations id delete - Delete a conversation [intent=reverse_etl availability=implemented write=delete_conversations_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Delete a conversation (DELETE /conversations/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write custom-objects schemas schema-id records record-id delete - Delete A Record [intent=reverse_etl availability=implemented write=delete_custom_objects_schemas_schema_id_records_record_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Delete A Record (DELETE /custom_objects/schemas/{schema_id}/records/{record_id}); requires reverse ETL plan, preview, approval, and execute.; flags: --schema-id, --record-id
  - write discussions categories id delete - Get details of Canned Responses in a Folder [intent=reverse_etl availability=implemented write=delete_discussions_categories_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Get details of Canned Responses in a Folder (DELETE /discussions/categories/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write discussions comments id delete - Get details of Canned Responses in a Folder [intent=reverse_etl availability=implemented write=delete_discussions_comments_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Get details of Canned Responses in a Folder (DELETE /discussions/comments/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write discussions forums forum-id follow delete - Get details of Canned Responses in a Folder [intent=reverse_etl availability=implemented write=delete_discussions_forums_forum_id_follow]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Get details of Canned Responses in a Folder (DELETE /discussions/forums/{forum_id}/follow?user_id={user_id}); requires reverse ETL plan, preview, approval, and execute.; flags: --forum-id, --user-id
  - write discussions forums id delete - Get details of Canned Responses in a Folder [intent=reverse_etl availability=implemented write=delete_discussions_forums_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Get details of Canned Responses in a Folder (DELETE /discussions/forums/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write discussions topics id delete - Get details of Canned Responses in a Folder [intent=reverse_etl availability=implemented write=delete_discussions_topics_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Get details of Canned Responses in a Folder (DELETE /discussions/topics/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write discussions topics topic-id follow delete - Get details of Canned Responses in a Folder [intent=reverse_etl availability=implemented write=delete_discussions_topics_topic_id_follow]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Get details of Canned Responses in a Folder (DELETE /discussions/topics/{topic_id}/follow?user_id={user_id}); requires reverse ETL plan, preview, approval, and execute.; flags: --topic-id, --user-id
  - write email mailboxes id delete - Delete an Email Mailbox [intent=reverse_etl availability=implemented write=delete_email_mailboxes_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Delete an Email Mailbox (DELETE /email/mailboxes/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write groups id delete - Delete a Group [intent=reverse_etl availability=implemented write=delete_groups_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Delete a Group (DELETE /groups/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write solutions articles id delete - Get details of Canned Responses in a Folder [intent=reverse_etl availability=implemented write=delete_solutions_articles_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Get details of Canned Responses in a Folder (DELETE /solutions/articles/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write solutions categories id delete - Get details of Canned Responses in a Folder [intent=reverse_etl availability=implemented write=delete_solutions_categories_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Get details of Canned Responses in a Folder (DELETE /solutions/categories/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write solutions folders id delete - Get details of Canned Responses in a Folder [intent=reverse_etl availability=implemented write=delete_solutions_folders_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Get details of Canned Responses in a Folder (DELETE /solutions/folders/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write ticket-forms form-id fields field-id delete - Delete a Ticket Form's Field [intent=reverse_etl availability=implemented write=delete_ticket_forms_form_id_fields_field_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Delete a Ticket Form's Field (DELETE /ticket-forms/{form_id}/fields/{field_id}); requires reverse ETL plan, preview, approval, and execute.; flags: --form-id, --field-id
  - write ticket-forms id delete - Delete a Ticket Form [intent=reverse_etl availability=implemented write=delete_ticket_forms_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Delete a Ticket Form (DELETE /ticket-forms/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write tickets archived id delete - List All Satisfaction Ratings of a Ticket [intent=reverse_etl availability=implemented write=delete_tickets_archived_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk List All Satisfaction Ratings of a Ticket (DELETE /tickets/archived/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write tickets id delete - Delete a Ticket [intent=reverse_etl availability=implemented write=delete_tickets_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Delete a Ticket (DELETE /tickets/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write tickets id accesses delete - List All Satisfaction Ratings of a Ticket [intent=reverse_etl availability=implemented write=delete_tickets_id_accesses]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk List All Satisfaction Ratings of a Ticket (DELETE /tickets/{id}/accesses); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write tickets id summary delete - List All Satisfaction Ratings of a Ticket [intent=reverse_etl availability=implemented write=delete_tickets_id_summary]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk List All Satisfaction Ratings of a Ticket (DELETE /tickets/{id}/summary); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write time-entries id delete - Delete a Time Entry [intent=reverse_etl availability=implemented write=delete_time_entries_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Delete a Time Entry (DELETE /time_entries/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write companies create - Create a Company [intent=reverse_etl availability=implemented write=post_companies]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Create a Company (POST /companies); requires reverse ETL plan, preview, approval, and execute.; flags: --description, --name, --note, --health-score, --account-tier, --renewal-date, --industry, --lookup-parameter
  - write contacts create - Create a Contact [intent=reverse_etl availability=implemented write=post_contacts]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Create a Contact (POST /contacts); requires reverse ETL plan, preview, approval, and execute.; flags: --name, --email, --phone, --mobile, --twitter-id, --unique-external-id, --other-emails, --company-id, --view-all-tickets, --address, --description, --job-title, --language, --tags, --time-zone, --lookup-parameter
  - write contacts export - Contact Export [intent=reverse_etl availability=implemented write=post_contacts_export]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Contact Export (POST /contacts/export); requires reverse ETL plan, preview, approval, and execute.
  - write contacts merge - Merge Contacts [intent=reverse_etl availability=implemented write=post_contacts_merge]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Merge Contacts (POST /contacts/merge); requires reverse ETL plan, preview, approval, and execute.; flags: --primary-contact-id
  - write tickets create - Create a Ticket [intent=reverse_etl availability=implemented write=post_tickets]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Create a Ticket (POST /tickets); requires reverse ETL plan, preview, approval, and execute.; flags: --name, --requester-id, --email, --facebook-id, --phone, --twitter-id, --unique-external-id, --subject, --type, --status, --priority, --description, --responder-id, --cc-emails, --due-by, --email-config-id, --fr-due-by, --group-id, --parent-id, --product-id, --source, --tags, --company-id, --internal-agent-id, --internal-group-id, --lookup-parameter
  - write tickets bulk-delete - Delete Multiple Tickets [intent=reverse_etl availability=implemented write=post_tickets_bulk_delete]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Delete Multiple Tickets (POST /tickets/bulk_delete); requires reverse ETL plan, preview, approval, and execute.
  - write tickets bulk-update - Update Multiple Tickets [intent=reverse_etl availability=implemented write=post_tickets_bulk_update]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Update Multiple Tickets (POST /tickets/bulk_update); requires reverse ETL plan, preview, approval, and execute.
  - write tickets outbound-email create - Create an Outbound Email [intent=reverse_etl availability=implemented write=post_tickets_outbound_email]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Create an Outbound Email (POST /tickets/outbound_email); requires reverse ETL plan, preview, approval, and execute.; flags: --name, --email, --subject, --type, --status, --priority, --description, --due-by, --email-config-id, --fr-due-by, --group-id, --tags
  - write tickets id forward - Forward a ticket [intent=reverse_etl availability=implemented write=post_tickets_id_forward]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Forward a ticket (POST /tickets/{id}/forward); requires reverse ETL plan, preview, approval, and execute.; flags: --id, --body
  - write agents id update - Update an Agent [intent=reverse_etl availability=implemented write=put_agents_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Update an Agent (PUT /agents/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id, --occasional, --signature, --ticket-scope, --email, --language, --time-zone, --focus-mode
  - write companies id update - Update a Company [intent=reverse_etl availability=implemented write=put_companies_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Update a Company (PUT /companies/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id, --description, --name, --note, --health-score, --account-tier, --renewal-date, --industry, --lookup-parameter
  - write contacts id update - Update a Contact [intent=reverse_etl availability=implemented write=put_contacts_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Update a Contact (PUT /contacts/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id, --name, --email, --phone, --mobile, --twitter-id, --unique-external-id, --other-emails, --company-id, --view-all-tickets, --address, --description, --job-title, --language, --tags, --time-zone, --lookup-parameter
  - write contacts id make-agent - Make Agent [intent=reverse_etl availability=implemented write=put_contacts_id_make_agent]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Make Agent (PUT /contacts/{id}/make_agent); requires reverse ETL plan, preview, approval, and execute.; flags: --id, --occasional, --signature, --ticket-scope, --type, --focus-mode
  - write contacts id restore - Restore a Contact [intent=reverse_etl availability=implemented write=put_contacts_id_restore]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Restore a Contact (PUT /contacts/{id}/restore); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write groups id update - Update a Group [intent=reverse_etl availability=implemented write=put_groups_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Update a Group (PUT /groups/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id, --auto-ticket-assign, --description, --escalate-to, --name, --unassigned-for
  - write tickets merge - Ticket Merge API [intent=reverse_etl availability=implemented write=put_tickets_merge]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Ticket Merge API (PUT /tickets/merge); requires reverse ETL plan, preview, approval, and execute.; flags: --primary-id, --convert-recepients-to-cc
  - write tickets id update - Update a Ticket [intent=reverse_etl availability=implemented write=put_tickets_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Update a Ticket (PUT /tickets/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id, --name, --requester-id, --email, --facebook-id, --phone, --twitter-id, --unique-external-id, --subject, --type, --status, --priority, --description, --responder-id, --due-by, --email-config-id, --fr-due-by, --group-id, --parent-id, --product-id, --source, --tags, --company-id, --internal-agent-id, --internal-group-id, --lookup-parameter
  - write tickets id restore - Restore a Ticket [intent=reverse_etl availability=implemented write=put_tickets_id_restore]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Restore a Ticket (PUT /tickets/{id}/restore); requires reverse ETL plan, preview, approval, and execute.; flags: --id
  - write time-entries id update - Update a Time Entry [intent=reverse_etl availability=implemented write=put_time_entries_id]; approval: reverse ETL writes require plan, preview, approval, and execute; destructive/high-risk Freshdesk writes also require typed confirmation.; risk: Freshdesk Update a Time Entry (PUT /time_entries/{id}); requires reverse ETL plan, preview, approval, and execute.; flags: --id, --agent-id, --billable, --executed-at, --note, --start-time, --time-spent, --timer-running
- Help topics:
  - freshdesk - Freshdesk streams, bounded direct reads, and gated reverse-ETL commands.
  - freshdesk-writes - Freshdesk write commands use named reverse-ETL actions with plan, preview, approval, execute, and destructive confirmation where required.
  - freshdesk-direct-reads - Freshdesk direct reads are fixed connector-relative GET operations with JSON output capped at 1 MiB.

## Commands

### Inspect as a manual

```bash
pm connectors inspect freshdesk
```

### Inspect as structured JSON

```bash
pm connectors inspect freshdesk --json
```

## Agent Rules

- Run pm connectors inspect freshdesk before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
