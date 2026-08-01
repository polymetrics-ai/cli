# pm connectors inspect shopify

```text
NAME
  pm connectors inspect shopify - Shopify connector manual

SYNOPSIS
  pm connectors inspect shopify
  pm connectors inspect shopify --json
  pm credentials add <name> --connector shopify [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Connector-owned Shopify Admin API parity ledger with a fixture-backed Shop stream, typed destructive REST delete actions, source-deprecated dispositions, and critical blocked destructive safeguards for state-destroying official operations.

ICON
  asset: icons/shopify.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://shopify.dev/docs/api/admin-rest

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  shop_domain
  access_token (secret)

ETL STREAMS
  shop:
    primary key: id
    fields: created_at(), currency(), domain(), email(), iana_timezone(), id(), myshopify_domain(), name(), timezone(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  delete_articles_id_metafields_id:
    endpoint: DELETE /admin/api/latest/articles/{{ record.article_id }}/metafields/{{ record.metafield_id }}.json
    required fields: article_id, metafield_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/articles/{article_id}/metafields/{metafield_id}.json.
  delete_blogs_id:
    endpoint: DELETE /admin/api/latest/blogs/{{ record.blog_id }}.json
    required fields: blog_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/blogs/{blog_id}.json.
  delete_blogs_id_articles_id:
    endpoint: DELETE /admin/api/latest/blogs/{{ record.blog_id }}/articles/{{ record.article_id }}.json
    required fields: blog_id, article_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/blogs/{blog_id}/articles/{article_id}.json.
  delete_blogs_id_metafields_id:
    endpoint: DELETE /admin/api/latest/blogs/{{ record.blog_id }}/metafields/{{ record.metafield_id }}.json
    required fields: blog_id, metafield_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/blogs/{blog_id}/metafields/{metafield_id}.json.
  delete_carrier_services_id:
    endpoint: DELETE /admin/api/latest/carrier_services/{{ record.carrier_service_id }}.json
    required fields: carrier_service_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/carrier_services/{carrier_service_id}.json.
  delete_collection_listings_id:
    endpoint: DELETE /admin/api/latest/collection_listings/{{ record.collection_listing_id }}.json
    required fields: collection_listing_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/collection_listings/{collection_listing_id}.json.
  delete_collections_id_metafields_id:
    endpoint: DELETE /admin/api/latest/collections/{{ record.collection_id }}/metafields/{{ record.metafield_id }}.json
    required fields: collection_id, metafield_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/collections/{collection_id}/metafields/{metafield_id}.json.
  delete_collects_id:
    endpoint: DELETE /admin/api/latest/collects/{{ record.collect_id }}.json
    required fields: collect_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/collects/{collect_id}.json.
  delete_countries_id:
    endpoint: DELETE /admin/api/latest/countries/{{ record.country_id }}.json
    required fields: country_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/countries/{country_id}.json.
  delete_custom_collections_id:
    endpoint: DELETE /admin/api/latest/custom_collections/{{ record.custom_collection_id }}.json
    required fields: custom_collection_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/custom_collections/{custom_collection_id}.json.
  delete_customers_id:
    endpoint: DELETE /admin/api/latest/customers/{{ record.customer_id }}.json
    required fields: customer_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/customers/{customer_id}.json.
  delete_customers_id_addresses_id:
    endpoint: DELETE /admin/api/latest/customers/{{ record.customer_id }}/addresses/{{ record.address_id }}.json
    required fields: customer_id, address_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/customers/{customer_id}/addresses/{address_id}.json.
  delete_customers_id_metafields_id:
    endpoint: DELETE /admin/api/latest/customers/{{ record.customer_id }}/metafields/{{ record.metafield_id }}.json
    required fields: customer_id, metafield_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/customers/{customer_id}/metafields/{metafield_id}.json.
  delete_draft_orders_id:
    endpoint: DELETE /admin/api/latest/draft_orders/{{ record.draft_order_id }}.json
    required fields: draft_order_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/draft_orders/{draft_order_id}.json.
  delete_draft_orders_id_metafields_id:
    endpoint: DELETE /admin/api/latest/draft_orders/{{ record.draft_order_id }}/metafields/{{ record.metafield_id }}.json
    required fields: draft_order_id, metafield_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/draft_orders/{draft_order_id}/metafields/{metafield_id}.json.
  delete_fulfillment_services_id:
    endpoint: DELETE /admin/api/latest/fulfillment_services/{{ record.fulfillment_service_id }}.json
    required fields: fulfillment_service_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/fulfillment_services/{fulfillment_service_id}.json.
  delete_marketing_events_id:
    endpoint: DELETE /admin/api/latest/marketing_events/{{ record.marketing_event_id }}.json
    required fields: marketing_event_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/marketing_events/{marketing_event_id}.json.
  delete_metafields_id:
    endpoint: DELETE /admin/api/latest/metafields/{{ record.metafield_id }}.json
    required fields: metafield_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/metafields/{metafield_id}.json.
  delete_mobile_platform_applications_id:
    endpoint: DELETE /admin/api/latest/mobile_platform_applications/{{ record.mobile_platform_application_id }}.json
    required fields: mobile_platform_application_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/mobile_platform_applications/{mobile_platform_application_id}.json.
  delete_orders_id:
    endpoint: DELETE /admin/api/latest/orders/{{ record.order_id }}.json
    required fields: order_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/orders/{order_id}.json.
  delete_orders_id_fulfillments_id_events_id:
    endpoint: DELETE /admin/api/latest/orders/{{ record.order_id }}/fulfillments/{{ record.fulfillment_id }}/events/{{ record.event_id }}.json
    required fields: order_id, fulfillment_id, event_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/orders/{order_id}/fulfillments/{fulfillment_id}/events/{event_id}.json.
  delete_orders_id_metafields_id:
    endpoint: DELETE /admin/api/latest/orders/{{ record.order_id }}/metafields/{{ record.metafield_id }}.json
    required fields: order_id, metafield_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/orders/{order_id}/metafields/{metafield_id}.json.
  delete_orders_id_risks_id:
    endpoint: DELETE /admin/api/latest/orders/{{ record.order_id }}/risks/{{ record.risk_id }}.json
    required fields: order_id, risk_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/orders/{order_id}/risks/{risk_id}.json.
  delete_pages_id:
    endpoint: DELETE /admin/api/latest/pages/{{ record.page_id }}.json
    required fields: page_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/pages/{page_id}.json.
  delete_pages_id_metafields_id:
    endpoint: DELETE /admin/api/latest/pages/{{ record.page_id }}/metafields/{{ record.metafield_id }}.json
    required fields: page_id, metafield_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/pages/{page_id}/metafields/{metafield_id}.json.
  delete_price_rules_id:
    endpoint: DELETE /admin/api/latest/price_rules/{{ record.price_rule_id }}.json
    required fields: price_rule_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/price_rules/{price_rule_id}.json.
  delete_price_rules_id_discount_codes_id:
    endpoint: DELETE /admin/api/latest/price_rules/{{ record.price_rule_id }}/discount_codes/{{ record.discount_code_id }}.json
    required fields: price_rule_id, discount_code_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/price_rules/{price_rule_id}/discount_codes/{discount_code_id}.json.
  delete_product_images_id_metafields_id:
    endpoint: DELETE /admin/api/latest/product_images/{{ record.product_image_id }}/metafields/{{ record.metafield_id }}.json
    required fields: product_image_id, metafield_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/product_images/{product_image_id}/metafields/{metafield_id}.json.
  delete_product_listings_id:
    endpoint: DELETE /admin/api/latest/product_listings/{{ record.product_listing_id }}.json
    required fields: product_listing_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/product_listings/{product_listing_id}.json.
  delete_products_id:
    endpoint: DELETE /admin/api/latest/products/{{ record.product_id }}.json
    required fields: product_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/products/{product_id}.json.
  delete_products_id_images_id:
    endpoint: DELETE /admin/api/latest/products/{{ record.product_id }}/images/{{ record.image_id }}.json
    required fields: product_id, image_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/products/{product_id}/images/{image_id}.json.
  delete_products_id_metafields_id:
    endpoint: DELETE /admin/api/latest/products/{{ record.product_id }}/metafields/{{ record.metafield_id }}.json
    required fields: product_id, metafield_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/products/{product_id}/metafields/{metafield_id}.json.
  delete_products_id_variants_id:
    endpoint: DELETE /admin/api/latest/products/{{ record.product_id }}/variants/{{ record.variant_id }}.json
    required fields: product_id, variant_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/products/{product_id}/variants/{variant_id}.json.
  delete_recurring_application_charges_id:
    endpoint: DELETE /admin/api/latest/recurring_application_charges/{{ record.recurring_application_charge_id }}.json
    required fields: recurring_application_charge_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/recurring_application_charges/{recurring_application_charge_id}.json.
  delete_redirects_id:
    endpoint: DELETE /admin/api/latest/redirects/{{ record.redirect_id }}.json
    required fields: redirect_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/redirects/{redirect_id}.json.
  delete_script_tags_id:
    endpoint: DELETE /admin/api/latest/script_tags/{{ record.script_tag_id }}.json
    required fields: script_tag_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/script_tags/{script_tag_id}.json.
  delete_shopify_payments_disputes_id_dispute_file_uploads_id:
    endpoint: DELETE /admin/api/latest/shopify_payments/disputes/{{ record.dispute_id }}/dispute_file_uploads/{{ record.dispute_file_upload_id }}.json
    required fields: dispute_id, dispute_file_upload_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/shopify_payments/disputes/{dispute_id}/dispute_file_uploads/{dispute_file_upload_id}.json.
  delete_smart_collections_id:
    endpoint: DELETE /admin/api/latest/smart_collections/{{ record.smart_collection_id }}.json
    required fields: smart_collection_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/smart_collections/{smart_collection_id}.json.
  delete_storefront_access_tokens_id:
    endpoint: DELETE /admin/api/latest/storefront_access_tokens/{{ record.storefront_access_token_id }}.json
    required fields: storefront_access_token_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/storefront_access_tokens/{storefront_access_token_id}.json.
  delete_themes_id:
    endpoint: DELETE /admin/api/latest/themes/{{ record.theme_id }}.json
    required fields: theme_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/themes/{theme_id}.json.
  delete_variants_id_metafields_id:
    endpoint: DELETE /admin/api/latest/variants/{{ record.variant_id }}/metafields/{{ record.metafield_id }}.json
    required fields: variant_id, metafield_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/variants/{variant_id}/metafields/{metafield_id}.json.
  delete_webhooks_id:
    endpoint: DELETE /admin/api/latest/webhooks/{{ record.webhook_id }}.json
    required fields: webhook_id
    risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/webhooks/{webhook_id}.json.

SECURITY
  read risk: Fixed Shopify Admin REST/GraphQL reads only; no raw query or arbitrary API passthrough.
  write risk: Typed reverse-ETL actions only; executable deletes and state-destroying blocked operation rows require `destructive` confirmation and preview approval before execution.
  approval: Reverse ETL remains plan -> preview -> explicit approval -> execute; fixture-only evidence is not certification.
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Shopify Admin API connector-owned parity metadata with no raw API passthrough.
  Usage: Use standard `pm connectors`, `pm etl`, and reverse ETL flows with the `shopify` connector; planned provider-style commands are metadata only until fixed operation bindings exist.
  Source CLI: Shopify Admin API docs (https://shopify.dev/docs/api/admin-graphql/latest/full-index)
  Fixed reads
  Typed writes
  Ledger metadata
  Other Commands
    ledger inspect - Inspect the connector-owned Shopify operation ledger and source inventory. [intent=docs_only availability=planned]; notes: Docs/metadata only in this slice; no provider call or raw API execution. Current source inventory contains 1166 operation rows, including 211 blocked state-destroying rows marked `destructive_action` with `confirm: destructive` and 59 source-deprecated GraphQL rows with explicit deprecated notes.
    shop read - Read the fixed Shopify Shop REST resource through the ETL stream. [intent=etl availability=implemented stream=shop]; notes: Uses the fixed `shop` stream; no arbitrary REST path is accepted.
    delete <fixed-delete-action> - Run one of 42 connector-owned typed Shopify REST DELETE actions after preview and destructive confirmation. [intent=reverse_etl availability=planned]; approval: plan -> preview -> explicit approval -> execute with typed destructive confirmation; risk: Destructive deletes require action-specific record schemas and `destructive` typed confirmation.; notes: Individual implemented delete actions are exposed through reverse ETL action names; this metadata is not a raw method/path command. Inventory-level and theme-asset DELETE shapes remain planned until shared write-query support exists.
    graphql query <fixed-operation> - Future fixed Shopify Admin GraphQL query commands generated only from reviewed operation documents. [intent=direct_read availability=planned]; notes: No arbitrary GraphQL document, variables blob, or passthrough is accepted.
    graphql mutation <fixed-operation> - Future fixed Shopify Admin GraphQL mutations generated only from reviewed typed write schemas. [intent=reverse_etl availability=planned]; approval: plan -> preview -> explicit approval -> execute; state-destroying operations require typed destructive confirmation; risk: Mutations require action-specific schemas, redaction, preview, approval, and `destructive` confirmation for state-destroying operations.; notes: No arbitrary GraphQL mutation passthrough is accepted.
  Help topics:
    shopify safety - No Shopify writes execute without plan, preview, explicit approval, and typed destructive confirmation for state-destroying operations.
    shopify ledger - Operation ledger rows are source-linked and blocked/planned unless covered by a stream or typed write action; destructive rows carry critical risk and `confirm: destructive`.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect shopify

  # Inspect as structured JSON
  pm connectors inspect shopify --json

AGENT WORKFLOW
  - Run pm connectors inspect shopify before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
