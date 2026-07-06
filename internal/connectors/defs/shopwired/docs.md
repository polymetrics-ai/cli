# Overview

Reads and writes ShopWired v1 catalog, order, customer, content, marketing, shipping, theme,
webhook, wishlist, and payment resources through the documented REST API.

Readable streams: `app`, `data`, `blog_categories`, `blog_category`, `blog_posts`, `blog_post`,
`blog_tags`, `blog_tag`, `brands`, `brand`, `business`, `features`, `categories`, `category`,
`choice_sets`, `choice_set`, `collect_locations`, `collect_location`, `custom_fields`,
`custom_field`, `customers`, `customer`, `digital_files`, `digital_file`, `events`, `event`,
`filter_groups`, `filter_group`, `gift_vouchers`, `gift_voucher`, `incomplete_orders`,
`incomplete_order`, `newsletter_subscribers`, `newsletter_subscriber`, `nexuses`, `offers`, `offer`,
`order_statuses`, `orders`, `search`, `order`, `pages`, `page`, `payment_methods`, `products`,
`search_2`, `product`, `bulk_prices`, `bulk_price`, `choices`, `choice`, `customization_fields`,
`extras`, `images`, `options`, `option`, `reviews`, `variations`, `variation`, `redirects`,
`redirect`, `sales`, `sale`, `shipping_rates`, `shipping_rate`, `shipping_zones`, `shipping_zone`,
`disputes`, `dispute`, `payouts`, `payout`, `transactions`, `transaction`, `stock`,
`stock_requests`, `theme_assets`, `themes`, `theme`, `trade_customer_product_prices`,
`trade_customer_product_price`, `trade_groups`, `trade_group`, `vouchers`, `voucher`, `webhooks`,
`webhook`, `wishlists`, `wishlist`.

Write actions: `deletes_all_app_data`, `create_new_app_data`, `create_blog_category`,
`delete_blog_category`, `update_blog_category`, `create_blog_post`, `delete_blog_post`,
`update_blog_post`, `create_blog_tag`, `delete_blog_tag`, `update_blog_tag`, `create_brand`,
`delete_brand`, `update_brand`, `change_business_feature_status`, `create_category`,
`delete_category_by_id`, `update_category_by_id`, `create_choice_set_value`,
`delete_choice_set_value`, `update_choice_set_value`, `create_choice_set`, `delete_choice_set`,
`update_choice_set`, `create_custom_field`, `delete_custom_field`, `update_custom_field`,
`create_customer`, `create_digital_file`, `delete_digital_file`, `update_digital_file`,
`create_filter_group`, `delete_filter_group`, `update_filter_group`, `create_gift_card`,
`delete_gift_card`, `update_gift_card`, `create_business_nexus`, `delete_business_nexus`,
`update_business_nexus`, `create_offer`, `delete_offer`, `update_offer`, `create_order_status`,
`delete_order_status`, `create_order`, `delete_order`, `create_an_order_admin_comment`,
`delete_an_order_admin_comment`, `request_a_pre_order_be_dispatched`, `update_an_order_s_status`,
`create_refund`, `create_page`, `delete_page`, `update_page`, `create_product`, `update_prices`,
`delete_product`, `update_product`, `create_bulk_price`, `delete_bulk_price`, `update_bulk_price`,
`assign_product_choice`, `delete_product_choice`, `update_product_choice`,
`create_product_customization_field`, `delete_product_customization_field`,
`update_product_customization_field`, `create_product_extra`, `delete_product_extra`,
`update_product_extra`, `create_product_image`, `delete_product_image`, `update_product_image`,
`create_product_option_value`, `delete_product_option_value`, `update_product_option_value`,
`create_product_option`, `delete_product_option`, `update_product_option`, and 39 more.

Service API documentation: https://shopwired.readme.io/reference.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string); Optional OAuth access token for bearer authentication.
- `active` (optional, string); Optional ShopWired API parameter active.
- `api_key` (required, secret, string); ShopWired API key.
- `api_secret` (optional, secret, string); ShopWired API secret, used as the Basic-auth password
  with api_key for the current documented API.
- `archived` (optional, string); Optional ShopWired API parameter archived.
- `base_url` (optional, string); default `https://api.ecommerceapi.uk/v1`; format `uri`; ShopWired
  API base URL.
- `code` (optional, string); Optional ShopWired API parameter code.
- `count` (optional, string); Optional ShopWired API parameter count.
- `country_id` (optional, string); Optional ShopWired API parameter country_id.
- `created_after` (optional, string); Optional ShopWired API parameter created_after.
- `created_before` (optional, string); Optional ShopWired API parameter created_before.
- `customer_id` (optional, string); Optional ShopWired API parameter customer_id.
- `dispute_id` (optional, string); Optional ShopWired API parameter dispute_id.
- `email` (optional, string); Optional ShopWired API parameter email.
- `embed` (optional, string); Optional ShopWired API parameter embed.
- `fields` (optional, string); Optional ShopWired API parameter fields.
- `from` (optional, string); Optional ShopWired API parameter from.
- `id` (optional, string); Optional ShopWired API parameter id.
- `ids` (optional, string); Optional ShopWired API parameter ids.
- `key` (optional, string); Optional ShopWired API parameter key.
- `keyword` (optional, string); Optional ShopWired API parameter keyword.
- `keywords` (optional, string); Optional ShopWired API parameter keywords.
- `name` (optional, string); Optional ShopWired API parameter name.
- `offset` (optional, string); Optional ShopWired API parameter offset.
- `path` (optional, string); Optional ShopWired API parameter path.
- `payout_id` (optional, string); Optional ShopWired API parameter payout_id.
- `product_id` (optional, string); Optional ShopWired API parameter product_id.
- `session_id` (optional, string); Optional ShopWired API parameter session_id.
- `since_id` (optional, string); Optional ShopWired API parameter since_id.
- `sku` (optional, string); Optional ShopWired API parameter sku.
- `sort` (optional, string); Optional ShopWired API parameter sort.
- `status` (optional, string); Optional ShopWired API parameter status.
- `subject_id` (optional, string); Optional ShopWired API parameter subject_id.
- `subject_type` (optional, string); Optional ShopWired API parameter subject_type.
- `theme_id` (optional, string); Optional ShopWired API parameter theme_id.
- `to` (optional, string); Optional ShopWired API parameter to.
- `topic` (optional, string); Optional ShopWired API parameter topic.
- `trade` (optional, string); Optional ShopWired API parameter trade.
- `transaction_id` (optional, string); Optional ShopWired API parameter transaction_id.

Secret fields are redacted in logs and write previews: `access_token`, `api_key`, `api_secret`.

Default configuration values: `base_url=https://api.ecommerceapi.uk/v1`.

Authentication behavior:

- HTTP Basic authentication using `secrets.api_key`, `secrets.api_secret` when `{{
  secrets.api_secret }}`.
- Bearer token authentication using `secrets.access_token` when `{{ secrets.access_token }}`.
- API key authentication in `X-API-Key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/products` with query `count`=`1`; `offset`=`0`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `count`;
page size 100.

