# HubSpot connector

## Overview

This bundle is the connector-local HubSpot official API operation ledger for parent issue #132. It inventories the HubSpot public OpenAPI 3.0 spec collection at commit `2bebde2dca45eaa1792931089c4e441c8e377594` and deduplicates versioned definitions by HTTP method and path. The inventory found 4466 versioned operation entries across 524 OpenAPI files and 3118 unique operations. Unique method counts: GET 1057, POST 1330, PUT 177, PATCH 239, DELETE 315.

The reconciled lane allocation matches the parent issue counts: ETL/read 925, reverse ETL write 1704, direct/provider search/query 260, binary/file 229, CDC 0, total 3118. No local row is claimed as implemented, fixture-tested, validated, or live-safe in this wave.

## Auth setup

HubSpot catalog and object reads use a credential profile containing a private app or OAuth access token. Do not pass token values in prompt text, command examples, issue bodies, fixtures, logs, schemas, or catalog caches. The optional `access_token` field is marked `x-secret` in `spec.json` and is used only by the runtime requester.

## Streams notes

At catalog time, the native adapter lists custom CRM schemas, combines them with a declared standard-object baseline, then asks HubSpot for every object's properties. The result is one dynamic stream per object type, including an account-created type the code has never named. Each stream schema contains only provider-described properties; a collection read requests and emits only those fields through `/crm/v3/objects/{objectType}`. Discovery uses a ten-worker bounded pool, provider-rate-limit retry/backoff, progress heartbeats, declared fallback, partial-status reporting, and an account-scoped cache keyed by connector plus opaque coordination identity. `pm catalog refresh` explicitly replaces that account catalog; multiple connections to the same account reuse it.

The 925 ETL/read operations are still represented in `execution bundle`, `operations.json`, and planned command metadata as blocked rows unless covered by that narrow discovered object collection route. Provider search/query operations remain blocked pending shared foundation #2985 and must stay distinct from warehouse-focused `pm query`.

## Write actions & risks

No HubSpot reverse ETL write actions are enabled yet. The 1,704 reverse/write operations, including 315 DELETE operations and 522 operations classified destructive by method or summary/path keywords, are intentionally included in scope. They are not blanket-excluded as unsafe. A future executable action must be named, schema-gated, redacted, risk documented, and routed through plan -> preview -> explicit approval -> execute; DELETE/destructive actions must additionally require typed destructive confirmation (for example `confirm: "destructive"` in a write action).

Binary/file operations are also blocked until a connector-owned fixed-target policy supplies size caps, path safety, redaction, and fixture/execution-contract evidence. No raw method/path/body, arbitrary GraphQL, shell, unrestricted file, generic SQL, or passthrough API command is exposed by this bundle.

## Known limits

- This is a complete documented operation ledger, not completed runtime parity. `metadata.json` enables only dynamic CRM object catalog/read; write, query, and CDC remain false until executable evidence exists.
- Shared provider search/query foundation #2985 is open, so search/query/direct commands remain planned and blocked.
- CDC truth/lab foundations #2986 and #2988 are open; HubSpot has no counted CDC operations in parent #132 and this bundle claims no CDC capability.
- The generated operation body schemas intentionally avoid examples/default values and preserve only structural OpenAPI type/ref information to prevent secret-like fixture/doc literals.
- No live credentials, live provider calls, live writes, or validation were used to produce this bundle. Fixture-backed native tests prove discovery and read mechanics, not a live account's provider behavior.
