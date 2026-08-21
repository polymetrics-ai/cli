---
name: pm-gorgias
description: Gorgias connector knowledge and safe action guide.
---

# pm-gorgias

## Purpose

Reads Gorgias helpdesk tickets, customers, messages, and satisfaction surveys through the Gorgias REST API; executes bounded direct reads and a file download across the account, custom fields, events, integrations, jobs, macros, metric cards, phone, rules, search, statistics, tags, teams, users, views, and widgets surface; models Gorgias mutations and a multipart file upload as typed reverse-ETL actions.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url (required)
- mode
- page_size
- username (required)
- password (secret) (required)

## ETL Streams

- tickets:
  - primary key: id
  - cursor: updated_datetime
  - fields: channel(string), closed_datetime(string), created_datetime(string), id(integer), is_unread(boolean), language(string), opened_datetime(string), priority(string), spam(boolean), status(string), subject(string), trashed_datetime(string), updated_datetime(string), via(string)
- customers:
  - primary key: id
  - cursor: updated_datetime
  - fields: channel(string), created_datetime(string), email(string), external_id(string), firstname(string), id(integer), language(string), lastname(string), name(string), timezone(string), updated_datetime(string)
- messages:
  - primary key: id
  - cursor: created_datetime
  - fields: body_text(string), channel(string), created_datetime(string), from_agent(boolean), id(integer), public(boolean), sent_datetime(string), stripped_text(string), subject(string), ticket_id(integer), via(string)
- satisfaction_surveys:
  - primary key: id
  - cursor: created_datetime
  - fields: body_text(string), created_datetime(string), customer_id(integer), id(integer), scale_range(integer), score(integer), scored_datetime(string), sent_datetime(string), ticket_id(integer)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_account_setting:
  - endpoint: POST /account/settings
  - required fields: type
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- update_account_setting:
  - endpoint: PUT /account/settings/{{ record.id }}
  - required fields: id, type
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- create_custom_field:
  - endpoint: POST /custom-fields
  - required fields: object_type, label, definition
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- update_custom_field:
  - endpoint: PUT /custom-fields/{{ record.id }}
  - required fields: id
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- update_custom_fields:
  - endpoint: PUT /custom-fields
  - required fields: custom_fields
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- create_customer:
  - endpoint: POST /customers
  - required fields: channels
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- update_customer:
  - endpoint: PUT /customers/{{ record.id }}
  - required fields: id
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- merge_customers:
  - endpoint: PUT /customers/merge
  - required fields: source_id, target_id
  - optional fields: email, name
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- update_customer_custom_field_values:
  - endpoint: PUT /customers/{{ record.customer_id }}/custom-fields
  - required fields: customer_id, values
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- update_customer_data:
  - endpoint: PUT /customers/{{ record.customer_id }}/data
  - required fields: customer_id, data
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- delete_customer:
  - endpoint: DELETE /customers/{{ record.id }}
  - required fields: id
  - risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
- delete_customers:
  - endpoint: DELETE /customers
  - required fields: ids
  - risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
- create_integration:
  - endpoint: POST /integrations
  - required fields: name, type
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- update_integration:
  - endpoint: PUT /integrations/{{ record.id }}
  - required fields: id, name
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- delete_integration:
  - endpoint: DELETE /integrations/{{ record.id }}
  - required fields: id
  - risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
- create_job:
  - endpoint: POST /jobs
  - required fields: type, params
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- update_job:
  - endpoint: PUT /jobs/{{ record.id }}
  - required fields: id
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- cancel_job:
  - endpoint: DELETE /jobs/{{ record.id }}
  - required fields: id
  - risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
- create_macro:
  - endpoint: POST /macros
  - required fields: name
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- update_macro:
  - endpoint: PUT /macros/{{ record.id }}
  - required fields: id
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- delete_macro:
  - endpoint: DELETE /macros/{{ record.id }}
  - required fields: id
  - risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
- bulk_archive_macros:
  - endpoint: PUT /macros/archive
  - required fields: ids
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- bulk_unarchive_macros:
  - endpoint: PUT /macros/unarchive
  - required fields: ids
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- create_metric_card_feedback:
  - endpoint: POST /metric-cards/{{ record.slug }}/feedback
  - required fields: slug, disagreement_type, metric_category
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- delete_voice_call_recording:
  - endpoint: DELETE /phone/voice-call-recordings/{{ record.id }}
  - required fields: id
  - risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
