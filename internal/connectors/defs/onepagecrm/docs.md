# Overview

Reads the documented OnePageCRM API v3 CRM surface and exposes write actions for
supported JSON/path mutations.

Readable streams: `contacts`, `deals`, `actions`, `companies`, `users`, `bootstrap`, `user`,
`lead_sources`, `lead_source`, `statuses`, `status`, `deal_fields`, `deal_field`, `custom_fields`,
`custom_field`, `company_fields`, `company_field`, `predefined_actions`, `predefined_action`,
`predefined_action_groups`, `predefined_action_group`, `predefined_items`, `predefined_item`,
`predefined_item_groups`, `predefined_item_group`, `notes`, `note`, `calls`, `call`, `call_results`,
`meetings`, `meeting`, `deal`, `relationship_types`, `relationship_type`, `countries`, `action`,
`filters`, `filter`, `company`, `company_actions`, `company_deals`, `company_notes`,
`company_calls`, `company_meetings`, `company_linked_contacts`, `company_pinned_attachments`,
`contact`, `filtered_contacts`, `contact_actions`, `contact_deals`, `contact_notes`,
`contact_calls`, `contact_meetings`, `contact_relationships`, `contact_relationship`,
`contact_pinned_attachments`, `contacts_cascade`, `contacts_cascade_after`, `action_stream`,
`team_stream`, `notifications`, `notification`, `webhooks`, `webhook`, `pipelines`, `pipeline`.

Write actions: `update_user`, `create_lead_source`, `update_lead_source`, `delete_lead_source`,
`create_status`, `update_status`, `delete_status`, `create_deal_field`, `update_deal_field`,
`delete_deal_field`, `create_custom_field`, `update_custom_field`, `delete_custom_field`,
`create_company_field`, `update_company_field`, `delete_company_field`, `create_predefined_action`,
`update_predefined_action`, `delete_predefined_action`, `create_predefined_action_group`,
`update_predefined_action_group`, `delete_predefined_action_group`, `create_predefined_item`,
`update_predefined_item`, `delete_predefined_item`, `create_predefined_item_group`,
`delete_predefined_item_group`, `create_note`, `update_note`, `delete_note`,
`create_note_attachment`, `create_call`, `update_call`, `delete_call`, `create_call_attachment`,
`create_meeting`, `update_meeting`, `delete_meeting`, `create_meeting_attachment`, `create_deal`,
`update_deal`, `delete_deal`, `create_deal_attachment`, `create_attachment`, `update_attachment`,
`delete_attachment`, `pin_attachment`, `unpin_attachment`, `create_relationship_type`,
`update_relationship_type`, `delete_relationship_type`, `create_action`, `update_action`,
`delete_action`, `unassign_action`, `mark_as_done_action`, `undo_completion_action`,
`promote_action`, `revert_promotion_action`, `swap_action`, `update_company`,
`create_company_linked_contact`, `delete_company_linked_contact`, `enable_company_synced_status`,
`delete_company_synced_status`, `delete_company_logo`, `create_contact`, `update_contact`,
`delete_contact`, `delete_contact_contact_photo`, `save_contact_to_google_contacts`,
`create_contact_action`, `create_contact_deal`, `create_contact_note`, `create_contact_call`,
`create_contact_meeting`, `create_contact_relationship`, `update_relationship`,
`delete_contact_relationship`, `assign_contact_tag`, and 12 more.

Service API documentation: https://developer.onepagecrm.com/.

## Auth setup

Connection fields:

- `action_id` (optional, string); action id for parameterized OnePageCRM streams.
- `attachment_id` (optional, string); attachment id for parameterized OnePageCRM streams.
- `base_url` (optional, string); default `https://app.onepagecrm.com/api/v3`; format `uri`;
  OnePageCRM API base URL override for tests or proxies.
- `call_id` (optional, string); call id for parameterized OnePageCRM streams.
- `company_field_id` (optional, string); company field id for parameterized OnePageCRM streams.
- `company_id` (optional, string); company id for parameterized OnePageCRM streams.
- `contact_id` (optional, string); contact id for parameterized OnePageCRM streams.
- `custom_field_id` (optional, string); custom field id for parameterized OnePageCRM streams.
- `deal_field_id` (optional, string); deal field id for parameterized OnePageCRM streams.
- `deal_id` (optional, string); deal id for parameterized OnePageCRM streams.
- `filter_id` (optional, string); filter id for parameterized OnePageCRM streams.
- `last_id` (optional, string); last id for parameterized OnePageCRM streams.
- `lead_source_id` (optional, string); lead source id for parameterized OnePageCRM streams.
- `meeting_id` (optional, string); meeting id for parameterized OnePageCRM streams.
- `mode` (optional, string).
- `note_id` (optional, string); note id for parameterized OnePageCRM streams.
- `notification_id` (optional, string); notification id for parameterized OnePageCRM streams.
- `owner_id` (optional, string); owner id for parameterized OnePageCRM streams.
- `password` (required, secret, string); OnePageCRM API key, sent as the HTTP Basic auth password.
  Never logged.
- `pipeline_id` (optional, string); pipeline id for parameterized OnePageCRM streams.
- `predefined_action_group_id` (optional, string); predefined action group id for parameterized
  OnePageCRM streams.
- `predefined_action_id` (optional, string); predefined action id for parameterized OnePageCRM
  streams.
- `predefined_item_group_id` (optional, string); predefined item group id for parameterized
  OnePageCRM streams.
- `predefined_item_id` (optional, string); predefined item id for parameterized OnePageCRM streams.
- `relationship_id` (optional, string); relationship id for parameterized OnePageCRM streams.
- `relationship_type_id` (optional, string); relationship type id for parameterized OnePageCRM
  streams.
- `status_id` (optional, string); status id for parameterized OnePageCRM streams.
- `tag_name` (optional, string); tag name for parameterized OnePageCRM streams.
- `user_id` (optional, string); user id for parameterized OnePageCRM streams.
- `username` (required, string); OnePageCRM API user ID, sent as the HTTP Basic auth username.
- `webhook_id` (optional, string); webhook id for parameterized OnePageCRM streams.

