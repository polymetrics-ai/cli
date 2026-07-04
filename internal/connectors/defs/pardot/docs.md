# Overview

Pardot reads and writes Salesforce Marketing Cloud Account Engagement API v5 JSON resources. The legacy streams for prospects, campaigns, and lists keep their record projections exactly: prospects emit id, email, firstName, lastName, createdAt, and updatedAt; campaigns and lists emit id, name, createdAt, and updatedAt.

## Auth setup

Provide a pre-issued Salesforce OAuth access token as the `access_token` secret. The connector sends it as `Authorization: Bearer <token>` and sends `business_unit_id` as the `Pardot-Business-Unit-Id` header on every request. Token acquisition and refresh remain outside the connector, matching legacy behavior.

## Streams notes

Query streams use Account Engagement API v5 list conventions: records are read from `values`, requests send `limit=200`, and pagination follows `nextPageUrl`. Detail and singleton streams emit the root response object. The shared optional `id` config value is used by all detail streams in this static bundle.

## Write actions & risks

JSON POST, PATCH, and DELETE operations are represented in `writes.json`. Delete actions are marked destructive and idempotent for 404 responses. High-impact actions such as sending emails, copying assets to CMS, restoring records, canceling imports, assigning visitors, and tag mutations require the normal reverse-ETL plan, preview, approval, and execute flow.

## Known limits

- Multipart/file upload endpoints are excluded as `binary_payload`: file creation, engagement-studio program creation, import batch upload, and bulk-action CSV creation require request shapes outside the declarative JSON write dialect.
- CSV/file download endpoints are excluded as `binary_payload`: import errors, bulk-action errors, and engagement-studio program structure downloads are not JSON list/detail responses.
- The API v5 docs require explicit `fields` parameters, so streams request stable documented field sets instead of asking for every relationship expansion.