- create_rule:
  - endpoint: POST /rules
  - required fields: code, name
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- update_rule:
  - endpoint: PUT /rules/{{ record.id }}
  - required fields: id
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- delete_rule:
  - endpoint: DELETE /rules/{{ record.id }}
  - required fields: id
  - risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
- update_rules_priorities:
  - endpoint: POST /rules/priorities
  - required fields: priorities
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- create_satisfaction_survey:
  - endpoint: POST /satisfaction-surveys
  - required fields: customer_id, ticket_id
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- update_satisfaction_survey:
  - endpoint: PUT /satisfaction-surveys/{{ record.id }}
  - required fields: id
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- create_tag:
  - endpoint: POST /tags
  - required fields: name
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- update_tag:
  - endpoint: PUT /tags/{{ record.id }}
  - required fields: id
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- delete_tag:
  - endpoint: DELETE /tags/{{ record.id }}
  - required fields: id
  - risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
- delete_tags:
  - endpoint: DELETE /tags
  - required fields: ids
  - risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
- merge_tags:
  - endpoint: PUT /tags/{{ record.destination_tag_id }}/merge
  - required fields: destination_tag_id, source_tags_ids
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- create_team:
  - endpoint: POST /teams
  - required fields: name
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- update_team:
  - endpoint: PUT /teams/{{ record.id }}
  - required fields: id
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- delete_team:
  - endpoint: DELETE /teams/{{ record.id }}
  - required fields: id
  - risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
- create_ticket:
  - endpoint: POST /tickets
  - required fields: messages
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- update_ticket:
  - endpoint: PUT /tickets/{{ record.id }}
  - required fields: id
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- delete_ticket:
  - endpoint: DELETE /tickets/{{ record.id }}
  - required fields: id
  - risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
- update_ticket_custom_fields:
  - endpoint: PUT /tickets/{{ record.ticket_id }}/custom-fields
  - required fields: ticket_id, values
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- delete_ticket_custom_field:
  - endpoint: DELETE /tickets/{{ record.ticket_id }}/custom-fields/{{ record.id }}
  - required fields: ticket_id, id
  - risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
- create_ticket_message:
  - endpoint: POST /tickets/{{ record.ticket_id }}/messages
  - required fields: ticket_id, channel, from_agent
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- update_ticket_message:
  - endpoint: PUT /tickets/{{ record.ticket_id }}/messages/{{ record.id }}
  - required fields: ticket_id, id, channel, from_agent
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- delete_ticket_message:
  - endpoint: DELETE /tickets/{{ record.ticket_id }}/messages/{{ record.id }}
  - required fields: ticket_id, id
  - risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
- create_ticket_tags:
  - endpoint: POST /tickets/{{ record.ticket_id }}/tags
  - required fields: ticket_id
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- update_ticket_tags:
  - endpoint: PUT /tickets/{{ record.ticket_id }}/tags
  - required fields: ticket_id
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- delete_ticket_tags:
  - endpoint: DELETE /tickets/{{ record.ticket_id }}/tags
  - required fields: ticket_id
  - risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
- create_user:
  - endpoint: POST /users
  - required fields: email, name, role
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- update_user:
  - endpoint: PUT /users/{{ record.id }}
  - required fields: id
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- delete_user:
  - endpoint: DELETE /users/{{ record.id }}
  - required fields: id
  - risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
- create_view:
  - endpoint: POST /views
  - required fields: slug
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- update_view:
  - endpoint: PUT /views/{{ record.id }}
  - required fields: id
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- delete_view:
  - endpoint: DELETE /views/{{ record.id }}
  - required fields: id
  - risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
- create_widget:
  - endpoint: POST /widgets
  - required fields: type, template
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- update_widget:
  - endpoint: PUT /widgets/{{ record.id }}
  - required fields: id
  - risk: medium: mutates Gorgias API state; requires reverse ETL approval
- delete_widget:
  - endpoint: DELETE /widgets/{{ record.id }}
  - required fields: id
  - risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
- delete_customer_custom_field_value:
  - endpoint: DELETE /customers/{{ record.customer_id }}/custom-fields/{{ record.id }}
  - required fields: customer_id, id
  - risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
- upload_file:
  - endpoint: POST /upload
  - required fields: file_path
  - risk: high: uploads a local file's bytes to Gorgias file storage; file path/content are redacted in plans and require reverse ETL approval

## Security

