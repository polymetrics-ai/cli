# pm connectors inspect yotpo

```text
NAME
  pm connectors inspect yotpo - Yotpo connector manual

SYNOPSIS
  pm connectors inspect yotpo
  pm connectors inspect yotpo --json
  pm credentials add <name> --connector yotpo [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Yotpo store products, product variants, collections, customers, orders, and webhook targets/filters/subscriptions, and writes product/variant/order/customer/fulfillment/collection-membership/webhook mutations through the Yotpo Core API v3.

ICON
  asset: icons/yotpo.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://apidocs.yotpo.com/reference

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  product_id
  store_id
  access_token (secret)

ETL STREAMS
  products:
    primary key: yotpo_id
    cursor: updated_at
    fields: brand(), compare_at_price(), created_at(), currency(), custom_properties(), description(), external_id(), group_name(), gtins(), image_url(), inventory_quantity(), is_discontinued(), is_valid_url(), mpn(), name(), price(), sku(), status(), updated_at(), url(), yotpo_id()
  product_variants:
    primary key: id
    cursor: updated_at
    fields: compare_at_price(), created_at(), currency(), description(), external_id(), gtins(), id(), image_url(), inventory_quantity(), is_discontinued(), is_valid_url(), name(), options(), price(), sku(), updated_at(), url(), yotpo_id()
  collections:
    primary key: id
    cursor: updated_at
    fields: created_at(), external_id(), id(), name(), updated_at(), yotpo_id()
  customers:
    primary key: external_id
    cursor: account_updated_at
    fields: accepts_email_marketing(), accepts_sms_marketing(), account_created_at(), account_status(), account_updated_at(), address(), custom_properties(), default_currency(), default_language(), email(), external_id(), first_name(), gender(), last_name(), phone_number(), tags()
  orders:
    primary key: yotpo_id
    cursor: order_date
    fields: billing_address(), cancellation(), checkout_token(), currency(), custom_properties(), customer(), customer_locale(), external_id(), fulfillments(), landing_site_url(), line_items(), order_date(), order_name(), order_number(), payment_method(), payment_status(), shipping_address(), status(), subtotal_price(), total_price(), yotpo_id()
  webhook_targets:
    primary key: id
    cursor: updated_at
    fields: created_at(), id(), updated_at(), url(), yotpo_id()
  webhook_filters:
    primary key: id
    cursor: updated_at
    fields: created_at(), events(), id(), updated_at(), yotpo_id()
  webhook_subscriptions:
    primary key: id
    cursor: updated_at
    fields: active(), created_at(), filter_id(), id(), target_id(), updated_at(), yotpo_id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_product:
    endpoint: POST /core/v3/stores/{{ config.store_id }}/products
    risk: external mutation; creates a new product in the store's catalog; approval required. Body is wrapped under a top-level "product" key (Yotpo Core API v3 convention) — the record itself carries that wrapper, since the engine's write dialect sends record fields verbatim as the JSON body with no nested-wrapper construction primitive (see teamwork/ynab precedent).
  update_product:
    endpoint: PATCH /core/v3/stores/{{ config.store_id }}/products/{{ record.yotpo_id }}
    required fields: yotpo_id
    risk: external mutation; updates an existing product's catalog fields; approval required
  create_product_variant:
    endpoint: POST /core/v3/stores/{{ config.store_id }}/products/{{ record.product_yotpo_id }}/variants
    required fields: product_yotpo_id
    risk: external mutation; creates a new variant under an existing product; approval required
  update_product_variant:
    endpoint: PATCH /core/v3/stores/{{ config.store_id }}/products/{{ record.product_yotpo_id }}/variants/{{ record.yotpo_id }}
    required fields: product_yotpo_id, yotpo_id
    risk: external mutation; updates an existing product variant's fields; approval required
  create_order:
    endpoint: POST /core/v3/stores/{{ config.store_id }}/orders
    risk: external mutation; creates a new order (may trigger Yotpo's automatic review-request email flow for the associated customer); approval required. Not possible to send automatic review-request emails for orders older than six months (Yotpo's own documented constraint).
  update_order:
    endpoint: PATCH /core/v3/stores/{{ config.store_id }}/orders/{{ record.yotpo_id }}
    required fields: yotpo_id
    risk: external mutation; updates an existing order's status/pricing/cancellation fields; approval required
  create_customer:
    endpoint: POST /core/v3/stores/{{ config.store_id }}/customers
    risk: external mutation; creates or updates (upsert-by-external_id) a customer profile; approval required. Yotpo's own endpoint is documented as create-or-update, keyed on external_id — there is no separate update_customer action since the same request both creates and upserts.
  create_order_fulfillment:
    endpoint: POST /core/v3/stores/{{ config.store_id }}/orders/{{ record.order_yotpo_id }}/fulfillments
    required fields: order_yotpo_id
    risk: external mutation; records a shipment/fulfillment event against an existing order; approval required
  update_order_fulfillment:
    endpoint: PATCH /core/v3/stores/{{ config.store_id }}/orders/{{ record.order_yotpo_id }}/fulfillments/{{ record.yotpo_id }}
    required fields: order_yotpo_id, yotpo_id
    risk: external mutation; updates the shipment status/tracking of an existing order fulfillment; approval required
  create_collection:
    endpoint: POST /core/v3/stores/{{ config.store_id }}/collections
    risk: external mutation; creates a new product collection; approval required
  update_collection:
    endpoint: PATCH /core/v3/stores/{{ config.store_id }}/collections/{{ record.yotpo_id }}
    required fields: yotpo_id
    risk: external mutation; renames an existing product collection; approval required
  add_product_to_collection:
    endpoint: POST /core/v3/stores/{{ config.store_id }}/collections/{{ record.collection_yotpo_id }}/products
    required fields: collection_yotpo_id
    risk: external mutation; adds a product to an existing collection; approval required
  remove_product_from_collection:
    endpoint: DELETE /core/v3/stores/{{ config.store_id }}/collections/{{ record.collection_yotpo_id }}/products
    required fields: collection_yotpo_id
    optional fields: product_id
    risk: irreversible external mutation; removes a product from an existing collection; approval required
  create_webhook_target:
    endpoint: POST /core/v3/stores/{{ config.store_id }}/webhooks/targets
    risk: external mutation; registers a webhook callback URL target; approval required
  update_webhook_target:
    endpoint: PATCH /core/v3/stores/{{ config.store_id }}/webhooks/targets/{{ record.yotpo_id }}
    required fields: yotpo_id
    risk: external mutation; changes an existing webhook target's callback URL; approval required
  delete_webhook_target:
    endpoint: DELETE /core/v3/stores/{{ config.store_id }}/webhooks/targets/{{ record.yotpo_id }}
    required fields: yotpo_id
    risk: irreversible external deletion; removes a registered webhook target (any subscription still referencing it becomes inactive); approval required
  create_webhook_filter:
    endpoint: POST /core/v3/stores/{{ config.store_id }}/webhooks/filters
    risk: external mutation; creates a webhook event filter (an event type cannot be used twice across filters, per Yotpo's own constraint); approval required
  update_webhook_filter:
    endpoint: PATCH /core/v3/stores/{{ config.store_id }}/webhooks/filters/{{ record.yotpo_id }}
    required fields: yotpo_id
    risk: external mutation; changes an existing webhook filter's subscribed event types; approval required
  delete_webhook_filter:
    endpoint: DELETE /core/v3/stores/{{ config.store_id }}/webhooks/filters/{{ record.yotpo_id }}
    required fields: yotpo_id
    risk: irreversible external deletion; removes a webhook filter (only unused filters can be deleted, per Yotpo's own constraint); approval required
  create_webhook_subscription:
    endpoint: POST /core/v3/stores/{{ config.store_id }}/webhooks/subscriptions
    risk: external mutation; activates webhook event delivery by combining an existing target and filter; approval required
  update_webhook_subscription:
    endpoint: PATCH /core/v3/stores/{{ config.store_id }}/webhooks/subscriptions/{{ record.yotpo_id }}
    required fields: yotpo_id
    risk: external mutation; retargets or (de)activates an existing webhook subscription; approval required
  delete_webhook_subscription:
    endpoint: DELETE /core/v3/stores/{{ config.store_id }}/webhooks/subscriptions/{{ record.yotpo_id }}
    required fields: yotpo_id
    risk: irreversible external deletion; stops webhook event delivery for an existing target/filter combination; approval required

SECURITY
  read risk: external Yotpo Core API read of product, variant, collection, customer, order, and webhook configuration data
  write risk: external mutation: creates/updates products, variants, orders, customers, order fulfillments, and collections; manages collection membership and webhook target/filter/subscription lifecycle
  approval: required for all write actions; reads require none
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect yotpo

  # Inspect as structured JSON
  pm connectors inspect yotpo --json

AGENT WORKFLOW
  - Run pm connectors inspect yotpo before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
