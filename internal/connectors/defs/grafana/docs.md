# Overview

Reads Grafana dashboards, folders, data sources, organization users, and provisioned alert rules
through the Grafana REST API (read-only).

Readable streams: `dashboards`, `folders`, `datasources`, `org_users`, `alert_rules`.

This connector is read-only; no write actions are declared.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Grafana service account token or API key, sent as
  Authorization: Bearer <api_key>; never logged.
- `base_url` (required, string); format `uri`; Your Grafana instance's URL, e.g.
  https://your-grafana.grafana.net.
- `mode` (optional, string).

Secret fields are redacted in logs and write previews: `api_key`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/org`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `limit`; starts at
1; page size 1000.

Pagination by stream: none: `datasources`, `org_users`, `alert_rules`; page_number: `dashboards`,
`folders`.

- `dashboards`: GET `/api/search` - records at response root; query `type`=`dash-db`; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 1000.
- `folders`: GET `/api/search` - records at response root; query `type`=`dash-folder`; page-number
  pagination; page parameter `page`; size parameter `limit`; starts at 1; page size 1000.
- `datasources`: GET `/api/datasources` - records at response root.
- `org_users`: GET `/api/org/users` - records at response root.
- `alert_rules`: GET `/api/v1/provisioning/alert-rules` - records at response root.

## Write actions & risks

This connector is read-only. Read behavior: external Grafana instance API read of dashboards,
folders, data sources, org users, and alert rules.

## Known limits

- Batch defaults: read_page_size=1000.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=6.
