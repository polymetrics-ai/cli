# Google Ads connector notes

## Overview

Google Ads is implemented as a declarative preview connector against the public Google Ads API v22 REST discovery document. This wave ships sanitized fixture coverage plus executable credential-backed reads, fixed direct reads, and guarded reverse/write actions, but does not claim validation.

Public source audit:

- Source: `https://googleads.googleapis.com/$discovery/rest?version=v22`
- API version: `v22`
- Discovery revision: `20260817`
- Canonical Discovery SHA-256: `c14a489015a3a4664addc58fa429c05b3bce26adc2a519a3a5469d475c18f8f8` (`2243707` canonical bytes; `2937930` source bytes)
- Raw discovery method count: `163` (`POST=151`, `GET=11`, `DELETE=1`)
- Local operation ledger rows: `164`. The row count is one greater than the raw method count because the single `customers.googleAds.search` method is intentionally represented by two fixed GAQL stream rows: `campaigns` and `ad_groups`.

## Auth setup

Provide `access_token` and `developer_token` through the credentials layer or environment. Optional `login_customer_id` is sent only when present. `customer_id` is required for customer-scoped streams, fixed direct reads, and reverse/write actions. Do not place secret values in plans, docs, fixtures, or command text.

## Streams notes

Implemented streams are `accessible_customers`, `campaigns`, and `ad_groups`. The campaign and ad group streams use fixed connector-owned GAQL statements; the connector does not expose arbitrary GAQL or raw search passthrough.

Direct reads: `33` fixed connector-owned operations with JSON-redacted output, bounded response size, and typed CLI body/query fields where a POST body or GET query parameters are required.

## Write actions & risks

Reverse/write actions: `104` guarded write actions whose request schemas are closed and connector-owned.

- Write actions use closed record schemas derived from public discovery fields that can be represented without raw operation objects.
- Destructive or account-admin actions carry explicit `confirm: destructive` metadata and remain subject to the platform reverse ETL plan -> preview -> approval -> execute lifecycle.
- Secret-like fields are redacted; `access_token` and `developer_token` are never stored in fixtures.
- No generic Google Ads SQL/GAQL shell, generic HTTP write, or raw request passthrough is exposed.

## Known limits

Blocked/planned operations: `24` rows. These are not advertised as executable. Reserved-expansion resource-name path variables, open-ended discovery write schemas, raw GAQL query commands, and direct reads with required complex request bodies remain blocked.

Google Ads methods whose REST paths use `{+resourceName}`, `{+name}`, `{+experiment}`, `{+campaignDraft}`, or `{+adGroupAd}` are blocked in `execution bundle`. These path variables are reserved expansions and may contain slash-separated Google Ads resource names. The current connector-local path interpolation intentionally URL-encodes slashes for safety, so enabling those methods without shared reserved-expansion support would call the wrong URL.
