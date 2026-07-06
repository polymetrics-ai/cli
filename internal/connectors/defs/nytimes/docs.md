# Overview

Reads New York Times Most Popular (viewed, emailed, shared) articles via the NYTimes Developer APIs.

Readable streams: `most_popular_viewed`, `most_popular_emailed`, `most_popular_shared`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.nytimes.com/apis.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); NYTimes Developer API key. Sent as the api-key query
  parameter; never logged.
- `base_url` (optional, string); default `https://api.nytimes.com/svc`; format `uri`; NYTimes API
  base URL override for tests or proxies.
- `mode` (optional, string).
- `period` (optional, string); default `7`; Most Popular period in days: 1, 7, or 30.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.nytimes.com/svc`, `period=7`.

Authentication behavior:

- API key authentication in query parameter `api-key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/mostpopular/v2/viewed/{{ config.period }}.json`.

## Streams notes

Default pagination: single request; no pagination.

- `most_popular_viewed`: GET `/mostpopular/v2/viewed/{{ config.period }}.json` - records path
  `results`.
- `most_popular_emailed`: GET `/mostpopular/v2/emailed/{{ config.period }}.json` - records path
  `results`.
- `most_popular_shared`: GET `/mostpopular/v2/shared/{{ config.period }}.json` - records path
  `results`.

## Write actions & risks

This connector is read-only. Read behavior: external NYTimes API read of published article metadata
(no PII).

## Known limits

- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=5.
