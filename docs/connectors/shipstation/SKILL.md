---
name: pm-shipstation
description: ShipStation connector knowledge and safe action guide.
---

# pm-shipstation

## Purpose

Reads ShipStation orders, shipments, products, and customers through the ShipStation REST API.

## Icon

- id: shipstation
- asset: icons/shipstation.svg
- source: official
- review_status: official_verified
- review_url: https://www.shipstation.com/docs/api/

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- api_key (secret) (required)
- api_secret (secret) (required)

## ETL Streams

- orders:
  - primary key: id
  - cursor: modified_at
  - fields: id(integer), modified_at(string), order_number(string), status(string)
- shipments:
  - primary key: id
  - cursor: modified_at
  - fields: id(integer), modified_at(string), order_number(string), status(string)
- products:
  - primary key: id
  - cursor: modified_at
  - fields: id(integer), modified_at(string), name(string), sku(string)
- customers:
  - primary key: id
  - cursor: modified_at
  - fields: email(string), id(integer), modified_at(string), name(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external ShipStation API read of order, shipment, product, and customer data
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Declared shipstation API commands.
- Usage: pm shipstation <command> [flags]
- Global flags:
  - --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
- Other Commands
  - operations get-accounts-listtags - Declared direct read: GET /accounts/listtags. [intent=direct_read availability=partial]; notes: Blocked: locked source operation shipstation.provider.list-account-tags-1 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
  - operations get-carriers - Declared direct read: GET /carriers. [intent=direct_read availability=partial]; notes: Blocked: locked source operation shipstation.provider.list-carriers-2 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
  - operations post-carriers-addfunds - Declared direct write: POST /carriers/addfunds. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Blocked: locked source operation shipstation.provider.add-funds-to-carrier-3 has no declaration-owned executable direct_write route.
  - operations get-carriers-getcarrier - Declared direct read: GET /carriers/getcarrier. [intent=direct_read availability=partial]; notes: Blocked: locked source operation shipstation.provider.get-carrier-4 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
  - operations get-carriers-listpackages - Declared direct read: GET /carriers/listpackages. [intent=direct_read availability=partial]; notes: Blocked: locked source operation shipstation.provider.list-packages-5 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
  - operations get-carriers-listservices - Declared direct read: GET /carriers/listservices. [intent=direct_read availability=partial]; notes: Blocked: locked source operation shipstation.provider.list-services-6 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
  - operations get-customers - Declared etl: GET /customers. [intent=etl availability=implemented stream=customers]; notes: Provider GET /customers is bound to the existing customers stream with its connector-owned schema and pagination contract.
  - operations get-customers-customer-id - Declared direct read: GET /customers/{customerId}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation shipstation.provider.get-customer-8 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
  - operations get-fulfillments - Declared direct read: GET /fulfillments. [intent=direct_read availability=partial]; notes: Blocked: locked source operation shipstation.provider.list-fulfillments-9 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
  - operations get-orders - Declared etl: GET /orders. [intent=etl availability=implemented stream=orders]; notes: Provider GET /orders is bound to the existing orders stream with its connector-owned schema and pagination contract.
  - operations delete-orders-order-id - Declared direct write: DELETE /orders/{orderId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Documented provider deletion is not promoted without an operation-specific plan, preview, approval, typed confirmation, and execution contract.; notes: Blocked: locked source operation shipstation.provider.delete-order-11 has no declaration-owned executable direct_write route.
  - operations get-orders-order-id - Declared direct read: GET /orders/{orderId}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation shipstation.provider.get-order-12 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
  - operations post-orders-addtag - Declared direct write: POST /orders/addtag. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Blocked: locked source operation shipstation.provider.add-tag-to-order-13 has no declaration-owned executable direct_write route.
  - operations post-orders-assignuser - Declared direct write: POST /orders/assignuser. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Blocked: locked source operation shipstation.provider.assign-user-to-order-14 has no declaration-owned executable direct_write route.
  - operations post-orders-createlabelfororder - Declared direct write: POST /orders/createlabelfororder. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Blocked: locked source operation shipstation.provider.create-label-for-order-15 has no declaration-owned executable direct_write route.
  - operations post-orders-createorder - Declared direct write: POST /orders/createorder. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Legacy connector never implemented order creation; notes: Blocked: locked source operation shipstation.provider.create-update-order-16 has no declaration-owned executable direct_write route.
  - operations post-orders-createorders - Declared direct write: POST /orders/createorders. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Blocked: locked source operation shipstation.provider.create-update-multiple-orders-17 has no declaration-owned executable direct_write route.
  - operations post-orders-holduntil - Declared direct write: POST /orders/holduntil. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Blocked: locked source operation shipstation.provider.hold-order-until-18 has no declaration-owned executable direct_write route.
  - operations get-orders-listbytag - Declared direct read: GET /orders/listbytag. [intent=direct_read availability=partial]; notes: Blocked: locked source operation shipstation.provider.list-orders-by-tag-19 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
  - operations post-orders-markasshipped - Declared direct write: POST /orders/markasshipped. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Blocked: locked source operation shipstation.provider.mark-order-as-shipped-20 has no declaration-owned executable direct_write route.
  - operations post-orders-removetag - Declared direct write: POST /orders/removetag. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Blocked: locked source operation shipstation.provider.remove-tag-from-order-21 has no declaration-owned executable direct_write route.
  - operations post-orders-restorefromhold - Declared direct write: POST /orders/restorefromhold. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Blocked: locked source operation shipstation.provider.restore-order-from-hold-22 has no declaration-owned executable direct_write route.
  - operations post-orders-unassignuser - Declared direct write: POST /orders/unassignuser. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Blocked: locked source operation shipstation.provider.unassign-user-from-order-23 has no declaration-owned executable direct_write route.
  - operations get-products - Declared etl: GET /products. [intent=etl availability=implemented stream=products]; notes: Provider GET /products is bound to the existing products stream with its connector-owned schema and pagination contract.
  - operations get-products-product-id - Declared direct read: GET /products/{productId}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation shipstation.provider.get-product-25 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
  - operations put-products-product-id - Declared direct write: PUT /products/{productId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Blocked: locked source operation shipstation.provider.update-product-26 has no declaration-owned executable direct_write route.
  - operations get-shipments - Declared etl: GET /shipments. [intent=etl availability=implemented stream=shipments]; notes: Provider GET /shipments is bound to the existing shipments stream with its connector-owned schema and pagination contract.
  - operations post-shipments-createlabel - Declared direct write: POST /shipments/createlabel. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Legacy connector never implemented label purchasing; notes: Blocked: locked source operation shipstation.provider.create-shipment-label-28 has no declaration-owned executable direct_write route.
  - operations post-shipments-getrates - Declared direct write: POST /shipments/getrates. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Blocked: locked source operation shipstation.provider.get-rates-29 has no declaration-owned executable direct_write route.
  - operations post-shipments-voidlabel - Declared direct write: POST /shipments/voidlabel. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Blocked: locked source operation shipstation.provider.void-label-30 has no declaration-owned executable direct_write route.
  - operations get-stores - Declared direct read: GET /stores. [intent=direct_read availability=partial]; notes: Blocked: locked source operation shipstation.provider.list-stores-31 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
  - operations get-stores-store-id - Declared direct read: GET /stores/{storeId}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation shipstation.provider.get-store-32 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
  - operations put-stores-store-id - Declared direct write: PUT /stores/{storeId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Blocked: locked source operation shipstation.provider.update-store-33 has no declaration-owned executable direct_write route.
  - operations post-stores-deactivate - Declared direct write: POST /stores/deactivate. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Blocked: locked source operation shipstation.provider.deactivate-store-34 has no declaration-owned executable direct_write route.
  - operations get-stores-getrefreshstatus - Declared direct read: GET /stores/getrefreshstatus. [intent=direct_read availability=partial]; notes: Blocked: locked source operation shipstation.provider.get-store-refresh-status-35 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
  - operations get-stores-marketplaces - Declared direct read: GET /stores/marketplaces. [intent=direct_read availability=partial]; notes: Blocked: locked source operation shipstation.provider.list-marketplaces-36 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
  - operations post-stores-reactivate - Declared direct write: POST /stores/reactivate. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Blocked: locked source operation shipstation.provider.reactivate-store-37 has no declaration-owned executable direct_write route.
  - operations post-stores-refreshstore - Declared direct write: POST /stores/refreshstore. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Blocked: locked source operation shipstation.provider.refresh-store-38 has no declaration-owned executable direct_write route.
  - operations get-users - Declared direct read: GET /users. [intent=direct_read availability=partial]; notes: Blocked: locked source operation shipstation.provider.list-users-39 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
  - operations get-warehouses - Declared direct read: GET /warehouses. [intent=direct_read availability=partial]; notes: Blocked: locked source operation shipstation.provider.list-warehouses-40 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
  - operations delete-warehouses-warehouse-id - Declared direct write: DELETE /warehouses/{warehouseId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Documented provider deletion is not promoted without an operation-specific plan, preview, approval, typed confirmation, and execution contract.; notes: Blocked: locked source operation shipstation.provider.delete-warehouse-41 has no declaration-owned executable direct_write route.
  - operations get-warehouses-warehouse-id - Declared direct read: GET /warehouses/{warehouseId}. [intent=direct_read availability=partial]; notes: Blocked: locked source operation shipstation.provider.get-warehouse-42 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
  - operations post-warehouses-createwarehouse - Declared direct write: POST /warehouses/createwarehouse. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Blocked: locked source operation shipstation.provider.create-warehouse-43 has no declaration-owned executable direct_write route.
  - operations put-warehouses-updatewarehouse - Declared direct write: PUT /warehouses/updatewarehouse. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Blocked: locked source operation shipstation.provider.update-warehouse-44 has no declaration-owned executable direct_write route.
  - operations get-webhooks - Declared direct read: GET /webhooks. [intent=direct_read availability=partial]; notes: Blocked: locked source operation shipstation.provider.list-webhooks-45 has no declaration-owned executable stream, direct-read, binary, or status route.; flags: --page, --page-cursor
  - operations delete-webhooks-webhook-id - Declared direct write: DELETE /webhooks/{webhookId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Documented provider deletion is not promoted without an operation-specific plan, preview, approval, typed confirmation, and execution contract.; notes: Blocked: locked source operation shipstation.provider.unsubscribe-webhook-46 has no declaration-owned executable direct_write route.
  - operations post-webhooks-subscribe - Declared direct write: POST /webhooks/subscribe. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: /webhooks/subscribe is a mutation endpoint; this bundle has no reviewed write record schema, approval risk, or legacy write contract for that operation; notes: Blocked: locked source operation shipstation.provider.subscribe-webhook-47 has no declaration-owned executable direct_write route.

## Commands

### Inspect as a manual

```bash
pm connectors inspect shipstation
```

### Inspect as structured JSON

```bash
pm connectors inspect shipstation --json
```

## Agent Rules

- Run pm connectors inspect shipstation before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