Secret fields are redacted in logs and write previews: `password`.

Default configuration values: `base_url=https://app.onepagecrm.com/api/v3`.

Authentication behavior:

- HTTP Basic authentication using `config.username`, `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/users`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 100.

Pagination by stream: none: `bootstrap`, `user`, `lead_source`, `status`, `deal_field`,
`custom_field`, `company_field`, `predefined_action`, `predefined_action_group`, `predefined_item`,
`predefined_item_group`, `note`, `call`, `meeting`, `deal`, `relationship_type`, `action`, `filter`,
`company`, `contact`, `contact_relationship`, `notification`, `webhook`, `pipeline`; page_number:
`contacts`, `deals`, `actions`, `companies`, `users`, `lead_sources`, `statuses`, `deal_fields`,
`custom_fields`, `company_fields`, `predefined_actions`, `predefined_action_groups`,
`predefined_items`, `predefined_item_groups`, `notes`, `calls`, `call_results`, `meetings`,
`relationship_types`, `countries`, `filters`, `company_actions`, `company_deals`, `company_notes`,
`company_calls`, `company_meetings`, `company_linked_contacts`, `company_pinned_attachments`,
`filtered_contacts`, `contact_actions`, `contact_deals`, `contact_notes`, `contact_calls`,
`contact_meetings`, `contact_relationships`, `contact_pinned_attachments`, `contacts_cascade`,
`contacts_cascade_after`, `action_stream`, `team_stream`, `notifications`, `webhooks`, `pipelines`.

- `contacts`: GET `/contacts` - records path `data.contacts`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; computed output fields
  `company_name`, `created_at`, `first_name`, `id`, `job_title`, `last_name`, `owner_id`, `starred`,
  `status_id`, `updated_at`.
- `deals`: GET `/deals` - records path `data.deals`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100; computed output fields `amount`,
  `contact_id`, `created_at`, `currency`, `expected_close_date`, `id`, `name`, `owner_id`, `stage`,
  `status`, `updated_at`.
- `actions`: GET `/actions` - records path `data.actions`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; computed output fields
  `assignee_id`, `contact_id`, `created_at`, `date`, `done`, `id`, `status`, `text`, `updated_at`.
- `companies`: GET `/companies` - records path `data.companies`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; computed output fields
  `created_at`, `description`, `id`, `name`, `phone`, `updated_at`, `url`.
- `users`: GET `/users` - records path `data`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 100; computed output fields `email`, `first_name`,
  `id`, `last_name`, `role`, `status`.
- `bootstrap`: GET `/bootstrap` - single-object response; records path `data`; computed output
  fields `id`; emits passthrough records.
- `user`: GET `/users/{{ config.user_id }}` - single-object response; records path `data`; computed
  output fields `id`; emits passthrough records.
- `lead_sources`: GET `/lead_sources` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; computed output fields `id`; emits
  passthrough records.
- `lead_source`: GET `/lead_sources/{{ config.lead_source_id }}` - single-object response; records
  path `data`; computed output fields `id`; emits passthrough records.
- `statuses`: GET `/statuses` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100; computed output fields `id`; emits
  passthrough records.
- `status`: GET `/statuses/{{ config.status_id }}` - single-object response; records path `data`;
  computed output fields `id`; emits passthrough records.
- `deal_fields`: GET `/deal_fields` - records path `data.deal_fields`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; computed output fields
  `id`; emits passthrough records.
- `deal_field`: GET `/deal_fields/{{ config.deal_field_id }}` - single-object response; records path
  `data`; computed output fields `id`; emits passthrough records.
- `custom_fields`: GET `/custom_fields` - records path `data.custom_fields`; page-number pagination;
  page parameter `page`; size parameter `per_page`; starts at 1; page size 100; computed output
  fields `id`; emits passthrough records.
- `custom_field`: GET `/custom_fields/{{ config.custom_field_id }}` - single-object response;
  records path `data`; computed output fields `id`; emits passthrough records.
- `company_fields`: GET `/company_fields` - records path `data.company_fields`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; computed
  output fields `id`; emits passthrough records.
- `company_field`: GET `/company_fields/{{ config.company_field_id }}` - single-object response;
  records path `data`; computed output fields `id`; emits passthrough records.
- `predefined_actions`: GET `/predefined_actions` - records path `data.predefined_actions`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `id`; emits passthrough records.
- `predefined_action`: GET `/predefined_actions/{{ config.predefined_action_id }}` - single-object
  response; records path `data`; computed output fields `id`; emits passthrough records.
- `predefined_action_groups`: GET `/predefined_action_groups` - records path
  `data.predefined_action_groups`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 100; computed output fields `id`; emits passthrough records.
- `predefined_action_group`: GET `/predefined_action_groups/{{ config.predefined_action_group_id }}`
  - single-object response; records path `data`; computed output fields `id`; emits passthrough
  records.
- `predefined_items`: GET `/predefined_items` - records path `data.predefined_items`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; computed
  output fields `id`; emits passthrough records.
- `predefined_item`: GET `/predefined_items/{{ config.predefined_item_id }}` - single-object
  response; records path `data`; computed output fields `id`; emits passthrough records.
- `predefined_item_groups`: GET `/predefined_item_groups` - records path
  `data.predefined_item_groups`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 100; computed output fields `id`; emits passthrough records.
- `predefined_item_group`: GET `/predefined_item_groups/{{ config.predefined_item_group_id }}` -
  single-object response; records path `data`; computed output fields `id`; emits passthrough
  records.
- `notes`: GET `/notes` - records path `data.notes`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100; computed output fields `id`; emits
  passthrough records.
- `note`: GET `/notes/{{ config.note_id }}` - single-object response; records path `data`; computed
  output fields `id`; emits passthrough records.
