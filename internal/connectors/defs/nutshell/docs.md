# Overview

Reads and writes documented Nutshell CRM REST resources through the Nutshell REST API.

Readable streams: `accounts`, `contacts`, `leads`, `activities`, `users`, `account`,
`account_custom_fields`, `account_custom_field_attributes`, `account_list_items`,
`account_list_fields`, `account_types`, `activity`, `activity_types`, `audiences`,
`competitor_maps`, `competitor_map`, `competitors`, `competitor`, `contact`,
`contact_custom_fields`, `contact_custom_field_attributes`, `contact_list_items`,
`contact_list_fields`, `editions`, `edition`, `email`, `events`, `deleted_events`, `saved_filters`,
`forms`, `form_field`, `form`, `industries`, `industry`, `invoices`, `invoice`, `lead`,
`lead_custom_fields`, `lead_installments`, `lead_stages`, `lead_custom_field_attributes`,
`lead_list_items`, `lead_list_fields`, `lead_reports`, `markets`, `market`, `notes`, `note`,
`lead_outcomes`, `lead_outcome`, `product_categories`, `product_category`, `product_maps`,
`product_map`, `products`, `product`, `quotes`, `quote`, `sources`, `stages`, `pipelines`, `tags`,
`tasks`, `task`, `territories`, `user`.

Write actions: `create_account`, `delete_account`, `undelete_account`,
`create_account_custom_field`, `create_activity`, `update_activity`, `create_audience`,
`delete_competitor_map`, `create_contact`, `delete_contact`, `undelete_contact`,
`create_contact_custom_field`, `create_lead`, `delete_lead`, `reopen_lead`, `set_lead_pipeline`,
`update_lead_status`, `undelete_lead`, `watch_lead`, `create_lead_custom_field`, `create_note`,
`delete_note`, `undelete_note`, `create_product_category`, `delete_product_map`, `delete_product`,
`undelete_product`, `create_source`, `delete_source`, `undelete_source`, `create_tag`, `delete_tag`,
`undelete_tag`, `create_task`, `delete_task`.

Service API documentation: https://developers.nutshell.com/docs/getting-started.

## Auth setup

Connection fields:

- `account_id` (optional, string); Optional account id used by the corresponding detail stream.
- `activity_id` (optional, string); Optional activity id used by the corresponding detail stream.
- `base_url` (optional, string); default `https://app.nutshell.com/rest`; format `uri`; Nutshell
  REST API base URL override for tests or proxies.
- `competitor_id` (optional, string); Optional competitor id used by the corresponding detail
  stream.
- `competitor_map_id` (optional, string); Optional competitor map id used by the corresponding
  detail stream.
- `contact_id` (optional, string); Optional contact id used by the corresponding detail stream.
- `edition_id` (optional, string); Optional edition id used by the corresponding detail stream.
- `email_id` (optional, string); Optional email id used by the corresponding detail stream.
- `form_field_id` (optional, string); Optional form field id used by the corresponding detail
  stream.
- `form_id` (optional, string); Optional form id used by the corresponding detail stream.
- `industry_id` (optional, string); Optional industry id used by the corresponding detail stream.
- `invoice_id` (optional, string); Optional invoice id used by the corresponding detail stream.
- `lead_id` (optional, string); Optional lead id used by the corresponding detail stream.
- `market_id` (optional, string); Optional market id used by the corresponding detail stream.
- `mode` (optional, string).
- `note_id` (optional, string); Optional note id used by the corresponding detail stream.
- `outcome_id` (optional, string); Optional outcome id used by the corresponding detail stream.
- `page_size` (optional, string); default `500`; Records per page (1-500) for paginated endpoints.
- `password` (required, secret, string); Nutshell API token, sent as the HTTP Basic auth password.
  Never logged.
- `product_category_id` (optional, string); Optional product category id used by the corresponding
  detail stream.
- `product_id` (optional, string); Optional product id used by the corresponding detail stream.
- `product_map_id` (optional, string); Optional product map id used by the corresponding detail
  stream.
- `quote_id` (optional, string); Optional quote id used by the corresponding detail stream.
- `task_id` (optional, string); Optional task id used by the corresponding detail stream.
- `user_id` (optional, string); Optional user id used by the corresponding detail stream.
- `username` (required, string); Nutshell account username/email used for HTTP Basic auth.

Secret fields are redacted in logs and write previews: `password`.

Default configuration values: `base_url=https://app.nutshell.com/rest`, `page_size=500`.

Authentication behavior:

- HTTP Basic authentication using `config.username`, `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/users` with query `page[limit]`=`1`; `page[page]`=`0`.

## Streams notes

Default pagination: page-number pagination; page parameter `page[page]`; size parameter
`page[limit]`; starts at 0; page size 500.

