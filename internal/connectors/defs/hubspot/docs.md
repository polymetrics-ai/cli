# HubSpot connector

## Overview

This bundle is the connector-local HubSpot official API operation ledger for parent issue #132. It inventories the HubSpot public OpenAPI 3.0 spec collection at commit `2bebde2dca45eaa1792931089c4e441c8e377594` and deduplicates versioned definitions by HTTP method and path. The inventory found 4466 versioned operation entries across 524 OpenAPI files and 3118 unique operations. Unique method counts: GET 1057, POST 1330, PUT 177, PATCH 239, DELETE 315.

The reconciled lane allocation matches the parent issue counts: ETL/read 925, reverse ETL write 1704, direct/provider search/query 260, binary/file 229, CDC 0, total 3118. No local row is claimed as implemented, fixture-tested, certified, or live-safe in this wave.

## Auth setup

Future executable HubSpot commands should use a credential profile containing a private app or OAuth access token. Do not pass token values in prompt text, command examples, issue bodies, fixtures, or logs. The optional `access_token` field is marked `x-secret` in `spec.json`; this ledger slice does not perform authenticated provider calls.

## Streams notes

No HubSpot ETL streams are enabled yet. The 925 ETL/read operations are represented in `api_surface.json`, `operations.json`, and planned command metadata as blocked rows until future connector-local lanes add named streams, schemas, pagination/cursor policy, sanitized fixtures, and conformance evidence. Provider search/query operations remain blocked pending shared foundation #2985 and must stay distinct from warehouse-focused `pm query`.

## Write actions & risks

No HubSpot reverse ETL write actions are enabled yet. The 1,704 reverse/write operations, including 315 DELETE operations and 522 operations classified destructive by method or summary/path keywords, are intentionally included in scope. They are not blanket-excluded as unsafe. A future executable action must be named, schema-gated, redacted, risk documented, and routed through plan -> preview -> explicit approval -> execute; DELETE/destructive actions must additionally require typed destructive confirmation (for example `confirm: "destructive"` in a write action).

Binary/file operations are also blocked until a connector-owned fixed-target policy supplies size caps, path safety, redaction, and fixture/conformance evidence. No raw method/path/body, arbitrary GraphQL, shell, unrestricted file, generic SQL, or passthrough API command is exposed by this bundle.

## Known limits

- This is a complete documented operation ledger, not completed runtime parity. `metadata.json` keeps `read`, `write`, `query`, and `cdc` false until executable evidence exists.
- Shared provider search/query foundation #2985 is open, so search/query/direct commands remain planned and blocked.
- CDC truth/lab foundations #2986 and #2988 are open; HubSpot has no counted CDC operations in parent #132 and this bundle claims no CDC capability.
- The generated operation body schemas intentionally avoid examples/default values and preserve only structural OpenAPI type/ref information to prevent secret-like fixture/doc literals.
- No live credentials, live provider calls, live writes, or certification were used to produce this bundle. Fixture-backed dynamic conformance skips because no stream or write action is executable in this wave.
