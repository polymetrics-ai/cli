# Overview

Reads JobNimbus CRM contacts, jobs, tasks, activities, and files through the JobNimbus REST API.

Readable streams: `contacts`, `jobs`, `tasks`, `activities`, `files`.

This connector is read-only; no write actions are declared.

Service API documentation: https://documenter.getpostman.com/view/3919598/S11PpG4x.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); JobNimbus API key, sent as a Bearer token (Authorization:
  Bearer <api_key>). Never logged.
- `base_url` (optional, string); default `https://app.jobnimbus.com/api1`; format `uri`; JobNimbus
  API base URL override for tests or proxies.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://app.jobnimbus.com/api1`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/contacts` with query `from`=`0`; `size`=`1`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `from`; limit parameter `size`; page
size 1000.

- `contacts`: GET `/contacts` - records path `results`; offset/limit pagination; offset parameter
  `from`; limit parameter `size`; page size 1000.
- `jobs`: GET `/jobs` - records path `results`; offset/limit pagination; offset parameter `from`;
  limit parameter `size`; page size 1000.
- `tasks`: GET `/tasks` - records path `results`; offset/limit pagination; offset parameter `from`;
  limit parameter `size`; page size 1000.
- `activities`: GET `/activities` - records path `activity`; offset/limit pagination; offset
  parameter `from`; limit parameter `size`; page size 1000.
- `files`: GET `/files` - records path `files`; offset/limit pagination; offset parameter `from`;
  limit parameter `size`; page size 1000.

## Write actions & risks

This connector is read-only. Read behavior: external JobNimbus API read of CRM contact, job, task,
activity, and file data.

## Known limits

- Batch defaults: read_page_size=1000.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=4.
