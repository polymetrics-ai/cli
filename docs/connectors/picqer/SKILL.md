---
name: pm-picqer
description: Picqer connector knowledge and safe action guide.
---

# pm-picqer

## Purpose

Reads Picqer products, customers, orders, picklists, warehouses, suppliers, purchase orders, returns, and warehouse-operations reference data, and writes order/purchase-order/return lifecycle and catalog mutations through the Picqer REST API.

## Icon

- id: picqer
- asset: icons/picqer.svg
- source: official
- review_status: official_verified
- review_url: https://picqer.com/en/api

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- max_pages
- mode
- organization_name (required)
- page_size
- username
- api_key (secret)
- password (secret)

## ETL Streams

- products:
  - primary key: id
  - fields: active(boolean), created(string), id(integer), idproduct(integer), name(string), price(number), productcode(string), stock(array), updated(string)
- customers:
  - primary key: id
  - fields: contactname(string), created(string), email(string), id(integer), idcustomer(integer), name(string), telephone(string), updated(string)
- orders:
  - primary key: id
  - fields: created(string), id(integer), idorder(integer), orderid(string), status(string), updated(string)
- picklists:
  - primary key: id
  - fields: created(string), id(integer), idpicklist(integer), picklistid(string), status(string), updated(string)
- warehouses:
  - primary key: id
  - fields: active(boolean), created(string), id(integer), idwarehouse(integer), name(string), updated(string)
- suppliers:
  - primary key: id
  - fields: contactname(string), created(string), emailaddress(string), id(integer), idsupplier(integer), name(string), updated(string)
- tags:
  - primary key: id
  - fields: color(string), id(integer), idtag(integer), inherit(boolean), textColor(string), title(string)
- purchaseorders:
  - primary key: id
  - fields: created(string), delivery_date(string), id(integer), idpurchaseorder(integer), idsupplier(integer), idwarehouse(integer), products(array), purchaseorderid(string), remarks(string), status(string), supplier_name(string), supplier_orderid(string), updated(string)
- receipts:
  - primary key: id
  - fields: amount_received(integer), completed_at(string), completed_by(object), created(string), id(integer), idreceipt(integer), idwarehouse(integer), products(array), purchaseorder(object), receiptid(string), remarks(string), status(string), supplier(object), updated(string), version(integer)
- returns:
  - primary key: id
  - fields: address(string), city(string), contactname(string), country(string), emailaddress(string), id(integer), idcustomer(integer), idorder(integer), idreturn(integer), idreturn_status(integer), idtemplate(integer), name(string), returnid(string), zipcode(string)
- return_statuses:
  - primary key: id
  - fields: color(string), completed(boolean), created_at(string), default(boolean), id(integer), idreturn_status(integer), name(string), updated_at(string)
- return_reasons:
  - primary key: id
  - fields: created_at(string), default(boolean), id(integer), idreturn_reason(integer), name(string), updated_at(string)
- backorders:
  - primary key: id
  - fields: amount(integer), amountavailable(integer), has_parts(boolean), id(integer), idbackorder(integer), idcustomer(integer), idorder(integer), idorder_product(integer), idproduct(integer), idreturn(integer), idwarehouse(integer), priority(integer)
- comments:
  - primary key: id
  - fields: author(object), author_type(string), body(string), id(integer), idcomment(integer), mentions(array), source(object), source_type(string)
- stockhistory:
  - primary key: id
  - fields: change_type(string), changed_at(string), id(integer), idlocation(integer), idproduct(integer), idproduct_stock_history(integer), iduser(integer), idwarehouse(integer), new_stock(integer), old_stock(integer), reason(string), stock_change(integer)
- users:
  - primary key: id
  - fields: active(boolean), admin(boolean), created_at(string), emailaddress(string), first_name(string), id(integer), idpacking_station(integer), iduser(integer), language(string), last_login_at(string), last_name(string), updated_at(string), username(string)
- product_fields:
  - primary key: id
  - fields: id(integer), idproductfield(integer), required(boolean), title(string), type(string), values(array), visible_invoice(boolean), visible_picklist(boolean), visible_portal(boolean), visible_purchase_order(boolean), visible_shippinglist(boolean)
- customer_fields:
  - primary key: id
  - fields: id(integer), idcustomerfield(integer), required(boolean), title(string), type(string), values(array)
- order_fields:
  - primary key: id
  - fields: id(integer), idorderfield(integer), only_accessible_via_api(boolean), required(boolean), title(string), type(string), values(array), visible_picklist(boolean), visible_portal(boolean)
- pricelists:
  - primary key: id
  - fields: id(integer), idpricelist(integer), name(string)
- shippingproviders:
  - primary key: id
  - fields: active(boolean), created(string), id(integer), idshippingprovider(integer), name(string), profiles(array), provider(string), updated(string)
- vatgroups:
  - primary key: id
  - fields: id(integer), idvatgroup(integer), name(string), percentage(number)
