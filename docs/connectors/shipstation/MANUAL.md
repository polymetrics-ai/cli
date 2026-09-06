# pm connectors inspect shipstation

```text
NAME
  pm connectors inspect shipstation - ShipStation connector manual

SYNOPSIS
  pm connectors inspect shipstation
  pm connectors inspect shipstation --json
  pm credentials add <name> --connector shipstation [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads ShipStation orders, shipments, products, and customers through the ShipStation REST API.

ICON
  id: shipstation
  asset: icons/shipstation.svg
  source: official
  review_status: official_verified
  review_url: https://www.shipstation.com/docs/api/

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  api_key (secret) (required)
  api_secret (secret) (required)

ETL STREAMS
  orders:
    primary key: id
    cursor: modified_at
    fields: id(integer), modified_at(string), order_number(string), status(string)
  shipments:
    primary key: id
    cursor: modified_at
    fields: id(integer), modified_at(string), order_number(string), status(string)
  products:
    primary key: id
    cursor: modified_at
    fields: id(integer), modified_at(string), name(string), sku(string)
  customers:
    primary key: id
    cursor: modified_at
    fields: email(string), id(integer), modified_at(string), name(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external ShipStation API read of order, shipment, product, and customer data
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Declared shipstation API commands.
  Usage: pm shipstation <command> [flags]
  Global flags:
    --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
  Other Commands
    operations get-accounts-listtags - Declared direct read: GET /accounts/listtags. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
    operations get-carriers - Declared direct read: GET /carriers. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
    operations post-carriers-addfunds - Declared direct write: POST /carriers/addfunds. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
    operations get-carriers-getcarrier - Declared direct read: GET /carriers/getcarrier. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
    operations get-carriers-listpackages - Declared direct read: GET /carriers/listpackages. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
    operations get-carriers-listservices - Declared direct read: GET /carriers/listservices. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
    operations get-customers - Declared etl: GET /customers. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
    operations get-customers-customer-id - Declared direct read: GET /customers/{customerId}. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
    operations get-fulfillments - Declared direct read: GET /fulfillments. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
    operations get-orders - Declared etl: GET /orders. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
    operations delete-orders-order-id - Declared direct write: DELETE /orders/{orderId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Documented provider deletion is not promoted without an operation-specific plan, preview, approval, typed confirmation, and execution contract.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
    operations get-orders-order-id - Declared direct read: GET /orders/{orderId}. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
    operations post-orders-addtag - Declared direct write: POST /orders/addtag. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
    operations post-orders-assignuser - Declared direct write: POST /orders/assignuser. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
    operations post-orders-createlabelfororder - Declared direct write: POST /orders/createlabelfororder. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
    operations post-orders-createorder - Declared direct write: POST /orders/createorder. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Legacy connector never implemented order creation; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
    operations post-orders-createorders - Declared direct write: POST /orders/createorders. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
    operations post-orders-holduntil - Declared direct write: POST /orders/holduntil. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
    operations get-orders-listbytag - Declared direct read: GET /orders/listbytag. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
    operations post-orders-markasshipped - Declared direct write: POST /orders/markasshipped. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
    operations post-orders-removetag - Declared direct write: POST /orders/removetag. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
    operations post-orders-restorefromhold - Declared direct write: POST /orders/restorefromhold. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
    operations post-orders-unassignuser - Declared direct write: POST /orders/unassignuser. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
    operations get-products - Declared etl: GET /products. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
    operations get-products-product-id - Declared direct read: GET /products/{productId}. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
    operations put-products-product-id - Declared direct write: PUT /products/{productId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
    operations get-shipments - Declared etl: GET /shipments. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
    operations post-shipments-createlabel - Declared direct write: POST /shipments/createlabel. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Legacy connector never implemented label purchasing; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
    operations post-shipments-getrates - Declared direct write: POST /shipments/getrates. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
    operations post-shipments-voidlabel - Declared direct write: POST /shipments/voidlabel. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
    operations get-stores - Declared direct read: GET /stores. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
    operations get-stores-store-id - Declared direct read: GET /stores/{storeId}. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
    operations put-stores-store-id - Declared direct write: PUT /stores/{storeId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
    operations post-stores-deactivate - Declared direct write: POST /stores/deactivate. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
    operations get-stores-getrefreshstatus - Declared direct read: GET /stores/getrefreshstatus. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
    operations get-stores-marketplaces - Declared direct read: GET /stores/marketplaces. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
    operations post-stores-reactivate - Declared direct write: POST /stores/reactivate. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
    operations post-stores-refreshstore - Declared direct write: POST /stores/refreshstore. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
    operations get-users - Declared direct read: GET /users. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
    operations get-warehouses - Declared direct read: GET /warehouses. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
    operations delete-warehouses-warehouse-id - Declared direct write: DELETE /warehouses/{warehouseId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Documented provider deletion is not promoted without an operation-specific plan, preview, approval, typed confirmation, and execution contract.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
    operations get-warehouses-warehouse-id - Declared direct read: GET /warehouses/{warehouseId}. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
    operations post-warehouses-createwarehouse - Declared direct write: POST /warehouses/createwarehouse. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
    operations put-warehouses-updatewarehouse - Declared direct write: PUT /warehouses/updatewarehouse. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
    operations get-webhooks - Declared direct read: GET /webhooks. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
    operations delete-webhooks-webhook-id - Declared direct write: DELETE /webhooks/{webhookId}. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute with destructive confirmation; risk: Documented provider deletion is not promoted without an operation-specific plan, preview, approval, typed confirmation, and execution contract.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
    operations post-webhooks-subscribe - Declared direct write: POST /webhooks/subscribe. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: /webhooks/subscribe is a mutation endpoint; this bundle has no reviewed write record schema, approval risk, or legacy write contract for that operation; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect shipstation

  # Inspect as structured JSON
  pm connectors inspect shipstation --json

AGENT WORKFLOW
  - Run pm connectors inspect shipstation before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
