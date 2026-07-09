# Overview

HubSpot is being added through the issue #132 CLI parity roadmap. This bundle is the issue #134 metadata scaffold: it records the connection spec and provider-like command surface, but it does not yet claim executable streams, direct reads, binary transfers, or reverse ETL writes.

Official operation inventory comes from the public HubSpot OpenAPI collection at <https://github.com/HubSpot/HubSpot-public-api-spec-collection>. The parent baseline is 3,060 unique method/path operations across 401 OpenAPI files. Issue #137 owns the full `api_surface.json` ledger; later lanes implement streams, direct reads, advanced POST body/query shapes, binary policy, and sensitive/admin reverse ETL actions.

# Auth setup

Use a HubSpot private app access token with the minimum scopes required for the operation being planned. Provide the token through an environment variable or stdin-backed credential flow. Never place token values in chat, shell history, documentation, fixtures, or JSON output.

# Streams notes

No HubSpot streams are executable in this metadata slice. Collection commands in `cli_surface.json` are marked `planned` and map to the future stream runner lane (#136). The initial `streams.json` only defines shared HTTP configuration and a bounded check request shape.

# Write actions & risks

No HubSpot write actions are executable in this metadata slice. Mutation commands in `cli_surface.json` are planned reverse ETL intents, never direct writes. Future write actions must use named actions in `writes.json`, fixed record schemas, path fields, fixture-backed request-shape tests, and plan → preview → approval → execute. Destructive/admin/sensitive operations require risk text, redaction policy, and typed confirmation.

# Known limits

- `api_surface.json` intentionally has no endpoints until issue #137 lands the full official operation ledger.
- `cli_surface.json` is command metadata only; no HubSpot command is runtime executable in this slice. Planned reverse ETL commands intentionally do not reference operation metadata until named `writes.json` actions exist.
- Direct reads need bounded output policies and redaction (#138).
- POST search/query bodies and binary/file transfers need fixed schemas and bounded policies (#139).
- Sensitive/admin/destructive writes need typed reverse ETL action metadata, redaction, and confirmation (#140).
