# pm connectors inspect gorgias

```text
NAME
  pm connectors inspect gorgias - Gorgias connector manual

SYNOPSIS
  pm connectors inspect gorgias
  pm connectors inspect gorgias --json
  pm credentials add <name> --connector gorgias [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Gorgias helpdesk tickets, customers, messages, and satisfaction surveys through the Gorgias REST API; executes bounded direct reads and a file download across the account, custom fields, events, integrations, jobs, macros, metric cards, phone, rules, search, statistics, tags, teams, users, views, and widgets surface; models Gorgias mutations and a multipart file upload as typed reverse-ETL actions.

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
  base_url (required)
  mode
  page_size
  username (required)
  password (secret) (required)

ETL STREAMS
  tickets:
    primary key: id
    cursor: updated_datetime
    fields: channel(string), closed_datetime(string), created_datetime(string), id(integer), is_unread(boolean), language(string), opened_datetime(string), priority(string), spam(boolean), status(string), subject(string), trashed_datetime(string), updated_datetime(string), via(string)
  customers:
    primary key: id
    cursor: updated_datetime
    fields: channel(string), created_datetime(string), email(string), external_id(string), firstname(string), id(integer), language(string), lastname(string), name(string), timezone(string), updated_datetime(string)
  messages:
    primary key: id
    cursor: created_datetime
    fields: body_text(string), channel(string), created_datetime(string), from_agent(boolean), id(integer), public(boolean), sent_datetime(string), stripped_text(string), subject(string), ticket_id(integer), via(string)
  satisfaction_surveys:
    primary key: id
    cursor: created_datetime
    fields: body_text(string), created_datetime(string), customer_id(integer), id(integer), scale_range(integer), score(integer), scored_datetime(string), sent_datetime(string), ticket_id(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_account_setting:
    endpoint: POST /account/settings
    required fields: type
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  update_account_setting:
    endpoint: PUT /account/settings/{{ record.id }}
    required fields: id, type
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  create_custom_field:
    endpoint: POST /custom-fields
    required fields: object_type, label, definition
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  update_custom_field:
    endpoint: PUT /custom-fields/{{ record.id }}
    required fields: id
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  update_custom_fields:
    endpoint: PUT /custom-fields
    required fields: custom_fields
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  create_customer:
    endpoint: POST /customers
    required fields: channels
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  update_customer:
    endpoint: PUT /customers/{{ record.id }}
    required fields: id
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  merge_customers:
    endpoint: PUT /customers/merge
    required fields: source_id, target_id
    optional fields: email, name
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  update_customer_custom_field_values:
    endpoint: PUT /customers/{{ record.customer_id }}/custom-fields
    required fields: customer_id, values
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  update_customer_data:
    endpoint: PUT /customers/{{ record.customer_id }}/data
    required fields: customer_id, data
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  delete_customer:
    endpoint: DELETE /customers/{{ record.id }}
    required fields: id
    risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
  delete_customers:
    endpoint: DELETE /customers
    required fields: ids
    risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
  create_integration:
    endpoint: POST /integrations
    required fields: name, type
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  update_integration:
    endpoint: PUT /integrations/{{ record.id }}
    required fields: id, name
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  delete_integration:
    endpoint: DELETE /integrations/{{ record.id }}
    required fields: id
    risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
  create_job:
    endpoint: POST /jobs
    required fields: type, params
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  update_job:
    endpoint: PUT /jobs/{{ record.id }}
    required fields: id
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  cancel_job:
    endpoint: DELETE /jobs/{{ record.id }}
    required fields: id
    risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
  create_macro:
    endpoint: POST /macros
    required fields: name
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  update_macro:
    endpoint: PUT /macros/{{ record.id }}
    required fields: id
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  delete_macro:
    endpoint: DELETE /macros/{{ record.id }}
    required fields: id
    risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
  bulk_archive_macros:
    endpoint: PUT /macros/archive
    required fields: ids
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  bulk_unarchive_macros:
    endpoint: PUT /macros/unarchive
    required fields: ids
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  create_metric_card_feedback:
    endpoint: POST /metric-cards/{{ record.slug }}/feedback
    required fields: slug, disagreement_type, metric_category
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  delete_voice_call_recording:
    endpoint: DELETE /phone/voice-call-recordings/{{ record.id }}
    required fields: id
    risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
  create_rule:
    endpoint: POST /rules
    required fields: code, name
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  update_rule:
    endpoint: PUT /rules/{{ record.id }}
    required fields: id
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  delete_rule:
    endpoint: DELETE /rules/{{ record.id }}
    required fields: id
    risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
  update_rules_priorities:
    endpoint: POST /rules/priorities
    required fields: priorities
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  create_satisfaction_survey:
    endpoint: POST /satisfaction-surveys
    required fields: customer_id, ticket_id
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  update_satisfaction_survey:
    endpoint: PUT /satisfaction-surveys/{{ record.id }}
    required fields: id
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  create_tag:
    endpoint: POST /tags
    required fields: name
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  update_tag:
    endpoint: PUT /tags/{{ record.id }}
    required fields: id
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  delete_tag:
    endpoint: DELETE /tags/{{ record.id }}
    required fields: id
    risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
  delete_tags:
    endpoint: DELETE /tags
    required fields: ids
    risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
  merge_tags:
    endpoint: PUT /tags/{{ record.destination_tag_id }}/merge
    required fields: destination_tag_id, source_tags_ids
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  create_team:
    endpoint: POST /teams
    required fields: name
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  update_team:
    endpoint: PUT /teams/{{ record.id }}
    required fields: id
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  delete_team:
    endpoint: DELETE /teams/{{ record.id }}
    required fields: id
    risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
  create_ticket:
    endpoint: POST /tickets
    required fields: messages
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  update_ticket:
    endpoint: PUT /tickets/{{ record.id }}
    required fields: id
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  delete_ticket:
    endpoint: DELETE /tickets/{{ record.id }}
    required fields: id
    risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
  update_ticket_custom_fields:
    endpoint: PUT /tickets/{{ record.ticket_id }}/custom-fields
    required fields: ticket_id, values
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  delete_ticket_custom_field:
    endpoint: DELETE /tickets/{{ record.ticket_id }}/custom-fields/{{ record.id }}
    required fields: ticket_id, id
    risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
  create_ticket_message:
    endpoint: POST /tickets/{{ record.ticket_id }}/messages
    required fields: ticket_id, channel, from_agent
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  update_ticket_message:
    endpoint: PUT /tickets/{{ record.ticket_id }}/messages/{{ record.id }}
    required fields: ticket_id, id, channel, from_agent
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  delete_ticket_message:
    endpoint: DELETE /tickets/{{ record.ticket_id }}/messages/{{ record.id }}
    required fields: ticket_id, id
    risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
  create_ticket_tags:
    endpoint: POST /tickets/{{ record.ticket_id }}/tags
    required fields: ticket_id
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  update_ticket_tags:
    endpoint: PUT /tickets/{{ record.ticket_id }}/tags
    required fields: ticket_id
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  delete_ticket_tags:
    endpoint: DELETE /tickets/{{ record.ticket_id }}/tags
    required fields: ticket_id
    risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
  create_user:
    endpoint: POST /users
    required fields: email, name, role
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  update_user:
    endpoint: PUT /users/{{ record.id }}
    required fields: id
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  delete_user:
    endpoint: DELETE /users/{{ record.id }}
    required fields: id
    risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
  create_view:
    endpoint: POST /views
    required fields: slug
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  update_view:
    endpoint: PUT /views/{{ record.id }}
    required fields: id
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  delete_view:
    endpoint: DELETE /views/{{ record.id }}
    required fields: id
    risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
  create_widget:
    endpoint: POST /widgets
    required fields: type, template
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  update_widget:
    endpoint: PUT /widgets/{{ record.id }}
    required fields: id
    risk: medium: mutates Gorgias API state; requires reverse ETL approval
  delete_widget:
    endpoint: DELETE /widgets/{{ record.id }}
    required fields: id
    risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
  delete_customer_custom_field_value:
    endpoint: DELETE /customers/{{ record.customer_id }}/custom-fields/{{ record.id }}
    required fields: customer_id, id
    risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation
  upload_file:
    endpoint: POST /upload
    required fields: file_path
    risk: high: uploads a local file's bytes to Gorgias file storage; file path/content are redacted in plans and require reverse ETL approval

SECURITY
  read risk: external Gorgias API read of helpdesk tickets, customers, messages, satisfaction surveys, and bounded direct reads across account, custom field, event, integration, job, macro, metric card, phone, rule, search, statistics, tag, team, user, view, and widget resources; a bounded file download follows the provider's cross-host storage redirect without forwarding the connector credential
  write risk: typed Gorgias reverse ETL mutations for tickets, customers, users, teams, tags, views, macros, rules, widgets, integrations, jobs, account settings, custom fields, satisfaction surveys, and a multipart file upload
  approval: reverse ETL writes require plan, preview, approval, execute; destructive actions require --confirm destructive
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Work with Gorgias tickets, customers, messages, users, teams, tags, views, macros, rules, widgets, integrations, jobs and files from the command line.
  Usage: pm gorgias <command> [flags]
  PM execution policy pm-request-contract-bounds-v1: each max N bytes qualifier is the effective PM request limit, not a provider schema assertion; path/query values are measured after exact wire encoding and rejected rather than truncated.
  Global flags:
    --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
  Account
  Custom Fields
  Customers
  Events
  Files
  Integrations
  Jobs
  Macros
  Messages
  Metric Cards
  Rules
  Satisfaction Surveys
  Satisfaction Surveys
  Search
    search - Search for resources [intent=direct_read availability=implemented operation=gorgias.search]; notes: Bounded Gorgias read-query; fixed method and path with typed request fields.; flags: --type (required), --query, --size, --page, --page-cursor
  Stats
  Tags
  Teams
  Tickets
  Users
  Views
  Voice Call Events
  Voice Call Recordings
  Voice Calls
  Widgets
  Other Commands
    tickets list - List Gorgias tickets as ETL records. [intent=etl availability=implemented stream=tickets]
    customers list - List Gorgias customers as ETL records. [intent=etl availability=implemented stream=customers]
    messages list - List Gorgias messages as ETL records. [intent=etl availability=implemented stream=messages]
    satisfaction-surveys list - List Gorgias satisfaction surveys as ETL records. [intent=etl availability=implemented stream=satisfaction_surveys]
    account get - Retrieve your account [intent=direct_read availability=implemented operation=gorgias.get_account]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    account settings list - List settings [intent=direct_read availability=implemented operation=gorgias.list_account_settings]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    custom-fields list - List custom fields [intent=direct_read availability=implemented operation=gorgias.list_custom_fields]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    custom-fields get - Retrieve a custom field [intent=direct_read availability=implemented operation=gorgias.get_custom_field]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    customers custom-fields list - List customer field values [intent=direct_read availability=implemented operation=gorgias.list_customer_custom_field_values]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    customers get - Retrieve a customer [intent=direct_read availability=implemented operation=gorgias.get_customer]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    events list - List events [intent=direct_read availability=implemented operation=gorgias.list_events]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    events get - Retrieve an event [intent=direct_read availability=implemented operation=gorgias.get_event]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    integrations list - List integrations [intent=direct_read availability=implemented operation=gorgias.list_integrations]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    integrations get - Retrieve an integration [intent=direct_read availability=implemented operation=gorgias.get_integration]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    jobs list - List jobs [intent=direct_read availability=implemented operation=gorgias.list_jobs]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    jobs get - Retrieve a job [intent=direct_read availability=implemented operation=gorgias.get_job]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    macros list - List macros [intent=direct_read availability=implemented operation=gorgias.list_macros]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    macros get - Retrieve a macro [intent=direct_read availability=implemented operation=gorgias.get_macro]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    metric-cards list - Search metric cards [intent=direct_read availability=implemented operation=gorgias.search_metric_cards]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    metric-cards get - Retrieve a metric card by slug [intent=direct_read availability=implemented operation=gorgias.get_metric_card]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    voice-call-events list - List voice call events [intent=direct_read availability=implemented operation=gorgias.list_voice_call_events]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    voice-call-events get - Retrieve a voice call event [intent=direct_read availability=implemented operation=gorgias.get_voice_call_event]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    voice-call-recordings list - List voice call recordings [intent=direct_read availability=implemented operation=gorgias.list_voice_call_recordings]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    voice-call-recordings get - Retrieve a voice call recording [intent=direct_read availability=implemented operation=gorgias.get_voice_call_recording]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    voice-calls list - List voice calls [intent=direct_read availability=implemented operation=gorgias.list_voice_calls]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    voice-calls get - Retrieve a voice call [intent=direct_read availability=implemented operation=gorgias.get_voice_call]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    rules list - List rules [intent=direct_read availability=implemented operation=gorgias.list_rules]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    rules get - Retrieve a rule [intent=direct_read availability=implemented operation=gorgias.get_rule]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    satisfaction-surveys get - Retrieve a survey [intent=direct_read availability=implemented operation=gorgias.get_satisfaction_survey]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    tags list - List tags [intent=direct_read availability=implemented operation=gorgias.list_tags]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    tags get - Retrieve a tag [intent=direct_read availability=implemented operation=gorgias.get_tag]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    teams list - List teams [intent=direct_read availability=implemented operation=gorgias.list_teams]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    teams get - Retrieve a team [intent=direct_read availability=implemented operation=gorgias.get_team]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    tickets get - Retrieve a ticket [intent=direct_read availability=implemented operation=gorgias.get_ticket]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    tickets custom-fields list - List ticket field values [intent=direct_read availability=implemented operation=gorgias.list_ticket_custom_fields]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    tickets messages get - Retrieve a message [intent=direct_read availability=implemented operation=gorgias.get_ticket_message]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    tickets tags list - List ticket tags [intent=direct_read availability=implemented operation=gorgias.list_ticket_tags]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    users list - List users [intent=direct_read availability=implemented operation=gorgias.list_users]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    users get - Retrieve a user [intent=direct_read availability=implemented operation=gorgias.get_user]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    views list - List views [intent=direct_read availability=implemented operation=gorgias.list_views]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    views get - Retrieve a view [intent=direct_read availability=implemented operation=gorgias.get_view]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    views items list - List view's items [intent=direct_read availability=implemented operation=gorgias.list_view_items]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    widgets list - List widgets [intent=direct_read availability=implemented operation=gorgias.list_widgets]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    widgets get - Retrieve a widget [intent=direct_read availability=implemented operation=gorgias.get_widget]; notes: Bounded Gorgias read; fixed method and path with typed request fields.; flags: --page, --page-cursor
    stats get - Retrieve a statistic [intent=direct_read availability=implemented operation=gorgias.get_legacy_statistic]; notes: Bounded Gorgias read-query; fixed method and path with typed request fields.; flags: --page, --page-cursor
    files download - Download a file [intent=binary_download availability=partial operation=gorgias.download_file]; notes: Blocked: the provider documents a signed cross-host redirect without a stable exact host allowlist or a bounded final media contract.; flags: --file-type (required, max 4096 bytes), --domain-hash (required, max 4096 bytes), --resource-name (required, max 4096 bytes), --dest-root (required), --file-name, --max-bytes
    account settings create - Create account setting. [intent=reverse_etl availability=implemented write=create_account_setting]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --type (required)
    account settings update - Update account setting. [intent=reverse_etl availability=implemented write=update_account_setting]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required), --type (required)
    custom-fields create - Create custom field. [intent=reverse_etl availability=partial write=create_custom_field]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: definition is a discriminated union over the custom field's data type (text/dropdown/checkbox/etc.) with no typed scalar leaf; supply it from a reverse-ETL source record; flags: --object-type (required), --label (required), --description, --priority, --required, --external-id
    custom-fields update - Update custom field. [intent=reverse_etl availability=implemented write=update_custom_field]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required), --label, --description, --priority, --required, --external-id
    custom-fields bulk-update - Update custom fields. [intent=reverse_etl availability=partial write=update_custom_fields]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: the entire request body is a bare JSON array of custom field objects, each itself carrying the same non-flaggable `definition` union; supply it from a reverse-ETL source record; flags: --custom-fields (required)
    customers create - Create customer. [intent=reverse_etl availability=partial write=create_customer]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: channels is a required array of channel objects with no typed scalar leaf; supply it from a reverse-ETL source record; flags: --channels (required), --email, --name, --external-id, --language, --timezone
    customers update - Update customer. [intent=reverse_etl availability=implemented write=update_customer]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required), --email, --name, --external-id, --language, --timezone
    customers merge - Merge customers. [intent=reverse_etl availability=implemented write=merge_customers]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: source_id and target_id are query parameters, mapped to query.source_id/query.target_id; flags: --source-id (required), --target-id (required), --email, --name
    customers custom-fields bulk-update - Update customer custom field values. [intent=reverse_etl availability=partial write=update_customer_custom_field_values]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: the entire request body is a bare JSON array of {id, value} objects; supply it from a reverse-ETL source record; flags: --customer-id (required), --values (required)
    customers data update - Update customer data. [intent=reverse_etl availability=partial write=update_customer_data]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: data is an arbitrary free-form document with no typed scalar leaf; supply it from a reverse-ETL source record; flags: --customer-id (required), --data (required), --version
    customers delete - Delete customer. [intent=reverse_etl availability=implemented write=delete_customer]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --id (required)
    customers bulk-delete - Delete customers. [intent=reverse_etl availability=partial write=delete_customers]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; notes: ids is a required array of integer customer IDs with no scalar leaf; supply it from a reverse-ETL source record; flags: --ids (required)
    integrations create - Create integration. [intent=reverse_etl availability=implemented write=create_integration]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --name (required), --type (required), --description, --managed
    integrations update - Update integration. [intent=reverse_etl availability=implemented write=update_integration]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required), --name (required), --description
    integrations delete - Delete integration. [intent=reverse_etl availability=implemented write=delete_integration]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --id (required)
    jobs create - Create job. [intent=reverse_etl availability=partial write=create_job]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: params is a job-type-specific object with no fixed shape or scalar leaf; supply it from a reverse-ETL source record; flags: --type (required), --params (required)
    jobs update - Update job. [intent=reverse_etl availability=implemented write=update_job]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required), --status
    jobs cancel - Cancel job. [intent=reverse_etl availability=implemented write=cancel_job]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --id (required)
    macros create - Create macro. [intent=reverse_etl availability=implemented write=create_macro]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --name (required), --external-id, --intent, --language
    macros update - Update macro. [intent=reverse_etl availability=implemented write=update_macro]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required), --name, --intent, --language
    macros delete - Delete macro. [intent=reverse_etl availability=implemented write=delete_macro]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --id (required)
    macros archive - Bulk archive macros. [intent=reverse_etl availability=partial write=bulk_archive_macros]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: ids is a required array of integer macro IDs with no scalar leaf; supply it from a reverse-ETL source record; flags: --ids (required)
    macros unarchive - Bulk unarchive macros. [intent=reverse_etl availability=partial write=bulk_unarchive_macros]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: ids is a required array of integer macro IDs with no scalar leaf; supply it from a reverse-ETL source record; flags: --ids (required)
    metric-cards feedback create - Create metric card feedback. [intent=reverse_etl availability=implemented write=create_metric_card_feedback]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --slug (required), --disagreement-type (required), --metric-category (required), --comment
    voice-call-recordings delete - Delete voice call recording. [intent=reverse_etl availability=implemented write=delete_voice_call_recording]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --id (required)
    rules create - Create rule. [intent=reverse_etl availability=implemented write=create_rule]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --code (required), --name (required), --description, --priority, --event-types
    rules update - Update rule. [intent=reverse_etl availability=implemented write=update_rule]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required), --code, --name, --description, --priority
    rules delete - Delete rule. [intent=reverse_etl availability=implemented write=delete_rule]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --id (required)
    rules priorities update - Update rules priorities. [intent=reverse_etl availability=partial write=update_rules_priorities]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: priorities is a required array of {id, priority} objects with no scalar leaf; supply it from a reverse-ETL source record; flags: --priorities (required)
    satisfaction-surveys create - Create satisfaction survey. [intent=reverse_etl availability=implemented write=create_satisfaction_survey]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --customer-id (required), --ticket-id (required), --score, --body-text
    satisfaction-surveys update - Update satisfaction survey. [intent=reverse_etl availability=implemented write=update_satisfaction_survey]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required), --score, --body-text
    tags create - Create tag. [intent=reverse_etl availability=implemented write=create_tag]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --name (required), --description
    tags update - Update tag. [intent=reverse_etl availability=implemented write=update_tag]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required), --name, --description
    tags delete - Delete tag. [intent=reverse_etl availability=implemented write=delete_tag]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --id (required)
    tags bulk-delete - Delete tags. [intent=reverse_etl availability=partial write=delete_tags]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; notes: ids is a required array of integer tag IDs with no scalar leaf; supply it from a reverse-ETL source record; flags: --ids (required)
    tags merge - Merge tags. [intent=reverse_etl availability=partial write=merge_tags]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: source_tags_ids is a required array of integer tag IDs with no scalar leaf; supply it from a reverse-ETL source record; flags: --destination-tag-id (required), --source-tags-ids (required)
    teams create - Create team. [intent=reverse_etl availability=implemented write=create_team]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --name (required), --description
    teams update - Update team. [intent=reverse_etl availability=implemented write=update_team]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required), --name, --description
    teams delete - Delete team. [intent=reverse_etl availability=implemented write=delete_team]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --id (required)
    tickets create - Create ticket. [intent=reverse_etl availability=partial write=create_ticket]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: messages is a required array of message objects (each with its own required, unioned `channel`) with no scalar leaf; supply it from a reverse-ETL source record; flags: --messages (required), --subject, --status, --priority
    tickets update - Update ticket. [intent=reverse_etl availability=implemented write=update_ticket]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required), --subject, --status, --priority, --is-unread
    tickets delete - Delete ticket. [intent=reverse_etl availability=implemented write=delete_ticket]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --id (required)
    tickets custom-fields bulk-update - Update ticket custom fields. [intent=reverse_etl availability=partial write=update_ticket_custom_fields]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: the entire request body is a bare JSON array of {id, value} objects; supply it from a reverse-ETL source record; flags: --ticket-id (required), --values (required)
    tickets custom-fields delete - Delete ticket custom field. [intent=reverse_etl availability=implemented write=delete_ticket_custom_field]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --ticket-id (required), --id (required)
    tickets messages create - Create ticket message. [intent=reverse_etl availability=implemented write=create_ticket_message]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: channel is documented as a discriminated union of channel-specific shapes; this bundle exposes it as a plain string flag (the channel name), which is the common, honest case; flags: --ticket-id (required), --channel (required), --from-agent (required), --body-text, --subject, --public
    tickets messages update - Update ticket message. [intent=reverse_etl availability=implemented write=update_ticket_message]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --ticket-id (required), --id (required), --channel (required), --from-agent (required), --body-text, --subject
    tickets messages delete - Delete ticket message. [intent=reverse_etl availability=implemented write=delete_ticket_message]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --ticket-id (required), --id (required)
    tickets tags create - Create ticket tags. [intent=reverse_etl availability=implemented write=create_ticket_tags]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --ticket-id (required), --names
    tickets tags update - Update ticket tags. [intent=reverse_etl availability=implemented write=update_ticket_tags]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --ticket-id (required), --names
    tickets tags delete - Delete ticket tags. [intent=reverse_etl availability=implemented write=delete_ticket_tags]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --ticket-id (required), --names
    users create - Create user. [intent=reverse_etl availability=implemented write=create_user]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --email (required), --name (required), --role (required), --bio, --country, --language, --timezone
    users update - Update user. [intent=reverse_etl availability=implemented write=update_user]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required), --email, --name, --role, --bio
    users delete - Delete user. [intent=reverse_etl availability=implemented write=delete_user]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --id (required)
    views create - Create view. [intent=reverse_etl availability=implemented write=create_view]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --slug (required), --name, --type, --order-by, --order-dir, --visibility
    views update - Update view. [intent=reverse_etl availability=implemented write=update_view]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required), --name, --order-by, --order-dir, --visibility
    views delete - Delete view. [intent=reverse_etl availability=implemented write=delete_view]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --id (required)
    widgets create - Create widget. [intent=reverse_etl availability=partial write=create_widget]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; notes: template is a widget-type-specific rendering object with no fixed shape or scalar leaf; supply it from a reverse-ETL source record; flags: --type (required), --template (required), --integration-id, --order
    widgets update - Update widget. [intent=reverse_etl availability=implemented write=update_widget]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: medium: mutates Gorgias API state; requires reverse ETL approval; flags: --id (required), --order, --integration-id
    widgets delete - Delete widget. [intent=reverse_etl availability=implemented write=delete_widget]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --id (required)
    customers custom-fields delete - Delete customer custom field value. [intent=reverse_etl availability=implemented write=delete_customer_custom_field_value]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: removes Gorgias state; requires reverse ETL approval and destructive confirmation; flags: --customer-id (required), --id (required)
    files upload - Upload a file to Gorgias file storage. [intent=reverse_etl availability=implemented write=upload_file]; approval: reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.; risk: high: uploads a local file's bytes to Gorgias file storage; file path/content are redacted in plans and require reverse ETL approval; notes: Uses typed multipart write support; no generic upload command is exposed.; flags: --file-path (required), --type
  Help topics:
    safety - Reverse ETL mutations require plan, preview, explicit approval, and execute; destructive actions require destructive confirmation.
    pagination - Gorgias paginates with cursor/limit and meta.next_cursor; ETL streams follow the cursor to exhaustion within the configured page bounds.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect gorgias

  # Inspect as structured JSON
  pm connectors inspect gorgias --json

AGENT WORKFLOW
  - Run pm connectors inspect gorgias before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