Pagination by stream: none: `users`, `account`, `account_custom_fields`,
`account_custom_field_attributes`, `account_list_items`, `account_list_fields`, `account_types`,
`activity`, `activity_types`, `audiences`, `competitor_maps`, `competitor_map`, `competitors`,
`competitor`, `contact`, `contact_custom_fields`, `contact_custom_field_attributes`,
`contact_list_items`, `contact_list_fields`, `editions`, `edition`, `email`, `events`,
`deleted_events`, `saved_filters`, `forms`, `form_field`, `form`, `industries`, `industry`,
`invoice`, `lead`, `lead_custom_fields`, `lead_installments`, `lead_stages`,
`lead_custom_field_attributes`, `lead_list_items`, `lead_list_fields`, `lead_reports`, `markets`,
`market`, `note`, `lead_outcomes`, `lead_outcome`, `product_categories`, `product_category`,
`product_maps`, `product_map`, `products`, `product`, `quote`, `sources`, `stages`, `pipelines`,
`tags`, `task`, `territories`, `user`; page_number: `accounts`, `contacts`, `leads`, `activities`,
`invoices`, `notes`, `quotes`, `tasks`.

- `accounts`: GET `/accounts` - records path `accounts`; page-number pagination; page parameter
  `page[page]`; size parameter `page[limit]`; starts at 0; page size 500.
- `contacts`: GET `/contacts` - records path `contacts`; page-number pagination; page parameter
  `page[page]`; size parameter `page[limit]`; starts at 0; page size 500.
- `leads`: GET `/leads` - records path `leads`; page-number pagination; page parameter `page[page]`;
  size parameter `page[limit]`; starts at 0; page size 500.
- `activities`: GET `/activities` - records path `activities`; page-number pagination; page
  parameter `page[page]`; size parameter `page[limit]`; starts at 0; page size 500.
- `users`: GET `/users` - records path `users`.
- `account`: GET `/accounts/{{ config.account_id }}` - records path `accounts`; emits passthrough
  records.
- `account_custom_fields`: GET `/accounts/{{ config.account_id }}/customfields` - records path
  `customFields`; emits passthrough records.
- `account_custom_field_attributes`: GET `/accounts/customfields/attributes` - records path
  `customFields`; emits passthrough records.
- `account_list_items`: GET `/accounts/list` - records path `listItems`; emits passthrough records.
- `account_list_fields`: GET `/accounts/list/fields` - single-object response; records at response
  root; emits passthrough records.
- `account_types`: GET `/accounttypes` - records at response root; emits passthrough records.
- `activity`: GET `/activities/{{ config.activity_id }}` - records path `activities`; emits
  passthrough records.
- `activity_types`: GET `/activitytypes` - records path `activityTypes`; emits passthrough records.
- `audiences`: GET `/audiences` - records at response root; emits passthrough records.
- `competitor_maps`: GET `/competitormaps` - records path `competitorMaps`; emits passthrough
  records.
- `competitor_map`: GET `/competitormaps/{{ config.competitor_map_id }}` - records path
  `competitorMaps`; emits passthrough records.
- `competitors`: GET `/competitors` - records path `competitors`; emits passthrough records.
- `competitor`: GET `/competitors/{{ config.competitor_id }}` - records path `competitors`; emits
  passthrough records.
- `contact`: GET `/contacts/{{ config.contact_id }}` - records path `contacts`; emits passthrough
  records.
- `contact_custom_fields`: GET `/contacts/{{ config.contact_id }}/customfields` - records at
  response root; emits passthrough records.
- `contact_custom_field_attributes`: GET `/contacts/customfields/attributes` - records at response
  root; emits passthrough records.
- `contact_list_items`: GET `/contacts/list` - records path `listItems`; emits passthrough records.
- `contact_list_fields`: GET `/contacts/list/fields` - single-object response; records at response
  root; emits passthrough records.
- `editions`: GET `/editions` - records at response root; emits passthrough records.
- `edition`: GET `/editions/{{ config.edition_id }}` - single-object response; records at response
  root; emits passthrough records.
- `email`: GET `/emails/{{ config.email_id }}` - single-object response; records at response root;
  emits passthrough records.
- `events`: GET `/events` - records path `events`; emits passthrough records.
- `deleted_events`: GET `/events/deleted` - records path `events`; emits passthrough records.
- `saved_filters`: GET `/filters` - records path `filters`; emits passthrough records.
- `forms`: GET `/forms` - records path `wfForms`; emits passthrough records.
- `form_field`: GET `/forms/{{ config.form_field_id }}` - records path `wfFields`; emits passthrough
  records.
- `form`: GET `/forms/{{ config.form_id }}` - records path `wfForms`; emits passthrough records.
- `industries`: GET `/industries` - records at response root; emits passthrough records.
- `industry`: GET `/industries/{{ config.industry_id }}` - single-object response; records at
  response root; emits passthrough records.
