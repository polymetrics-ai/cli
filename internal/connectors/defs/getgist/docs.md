# Overview

Reads Gist contacts, tags, segments, campaigns, forms, teammates, articles, collections,
conversations, teams, workspace metadata, and e-commerce resources through the Gist REST API; writes
regular JSON Gist resources and relationship actions.

Readable streams: `contacts`, `contact_details`, `contact_batch_status`, `articles`,
`article_details`, `article_search`, `article_settings`, `collections`, `collection_details`,
`events`, `tags`, `segments`, `segment_details`, `forms`, `form_details`, `form_submissions`,
`campaigns`, `campaign_details`, `subscription_types`, `subscription_type_details`, `conversations`,
`conversation_details`, `conversation_messages`, `conversation_global_counts`,
`conversation_team_counts`, `conversation_teammate_counts`, `teams`, `team_details`, `teammates`,
`teammate_details`, `token_info`, `ecommerce_stores`, `ecommerce_store_details`,
`ecommerce_customer_details`, `ecommerce_product_details`, `ecommerce_product_variant_details`,
`ecommerce_product_category_details`.

Write actions: `create_article`, `update_article`, `delete_article`, `create_collection`,
`update_collection`, `delete_collection`, `upsert_contact`, `upsert_contacts_batch`,
`delete_contact`, `track_event`, `upsert_tag`, `delete_tag`, `add_tag_to_contacts`,
`remove_tag_from_contacts`, `subscribe_contact_to_form`, `subscribe_contact_to_campaign`,
`unsubscribe_contact_from_campaign`, `attach_contact_to_subscription_type`,
`detach_contact_from_subscription_type`, `create_conversation`, `update_conversation`,
`reply_to_conversation`, `delete_conversation`, `unassign_conversation`, `assign_conversation`,
`snooze_conversation`, `unsnooze_conversation`, `close_conversation`, `prioritize_conversation`,
`tag_conversation`, `untag_conversation`, `create_store`, `update_store`, `create_customer`,
`update_customer`, `create_product`, `update_product`, `create_product_variant`,
`update_product_variant`, `create_product_category`, `update_product_category`, `upsert_cart`,
`delete_cart`, `create_order`, `update_order`.

Service API documentation: https://developers.getgist.com/api/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Gist API key. Used only for Bearer auth; never logged.
- `article_id` (optional, string); Article id for article_details, update_article, and
  delete_article.
- `article_query` (optional, string); Search text for article_search.
- `base_url` (optional, string); default `https://api.getgist.com`; format `uri`; Gist API base URL
  override for tests or proxies.
- `batch_id` (optional, string); Contact batch import id for contact_batch_status.
- `campaign_id` (optional, string); Campaign id for campaign_details.
- `category_id` (optional, string); Product category id for product_category_details.
- `collection_id` (optional, string); Collection id for collection_details.
- `contact_email` (optional, string); Optional contact email lookup filter.
- `contact_id` (optional, string); Contact id for contact_details.
- `contact_user_id` (optional, string); Optional contact user_id lookup filter.
- `conversation_id` (optional, string); Conversation id for conversation detail, message, and
  mutation endpoints.
- `customer_id` (optional, string); E-commerce customer id for customer and cart resources.
- `event_type` (optional, string); Optional event type filter for events.
- `form_id` (optional, string); Form id for form_details and form_submissions.
- `include_count` (optional, string); Optional include_count query value for segment streams.
- `mode` (optional, string).
- `order_id` (optional, string); Order id for update_order.
- `product_id` (optional, string); E-commerce product id for product and variant resources.
- `product_variant_id` (optional, string); Product variant id for update_product_variant.
- `segment_id` (optional, string); Segment id for segment_details.
- `store_id` (optional, string); E-commerce store id for store-scoped resources.
- `subscription_type_id` (optional, string); Subscription type id for subscription_type_details.
- `team_id` (optional, string); Team id for team_details.
- `teammate_id` (optional, string); Teammate id for teammate_details.
- `variant_id` (optional, string); Product variant id for product_variant_details.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.getgist.com`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/contacts` with query `per_page`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 50.

