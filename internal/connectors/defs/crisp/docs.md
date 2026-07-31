# Crisp connector

## Overview

This bundle is the connector-local Crisp REST API V1 official operation ledger for parent issue #204. It inventories the official Crisp REST API reference at `https://docs.crisp.chat/references/rest-api/v1/` and deduplicates operations by HTTP method and canonical path.

The current documentation parse found 234 documented method/path rows. Method counts: DELETE 26, GET 91, HEAD 14, PATCH 44, POST 47, PUT 12. Lane counts: ETL/read 82, direct/provider-search/query 12, binary/file 8, changefeed/events 4, reverse/write 114, HEAD checks 14. The parent r2 audit allocated 220 non-HEAD rows; this ledger additionally records the 14 current official HEAD existence checks as planned/blocked non-data operations.

No local row is claimed as implemented, fixture-tested, certified, or live-safe in this wave. `metadata.json` keeps `check`, `read`, `write`, `query`, and `cdc` false until executable connector-local evidence exists.

## Auth setup

Crisp REST API authentication uses Basic authentication with a token keypair plus the `X-Crisp-Tier` header. The official authentication guides document website-token requests as `Authorization: Basic BASE64(token_id:token_key)` with `X-Crisp-Tier: website`, and plugin-token requests as `Authorization: Basic BASE64(identifier:key)` with `X-Crisp-Tier: plugin`.

Future executable Crisp commands should use a credential profile containing the token identifier and token key. Do not pass token values in prompt text, command examples, issue bodies, fixtures, or logs. Both `token_id` and `token_key` are marked `x-secret` in `spec.json`; this ledger slice does not perform authenticated provider calls.

## Streams notes

No Crisp ETL streams are enabled yet. The 82 ETL/read rows and 4 changefeed/event rows are represented in `api_surface.json`, `operations.json`, and planned command metadata as blocked rows until future connector-local lanes add named streams, schemas, pagination/cursor policy, sanitized fixtures, and conformance evidence.

Provider search/query/direct operations remain blocked pending shared foundation #2985 and must stay distinct from warehouse-focused `pm query`. CDC/changefeed rows remain blocked pending #2986/#2988 and connector-owned replay evidence.

## Write actions & risks

No Crisp reverse ETL write actions are enabled yet. The 114 reverse/write rows include DELETE, destructive, sensitive, and admin operations in scope. They are not blanket-excluded as unsafe. A future executable action must be named, schema-gated, redacted, risk documented, and routed through plan -> preview -> explicit approval -> execute; DELETE/destructive/admin actions must additionally require typed destructive confirmation such as `confirm: "destructive"` in `writes.json` plus idempotency notes where the provider supports safe retry/missing-resource semantics.

Binary/file/import/export rows are also blocked until a connector-owned fixed-target policy supplies size caps, path safety, redaction, and fixture/conformance evidence. No raw method/path/body, arbitrary GraphQL, shell, unrestricted file, generic SQL, or passthrough API command is exposed by this bundle.

## Known limits

- This is a complete documented operation ledger and planned typed command surface, not completed runtime parity.
- Shared provider search/query foundation #2985 is open, so search/query/direct commands remain planned and blocked.
- CDC truth/lab foundations #2986 and #2988 are open, so changefeed/event operations remain planned and blocked.
- `operations.json` preserves typed top-level request-body schemas parsed from the documentation, but nested provider-specific object fields remain bounded generic objects until future executable lanes add full schemas and fixtures.
- No live credentials, live provider calls, live writes, certification, VPS/Thaalam work, or merge activity were used to produce this bundle. Fixture-backed dynamic conformance has no executable stream or write action to run in this wave.