Pagination by stream: none: `app`, `data`, `blog_category`, `blog_post`, `blog_tag`, `brand`,
`business`, `category`, `choice_set`, `collect_location`, `custom_field`, `customer`,
`digital_file`, `event`, `filter_group`, `gift_voucher`, `incomplete_order`,
`newsletter_subscriber`, `offer`, `search`, `order`, `page`, `product`, `bulk_price`, `choice`,
`option`, `variation`, `redirect`, `sale`, `shipping_rate`, `shipping_zone`, `dispute`, `payout`,
`transaction`, `stock`, `theme_assets`, `theme`, `trade_customer_product_price`, `trade_group`,
`voucher`, `webhook`, `wishlist`; offset_limit: `blog_categories`, `blog_posts`, `blog_tags`,
`brands`, `features`, `categories`, `choice_sets`, `collect_locations`, `custom_fields`,
`customers`, `digital_files`, `events`, `filter_groups`, `gift_vouchers`, `incomplete_orders`,
`newsletter_subscribers`, `nexuses`, `offers`, `order_statuses`, `orders`, `pages`,
`payment_methods`, `products`, `search_2`, `bulk_prices`, `choices`, `customization_fields`,
`extras`, `images`, `options`, `reviews`, `variations`, `redirects`, `sales`, `shipping_rates`,
`shipping_zones`, `disputes`, `payouts`, `transactions`, `stock_requests`, `themes`,
`trade_customer_product_prices`, `trade_groups`, `vouchers`, `webhooks`, `wishlists`.

- `app`: GET `/app` - single-object response; records path `.`; computed output fields `id`; emits
  passthrough records.
- `data`: GET `/app/data` - single-object response; records path `.`; computed output fields `id`;
  emits passthrough records.
- `blog_categories`: GET `/blog-categories` - records path `.`; query `fields` from template `{{
  config.fields }}`, omitted when absent; `sort` from template `{{ config.sort }}`, omitted when
  absent; offset/limit pagination; offset parameter `offset`; limit parameter `count`; page size
  100; computed output fields `id`; emits passthrough records.
- `blog_category`: GET `/blog-categories/{{ config.id }}` - single-object response; records path
  `.`; query `fields` from template `{{ config.fields }}`, omitted when absent; computed output
  fields `id`; emits passthrough records.
- `blog_posts`: GET `/blog-posts` - records path `.`; query `embed` from template `{{ config.embed
  }}`, omitted when absent; `fields` from template `{{ config.fields }}`, omitted when absent;
  `sort` from template `{{ config.sort }}`, omitted when absent; offset/limit pagination; offset
  parameter `offset`; limit parameter `count`; page size 100; computed output fields `id`; emits
  passthrough records.
- `blog_post`: GET `/blog-posts/{{ config.id }}` - single-object response; records path `.`; query
  `embed` from template `{{ config.embed }}`, omitted when absent; `fields` from template `{{
  config.fields }}`, omitted when absent; computed output fields `id`; emits passthrough records.
- `blog_tags`: GET `/blog-tags` - records path `.`; query `fields` from template `{{ config.fields
  }}`, omitted when absent; offset/limit pagination; offset parameter `offset`; limit parameter
  `count`; page size 100; computed output fields `id`; emits passthrough records.
- `blog_tag`: GET `/blog-tags/{{ config.id }}` - single-object response; records path `.`; query
  `fields` from template `{{ config.fields }}`, omitted when absent; computed output fields `id`;
  emits passthrough records.
- `brands`: GET `/brands` - records path `.`; query `fields` from template `{{ config.fields }}`,
  omitted when absent; `sort` from template `{{ config.sort }}`, omitted when absent; offset/limit
  pagination; offset parameter `offset`; limit parameter `count`; page size 100; computed output
  fields `id`; emits passthrough records.
- `brand`: GET `/brands/{{ config.id }}` - single-object response; records path `.`; query `fields`
  from template `{{ config.fields }}`, omitted when absent; computed output fields `id`; emits
  passthrough records.
- `business`: GET `/business` - single-object response; records path `.`; query `embed` from
  template `{{ config.embed }}`, omitted when absent; `fields` from template `{{ config.fields }}`,
  omitted when absent; computed output fields `id`; emits passthrough records.
- `features`: GET `/business/features` - records path `.`; query `name` from template `{{
  config.name }}`, omitted when absent; offset/limit pagination; offset parameter `offset`; limit
  parameter `count`; page size 100; computed output fields `id`; emits passthrough records.
- `categories`: GET `/categories` - records path `.`; query `embed` from template `{{ config.embed
  }}`, omitted when absent; `fields` from template `{{ config.fields }}`, omitted when absent;
  `sort` from template `{{ config.sort }}`, omitted when absent; offset/limit pagination; offset
  parameter `offset`; limit parameter `count`; page size 100; computed output fields `id`, `name`,
  `updated_at`.
- `category`: GET `/categories/{{ config.id }}` - single-object response; records path `.`; query
  `embed` from template `{{ config.embed }}`, omitted when absent; computed output fields `id`;
  emits passthrough records.
- `choice_sets`: GET `/choice-sets` - records path `.`; query `fields` from template `{{
  config.fields }}`, omitted when absent; `sort` from template `{{ config.sort }}`, omitted when
  absent; offset/limit pagination; offset parameter `offset`; limit parameter `count`; page size
  100; computed output fields `id`; emits passthrough records.
- `choice_set`: GET `/choice-sets/{{ config.id }}` - single-object response; records path `.`; query
  `fields` from template `{{ config.fields }}`, omitted when absent; computed output fields `id`;
  emits passthrough records.
- `collect_locations`: GET `/collect-locations` - records path `.`; query `fields` from template `{{
  config.fields }}`, omitted when absent; offset/limit pagination; offset parameter `offset`; limit
  parameter `count`; page size 100; computed output fields `id`; emits passthrough records.
- `collect_location`: GET `/collect-locations/{{ config.id }}` - single-object response; records
  path `.`; query `fields` from template `{{ config.fields }}`, omitted when absent; computed output
  fields `id`; emits passthrough records.
- `custom_fields`: GET `/custom-fields` - records path `.`; offset/limit pagination; offset
  parameter `offset`; limit parameter `count`; page size 100; computed output fields `id`; emits
  passthrough records.
- `custom_field`: GET `/custom-fields/{{ config.id }}` - single-object response; records path `.`;
  computed output fields `id`; emits passthrough records.
- `customers`: GET `/customers` - records path `.`; query `email` from template `{{ config.email
  }}`, omitted when absent; `embed` from template `{{ config.embed }}`, omitted when absent;
  `fields` from template `{{ config.fields }}`, omitted when absent; `sort` from template `{{
  config.sort }}`, omitted when absent; `trade` from template `{{ config.trade }}`, omitted when
  absent; offset/limit pagination; offset parameter `offset`; limit parameter `count`; page size
  100; computed output fields `id`, `name`, `updated_at`.
- `customer`: GET `/customers/{{ config.id }}` - single-object response; records path `.`; query
  `embed` from template `{{ config.embed }}`, omitted when absent; `fields` from template `{{
  config.fields }}`, omitted when absent; computed output fields `id`; emits passthrough records.
- `digital_files`: GET `/digital-files` - records path `.`; query `created_after` from template `{{
  config.created_after }}`, omitted when absent; `created_before` from template `{{
  config.created_before }}`, omitted when absent; `fields` from template `{{ config.fields }}`,
  omitted when absent; `since_id` from template `{{ config.since_id }}`, omitted when absent; `sort`
  from template `{{ config.sort }}`, omitted when absent; offset/limit pagination; offset parameter
  `offset`; limit parameter `count`; page size 100; computed output fields `id`; emits passthrough
  records.
- `digital_file`: GET `/digital-files/{{ config.id }}` - single-object response; records path `.`;
  query `fields` from template `{{ config.fields }}`, omitted when absent; computed output fields
  `id`; emits passthrough records.
- `events`: GET `/events` - records path `.`; query `created_after` from template `{{
  config.created_after }}`, omitted when absent; `created_before` from template `{{
  config.created_before }}`, omitted when absent; `fields` from template `{{ config.fields }}`,
  omitted when absent; `since_id` from template `{{ config.since_id }}`, omitted when absent;
  `subject_id` from template `{{ config.subject_id }}`, omitted when absent; `subject_type` from
  template `{{ config.subject_type }}`, omitted when absent; `topic` from template `{{ config.topic
  }}`, omitted when absent; offset/limit pagination; offset parameter `offset`; limit parameter
  `count`; page size 100; computed output fields `id`; emits passthrough records.
