---
name: pm-easypost
description: EasyPost connector knowledge and safe action guide.
---

# pm-easypost

## Purpose

Reads and writes EasyPost shipping resources including shipments, trackers, addresses, parcels, batches, events, claims, pickups, refunds, scan forms, end shippers, users, and webhooks through the EasyPost REST API.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- start_date
- username (secret) (required)

## ETL Streams

- shipments:
  - primary key: id
  - cursor: created_at
  - fields: batch_id(string), batch_status(string), created_at(string), id(string), is_return(boolean), mode(string), object(string), reference(string), status(string), tracking_code(string), updated_at(string)
- trackers:
  - primary key: id
  - cursor: created_at
  - fields: carrier(string), created_at(string), est_delivery_date(string), id(string), mode(string), object(string), shipment_id(string), signed_by(string), status(string), status_detail(string), tracking_code(string), updated_at(string)
- addresses:
  - primary key: id
  - cursor: created_at
  - fields: city(string), company(string), country(string), created_at(string), email(string), id(string), mode(string), name(string), object(string), phone(string), residential(boolean), state(string), street1(string), street2(string), updated_at(string), zip(string)
- parcels:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), height(number), id(string), length(number), mode(string), object(string), predefined_package(string), updated_at(string), weight(number), width(number)
- insurances:
  - primary key: id
  - cursor: created_at
  - fields: amount(string), created_at(string), id(string), mode(string), object(string), provider(string), reference(string), shipment_id(string), status(string), tracking_code(string), updated_at(string)
- batches:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), id(string), label_url(string), mode(string), num_shipments(integer), object(string), pickup(object), reference(string), scan_form(object), shipments(array), state(string), status(object), updated_at(string)
- carrier_accounts:
  - primary key: id
  - fields: billing_type(string), clone(boolean), created_at(string), description(string), fields(object), id(string), logo(string), object(string), readable(string), reference(string), type(string), updated_at(string)
- carrier_metadata:
  - primary key: name
  - fields: human_readable(string), name(string), predefined_packages(array), service_levels(array)
- carrier_types:
  - primary key: type
  - fields: fields(object), logo(string), object(string), readable(string), type(string)
- end_shippers:
  - primary key: id
  - fields: city(string), company(string), country(string), created_at(string), email(string), id(string), mode(string), name(string), object(string), phone(string), state(string), street1(string), street2(string), updated_at(string), zip(string)
- events:
  - primary key: id
  - cursor: created_at
  - fields: created_at(string), description(string), id(string), mode(string), object(string), status(string), user_id(string)
- claims:
  - primary key: id
  - cursor: created_at
  - fields: approved_amount(string), attachments(array), created_at(string), history(array), id(string), insurance_id(string), mode(string), object(string), requested_amount(string), shipment_id(string), status(string), status_detail(string), tracking_code(string), type(string), updated_at(string)
- pickups:
  - primary key: id
  - cursor: created_at
  - fields: address(object), confirmation(string), created_at(string), id(string), instructions(string), is_account_address(boolean), max_datetime(string), min_datetime(string), mode(string), object(string), pickup_rates(array), reference(string), status(string), updated_at(string)
- refunds:
  - primary key: id
  - cursor: created_at
  - fields: carrier(string), confirmation_number(string), created_at(string), id(string), object(string), shipment_id(string), status(string), tracking_code(string), updated_at(string)
- scan_forms:
  - primary key: id
  - cursor: created_at
  - fields: address(object), batch_id(string), confirmation(string), created_at(string), form_file_type(string), form_url(string), id(string), message(string), object(string), status(string), tracking_codes(array), updated_at(string)
- child_users:
  - primary key: id
  - fields: children(array), created_at(string), default_carbon_offset(boolean), id(string), name(string), object(string), parent_id(string), phone_number(string), verified(boolean)
- referral_customers:
  - primary key: id
  - fields: balance(string), children(array), created_at(string), default_carbon_offset(boolean), email(string), id(string), name(string), object(string), parent_id(string), phone_number(string), price_per_shipment(string), verified(boolean)