- locations:
  - primary key: id
  - fields: id(integer), idlocation(integer), idwarehouse(integer), is_bulk_location(boolean), is_exclusive_location(boolean), location_type(object), name(string), parent_idlocation(integer), remarks(string), type(string), unlink_on_empty(boolean)
- location_types:
  - primary key: id
  - fields: color(string), default(boolean), id(integer), idlocation_type(integer), name(string)
- picking_containers:
  - primary key: id
  - fields: created_at(string), id(integer), idpicking_container(integer), idpicklist(integer), name(string), updated_at(string)
- picklist_batches:
  - primary key: id
  - fields: assigned_to(object), completed_at(string), completed_by(object), created_at(string), display_title(string), id(integer), idpicklist_batch(integer), idwarehouse(integer), picklist_batchid(integer), status(string), title(string), total_picklists(integer), total_products(integer), type(string), updated_at(string)
- shipments:
  - primary key: id
  - fields: cancelled(boolean), created(string), id(integer), idorder(integer), idpackaging(integer), idpicklist(integer), idreturn(integer), idshipment(integer), idshippingprovider(integer), parcels(array), provider(string), providername(string), updated(string), weight(integer)
- packagings:
  - primary key: id
  - fields: active(boolean), barcode(string), created_at(string), height(integer), id(integer), idpackaging(integer), length(integer), name(string), updated_at(string), use_in_auto_advice(boolean), width(integer)
- packingstations:
  - primary key: id
  - fields: id(integer), idpacking_station(integer), name(string), printer_packinglists(object), printer_product_labels(object), printer_shipping_documents(object), printer_shipping_labels(object)
- webshoporders:
  - primary key: id
  - fields: created(string), foreign_id(string), foreign_number(string), foreign_status(string), id(integer), idcompany_webshop(integer), idcompany_webshop_order(integer), idorder(integer), ordered(string), reason(string), status(string), updated(string)
- hooks:
  - primary key: id
  - fields: active(boolean), address(string), event(string), id(integer), idhook(integer), name(string), secret(boolean)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite

## Reverse ETL Actions

- create_customer:
  - endpoint: POST /customers
  - required fields: name
  - risk: creates a new WMS customer record; low-risk external mutation, no approval required
- update_customer:
  - endpoint: PUT /customers/{{ record.idcustomer }}
  - required fields: idcustomer
  - risk: updates an existing customer's contact details; external mutation, approval required
- delete_customer:
  - endpoint: DELETE /customers/{{ record.idcustomer }}
  - required fields: idcustomer
  - risk: permanently deletes a customer record; destructive external mutation, approval required
- create_supplier:
  - endpoint: POST /suppliers
  - required fields: name
  - risk: creates a new supplier record; low-risk external mutation, no approval required
- update_supplier:
  - endpoint: PUT /suppliers/{{ record.idsupplier }}
  - required fields: idsupplier
  - risk: updates an existing supplier's contact details; external mutation, approval required
- create_tag:
  - endpoint: POST /tags
  - required fields: title, color, inherit
  - risk: creates a new tag; low-risk external mutation, no approval required
- update_tag:
  - endpoint: PUT /tags/{{ record.idtag }}
  - required fields: idtag
  - risk: updates an existing tag's title/color/inherit setting; external mutation, approval required
- delete_tag:
  - endpoint: DELETE /tags/{{ record.idtag }}
  - required fields: idtag
  - risk: permanently deletes a tag and removes it from every linked order/customer/product; destructive external mutation, approval required
- update_product:
  - endpoint: PUT /products/{{ record.idproduct }}
  - required fields: idproduct
  - risk: updates an existing product's catalog fields (name/price/active status); external mutation, approval required
- pause_order:
  - endpoint: POST /orders/{{ record.idorder }}/pause
  - required fields: idorder
  - risk: pauses picking/fulfillment of an order; external mutation, approval required
- resume_order:
  - endpoint: POST /orders/{{ record.idorder }}/resume
  - required fields: idorder
  - risk: resumes picking/fulfillment of a paused order; external mutation, approval required
- reopen_order:
  - endpoint: POST /orders/{{ record.idorder }}/reopen
  - required fields: idorder
  - risk: reopens a completed/closed order for further processing; external mutation, approval required