- `event`: GET `/events/{{ config.id }}` - single-object response; records path `.`; computed output
  fields `id`; emits passthrough records.
- `filter_groups`: GET `/filter-groups` - records path `.`; query `fields` from template `{{
  config.fields }}`, omitted when absent; `sort` from template `{{ config.sort }}`, omitted when
  absent; offset/limit pagination; offset parameter `offset`; limit parameter `count`; page size
  100; computed output fields `id`; emits passthrough records.
- `filter_group`: GET `/filter-groups/{{ config.id }}` - single-object response; records path `.`;
  query `fields` from template `{{ config.fields }}`, omitted when absent; computed output fields
  `id`; emits passthrough records.
- `gift_vouchers`: GET `/gift-vouchers` - records path `.`; query `fields` from template `{{
  config.fields }}`, omitted when absent; offset/limit pagination; offset parameter `offset`; limit
  parameter `count`; page size 100; computed output fields `id`; emits passthrough records.
- `gift_voucher`: GET `/gift-vouchers/{{ config.id }}` - single-object response; records path `.`;
  computed output fields `id`; emits passthrough records.
- `incomplete_orders`: GET `/incomplete-orders` - records path `.`; query `embed` from template `{{
  config.embed }}`, omitted when absent; `fields` from template `{{ config.fields }}`, omitted when
  absent; `from` from template `{{ config.from }}`, omitted when absent; `sort` from template `{{
  config.sort }}`, omitted when absent; offset/limit pagination; offset parameter `offset`; limit
  parameter `count`; page size 100; computed output fields `id`; emits passthrough records.
- `incomplete_order`: GET `/incomplete-orders/{{ config.id }}` - single-object response; records
  path `.`; query `embed` from template `{{ config.embed }}`, omitted when absent; computed output
  fields `id`; emits passthrough records.
- `newsletter_subscribers`: GET `/newsletter-subscribers` - records path `.`; query `fields` from
  template `{{ config.fields }}`, omitted when absent; `sort` from template `{{ config.sort }}`,
  omitted when absent; offset/limit pagination; offset parameter `offset`; limit parameter `count`;
  page size 100; computed output fields `id`; emits passthrough records.
- `newsletter_subscriber`: GET `/newsletter-subscribers/{{ config.id }}` - single-object response;
  records path `.`; query `fields` from template `{{ config.fields }}`, omitted when absent;
  computed output fields `id`; emits passthrough records.
- `nexuses`: GET `/nexuses` - records path `.`; query `fields` from template `{{ config.fields }}`,
  omitted when absent; offset/limit pagination; offset parameter `offset`; limit parameter `count`;
  page size 100; computed output fields `id`; emits passthrough records.
- `offers`: GET `/offers` - records path `.`; query `embed` from template `{{ config.embed }}`,
  omitted when absent; `fields` from template `{{ config.fields }}`, omitted when absent; `sort`
  from template `{{ config.sort }}`, omitted when absent; offset/limit pagination; offset parameter
  `offset`; limit parameter `count`; page size 100; computed output fields `id`; emits passthrough
  records.
- `offer`: GET `/offers/{{ config.id }}` - single-object response; records path `.`; query `embed`
  from template `{{ config.embed }}`, omitted when absent; `fields` from template `{{ config.fields
  }}`, omitted when absent; computed output fields `id`; emits passthrough records.
- `order_statuses`: GET `/order-statuses` - records path `.`; query `fields` from template `{{
  config.fields }}`, omitted when absent; offset/limit pagination; offset parameter `offset`; limit
  parameter `count`; page size 100; computed output fields `id`; emits passthrough records.
- `orders`: GET `/orders` - records path `.`; query `archived` from template `{{ config.archived
  }}`, omitted when absent; `embed` from template `{{ config.embed }}`, omitted when absent;
  `fields` from template `{{ config.fields }}`, omitted when absent; `from` from template `{{
  config.from }}`, omitted when absent; `ids` from template `{{ config.ids }}`, omitted when absent;
  `sort` from template `{{ config.sort }}`, omitted when absent; `status` from template `{{
  config.status }}`, omitted when absent; `to` from template `{{ config.to }}`, omitted when absent;
  offset/limit pagination; offset parameter `offset`; limit parameter `count`; page size 100;
  computed output fields `id`, `name`, `updated_at`.
- `search`: GET `/orders/search` - single-object response; records path `.`; query `keywords` from
  template `{{ config.keywords }}`, omitted when absent; `session_id` from template `{{
  config.session_id }}`, omitted when absent; `sort` from template `{{ config.sort }}`, omitted when
  absent; computed output fields `id`; emits passthrough records.
- `order`: GET `/orders/{{ config.id }}` - single-object response; records path `.`; query `embed`
  from template `{{ config.embed }}`, omitted when absent; `fields` from template `{{ config.fields
  }}`, omitted when absent; computed output fields `id`; emits passthrough records.
- `pages`: GET `/pages` - records path `.`; query `fields` from template `{{ config.fields }}`,
  omitted when absent; `sort` from template `{{ config.sort }}`, omitted when absent; offset/limit
  pagination; offset parameter `offset`; limit parameter `count`; page size 100; computed output
  fields `id`; emits passthrough records.
- `page`: GET `/pages/{{ config.id }}` - single-object response; records path `.`; query `fields`
  from template `{{ config.fields }}`, omitted when absent; computed output fields `id`; emits
  passthrough records.
- `payment_methods`: GET `/payment-methods` - records path `.`; query `fields` from template `{{
  config.fields }}`, omitted when absent; offset/limit pagination; offset parameter `offset`; limit
  parameter `count`; page size 100; computed output fields `id`; emits passthrough records.
- `products`: GET `/products` - records path `.`; query `active` from template `{{ config.active
  }}`, omitted when absent; `embed` from template `{{ config.embed }}`, omitted when absent;
  `fields` from template `{{ config.fields }}`, omitted when absent; `sort` from template `{{
  config.sort }}`, omitted when absent; offset/limit pagination; offset parameter `offset`; limit
  parameter `count`; page size 100; computed output fields `id`, `name`, `updated_at`.
- `search_2`: GET `/products/search` - records path `.`; query `embed` from template `{{
  config.embed }}`, omitted when absent; `fields` from template `{{ config.fields }}`, omitted when
  absent; `keywords` from template `{{ config.keywords }}`, omitted when absent; `session_id` from
  template `{{ config.session_id }}`, omitted when absent; `sort` from template `{{ config.sort }}`,
  omitted when absent; offset/limit pagination; offset parameter `offset`; limit parameter `count`;
  page size 100; computed output fields `id`; emits passthrough records.
- `product`: GET `/products/{{ config.id }}` - single-object response; records path `.`; query
  `embed` from template `{{ config.embed }}`, omitted when absent; `fields` from template `{{
  config.fields }}`, omitted when absent; computed output fields `id`; emits passthrough records.
- `bulk_prices`: GET `/products/{{ config.id }}/bulk-prices` - records path `.`; offset/limit
  pagination; offset parameter `offset`; limit parameter `count`; page size 100; computed output
  fields `id`; emits passthrough records.
- `bulk_price`: GET `/products/{{ config.product_id }}/bulk-prices/{{ config.id }}` - single-object
  response; records path `.`; computed output fields `id`; emits passthrough records.
- `choices`: GET `/products/{{ config.product_id }}/choices` - records path `.`; offset/limit
  pagination; offset parameter `offset`; limit parameter `count`; page size 100; computed output
  fields `id`; emits passthrough records.