- webhooks:
  - primary key: id
  - fields: created_at(string), custom_headers(array), disabled_at(string), id(string), mode(string), object(string), url(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_address:
  - endpoint: POST /addresses
  - required fields: address
  - risk: creates a reusable EasyPost Address object; low-risk external mutation, approval required
- create_and_verify_address:
  - endpoint: POST /addresses/create_and_verify
  - required fields: address
  - risk: creates and verifies an EasyPost Address object; may return verification failures but does not buy postage, approval required
- create_parcel:
  - endpoint: POST /parcels
  - required fields: parcel
  - risk: creates a Parcel object describing package dimensions and weight; low-risk external mutation, approval required
- create_customs_item:
  - endpoint: POST /customs_items
  - required fields: customs_item
  - risk: creates a CustomsItem declaration object used by international shipments; approval required
- create_customs_info:
  - endpoint: POST /customs_infos
  - required fields: customs_info
  - risk: creates a CustomsInfo declaration object used by international shipments; approval required
- create_shipment:
  - endpoint: POST /shipments
  - required fields: shipment
  - risk: creates and rates a Shipment object; does not purchase postage by itself, approval required
- buy_shipment:
  - endpoint: POST /shipments/{{ record.id }}/buy
  - required fields: id, rate
  - risk: purchases a live postage label for an existing Shipment and may incur carrier/account charges; approval required
- rerate_shipment:
  - endpoint: POST /shipments/{{ record.id }}/rerate
  - required fields: id
  - risk: refreshes rates on an existing Shipment; external mutation of rated shipment state, approval required
- insure_shipment:
  - endpoint: POST /shipments/{{ record.id }}/insure
  - required fields: id, amount
  - risk: adds shipping insurance to an existing Shipment and may incur a charge; approval required
- refund_shipment:
  - endpoint: POST /shipments/{{ record.id }}/refund
  - required fields: id
  - risk: requests a refund for an existing Shipment label; changes shipment refund state, approval required
- create_shipment_form:
  - endpoint: POST /shipments/{{ record.id }}/forms
  - required fields: id, form
  - risk: creates a shipment-associated form/document metadata object; approval required
- create_tracker:
  - endpoint: POST /trackers
  - required fields: tracker
  - risk: creates a Tracker for a carrier tracking code; low-risk external mutation, approval required
- delete_tracker:
  - endpoint: DELETE /trackers/{{ record.id }}
  - required fields: id
  - risk: deletes an EasyPost Tracker object; destructive external mutation, approval required
- create_batch:
  - endpoint: POST /batches
  - required fields: batch
  - risk: creates a Batch grouping shipments; does not buy postage by itself, approval required
- add_shipments_to_batch:
  - endpoint: POST /batches/{{ record.id }}/add_shipments
  - required fields: id, shipments
  - risk: adds Shipment references to an existing Batch; approval required
- remove_shipments_from_batch:
  - endpoint: POST /batches/{{ record.id }}/remove_shipments
  - required fields: id, shipments
  - risk: removes Shipment references from an existing Batch; approval required
- buy_batch:
  - endpoint: POST /batches/{{ record.id }}/buy
  - required fields: id
  - risk: purchases postage for all eligible shipments in a Batch and may incur charges; approval required
- label_batch:
  - endpoint: POST /batches/{{ record.id }}/label
  - required fields: id
  - risk: generates a batch label file after purchase; external mutation of batch artifact state, approval required
- create_batch_scan_form:
  - endpoint: POST /batches/{{ record.id }}/scan_form
  - required fields: id
  - risk: creates a ScanForm for an existing Batch; approval required
- create_end_shipper:
  - endpoint: POST /end_shippers
  - required fields: end_shipper
  - risk: creates an EndShipper sender identity/address record; approval required
- update_end_shipper:
  - endpoint: PUT /end_shippers/{{ record.id }}
  - required fields: id, end_shipper
  - risk: updates an EndShipper sender identity/address record; approval required
- create_insurance:
  - endpoint: POST /insurances
  - required fields: insurance
  - risk: creates standalone shipping insurance and may incur a charge; approval required
- refund_insurance:
  - endpoint: POST /insurances/{{ record.id }}/refund
  - required fields: id
  - risk: requests a refund for standalone insurance; approval required
- create_order:
  - endpoint: POST /orders
  - required fields: order
  - risk: creates an Order grouping multiple shipments; does not buy postage by itself, approval required
- cancel_claim:
  - endpoint: POST /claims/{{ record.id }}/cancel
  - required fields: id
  - risk: cancels an existing EasyPost insurance claim; external claim workflow mutation, approval required
- buy_order:
  - endpoint: POST /orders/{{ record.id }}/buy
  - required fields: id, carrier, service
  - risk: purchases postage for an Order and may incur charges; approval required
- create_pickup:
  - endpoint: POST /pickups
  - required fields: pickup
  - risk: creates a carrier pickup request for a shipment/address window; approval required
- buy_pickup:
  - endpoint: POST /pickups/{{ record.id }}/buy
  - required fields: id, carrier, service
  - risk: buys/schedules a carrier pickup and may incur carrier charges; approval required
- cancel_pickup:
  - endpoint: POST /pickups/{{ record.id }}/cancel
  - required fields: id
  - risk: cancels a scheduled Pickup; external operational mutation, approval required
- create_refund:
  - endpoint: POST /refunds
  - required fields: refund
  - risk: creates one or more shipment refund requests; approval required
- create_scan_form:
  - endpoint: POST /scan_forms
  - required fields: shipments
  - risk: creates a ScanForm manifest for shipment IDs; approval required
- create_report:
  - endpoint: POST /reports/{{ record.type }}
  - required fields: type, start_date, end_date
  - risk: starts an asynchronous EasyPost report export for the requested report type/date range; approval required
- create_luma_promise:
  - endpoint: POST /luma/promise
  - required fields: shipment
  - risk: requests a Luma delivery promise/rating calculation; no label purchase by itself, approval required
- create_luma_shipment:
  - endpoint: POST /shipments/luma
  - required fields: shipment
  - risk: creates and buys a Shipment through Luma one-call buy and may incur postage charges; approval required
- buy_luma_shipment:
  - endpoint: POST /shipments/{{ record.id }}/luma
  - required fields: id, ruleset_name
  - risk: buys postage for an existing Shipment through Luma and may incur charges; approval required
- create_child_user:
  - endpoint: POST /users
  - required fields: user
  - risk: creates a production-only child user/sub-account under the authenticated EasyPost account; elevated account-management mutation, approval required
- update_user:
  - endpoint: PATCH /users/{{ record.id }}
  - required fields: id, user
  - risk: updates EasyPost user/sub-account settings such as child account name; elevated account-management mutation, approval required
- delete_child_user:
  - endpoint: DELETE /users/{{ record.child_id }}
  - required fields: child_id
  - risk: removes a child user from the parent account; destructive account-management mutation, approval required
- create_webhook:
  - endpoint: POST /webhooks
  - required fields: webhook
  - risk: registers an outbound Webhook URL/custom headers for EasyPost events; approval required
- update_webhook:
  - endpoint: PATCH /webhooks/{{ record.id }}
  - required fields: id
  - risk: updates Webhook delivery metadata such as custom headers; approval required
- delete_webhook:
  - endpoint: DELETE /webhooks/{{ record.id }}
  - required fields: id
  - risk: deletes a Webhook subscription, stopping outbound event delivery; destructive external mutation, approval required

## Security

- read risk: external EasyPost API read of shipping, tracking, carrier metadata/account, event, claim, pickup, refund, scan-form, user, referral customer, and webhook data
- write risk: external EasyPost API mutation of live shipping objects, labels, pickups, insurance/refund workflows, reports, users, and webhooks; buy/refund/insurance/pickup/order/Luma actions may incur charges or alter operational shipping state
- approval: reverse ETL plan preview and approval required for every write action; charge-bearing/destructive actions are flagged in writes.json
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect easypost
```

### Inspect as structured JSON

```bash
pm connectors inspect easypost --json
```

## Agent Rules

- Run pm connectors inspect easypost before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