- read risk: external Gorgias API read of helpdesk tickets, customers, messages, satisfaction surveys, and bounded direct reads across account, custom field, event, integration, job, macro, metric card, phone, rule, search, statistics, tag, team, user, view, and widget resources; a bounded file download follows the provider's cross-host storage redirect without forwarding the connector credential
- write risk: typed Gorgias reverse ETL mutations for tickets, customers, users, teams, tags, views, macros, rules, widgets, integrations, jobs, account settings, custom fields, satisfaction surveys, and a multipart file upload
- approval: reverse ETL writes require plan, preview, approval, execute; destructive actions require --confirm destructive
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Work with Gorgias tickets, customers, messages, users, teams, tags, views, macros, rules, widgets, integrations, jobs and files from the command line.
- Usage: pm gorgias <command> [flags]
- Source CLI: Gorgias REST API (https://dash.readme.com/api/v1/api-registry/1qfhqbgmshn434r)
- Global flags:
  - --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
- Account
- Custom Fields
- Customers
- Events
- Files
- Integrations
- Jobs
- Macros
- Messages
- Metric Cards
- Rules
- Satisfaction Surveys
- Satisfaction Surveys
- Search
  - search - Search for resources [intent=direct_read availability=implemented operation=gorgias.search]; notes: Bounded Gorgias read-query; fixed method and path with typed request fields.; flags: --type (required) (enum): The type of search.: values=agent|customer|customer_profile|customer_channel|customer_channel_email|customer_channel_phone|customers_by_phone|integration|team|tag: maps_to=body.type, --query (string): Text query used to search for resources.: maps_to=body.query, --size (integer): Maximum number of results returned (1-50).: maps_to=body.size, --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
- Stats
- Tags
- Teams
- Tickets
- Users
- Views
- Voice Call Events
- Voice Call Recordings
- Voice Calls
- Widgets
- Other Commands
  - tickets list - List Gorgias tickets as ETL records. [intent=etl availability=implemented stream=tickets]
  - customers list - List Gorgias customers as ETL records. [intent=etl availability=implemented stream=customers]
  - messages list - List Gorgias messages as ETL records. [intent=etl availability=implemented stream=messages]
  - satisfaction-surveys list - List Gorgias satisfaction surveys as ETL records. [intent=etl availability=implemented stream=satisfaction_surveys]
  - account get - Retrieve your account [intent=direct_read availability=implemented operation=gorgias.get_account]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - account settings list - List settings [intent=direct_read availability=implemented operation=gorgias.list_account_settings]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - custom-fields list - List custom fields [intent=direct_read availability=implemented operation=gorgias.list_custom_fields]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - custom-fields get - Retrieve a custom field [intent=direct_read availability=implemented operation=gorgias.get_custom_field]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - customers custom-fields list - List customer field values [intent=direct_read availability=implemented operation=gorgias.list_customer_custom_field_values]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - customers get - Retrieve a customer [intent=direct_read availability=implemented operation=gorgias.get_customer]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - events list - List events [intent=direct_read availability=implemented operation=gorgias.list_events]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - events get - Retrieve an event [intent=direct_read availability=implemented operation=gorgias.get_event]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - integrations list - List integrations [intent=direct_read availability=implemented operation=gorgias.list_integrations]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - integrations get - Retrieve an integration [intent=direct_read availability=implemented operation=gorgias.get_integration]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - jobs list - List jobs [intent=direct_read availability=implemented operation=gorgias.list_jobs]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - jobs get - Retrieve a job [intent=direct_read availability=implemented operation=gorgias.get_job]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - macros list - List macros [intent=direct_read availability=implemented operation=gorgias.list_macros]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - macros get - Retrieve a macro [intent=direct_read availability=implemented operation=gorgias.get_macro]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - metric-cards list - Search metric cards [intent=direct_read availability=implemented operation=gorgias.search_metric_cards]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - metric-cards get - Retrieve a metric card by slug [intent=direct_read availability=implemented operation=gorgias.get_metric_card]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - voice-call-events list - List voice call events [intent=direct_read availability=implemented operation=gorgias.list_voice_call_events]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - voice-call-events get - Retrieve a voice call event [intent=direct_read availability=implemented operation=gorgias.get_voice_call_event]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - voice-call-recordings list - List voice call recordings [intent=direct_read availability=implemented operation=gorgias.list_voice_call_recordings]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - voice-call-recordings get - Retrieve a voice call recording [intent=direct_read availability=implemented operation=gorgias.get_voice_call_recording]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - voice-calls list - List voice calls [intent=direct_read availability=implemented operation=gorgias.list_voice_calls]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - voice-calls get - Retrieve a voice call [intent=direct_read availability=implemented operation=gorgias.get_voice_call]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - rules list - List rules [intent=direct_read availability=implemented operation=gorgias.list_rules]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - rules get - Retrieve a rule [intent=direct_read availability=implemented operation=gorgias.get_rule]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - satisfaction-surveys get - Retrieve a survey [intent=direct_read availability=implemented operation=gorgias.get_satisfaction_survey]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - tags list - List tags [intent=direct_read availability=implemented operation=gorgias.list_tags]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - tags get - Retrieve a tag [intent=direct_read availability=implemented operation=gorgias.get_tag]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - teams list - List teams [intent=direct_read availability=implemented operation=gorgias.list_teams]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - teams get - Retrieve a team [intent=direct_read availability=implemented operation=gorgias.get_team]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - tickets get - Retrieve a ticket [intent=direct_read availability=implemented operation=gorgias.get_ticket]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - tickets custom-fields list - List ticket field values [intent=direct_read availability=implemented operation=gorgias.list_ticket_custom_fields]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - tickets messages get - Retrieve a message [intent=direct_read availability=implemented operation=gorgias.get_ticket_message]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - tickets tags list - List ticket tags [intent=direct_read availability=implemented operation=gorgias.list_ticket_tags]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - users list - List users [intent=direct_read availability=implemented operation=gorgias.list_users]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - users get - Retrieve a user [intent=direct_read availability=implemented operation=gorgias.get_user]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - views list - List views [intent=direct_read availability=implemented operation=gorgias.list_views]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - views get - Retrieve a view [intent=direct_read availability=implemented operation=gorgias.get_view]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - views items list - List view's items [intent=direct_read availability=implemented operation=gorgias.list_view_items]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - widgets list - List widgets [intent=direct_read availability=implemented operation=gorgias.list_widgets]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - widgets get - Retrieve a widget [intent=direct_read availability=implemented operation=gorgias.get_widget]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - stats get - Retrieve a statistic [intent=direct_read availability=implemented operation=gorgias.get_legacy_statistic]; notes: Bounded Gorgias read-query; fixed method and path with typed request fields.; flags: --page (integer): Page number to return. Only for connectors whose declared pagination addresses pages by number (page_number, offset_limit); defaults to the first page., --page-cursor (string): Continuation token to return the next page. Use the next_cursor value from a previous page, for connectors whose declared pagination addresses pages by token (cursor, next_url, link_header).
  - files download - Download a file [intent=binary_download availability=partial operation=gorgias.download_file]; notes: Blocked: the provider documents a signed cross-host redirect without a stable exact host allowlist or a bounded final media contract.; flags: --file-type (required, max 4096 bytes) (string): The type of file to download.: maps_to=path.file_type, --domain-hash (required, max 4096 bytes) (string): The domain identifier of the account that owns the resource.: maps_to=path.domain_hash, --resource-name (required, max 4096 bytes) (string): The resource identifier of the attachment you are trying to retrieve.: maps_to=path.resource_name, --dest-root (required) (string): directory the download is written beneath; traversal outside it is refused., --file-name (string): name for the downloaded file within --dest-root; must be a single path segment., --max-bytes (integer): lower the operation's declared size cap; it can never raise it.
  - account settings create - Create account setting. [intent=reverse_etl availability=implemented write=create_account_setting]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --type (required) (string): The type of setting.: maps_to=record.type
  - account settings update - Update account setting. [intent=reverse_etl availability=implemented write=update_account_setting]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id, --type (required) (string): The type of setting.: maps_to=record.type
  - custom-fields create - Create custom field. [intent=reverse_etl availability=partial write=create_custom_field]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: definition is a discriminated union over the custom field's data type (text/dropdown/checkbox/etc.) with no typed scalar leaf; supply it from a reverse-ETL source record; flags: --object-type (required) (string): The object type the custom field applies to.: maps_to=record.object_type, --label (required) (string): The label of the custom field.: maps_to=record.label, --description (string): The description of the custom field.: maps_to=record.description, --priority (integer): The priority of the custom field.: maps_to=record.priority, --required (boolean): Whether the custom field is required.: maps_to=record.required, --external-id (string): An external identifier for the custom field.: maps_to=record.external_id
  - custom-fields update - Update custom field. [intent=reverse_etl availability=implemented write=update_custom_field]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id, --label (string): The label of the custom field.: maps_to=record.label, --description (string): The description of the custom field.: maps_to=record.description, --priority (integer): The priority of the custom field.: maps_to=record.priority, --required (boolean): Whether the custom field is required.: maps_to=record.required, --external-id (string): An external identifier for the custom field.: maps_to=record.external_id
  - custom-fields bulk-update - Update custom fields. [intent=reverse_etl availability=partial write=update_custom_fields]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: the entire request body is a bare JSON array of custom field objects, each itself carrying the same non-flaggable `definition` union; supply it from a reverse-ETL source record; flags: --custom-fields (required) (string): The list of custom fields to bulk update.: maps_to=record.custom_fields
  - customers create - Create customer. [intent=reverse_etl availability=partial write=create_customer]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: channels is a required array of channel objects with no typed scalar leaf; supply it from a reverse-ETL source record; flags: --channels (required) (string): The customer's channels (email, phone, etc.).: maps_to=record.channels, --email (string): The customer's email address.: maps_to=record.email, --name (string): The customer's name.: maps_to=record.name, --external-id (string): An external identifier for the customer.: maps_to=record.external_id, --language (string): The customer's preferred language.: maps_to=record.language, --timezone (string): The customer's timezone.: maps_to=record.timezone
  - customers update - Update customer. [intent=reverse_etl availability=implemented write=update_customer]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id, --email (string): The customer's email address.: maps_to=record.email, --name (string): The customer's name.: maps_to=record.name, --external-id (string): An external identifier for the customer.: maps_to=record.external_id, --language (string): The customer's preferred language.: maps_to=record.language, --timezone (string): The customer's timezone.: maps_to=record.timezone
  - customers merge - Merge customers. [intent=reverse_etl availability=implemented write=merge_customers]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: source_id and target_id are query parameters, mapped to query.source_id/query.target_id; flags: --source-id (required) (integer): The ID of the customer to merge (merged into the target).: maps_to=record.source_id, --target-id (required) (integer): The ID of the targeted customer.: maps_to=record.target_id, --email (string): The customer's email address.: maps_to=record.email, --name (string): The customer's name.: maps_to=record.name
  - customers custom-fields bulk-update - Update customer custom field values. [intent=reverse_etl availability=partial write=update_customer_custom_field_values]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: the entire request body is a bare JSON array of {id, value} objects; supply it from a reverse-ETL source record; flags: --customer-id (required) (string): Path parameter `customer_id`.: maps_to=record.customer_id, --values (required) (string): The list of {id, value} custom field values to set.: maps_to=record.values
  - customers data update - Update customer data. [intent=reverse_etl availability=partial write=update_customer_data]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: data is an arbitrary free-form document with no typed scalar leaf; supply it from a reverse-ETL source record; flags: --customer-id (required) (string): Path parameter `customer_id`.: maps_to=record.customer_id, --data (required) (string): Arbitrary customer data (free-form key/value document).: maps_to=record.data, --version (string): The date of the customer data being written.: maps_to=record.version
  - customers delete - Delete customer. [intent=reverse_etl availability=implemented write=delete_customer]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id
  - customers bulk-delete - Delete customers. [intent=reverse_etl availability=partial write=delete_customers]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; notes: ids is a required array of integer customer IDs with no scalar leaf; supply it from a reverse-ETL source record; flags: --ids (required) (string): The IDs of the customers to delete.: maps_to=record.ids
  - integrations create - Create integration. [intent=reverse_etl availability=implemented write=create_integration]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --name (required) (string): The name of the integration.: maps_to=record.name, --type (required) (string): The type of integration (e.g. `http`).: maps_to=record.type, --description (string): The description of the integration.: maps_to=record.description, --managed (boolean): Whether the integration is managed by Gorgias.: maps_to=record.managed
  - integrations update - Update integration. [intent=reverse_etl availability=implemented write=update_integration]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id, --name (required) (string): The name of the integration.: maps_to=record.name, --description (string): The description of the integration.: maps_to=record.description
  - integrations delete - Delete integration. [intent=reverse_etl availability=implemented write=delete_integration]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id
  - jobs create - Create job. [intent=reverse_etl availability=partial write=create_job]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: params is a job-type-specific object with no fixed shape or scalar leaf; supply it from a reverse-ETL source record; flags: --type (required) (enum): The type of the job.: values=applyMacro|deleteTicket|exportTicket|importMacro|exportMacro|updateTicket|exportTicketDrilldown|exportConvertCampaignSalesDrilldown: maps_to=record.type, --params (required) (string): Job-type-specific parameters.: maps_to=record.params
  - jobs update - Update job. [intent=reverse_etl availability=implemented write=update_job]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id, --status (string): The new status of the job.: maps_to=record.status
  - jobs cancel - Cancel job. [intent=reverse_etl availability=implemented write=cancel_job]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id
  - macros create - Create macro. [intent=reverse_etl availability=implemented write=create_macro]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --name (required) (string): The name of the macro.: maps_to=record.name, --external-id (string): An external identifier for the macro.: maps_to=record.external_id, --intent (string): The macro's intent.: maps_to=record.intent, --language (string): The macro's language.: maps_to=record.language
  - macros update - Update macro. [intent=reverse_etl availability=implemented write=update_macro]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id, --name (string): The name of the macro.: maps_to=record.name, --intent (string): The macro's intent.: maps_to=record.intent, --language (string): The macro's language.: maps_to=record.language
  - macros delete - Delete macro. [intent=reverse_etl availability=implemented write=delete_macro]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id
  - macros archive - Bulk archive macros. [intent=reverse_etl availability=partial write=bulk_archive_macros]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: ids is a required array of integer macro IDs with no scalar leaf; supply it from a reverse-ETL source record; flags: --ids (required) (string): The IDs of the macros to archive.: maps_to=record.ids
  - macros unarchive - Bulk unarchive macros. [intent=reverse_etl availability=partial write=bulk_unarchive_macros]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: ids is a required array of integer macro IDs with no scalar leaf; supply it from a reverse-ETL source record; flags: --ids (required) (string): The IDs of the macros to unarchive.: maps_to=record.ids
  - metric-cards feedback create - Create metric card feedback. [intent=reverse_etl availability=implemented write=create_metric_card_feedback]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --slug (required) (string): Path parameter `slug`.: maps_to=record.slug, --disagreement-type (required) (string): The type of disagreement being reported.: maps_to=record.disagreement_type, --metric-category (required) (string): The metric category the feedback applies to.: maps_to=record.metric_category, --comment (string): Free-text feedback comment.: maps_to=record.comment
  - voice-call-recordings delete - Delete voice call recording. [intent=reverse_etl availability=implemented write=delete_voice_call_recording]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id
  - rules create - Create rule. [intent=reverse_etl availability=implemented write=create_rule]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --code (required) (string): The rule's logic, as Gorgias rule-language source.: maps_to=record.code, --name (required) (string): The name of the rule.: maps_to=record.name, --description (string): The description of the rule.: maps_to=record.description, --priority (integer): The priority of the rule.: maps_to=record.priority, --event-types (string): The event type(s) that trigger the rule.: maps_to=record.event_types
  - rules update - Update rule. [intent=reverse_etl availability=implemented write=update_rule]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id, --code (string): The rule's logic, as Gorgias rule-language source.: maps_to=record.code, --name (string): The name of the rule.: maps_to=record.name, --description (string): The description of the rule.: maps_to=record.description, --priority (integer): The priority of the rule.: maps_to=record.priority
  - rules delete - Delete rule. [intent=reverse_etl availability=implemented write=delete_rule]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id
  - rules priorities update - Update rules priorities. [intent=reverse_etl availability=partial write=update_rules_priorities]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: priorities is a required array of {id, priority} objects with no scalar leaf; supply it from a reverse-ETL source record; flags: --priorities (required) (string): The ordered list of {id, priority} rule priorities.: maps_to=record.priorities
  - satisfaction-surveys create - Create satisfaction survey. [intent=reverse_etl availability=implemented write=create_satisfaction_survey]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --customer-id (required) (integer): The ID of the customer the survey is for.: maps_to=record.customer_id, --ticket-id (required) (integer): The ID of the ticket the survey is for.: maps_to=record.ticket_id, --score (integer): The satisfaction score.: maps_to=record.score, --body-text (string): The free-text body of the survey response.: maps_to=record.body_text
  - satisfaction-surveys update - Update satisfaction survey. [intent=reverse_etl availability=implemented write=update_satisfaction_survey]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id, --score (integer): The satisfaction score.: maps_to=record.score, --body-text (string): The free-text body of the survey response.: maps_to=record.body_text
  - tags create - Create tag. [intent=reverse_etl availability=implemented write=create_tag]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --name (required) (string): The name of the tag.: maps_to=record.name, --description (string): The description of the tag.: maps_to=record.description
  - tags update - Update tag. [intent=reverse_etl availability=implemented write=update_tag]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id, --name (string): The name of the tag.: maps_to=record.name, --description (string): The description of the tag.: maps_to=record.description
  - tags delete - Delete tag. [intent=reverse_etl availability=implemented write=delete_tag]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id
  - tags bulk-delete - Delete tags. [intent=reverse_etl availability=partial write=delete_tags]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; notes: ids is a required array of integer tag IDs with no scalar leaf; supply it from a reverse-ETL source record; flags: --ids (required) (string): The IDs of the tags to delete.: maps_to=record.ids
  - tags merge - Merge tags. [intent=reverse_etl availability=partial write=merge_tags]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: source_tags_ids is a required array of integer tag IDs with no scalar leaf; supply it from a reverse-ETL source record; flags: --destination-tag-id (required) (string): Path parameter `destination_tag_id`.: maps_to=record.destination_tag_id, --source-tags-ids (required) (string): The IDs of the tags to merge into the destination tag.: maps_to=record.source_tags_ids
  - teams create - Create team. [intent=reverse_etl availability=implemented write=create_team]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --name (required) (string): The name of the team.: maps_to=record.name, --description (string): The description of the team.: maps_to=record.description
  - teams update - Update team. [intent=reverse_etl availability=implemented write=update_team]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id, --name (string): The name of the team.: maps_to=record.name, --description (string): The description of the team.: maps_to=record.description
  - teams delete - Delete team. [intent=reverse_etl availability=implemented write=delete_team]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id
  - tickets create - Create ticket. [intent=reverse_etl availability=partial write=create_ticket]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: messages is a required array of message objects (each with its own required, unioned `channel`) with no scalar leaf; supply it from a reverse-ETL source record; flags: --messages (required) (string): The ticket's initial messages.: maps_to=record.messages, --subject (string): The subject of the ticket.: maps_to=record.subject, --status (string): The status of the ticket.: maps_to=record.status, --priority (string): The priority of the ticket.: maps_to=record.priority
  - tickets update - Update ticket. [intent=reverse_etl availability=implemented write=update_ticket]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id, --subject (string): The subject of the ticket.: maps_to=record.subject, --status (string): The status of the ticket.: maps_to=record.status, --priority (string): The priority of the ticket.: maps_to=record.priority, --is-unread (boolean): Whether the ticket is unread.: maps_to=record.is_unread
  - tickets delete - Delete ticket. [intent=reverse_etl availability=implemented write=delete_ticket]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id
  - tickets custom-fields bulk-update - Update ticket custom fields. [intent=reverse_etl availability=partial write=update_ticket_custom_fields]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: the entire request body is a bare JSON array of {id, value} objects; supply it from a reverse-ETL source record; flags: --ticket-id (required) (string): Path parameter `ticket_id`.: maps_to=record.ticket_id, --values (required) (string): The list of {id, value} custom field values to set.: maps_to=record.values
  - tickets custom-fields delete - Delete ticket custom field. [intent=reverse_etl availability=implemented write=delete_ticket_custom_field]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --ticket-id (required) (string): Path parameter `ticket_id`.: maps_to=record.ticket_id, --id (required) (string): Path parameter `id`.: maps_to=record.id
  - tickets messages create - Create ticket message. [intent=reverse_etl availability=implemented write=create_ticket_message]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: channel is documented as a discriminated union of channel-specific shapes; this bundle exposes it as a plain string flag (the channel name), which is the common, honest case; flags: --ticket-id (required) (string): Path parameter `ticket_id`.: maps_to=record.ticket_id, --channel (required) (string): The channel the message was sent through.: maps_to=record.channel, --from-agent (required) (boolean): Whether the message was sent by an agent.: maps_to=record.from_agent, --body-text (string): The plain-text body of the message.: maps_to=record.body_text, --subject (string): The subject of the message.: maps_to=record.subject, --public (boolean): Whether the message is public.: maps_to=record.public
  - tickets messages update - Update ticket message. [intent=reverse_etl availability=implemented write=update_ticket_message]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --ticket-id (required) (string): Path parameter `ticket_id`.: maps_to=record.ticket_id, --id (required) (string): Path parameter `id`.: maps_to=record.id, --channel (required) (string): The channel the message was sent through.: maps_to=record.channel, --from-agent (required) (boolean): Whether the message was sent by an agent.: maps_to=record.from_agent, --body-text (string): The plain-text body of the message.: maps_to=record.body_text, --subject (string): The subject of the message.: maps_to=record.subject
  - tickets messages delete - Delete ticket message. [intent=reverse_etl availability=implemented write=delete_ticket_message]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --ticket-id (required) (string): Path parameter `ticket_id`.: maps_to=record.ticket_id, --id (required) (string): Path parameter `id`.: maps_to=record.id
  - tickets tags create - Create ticket tags. [intent=reverse_etl availability=implemented write=create_ticket_tags]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --ticket-id (required) (string): Path parameter `ticket_id`.: maps_to=record.ticket_id, --names (string_array): The names of the tags to add.: maps_to=record.names
  - tickets tags update - Update ticket tags. [intent=reverse_etl availability=implemented write=update_ticket_tags]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --ticket-id (required) (string): Path parameter `ticket_id`.: maps_to=record.ticket_id, --names (string_array): The names of the tags to set.: maps_to=record.names
  - tickets tags delete - Delete ticket tags. [intent=reverse_etl availability=implemented write=delete_ticket_tags]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --ticket-id (required) (string): Path parameter `ticket_id`.: maps_to=record.ticket_id, --names (string_array): The names of the tags to remove.: maps_to=record.names
  - users create - Create user. [intent=reverse_etl availability=implemented write=create_user]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --email (required) (string): The user's email address.: maps_to=record.email, --name (required) (string): The user's name.: maps_to=record.name, --role (required) (string): The user's role.: maps_to=record.role, --bio (string): The user's bio.: maps_to=record.bio, --country (string): The user's country.: maps_to=record.country, --language (string): The user's preferred language.: maps_to=record.language, --timezone (string): The user's timezone.: maps_to=record.timezone
  - users update - Update user. [intent=reverse_etl availability=implemented write=update_user]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id, --email (string): The user's email address.: maps_to=record.email, --name (string): The user's name.: maps_to=record.name, --role (string): The user's role.: maps_to=record.role, --bio (string): The user's bio.: maps_to=record.bio
  - users delete - Delete user. [intent=reverse_etl availability=implemented write=delete_user]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id
  - views create - Create view. [intent=reverse_etl availability=implemented write=create_view]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --slug (required) (string): The unique slug of the view.: maps_to=record.slug, --name (string): The name of the view.: maps_to=record.name, --type (string): The type of view.: maps_to=record.type, --order-by (string): The attribute used to order the view.: maps_to=record.order_by, --order-dir (string): The direction used to order the view.: maps_to=record.order_dir, --visibility (string): The visibility of the view.: maps_to=record.visibility
  - views update - Update view. [intent=reverse_etl availability=implemented write=update_view]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id, --name (string): The name of the view.: maps_to=record.name, --order-by (string): The attribute used to order the view.: maps_to=record.order_by, --order-dir (string): The direction used to order the view.: maps_to=record.order_dir, --visibility (string): The visibility of the view.: maps_to=record.visibility
  - views delete - Delete view. [intent=reverse_etl availability=implemented write=delete_view]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id
  - widgets create - Create widget. [intent=reverse_etl availability=partial write=create_widget]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: template is a widget-type-specific rendering object with no fixed shape or scalar leaf; supply it from a reverse-ETL source record; flags: --type (required) (string): The type of widget.: maps_to=record.type, --template (required) (string): The widget's rendering template.: maps_to=record.template, --integration-id (integer): The ID of the integration the widget belongs to.: maps_to=record.integration_id, --order (integer): The display order of the widget.: maps_to=record.order
  - widgets update - Update widget. [intent=reverse_etl availability=implemented write=update_widget]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id, --order (integer): The display order of the widget.: maps_to=record.order, --integration-id (integer): The ID of the integration the widget belongs to.: maps_to=record.integration_id
  - widgets delete - Delete widget. [intent=reverse_etl availability=implemented write=delete_widget]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --id (required) (string): Path parameter `id`.: maps_to=record.id
  - customers custom-fields delete - Delete customer custom field value. [intent=reverse_etl availability=implemented write=delete_customer_custom_field_value]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --customer-id (required) (string): Path parameter `customer_id`.: maps_to=record.customer_id, --id (required) (string): Path parameter `id`.: maps_to=record.id
  - files upload - Upload a file to Gorgias file storage. [intent=reverse_etl availability=implemented write=upload_file]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: uploads a local file's bytes to Gorgias file storage; file path/content are redacted in plans and require reverse ETL approval; notes: Uses typed multipart write support; no generic upload command is exposed.; flags: --file-path (required) (string): Project-relative path to the file to upload.: maps_to=record.file_path, --type (enum): The type of file to upload.: values=public_attachment|attachment|profile_picture|widget_picture: maps_to=record.type
- Help topics:
  - safety - Reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.
  - pagination - Gorgias paginates with cursor/limit and meta.next_cursor; ETL streams follow the cursor to exhaustion within the configured page bounds.

## Commands

### Inspect as a manual

```bash
pm connectors inspect gorgias
```

### Inspect as structured JSON

```bash
pm connectors inspect gorgias --json
```

## Agent Rules

- Run pm connectors inspect gorgias before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