- `choice`: GET `/products/{{ config.product_id }}/choices/{{ config.id }}` - single-object
  response; records path `.`; computed output fields `id`; emits passthrough records.
- `customization_fields`: GET `/products/{{ config.product_id }}/customization-fields` - records
  path `.`; query `fields` from template `{{ config.fields }}`, omitted when absent; offset/limit
  pagination; offset parameter `offset`; limit parameter `count`; page size 100; computed output
  fields `id`; emits passthrough records.
- `extras`: GET `/products/{{ config.product_id }}/extras` - records path `.`; query `fields` from
  template `{{ config.fields }}`, omitted when absent; offset/limit pagination; offset parameter
  `offset`; limit parameter `count`; page size 100; computed output fields `id`; emits passthrough
  records.
- `images`: GET `/products/{{ config.product_id }}/images` - records path `.`; query `fields` from
  template `{{ config.fields }}`, omitted when absent; offset/limit pagination; offset parameter
  `offset`; limit parameter `count`; page size 100; computed output fields `id`; emits passthrough
  records.
- `options`: GET `/products/{{ config.product_id }}/options` - records path `.`; query `embed` from
  template `{{ config.embed }}`, omitted when absent; offset/limit pagination; offset parameter
  `offset`; limit parameter `count`; page size 100; computed output fields `id`; emits passthrough
  records.
- `option`: GET `/products/{{ config.product_id }}/options/{{ config.id }}` - single-object
  response; records path `.`; query `embed` from template `{{ config.embed }}`, omitted when absent;
  computed output fields `id`; emits passthrough records.
- `reviews`: GET `/products/{{ config.product_id }}/reviews` - records path `.`; query `fields` from
  template `{{ config.fields }}`, omitted when absent; `sort` from template `{{ config.sort }}`,
  omitted when absent; offset/limit pagination; offset parameter `offset`; limit parameter `count`;
  page size 100; computed output fields `id`; emits passthrough records.
- `variations`: GET `/products/{{ config.product_id }}/variations` - records path `.`; query
  `fields` from template `{{ config.fields }}`, omitted when absent; offset/limit pagination; offset
  parameter `offset`; limit parameter `count`; page size 100; computed output fields `id`; emits
  passthrough records.
- `variation`: GET `/products/{{ config.product_id }}/variations/{{ config.id }}` - single-object
  response; records path `.`; query `fields` from template `{{ config.fields }}`, omitted when
  absent; computed output fields `id`; emits passthrough records.
- `redirects`: GET `/redirects` - records path `.`; query `fields` from template `{{ config.fields
  }}`, omitted when absent; `path` from template `{{ config.path }}`, omitted when absent;
  offset/limit pagination; offset parameter `offset`; limit parameter `count`; page size 100;
  computed output fields `id`; emits passthrough records.
- `redirect`: GET `/redirects/{{ config.id }}` - single-object response; records path `.`; query
  `fields` from template `{{ config.fields }}`, omitted when absent; computed output fields `id`;
  emits passthrough records.
- `sales`: GET `/sales` - records path `.`; query `fields` from template `{{ config.fields }}`,
  omitted when absent; offset/limit pagination; offset parameter `offset`; limit parameter `count`;
  page size 100; computed output fields `id`; emits passthrough records.
- `sale`: GET `/sales/{{ config.id }}` - single-object response; records path `.`; query `fields`
  from template `{{ config.fields }}`, omitted when absent; computed output fields `id`; emits
  passthrough records.
- `shipping_rates`: GET `/shipping-rates` - records path `.`; query `embed` from template `{{
  config.embed }}`, omitted when absent; `fields` from template `{{ config.fields }}`, omitted when
  absent; offset/limit pagination; offset parameter `offset`; limit parameter `count`; page size
  100; computed output fields `id`; emits passthrough records.
- `shipping_rate`: GET `/shipping-rates/{{ config.id }}` - single-object response; records path `.`;
  query `embed` from template `{{ config.embed }}`, omitted when absent; `fields` from template `{{
  config.fields }}`, omitted when absent; computed output fields `id`; emits passthrough records.
- `shipping_zones`: GET `/shipping-zones` - records path `.`; query `embed` from template `{{
  config.embed }}`, omitted when absent; `fields` from template `{{ config.fields }}`, omitted when
  absent; offset/limit pagination; offset parameter `offset`; limit parameter `count`; page size
  100; computed output fields `id`; emits passthrough records.
- `shipping_zone`: GET `/shipping-zones/{{ config.country_id }}` - single-object response; records
  path `.`; query `embed` from template `{{ config.embed }}`, omitted when absent; `fields` from
  template `{{ config.fields }}`, omitted when absent; computed output fields `country_id`, `id`;
  emits passthrough records.
- `disputes`: GET `/shopwired-payments/disputes` - records path `.`; query `fields` from template
  `{{ config.fields }}`, omitted when absent; `sort` from template `{{ config.sort }}`, omitted when
  absent; offset/limit pagination; offset parameter `offset`; limit parameter `count`; page size
  100; computed output fields `id`; emits passthrough records.
- `dispute`: GET `/shopwired-payments/disputes/{{ config.dispute_id }}` - single-object response;
  records path `.`; query `fields` from template `{{ config.fields }}`, omitted when absent;
  computed output fields `dispute_id`, `id`; emits passthrough records.
- `payouts`: GET `/shopwired-payments/payouts` - records path `.`; query `fields` from template `{{
  config.fields }}`, omitted when absent; `sort` from template `{{ config.sort }}`, omitted when
  absent; offset/limit pagination; offset parameter `offset`; limit parameter `count`; page size
  100; computed output fields `id`; emits passthrough records.
- `payout`: GET `/shopwired-payments/payouts/{{ config.payout_id }}` - single-object response;
  records path `.`; query `fields` from template `{{ config.fields }}`, omitted when absent;
  computed output fields `id`, `payout_id`; emits passthrough records.
- `transactions`: GET `/shopwired-payments/transactions` - records path `.`; query `fields` from
  template `{{ config.fields }}`, omitted when absent; `sort` from template `{{ config.sort }}`,
  omitted when absent; offset/limit pagination; offset parameter `offset`; limit parameter `count`;
  page size 100; computed output fields `id`; emits passthrough records.
- `transaction`: GET `/shopwired-payments/transactions/{{ config.transaction_id }}` - single-object
  response; records path `.`; query `fields` from template `{{ config.fields }}`, omitted when
  absent; computed output fields `id`, `transaction_id`; emits passthrough records.
- `stock`: GET `/stock` - single-object response; records path `.`; query `sku` from template `{{
  config.sku }}`, omitted when absent; computed output fields `id`; emits passthrough records.
- `stock_requests`: GET `/stock-requests` - records path `.`; query `fields` from template `{{
  config.fields }}`, omitted when absent; `sort` from template `{{ config.sort }}`, omitted when
  absent; offset/limit pagination; offset parameter `offset`; limit parameter `count`; page size
  100; computed output fields `id`; emits passthrough records.
- `theme_assets`: GET `/theme-assets/{{ config.theme_id }}` - single-object response; records path
  `.`; query `fields` from template `{{ config.fields }}`, omitted when absent; `key` from template
  `{{ config.key }}`, omitted when absent; computed output fields `id`, `theme_id`; emits
  passthrough records.
- `themes`: GET `/themes` - records path `.`; query `fields` from template `{{ config.fields }}`,
  omitted when absent; `sort` from template `{{ config.sort }}`, omitted when absent; offset/limit
  pagination; offset parameter `offset`; limit parameter `count`; page size 100; computed output
  fields `id`; emits passthrough records.
- `theme`: GET `/themes/{{ config.id }}` - single-object response; records path `.`; query `fields`
  from template `{{ config.fields }}`, omitted when absent; computed output fields `id`; emits
  passthrough records.
