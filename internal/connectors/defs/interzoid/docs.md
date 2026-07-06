# Overview

Reads Interzoid data-matching lookups: company-name, individual-name, and street-address similarity
keys, plus organization-name standardization, via the Interzoid REST API.

Readable streams: `company_name_matching`, `individual_name_matching`, `street_address_matching`,
`standardize_company_names`.

This connector is read-only; no write actions are declared.

Service API documentation: https://docs.interzoid.com/entries/interzoid-core-apis.

## Auth setup

Connection fields:

- `address` (optional, string); Street address input for the street_address_matching stream
  (getaddressmatchadvanced). Required to read that stream.
- `address_match_algorithm` (optional, string); Optional matching algorithm override for the
  street_address_matching stream.
- `api_key` (required, secret, string); Interzoid license key. Sent as the `license` query parameter
  on every request; never logged.
- `base_url` (optional, string); default `https://api.interzoid.com`; format `uri`; Interzoid API
  base URL override for tests or proxies.
- `company` (optional, string); Company name input for the company_name_matching stream
  (getcompanymatchadvanced). Required to read that stream.
- `company_match_algorithm` (optional, string); Optional matching algorithm override for the
  company_name_matching stream.
- `fullname` (optional, string); Full name input for the individual_name_matching stream
  (getfullnamematch). Required to read that stream.
- `org` (optional, string); Organization name input for the standardize_company_names stream
  (getorgstandard). Required to read that stream.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.interzoid.com`.

Authentication behavior:

- API key authentication in query parameter `license` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

## Streams notes

Default pagination: single request; no pagination.

- `company_name_matching`: GET `/getcompanymatchadvanced` - single-object response; records at
  response root; query `algorithm` from template `{{ config.company_match_algorithm }}`, omitted
  when absent; `company`=`{{ config.company }}`; computed output fields `query_company`.
- `individual_name_matching`: GET `/getfullnamematch` - single-object response; records at response
  root; query `fullname`=`{{ config.fullname }}`; computed output fields `query_fullname`.
- `street_address_matching`: GET `/getaddressmatchadvanced` - single-object response; records at
  response root; query `address`=`{{ config.address }}`; `algorithm` from template `{{
  config.address_match_algorithm }}`, omitted when absent; computed output fields `query_address`.
- `standardize_company_names`: GET `/getorgstandard` - single-object response; records at response
  root; query `org`=`{{ config.org }}`; computed output fields `query_org`.

## Write actions & risks

This connector is read-only. Read behavior: external Interzoid API single-lookup read; each read
spends an API credit.

## Known limits

- Batch defaults: read_page_size=1.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=3.
