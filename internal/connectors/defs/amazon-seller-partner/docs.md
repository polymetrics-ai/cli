# Overview

Amazon Seller Partner covers the official Amazon Selling Partner API (SP-API) JSON surface from the documentation site and the official `amzn/selling-partner-api-models` Swagger models. The bundle keeps the three legacy-parity streams first and expands Pass B coverage with additional passthrough streams for JSON GET operations plus write actions for mutations that the current declarative engine can express as path/body JSON requests.

The surface manifest covers 353 documented operations. Implemented reads use schema projection only for the legacy streams and `projection: "passthrough"` for generated Pass B streams so the raw documented response objects are not truncated while schemas stabilize. Endpoints that need restricted buyer data, binary documents, dynamic POST read bodies, mutation query parameters, or elevated application administration are excluded in `api_surface.json` with a closed-vocabulary category and a concrete reason.

## Auth setup

SP-API uses Login with Amazon (LWA). The configured `refresh_token`, `lwa_app_id`, and `lwa_client_secret` are exchanged by the existing Tier-2 AuthHook for an access token, which is sent as `x-amz-access-token` on SP-API requests. The declarative engine cannot perform this refresh-token exchange or route the token into that header without the hook, so bundle-level dynamic conformance remains skipped and the parity suite named in `metadata.json` remains the auth proof.

`base_url` defaults to the North America endpoint. Set it to the EU or FE endpoint for those sellers. Operation-specific stream inputs such as `asin`, `seller_sku`, `order_id`, `shipment_id`, `report_id`, `created_after`, and similar values are optional in `spec.json` because they are only required when selecting the corresponding stream.

## Streams notes

Legacy streams:

- `orders` reads `GET /orders/v0/orders` with incremental `LastUpdateDate` filtering.
- `inventory_summaries` reads `GET /fba/inventory/v1/summaries` with incremental `lastUpdatedTime` filtering.
- `financial_event_groups` reads `GET /finances/v0/financialEventGroups` with incremental `FinancialEventGroupEnd` filtering.

Generated Pass B streams are named from the official operation id, with a path-derived suffix where Amazon reuses an operation id across API versions. Required query parameters are wired from config keys and therefore fail closed when a caller selects that stream without the necessary value. Common pagination token shapes such as `nextToken`, `NextToken`, `pageToken`, `paginationToken`, and `nextPageToken` are modeled with cursor pagination when the response model exposes a matching token path.

## Write actions & risks

This connector now declares write capability for SP-API mutations that fit the engine's declarative write path: one request per record, path fields from `record.*`, and either a JSON body or no body. All live writes require the normal reverse-ETL plan, preview, approval, and execute workflow. Cancel, delete, submit, confirm, purchase, generate, process, schedule, and update actions are marked `confirm: "destructive"` because they can change seller/vendor workflow state or create external artifacts.

Mutations with required query parameters are excluded rather than approximated because `writes.json` has no query map. Upload/document/token/secret-rotation operations are also excluded because they need binary transfer, presigned upload handling, or elevated application/security scope.

## Known limits

- Dynamic conformance is skipped at bundle level because the custom LWA AuthHook cannot be exercised with synthetic token endpoint config. Static validation and API-surface checks still run.
- Generated Pass B stream schemas are intentionally permissive with `additionalProperties: true` and passthrough projection. They preserve raw SP-API JSON objects instead of trying to flatten every nested model in the first Pass B expansion.
- POST search/read endpoints are excluded: the current read engine does not send dynamic request bodies, even though it has a method field.
- Binary/document endpoints are excluded unless the operation returns ordinary JSON metadata. Downloading presigned report/feed documents, labels, packing slips, invoices, or bill-of-lading content requires a hook or a future engine feature.
- Mutations that require query parameters are excluded until the write dialect has a query-parameter map.
- Existing SP-API cursor pagination continues to resend base filters alongside cursor tokens, matching the generic engine cursor paginator rather than legacy's SP-API-specific continuation-query replacement.