- `trade_customer_product_prices`: GET `/trade-customer-product-prices` - records path `.`; query
  `customer_id` from template `{{ config.customer_id }}`, omitted when absent; `fields` from
  template `{{ config.fields }}`, omitted when absent; `product_id` from template `{{
  config.product_id }}`, omitted when absent; offset/limit pagination; offset parameter `offset`;
  limit parameter `count`; page size 100; computed output fields `id`; emits passthrough records.
- `trade_customer_product_price`: GET `/trade-customer-product-prices/{{ config.id }}` -
  single-object response; records path `.`; computed output fields `id`; emits passthrough records.
- `trade_groups`: GET `/trade-groups` - records path `.`; query `embed` from template `{{
  config.embed }}`, omitted when absent; `fields` from template `{{ config.fields }}`, omitted when
  absent; offset/limit pagination; offset parameter `offset`; limit parameter `count`; page size
  100; computed output fields `id`; emits passthrough records.
- `trade_group`: GET `/trade-groups/{{ config.id }}` - single-object response; records path `.`;
  query `fields` from template `{{ config.fields }}`, omitted when absent; computed output fields
  `id`; emits passthrough records.
- `vouchers`: GET `/vouchers` - records path `.`; query `code` from template `{{ config.code }}`,
  omitted when absent; `embed` from template `{{ config.embed }}`, omitted when absent; `fields`
  from template `{{ config.fields }}`, omitted when absent; `sort` from template `{{ config.sort
  }}`, omitted when absent; offset/limit pagination; offset parameter `offset`; limit parameter
  `count`; page size 100; computed output fields `id`; emits passthrough records.
- `voucher`: GET `/vouchers/{{ config.id }}` - single-object response; records path `.`; query
  `fields` from template `{{ config.fields }}`, omitted when absent; computed output fields `id`;
  emits passthrough records.
- `webhooks`: GET `/webhooks` - records path `.`; query `fields` from template `{{ config.fields
  }}`, omitted when absent; `sort` from template `{{ config.sort }}`, omitted when absent;
  offset/limit pagination; offset parameter `offset`; limit parameter `count`; page size 100;
  computed output fields `id`; emits passthrough records.
- `webhook`: GET `/webhooks/{{ config.id }}` - single-object response; records path `.`; query
  `fields` from template `{{ config.fields }}`, omitted when absent; computed output fields `id`;
  emits passthrough records.
- `wishlists`: GET `/wishlists` - records path `.`; query `customer_id` from template `{{
  config.customer_id }}`, omitted when absent; `embed` from template `{{ config.embed }}`, omitted
  when absent; `fields` from template `{{ config.fields }}`, omitted when absent; offset/limit
  pagination; offset parameter `offset`; limit parameter `count`; page size 100; computed output
  fields `id`; emits passthrough records.
- `wishlist`: GET `/wishlists/{{ config.id }}` - single-object response; records path `.`; query
  `embed` from template `{{ config.embed }}`, omitted when absent; `fields` from template `{{
  config.fields }}`, omitted when absent; computed output fields `id`; emits passthrough records.

## Write actions & risks

Overall write risk: external ShopWired API mutations that create, update, delete, refund, dispatch,
verify, modify, or otherwise alter store data.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `deletes_all_app_data`: DELETE `/app/data` - kind `delete`; body type `none`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Deletes all app data in
  ShopWired.
- `create_new_app_data`: POST `/app/data` - kind `create`; body type `json`; risk: Create new app
  data in ShopWired.
- `create_blog_category`: POST `/blog-categories` - kind `create`; body type `json`; required record
  fields `title`; accepted fields `slug`, `title`; risk: Create a new blog category in ShopWired.
- `delete_blog_category`: DELETE `/blog-categories/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Delete a blog category in
  ShopWired.
- `update_blog_category`: PUT `/blog-categories/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `id`, `slug`, `title`; risk: Update
  a blog category in ShopWired.
- `create_blog_post`: POST `/blog-posts` - kind `create`; body type `json`; required record fields
  `title`, `slug`; accepted fields `active`, `categoryId`, `categoryTitle`, `content`, `customUrl`,
  `excerpt`, `image`, `metaDescription`, `metaKeywords`, `metaTitle`, `slug`, `tags`, `title`; risk:
  Create a new blog post in ShopWired.
- `delete_blog_post`: DELETE `/blog-posts/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: Delete a blog post in ShopWired.
- `update_blog_post`: PUT `/blog-posts/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `title`, `slug`; accepted fields `active`, `categoryId`,
  `categoryTitle`, `content`, `customUrl`, `excerpt`, `id`, `image`, `metaDescription`,
  `metaKeywords`, `metaTitle`, `slug`, `tags`, `title`; risk: Update a blog post in ShopWired.
- `create_blog_tag`: POST `/blog-tags` - kind `create`; body type `json`; required record fields
  `title`; accepted fields `slug`, `title`; risk: Create a new blog tag in ShopWired.
- `delete_blog_tag`: DELETE `/blog-tags/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: Delete a blog tag in ShopWired.
- `update_blog_tag`: PUT `/blog-tags/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `id`, `slug`, `title`; risk: Update a blog tag
  in ShopWired.
- `create_brand`: POST `/brands` - kind `create`; body type `json`; required record fields `title`;
  accepted fields `active`, `description`, `image`, `metaDescription`, `metaKeywords`, `metaTitle`,
  `slug`, `title`; risk: Create a new brand in ShopWired.
- `delete_brand`: DELETE `/brands/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: Delete a brand in ShopWired.
- `update_brand`: PUT `/brands/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `active`, `description`, `id`, `image`,
  `metaDescription`, `metaKeywords`, `metaTitle`, `slug`, `title`; risk: Update a brand in
  ShopWired.
- `change_business_feature_status`: POST `/business/features/change-status` - kind `create`; body
  type `json`; required record fields `name`, `status`; accepted fields `name`, `status`; risk:
  Change business feature status in ShopWired.
- `create_category`: POST `/categories` - kind `create`; body type `json`; required record fields
  `title`; accepted fields `active`, `description`, `description2`, `image`, `metaDescription`,
  `metaKeywords`, `metaTitle`, `parents`, `slug`, `title`; risk: Create a new category in ShopWired.
- `delete_category_by_id`: DELETE `/categories/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Delete a category in ShopWired.
- `update_category_by_id`: PUT `/categories/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `title`; accepted fields `active`, `description`,
  `description2`, `id`, `image`, `metaDescription`, `metaKeywords`, `metaTitle`, `parents`, `slug`,
  `title`; risk: Update a category in ShopWired.
- `create_choice_set_value`: POST `/choice-set-values` - kind `create`; body type `json`; required
  record fields `set`, `name`; accepted fields `name`, `price`, `set`, `sortOrder`; risk: Create a
  choice set value in ShopWired.
- `delete_choice_set_value`: DELETE `/choice-set-values/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Delete a choice set value
  in ShopWired.
- `update_choice_set_value`: PUT `/choice-set-values/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `id`, `name`, `price`,
  `sortOrder`; risk: Update a choice set value in ShopWired.
- `create_choice_set`: POST `/choice-sets` - kind `create`; body type `json`; required record fields
  `displayName`, `internalName`; accepted fields `displayName`, `internalName`; risk: Create a
  choice set in ShopWired.
- `delete_choice_set`: DELETE `/choice-sets/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: Delete a choice set in ShopWired.
- `update_choice_set`: PUT `/choice-sets/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `displayName`, `id`, `internalName`;
  risk: Update a choice set in ShopWired.
