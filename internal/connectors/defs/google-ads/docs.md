# Google Ads connector notes

## Overview

Google Ads is implemented as a declarative preview connector against the public Google Ads API v22 REST discovery document. This wave is fixture-only: it does not make live Google Ads calls, request credentials, execute provider writes, or claim certification.

Public source audit:

- Source: `https://googleads.googleapis.com/$discovery/rest?version=v22`
- API version: `v22`
- Discovery revision: `20260721`
- Raw discovery method count: `163` (`POST=151`, `GET=11`, `DELETE=1`)
- Local operation ledger rows: `164`. The row count is one greater than the raw method count because the single `customers.googleAds.search` method is intentionally represented by two fixed GAQL stream rows: `campaigns` and `ad_groups`.

## Auth setup

Provide `access_token` and `developer_token` through the credentials layer or environment. Optional `login_customer_id` is sent only when present. `customer_id` is required for customer-scoped streams, fixed direct reads, and reverse/write actions. Do not place secret values in plans, docs, fixtures, or command text.

## Streams notes

Implemented streams are `accessible_customers`, `campaigns`, and `ad_groups`. The campaign and ad group streams use fixed connector-owned GAQL statements; the connector does not expose arbitrary GAQL or raw search passthrough.

Direct reads: `34` fixed connector-owned operations with JSON-redacted output and bounded response size.

## Write actions & risks

Reverse/write actions: `104` guarded write actions generated from v22 methods whose path variables are representable by the current connector contract.

- Write actions use typed top-level schemas and closed top-level request objects derived from the public discovery schema.
- Destructive or account-admin actions carry explicit `confirm: destructive` metadata and remain subject to the platform reverse ETL plan -> preview -> approval -> execute lifecycle.
- Secret-like fields are redacted; `access_token` and `developer_token` are never stored in fixtures.
- No generic Google Ads SQL/GAQL shell, generic HTTP write, or raw request passthrough is exposed.

## Known limits

Blocked/planned operations: `23` rows. These are not advertised as executable. Most require reserved-expansion resource-name path variables whose values contain slashes.

Google Ads methods whose REST paths use `{+resourceName}`, `{+name}`, `{+experiment}`, `{+campaignDraft}`, or `{+adGroupAd}` are blocked in `api_surface.json`. These path variables are reserved expansions and may contain slash-separated Google Ads resource names. The current connector-local path interpolation intentionally URL-encodes slashes for safety, so enabling those methods without shared reserved-expansion support would call the wrong URL.
