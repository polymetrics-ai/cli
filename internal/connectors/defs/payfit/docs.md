# Overview

PayFit reads the legacy customer `/v1` endpoints used by `internal/connectors/payfit` and the current company-scoped PayFit API documented at `developers.payfit.io`. The legacy streams (`employees`, `contracts`, `companies`) keep their schema projection, field names, and emitted records; new Pass B streams use passthrough projection for PayFit's documented JSON response objects.

## Auth setup

Provide a PayFit API key via the `api_key` secret. It is sent as `Authorization: Bearer <api_key>` and is never logged. Current company-scoped endpoints require `company_id`; PayFit documents retrieving that value via `https://oauth.payfit.com/introspect`, which is listed in `api_surface.json` as a non-data helper rather than modeled as a stream or write.

## Streams notes

There are 21 streams: 3 legacy `/v1` streams plus current company, collaborator, contract, absence, payroll-status, accounting-v2, meal-voucher, payslip-metadata, and insurance/provident-fund metadata streams. Legacy streams continue using PayFit's old `limit` + `offset`/`meta.next_offset` pagination. Current paginated endpoints use `maxResults=50`, `nextPageToken`, and `meta.nextPageToken`. Endpoints returning PDF, CSV, or octet-stream files are excluded as `binary_payload` and the JSON metadata endpoints that link to those files are streamed instead.

## Write actions & risks

There are 7 write actions for documented JSON customer-key mutations: create collaborator, create contract, create absence, cancel absence, update contract health insurance, update contract provident fund, and request health-insurance regularization. Multipart setup-sheet upload is excluded as `binary_payload`; partner-only billing declarations are excluded as `requires_elevated_scope` because this migrated connector authenticates with a customer API key.

## Known limits

- Legacy `/v1` streams are retained for emitted-record parity even though current PayFit documentation now centers on `/companies/{companyId}/...` endpoints.
- Optional filters such as collaborator email, absence status/date range, and include-in-progress contracts are not modeled because the read dialect lacks optional per-stream query omission for arbitrary config keys. Required date/pay-period parameters are modeled with `pay_period`.
- Binary download endpoints for payslips, company documents, payment files, and CSV accounting exports are excluded; their JSON list/metadata endpoints are covered where documented.
- The OAuth introspection and webhook-dashboard helper URLs are documented operational endpoints, not syncable data resources; they are listed as `non_data_endpoint` exclusions.
