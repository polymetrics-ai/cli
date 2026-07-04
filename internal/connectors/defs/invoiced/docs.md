# Overview

Invoiced is a billing API for customers, invoices, payments, subscriptions, estimates, and related accounting resources. This Pass B bundle expands the legacy five-stream source into the documented REST API surface published at https://developer.invoiced.com/api. The legacy stream names and projected fields for `customers`, `invoices`, `payments`, `subscriptions`, and `estimates` remain first in the catalog and unchanged.

## Auth setup

Provide an Invoiced API key via the `api_key` secret. The API key is sent as the HTTP Basic username with a blank password, matching the legacy connector. `base_url` defaults to `https://api.invoiced.com`; use `https://api.sandbox.invoiced.com` for sandbox accounts.

## Streams notes

List endpoints read top-level JSON arrays with page-number pagination using `page` and `per_page`, starting at page 1 with a default page size of 100. Detail and singleton endpoints read top-level JSON objects with pagination disabled. New Pass B streams use passthrough projection with permissive schemas so the engine preserves the documented response object without narrowing fields that legacy never modeled.

The original five legacy streams keep their schema projection and client-filtered `updated_at` cursor behavior because the legacy connector does not send a server-side updated-at filter to Invoiced.

## Write actions & risks

`writes.json` covers all JSON/form-expressible documented POST, PUT, PATCH, and DELETE endpoints. Create and update actions send JSON request bodies using all non-path fields from the write record. Delete, void, cancel, charge, pay, and refund actions are marked as high-impact or destructive and require the normal reverse-ETL approval flow; DELETE actions are bodyless and treat 404 as idempotent missing-ok.

## Known limits

- `POST /files` is excluded as `binary_payload` because the official API reference requires multipart file upload with a `file` part, which is outside the declarative JSON/form write dialect.
- The API reference describes generic query helpers such as filtering, metadata filtering, expansion, and PDF variants; these are request modifiers, not separate endpoints, and are not represented as additional streams.
