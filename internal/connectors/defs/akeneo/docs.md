# Overview

Reads Akeneo PIM products, categories, families, attributes, channels, product models, family
variants, attribute groups, association types, locales, currencies, and measure families, and writes
create-or-update upserts for the 9 catalog-structure resources, through the Akeneo REST API (OAuth2
password grant).

Readable streams: `products`, `categories`, `families`, `attributes`, `channels`, `product_models`,
`family_variants`, `attribute_groups`, `association_types`, `locales`, `currencies`,
`measure_families`.

Write actions: `create_or_update_product`, `create_or_update_category`, `create_or_update_family`,
`create_or_update_attribute`, `create_or_update_channel`, `create_or_update_product_model`,
`create_or_update_family_variant`, `create_or_update_attribute_group`,
`create_or_update_association_type`.

Service API documentation: https://api.akeneo.com/api-reference.html.

## Auth setup

Connection fields:

- `api_username` (required, string); Akeneo API user username, sent in the password-grant
  token-exchange JSON body.
- `base_url` (required, string); format `uri`; Akeneo PIM host URL (e.g.
  https://example.trial.akeneo.cloud). An absolute http(s) URL with a host is required; the
  connector fails closed otherwise (SSRF guard).
- `client_id` (required, string); Akeneo OAuth2 API client ID (Basic-auth'd, paired with the client
  secret, on the token-exchange request only).
- `page_size` (optional, string); default `100`; Records per page (1-100, the limit query param).
- `password` (optional, secret, string); Akeneo API user password, sent in the password-grant
  token-exchange JSON body. Never logged.
- `secret` (optional, secret, string); Akeneo OAuth2 API client secret, paired with client_id as
  HTTP Basic auth on the token-exchange request. Never logged.

Secret fields are redacted in logs and write previews: `password`, `secret`.

Default configuration values: `page_size=100`.

Authentication behavior:

- Connector-specific authentication using `config.api_username`, `secrets.password`,
  `config.base_url`, `config.client_id`, `secrets.secret`.
- OAuth2 access tokens are cached and reused until 60 seconds before their declared expiration
  time (`expires_in`), at which point a new token is requested.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/rest/v1/channels` with query `limit`=`1`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `_links.next.href`;
next URLs stay on the configured API host.

The computed `id` output field is derived by coalescing `record.code`, `record.identifier`, and
`record.uuid` in that order: products normally resolve through `identifier`, other streams through
`code`, and `uuid` serves as a fallback.

- `products`: GET `/api/rest/v1/products` - records path `_embedded.items`; query `limit`=`{{
  config.page_size }}`; follows a next-page URL from the response body; URL path `_links.next.href`;
  next URLs stay on the configured API host; computed output fields `id`.
- `categories`: GET `/api/rest/v1/categories` - records path `_embedded.items`; query `limit`=`{{
  config.page_size }}`; follows a next-page URL from the response body; URL path `_links.next.href`;
  next URLs stay on the configured API host; computed output fields `id`.
- `families`: GET `/api/rest/v1/families` - records path `_embedded.items`; query `limit`=`{{
  config.page_size }}`; follows a next-page URL from the response body; URL path `_links.next.href`;
  next URLs stay on the configured API host; computed output fields `id`.
- `attributes`: GET `/api/rest/v1/attributes` - records path `_embedded.items`; query `limit`=`{{
  config.page_size }}`; follows a next-page URL from the response body; URL path `_links.next.href`;
  next URLs stay on the configured API host; computed output fields `id`.
- `channels`: GET `/api/rest/v1/channels` - records path `_embedded.items`; query `limit`=`{{
  config.page_size }}`; follows a next-page URL from the response body; URL path `_links.next.href`;
  next URLs stay on the configured API host; computed output fields `id`.
- `product_models`: GET `/api/rest/v1/product-models` - records path `_embedded.items`; query
  `limit`=`{{ config.page_size }}`; follows a next-page URL from the response body; URL path
  `_links.next.href`; next URLs stay on the configured API host; computed output fields `id`.
- `family_variants`: GET `/api/rest/v1/family-variants` - records path `_embedded.items`; query
  `limit`=`{{ config.page_size }}`; follows a next-page URL from the response body; URL path
  `_links.next.href`; next URLs stay on the configured API host; computed output fields `id`.
- `attribute_groups`: GET `/api/rest/v1/attribute-groups` - records path `_embedded.items`; query
  `limit`=`{{ config.page_size }}`; follows a next-page URL from the response body; URL path
  `_links.next.href`; next URLs stay on the configured API host; computed output fields `id`.
- `association_types`: GET `/api/rest/v1/association-types` - records path `_embedded.items`; query
  `limit`=`{{ config.page_size }}`; follows a next-page URL from the response body; URL path
  `_links.next.href`; next URLs stay on the configured API host; computed output fields `id`.
- `locales`: GET `/api/rest/v1/locales` - records path `_embedded.items`; query `limit`=`{{
  config.page_size }}`; follows a next-page URL from the response body; URL path `_links.next.href`;
  next URLs stay on the configured API host; computed output fields `id`.
- `currencies`: GET `/api/rest/v1/currencies` - records path `_embedded.items`; query `limit`=`{{
  config.page_size }}`; follows a next-page URL from the response body; URL path `_links.next.href`;
  next URLs stay on the configured API host; computed output fields `id`.
- `measure_families`: GET `/api/rest/v1/measure-families` - records path `_embedded.items`; query
  `limit`=`{{ config.page_size }}`; follows a next-page URL from the response body; URL path
  `_links.next.href`; next URLs stay on the configured API host; computed output fields `id`.

## Write actions & risks

Overall write risk: external Akeneo PIM API upsert (create-or-update, PATCH-based) of products,
categories, families, attributes, channels, product models, family variants, attribute groups, and
association types; schema-shaping mutations (family/attribute/attribute-group) affect every product
referencing them, approval required.

The `locales`, `currencies`, and `measure_families` streams are read-only: Akeneo's API provides
no per-record write endpoints for these resources.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_or_update_product`: PATCH `/api/rest/v1/products/{{ record.id }}` - kind `upsert`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `associations`,
  `categories`, `enabled`, `family`, `groups`, `id`, `parent`, `values`; risk: creates a new product
  (201) or updates an existing one (204) in the connected Akeneo PIM catalog; visible to every
  downstream channel the product is enabled/categorized for.
- `create_or_update_category`: PATCH `/api/rest/v1/categories/{{ record.id }}` - kind `upsert`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `id`, `labels`,
  `parent`; risk: creates or updates a category node; re-parenting an existing category changes the
  catalog tree for every product classified under it.
- `create_or_update_family`: PATCH `/api/rest/v1/families/{{ record.id }}` - kind `upsert`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `attribute_as_image`,
  `attribute_as_label`, `attributes`, `id`, `labels`; risk: creates or updates a product family
  definition; changes the required/optional attribute set for every product assigned to this family.
- `create_or_update_attribute`: PATCH `/api/rest/v1/attributes/{{ record.id }}` - kind `upsert`;
  body type `json`; path fields `id`; required record fields `id`; accepted fields `group`, `id`,
  `labels`, `localizable`, `scopable`, `type`; risk: creates a new attribute (schema mutation,
  affects every family/product referencing it) or updates an existing one's non-structural
  properties (labels/group); some attribute properties are immutable after creation per Akeneo's own
  API rules.
- `create_or_update_channel`: PATCH `/api/rest/v1/channels/{{ record.id }}` - kind `upsert`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `category_tree`,
  `currencies`, `id`, `labels`, `locales`; risk: creates or updates a distribution channel; changes
  which locales/currencies/category tree every product exported to this channel uses.
- `create_or_update_product_model`: PATCH `/api/rest/v1/product-models/{{ record.id }}` - kind
  `upsert`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `categories`, `family_variant`, `id`, `parent`, `values`; risk: creates or updates a product
  model; a shared parent for variant products, changes propagate to every variant beneath it.
- `create_or_update_family_variant`: PATCH `/api/rest/v1/family-variants/{{ record.id }}` - kind
  `upsert`; body type `json`; path fields `id`; required record fields `id`; accepted fields `id`,
  `labels`, `variant_attribute_sets`; risk: creates or updates a family variant's axis/attribute-set
  structure; changes which attributes distinguish variant products under this family.
- `create_or_update_attribute_group`: PATCH `/api/rest/v1/attribute-groups/{{ record.id }}` - kind
  `upsert`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `attributes`, `id`, `labels`, `sort_order`; risk: creates or updates an attribute group;
  reorganizes attribute grouping in the PIM's data-entry UI, a low-risk organizational mutation.
- `create_or_update_association_type`: PATCH `/api/rest/v1/association-types/{{ record.id }}` - kind
  `upsert`; body type `json`; path fields `id`; required record fields `id`; accepted fields `id`,
  `is_two_way`, `labels`; risk: creates or updates an association type (e.g. cross-sell/up-sell
  relationship definition); low-risk organizational mutation, no product data changes on its own.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 12 stream-backed endpoint group(s), 9 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=9, destructive_admin=4, duplicate_of=39, non_data_endpoint=5, out_of_scope=42,
  requires_elevated_scope=32.
