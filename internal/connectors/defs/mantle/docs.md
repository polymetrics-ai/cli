# Overview

Mantle is a billing, CRM, helpdesk, documentation, email, workflow, and analytics API for Shopify apps. This Pass B bundle covers the Mantle Core API documented at https://coreapi.heymantle.dev/ while preserving legacy parity for the original `customers` and `subscriptions` streams.

## Auth setup

Provide a Mantle API key via the `api_key` secret. The bundle sends it with Bearer authentication and does not place it in record data, fixtures, or logs.

## Streams notes

The bundle declares 133 GET streams. The legacy `customers` and `subscriptions` streams retain the legacy record projection and `take=500` request shape. Newly added streams use passthrough projection so documented response fields are not dropped by a narrow generated schema.

Streams with Mantle cursor metadata declare per-stream cursor pagination using `cursor` or `nextCursor` plus `hasNextPage`. Other detail or singleton endpoints explicitly use non-paginated reads so the base cursor paginator is not accidentally applied to one-record responses.

Streams whose documented path or required query needs an ID use optional config properties such as `id`, `app_id`, `page_id`, `resource_id`, and `repository_id`. They are optional at connection-spec level because they are stream-specific; reading the corresponding stream requires setting the relevant value.

## Write actions & risks

The bundle declares 148 dialect-expressible write actions for documented POST, PUT, PATCH, and DELETE operations. Delete actions are marked destructive and treat 404 as an idempotent missing-ok response when Mantle reports a resource is already gone. Other actions still require the normal write approval path because they mutate Mantle resources or trigger side effects such as email delivery, AI generation, webhooks, workflow runs, uploads, or transcription jobs.

## Known limits

- Deprecated Mantle docs endpoints are excluded with category `deprecated` and point to their documented successors where applicable.
- Webhook event callback payload pages under `/webhooks/<topic>` are excluded as `non_data_endpoint`; they document requests Mantle sends to receivers, not client-callable Mantle API mutations.
- `POST /v1/attachments` creates a presigned upload URL and is modeled as a write action for that API call only; uploading file bytes to the returned signed URL is outside the Mantle API connector surface.
- Generated schemas intentionally flatten OpenAPI `allOf` and use broad types for unsupported OpenAPI constructs such as `oneOf`. New streams use passthrough projection so this does not discard documented fields.
