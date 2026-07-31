# Chargebee connector

## Overview

Fixture-backed Chargebee API v2 Product Catalog 2.0 bundle generated from the pinned official OpenAPI source.

- Source: `https://raw.githubusercontent.com/chargebee/openapi/fbd261f5383317cdc98d00d448ba038cc0659df1/spec/chargebee_api_v2_pc_v2_spec.json`
- Spec version: `2026-07-21.2a6a65b3e1a8ff29840466a7bfdb5cdd778d0634`
- Reviewed at: `2026-07-31`
- Operation ledger: 655 official operations total: 125 read streams, 264 reverse-ETL writes, 18 direct/query reads blocked/planned, 14 binary/file operations blocked/planned, and 234 CDC/webhook/changefeed operations blocked/planned.
- Certification: fixture-only and uncertified; this slice performs no live Chargebee calls.

## Auth setup

Provide `site_api_key` as a secret and `base_url` as the fully formed Chargebee API v2 base URL, for example `https://example.chargebee.com/api/v2`. The API key is sent as the HTTP Basic username with an empty password and must not be logged or stored in prompts.

## Streams notes

All executable streams have sanitized replay fixtures and schemas. Parameterized detail/list streams use synthetic non-secret path placeholders in fixture conformance.

## Write actions & risks

Chargebee writes use closed `record_schema` definitions and remain under the reverse-ETL plan -> preview -> explicit approval -> execute contract. Destructive delete-style actions include typed confirmation metadata and provider-missing `404` idempotency where applicable.

## Known limits

`api_surface.json` records every official operation. Direct/query, binary/file, and CDC/webhook operations are included as blocked/planned ledger entries rather than executable raw passthroughs because this branch does not add shared runtime foundations, file writes, webhook receivers, or live certification.