- `calls`: GET `/calls` - records path `data.calls`; page-number pagination; page parameter `page`;
  size parameter `per_page`; starts at 1; page size 100; computed output fields `id`; emits
  passthrough records.
- `call`: GET `/calls/{{ config.call_id }}` - single-object response; records path `data`; computed
  output fields `id`; emits passthrough records.
- `call_results`: GET `/call_results` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; computed output fields `id`; emits
  passthrough records.
- `meetings`: GET `/meetings` - records path `data.meetings`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; computed output fields `id`; emits
  passthrough records.
- `meeting`: GET `/meetings/{{ config.meeting_id }}` - single-object response; records path `data`;
  computed output fields `id`; emits passthrough records.
- `deal`: GET `/deals/{{ config.deal_id }}` - single-object response; records path `data`; computed
  output fields `id`; emits passthrough records.
- `relationship_types`: GET `/relationship_types` - records path `data.relationship_types`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `id`; emits passthrough records.
- `relationship_type`: GET `/relationship_types/{{ config.relationship_type_id }}` - single-object
  response; records path `data`; computed output fields `id`; emits passthrough records.
- `countries`: GET `/countries` - records path `data.countries`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; computed output fields
  `id`; emits passthrough records.
- `action`: GET `/actions/{{ config.action_id }}` - single-object response; records path `data`;
  computed output fields `id`; emits passthrough records.
- `filters`: GET `/filters` - records path `data.filters`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; computed output fields `id`; emits
  passthrough records.
- `filter`: GET `/filters/{{ config.filter_id }}` - single-object response; records path `data`;
  computed output fields `id`; emits passthrough records.
- `company`: GET `/companies/{{ config.company_id }}` - single-object response; records path `data`;
  computed output fields `id`; emits passthrough records.
- `company_actions`: GET `/companies/{{ config.company_id }}/actions` - records path `data.actions`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `id`; emits passthrough records.
- `company_deals`: GET `/companies/{{ config.company_id }}/deals` - records path `data.deals`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `id`; emits passthrough records.
- `company_notes`: GET `/companies/{{ config.company_id }}/notes` - records path `data.notes`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `id`; emits passthrough records.
- `company_calls`: GET `/companies/{{ config.company_id }}/calls` - records path `data.calls`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `id`; emits passthrough records.
- `company_meetings`: GET `/companies/{{ config.company_id }}/meetings` - records path
  `data.meetings`; page-number pagination; page parameter `page`; size parameter `per_page`; starts
  at 1; page size 100; computed output fields `id`; emits passthrough records.
- `company_linked_contacts`: GET `/companies/{{ config.company_id }}/linked_contacts` - records path
  `data.linked_contacts`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 100; computed output fields `id`; emits passthrough records.
- `company_pinned_attachments`: GET `/companies/{{ config.company_id }}/pinned_attachments` -
  records path `data.pinned_attachments`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 100; computed output fields `id`; emits passthrough
  records.
- `contact`: GET `/contacts/{{ config.contact_id }}` - single-object response; records path `data`;
  computed output fields `id`; emits passthrough records.
- `filtered_contacts`: GET `/contacts/filters/{{ config.filter_id }}` - records path
  `data.contacts`; page-number pagination; page parameter `page`; size parameter `per_page`; starts
  at 1; page size 100; computed output fields `id`; emits passthrough records.
- `contact_actions`: GET `/contacts/{{ config.contact_id }}/actions` - records path `data.actions`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `id`; emits passthrough records.
- `contact_deals`: GET `/contacts/{{ config.contact_id }}/deals` - records path `data.deals`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `id`; emits passthrough records.
- `contact_notes`: GET `/contacts/{{ config.contact_id }}/notes` - records path `data.notes`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `id`; emits passthrough records.
- `contact_calls`: GET `/contacts/{{ config.contact_id }}/calls` - records path `data.calls`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  100; computed output fields `id`; emits passthrough records.
- `contact_meetings`: GET `/contacts/{{ config.contact_id }}/meetings` - records path
  `data.meetings`; page-number pagination; page parameter `page`; size parameter `per_page`; starts
  at 1; page size 100; computed output fields `id`; emits passthrough records.
- `contact_relationships`: GET `/contacts/{{ config.contact_id }}/relationships` - records path
  `data.relationships`; page-number pagination; page parameter `page`; size parameter `per_page`;
  starts at 1; page size 100; computed output fields `id`; emits passthrough records.
- `contact_relationship`: GET `/contacts/{{ config.contact_id }}/relationships/{{
  config.relationship_id }}` - single-object response; records path `data`; computed output fields
  `id`; emits passthrough records.
- `contact_pinned_attachments`: GET `/contacts/{{ config.contact_id }}/pinned_attachments` - records
  path `data.pinned_attachments`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 100; computed output fields `id`; emits passthrough records.
- `contacts_cascade`: GET `/contacts/cascade` - records path `data.contacts`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 100; computed
  output fields `id`; emits passthrough records.
- `contacts_cascade_after`: GET `/contacts/cascade/{{ config.last_id }}` - records path
  `data.contacts`; page-number pagination; page parameter `page`; size parameter `per_page`; starts
  at 1; page size 100; computed output fields `id`; emits passthrough records.
- `action_stream`: GET `/action_stream` - records path `data.contacts`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; computed output fields
  `id`; emits passthrough records.
- `team_stream`: GET `/team_stream` - records path `data.contacts`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; computed output fields
  `id`; emits passthrough records.
- `notifications`: GET `/notifications` - records path `data.notifications`; page-number pagination;
  page parameter `page`; size parameter `per_page`; starts at 1; page size 100; computed output
  fields `id`; emits passthrough records.
- `notification`: GET `/notifications/{{ config.notification_id }}` - single-object response;
  records path `data`; computed output fields `id`; emits passthrough records.
- `webhooks`: GET `/webhooks` - records path `data.webhooks`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 100; computed output fields `id`; emits
  passthrough records.
