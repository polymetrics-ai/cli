# pm connectors inspect printify

```text
NAME
  pm connectors inspect printify - Printify connector manual

SYNOPSIS
  pm connectors inspect printify
  pm connectors inspect printify --json
  pm credentials add <name> --connector printify [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes Printify shops, catalog resources, products, orders, uploads, and webhooks through the Printify public API.

ICON
  asset: icons/printify.svg
  source: official
  review_status: official_verified
  review_url: https://developers.printify.com/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  blueprint_id
  image_id
  order_id
  order_sku
  order_status
  print_provider_id
  product_id
  shop_id
  show_out_of_stock
  webhook_id
  api_token (secret)

ETL STREAMS
  shops:
    primary key: id
    fields: id(), sales_channel(), title()
  products:
    primary key: id
    cursor: updated_at
    fields: created_at(), id(), sales_channel(), status(), title(), updated_at(), visible()
  orders:
    primary key: id
    cursor: updated_at
    fields: created_at(), id(), sales_channel(), status(), title(), updated_at(), visible()
  blueprints:
    primary key: id
    fields: id(), title()
  print_providers:
    primary key: id
    fields: id(), title()
  blueprint_detail:
    primary key: id
    fields: brand(), description(), id(), images(), model(), title()
  blueprint_print_providers:
    primary key: id
    fields: decoration_methods(), id(), title()
  blueprint_variants:
    primary key: id
    fields: blueprint_id(), decoration_methods(), id(), is_available(), options(), placeholders(), print_provider_id(), title()
  shipping_profiles:
    primary key: variant_ids
    fields: additional_items(), blueprint_id(), countries(), first_item(), handling_time(), print_provider_id(), variant_ids()
  print_provider_detail:
    primary key: id
    fields: blueprints(), id(), location(), title()
  product_detail:
    primary key: id
    cursor: updated_at
    fields: blueprint_id(), created_at(), description(), external(), id(), images(), is_locked(), options(), print_areas(), print_provider_id(), sales_channel(), shop_id(), tags(), title(), updated_at(), user_id(), variants(), visible()
  product_gpsr:
    primary key: title
    fields: text(), title()
  order_detail:
    primary key: id
    cursor: updated_at
    fields: address_to(), app_order_id(), created_at(), id(), line_items(), metadata(), shipping_method(), status(), updated_at()
  uploads:
    primary key: id
    cursor: upload_time
    fields: file_name(), height(), id(), mime_type(), preview_url(), size(), upload_time(), width()
  upload_detail:
    primary key: id
    cursor: upload_time
    fields: file_name(), height(), id(), mime_type(), preview_url(), size(), upload_time(), width()
  webhooks:
    primary key: id
    fields: id(), shop_id(), topic(), url()
  v2_shipping_methods:
    primary key: id
    fields: attributes(), blueprint_id(), id(), print_provider_id(), type()
  v2_shipping_standard:
    primary key: id
    fields: attributes(), blueprint_id(), id(), print_provider_id(), shipping_method(), type()
  v2_shipping_priority:
    primary key: id
    fields: attributes(), blueprint_id(), id(), print_provider_id(), shipping_method(), type()
  v2_shipping_express:
    primary key: id
    fields: attributes(), blueprint_id(), id(), print_provider_id(), shipping_method(), type()
  v2_shipping_economy:
    primary key: id
    fields: attributes(), blueprint_id(), id(), print_provider_id(), shipping_method(), type()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  disconnect_shop:
    endpoint: DELETE /v1/shops/{{ config.shop_id }}/connection.json
    risk: disconnects the configured shop from the Printify account
  create_product:
    endpoint: POST /v1/shops/{{ config.shop_id }}/products.json
    risk: creates a product in the configured shop
  update_product:
    endpoint: PUT /v1/shops/{{ config.shop_id }}/products/{{ record.product_id }}.json
    required fields: product_id
    risk: updates an existing product in the configured shop
  delete_product:
    endpoint: DELETE /v1/shops/{{ config.shop_id }}/products/{{ record.product_id }}.json
    required fields: product_id
    risk: deletes a product from the configured shop
  publish_product:
    endpoint: POST /v1/shops/{{ config.shop_id }}/products/{{ record.product_id }}/publish.json
    required fields: product_id
    risk: publishes a product to the connected sales channel
  mark_product_publishing_succeeded:
    endpoint: POST /v1/shops/{{ config.shop_id }}/products/{{ record.product_id }}/publishing_succeeded.json
    required fields: product_id
    risk: marks product publishing as succeeded and stores an external handle
  mark_product_publishing_failed:
    endpoint: POST /v1/shops/{{ config.shop_id }}/products/{{ record.product_id }}/publishing_failed.json
    required fields: product_id
    risk: marks product publishing as failed
  unpublish_product:
    endpoint: POST /v1/shops/{{ config.shop_id }}/products/{{ record.product_id }}/unpublish.json
    required fields: product_id
    risk: notifies Printify that a product has been unpublished
  submit_order:
    endpoint: POST /v1/shops/{{ config.shop_id }}/orders.json
    risk: submits an order to Printify
  submit_express_order:
    endpoint: POST /v1/shops/{{ config.shop_id }}/orders/express.json
    risk: submits a Printify Express order
  send_order_to_production:
    endpoint: POST /v1/shops/{{ config.shop_id }}/orders/{{ record.order_id }}/send_to_production.json
    required fields: order_id
    risk: sends an existing order to production
  calculate_order_shipping:
    endpoint: POST /v1/shops/{{ config.shop_id }}/orders/shipping.json
    risk: calculates shipping costs for a prospective order without submitting it
  cancel_order:
    endpoint: POST /v1/shops/{{ config.shop_id }}/orders/{{ record.order_id }}/cancel.json
    required fields: order_id
    risk: cancels an unpaid order
  upload_image:
    endpoint: POST /v1/uploads/images.json
    risk: uploads an image into the Printify media library
  archive_uploaded_image:
    endpoint: POST /v1/uploads/{{ record.image_id }}/archive.json
    required fields: image_id
    risk: archives an uploaded image
  create_webhook:
    endpoint: POST /v1/shops/{{ config.shop_id }}/webhooks.json
    risk: creates a webhook subscription for the configured shop
  update_webhook:
    endpoint: PUT /v1/shops/{{ config.shop_id }}/webhooks/{{ record.webhook_id }}.json
    required fields: webhook_id
    risk: updates an existing webhook subscription
  delete_webhook:
    endpoint: DELETE /v1/shops/{{ config.shop_id }}/webhooks/{{ record.webhook_id }}.json?host={{ record.host }}
    required fields: webhook_id, host
    risk: deletes a webhook subscription after host safeguard matching
  simulate_webhook:
    endpoint: POST /v1/shops/{{ config.shop_id }}/webhooks/{{ record.webhook_id }}/simulate
    required fields: webhook_id
    risk: sends a webhook simulation event for testing

SECURITY
  read risk: external Printify API read of shop, catalog, product, order, upload, and webhook metadata
  write risk: creates, updates, publishes, unpublishes, deletes, archives, disconnects, submits, cancels, and simulates Printify resources depending on the selected write action
  approval: reverse ETL writes require plan preview and approval token; destructive product/order/shop/upload/webhook actions are marked destructive
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect printify

  # Inspect as structured JSON
  pm connectors inspect printify --json

AGENT WORKFLOW
  - Run pm connectors inspect printify before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
