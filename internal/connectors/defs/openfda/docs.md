# Overview

openFDA is a public, read-only REST API over FDA regulatory datasets. This bundle keeps the five legacy-parity streams (`drug_event`, `drug_label`, `drug_enforcement`, `device_event`, and `food_enforcement`) with their exact projected record fields and adds the rest of the currently documented openFDA JSON dataset endpoints: animal/veterinary adverse events, cosmetics adverse events, food adverse events, drug NDC/Drugs@FDA/shortages/Orange Book, medical-device 510(k)/PMA/UDI/enforcement/recall/classification/registration-listing/COVID-19 serology, tobacco problem/research datasets, transparency complete-response letters, and other historical-document/NSDE/substance/UNII datasets.

## Auth setup

An API key is optional. When the `api_key` secret is set, it is sent as the `api_key` query parameter to raise rate limits; otherwise requests use the public unauthenticated tier.

## Streams notes

All streams use openFDA's shared response envelope, `{"meta": {"results": ...}, "results": [...] }`, with offset pagination via `limit` and `skip`. The base page size remains `100`, matching the legacy connector default. Every stream accepts the optional `search` config value through openFDA's standard search query parameter.

The original five legacy streams retain schema projection to preserve exactly the fields emitted by `internal/connectors/openfda`. Newly added streams use passthrough projection with permissive schemas because their record shapes are dataset-specific and already exposed as public openFDA JSON objects.

## Write actions & risks

None. openFDA is a read-only public regulatory API; `capabilities.write` is `false` and no `writes.json` is shipped.

## Known limits

- The declarative offset paginator does not encode legacy's hard `skip <= 25000` guard. openFDA itself enforces the skip ceiling, while normal reads still stop on short pages.
- Download archives listed under openFDA's downloads pages are not separate API streams here; this bundle covers the JSON API endpoints under `https://api.fda.gov/*.json`.
