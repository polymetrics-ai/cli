# Overview

Flexport reads the official default Flexport API v2 REST surface linked from
`https://developers.flexport.com/` to `https://apidocs.flexport.com/`. The bundle covers paginated
JSON lists for bookings, purchase orders, shipment status resources, customs and freight invoices,
documents, ports, network objects, products, and webhook events. It also covers the v2 JSON create
or update actions that the declarative write engine can express as one request per record.

## Auth setup

Provide a Flexport API credential in the `api_key` secret. The engine sends it as
`Authorization: Bearer <api_key>`, matching the legacy connector's bearer-token behavior.
`base_url` defaults to `https://api.flexport.com`. `page_size` defaults to `100` and is sent as the
`per` query parameter on list requests.

## Streams notes

List streams use Flexport's paginated envelope: records are under `data.data` and the next page URL
is under `data.next`. `my_company` is the only single-object stream; it reads
`GET /network/me/companies` and extracts `data`.

The five legacy-backed streams (`companies`, `locations`, `products`, `invoices`, and `shipments`)
use schema projection matching the legacy Go mappers. In particular, `companies` and `locations`
read the legacy `/companies` and `/locations` list paths rather than the newer `/network/*` list
objects because the newer network resources expose a different shape. The Pass B-added streams use
`projection: "passthrough"` with permissive schemas keyed by `id` to preserve their broad v2
objects.

## Write actions & risks

Write actions are JSON-only, one request per input record:

- `create_booking_amendment`, `create_booking_line_item`, `create_booking`
- `create_document`
- `create_company`, `update_company`
- `create_company_entity`, `update_company_entity`
- `create_contact`, `update_contact`
- `create_location`, `update_location`
- `create_product`, `update_product`
- `update_shipment`
- `create_shipments_shareable`

The highest-risk writes are `create_booking`, `create_booking_amendment`, `create_document`, and
`update_shipment` because they can affect operational freight records. All reverse ETL writes still
flow through plan, preview, approval, and execute.

## Known limits

- The default official docs page is Flexport API v2. The `2023-07-01`, `v3`, and `v1` variants in
  the same Redocly state are not enumerated in this bundle.
- Detail GET endpoints that require a caller-supplied id are excluded as `duplicate_of` when the
  corresponding list stream covers the same resource family. `GET /container_load_results/{id}` is
  excluded as `out_of_scope` because the v2 docs expose no list endpoint to discover ids.
- `GET /documents/{id}/download` is excluded as `binary_payload`; it returns document bytes, not a
  JSON record stream.
- Optional filter parameters such as `f.updated_at.gt`, `f.shipment.id`, metadata filters, and
  sort/direction are not surfaced as connector config. The streams read the unfiltered list surface
  with `per={{ config.page_size }}`.
- The fixture files assert method and path but intentionally do not assert `per`: conformance
  fills every non-secret config key with `synthetic-conformance-value`, so templated page-size query
  values are not stable fixture literals.