- `invoices`: GET `/invoices` - records path `invoices`; page-number pagination; page parameter
  `page[page]`; size parameter `page[limit]`; starts at 0; page size 500; emits passthrough records.
- `invoice`: GET `/invoices/{{ config.invoice_id }}` - records path `invoices`; emits passthrough
  records.
- `lead`: GET `/leads/{{ config.lead_id }}` - records path `leads`; emits passthrough records.
- `lead_custom_fields`: GET `/leads/{{ config.lead_id }}/customfields` - records at response root;
  emits passthrough records.
- `lead_installments`: GET `/leads/{{ config.lead_id }}/installments` - records path `installments`;
  emits passthrough records.
- `lead_stages`: GET `/leads/{{ config.lead_id }}/stages` - records path `stages`; emits passthrough
  records.
- `lead_custom_field_attributes`: GET `/leads/customfields/attributes` - records at response root;
  emits passthrough records.
- `lead_list_items`: GET `/leads/list` - records path `listItems`; emits passthrough records.
- `lead_list_fields`: GET `/leads/list/fields` - single-object response; records at response root;
  emits passthrough records.
- `lead_reports`: GET `/leads/report` - records path `reports`; emits passthrough records.
- `markets`: GET `/markets` - records at response root; emits passthrough records.
- `market`: GET `/markets/{{ config.market_id }}` - single-object response; records at response
  root; emits passthrough records.
- `notes`: GET `/notes` - records path `notes`; page-number pagination; page parameter `page[page]`;
  size parameter `page[limit]`; starts at 0; page size 500; emits passthrough records.
- `note`: GET `/notes/{{ config.note_id }}` - single-object response; records at response root;
  emits passthrough records.
- `lead_outcomes`: GET `/outcomes` - records at response root; emits passthrough records.
- `lead_outcome`: GET `/outcomes/{{ config.outcome_id }}` - single-object response; records at
  response root; emits passthrough records.
- `product_categories`: GET `/productcategories` - records path `productCategories`; emits
  passthrough records.
- `product_category`: GET `/productcategories/{{ config.product_category_id }}` - records path
  `productCategories`; emits passthrough records.
- `product_maps`: GET `/productmaps` - records path `productMaps`; emits passthrough records.
- `product_map`: GET `/productmaps/{{ config.product_map_id }}` - records path `productMaps`; emits
  passthrough records.
- `products`: GET `/products` - records path `products`; emits passthrough records.
- `product`: GET `/products/{{ config.product_id }}` - records path `products`; emits passthrough
  records.
- `quotes`: GET `/quotes` - records path `quotes`; page-number pagination; page parameter
  `page[page]`; size parameter `page[limit]`; starts at 0; page size 500; emits passthrough records.
- `quote`: GET `/quotes/{{ config.quote_id }}` - records path `quotes`; emits passthrough records.
- `sources`: GET `/sources` - records path `sources`; emits passthrough records.
- `stages`: GET `/stages` - records at response root; emits passthrough records.
- `pipelines`: GET `/stagesets` - records at response root; emits passthrough records.
- `tags`: GET `/tags` - records path `tags`; emits passthrough records.
- `tasks`: GET `/tasks` - records path `tasks`; page-number pagination; page parameter `page[page]`;
  size parameter `page[limit]`; starts at 0; page size 500; emits passthrough records.
- `task`: GET `/tasks/{{ config.task_id }}` - single-object response; records at response root;
  emits passthrough records.
- `territories`: GET `/territories` - records at response root; emits passthrough records.
- `user`: GET `/users/{{ config.user_id }}` - records path `users`; emits passthrough records.

## Write actions & risks

Overall write risk: external Nutshell CRM mutations including creates, updates, undeletes, watches,
and destructive deletes.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_account`: POST `/accounts` - kind `create`; body type `json`; required record fields
  `accounts`; accepted fields `accounts`; risk: creates company/account records in Nutshell.
- `delete_account`: DELETE `/accounts/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: deletes a Nutshell account/company;
  recoverable only through the undelete endpoint for a limited period.
- `undelete_account`: POST `/accounts/{{ record.id }}/undelete` - kind `update`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: restores a deleted
  Nutshell account/company.
- `create_account_custom_field`: POST `/accounts/customfield` - kind `create`; body type `json`;
  required record fields `name`, `type`; accepted fields `choices`, `id`, `isMultiple`, `name`,
  `title`, `type`; risk: creates an account custom field definition.
- `create_activity`: POST `/activities` - kind `create`; body type `json`; required record fields
  `activities`; accepted fields `activities`; risk: creates Nutshell activity records.
