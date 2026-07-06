# Overview

Reads Northpass LMS people, courses, course enrollments, and groups through the Northpass REST API.
Read-only.

Readable streams: `people`, `courses`, `course_enrollments`, `groups`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.northpass.com/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Northpass API key, sent as the X-Api-Key header. Never
  logged.
- `base_url` (optional, string); default `https://api.northpass.com/v2`; format `uri`; Northpass API
  base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.northpass.com/v2`.

Authentication behavior:

- API key authentication in `X-Api-Key` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/courses`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `links.next`; next URLs
stay on the configured API host.

- `people`: GET `/people` - records path `data`; query `limit`=`100`; follows a next-page URL from
  the response body; URL path `links.next`; next URLs stay on the configured API host; computed
  output fields `created_at`, `email`, `first_name`, `last_name`, `status`, `updated_at`.
- `courses`: GET `/courses` - records path `data`; query `limit`=`100`; follows a next-page URL from
  the response body; URL path `links.next`; next URLs stay on the configured API host; computed
  output fields `created_at`, `name`, `slug`, `status`, `updated_at`.
- `course_enrollments`: GET `/course_enrollments` - records path `data`; query `limit`=`100`;
  follows a next-page URL from the response body; URL path `links.next`; next URLs stay on the
  configured API host; computed output fields `completed_at`, `course_id`, `created_at`,
  `learner_id`, `percentage`, `status`, `updated_at`.
- `groups`: GET `/groups` - records path `data`; query `limit`=`100`; follows a next-page URL from
  the response body; URL path `links.next`; next URLs stay on the configured API host; computed
  output fields `created_at`, `name`, `slug`, `updated_at`.

## Write actions & risks

This connector is read-only. Read behavior: external Northpass LMS API read of learner and course
data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=3.