- `webhook`: GET `/webhooks/{{ config.webhook_id }}` - single-object response; records path `data`;
  computed output fields `id`; emits passthrough records.
- `pipelines`: GET `/pipelines` - records path `data.pipelines`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 100; computed output fields
  `id`; emits passthrough records.
- `pipeline`: GET `/pipelines/{{ config.pipeline_id }}` - single-object response; records path
  `data`; computed output fields `id`; emits passthrough records.

## Write actions & risks

Overall write risk: external OnePageCRM API mutations can create, update, complete, tag, export,
disable, or delete live CRM records and account configuration.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `update_user`: PUT `/users/{{ record.user_id }}` - kind `update`; body type `json`; path fields
  `user_id`; required record fields `user_id`, `first_name`; accepted fields `bcc_email`,
  `company_name`, `first_name`, `last_name`, `user_id`; risk: Update a specific user; external
  OnePageCRM mutation, approval required.
- `create_lead_source`: POST `/lead_sources` - kind `create`; body type `json`; required record
  fields `id`; accepted fields `counts`, `id`, `text`, `total_count`; risk: Create a new lead
  source; external OnePageCRM mutation, approval required.
- `update_lead_source`: PUT `/lead_sources/{{ record.lead_source_id }}` - kind `update`; body type
  `json`; path fields `lead_source_id`; required record fields `lead_source_id`, `id`; accepted
  fields `counts`, `id`, `lead_source_id`, `text`, `total_count`; risk: Update a specific lead
  source; external OnePageCRM mutation, approval required.
- `delete_lead_source`: DELETE `/lead_sources/{{ record.lead_source_id }}` - kind `delete`; body
  type `none`; path fields `lead_source_id`; required record fields `lead_source_id`; accepted
  fields `lead_source_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Delete a specific lead source; external OnePageCRM mutation, approval
  required.
- `create_status`: POST `/statuses` - kind `create`; body type `json`; required record fields `id`;
  accepted fields `description`, `id`, `status`, `text`; risk: Create a new status; external
  OnePageCRM mutation, approval required.
- `update_status`: PUT `/statuses/{{ record.status_id }}` - kind `update`; body type `json`; path
  fields `status_id`; required record fields `status_id`, `id`; accepted fields `description`, `id`,
  `status`, `status_id`, `text`; risk: Update a specific status; external OnePageCRM mutation,
  approval required.
- `delete_status`: DELETE `/statuses/{{ record.status_id }}` - kind `delete`; body type `none`; path
  fields `status_id`; required record fields `status_id`; accepted fields `status_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: Delete a specific
  status; external OnePageCRM mutation, approval required.
- `create_deal_field`: POST `/deal_fields` - kind `create`; body type `json`; required record fields
  `id`; accepted fields `id`, `name`, `position`, `type`; risk: Create a new deal field; external
  OnePageCRM mutation, approval required.
- `update_deal_field`: PUT `/deal_fields/{{ record.deal_field_id }}` - kind `update`; body type
  `json`; path fields `deal_field_id`; required record fields `deal_field_id`, `id`; accepted fields
  `deal_field_id`, `id`, `name`, `position`, `type`; risk: Update a specific deal field; external
  OnePageCRM mutation, approval required.
- `delete_deal_field`: DELETE `/deal_fields/{{ record.deal_field_id }}` - kind `delete`; body type
  `none`; path fields `deal_field_id`; required record fields `deal_field_id`; accepted fields
  `deal_field_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Delete a specific deal field; external OnePageCRM mutation, approval required.
- `create_custom_field`: POST `/custom_fields` - kind `create`; body type `json`; required record
  fields `id`; accepted fields `id`, `name`, `position`, `type`; risk: Create a new custom field;
  external OnePageCRM mutation, approval required.
- `update_custom_field`: PUT `/custom_fields/{{ record.custom_field_id }}` - kind `update`; body
  type `json`; path fields `custom_field_id`; required record fields `custom_field_id`, `id`;
  accepted fields `custom_field_id`, `id`, `name`, `position`, `type`; risk: Update a specific
  custom field; external OnePageCRM mutation, approval required.
- `delete_custom_field`: DELETE `/custom_fields/{{ record.custom_field_id }}` - kind `delete`; body
  type `none`; path fields `custom_field_id`; required record fields `custom_field_id`; accepted
  fields `custom_field_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Delete a specific custom field; external OnePageCRM mutation, approval
  required.
- `create_company_field`: POST `/company_fields` - kind `create`; body type `json`; required record
  fields `id`; accepted fields `id`, `name`, `position`, `type`; risk: Create a new company field;
  external OnePageCRM mutation, approval required.
- `update_company_field`: PUT `/company_fields/{{ record.company_field_id }}` - kind `update`; body
  type `json`; path fields `company_field_id`; required record fields `company_field_id`, `id`;
  accepted fields `company_field_id`, `id`, `name`, `position`, `type`; risk: Update a specific
  company field; external OnePageCRM mutation, approval required.
- `delete_company_field`: DELETE `/company_fields/{{ record.company_field_id }}` - kind `delete`;
  body type `none`; path fields `company_field_id`; required record fields `company_field_id`;
  accepted fields `company_field_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Delete a specific company field; external OnePageCRM mutation,
  approval required.
- `create_predefined_action`: POST `/predefined_actions` - kind `create`; body type `json`; required
  record fields `id`; accepted fields `days`, `id`, `text`; risk: Create a new predefined action;
  external OnePageCRM mutation, approval required.
- `update_predefined_action`: PUT `/predefined_actions/{{ record.predefined_action_id }}` - kind
  `update`; body type `json`; path fields `predefined_action_id`; required record fields
  `predefined_action_id`, `id`; accepted fields `days`, `id`, `predefined_action_id`, `text`; risk:
  Update a specific predefined action; external OnePageCRM mutation, approval required.
- `delete_predefined_action`: DELETE `/predefined_actions/{{ record.predefined_action_id }}` - kind
  `delete`; body type `none`; path fields `predefined_action_id`; required record fields
  `predefined_action_id`; accepted fields `predefined_action_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: Delete a specific predefined action; external
  OnePageCRM mutation, approval required.