- `create_custom_field`: POST `/custom-fields` - kind `create`; body type `json`; required record
  fields `name`, `itemType`; accepted fields `allowedValues`, `itemType`, `label`, `name`,
  `sortOrder`, `type`; risk: Create a custom field in ShopWired.
- `delete_custom_field`: DELETE `/custom-fields/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Delete a custom field in ShopWired.
- `update_custom_field`: PUT `/custom-fields/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `allowedValues`, `id`, `label`,
  `name`, `sortOrder`, `type`; risk: Update a custom field in ShopWired.
- `create_customer`: POST `/customers` - kind `create`; body type `json`; required record fields
  `firstName`, `lastName`, `email`, `password`; accepted fields `acceptsMarketing`, `active`,
  `additionalData`, `addressLine1`, `addressLine2`, `addressLine3`, `city`, `companyName`,
  `costPriceMultiplier`, `country`, `credit`, `discount`, `email`, `firstName`, `lastName`,
  `mobilePhone`, `notes`, `password`, and 9 more; risk: Create a new customer in ShopWired.
- `create_digital_file`: POST `/digital-files` - kind `create`; body type `json`; required record
  fields `name`, `extension`, `sourceUrl`; accepted fields `extension`, `name`, `sourceUrl`, `tag`;
  risk: Create a new digital file in ShopWired.
- `delete_digital_file`: DELETE `/digital-files/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Delete a digital file in ShopWired.
- `update_digital_file`: PUT `/digital-files/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `id`, `name`, `sourceUrl`, `tag`;
  risk: Update a digital file in ShopWired.
- `create_filter_group`: POST `/filter-groups` - kind `create`; body type `json`; required record
  fields `title`; accepted fields `title`; risk: Create a new filter group in ShopWired.
- `delete_filter_group`: DELETE `/filter-groups/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Delete a filter group in ShopWired.
- `update_filter_group`: PUT `/filter-groups/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`, `title`; accepted fields `id`, `title`; risk:
  Update a filter group in ShopWired.
- `create_gift_card`: POST `/gift-vouchers` - kind `create`; body type `json`; required record
  fields `code`, `amount`, `amountUsed`; accepted fields `amount`, `amountUsed`, `code`; risk:
  Create a gift card in ShopWired.
- `delete_gift_card`: DELETE `/gift-vouchers/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Delete a gift card in ShopWired.
- `update_gift_card`: PUT `/gift-vouchers/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `amount`, `amountUsed`, `code`, `id`;
  risk: Update a gift card in ShopWired.
- `create_business_nexus`: POST `/nexuses` - kind `create`; body type `json`; required record fields
  `countryId`, `stateId`, `type`, `name`, `addressLine1`, `city`, `postcode`; accepted fields
  `addressLine1`, `addressLine2`, `city`, `countryId`, `name`, `postcode`, `salesTaxId`, `stateId`,
  `type`; risk: Create a new business nexus item in ShopWired.
- `delete_business_nexus`: DELETE `/nexuses/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: Delete a business nexus item in ShopWired.
- `update_business_nexus`: PUT `/nexuses/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `addressLine1`, `addressLine2`, `city`,
  `id`, `name`, `postcode`, `salesTaxId`; risk: Update a business nexus item in ShopWired.
- `create_offer`: POST `/offers` - kind `create`; body type `json`; required record fields `title`,
  `itemCount`; accepted fields `active`, `description`, `discountAmount`, `discountType`, `endDate`,
  `excludedProducts`, `itemCount`, `itemSortOrder`, `maximumDiscountAmount`,
  `productShippingEndDate`, `productShippingStartDate`, `startDate`, `supportsPreOrderProducts`,
  `targets`, `title`; risk: Create a new offer in ShopWired.
- `delete_offer`: DELETE `/offers/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: Delete an offer in ShopWired.
- `update_offer`: PUT `/offers/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `active`, `description`, `discountAmount`,
  `discountType`, `endDate`, `excludedProducts`, `id`, `itemCount`, `itemSortOrder`,
  `maximumDiscountAmount`, `productShippingEndDate`, `productShippingStartDate`, `startDate`,
  `supportsPreOrderProducts`, `targets`, `title`; risk: Update an offer in ShopWired.
- `create_order_status`: POST `/order-statuses` - kind `create`; body type `json`; required record
  fields `name`, `sortOrder`; accepted fields `name`, `sortOrder`; risk: Create an order status in
  ShopWired.
- `delete_order_status`: DELETE `/order-statuses/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Delete an order status in ShopWired.
- `create_order`: POST `/orders` - kind `create`; body type `json`; required record fields `status`,
  `billingAddress`, `shippingAddress`, `shippingRate`, `products`; accepted fields
  `acceptsMarketing`, `archived`, `billingAddress`, `comments`, `customer`, `deliveryDate`,
  `discounts`, `earnedRewardPoints`, `externalFulfillment`, `fees`, `originalShippingTotal`,
  `partialPaymentTotal`, `paymentMethod`, `postcode`, `products`, `reference`, `refunds`,
  `sendEmails`, and 10 more; risk: Create a new order in ShopWired.
- `delete_order`: DELETE `/orders/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: Delete an order in ShopWired.
- `create_an_order_admin_comment`: POST `/orders/{{ record.id }}/comments` - kind `create`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `content`, `id`; risk:
  Create an order admin comment in ShopWired.
- `delete_an_order_admin_comment`: DELETE `/orders/{{ record.id }}/comments/{{ record.comment_id }}`
  - kind `delete`; body type `none`; path fields `id`, `comment_id`; required record fields `id`,
  `comment_id`; accepted fields `comment_id`, `id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: Delete an order admin comment in ShopWired.
- `request_a_pre_order_be_dispatched`: POST `/orders/{{ record.id }}/dispatch-pre-order` - kind
  `create`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `charge`, `id`; confirmation `destructive`; risk: Request a pre-order be dispatched in ShopWired.
- `update_an_order_s_status`: POST `/orders/{{ record.id }}/status` - kind `create`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields
  `eBayShipmentTrackingNumber`, `eBayShippingCarrier`, `id`, `sendEmail`, `sendToEbay`, `status`,
  `trackingUrl`; risk: Update an order's status in ShopWired.
- `create_refund`: POST `/orders/{{ record.order_id }}/refunds` - kind `create`; body type `json`;
  path fields `order_id`; required record fields `order_id`, `amount`, `comment`; accepted fields
  `amount`, `comment`, `order_id`; confirmation `destructive`; risk: Create a refund for an order in
  ShopWired.
- `create_page`: POST `/pages` - kind `create`; body type `json`; required record fields `title`;
  accepted fields `content`, `customContent`, `metaDescription`, `metaKeywords`, `metaTitle`,
  `slug`, `title`; risk: Create a new page in ShopWired.
