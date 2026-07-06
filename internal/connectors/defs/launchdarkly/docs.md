# Overview

Reads LaunchDarkly projects, members, audit log entries, feature flags, and environments through the
LaunchDarkly REST API.

Readable streams: `projects`, `members`, `auditlog`, `flags`, `environments`.

This connector is read-only; no write actions are declared.

Service API documentation: https://apidocs.launchdarkly.com/.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); LaunchDarkly access token. Sent verbatim (no "Bearer"
  prefix) as the Authorization header; never logged.
- `base_url` (optional, string); default `https://app.launchdarkly.com/api/v2`; format `uri`;
  LaunchDarkly API base URL override for tests or proxies.
- `mode` (optional, string).
- `project_key` (optional, string); LaunchDarkly project key. Required only for the flags and
  environments streams, which are scoped to a single project.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://app.launchdarkly.com/api/v2`.

Authentication behavior:

- API key authentication in `Authorization` using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/projects` with query `limit`=`1`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 20.

- `projects`: GET `/projects` - records path `items`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 20.
- `members`: GET `/members` - records path `items`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 20.
- `auditlog`: GET `/auditlog` - records path `items`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 20.
- `flags`: GET `/flags/{{ config.project_key }}` - records path `items`; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 20.
- `environments`: GET `/projects/{{ config.project_key }}/environments` - records path `items`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 20.

## Write actions & risks

This connector is read-only. Read behavior: external LaunchDarkly API read of project, membership,
audit, and feature-flag configuration data.

## Known limits

- Batch defaults: read_page_size=20.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=2.