- `create_predefined_action_group`: POST `/predefined_action_groups` - kind `create`; body type
  `json`; required record fields `text`; accepted fields `action_ids`, `position`, `text`; risk:
  Create a new predefined action group; external OnePageCRM mutation, approval required.
- `update_predefined_action_group`: PUT `/predefined_action_groups/{{
  record.predefined_action_group_id }}` - kind `update`; body type `json`; path fields
  `predefined_action_group_id`; required record fields `predefined_action_group_id`, `text`;
  accepted fields `action_ids`, `position`, `predefined_action_group_id`, `text`; risk: Update a
  specific predefined action group; external OnePageCRM mutation, approval required.
- `delete_predefined_action_group`: DELETE `/predefined_action_groups/{{
  record.predefined_action_group_id }}` - kind `delete`; body type `none`; path fields
  `predefined_action_group_id`; required record fields `predefined_action_group_id`; accepted fields
  `predefined_action_group_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Delete a specific predefined action group; external OnePageCRM mutation,
  approval required.
- `create_predefined_item`: POST `/predefined_items` - kind `create`; body type `json`; required
  record fields `name`; accepted fields `cost`, `description`, `name`, `price`; risk: Create a new
  predefined item; external OnePageCRM mutation, approval required.
- `update_predefined_item`: PUT `/predefined_items/{{ record.predefined_item_id }}` - kind `update`;
  body type `json`; path fields `predefined_item_id`; required record fields `predefined_item_id`,
  `name`; accepted fields `cost`, `description`, `name`, `predefined_item_id`, `price`; risk: Update
  a specific predefined item; external OnePageCRM mutation, approval required.
- `delete_predefined_item`: DELETE `/predefined_items/{{ record.predefined_item_id }}` - kind
  `delete`; body type `none`; path fields `predefined_item_id`; required record fields
  `predefined_item_id`; accepted fields `predefined_item_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: Delete a specific predefined item; external
  OnePageCRM mutation, approval required.
- `create_predefined_item_group`: POST `/predefined_item_groups` - kind `create`; body type `json`;
  required record fields `name`; accepted fields `deal_items`, `name`; risk: Create a new predefined
  item group; external OnePageCRM mutation, approval required.
- `delete_predefined_item_group`: DELETE `/predefined_item_groups/{{ record.predefined_item_group_id
  }}` - kind `delete`; body type `none`; path fields `predefined_item_group_id`; required record
  fields `predefined_item_group_id`; accepted fields `predefined_item_group_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Delete a specific
  predefined item group; external OnePageCRM mutation, approval required.
- `create_note`: POST `/notes` - kind `create`; body type `json`; required record fields
  `contact_id`; accepted fields `contact_id`, `date`, `linked_deal_id`, `text`; risk: Create a new
  note; external OnePageCRM mutation, approval required.
- `update_note`: PUT `/notes/{{ record.note_id }}` - kind `update`; body type `json`; path fields
  `note_id`; required record fields `note_id`, `contact_id`; accepted fields `contact_id`, `date`,
  `linked_deal_id`, `note_id`, `text`; risk: Update a specific note; external OnePageCRM mutation,
  approval required.
- `delete_note`: DELETE `/notes/{{ record.note_id }}` - kind `delete`; body type `none`; path fields
  `note_id`; required record fields `note_id`; accepted fields `note_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Delete a specific note; external
  OnePageCRM mutation, approval required.
- `create_note_attachment`: POST `/notes/{{ record.note_id }}/attachments` - kind `create`; body
  type `json`; path fields `note_id`; required record fields `note_id`, `reference_id`; accepted
  fields `contact_id`, `name`, `note_id`, `reference_id`, `reference_type`; risk: Create attachment
  and assign it to an existing note; external OnePageCRM mutation, approval required.
- `create_call`: POST `/calls` - kind `create`; body type `json`; required record fields
  `contact_id`; accepted fields `call_time_int`, `contact_id`, `phone_number`, `text`; risk: Create
  a call; external OnePageCRM mutation, approval required.
- `update_call`: PUT `/calls/{{ record.call_id }}` - kind `update`; body type `json`; path fields
  `call_id`; required record fields `call_id`, `contact_id`; accepted fields `call_id`,
  `call_time_int`, `contact_id`, `phone_number`, `text`; risk: Update a specific call; external
  OnePageCRM mutation, approval required.
- `delete_call`: DELETE `/calls/{{ record.call_id }}` - kind `delete`; body type `none`; path fields
  `call_id`; required record fields `call_id`; accepted fields `call_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Delete a specific call; external
  OnePageCRM mutation, approval required.
- `create_call_attachment`: POST `/calls/{{ record.call_id }}/attachments` - kind `create`; body
  type `json`; path fields `call_id`; required record fields `call_id`, `reference_id`; accepted
  fields `call_id`, `contact_id`, `name`, `reference_id`, `reference_type`; risk: Create attachment
  and assign it to an existing call; external OnePageCRM mutation, approval required.
- `create_meeting`: POST `/meetings` - kind `create`; body type `json`; required record fields
  `contact_id`; accepted fields `contact_id`, `meeting_time_int`, `place`, `text`; risk: Create a
  meeting; external OnePageCRM mutation, approval required.
- `update_meeting`: PUT `/meetings/{{ record.meeting_id }}` - kind `update`; body type `json`; path
  fields `meeting_id`; required record fields `meeting_id`, `contact_id`; accepted fields
  `contact_id`, `meeting_id`, `meeting_time_int`, `place`, `text`; risk: Update a specific meeting;
  external OnePageCRM mutation, approval required.
- `delete_meeting`: DELETE `/meetings/{{ record.meeting_id }}` - kind `delete`; body type `none`;
  path fields `meeting_id`; required record fields `meeting_id`; accepted fields `meeting_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Delete a
  specific meeting; external OnePageCRM mutation, approval required.