- cancel_order:
  - endpoint: DELETE /orders/{{ record.idorder }}
  - required fields: idorder
  - risk: cancels an order (Picqer's DELETE-shaped cancel, reversible via undo-cancellation, but stops all further fulfillment immediately); destructive external mutation, approval required
- create_purchaseorder:
  - endpoint: POST /purchaseorders
  - required fields: idsupplier, idwarehouse
  - risk: creates a new purchase order (concept status) with an optional initial product line list; low-risk external mutation, no approval required
- mark_purchaseorder_as_purchased:
  - endpoint: POST /purchaseorders/{{ record.idpurchaseorder }}/mark-as-purchased
  - required fields: idpurchaseorder
  - risk: transitions a concept purchase order to purchased status, signalling it has been placed with the supplier; external mutation, approval required
- close_purchaseorder:
  - endpoint: POST /purchaseorders/{{ record.idpurchaseorder }}/close
  - required fields: idpurchaseorder
  - risk: closes a purchase order, marking it as finished even if not all products were received; external mutation, approval required
- cancel_purchaseorder:
  - endpoint: POST /purchaseorders/{{ record.idpurchaseorder }}/cancel
  - required fields: idpurchaseorder
  - risk: cancels a purchase order; destructive external mutation, approval required
- create_receipt:
  - endpoint: POST /receipts
  - required fields: idpurchaseorder
  - risk: starts a new goods-receiving session against a purchase order (Picqer's v2 receipts API also accepts idsupplier in place of idpurchaseorder for supplier-only receiving; this action only models the idpurchaseorder-required shape, see docs.md Known limits); low-risk external mutation, no approval required
- complete_receipt:
  - endpoint: PUT /receipts/{{ record.idreceipt }}
  - required fields: idreceipt, status
  - risk: marks a goods-receiving session as complete (Picqer's documented PUT /receipts/{idreceipt} {"status": "completed"} shape), finalizing received stock quantities in the background; external mutation, approval required
- create_return:
  - endpoint: POST /returns
  - required fields: name
  - risk: creates a new customer return record; low-risk external mutation, no approval required
- update_return:
  - endpoint: PUT /returns/{{ record.idreturn }}
  - required fields: idreturn
  - risk: updates an existing return's status/contact details; external mutation, approval required
- delete_return:
  - endpoint: DELETE /returns/{{ record.idreturn }}
  - required fields: idreturn
  - risk: permanently deletes a return record; destructive external mutation, approval required
- process_backorders:
  - endpoint: POST /backorders/process
  - risk: processes one or more backorders for fulfillment now that stock is available; external mutation, approval required
- create_location:
  - endpoint: POST /locations
  - required fields: name, idwarehouse
  - risk: creates a new warehouse storage location; low-risk external mutation, no approval required
- update_location:
  - endpoint: PUT /locations/{{ record.idlocation }}
  - required fields: idlocation
  - risk: updates an existing warehouse location's name/remarks; external mutation, approval required
- delete_location:
  - endpoint: DELETE /locations/{{ record.idlocation }}
  - required fields: idlocation
  - risk: permanently deletes a warehouse storage location; destructive external mutation, approval required
- create_location_type:
  - endpoint: POST /location_types
  - required fields: name, color
  - risk: creates a new location type; low-risk external mutation, no approval required
- update_location_type:
  - endpoint: PUT /location_types/{{ record.id }}
  - required fields: id
  - risk: updates an existing location type's name/color; external mutation, approval required
- create_picking_container:
  - endpoint: POST /picking-containers
  - required fields: name
  - risk: creates a new picking container; low-risk external mutation, no approval required
- update_picking_container:
  - endpoint: PUT /picking-containers/{{ record.idpicking_container }}
  - required fields: idpicking_container
  - risk: updates an existing picking container's name; external mutation, approval required
- create_picklist_batch:
  - endpoint: POST /picklists/batches
  - required fields: idwarehouse
  - risk: creates a new picklist batch for warehouse picking; low-risk external mutation, no approval required
- create_shipment:
  - endpoint: POST /picklists/{{ record.idpicklist }}/shipments
  - required fields: idpicklist, idshippingprovider_profile
  - risk: creates a shipment for a picklist, booking it with the configured shipping provider and generating a shipping label; external mutation, approval required
- create_packaging:
  - endpoint: POST /packagings
  - required fields: name
  - risk: creates a new packaging type; low-risk external mutation, no approval required
- update_packaging:
  - endpoint: PUT /packagings/{{ record.id }}
  - required fields: id
  - risk: updates an existing packaging type's dimensions/active status; external mutation, approval required
- create_hook:
  - endpoint: POST /hooks
  - required fields: name, event, address
  - risk: registers a new webhook subscription that will receive event notifications; low-risk external mutation, no approval required
- delete_hook:
  - endpoint: DELETE /hooks/{{ record.idhook }}
  - required fields: idhook
  - risk: deactivates a webhook subscription (Picqer's DELETE call deactivates rather than hard-deletes the record); destructive external mutation, approval required

## Security

- read risk: external Picqer API read of warehouse management data
- write risk: creates/updates/deletes customers, suppliers, tags, locations, purchase orders, returns, and warehouse-operations records, and mutates order fulfillment lifecycle (pause/resume/reopen/cancel)
- approval: required for update/delete/cancel-shaped actions; create_customer/create_supplier/create_tag/create_purchaseorder/create_receipt/create_return/create_location/create_location_type/create_picking_container/create_picklist_batch/create_packaging/create_hook require no approval (low-risk, non-destructive)
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect picqer
```

### Inspect as structured JSON

```bash
pm connectors inspect picqer --json
```

## Agent Rules

- Run pm connectors inspect picqer before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