Pagination by stream: none: `contact_details`, `contact_batch_status`, `article_details`,
`article_settings`, `collection_details`, `segment_details`, `form_details`, `campaign_details`,
`subscription_type_details`, `conversation_details`, `conversation_global_counts`, `team_details`,
`teammate_details`, `token_info`, `ecommerce_store_details`, `ecommerce_customer_details`,
`ecommerce_product_details`, `ecommerce_product_variant_details`,
`ecommerce_product_category_details`; page_number: `contacts`, `articles`, `article_search`,
`collections`, `events`, `tags`, `segments`, `forms`, `form_submissions`, `campaigns`,
`subscription_types`, `conversations`, `conversation_messages`, `conversation_team_counts`,
`conversation_teammate_counts`, `teams`, `teammates`, `ecommerce_stores`.

- `contacts`: GET `/contacts` - records path `contacts`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 50.
- `contact_details`: GET `/contacts/{{ config.contact_id }}` - records path `contact`.
- `contact_batch_status`: GET `/contacts/batch/{{ config.batch_id }}` - single-object response;
  records path `.`.
- `articles`: GET `/articles` - records path `articles`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 50.
- `article_details`: GET `/articles/{{ config.article_id }}` - records path `article`.
- `article_search`: GET `/articles/search` - records path `articles`; query `query`=`{{
  config.article_query }}`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 50.
- `article_settings`: GET `/articles/settings` - single-object response; records path `.`; computed
  output fields `id`.
- `collections`: GET `/collections` - records path `collections`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 50.
- `collection_details`: GET `/collections/{{ config.collection_id }}` - records path `collection`.
- `events`: GET `/events` - records path `events`; query `event_type` from template `{{
  config.event_type }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 50.
- `tags`: GET `/tags` - records path `tags`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 50.
- `segments`: GET `/segments` - records path `segments`; query `include_count` from template `{{
  config.include_count }}`, omitted when absent; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 50.
- `segment_details`: GET `/segments/{{ config.segment_id }}` - records path `segment`; query
  `include_count` from template `{{ config.include_count }}`, omitted when absent.
- `forms`: GET `/forms` - records path `forms`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 50.
- `form_details`: GET `/forms` - records path `form`; query `form_id`=`{{ config.form_id }}`.
- `form_submissions`: GET `/forms/{{ config.form_id }}/submissions` - records path `submissions`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  50.
- `campaigns`: GET `/campaigns` - records path `campaigns`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 50.
- `campaign_details`: GET `/campaigns` - records path `campaign`; query `campaign_id`=`{{
  config.campaign_id }}`.
- `subscription_types`: GET `/subscription_types` - records path `subscription_types`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 50.
- `subscription_type_details`: GET `/subscription_types/{{ config.subscription_type_id }}` - records
  path `subscription_type`.
- `conversations`: GET `/conversations` - records path `conversations`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 50.
- `conversation_details`: GET `/conversations/{{ config.conversation_id }}` - records path
  `conversation`.
- `conversation_messages`: GET `/conversations/{{ config.conversation_id }}/messages` - records path
  `messages`; page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1;
  page size 50.
- `conversation_global_counts`: GET `/conversations/count` - single-object response; records path
  `.`; computed output fields `scope`.
- `conversation_team_counts`: GET `/conversations/count/teams` - records path `teams`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 50; computed
  output fields `scope`.
- `conversation_teammate_counts`: GET `/conversations/count/teammates` - records path `teammates`;
  page-number pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size
  50; computed output fields `scope`.
- `teams`: GET `/teams` - records path `teams`; page-number pagination; page parameter `page`; size
  parameter `per_page`; starts at 1; page size 50.
- `team_details`: GET `/teams/{{ config.team_id }}` - records path `team`.
- `teammates`: GET `/teammates` - records path `teammates`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 50.
- `teammate_details`: GET `/teammates/{{ config.teammate_id }}` - records path `teammate`.
- `token_info`: GET `/token/info` - single-object response; records path `.`; computed output fields
  `id`.
- `ecommerce_stores`: GET `/ecommerce/stores` - records path `stores`; page-number pagination; page
  parameter `page`; size parameter `per_page`; starts at 1; page size 50.
- `ecommerce_store_details`: GET `/ecommerce/stores/{{ config.store_id }}` - records path `store`.
- `ecommerce_customer_details`: GET `/ecommerce/stores/{{ config.store_id }}/customers/{{
  config.customer_id }}` - records path `customer`.
- `ecommerce_product_details`: GET `/ecommerce/stores/{{ config.store_id }}/products/{{
  config.product_id }}` - records path `product`.
- `ecommerce_product_variant_details`: GET `/ecommerce/stores/{{ config.store_id }}/products/{{
  config.product_id }}/variants/{{ config.variant_id }}` - records path `variant`.
- `ecommerce_product_category_details`: GET `/ecommerce/stores/{{ config.store_id }}/categories/{{
  config.category_id }}` - records path `category`.

## Write actions & risks

Overall write risk: creates, updates, deletes, tags, subscribes, assigns, replies to, and otherwise
mutates Gist workspace resources.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_article`: POST `/articles` - kind `create`; body type `json`; risk: create through the
  Gist API.