- `create_meeting_attachment`: POST `/meetings/{{ record.meeting_id }}/attachments` - kind `create`;
  body type `json`; path fields `meeting_id`; required record fields `meeting_id`, `reference_id`;
  accepted fields `contact_id`, `meeting_id`, `name`, `reference_id`, `reference_type`; risk: Create
  attachment and assign it to an existing meeting; external OnePageCRM mutation, approval required.
- `create_deal`: POST `/deals` - kind `create`; body type `json`; required record fields
  `contact_id`; accepted fields `contact_id`, `owner_id`, `pipeline_id`, `sales_pipeline_id`; risk:
  Create a new deal; external OnePageCRM mutation, approval required.
- `update_deal`: PUT `/deals/{{ record.deal_id }}` - kind `update`; body type `json`; path fields
  `deal_id`; required record fields `deal_id`, `contact_id`; accepted fields `contact_id`,
  `deal_id`, `owner_id`, `pipeline_id`, `sales_pipeline_id`; risk: Update a specific deal; external
  OnePageCRM mutation, approval required.
- `delete_deal`: DELETE `/deals/{{ record.deal_id }}` - kind `delete`; body type `none`; path fields
  `deal_id`; required record fields `deal_id`; accepted fields `deal_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Delete a specific deal; external
  OnePageCRM mutation, approval required.
- `create_deal_attachment`: POST `/deals/{{ record.deal_id }}/attachments` - kind `create`; body
  type `json`; path fields `deal_id`; required record fields `deal_id`, `reference_id`; accepted
  fields `contact_id`, `deal_id`, `name`, `reference_id`, `reference_type`; risk: Create attachment
  and assign it to an existing deal; external OnePageCRM mutation, approval required.
- `create_attachment`: POST `/attachments` - kind `create`; body type `json`; required record fields
  `reference_id`; accepted fields `contact_id`, `name`, `reference_id`, `reference_type`; risk:
  Create a new attachment; external OnePageCRM mutation, approval required.
- `update_attachment`: PATCH `/attachments/{{ record.attachment_id }}` - kind `update`; body type
  `json`; path fields `attachment_id`; required record fields `attachment_id`, `attachment`;
  accepted fields `attachment`, `attachment_id`; risk: Sets/updates attachment custom file name;
  external OnePageCRM mutation, approval required.
- `delete_attachment`: DELETE `/attachments/{{ record.attachment_id }}` - kind `delete`; body type
  `none`; path fields `attachment_id`; required record fields `attachment_id`; accepted fields
  `attachment_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Delete a specific attachment; external OnePageCRM mutation, approval required.
- `pin_attachment`: PATCH `/attachments/{{ record.attachment_id }}/pin` - kind `update`; body type
  `none`; path fields `attachment_id`; required record fields `attachment_id`; accepted fields
  `attachment_id`; risk: Pin attachment to its owner contact through its note/call/deal; external
  OnePageCRM mutation, approval required.
- `unpin_attachment`: PATCH `/attachments/{{ record.attachment_id }}/unpin` - kind `update`; body
  type `none`; path fields `attachment_id`; required record fields `attachment_id`; accepted fields
  `attachment_id`; risk: Unpin attachment from its owner contact through its note/call/deal;
  external OnePageCRM mutation, approval required.
- `create_relationship_type`: POST `/relationship_types` - kind `create`; body type `json`; required
  record fields `variants`; accepted fields `variants`; risk: Create a new relationship type;
  external OnePageCRM mutation, approval required.
- `update_relationship_type`: PUT `/relationship_types/{{ record.relationship_type_id }}` - kind
  `update`; body type `json`; path fields `relationship_type_id`; required record fields
  `relationship_type_id`, `variants`; accepted fields `relationship_type_id`, `variants`; risk:
  Update a specific relationship type; external OnePageCRM mutation, approval required.
- `delete_relationship_type`: DELETE `/relationship_types/{{ record.relationship_type_id }}` - kind
  `delete`; body type `none`; path fields `relationship_type_id`; required record fields
  `relationship_type_id`; accepted fields `relationship_type_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: Delete a relationship type; external
  OnePageCRM mutation, approval required.
- `create_action`: POST `/actions` - kind `create`; body type `json`; required record fields
  `contact_id`; accepted fields `assignee_id`, `contact_id`, `status`, `text`; risk: Create a new
  action; external OnePageCRM mutation, approval required.
- `update_action`: PUT `/actions/{{ record.action_id }}` - kind `update`; body type `json`; path
  fields `action_id`; required record fields `action_id`, `contact_id`; accepted fields `action_id`,
  `assignee_id`, `contact_id`, `status`, `text`; risk: Update a specific action; external OnePageCRM
  mutation, approval required.
- `delete_action`: DELETE `/actions/{{ record.action_id }}` - kind `delete`; body type `none`; path
  fields `action_id`; required record fields `action_id`; accepted fields `action_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: Delete a specific
  action; external OnePageCRM mutation, approval required.
- `unassign_action`: PUT `/actions/{{ record.action_id }}/unassign` - kind `update`; body type
  `none`; path fields `action_id`; required record fields `action_id`; accepted fields `action_id`;
  risk: Unassign a specific action (from the currently assigned user); external OnePageCRM mutation,
  approval required.
- `mark_as_done_action`: PUT `/actions/{{ record.action_id }}/mark_as_done` - kind `update`; body
  type `none`; path fields `action_id`; required record fields `action_id`; accepted fields
  `action_id`; risk: Mark a specific action as complete; external OnePageCRM mutation, approval
  required.
- `undo_completion_action`: PUT `/actions/{{ record.action_id }}/undo_completion` - kind `update`;
  body type `none`; path fields `action_id`; required record fields `action_id`; accepted fields
  `action_id`; risk: Undo action completion; external OnePageCRM mutation, approval required.