- `update_activity`: PUT `/activities/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `activities`; accepted fields `activities`, `id`; risk:
  updates an existing Nutshell activity.
- `create_audience`: POST `/audiences` - kind `create`; body type `json`; required record fields
  `emAudiences`; accepted fields `emAudiences`; risk: creates a Nutshell email marketing audience.
- `delete_competitor_map`: DELETE `/competitormaps/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: deletes a lead-competitor
  relationship.
- `create_contact`: POST `/contacts` - kind `create`; body type `json`; required record fields
  `contacts`; accepted fields `contacts`; risk: creates person/contact records in Nutshell.
- `delete_contact`: DELETE `/contacts/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: deletes a Nutshell contact/person; recoverable
  only through the undelete endpoint for a limited period.
- `undelete_contact`: POST `/contacts/{{ record.id }}/undelete` - kind `update`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: restores a deleted
  Nutshell contact/person.
- `create_contact_custom_field`: POST `/contacts/customfield` - kind `create`; body type `json`;
  required record fields `name`, `type`; accepted fields `choices`, `id`, `isMultiple`, `name`,
  `title`, `type`; risk: creates a contact custom field definition.
- `create_lead`: POST `/leads` - kind `create`; body type `json`; required record fields `leads`;
  accepted fields `leads`; risk: creates Nutshell lead records.
- `delete_lead`: DELETE `/leads/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: deletes a Nutshell lead; recoverable only through
  the undelete endpoint for a limited period.
- `reopen_lead`: POST `/leads/{{ record.id }}/reopen` - kind `update`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; risk: reopens a previously closed
  Nutshell lead.
- `set_lead_pipeline`: POST `/leads/{{ record.id }}/stageset` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`, `stageset`; accepted fields `id`, `stageset`; risk:
  changes the pipeline/stageset assigned to a lead.
- `update_lead_status`: POST `/leads/{{ record.id }}/status` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `competitorMaps`, `id`, `outcomeId`,
  `productMaps`; risk: updates a lead status/outcome and optional competitor/product maps.
- `undelete_lead`: POST `/leads/{{ record.id }}/undelete` - kind `update`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: restores a deleted Nutshell
  lead.
- `watch_lead`: POST `/leads/{{ record.id }}/watch` - kind `update`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; risk: toggles watch notifications for the
  authenticated user on a lead.
- `create_lead_custom_field`: POST `/leads/customfield` - kind `create`; body type `json`; required
  record fields `name`, `type`; accepted fields `choices`, `id`, `isMultiple`, `name`, `title`,
  `type`; risk: creates a lead custom field definition.
- `create_note`: POST `/notes` - kind `create`; body type `json`; required record fields `data`;
  accepted fields `data`; risk: creates a note attached to a Nutshell entity.
- `delete_note`: DELETE `/notes/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: deletes a Nutshell note; recoverable only through
  the undelete endpoint for a limited period.
- `undelete_note`: POST `/notes/{{ record.id }}/undelete` - kind `update`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: restores a deleted Nutshell
  note.
- `create_product_category`: POST `/productcategories` - kind `create`; body type `json`; required
  record fields `productCategories`; accepted fields `productCategories`; risk: creates a Nutshell
  product category.
- `delete_product_map`: DELETE `/productMaps/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: deletes a product mapping from a lead.
- `delete_product`: DELETE `/products/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: deletes a Nutshell product; recoverable only
  through the undelete endpoint for a limited period.
- `undelete_product`: POST `/products/{{ record.id }}/undelete` - kind `update`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: restores a deleted
  Nutshell product.
- `create_source`: POST `/sources` - kind `create`; body type `json`; required record fields
  `sources`; accepted fields `sources`; risk: creates a lead source in Nutshell.
- `delete_source`: DELETE `/sources/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: deletes a lead source; recoverable only through
  the undelete endpoint for a limited period.
- `undelete_source`: POST `/sources/{{ record.id }}/undelete` - kind `update`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: restores a deleted lead
  source.
- `create_tag`: POST `/tags` - kind `create`; body type `json`; required record fields `tags`;
  accepted fields `tags`; risk: creates a Nutshell tag and optionally links entities.
- `delete_tag`: DELETE `/tags/{{ record.id }}` - kind `delete`; body type `none`; path fields `id`;
  required record fields `id`; accepted fields `id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: deletes a Nutshell tag; recoverable only through the
  undelete endpoint for a limited period.
- `undelete_tag`: POST `/tags/{{ record.id }}/undelete` - kind `update`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: restores a deleted Nutshell
  tag.
- `create_task`: POST `/tasks` - kind `create`; body type `json`; required record fields `title`;
  accepted fields `description`, `dueTime`, `links`, `recurrenceRule`, `title`; risk: creates a task
  in Nutshell.
- `delete_task`: DELETE `/tasks/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: deletes a Nutshell task.

## Known limits

- Batch defaults: read_page_size=500.
- API coverage includes 66 stream-backed endpoint group(s), 35 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, out_of_scope=7.