- `update_article`: PATCH `/articles/{{ record.article_id }}` - kind `update`; body type `json`;
  path fields `article_id`; required record fields `article_id`; accepted fields `article_id`; risk:
  update through the Gist API.
- `delete_article`: DELETE `/articles/{{ record.article_id }}` - kind `delete`; body type `none`;
  path fields `article_id`; required record fields `article_id`; accepted fields `article_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: delete
  through the Gist API.
- `create_collection`: POST `/collections` - kind `create`; body type `json`; risk: create through
  the Gist API.
- `update_collection`: POST `/collections/{{ record.collection_id }}` - kind `update`; body type
  `json`; path fields `collection_id`; required record fields `collection_id`; accepted fields
  `collection_id`; risk: update through the Gist API.
- `delete_collection`: DELETE `/collections/{{ record.collection_id }}` - kind `delete`; body type
  `none`; path fields `collection_id`; required record fields `collection_id`; accepted fields
  `collection_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: delete through the Gist API.
- `upsert_contact`: POST `/contacts` - kind `upsert`; body type `json`; risk: upsert through the
  Gist API.
- `upsert_contacts_batch`: POST `/contacts/batch` - kind `upsert`; body type `json`; risk: upsert
  through the Gist API.
- `delete_contact`: DELETE `/contacts/{{ record.contact_id }}` - kind `delete`; body type `none`;
  path fields `contact_id`; required record fields `contact_id`; accepted fields `contact_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: delete
  through the Gist API.
- `track_event`: POST `/events` - kind `create`; body type `json`; risk: create through the Gist
  API.
- `upsert_tag`: POST `/tags` - kind `upsert`; body type `json`; risk: upsert through the Gist API.
- `delete_tag`: DELETE `/tags/{{ record.tag_id }}` - kind `delete`; body type `none`; path fields
  `tag_id`; required record fields `tag_id`; accepted fields `tag_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: delete through the Gist API.
- `add_tag_to_contacts`: POST `/tags` - kind `update`; body type `json`; risk: update through the
  Gist API.
- `remove_tag_from_contacts`: POST `/tags` - kind `update`; body type `json`; risk: update through
  the Gist API.
- `subscribe_contact_to_form`: POST `/forms/{{ record.form_id }}/subscribe` - kind `update`; body
  type `json`; path fields `form_id`; required record fields `form_id`; accepted fields `form_id`;
  risk: update through the Gist API.
