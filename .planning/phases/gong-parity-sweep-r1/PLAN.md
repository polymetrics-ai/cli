# Gong documented-operation parity — plan

Part of `cli-top50-fixed-schema-sweep-r1`. One connector, one PR.

## Goal

Bring `internal/connectors/defs/gong` from **67 declared endpoints** to the **full 69-operation
documented surface**, with every operation partitioned exactly once and individually reachable as
`pm gong <command>`.

## Operation surface, derived before authoring

Artifact: Gong's own OpenAPI document, served by the official API documentation page at
`https://gong.app.gong.io/ajax/settings/api/documentation/specs?version=` — fetched 2026-08-07,
HTTP 200, 453,605 bytes, `openapi: 3.0.1`, `info.title: "Gong API"`, `info.version: V2`.
This is the exact URL the provider-artifact ledger records, and it is publicly reachable
without authentication.

**69 operations across 59 paths. GET 29, POST 28, PUT 8, DELETE 3, PATCH 1.**

The ledger's carried-forward **69 reconciles exactly** with the live artifact — no drift. This is
the second OpenAPI-backed connector in this sweep to reconcile at zero, consistent with finding F1
(drift concentrates in `html_reference` records, not machine-readable ones).

### Read/write split — a correction to the ledger's split, not its total

The ledger records `29 read / 40 write`, which is a naive HTTP-method split. Gong documents three
**POST read-query** endpoints that the sweep's counting policy classifies semantically as reads:

- `POST /v2/calls/extensive`
- `POST /v2/calls/transcript`
- `POST /v2/stats/interaction`

The existing surface test already pins these as "official POST read-query endpoints". Counting
policy is "GET/HEAD are reads; POST classified semantically", so the honest split is
**32 read / 37 write**. The **total of 69 is unaffected**, and the ledger's total is confirmed.

### Baseline, measured by running the binary — not by reading files

`go build ./cmd/pm` then `--help` on every one of the 67 declared commands: **67 reachable, 0
failed**. Intents partition 12 `etl` / 29 `direct_read` / 26 `reverse_etl`, matching
`api_surface.json` `covered_by` exactly (12 stream, 29 direct_read, 26 write). Availability is
43 `implemented` / 24 `partial`. No blank dispositions: all 67 rows carry `covered_by`.

So gong is **not** a rebuild like notion was. Its gap is exactly two operations.

## The gap: 2 operations, both in the Targets area

Set-differenced spec against bundle: 2 operations in the spec are missing from the bundle, and
**zero rows in the bundle are absent from the spec** (no stale rows to retire).

| Operation | opId | Shape | Planned bucket |
| --- | --- | --- | --- |
| `GET /v2/targets` | `listTargetDefinitions` | required query `workspaceId`; response `{requestId, targets}` — **no cursor** | **Direct read** |
| `POST /v2/targets/{targetId}/assignments` | `uploadAssignments` | path `targetId`, required query `workspaceId`, optional `validateOnly`; body **`multipart/form-data`, single `file` part, `format: binary`** | **Binary upload → reverse-ETL write** |

`GET /v2/targets` returns an unpaginated envelope, so it is a **direct read, not an ETL stream**.
Declaring it a stream would advertise pagination the provider does not offer.

`POST /v2/targets/{targetId}/assignments` is a genuine **binary** operation. It goes through the
**existing engine multipart capability** (`engine.MultipartSpec`, `body_type: "multipart"`,
`bundle.go:566/747`), never ad-hoc HTTP. gong already ships two working exemplars of exactly this
contract in its own `writes.json` — `upload_call_media` and `upload_crm_entities` — so this is
copying a shape that already runs, not inventing one.

`Targets` is currently absent from `pm gong`'s command groups; it becomes a new group.

## Planned partition after the change

| Bucket | Before | After |
| --- | --- | --- |
| ETL streams | 12 | 12 |
| Direct reads | 29 | **30** |
| Reverse-ETL / direct writes | 26 | **27** (of which 3 binary multipart) |
| **Total** | **67** | **69** |

## Deprecated operation — counted, already covered

The artifact marks exactly one operation deprecated: `POST /v2/flows/prospects/assign/cool-off-override`
(`assignProspectsCoolOffOverrides`). It is documented, so it counts toward 69, and it is already
present in the bundle. No change.

## Webhook inventory (input for `cli-webhook-surface-sweep-r1`)

The artifact declares **no top-level `webhooks` block** (OAS 3.0.1 has no such construct) and
**0 webhook events**. Webhook *management* stays in scope and is already covered: the existing
surface test asserts `GET /v2/settings/webhooks` is deliberately absent as a wrong-method row, and
the current bundle carries the correct webhook-settings rows. Nothing deferred for gong.

Because the artifact is OAS **3.0.1**, finding F2 does not apply — `connectorgen batch materialize`
is not blocked by top-level `webhooks` here.

## TDD sequence

1. **RED** — update `cmd/connectorgen/gong_api_surface_test.go` from 67 to **69**, with the method
   split GET 29 / POST 28 / PUT 8 / DELETE 3 / PATCH 1, and add explicit presence assertions for the
   two Targets endpoints. It must fail against today's 67-row bundle. Capture failure text in
   `TDD-LEDGER.md`.
2. **GREEN** — author the two operations across `api_surface.json`, `cli_surface.json`,
   `writes.json`, `operations.json`, and schemas.
3. **REFACTOR** — docs, catalogs, operation endpoint ledger resync.
4. Gates, then no-mistakes.

## Safety notes

- Do not loosen `connectorgen validate`, the connector boundary gate, `certify`, or
  `TestEveryImplementedCommandPassesRuntimePreflight` to make this pass.
- Nothing is marked `implemented` unless its command runs; blocked rows carry a **named** dependency.
- The binary upload takes a bounded **local file path**, and file path/content are redacted in plans
  exactly as `upload_call_media` and `upload_crm_entities` already do. No credential or
  token-derived value is ever emitted.
- Keep the diff scoped to gong; revert unrelated generator churn.
