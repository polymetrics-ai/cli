# Overview

Employment Hero is a declarative REST connector for the official Employment Hero API reference at https://developer.employmenthero.com/api-references. The bundle uses the official Postman collection base URL (https://api.employmenthero.com/api) and declares versioned /v1 and /v2 paths per endpoint.

Pass B expands beyond the original four legacy streams. It now reads organisations, employees, teams, leave, certifications, cost centres, custom fields, employing entities, forms, goals, kiosk members, payroll-reference resources, rosters, unavailability, work locations/sites/types, and employee-scoped resources such as bank accounts, documents, emergency contacts, employment histories, leave balances, pay details, payslips, timesheet entries, tax declarations, superannuation details, and work eligibility. It also exposes official JSON mutations as reverse-ETL write actions.

## Auth setup

Provide an Employment Hero access token via the `api_key` secret. The engine sends it as `Authorization: Bearer <api_key>` and never records the value in fixtures or previews. OAuth authorize/token endpoints are not implemented as connector actions because they are credential-acquisition flows, not data sync endpoints.

## Streams notes

The connector base URL is `https://api.employmenthero.com/api`; stream paths include `/v1` or `/v2`. For the legacy-migrated list streams, requests use the legacy Go connector's `page_index`/`items_per_page` pagination spelling.

`organisations` is the root discovery stream. Most streams require `organization_id` (the official docs spell the path parameter `organisation_id`). Collection-style employee subresources fan out by listing `/v1/organisations/{organization_id}/employees` and then reading each employee child endpoint. Single-object employee detail endpoints (superannuation, tax declaration, work eligibility) use `employee_ids`, a comma-separated config list, because the current fan-out dialect cannot use one pagination strategy for the ID discovery request and a different no-pagination strategy for the child single-object request.

Form, goal, and team child collection streams fan out from their parent lists. Single-object streams use the corresponding config id (for example `form_id`, `goal_id`, `leave_request_id`, `payslip_id`). Optional filter query parameters from the API reference are not modeled unless required for addressing; streams perform broad full-refresh reads.

## Write actions & risks

Write actions cover the official JSON body mutations: certification create/update/archive/delete, department create/update, employee quick-add/onboarding/update/delete, employee certification update, form and form category/template mutations, goal status updates, kiosk access bulk grant/revoke, leave balance adjustment, leave request creation, position create/update, rostered shift bulk create, timesheet entry creation, and work site create/update.

All writes execute live Employment Hero mutations only after reverse-ETL plan preview and approval. Delete actions are marked destructive and treat 404 as idempotent success where the API resource is already absent. Async creation actions such as employee onboarding and rostered-shift bulk create are write-only submissions here; their polling/status endpoints are excluded as non-data endpoints.

## Known limits

- Multipart or binary endpoints are excluded: certification file uploads, employee certification file uploads, document creation/upload, and payslip PDF generation require binary upload/download semantics that the declarative JSON write engine does not express.
- OAuth authorize/token/refresh endpoints are excluded as non-data credential flows.
- Polling/status endpoints for async onboarding, certification polling, and rostered-shift bulk-create jobs are excluded as non-data endpoints that require ephemeral keys from prior mutations.
- The rostered shift cost endpoint is excluded as an aggregate calculation endpoint rather than a list or object data stream.
- Optional filters from the official collection are intentionally not surfaced as spec fields; this avoids dead config and keeps reads broad/full-refresh unless an ID is required in the path.
- Legacy's `organization_configids` comma-list fallback and runtime `items_per_page`/`max_pages` overrides remain unsupported in this declarative bundle; the engine has fixed bundle-authored pagination settings.