- `promote_action`: PUT `/actions/{{ record.action_id }}/promote` - kind `update`; body type `none`;
  path fields `action_id`; required record fields `action_id`; accepted fields `action_id`; risk:
  Specify action to be promoted as the logged API users next action; external OnePageCRM mutation,
  approval required.
- `revert_promotion_action`: PUT `/actions/{{ record.action_id }}/revert_promotion` - kind `update`;
  body type `none`; path fields `action_id`; required record fields `action_id`; accepted fields
  `action_id`; risk: Undo action promotion; external OnePageCRM mutation, approval required.
- `swap_action`: PUT `/actions/{{ record.action_id }}/swap` - kind `update`; body type `none`; path
  fields `action_id`; required record fields `action_id`; accepted fields `action_id`; risk: Specify
  action to be swapped in as the logged API users next action; external OnePageCRM mutation,
  approval required.
- `update_company`: PUT `/companies/{{ record.company_id }}` - kind `update`; body type `json`; path
  fields `company_id`; required record fields `company_id`, `name`; accepted fields `company_id`,
  `description`, `name`, `phone`, `url`; risk: Update a specific company; external OnePageCRM
  mutation, approval required.
- `create_company_linked_contact`: POST `/companies/{{ record.company_id }}/linked_contacts` - kind
  `create`; body type `json`; path fields `company_id`; required record fields `company_id`,
  `contact_id`; accepted fields `company_id`, `contact_id`; risk: Link a contact to a specific
  company; external OnePageCRM mutation, approval required.
- `delete_company_linked_contact`: DELETE `/companies/{{ record.company_id }}/linked_contacts/{{
  record.contact_id }}` - kind `delete`; body type `none`; path fields `company_id`, `contact_id`;
  required record fields `company_id`, `contact_id`; accepted fields `company_id`, `contact_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Unlink a
  contact from a company; external OnePageCRM mutation, approval required.
- `enable_company_synced_status`: POST `/companies/{{ record.company_id }}/synced_status` - kind
  `create`; body type `json`; path fields `company_id`; required record fields `company_id`,
  `status_id`; accepted fields `company_id`, `status_id`; risk: Enable company status sync; external
  OnePageCRM mutation, approval required.
- `delete_company_synced_status`: DELETE `/companies/{{ record.company_id }}/synced_status` - kind
  `delete`; body type `none`; path fields `company_id`; required record fields `company_id`;
  accepted fields `company_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Disable company status sync; external OnePageCRM mutation, approval required.
- `delete_company_logo`: DELETE `/companies/{{ record.company_id }}/logo` - kind `delete`; body type
  `none`; path fields `company_id`; required record fields `company_id`; accepted fields
  `company_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Delete logo in then given company; external OnePageCRM mutation, approval required.
- `create_contact`: POST `/contacts` - kind `create`; body type `json`; required record fields
  `title`; accepted fields `first_name`, `job_title`, `last_name`, `title`; risk: Create a contact;
  external OnePageCRM mutation, approval required.
- `update_contact`: PUT `/contacts/{{ record.contact_id }}` - kind `update`; body type `json`; path
  fields `contact_id`; required record fields `contact_id`, `title`; accepted fields `contact_id`,
  `first_name`, `job_title`, `last_name`, `title`; risk: Update a specific contact; external
  OnePageCRM mutation, approval required.
- `delete_contact`: DELETE `/contacts/{{ record.contact_id }}` - kind `delete`; body type `none`;
  path fields `contact_id`; required record fields `contact_id`; accepted fields `contact_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Delete a
  specific contact; external OnePageCRM mutation, approval required.
- `delete_contact_contact_photo`: DELETE `/contacts/{{ record.contact_id }}/contact_photo` - kind
  `delete`; body type `none`; path fields `contact_id`; required record fields `contact_id`;
  accepted fields `contact_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Remove a contact's photo; external OnePageCRM mutation, approval required.
- `save_contact_to_google_contacts`: POST `/contacts/{{ record.contact_id }}/google_contacts` - kind
  `create`; body type `none`; path fields `contact_id`; required record fields `contact_id`;
  accepted fields `contact_id`; risk: Save a specific OnePageCRM contact to Google Contacts;
  external OnePageCRM mutation, approval required.
- `create_contact_action`: POST `/contacts/{{ record.contact_id }}/actions` - kind `create`; body
  type `json`; path fields `contact_id`; required record fields `contact_id`, `assignee_id`;
  accepted fields `assignee_id`, `contact_id`, `date`, `status`, `text`; risk: Create an action for
  a specific contact; external OnePageCRM mutation, approval required.
- `create_contact_deal`: POST `/contacts/{{ record.contact_id }}/deals` - kind `create`; body type
  `json`; path fields `contact_id`; required record fields `contact_id`, `owner_id`; accepted fields
  `contact_id`, `name`, `owner_id`, `stage`, `text`; risk: Create a deal for a specific contact;
  external OnePageCRM mutation, approval required.
- `create_contact_note`: POST `/contacts/{{ record.contact_id }}/notes` - kind `create`; body type
  `json`; path fields `contact_id`; required record fields `contact_id`, `text`; accepted fields
  `contact_id`, `date`, `linked_deal_id`, `text`, `user_ids_to_notify`; risk: Create a note for a
  specific contact; external OnePageCRM mutation, approval required.
- `create_contact_call`: POST `/contacts/{{ record.contact_id }}/calls` - kind `create`; body type
  `json`; path fields `contact_id`; required record fields `contact_id`, `call_time_int`; accepted
  fields `call_result`, `call_time_int`, `contact_id`, `phone_number`, `text`; risk: Create a call
  for a specific contact; external OnePageCRM mutation, approval required.
- `create_contact_meeting`: POST `/contacts/{{ record.contact_id }}/meetings` - kind `create`; body
  type `json`; path fields `contact_id`; required record fields `contact_id`, `meeting_time_int`;
  accepted fields `contact_id`, `meeting_time_int`, `place`, `text`, `user_ids_to_notify`; risk:
  Create a meeting for a specific contact; external OnePageCRM mutation, approval required.
- `create_contact_relationship`: POST `/contacts/{{ record.contact_id }}/relationships` - kind
  `create`; body type `json`; path fields `contact_id`; required record fields `contact_id`,
  `relationship_type_id`; accepted fields `contact_id`, `related_contacts`, `relationship_type_id`;
  risk: Create a relationships for a specific contact; external OnePageCRM mutation, approval
  required.
- `update_relationship`: PUT `/contacts/{{ record.contact_id }}/relationships/{{
  record.relationship_id }}` - kind `update`; body type `json`; path fields `contact_id`,
  `relationship_id`; required record fields `contact_id`, `relationship_id`, `relationship_type_id`;
  accepted fields `contact_id`, `related_contacts`, `relationship_id`, `relationship_type_id`; risk:
  Update a specific relationship; external OnePageCRM mutation, approval required.
- `delete_contact_relationship`: DELETE `/contacts/{{ record.contact_id }}/relationships/{{
  record.relationship_id }}` - kind `delete`; body type `none`; path fields `contact_id`,
  `relationship_id`; required record fields `contact_id`, `relationship_id`; accepted fields
  `contact_id`, `relationship_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Delete a relationship; external OnePageCRM mutation, approval required.