- `delete_page`: DELETE `/pages/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: Delete a page in ShopWired.
- `update_page`: PUT `/pages/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`; accepted fields `content`, `customContent`, `id`, `metaDescription`,
  `metaKeywords`, `metaTitle`, `slug`, `title`; risk: Update a page in ShopWired.
- `create_product`: POST `/products` - kind `create`; body type `json`; required record fields
  `title`; accepted fields `active`, `additionalData`, `brand`, `categories`, `choices`,
  `comparePrice`, `customFields`, `customizationFields`, `deliveryPrice`, `description`,
  `description2`, `description3`, `description4`, `description5`, `digitalFiles`, `eBayBestOffer`,
  `eBayCategory`, `eBayShippingRates`, and 36 more; risk: Create a new product in ShopWired.
- `update_prices`: POST `/products/prices` - kind `create`; body type `json`; accepted fields
  `items`, `price`, `salePrice`, `sendToEbay`, `sku`; risk: Update product prices in ShopWired.
- `delete_product`: DELETE `/products/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: Delete a product in ShopWired.
- `update_product`: PUT `/products/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `active`, `additionalData`, `brand`,
  `categories`, `choices`, `comparePrice`, `customFields`, `customizationFields`, `deliveryPrice`,
  `description`, `description2`, `description3`, `description4`, `description5`, `digitalFiles`,
  `eBayBestOffer`, `eBayCategory`, `eBayShippingRates`, and 37 more; risk: Update a product in
  ShopWired.
- `create_bulk_price`: POST `/products/{{ record.id }}/bulk-prices` - kind `create`; body type
  `json`; path fields `id`; required record fields `id`, `fromQuantity`, `toQuantity`, `price`;
  accepted fields `fromQuantity`, `id`, `price`, `toQuantity`, `variationId`; risk: Create a bulk
  price for a product in ShopWired.
- `delete_bulk_price`: DELETE `/products/{{ record.product_id }}/bulk-prices/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `product_id`, `id`; required record fields `product_id`,
  `id`; accepted fields `id`, `product_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Delete a specific bulk price in ShopWired.
- `update_bulk_price`: PUT `/products/{{ record.product_id }}/bulk-prices/{{ record.id }}` - kind
  `update`; body type `json`; path fields `product_id`, `id`; required record fields `product_id`,
  `id`; accepted fields `fromQuantity`, `id`, `price`, `product_id`, `toQuantity`; risk: Update a
  specific bulk price in ShopWired.
- `assign_product_choice`: POST `/products/{{ record.product_id }}/choices` - kind `create`; body
  type `json`; path fields `product_id`; required record fields `product_id`, `set`, `value`;
  accepted fields `costPrice`, `customPrice`, `product_id`, `set`, `value`; risk: Assign a choice to
  a product in ShopWired.
- `delete_product_choice`: DELETE `/products/{{ record.product_id }}/choices/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `product_id`, `id`; required record fields `product_id`,
  `id`; accepted fields `id`, `product_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Delete a product choice in ShopWired.
- `update_product_choice`: PUT `/products/{{ record.product_id }}/choices/{{ record.id }}` - kind
  `update`; body type `json`; path fields `product_id`, `id`; required record fields `product_id`,
  `id`; accepted fields `costPrice`, `customPrice`, `id`, `product_id`; risk: Update a product
  choice in ShopWired.
- `create_product_customization_field`: POST `/products/{{ record.product_id
  }}/customization-fields` - kind `create`; body type `json`; path fields `product_id`; required
  record fields `product_id`, `label`, `type`; accepted fields `label`, `max_length`, `product_id`,
  `type`; risk: Create a product customisation field in ShopWired.
- `delete_product_customization_field`: DELETE `/products/{{ record.product_id
  }}/customization-fields/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `product_id`, `id`; required record fields `product_id`, `id`; accepted fields `id`, `product_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Delete a
  product customisation field in ShopWired.
- `update_product_customization_field`: PUT `/products/{{ record.product_id
  }}/customization-fields/{{ record.id }}` - kind `update`; body type `json`; path fields
  `product_id`, `id`; required record fields `product_id`, `id`; accepted fields `id`, `label`,
  `max_length`, `product_id`, `type`; risk: Update a product customisation field in ShopWired.
- `create_product_extra`: POST `/products/{{ record.product_id }}/extras` - kind `create`; body type
  `json`; path fields `product_id`; required record fields `product_id`, `name`, `price`, `sku`;
  accepted fields `costPrice`, `name`, `price`, `product_id`, `sku`; risk: Create a product extra in
  ShopWired.
- `delete_product_extra`: DELETE `/products/{{ record.product_id }}/extras/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `product_id`, `id`; required record fields `product_id`,
  `id`; accepted fields `id`, `product_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Delete a product extra in ShopWired.
- `update_product_extra`: PUT `/products/{{ record.product_id }}/extras/{{ record.id }}` - kind
  `update`; body type `json`; path fields `product_id`, `id`; required record fields `product_id`,
  `id`; accepted fields `costPrice`, `id`, `name`, `price`, `product_id`, `sku`; risk: Update a
  product extra in ShopWired.
- `create_product_image`: POST `/products/{{ record.product_id }}/images` - kind `create`; body type
  `json`; path fields `product_id`; required record fields `product_id`, `image`; accepted fields
  `description`, `image`, `imageName`, `product_id`, `sortOrder`; risk: Create a product image in
  ShopWired.
- `delete_product_image`: DELETE `/products/{{ record.product_id }}/images/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `product_id`, `id`; required record fields `product_id`,
  `id`; accepted fields `id`, `product_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Delete a product image in ShopWired.
- `update_product_image`: PUT `/products/{{ record.product_id }}/images/{{ record.id }}` - kind
  `update`; body type `json`; path fields `product_id`, `id`; required record fields `product_id`,
  `id`; accepted fields `description`, `id`, `product_id`, `sortOrder`; risk: Update a product image
  in ShopWired.
- `create_product_option_value`: POST `/products/{{ record.product_id }}/option-values` - kind
  `create`; body type `json`; path fields `product_id`; required record fields `product_id`, `name`,
  `option`; accepted fields `name`, `option`, `product_id`, `sortOrder`; risk: Create a product
  option value in ShopWired.
- `delete_product_option_value`: DELETE `/products/{{ record.product_id }}/option-values/{{
  record.id }}` - kind `delete`; body type `none`; path fields `product_id`, `id`; required record
  fields `product_id`, `id`; accepted fields `id`, `product_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: Delete a product option value in ShopWired.
- `update_product_option_value`: PUT `/products/{{ record.product_id }}/option-values/{{ record.id
  }}` - kind `update`; body type `json`; path fields `product_id`, `id`; required record fields
  `product_id`, `id`; accepted fields `id`, `name`, `product_id`, `sortOrder`; risk: Update a
  product option value in ShopWired.
- `create_product_option`: POST `/products/{{ record.product_id }}/options` - kind `create`; body
  type `json`; path fields `product_id`; required record fields `product_id`, `name`; accepted
  fields `name`, `product_id`, `sortOrder`; risk: Create a product option in ShopWired.
- `delete_product_option`: DELETE `/products/{{ record.product_id }}/options/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `product_id`, `id`; required record fields `product_id`,
  `id`; accepted fields `id`, `product_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Delete a product option in ShopWired.
- `update_product_option`: PUT `/products/{{ record.product_id }}/options/{{ record.id }}` - kind
  `update`; body type `json`; path fields `product_id`, `id`; required record fields `product_id`,
  `id`; accepted fields `id`, `name`, `product_id`, `sortOrder`; risk: Update a product option in
  ShopWired.
- `create_product_review`: POST `/products/{{ record.product_id }}/reviews` - kind `create`; body
  type `json`; path fields `product_id`; required record fields `product_id`, `name`, `content`,
  `rating`; accepted fields `active`, `content`, `name`, `product_id`, `rating`; risk: Create a
  product review in ShopWired.
- `delete_product_review`: DELETE `/products/{{ record.product_id }}/reviews/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `product_id`, `id`; required record fields `product_id`,
  `id`; accepted fields `id`, `product_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: Delete a product review in ShopWired.
- `update_product_review`: PUT `/products/{{ record.product_id }}/reviews/{{ record.id }}` - kind
  `update`; body type `json`; path fields `product_id`, `id`; required record fields `product_id`,
  `id`; accepted fields `active`, `content`, `id`, `name`, `product_id`, `rating`; risk: Update a
  product review in ShopWired.
- `create_product_variation`: POST `/products/{{ record.product_id }}/variations` - kind `create`;
  body type `json`; path fields `product_id`; required record fields `product_id`, `values`;
  accepted fields `costPrice`, `gtin`, `image`, `mpn`, `price`, `product_id`, `rewardPoints`,
  `salePrice`, `sku`, `stock`, `values`, `weight`; risk: Create a new product variation in
  ShopWired.
