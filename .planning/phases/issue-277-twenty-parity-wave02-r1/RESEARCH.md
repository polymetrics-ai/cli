# Research: Twenty CRM connector parity wave02 r1

## GitHub issue family

- #277 — parent, open: Twenty CRM all-ops CLI parity parent. Declares `https://api.twenty.com` base, `/rest` + `/graphql`, bearer auth, 28 objects, 546 fields, 56 reads, 112 writes, 168 operations.
- #278 — S1 foundation: `metadata.json`, `spec.json`, `api_surface.json`.
- #279 — S2 object schemas: 28 objects / 546 fields.
- #280 — S3 read streams: 28 stream reads + 28 direct-read get-by-id rows. Later ratified same-stream coverage for get-by-id rows because generic direct-read execution was not part of the slice.
- #281 — S4 reverse-ETL writes: 84 create/update/batch actions. Operator ratified shared `body_field` support for raw JSON array batch bodies; current main already has `body_field` in engine/schema.
- #282 — S5 destructive writes: 28 `DELETE /rest/<object>/{id}` actions with typed confirmation.
- #283 — S6 CLI surface + help/manual/website parity. Closed as complete on parent branch.
- #284 — S7 fixtures + docs.md + conformance/certify.
- #323 — nested follow-up under #277 for autonomous delivery control-plane hardening; not part of the connector bundle addendum scope in this worker because the task names #278-#284.

Parent PR #285 remains open, unmerged, and human-gated. It recorded integrated sub-PRs #287, #288, #290, #304, #317, #320, #322 and parent head `5199c0bb6519155cb0456fb3476e323ba9347d40`.

## Official Twenty sources inventoried

- `https://docs.twenty.com/developers/extend/api`
  - Twenty has schema-per-tenant APIs and no static public API reference for a specific workspace.
  - Core API is `/rest/` and `/graphql/` for CRUD on records such as People, Companies, Opportunities, and custom objects.
  - Metadata API is `/rest/metadata/` and `/metadata/` for schema management; this connector does not expose metadata admin actions as executable operations in this slice.
  - Base URL for cloud is `https://api.twenty.com/`; self-hosted is `https://{your-domain}/`.
  - Auth is `Authorization: Bearer YOUR_API_KEY`.
  - REST and GraphQL both support creating, reading, updating, deleting records and batch operations up to 60 records per request; GraphQL adds relation traversal and batch upsert.
  - Rate limits documented in the API overview: 100 requests/minute and batch size 60 records/call.
- `https://docs.twenty.com/llms.txt` and `https://docs.twenty.com/llms-full.txt`
  - Used to discover the current API overview and standard-object documentation snippets without a live workspace key.
- `gh-axi api '/repos/twentyhq/twenty/git/trees/main?recursive=1'`
  - Used to inventory open-source standard-object metadata paths under `packages/twenty-server/src/engine/workspace-manager/twenty-standard-application/`, including object, field, index, view, and page-layout metadata utilities.

## Operation ledger target

The connector bundle restored from the reviewed parent branch has these counts:

| Surface | Count | Representation |
| --- | ---: | --- |
| Standard objects | 28 | One stream/schema per object. |
| Read operations | 56 | `GET /rest/<object>` and `GET /rest/<object>/{id}`, both covered by the object stream. |
| Non-destructive reverse ETL | 84 | `create_*`, `update_*`, `batch_*` actions. |
| Destructive/admin DELETE | 28 | `delete_*` actions with `confirm: "destructive"`. |
| Total operations | 168 | `api_surface.json` rows. |

## Safety conclusions

- No live provider call or credential is required to restore and validate the connector-local bundle.
- The Twenty public docs explicitly describe both REST and GraphQL as schema-generated; a static complete workspace-specific GraphQL operation set cannot be invented without a workspace schema. The connector represents the standard-object REST CRUD surface and records GraphQL as a documented source/limitation, not a separate duplicate executable surface.
- Destructive DELETE operations are included in the ledger and executable only as named reverse-ETL delete actions with typed schemas, `confirm: "destructive"`, fixtures, and plan -> preview -> explicit approval -> execute.