- `assign_contact_tag`: PUT `/contacts/{{ record.contact_id }}/assign_tag/{{ record.tag_name }}` -
  kind `update`; body type `none`; path fields `contact_id`, `tag_name`; required record fields
  `contact_id`, `tag_name`; accepted fields `contact_id`, `tag_name`; risk: Assign a tag to a
  specific contact; external OnePageCRM mutation, approval required.
- `unassign_contact_tag`: PUT `/contacts/{{ record.contact_id }}/unassign_tag/{{ record.tag_name }}`
  - kind `update`; body type `none`; path fields `contact_id`, `tag_name`; required record fields
  `contact_id`, `tag_name`; accepted fields `contact_id`, `tag_name`; risk: Remove a tag from a
  specific contact; external OnePageCRM mutation, approval required.
- `change_contact_status`: PUT `/contacts/{{ record.contact_id }}/change_status/{{ record.status_id
  }}` - kind `update`; body type `none`; path fields `contact_id`, `status_id`; required record
  fields `contact_id`, `status_id`; accepted fields `contact_id`, `status_id`; risk: Change the
  status of a specific contact; external OnePageCRM mutation, approval required.
- `change_contact_owner`: PUT `/contacts/{{ record.contact_id }}/change_owner/{{ record.owner_id }}`
  - kind `update`; body type `none`; path fields `contact_id`, `owner_id`; required record fields
  `contact_id`, `owner_id`; accepted fields `contact_id`, `owner_id`; risk: Change the owner of a
  specific contact; external OnePageCRM mutation, approval required.
- `star_contact`: PUT `/contacts/{{ record.contact_id }}/star` - kind `update`; body type `none`;
  path fields `contact_id`; required record fields `contact_id`; accepted fields `contact_id`; risk:
  Apply a star to a specific contact; external OnePageCRM mutation, approval required.
- `unstar_contact`: PUT `/contacts/{{ record.contact_id }}/unstar` - kind `update`; body type
  `none`; path fields `contact_id`; required record fields `contact_id`; accepted fields
  `contact_id`; risk: Remove star from a specific contact; external OnePageCRM mutation, approval
  required.
- `close_sales_cycle_contact`: PUT `/contacts/{{ record.contact_id }}/close_sales_cycle` - kind
  `update`; body type `json`; path fields `contact_id`; required record fields `contact_id`,
  `comment`; accepted fields `comment`, `contact_id`; risk: Close the sales cycle for a specific
  contact; external OnePageCRM mutation, approval required.
- `force_close_sales_cycle_contact`: PUT `/contacts/{{ record.contact_id }}/force_close_sales_cycle`
  - kind `update`; body type `json`; path fields `contact_id`; required record fields `contact_id`,
  `comment`; accepted fields `comment`, `contact_id`; risk: Force close the sales cycle for a
  specific contact; external OnePageCRM mutation, approval required.
- `reopen_sales_cycle_contact`: PUT `/contacts/{{ record.contact_id }}/reopen_sales_cycle` - kind
  `update`; body type `none`; path fields `contact_id`; required record fields `contact_id`;
  accepted fields `contact_id`; risk: Reopen the sales cycle for a specific contact; external
  OnePageCRM mutation, approval required.
- `split_contact`: PUT `/contacts/{{ record.contact_id }}/split` - kind `update`; body type `json`;
  path fields `contact_id`; required record fields `contact_id`, `company_name`; accepted fields
  `company_name`, `contact_id`; risk: Split a contact from their current company (and potentially to
  a new company); external OnePageCRM mutation, approval required.
- `mark_as_read_notification`: POST `/notifications/{{ record.notification_id }}/mark_as_read` -
  kind `create`; body type `none`; path fields `notification_id`; required record fields
  `notification_id`; accepted fields `notification_id`; risk: Marks given notification as read;
  external OnePageCRM mutation, approval required.
- `mark_all_notifications_as_read`: POST `/notifications/mark_all_as_read` - kind `create`; body
  type `none`; risk: Marks all users' notifications as read; external OnePageCRM mutation, approval
  required.
- `delete_webhook`: DELETE `/webhooks/{{ record.webhook_id }}` - kind `delete`; body type `none`;
  path fields `webhook_id`; required record fields `webhook_id`; accepted fields `webhook_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Delete a
  specific webhook; external OnePageCRM mutation, approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 67 stream-backed endpoint group(s), 92 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=4, destructive_admin=1, non_data_endpoint=1, requires_elevated_scope=1.
