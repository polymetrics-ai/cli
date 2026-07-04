# Overview

Workable is a declarative HTTP connector for the official Workable SPI v3 API. The Pass B bundle covers the documented authenticated recruiting, account, employee, time tracking, time off, review, subscription, requisition, and offer GET surfaces as streams, and the JSON-expressible mutation endpoints as reverse-ETL write actions.

## Auth setup

Set `base_url` to the full SPI v3 base URL, such as `https://example.workable.com/spi/v3`, and provide the Workable API access token in the `api_key` secret. Requests use Bearer authentication, matching the legacy connector. The legacy `account_subdomain`-derived base URL shorthand is still not modeled because this dialect cannot derive one config value from another; pass the full `base_url` directly.

## Streams notes

The legacy parity streams `jobs`, `candidates`, and `members` keep Workable's `paging.next` pagination and send `limit=100`; `start_date`, when configured, is sent as `created_after` to preserve the legacy request shape. Pass B adds account, department, permission, recruiter, subscription, employee, review, time tracking, time off, candidate sub-resource, event, job sub-resource, requisition, and offer streams. Detail and sub-resource streams use optional config fields such as `candidate_id`, `job_shortcode`, `employee_id`, `event_id`, `offer_id`, and `requisition_code`.

Most streams use `projection: passthrough` because the legacy Workable read path emitted raw API objects. Schemas declare stable catalog keys and common documented fields but intentionally allow additional properties so new Workable response fields continue to pass through. Object-valued configuration endpoints that do not expose per-record IDs are emitted as singleton records with a computed `_pm_id`.

## Write actions & risks

Write actions cover departments, members, subscriptions, employees, review templates, time entries, time off approvals/requests, candidates, offers, requisitions, and talent pool candidates. These actions can create, update, approve, reject, archive, deactivate, or delete Workable resources. Destructive or deactivation-style actions are marked destructive where appropriate and still run through the plan, preview, approval, execute flow.

## Known limits

- The public careers API endpoint `/api/accounts/{subdomain}` is excluded as a duplicate of authenticated job data because it is served from a different public API host and does not share the configured SPI v3 base/auth model.
- `POST /employees/{id}/documents` is excluded because it is a multipart document upload, which the declarative JSON write dialect cannot express.
- Detail and sub-resource streams require the relevant id/shortcode config value. This bundle does not fan out every detail endpoint from parent lists because the legacy connector did not do that and no shared hook should be added for this Pass B expansion.
- Runtime-configurable `page_size` and `max_pages` remain unsupported for Workable for the same dialect reason documented in the prior bundle: pagination fields are static bundle values, not templates.
