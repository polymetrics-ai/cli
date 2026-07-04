# Overview

Pendo reads Engage API v1 product analytics objects, visitor/account details, deletion-job status, segments, reports, guides, metadata schemas/dependencies, exclusion lists, servers, and Listen feedback options. The four original streams (visitors, accounts, pages, features) retain legacy projection and the legacy data-envelope/cursor pagination shape. Pass B adds documented JSON GET list/detail streams and safe segment, guide, and feedback write actions.

## Auth setup

Provide a Pendo integration key via the `integration_key` secret. It is sent as `x-pendo-integration-key` and is never logged. `base_url` defaults to `https://app.pendo.io/api/v1`; override it only for tests, regional API hosts, or controlled proxies.

## Streams notes

The legacy streams `visitors`, `accounts`, `pages`, and `features` keep the hand-written connector behavior: records are read from the top-level `data` array and paginated with a body `next` token sent back as the `page` query parameter. Newly added documented endpoints use passthrough projection so the raw Pendo object shape is preserved. Detail streams require the corresponding config value, such as `visitor_id`, `account_id`, `segment_id`, `guide_id`, or `metadata_kind`.

## Write actions & risks

Write actions cover expressible Pendo mutations whose request body is JSON object-shaped or empty: segment export/job creation and membership changes, guide seen/state/segment resets, and Listen feedback create/update/delete. Destructive actions are marked with destructive confirmation semantics where applicable. Bulk GDPR/CCPA erasure, metadata schema deletion, data-sync credential rotation, raw XML/CSV import/export, and event-ingestion writes requiring a different per-action secret are not exposed.

## Known limits

- Base64 visitor/account lookup endpoints require `x-pendo-base64-encoded-params: true` on only those requests. The current stream dialect has global base headers, not per-stream headers, so those detail endpoints are excluded and recorded in `docs/migration/quarantine.json`.
- Pendo aggregation is a POST-body semantic read/report DSL. The generic declarative read path does not transmit `stream.body`, and exposing arbitrary aggregation bodies as writes would create a generic query/write tool surface.
- CSV, XLIFF/XML, and scalar response endpoints are excluded because declarative streams emit JSON object records.
- Metadata bulk updates and metadata field creation use root JSON array bodies; single metadata value updates and privacy opt-outs can use raw scalar bodies. The current write dialect constructs JSON objects from record maps, so these body shapes are not expressible without a hook or engine extension.
