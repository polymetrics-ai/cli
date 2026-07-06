# Overview

Reads and writes RD Station Marketing platform contacts, segmentation contacts, analytics, contact
fields, product catalog feeds, and workflows.

Readable streams: `contacts`, `segmentations`, `events`, `landing_pages`, `email_templates`,
`contact_detail`, `segmentation_contacts`, `contact_conversion_events`,
`contact_opportunity_events`, `contact_funnel`, `contact_fields`, `analytics_conversions`,
`analytics_emails`, `analytics_funnel`, `analytics_workflow_emails`, `catalog_feeds`,
`catalog_feed`.

Write actions: `create_contact`, `update_contact`, `delete_contact`, `add_contact_tags`,
`update_contact_funnel`, `insert_workflow_leads`, `create_contact_field`, `create_catalog_feed`,
`update_catalog_feed`, `delete_catalog_feed`.

Service API documentation: https://developers.rdstation.com/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); RD Station Marketing OAuth access token, sent as a
  Bearer token (Authorization: Bearer <access_token>). Never logged.
- `base_url` (optional, string); default `https://api.rd.services/platform`; format `uri`; RD
  Station Marketing platform API base URL override for tests or proxies.
- `catalog_feed_id` (optional, string); Product catalog feed ID for the catalog_feed detail stream.
- `contact_identifier` (optional, string); default `uuid`; allowed values `uuid`, `email`, `phone`;
  Identifier kind for contact detail/funnel streams.
- `contact_identifier_value` (optional, string); Identifier value for contact detail/funnel streams.
- `contact_uuid` (optional, string); Contact UUID for contact event streams.
- `segmentation_id` (optional, string); Segmentation ID for the segmentation_contacts stream.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.rd.services/platform`,
`contact_identifier=uuid`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/contacts` with query `page`=`1`; `page_size`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `page_size`;
starts at 1; page size 125.

Pagination by stream: none: `contact_detail`, `contact_funnel`, `contact_fields`,
`analytics_conversions`, `analytics_emails`, `analytics_funnel`, `analytics_workflow_emails`,
`catalog_feeds`, `catalog_feed`; page_number: `contacts`, `segmentations`, `events`,
`landing_pages`, `email_templates`, `segmentation_contacts`, `contact_conversion_events`,
`contact_opportunity_events`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `contacts`: GET `/contacts` - records path `contacts`; page-number pagination; page parameter
  `page`; size parameter `page_size`; starts at 1; page size 125; incremental cursor `updated_at`;
  formatted as `rfc3339`; computed output fields `id`.
- `segmentations`: GET `/segmentations` - records path `segmentations`; page-number pagination; page
  parameter `page`; size parameter `page_size`; starts at 1; page size 125; computed output fields
  `id`.
- `events`: GET `/events` - records path `events`; page-number pagination; page parameter `page`;
  size parameter `page_size`; starts at 1; page size 125; incremental cursor `created_at`; formatted
  as `rfc3339`; computed output fields `event_type`, `id`.
- `landing_pages`: GET `/landing_pages` - records path `landing_pages`; page-number pagination; page
  parameter `page`; size parameter `page_size`; starts at 1; page size 125; computed output fields
  `id`.
- `email_templates`: GET `/email_templates` - records path `email_templates`; page-number
  pagination; page parameter `page`; size parameter `page_size`; starts at 1; page size 125;
  computed output fields `id`.
- `contact_detail`: GET `/contacts/{{ config.contact_identifier }}:{{
  config.contact_identifier_value }}` - single-object response; records path `.`; computed output
  fields `id`.
- `segmentation_contacts`: GET `/segmentations/{{ config.segmentation_id }}/contacts` - records path
  `contacts`; page-number pagination; page parameter `page`; size parameter `page_size`; starts at
  1; page size 125; computed output fields `id`.
- `contact_conversion_events`: GET `/contacts/{{ config.contact_uuid }}/events` - records path `.`;
  query `event_type`=`CONVERSION`; page-number pagination; page parameter `page`; no page-size
  parameter; starts at 1; page size 10; computed output fields `email`, `id`, `name`, `phone`.
- `contact_opportunity_events`: GET `/contacts/{{ config.contact_uuid }}/events` - records path `.`;
  query `event_type`=`OPPORTUNITY`; page-number pagination; page parameter `page`; no page-size
  parameter; starts at 1; page size 10; computed output fields `email`, `id`, `name`, `phone`.
- `contact_funnel`: GET `/contacts/{{ config.contact_identifier }}:{{
  config.contact_identifier_value }}/funnels/default` - single-object response; records path `.`;
  computed output fields `id`.