- `subscribe_contact_to_campaign`: POST `/campaigns` - kind `update`; body type `json`; risk: update
  through the Gist API.
- `unsubscribe_contact_from_campaign`: POST `/campaigns` - kind `update`; body type `json`; risk:
  update through the Gist API.
- `attach_contact_to_subscription_type`: POST `/subscription_types/{{ record.subscription_type_id
  }}` - kind `update`; body type `json`; path fields `subscription_type_id`; required record fields
  `subscription_type_id`; accepted fields `subscription_type_id`; risk: update through the Gist API.
- `detach_contact_from_subscription_type`: POST `/subscription_types/{{ record.subscription_type_id
  }}` - kind `update`; body type `json`; path fields `subscription_type_id`; required record fields
  `subscription_type_id`; accepted fields `subscription_type_id`; risk: update through the Gist API.
- `create_conversation`: POST `/conversations` - kind `create`; body type `json`; risk: create
  through the Gist API.
- `update_conversation`: PATCH `/conversations/{{ record.conversation_id }}` - kind `update`; body
  type `json`; path fields `conversation_id`; required record fields `conversation_id`; accepted
  fields `conversation_id`; risk: update through the Gist API.
- `reply_to_conversation`: POST `/conversations/{{ record.conversation_id }}/messages` - kind
  `create`; body type `json`; path fields `conversation_id`; required record fields
  `conversation_id`; accepted fields `conversation_id`; risk: create through the Gist API.
- `delete_conversation`: DELETE `/conversations/{{ record.conversation_id }}` - kind `delete`; body
  type `none`; path fields `conversation_id`; required record fields `conversation_id`; accepted
  fields `conversation_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: delete through the Gist API.
- `unassign_conversation`: PATCH `/conversations/{{ record.conversation_id }}/assign` - kind
  `update`; body type `json`; path fields `conversation_id`; required record fields
  `conversation_id`; accepted fields `conversation_id`; risk: update through the Gist API.
- `assign_conversation`: PATCH `/conversations/{{ record.conversation_id }}/assign` - kind `update`;
  body type `json`; path fields `conversation_id`; required record fields `conversation_id`;
  accepted fields `conversation_id`; risk: update through the Gist API.
- `snooze_conversation`: PATCH `/conversations/{{ record.conversation_id }}` - kind `update`; body
  type `json`; path fields `conversation_id`; required record fields `conversation_id`; accepted
  fields `conversation_id`; risk: update through the Gist API.
- `unsnooze_conversation`: PATCH `/conversations/{{ record.conversation_id }}` - kind `update`; body
  type `json`; path fields `conversation_id`; required record fields `conversation_id`; accepted
  fields `conversation_id`; risk: update through the Gist API.
- `close_conversation`: PATCH `/conversations/{{ record.conversation_id }}` - kind `update`; body
  type `json`; path fields `conversation_id`; required record fields `conversation_id`; accepted
  fields `conversation_id`; risk: update through the Gist API.
- `prioritize_conversation`: PATCH `/conversations/{{ record.conversation_id }}/priority` - kind
  `update`; body type `json`; path fields `conversation_id`; required record fields
  `conversation_id`; accepted fields `conversation_id`; risk: update through the Gist API.
- `tag_conversation`: POST `/conversations/{{ record.conversation_id }}/tags` - kind `update`; body
  type `json`; path fields `conversation_id`; required record fields `conversation_id`; accepted
  fields `conversation_id`; risk: update through the Gist API.
- `untag_conversation`: DELETE `/conversations/{{ record.conversation_id }}/tags` - kind `delete`;
  body type `none`; path fields `conversation_id`; required record fields `conversation_id`;
  accepted fields `conversation_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: delete through the Gist API.
- `create_store`: POST `/ecommerce/stores` - kind `create`; body type `json`; risk: create through
  the Gist API.
- `update_store`: PATCH `/ecommerce/stores/{{ record.store_id }}` - kind `update`; body type `json`;
  path fields `store_id`; required record fields `store_id`; accepted fields `store_id`; risk:
  update through the Gist API.
