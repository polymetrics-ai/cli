# Overview

HubSpot is being added through the issue #132 CLI parity roadmap. This bundle is the issue #134 metadata scaffold: it records the connection spec and provider-like command surface, but it does not yet claim executable streams, direct reads, binary transfers, or reverse ETL writes.

Official operation inventory comes from the public HubSpot OpenAPI collection at <https://github.com/HubSpot/HubSpot-public-api-spec-collection>. Issue #137 generated the full non-executable operation ledger from 401 OpenAPI JSON files: 4,396 raw versioned operations deduplicated to 3,060 unique method/path operations (GET 1,038; POST 1,314; PUT 169; PATCH 232; DELETE 307). Later lanes implement streams, direct reads, advanced POST body/query shapes, binary policy, and sensitive/admin reverse ETL actions.

# Auth setup

Use a HubSpot private app access token with the minimum scopes required for the operation being planned. Provide the token through an environment variable or stdin-backed credential flow. Never place token values in chat, shell history, documentation, fixtures, or JSON output.

# Streams notes

No HubSpot streams are executable in this metadata slice. Collection commands in `cli_surface.json` are marked `planned` and map to the future stream runner lane (#136). The initial `streams.json` only defines shared HTTP configuration and a bounded check request shape.

# Write actions & risks

No HubSpot write actions are executable in this metadata slice. Mutation commands in `cli_surface.json` are planned reverse ETL intents, never direct writes. Future write actions must use named actions in `writes.json`, fixed record schemas, path fields, fixture-backed request-shape tests, and plan → preview → approval → execute. Destructive/admin/sensitive operations require risk text, redaction policy, and typed confirmation.

# Operation ledger status

`api_surface.json` uses `operation_ledger_version: 1`. Every official operation row is represented as a blocked-by-default typed app-operation candidate. The current classifier counts are:

| Model | Count |
| --- | ---: |
| `stream_etl` | 244 |
| `query_etl` | 223 |
| `direct_read` | 759 |
| `reverse_etl` | 850 |
| `binary_read` | 30 |
| `binary_write` | 31 |
| `sensitive_reverse_etl` | 40 |
| `admin_reverse_etl` | 291 |
| `destructive_action` | 556 |
| `deprecated` | 22 |
| `disallowed` | 14 |

# Known limits

- `api_surface.json` is complete inventory/classification metadata, not runtime dispatch. Rows remain blocked by default until their safe execution lane lands.
- `cli_surface.json` is command metadata only; no HubSpot command is runtime executable in this slice. Planned reverse ETL commands intentionally do not reference operation metadata until named `writes.json` actions exist.
- Stream candidates need bounded schemas, fixtures, and runner coverage (#136).
- Direct reads need bounded output policies and redaction (#138).
- POST search/query bodies and binary/file transfers need fixed schemas and bounded policies (#139).
- Sensitive/admin/destructive writes need typed reverse ETL action metadata, redaction, and confirmation (#140).