- `delete_product_variation`: DELETE `/products/{{ record.product_id }}/variations/{{ record.id }}`
  - kind `delete`; body type `none`; path fields `product_id`, `id`; required record fields
  `product_id`, `id`; accepted fields `id`, `product_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: Delete a product variation in ShopWired.
- `update_product_variation`: PUT `/products/{{ record.product_id }}/variations/{{ record.id }}` -
  kind `update`; body type `json`; path fields `product_id`, `id`; required record fields
  `product_id`, `id`; accepted fields `costPrice`, `gtin`, `id`, `image`, `mpn`, `price`,
  `product_id`, `rewardPoints`, `salePrice`, `sku`, `stock`, `values`, `weight`; risk: Update a
  product variation in ShopWired.
- `create_redirect`: POST `/redirects` - kind `create`; body type `json`; required record fields
  `oldPath`, `newPath`; accepted fields `newPath`, `oldPath`; risk: Create a new 301 redirect in
  ShopWired.
- `delete_redirect`: DELETE `/redirects/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: Delete a 301 redirect in ShopWired.
- `create_sale`: POST `/sales` - kind `create`; body type `json`; required record fields
  `targetType`, `targetId`, `discount`, `validFrom`, `active`; accepted fields `active`, `discount`,
  `expiresOn`, `targetId`, `targetType`, `validFrom`; risk: Create a sale in ShopWired.
- `delete_sale`: DELETE `/sales/{{ record.id }}` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: Delete a sale in ShopWired.
- `update_sale`: PUT `/sales/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`, `targetType`, `targetId`, `discount`, `validFrom`, `active`; accepted
  fields `active`, `discount`, `expiresOn`, `id`, `targetId`, `targetType`, `validFrom`; risk:
  Update a sale in ShopWired.
- `create_shipping_rate`: POST `/shipping-rates` - kind `create`; body type `json`; required record
  fields `country`, `name`, `criteria`, `from`, `to`, `cost`, `vatExclusive`; accepted fields
  `cost`, `country`, `criteria`, `from`, `name`, `postcodes`, `states`, `to`, `vatExclusive`; risk:
  Create a new shipping rate in ShopWired.
- `delete_shipping_rate`: DELETE `/shipping-rates/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Delete a shipping rate in
  ShopWired.
- `update_shipping_rate`: PUT `/shipping-rates/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `cost`, `criteria`, `enabled`,
  `from`, `id`, `name`, `postcodes`, `states`, `to`, `vatExclusive`; risk: Update a shipping rate in
  ShopWired.
- `create_shipping_zone`: POST `/shipping-zones` - kind `create`; body type `json`; required record
  fields `country`, `vat`; accepted fields `country`, `vat`; risk: Create a new shipping zone in
  ShopWired.
- `delete_shipping_zone`: DELETE `/shipping-zones/{{ record.country_id }}` - kind `delete`; body
  type `none`; path fields `country_id`; required record fields `country_id`; accepted fields
  `country_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: Delete a shipping zone in ShopWired.
- `update_shipping_zone`: PUT `/shipping-zones/{{ record.country_id }}` - kind `update`; body type
  `json`; path fields `country_id`; required record fields `country_id`, `vat`; accepted fields
  `country_id`, `vat`; risk: Update an existing shipping zone in ShopWired.
- `update_stock`: POST `/stock` - kind `create`; body type `json`; accepted fields `items`,
  `quantity`, `sku`; risk: Update stock quantity in ShopWired.
- `return_stock`: POST `/stock/return` - kind `custom`; body type `json`; required record fields
  `orderProductItemId`; accepted fields `orderProductItemId`; confirmation `destructive`; risk:
  Return stock for cancelled orders in ShopWired.
- `delete_theme_asset`: DELETE `/theme-assets/{{ record.theme_id }}` - kind `delete`; body type
  `none`; path fields `theme_id`; required record fields `theme_id`; accepted fields `theme_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Delete a
  theme asset in ShopWired.
- `create_or_update_theme_asset`: POST `/theme-assets/{{ record.theme_id }}` - kind `create`; body
  type `json`; path fields `theme_id`; required record fields `theme_id`, `key`; accepted fields
  `attachment`, `content`, `key`, `source_key`, `theme_id`; risk: Create or update a theme asset in
  ShopWired.
- `update_theme`: PUT `/themes/{{ record.id }}` - kind `update`; body type `json`; path fields `id`;
  required record fields `id`, `title`; accepted fields `id`, `title`; risk: Update a theme in
  ShopWired.
- `create_trade_customer_product_price`: POST `/trade-customer-product-prices` - kind `create`; body
  type `json`; required record fields `customerId`, `productId`, `price`; accepted fields
  `customerId`, `price`, `productId`, `sku`; risk: Create a new trade customer product price in
  ShopWired.
- `bulk_trade_customer_product_prices`: POST `/trade-customer-product-prices/bulk` - kind `create`;
  body type `json`; accepted fields `actions`; risk: Bulk create or remove trade customer product
  prices in ShopWired.
- `delete_trade_customer_product_price`: DELETE `/trade-customer-product-prices/{{ record.id }}` -
  kind `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  Delete a trade customer product price in ShopWired.
- `update_trade_customer_product_price`: PUT `/trade-customer-product-prices/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields `id`,
  `price`, `sku`; risk: Update a trade customer product price in ShopWired.
- `create_trade_group`: POST `/trade-groups` - kind `create`; body type `json`; required record
  fields `title`, `products`; accepted fields `products`, `title`; risk: Create a trade pricing band
  in ShopWired.
- `delete_trade_group`: DELETE `/trade-groups/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Delete a trade pricing band in
  ShopWired.
- `update_trade_group`: PUT `/trade-groups/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `products`; accepted fields `id`, `products`, `title`;
  risk: Update a trade pricing band in ShopWired.
- `create_voucher`: POST `/vouchers` - kind `create`; body type `json`; accepted fields `active`,
  `automatic`, `code`, `discount`, `dynamicAmount`, `expiresOn`, `target`, `validFrom`; risk: Create
  a new voucher in ShopWired.
- `delete_voucher`: DELETE `/vouchers/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: Delete a voucher in ShopWired.
- `update_voucher`: PUT `/vouchers/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `id`; risk: Update a voucher in ShopWired.
- `create_a_webhook`: POST `/webhooks` - kind `create`; body type `json`; required record fields
  `topic`, `url`; accepted fields `enabled`, `topic`, `url`; risk: Create a webhook in ShopWired.
- `delete_a_webhook`: DELETE `/webhooks/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: Delete a webhook in ShopWired.
- `update_a_webhook`: PUT `/webhooks/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `enabled`, `id`, `topic`, `url`; risk: Update a
  webhook in ShopWired.
- `verify_a_webhook`: POST `/webhooks/{{ record.id }}/verify` - kind `create`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: Verify a webhook in
  ShopWired.
- `create_a_wishlist`: POST `/wishlists` - kind `create`; body type `json`; required record fields
  `customerId`; accepted fields `customerId`, `public`; risk: Create a wishlist in ShopWired.
- `update_a_wishlist`: PUT `/wishlists/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`, `public`; accepted fields `id`, `public`; risk: Update a
  wishlist in ShopWired.
- `modify_a_wishlist`: POST `/wishlists/{{ record.id }}/modify` - kind `create`; body type `json`;
  path fields `id`; required record fields `id`, `action`; accepted fields `action`, `id`, `items`;
  confirmation `destructive`; risk: Modify a wishlist in ShopWired.

## Known limits

- Batch defaults: read_page_size=100, write_batch_size=1.
- API coverage includes 88 stream-backed endpoint group(s), 119 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=30.