- `create_customer`: POST `/ecommerce/stores/{{ record.store_id }}/customers` - kind `create`; body
  type `json`; path fields `store_id`; required record fields `store_id`; accepted fields
  `store_id`; risk: create through the Gist API.
- `update_customer`: PATCH `/ecommerce/stores/{{ record.store_id }}/customers/{{ record.customer_id
  }}` - kind `update`; body type `json`; path fields `store_id`, `customer_id`; required record
  fields `store_id`, `customer_id`; accepted fields `customer_id`, `store_id`; risk: update through
  the Gist API.
- `create_product`: POST `/ecommerce/stores/{{ record.store_id }}/products` - kind `create`; body
  type `json`; path fields `store_id`; required record fields `store_id`; accepted fields
  `store_id`; risk: create through the Gist API.
- `update_product`: PATCH `/ecommerce/stores/{{ record.store_id }}/products/{{ record.product_id }}`
  - kind `update`; body type `json`; path fields `store_id`, `product_id`; required record fields
  `store_id`, `product_id`; accepted fields `product_id`, `store_id`; risk: update through the Gist
  API.
- `create_product_variant`: POST `/ecommerce/stores/{{ record.store_id }}/products/{{
  record.product_id }}/variants` - kind `create`; body type `json`; path fields `store_id`,
  `product_id`; required record fields `store_id`, `product_id`; accepted fields `product_id`,
  `store_id`; risk: create through the Gist API.
- `update_product_variant`: PATCH `/ecommerce/stores/{{ record.store_id }}/products/{{
  record.product_id }}/variants/{{ record.product_variant_id }}` - kind `update`; body type `json`;
  path fields `store_id`, `product_id`, `product_variant_id`; required record fields `store_id`,
  `product_id`, `product_variant_id`; accepted fields `product_id`, `product_variant_id`,
  `store_id`; risk: update through the Gist API.
- `create_product_category`: POST `/ecommerce/stores/{{ record.store_id }}/categories` - kind
  `create`; body type `json`; path fields `store_id`; required record fields `store_id`; accepted
  fields `store_id`; risk: create through the Gist API.
- `update_product_category`: PATCH `/ecommerce/stores/{{ record.store_id }}/categories/{{
  record.category_id }}` - kind `update`; body type `json`; path fields `store_id`, `category_id`;
  required record fields `store_id`, `category_id`; accepted fields `category_id`, `store_id`; risk:
  update through the Gist API.
- `upsert_cart`: POST `/ecommerce/stores/{{ record.store_id }}/customers/{{ record.customer_id
  }}/cart` - kind `upsert`; body type `json`; path fields `store_id`, `customer_id`; required record
  fields `store_id`, `customer_id`; accepted fields `customer_id`, `store_id`; risk: upsert through
  the Gist API.
- `delete_cart`: DELETE `/ecommerce/stores/{{ record.store_id }}/customers/{{ record.customer_id
  }}/cart` - kind `delete`; body type `none`; path fields `store_id`, `customer_id`; required record
  fields `store_id`, `customer_id`; accepted fields `customer_id`, `store_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: delete through the Gist
  API.
- `create_order`: POST `/ecommerce/stores/{{ record.store_id }}/orders` - kind `create`; body type
  `json`; path fields `store_id`; required record fields `store_id`; accepted fields `store_id`;
  risk: create through the Gist API.
- `update_order`: PATCH `/ecommerce/stores/{{ record.store_id }}/orders/{{ record.order_id }}` -
  kind `update`; body type `json`; path fields `store_id`, `order_id`; required record fields
  `store_id`, `order_id`; accepted fields `order_id`, `store_id`; risk: update through the Gist API.

## Known limits

- Batch defaults: read_page_size=50, write_batch_size=1.
- API coverage includes 37 stream-backed endpoint group(s), 45 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=1, out_of_scope=2.
