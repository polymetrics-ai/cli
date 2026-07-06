# Overview

Reads Survicate surveys, survey questions, responses, and respondent attributes, and manages GDPR
personal-data requests, through the Survicate Data Export API v2. Read-only.

Readable streams: `surveys`, `survey_questions`, `responses`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.survicate.com/data-export/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Survicate Data Export API key (Survicate panel > Surveys
  Settings > Access Keys), sent verbatim as 'Authorization: Basic <api_key>' per Survicate's
  documented scheme (docs.survicate.com/data-export/setup: 'The format for this should be Basic
  {{apiKey}}' -- a raw key, NOT base64-encoded user:pass HTTP Basic auth). Never logged.
- `base_url` (optional, string); default `https://data-api.survicate.com/v2`; format `uri`;
  Survicate Data Export API base URL override for tests or proxies. Current documented API version
  is v2.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://data-api.survicate.com/v2`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `Basic` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/surveys` with query `items_per_page`=`1`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path
`pagination_data.next_url`; cross-host next URLs are allowed.

- `surveys`: GET `/surveys` - records path `data`; query `items_per_page`=`100`; follows a next-page
  URL from the response body; URL path `pagination_data.next_url`; cross-host next URLs are allowed.
- `survey_questions`: GET `/surveys/{{ fanout.id }}/questions` - records path `data`; query
  `items_per_page`=`100`; follows a next-page URL from the response body; URL path
  `pagination_data.next_url`; cross-host next URLs are allowed; fan-out; ids from request
  `/surveys`; id-list records path `data`; id field `id`; id inserted into the request path; stamps
  `survey_id`.
- `responses`: GET `/surveys/{{ fanout.id }}/responses` - records path `data`; query
  `items_per_page`=`100`; follows a next-page URL from the response body; URL path
  `pagination_data.next_url`; cross-host next URLs are allowed; fan-out; ids from request
  `/surveys`; id-list records path `data`; id field `id`; id inserted into the request path; stamps
  `survey_id`.

## Write actions & risks

This connector is read-only. Read behavior: external Survicate API read of survey, response, and
respondent data.

## Known limits

- Published rate limit metadata: requests_per_minute=1000.
- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, duplicate_of=2, requires_elevated_scope=3.