- `contact_fields`: GET `/contacts/fields` - records path `fields`; computed output fields `id`.
- `analytics_conversions`: GET `/analytics/conversions` - records path `conversions`.
- `analytics_emails`: GET `/analytics/emails` - records path `emails`.
- `analytics_funnel`: GET `/analytics/funnel` - records path `funnel`.
- `analytics_workflow_emails`: GET `/analytics/workflow_emails` - records path
  `workflow_email_statistics`.
- `catalog_feeds`: GET `/catalog_feeds` - records path `.`.
- `catalog_feed`: GET `/catalog_feeds/{{ config.catalog_feed_id }}` - single-object response;
  records path `.`.

## Write actions & risks

Overall write risk: creates, updates, and deletes RD Station Marketing contacts, contact fields, and
catalog feeds; mutates contact funnels, tags, and workflow membership.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_contact`: POST `/contacts` - kind `create`; body type `json`; accepted fields `bio`,
  `birthdate`, `city`, `country`, `email`, `facebook`, `job_title`, `legal_bases`, `linkedin`,
  `mobile_phone`, `name`, `personal_phone`, `phone`, `state`, `tags`, `twitter`, `website`; risk:
  creates a contact in the RD Station Marketing lead base.
- `update_contact`: PATCH `/contacts/{{ record.identifier }}:{{ record.value }}` - kind `update`;
  body type `json`; path fields `identifier`, `value`; required record fields `identifier`, `value`;
  accepted fields `bio`, `birthdate`, `city`, `country`, `email`, `facebook`, `identifier`,
  `job_title`, `legal_bases`, `linkedin`, `mobile_phone`, `name`, `personal_phone`, `phone`,
  `state`, `tags`, `twitter`, `value`, and 1 more; risk: updates an existing RD Station Marketing
  contact.
- `delete_contact`: DELETE `/contacts/{{ record.identifier }}:{{ record.value }}` - kind `delete`;
  body type `none`; path fields `identifier`, `value`; required record fields `identifier`, `value`;
  accepted fields `identifier`, `value`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: deletes a contact from the RD Station Marketing lead base.
- `add_contact_tags`: POST `/contacts/{{ record.identifier }}:{{ record.value }}/tag` - kind
  `update`; body type `json`; path fields `identifier`, `value`; required record fields
  `identifier`, `value`, `tags`; accepted fields `identifier`, `tags`, `value`; risk: adds tags to
  an existing RD Station Marketing contact.
- `update_contact_funnel`: PUT `/contacts/{{ record.identifier }}:{{ record.value
  }}/funnels/default` - kind `update`; body type `json`; path fields `identifier`, `value`; required
  record fields `identifier`, `value`; accepted fields `contact_owner_email`, `identifier`,
  `lifecycle_stage`, `opportunity`, `value`; risk: updates lifecycle/opportunity ownership fields in
  the default contact funnel.
- `insert_workflow_leads`: POST `/workflows/{{ record.workflow_id }}/leads` - kind `update`; body
  type `json`; path fields `workflow_id`; required record fields `workflow_id`, `leads`; accepted
  fields `leads`, `workflow_id`; risk: inserts one or more leads into a marketing automation
  workflow.
- `create_contact_field`: POST `/contacts/fields` - kind `create`; body type `json`; required record
  fields `api_identifier`, `data_type`, `label`, `name`, `presentation_type`; accepted fields
  `api_identifier`, `data_type`, `label`, `name`, `presentation_type`, `validation_rules`; risk:
  creates a custom contact field in the RD Station Marketing account.
- `create_catalog_feed`: POST `/catalog_feeds` - kind `create`; body type `json`; required record
  fields `name`, `url`, `format`; accepted fields `format`, `name`, `password`, `url`, `username`;
  risk: creates a product catalog feed configuration.
- `update_catalog_feed`: PATCH `/catalog_feeds/{{ record.catalog_feed_id }}` - kind `update`; body
  type `json`; path fields `catalog_feed_id`; required record fields `catalog_feed_id`, `name`,
  `url`, `format`; accepted fields `catalog_feed_id`, `format`, `name`, `password`, `url`,
  `username`; risk: updates a product catalog feed configuration.
- `delete_catalog_feed`: DELETE `/catalog_feeds/{{ record.catalog_feed_id }}` - kind `delete`; body
  type `none`; path fields `catalog_feed_id`; required record fields `catalog_feed_id`; accepted
  fields `catalog_feed_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes a product catalog feed configuration.

## Known limits

- Batch defaults: read_page_size=125.
- API coverage includes 17 stream-backed endpoint group(s), 10 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=3, out_of_scope=7.
