---
name: pm-shopify
description: Shopify connector knowledge and safe action guide.
---

# pm-shopify

## Purpose

Connector-owned Shopify Admin API parity ledger with a fixture-backed Shop stream and typed destructive REST delete actions; remaining official operations are blocked/planned with source evidence.

## Icon

- asset: icons/shopify.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://shopify.dev/docs/api/admin-rest

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- shop_domain
- access_token (secret)

## ETL Streams

- shop:
  - primary key: id
  - fields: created_at(), currency(), domain(), email(), iana_timezone(), id(), myshopify_domain(), name(), timezone(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- delete_blogs_id:
  - endpoint: DELETE /admin/api/latest/blogs/{{ record.id1 }}.json
  - required fields: id1
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/blogs/{id1}.json.
- delete_blogs_id_articles_id:
  - endpoint: DELETE /admin/api/latest/blogs/{{ record.id1 }}/articles/{{ record.id2 }}.json
  - required fields: id1, id2
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/blogs/{id1}/articles/{id2}.json.
- delete_blogs_id_metafields_id:
  - endpoint: DELETE /admin/api/latest/blogs/{{ record.id1 }}/metafields/{{ record.id2 }}.json
  - required fields: id1, id2
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/blogs/{id1}/metafields/{id2}.json.
- delete_carrier_services_id:
  - endpoint: DELETE /admin/api/latest/carrier_services/{{ record.id1 }}.json
  - required fields: id1
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/carrier_services/{id1}.json.
- delete_collection_listings_id:
  - endpoint: DELETE /admin/api/latest/collection_listings/{{ record.id1 }}.json
  - required fields: id1
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/collection_listings/{id1}.json.
- delete_collects_id:
  - endpoint: DELETE /admin/api/latest/collects/{{ record.id1 }}.json
  - required fields: id1
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/collects/{id1}.json.
- delete_countries_id:
  - endpoint: DELETE /admin/api/latest/countries/{{ record.id1 }}.json
  - required fields: id1
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/countries/{id1}.json.
- delete_custom_collections_id:
  - endpoint: DELETE /admin/api/latest/custom_collections/{{ record.id1 }}.json
  - required fields: id1
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/custom_collections/{id1}.json.
- delete_customers_id:
  - endpoint: DELETE /admin/api/latest/customers/{{ record.id1 }}.json
  - required fields: id1
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/customers/{id1}.json.
- delete_customers_id_addresses_id:
  - endpoint: DELETE /admin/api/latest/customers/{{ record.id1 }}/addresses/{{ record.id2 }}.json
  - required fields: id1, id2
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/customers/{id1}/addresses/{id2}.json.
- delete_draft_orders_id:
  - endpoint: DELETE /admin/api/latest/draft_orders/{{ record.id1 }}.json
  - required fields: id1
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/draft_orders/{id1}.json.
- delete_fulfillment_services_id:
  - endpoint: DELETE /admin/api/latest/fulfillment_services/{{ record.id1 }}.json
  - required fields: id1
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/fulfillment_services/{id1}.json.
- delete_marketing_events_id:
  - endpoint: DELETE /admin/api/latest/marketing_events/{{ record.id1 }}.json
  - required fields: id1
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/marketing_events/{id1}.json.
- delete_mobile_platform_applications_id:
  - endpoint: DELETE /admin/api/latest/mobile_platform_applications/{{ record.id1 }}.json
  - required fields: id1
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/mobile_platform_applications/{id1}.json.
- delete_orders_id:
  - endpoint: DELETE /admin/api/latest/orders/{{ record.id1 }}.json
  - required fields: id1
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/orders/{id1}.json.
- delete_orders_id_fulfillments_id_events_id:
  - endpoint: DELETE /admin/api/latest/orders/{{ record.id1 }}/fulfillments/{{ record.id2 }}/events/{{ record.id3 }}.json
  - required fields: id1, id2, id3
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/orders/{id1}/fulfillments/{id2}/events/{id3}.json.
- delete_orders_id_risks_id:
  - endpoint: DELETE /admin/api/latest/orders/{{ record.id1 }}/risks/{{ record.id2 }}.json
  - required fields: id1, id2
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/orders/{id1}/risks/{id2}.json.
- delete_pages_id:
  - endpoint: DELETE /admin/api/latest/pages/{{ record.id1 }}.json
  - required fields: id1
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/pages/{id1}.json.
- delete_price_rules_id:
  - endpoint: DELETE /admin/api/latest/price_rules/{{ record.id1 }}.json
  - required fields: id1
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/price_rules/{id1}.json.
- delete_price_rules_id_discount_codes_id:
  - endpoint: DELETE /admin/api/latest/price_rules/{{ record.id1 }}/discount_codes/{{ record.id2 }}.json
  - required fields: id1, id2
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/price_rules/{id1}/discount_codes/{id2}.json.
- delete_product_listings_id:
  - endpoint: DELETE /admin/api/latest/product_listings/{{ record.id1 }}.json
  - required fields: id1
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/product_listings/{id1}.json.
- delete_products_id:
  - endpoint: DELETE /admin/api/latest/products/{{ record.id1 }}.json
  - required fields: id1
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/products/{id1}.json.
- delete_products_id_images_id:
  - endpoint: DELETE /admin/api/latest/products/{{ record.id1 }}/images/{{ record.id2 }}.json
  - required fields: id1, id2
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/products/{id1}/images/{id2}.json.
- delete_products_id_variants_id:
  - endpoint: DELETE /admin/api/latest/products/{{ record.id1 }}/variants/{{ record.id2 }}.json
  - required fields: id1, id2
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/products/{id1}/variants/{id2}.json.
- delete_recurring_application_charges_id:
  - endpoint: DELETE /admin/api/latest/recurring_application_charges/{{ record.id1 }}.json
  - required fields: id1
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/recurring_application_charges/{id1}.json.
- delete_redirects_id:
  - endpoint: DELETE /admin/api/latest/redirects/{{ record.id1 }}.json
  - required fields: id1
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/redirects/{id1}.json.
- delete_script_tags_id:
  - endpoint: DELETE /admin/api/latest/script_tags/{{ record.id1 }}.json
  - required fields: id1
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/script_tags/{id1}.json.
- delete_shopify_payments_disputes_id_dispute_file_uploads_id:
  - endpoint: DELETE /admin/api/latest/shopify_payments/disputes/{{ record.id1 }}/dispute_file_uploads/{{ record.id2 }}.json
  - required fields: id1, id2
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/shopify_payments/disputes/{id1}/dispute_file_uploads/{id2}.json.
- delete_smart_collections_id:
  - endpoint: DELETE /admin/api/latest/smart_collections/{{ record.id1 }}.json
  - required fields: id1
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/smart_collections/{id1}.json.
- delete_storefront_access_tokens_id:
  - endpoint: DELETE /admin/api/latest/storefront_access_tokens/{{ record.id1 }}.json
  - required fields: id1
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/storefront_access_tokens/{id1}.json.
- delete_themes_id:
  - endpoint: DELETE /admin/api/latest/themes/{{ record.id1 }}.json
  - required fields: id1
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/themes/{id1}.json.
- delete_webhooks_id:
  - endpoint: DELETE /admin/api/latest/webhooks/{{ record.id1 }}.json
  - required fields: id1
  - risk: Permanently deletes the Shopify Admin REST resource at /admin/api/latest/webhooks/{id1}.json.

## Security

- read risk: Fixed Shopify Admin REST/GraphQL reads only; no raw query or arbitrary API passthrough.
- write risk: Typed reverse-ETL actions only; destructive deletes require `destructive` confirmation and preview approval.
- approval: Reverse ETL remains plan -> preview -> explicit approval -> execute; fixture-only evidence is not certification.
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Shopify Admin API connector-owned parity metadata with no raw API passthrough.
- Usage: Use standard `pm connectors`, `pm etl`, and reverse ETL flows with the `shopify` connector; planned provider-style commands are metadata only until fixed operation bindings exist.
- Source CLI: Shopify Admin API docs (https://shopify.dev/docs/api/admin-graphql/latest/full-index)
- Fixed reads
- Typed writes
- Ledger metadata
- Other Commands
  - shopify ledger inspect - Inspect the connector-owned Shopify operation ledger and source inventory. [intent=docs_only availability=planned]; notes: Docs/metadata only in this slice; no provider call or raw API execution.
  - shopify shop read - Read the fixed Shopify Shop REST resource through the ETL stream. [intent=etl availability=implemented stream=shop]; notes: Uses the fixed `shop` stream; no arbitrary REST path is accepted.
  - shopify delete <fixed-delete-action> - Run a connector-owned typed Shopify REST DELETE action after preview and destructive confirmation. [intent=reverse_etl availability=planned]; approval: plan -> preview -> explicit approval -> execute with typed destructive confirmation; risk: Destructive deletes require action-specific record schemas and `destructive` typed confirmation.; notes: Individual implemented delete actions are exposed through reverse ETL action names; this metadata is not a raw method/path command.
  - shopify graphql query <fixed-operation> - Future fixed Shopify Admin GraphQL query commands generated only from reviewed operation documents. [intent=direct_read availability=planned]; notes: No arbitrary GraphQL document, variables blob, or passthrough is accepted.
  - shopify graphql mutation <fixed-operation> - Future fixed Shopify Admin GraphQL mutations generated only from reviewed typed write schemas. [intent=reverse_etl availability=planned]; approval: plan -> preview -> explicit approval -> execute; risk: Mutations require action-specific schemas, redaction, preview, approval, and destructive confirmation where applicable.; notes: No arbitrary GraphQL mutation passthrough is accepted.
- Help topics:
  - shopify safety - No Shopify writes execute without plan, preview, explicit approval, and typed destructive confirmation where applicable.
  - shopify ledger - Operation ledger rows are source-linked and blocked/planned unless covered by a stream or typed write action.

## Commands

### Inspect as a manual

```bash
pm connectors inspect shopify
```

### Inspect as structured JSON

```bash
pm connectors inspect shopify --json
```

## Agent Rules

- Run pm connectors inspect shopify before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
