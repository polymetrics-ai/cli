# Zoho Books Connector

## Overview
This bundle targets Zoho Books API v3 at `https://www.zohoapis.com/books/v3`. It preserves the legacy read shape for `contacts`, `invoices`, and `items` and expands Pass B coverage from Zoho's OpenAPI archive (`openapi-all.zip`, dated 2026-06-25) for JSON list/detail operations and declarative JSON/form/no-body mutations.

## Auth setup
Provide a Zoho OAuth access token as `access_token`. The connector sends it in the `Authorization` header with Zoho's documented `Zoho-oauthtoken ` prefix. Do not put token values in command arguments or docs.

## Streams notes
Reads use Zoho's page-number pagination with `page` and `per_page`, matching the legacy 200-record page size. `organization_id` remains optional for read requests to preserve legacy behavior, and path or required-query parameters for detail/sub-resource streams are declared as config fields for conformance and targeted reads. Streams use passthrough projection so Zoho's raw accounting fields are preserved; the legacy streams synthesize `id`, `name`, and `updated_at` with the same first-non-null fallback order as the Go mapper.

## Write actions & risks
Write actions cover OpenAPI POST, PUT, PATCH, and DELETE operations whose request can be expressed as JSON, form, or an empty body. Write bodies are sent verbatim from the input record, so endpoints with documented wrapper objects must include those wrappers in the record. Organization-scoped write actions include `organization_id` in the request URL from config. Deletes are marked destructive and include idempotent 404 handling.

## Known limits
Multipart, file upload/download, PDF/print/export, and binary document operations are excluded as `binary_payload`. Operations that require custom per-action headers, such as custom-field unique-value update variants, are excluded because the current write dialect does not model action-specific headers. Non-record utility responses that do not expose a JSON object or array root are marked `non_data_endpoint` in `api_surface.json`.
