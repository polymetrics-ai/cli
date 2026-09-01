# Overview

Xero (`xero`) reads and writes the provider-owned Xero Accounting API using the official XeroAPI/Xero-OpenAPI `xero_accounting.yaml` document. This connector intentionally scopes to Accounting API operations only; Xero assets, bankfeeds, files, finance, payroll, projects, webhooks, identity, and app-store OpenAPI products are not counted here.

Current fixture-backed operation ledger:

| Lane | Official operations | Connector disposition |
| --- | ---: | --- |
| ETL/read streams | 78 | 78 covered by typed streams and sanitized stream fixtures |
| Direct/provider report reads | 11 | 11 covered by bounded `rest_read` operations and provider-style CLI commands |
| Reverse ETL writes | 87 | 87 covered by typed write actions and sanitized write fixtures |
| Binary/file attachment and PDF operations | 59 | 11 attachment metadata list streams implemented; 48 binary/PDF download or attachment upload operations are bounded in `operations.json` and blocked on the shared binary/file executor |
| CDC/changefeed | 0 | No Accounting API CDC/changefeed operation is present in the official source |

No live Xero call, credentialed check, validation, or provider write was performed to produce these fixtures.

## Auth setup

Create a Xero OAuth2 credential outside prompt text. The connector expects:

- `access_token` (`x-secret: true`): OAuth2 bearer access token.
- `tenant_id` (`x-secret: true`): Xero tenant/organisation id sent as the `Xero-tenant-id` header.
- `base_url` (optional): defaults to `https://api.xero.com/api.xro/2.0`.

Add credentials from environment variables or stdin, not chat or shell history. Secret fields are redacted in logs, direct-read output, and write previews.

## Streams notes

The official Accounting API has 78 non-report, non-binary GET operations and all are represented in `streams.json`. The bundle also retains report and attachment metadata streams for ETL compatibility, but `execution bundle` classifies official report endpoints as direct/provider reads and attachment/PDF endpoints as binary/file lane operations.

Pagination is bounded by the declarative `page_number` paginator for Xero list endpoints. Generated execution-contract fixtures include at least one sanitized page for every stream; the first eligible paginated stream includes a second page to exercise pagination termination without inflating every embedded fixture.

## Write actions & risks

All 87 non-attachment Accounting API mutations are named reverse-ETL write actions in `writes.json`. Write actions use closed record schemas, deterministic request construction, redaction for path identifiers, fixture-backed request-shape coverage, and the CLI reverse-ETL lifecycle: plan → preview → explicit approval → execute.

Destructive delete/status-delete actions declare `confirm: "destructive"`. Xero exposes several delete operations as `POST` requests that set `Status: DELETED`; where the provider exposes idempotent HTTP delete semantics the action records `missing_ok_status: [404]`. For provider status-delete calls without a documented idempotency key, operators must de-duplicate records by the Xero identifier before approval.

Attachment uploads are not exposed as generic write actions. They are tracked as bounded file-upload operations in `operations.json` and blocked until the shared binary/file runner supports approved payload digests and provider-bound paths.

## Known limits

- Fixture-only evidence is not live validation. the execution bundle records safe defaults and one direct-read candidate, but this branch does not claim validated live Xero behavior.
- The 48 attachment/PDF download or attachment upload operations that transfer file bytes are bounded and source-linked, but blocked on shared runtime support for binary/file execution. No connector-local generic HTTP/file escape hatch was added.
- Provider report commands are bounded direct reads with fixed operation definitions. They do not alter warehouse `pm query` semantics and do not expose raw query/path/body passthrough.
- Official source re-audit used Xero Accounting OpenAPI `info.version=16.1.0` fetched from `https://raw.githubusercontent.com/XeroAPI/Xero-OpenAPI/master/xero_accounting.yaml` on 2026-07-31. The semantic binary count treats PDF endpoints as binary/file operations even though the landed r2 issue table grouped only attachment paths under binary_file.
