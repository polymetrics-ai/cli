# Overview

Zoho Expense is a connector-architecture-v2 declarative bundle for the official Zoho Expense API v1 surface at `https://www.zohoapis.com/expense/v1`. It preserves the original legacy streams for reports, expenses, and users, and expands the bundle to cover the documented JSON resources for currencies, contacts/customers, employee advances, expense categories, expense reports, expenses, organizations, projects, reporting tags, taxes, trips, and users.

## Auth setup

Provide a Zoho OAuth access token via the `access_token` secret. The bundle sends it as the `Authorization` header with Zoho's `Zoho-oauthtoken ` prefix. Provide `organization_id` for organization-scoped endpoints; it is sent as the documented `X-com-zoho-expense-organizationid` header and, for reporting tag endpoints whose OpenAPI file still declares it as a query parameter, as `organization_id`. The value is configuration, not a secret, and no token values are stored in fixtures or docs.

## Streams notes

Every documented JSON `GET` list/detail endpoint in the API v1 OpenAPI files is represented as a stream. The original legacy streams remain first and keep their public names: `reports` maps `/expensereports` and extracts `expense_reports`, `expenses` maps `/expenses`, and `users` maps `/users`. New detail streams use config path parameters such as `expense_report_id`, `project_id`, `tag_id`, and `tax_id`; list streams with documented `page`/`per_page` parameters use the shared page-number pagination. Streams use `projection: passthrough` so Zoho response fields are retained.

Receipt and attachment retrieval endpoints are not modeled as streams because they transfer files rather than JSON records.

## Write actions & risks

`writes.json` covers every documented dialect-expressible JSON/no-body mutation endpoint. Create/update/status/approval/reimbursement/archive/share/assignment actions send JSON bodies when the docs declare `application/json`, and send no body for documented body-less actions. Path and required query identifiers are record fields, while `organization_id` stays in config. Delete actions are marked destructive; single-resource deletes include idempotent missing-404 handling.

These writes can create, modify, submit, approve, reject, reimburse, archive, export, and delete Zoho Expense resources. They require the normal reverse ETL plan, preview, approval, execute flow.

## Known limits

- Receipt and attachment upload/download endpoints are excluded as `binary_payload`; the current declarative write dialect does not send multipart file bodies, and binary responses are not JSON record streams.
- Zoho's reporting tag OpenAPI file contains regex path placeholders for option/tag IDs. The bundle normalizes those placeholders to `tag_id` and `option_id` path variables while preserving the documented endpoint coverage in `api_surface.json`.
- Legacy runtime `page_size` and `max_pages` overrides are not declared because the current declarative pagination block cannot apply config-driven page sizing or page caps; it uses the default 200 records per page.
