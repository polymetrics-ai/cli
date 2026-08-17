# pm connectors inspect picqer

```text
NAME
  pm connectors inspect picqer - Picqer connector manual

SYNOPSIS
  pm connectors inspect picqer
  pm connectors inspect picqer --json
  pm credentials add <name> --connector picqer [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Picqer products, customers, orders, picklists, warehouses, suppliers, purchase orders, returns, and warehouse-operations reference data, and writes order/purchase-order/return lifecycle and catalog mutations through the Picqer REST API.

ICON
  asset: icons/picqer.svg
  source: official
  review_status: official_verified
  review_url: https://picqer.com/en/api

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  max_pages
  mode
  organization_name
  page_size
  username
  api_key (secret)
  password (secret)

ETL STREAMS
  products:
    primary key: id
    fields: active(), created(), id(), idproduct(), name(), price(), productcode(), stock(), updated()
  customers:
    primary key: id
    fields: contactname(), created(), email(), id(), idcustomer(), name(), telephone(), updated()
  orders:
    primary key: id
    fields: created(), id(), idorder(), orderid(), status(), updated()
  picklists:
    primary key: id
    fields: created(), id(), idpicklist(), picklistid(), status(), updated()
  warehouses:
    primary key: id
    fields: active(), created(), id(), idwarehouse(), name(), updated()
  suppliers:
    primary key: id
    fields: contactname(), created(), emailaddress(), id(), idsupplier(), name(), updated()
  tags:
    primary key: id
    fields: color(), id(), idtag(), inherit(), textColor(), title()
  purchaseorders:
    primary key: id
    fields: created(), delivery_date(), id(), idpurchaseorder(), idsupplier(), idwarehouse(), products(), purchaseorderid(), remarks(), status(), supplier_name(), supplier_orderid(), updated()
  receipts:
    primary key: id
    fields: amount_received(), completed_at(), completed_by(), created(), id(), idreceipt(), idwarehouse(), products(), purchaseorder(), receiptid(), remarks(), status(), supplier(), updated(), version()
  returns:
    primary key: id
    fields: address(), city(), contactname(), country(), emailaddress(), id(), idcustomer(), idorder(), idreturn(), idreturn_status(), idtemplate(), name(), returnid(), zipcode()
  return_statuses:
    primary key: id
    fields: color(), completed(), created_at(), default(), id(), idreturn_status(), name(), updated_at()
  return_reasons:
    primary key: id
    fields: created_at(), default(), id(), idreturn_reason(), name(), updated_at()
  backorders:
    primary key: id
    fields: amount(), amountavailable(), has_parts(), id(), idbackorder(), idcustomer(), idorder(), idorder_product(), idproduct(), idreturn(), idwarehouse(), priority()
  comments:
    primary key: id
    fields: author(), author_type(), body(), id(), idcomment(), mentions(), source(), source_type()
  stockhistory:
    primary key: id
    fields: change_type(), changed_at(), id(), idlocation(), idproduct(), idproduct_stock_history(), iduser(), idwarehouse(), new_stock(), old_stock(), reason(), stock_change()
  users:
    primary key: id
    fields: active(), admin(), created_at(), emailaddress(), first_name(), id(), idpacking_station(), iduser(), language(), last_login_at(), last_name(), updated_at(), username()
  product_fields:
    primary key: id
    fields: id(), idproductfield(), required(), title(), type(), values(), visible_invoice(), visible_picklist(), visible_portal(), visible_purchase_order(), visible_shippinglist()
  customer_fields:
    primary key: id
    fields: id(), idcustomerfield(), required(), title(), type(), values()
  order_fields:
    primary key: id
    fields: id(), idorderfield(), only_accessible_via_api(), required(), title(), type(), values(), visible_picklist(), visible_portal()
  pricelists:
    primary key: id
    fields: id(), idpricelist(), name()
  shippingproviders:
    primary key: id
    fields: active(), created(), id(), idshippingprovider(), name(), profiles(), provider(), updated()
  vatgroups:
    primary key: id
    fields: id(), idvatgroup(), name(), percentage()
  locations:
    primary key: id
    fields: id(), idlocation(), idwarehouse(), is_bulk_location(), is_exclusive_location(), location_type(), name(), parent_idlocation(), remarks(), type(), unlink_on_empty()
  location_types:
    primary key: id
    fields: color(), default(), id(), idlocation_type(), name()
  picking_containers:
    primary key: id
    fields: created_at(), id(), idpicking_container(), idpicklist(), name(), updated_at()
  picklist_batches:
    primary key: id
    fields: assigned_to(), completed_at(), completed_by(), created_at(), display_title(), id(), idpicklist_batch(), idwarehouse(), picklist_batchid(), status(), title(), total_picklists(), total_products(), type(), updated_at()
  shipments:
    primary key: id
    fields: cancelled(), created(), id(), idorder(), idpackaging(), idpicklist(), idreturn(), idshipment(), idshippingprovider(), parcels(), provider(), providername(), updated(), weight()
  packagings:
    primary key: id
    fields: active(), barcode(), created_at(), height(), id(), idpackaging(), length(), name(), updated_at(), use_in_auto_advice(), width()
  packingstations:
    primary key: id
    fields: id(), idpacking_station(), name(), printer_packinglists(), printer_product_labels(), printer_shipping_documents(), printer_shipping_labels()
  webshoporders:
    primary key: id
    fields: created(), foreign_id(), foreign_number(), foreign_status(), id(), idcompany_webshop(), idcompany_webshop_order(), idorder(), ordered(), reason(), status(), updated()
  hooks:
    primary key: id
    fields: active(), address(), event(), id(), idhook(), name(), secret()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_customer:
    endpoint: POST /customers
    risk: creates a new WMS customer record; low-risk external mutation, no approval required
  update_customer:
    endpoint: PUT /customers/{{ record.idcustomer }}
    required fields: idcustomer
    risk: updates an existing customer's contact details; external mutation, approval required
  delete_customer:
    endpoint: DELETE /customers/{{ record.idcustomer }}
    required fields: idcustomer
    risk: permanently deletes a customer record; destructive external mutation, approval required
  create_supplier:
    endpoint: POST /suppliers
    risk: creates a new supplier record; low-risk external mutation, no approval required
  update_supplier:
    endpoint: PUT /suppliers/{{ record.idsupplier }}
    required fields: idsupplier
    risk: updates an existing supplier's contact details; external mutation, approval required
  create_tag:
    endpoint: POST /tags
    risk: creates a new tag; low-risk external mutation, no approval required
  update_tag:
    endpoint: PUT /tags/{{ record.idtag }}
    required fields: idtag
    risk: updates an existing tag's title/color/inherit setting; external mutation, approval required
  delete_tag:
    endpoint: DELETE /tags/{{ record.idtag }}
    required fields: idtag
    risk: permanently deletes a tag and removes it from every linked order/customer/product; destructive external mutation, approval required
  update_product:
    endpoint: PUT /products/{{ record.idproduct }}
    required fields: idproduct
    risk: updates an existing product's catalog fields (name/price/active status); external mutation, approval required
  pause_order:
    endpoint: POST /orders/{{ record.idorder }}/pause
    required fields: idorder
    risk: pauses picking/fulfillment of an order; external mutation, approval required
  resume_order:
    endpoint: POST /orders/{{ record.idorder }}/resume
    required fields: idorder
    risk: resumes picking/fulfillment of a paused order; external mutation, approval required
  reopen_order:
    endpoint: POST /orders/{{ record.idorder }}/reopen
    required fields: idorder
    risk: reopens a completed/closed order for further processing; external mutation, approval required
  cancel_order:
    endpoint: DELETE /orders/{{ record.idorder }}
    required fields: idorder
    risk: cancels an order (Picqer's DELETE-shaped cancel, reversible via undo-cancellation, but stops all further fulfillment immediately); destructive external mutation, approval required
  create_purchaseorder:
    endpoint: POST /purchaseorders
    risk: creates a new purchase order (concept status) with an optional initial product line list; low-risk external mutation, no approval required
  mark_purchaseorder_as_purchased:
    endpoint: POST /purchaseorders/{{ record.idpurchaseorder }}/mark-as-purchased
    required fields: idpurchaseorder
    risk: transitions a concept purchase order to purchased status, signalling it has been placed with the supplier; external mutation, approval required
  close_purchaseorder:
    endpoint: POST /purchaseorders/{{ record.idpurchaseorder }}/close
    required fields: idpurchaseorder
    risk: closes a purchase order, marking it as finished even if not all products were received; external mutation, approval required
  cancel_purchaseorder:
    endpoint: POST /purchaseorders/{{ record.idpurchaseorder }}/cancel
    required fields: idpurchaseorder
    risk: cancels a purchase order; destructive external mutation, approval required
  create_receipt:
    endpoint: POST /receipts
    risk: starts a new goods-receiving session against a purchase order (Picqer's v2 receipts API also accepts idsupplier in place of idpurchaseorder for supplier-only receiving; this action only models the idpurchaseorder-required shape, see docs.md Known limits); low-risk external mutation, no approval required
  complete_receipt:
    endpoint: PUT /receipts/{{ record.idreceipt }}
    required fields: idreceipt
    optional fields: status
    risk: marks a goods-receiving session as complete (Picqer's documented PUT /receipts/{idreceipt} {"status": "completed"} shape), finalizing received stock quantities in the background; external mutation, approval required
  create_return:
    endpoint: POST /returns
    risk: creates a new customer return record; low-risk external mutation, no approval required
  update_return:
    endpoint: PUT /returns/{{ record.idreturn }}
    required fields: idreturn
    risk: updates an existing return's status/contact details; external mutation, approval required
  delete_return:
    endpoint: DELETE /returns/{{ record.idreturn }}
    required fields: idreturn
    risk: permanently deletes a return record; destructive external mutation, approval required
  process_backorders:
    endpoint: POST /backorders/process
    risk: processes one or more backorders for fulfillment now that stock is available; external mutation, approval required
  create_location:
    endpoint: POST /locations
    risk: creates a new warehouse storage location; low-risk external mutation, no approval required
  update_location:
    endpoint: PUT /locations/{{ record.idlocation }}
    required fields: idlocation
    risk: updates an existing warehouse location's name/remarks; external mutation, approval required
  delete_location:
    endpoint: DELETE /locations/{{ record.idlocation }}
    required fields: idlocation
    risk: permanently deletes a warehouse storage location; destructive external mutation, approval required
  create_location_type:
    endpoint: POST /location_types
    risk: creates a new location type; low-risk external mutation, no approval required
  update_location_type:
    endpoint: PUT /location_types/{{ record.id }}
    required fields: id
    risk: updates an existing location type's name/color; external mutation, approval required
  create_picking_container:
    endpoint: POST /picking-containers
    risk: creates a new picking container; low-risk external mutation, no approval required
  update_picking_container:
    endpoint: PUT /picking-containers/{{ record.idpicking_container }}
    required fields: idpicking_container
    risk: updates an existing picking container's name; external mutation, approval required
  create_picklist_batch:
    endpoint: POST /picklists/batches
    risk: creates a new picklist batch for warehouse picking; low-risk external mutation, no approval required
  create_shipment:
    endpoint: POST /picklists/{{ record.idpicklist }}/shipments
    required fields: idpicklist
    risk: creates a shipment for a picklist, booking it with the configured shipping provider and generating a shipping label; external mutation, approval required
  create_packaging:
    endpoint: POST /packagings
    risk: creates a new packaging type; low-risk external mutation, no approval required
  update_packaging:
    endpoint: PUT /packagings/{{ record.id }}
    required fields: id
    risk: updates an existing packaging type's dimensions/active status; external mutation, approval required
  create_hook:
    endpoint: POST /hooks
    risk: registers a new webhook subscription that will receive event notifications; low-risk external mutation, no approval required
  delete_hook:
    endpoint: DELETE /hooks/{{ record.idhook }}
    required fields: idhook
    risk: deactivates a webhook subscription (Picqer's DELETE call deactivates rather than hard-deletes the record); destructive external mutation, approval required

SECURITY
  read risk: external Picqer API read of warehouse management data
  write risk: creates/updates/deletes customers, suppliers, tags, locations, purchase orders, returns, and warehouse-operations records, and mutates order fulfillment lifecycle (pause/resume/reopen/cancel)
  approval: required for update/delete/cancel-shaped actions; create_customer/create_supplier/create_tag/create_purchaseorder/create_receipt/create_return/create_location/create_location_type/create_picking_container/create_picklist_batch/create_packaging/create_hook require no approval (low-risk, non-destructive)
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect picqer

  # Inspect as structured JSON
  pm connectors inspect picqer --json

AGENT WORKFLOW
  - Run pm connectors inspect picqer before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
