# pm connectors inspect easypost

```text
NAME
  pm connectors inspect easypost - EasyPost connector manual

SYNOPSIS
  pm connectors inspect easypost
  pm connectors inspect easypost --json
  pm credentials add <name> --connector easypost [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes EasyPost shipping resources including shipments, trackers, addresses, parcels, batches, events, claims, pickups, refunds, scan forms, end shippers, users, and webhooks through the EasyPost REST API.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  start_date
  username (secret)

ETL STREAMS
  shipments:
    primary key: id
    cursor: created_at
    fields: batch_id(), batch_status(), created_at(), id(), is_return(), mode(), object(), reference(), status(), tracking_code(), updated_at()
  trackers:
    primary key: id
    cursor: created_at
    fields: carrier(), created_at(), est_delivery_date(), id(), mode(), object(), shipment_id(), signed_by(), status(), status_detail(), tracking_code(), updated_at()
  addresses:
    primary key: id
    cursor: created_at
    fields: city(), company(), country(), created_at(), email(), id(), mode(), name(), object(), phone(), residential(), state(), street1(), street2(), updated_at(), zip()
  parcels:
    primary key: id
    cursor: created_at
    fields: created_at(), height(), id(), length(), mode(), object(), predefined_package(), updated_at(), weight(), width()
  insurances:
    primary key: id
    cursor: created_at
    fields: amount(), created_at(), id(), mode(), object(), provider(), reference(), shipment_id(), status(), tracking_code(), updated_at()
  batches:
    primary key: id
    cursor: created_at
    fields: created_at(), id(), label_url(), mode(), num_shipments(), object(), pickup(), reference(), scan_form(), shipments(), state(), status(), updated_at()
  carrier_accounts:
    primary key: id
    fields: billing_type(), clone(), created_at(), description(), fields(), id(), logo(), object(), readable(), reference(), type(), updated_at()
  carrier_metadata:
    primary key: name
    fields: human_readable(), name(), predefined_packages(), service_levels()
  carrier_types:
    primary key: type
    fields: fields(), logo(), object(), readable(), type()
  end_shippers:
    primary key: id
    fields: city(), company(), country(), created_at(), email(), id(), mode(), name(), object(), phone(), state(), street1(), street2(), updated_at(), zip()
  events:
    primary key: id
    cursor: created_at
    fields: created_at(), description(), id(), mode(), object(), status(), user_id()
  claims:
    primary key: id
    cursor: created_at
    fields: approved_amount(), attachments(), created_at(), history(), id(), insurance_id(), mode(), object(), requested_amount(), shipment_id(), status(), status_detail(), tracking_code(), type(), updated_at()
  pickups:
    primary key: id
    cursor: created_at
    fields: address(), confirmation(), created_at(), id(), instructions(), is_account_address(), max_datetime(), min_datetime(), mode(), object(), pickup_rates(), reference(), status(), updated_at()
  refunds:
    primary key: id
    cursor: created_at
    fields: carrier(), confirmation_number(), created_at(), id(), object(), shipment_id(), status(), tracking_code(), updated_at()
  scan_forms:
    primary key: id
    cursor: created_at
    fields: address(), batch_id(), confirmation(), created_at(), form_file_type(), form_url(), id(), message(), object(), status(), tracking_codes(), updated_at()
  child_users:
    primary key: id
    fields: children(), created_at(), default_carbon_offset(), id(), name(), object(), parent_id(), phone_number(), verified()
  referral_customers:
    primary key: id
    fields: balance(), children(), created_at(), default_carbon_offset(), email(), id(), name(), object(), parent_id(), phone_number(), price_per_shipment(), verified()
  webhooks:
    primary key: id
    fields: created_at(), custom_headers(), disabled_at(), id(), mode(), object(), url()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_address:
    endpoint: POST /addresses
    risk: creates a reusable EasyPost Address object; low-risk external mutation, approval required
  create_and_verify_address:
    endpoint: POST /addresses/create_and_verify
    risk: creates and verifies an EasyPost Address object; may return verification failures but does not buy postage, approval required
  create_parcel:
    endpoint: POST /parcels
    risk: creates a Parcel object describing package dimensions and weight; low-risk external mutation, approval required
  create_customs_item:
    endpoint: POST /customs_items
    risk: creates a CustomsItem declaration object used by international shipments; approval required
  create_customs_info:
    endpoint: POST /customs_infos
    risk: creates a CustomsInfo declaration object used by international shipments; approval required
  create_shipment:
    endpoint: POST /shipments
    risk: creates and rates a Shipment object; does not purchase postage by itself, approval required
  buy_shipment:
    endpoint: POST /shipments/{{ record.id }}/buy
    required fields: id
    risk: purchases a live postage label for an existing Shipment and may incur carrier/account charges; approval required
  rerate_shipment:
    endpoint: POST /shipments/{{ record.id }}/rerate
    required fields: id
    risk: refreshes rates on an existing Shipment; external mutation of rated shipment state, approval required
  insure_shipment:
    endpoint: POST /shipments/{{ record.id }}/insure
    required fields: id
    risk: adds shipping insurance to an existing Shipment and may incur a charge; approval required
  refund_shipment:
    endpoint: POST /shipments/{{ record.id }}/refund
    required fields: id
    risk: requests a refund for an existing Shipment label; changes shipment refund state, approval required
  create_shipment_form:
    endpoint: POST /shipments/{{ record.id }}/forms
    required fields: id
    risk: creates a shipment-associated form/document metadata object; approval required
  create_tracker:
    endpoint: POST /trackers
    risk: creates a Tracker for a carrier tracking code; low-risk external mutation, approval required
  delete_tracker:
    endpoint: DELETE /trackers/{{ record.id }}
    required fields: id
    risk: deletes an EasyPost Tracker object; destructive external mutation, approval required
  create_batch:
    endpoint: POST /batches
    risk: creates a Batch grouping shipments; does not buy postage by itself, approval required
  add_shipments_to_batch:
    endpoint: POST /batches/{{ record.id }}/add_shipments
    required fields: id
    risk: adds Shipment references to an existing Batch; approval required
  remove_shipments_from_batch:
    endpoint: POST /batches/{{ record.id }}/remove_shipments
    required fields: id
    risk: removes Shipment references from an existing Batch; approval required
  buy_batch:
    endpoint: POST /batches/{{ record.id }}/buy
    required fields: id
    risk: purchases postage for all eligible shipments in a Batch and may incur charges; approval required
  label_batch:
    endpoint: POST /batches/{{ record.id }}/label
    required fields: id
    risk: generates a batch label file after purchase; external mutation of batch artifact state, approval required
  create_batch_scan_form:
    endpoint: POST /batches/{{ record.id }}/scan_form
    required fields: id
    risk: creates a ScanForm for an existing Batch; approval required
  create_end_shipper:
    endpoint: POST /end_shippers
    risk: creates an EndShipper sender identity/address record; approval required
  update_end_shipper:
    endpoint: PUT /end_shippers/{{ record.id }}
    required fields: id
    risk: updates an EndShipper sender identity/address record; approval required
  create_insurance:
    endpoint: POST /insurances
    risk: creates standalone shipping insurance and may incur a charge; approval required
  refund_insurance:
    endpoint: POST /insurances/{{ record.id }}/refund
    required fields: id
    risk: requests a refund for standalone insurance; approval required
  create_order:
    endpoint: POST /orders
    risk: creates an Order grouping multiple shipments; does not buy postage by itself, approval required
  cancel_claim:
    endpoint: POST /claims/{{ record.id }}/cancel
    required fields: id
    risk: cancels an existing EasyPost insurance claim; external claim workflow mutation, approval required
  buy_order:
    endpoint: POST /orders/{{ record.id }}/buy
    required fields: id
    risk: purchases postage for an Order and may incur charges; approval required
  create_pickup:
    endpoint: POST /pickups
    risk: creates a carrier pickup request for a shipment/address window; approval required
  buy_pickup:
    endpoint: POST /pickups/{{ record.id }}/buy
    required fields: id
    risk: buys/schedules a carrier pickup and may incur carrier charges; approval required
  cancel_pickup:
    endpoint: POST /pickups/{{ record.id }}/cancel
    required fields: id
    risk: cancels a scheduled Pickup; external operational mutation, approval required
  create_refund:
    endpoint: POST /refunds
    risk: creates one or more shipment refund requests; approval required
  create_scan_form:
    endpoint: POST /scan_forms
    risk: creates a ScanForm manifest for shipment IDs; approval required
  create_report:
    endpoint: POST /reports/{{ record.type }}
    required fields: type
    risk: starts an asynchronous EasyPost report export for the requested report type/date range; approval required
  create_luma_promise:
    endpoint: POST /luma/promise
    risk: requests a Luma delivery promise/rating calculation; no label purchase by itself, approval required
  create_luma_shipment:
    endpoint: POST /shipments/luma
    risk: creates and buys a Shipment through Luma one-call buy and may incur postage charges; approval required
  buy_luma_shipment:
    endpoint: POST /shipments/{{ record.id }}/luma
    required fields: id
    risk: buys postage for an existing Shipment through Luma and may incur charges; approval required
  create_child_user:
    endpoint: POST /users
    risk: creates a production-only child user/sub-account under the authenticated EasyPost account; elevated account-management mutation, approval required
  update_user:
    endpoint: PATCH /users/{{ record.id }}
    required fields: id
    risk: updates EasyPost user/sub-account settings such as child account name; elevated account-management mutation, approval required
  delete_child_user:
    endpoint: DELETE /users/{{ record.child_id }}
    required fields: child_id
    risk: removes a child user from the parent account; destructive account-management mutation, approval required
  create_webhook:
    endpoint: POST /webhooks
    risk: registers an outbound Webhook URL/custom headers for EasyPost events; approval required
  update_webhook:
    endpoint: PATCH /webhooks/{{ record.id }}
    required fields: id
    risk: updates Webhook delivery metadata such as custom headers; approval required
  delete_webhook:
    endpoint: DELETE /webhooks/{{ record.id }}
    required fields: id
    risk: deletes a Webhook subscription, stopping outbound event delivery; destructive external mutation, approval required

SECURITY
  read risk: external EasyPost API read of shipping, tracking, carrier metadata/account, event, claim, pickup, refund, scan-form, user, referral customer, and webhook data
  write risk: external EasyPost API mutation of live shipping objects, labels, pickups, insurance/refund workflows, reports, users, and webhooks; buy/refund/insurance/pickup/order/Luma actions may incur charges or alter operational shipping state
  approval: reverse ETL plan preview and approval required for every write action; charge-bearing/destructive actions are flagged in writes.json
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect easypost

  # Inspect as structured JSON
  pm connectors inspect easypost --json

AGENT WORKFLOW
  - Run pm connectors inspect easypost before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
