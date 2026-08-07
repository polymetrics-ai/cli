# Asana documented-operation parity — plan

> **Generated mechanically** from `data/cli-top50-fixed-schema-sweep-r1/MASTER-PLAN.json`
> (`asana`, landing order 14). The operation count below was derived in the sweep-wide
> artifact pass, not re-reasoned here. **Per-connector findings are NOT pre-planned** — if this
> connector surprises you during implementation, STOP and record it rather than forcing it into
> this shape.

Part of `cli-top50-fixed-schema-sweep-r1`. One connector, one PR.

## Operation surface, derived before authoring

- **Artifact**: `https://raw.githubusercontent.com/Asana/openapi/56796a67a3c093eedf55fd9682357957a2ebfd85/defs/asana_oas.yaml`
- **Kind**: openapi, OAS `3.0.0`
- **Retrieved**: 2026-08-07, 3066750 bytes
- **Documented operations: 249**
- **By method**: DELETE 23, GET 119, POST 81, PUT 26
- **Read / write split**: 119 read, 130 write
- **Deprecated (still counted)**: 0

### Against the provider-artifact ledger

| | |
| --- | ---: |
| Ledger recorded | 249 |
| Re-derived | 249 |
| Delta | 0 |

The ledger's figure **reconciles exactly** with the live artifact.

**How it was counted.** Parsed with yaml.safe_load (pinned to commit 56796a67a3c093eedf55fd9682357957a2ebfd85), walked `paths`, counted one operation per (path, method) pair for real HTTP methods only. All 81 POST paths were individually reviewed by name; the six semantically ambiguous ones (/batch, /exports/graph, /exports/resource, /organization_exports, /custom_fields/{gid}/enum_options/insert, /projects/{gid}/sections/insert) were checked against their actual summary/description text -- each creates a resource (a batch-request object, an export job, a reordered enum/section list) rather than retrieving/listing one, so all remain writes; none match the search/query/batchGet read pattern (confirmed the true search-shaped paths, /workspaces/{id}/projects/search, /workspaces/{id}/tasks/search, and /workspaces/{id}/typeahead, are all GETs already counted as reads). Cross-checked with a duplicate-key-detecting YAML loader (0 duplicates) and an independent recount; both reproduced 249/119/81/26/23 exactly.

## Hazards

- None recorded during derivation beyond the standard bar.

## Webhook inventory (input for `cli-webhook-surface-sweep-r1`)

- **Webhook events: 0** — excluded from the operation total per the counting policy.
- **Webhook management endpoints: 5** — these stay **in scope** and are counted in the
  total. Create/list/update/delete of a webhook *subscription* is an ordinary REST operation; only
  webhook *events* are deferred.

## Required scopes — all five must be covered

Every documented operation must land as an ETL stream, a reverse-ETL write, a direct read, a direct
write, or a binary transfer, **and** be individually reachable as its own `pm asana <command>`.
Every `api_surface.json` row must carry exactly one of `executable`,
`blocked-with-named-dependency`, or `unsupported-with-source-citation` — **never blank, and never
the legacy `excluded` category**.

## Issues

Parent **#380** (old generation); children are expected at **#381–#386** (old-generation pattern) — **CONFIRM with `gh-axi` before using them in a PR body**, this is inferred from the pattern, not verified.

Use `Closes` only for what this PR genuinely completes and `Refs` for the rest.

## TDD sequence — the red test is NOT generated, by design

1. **RED** — write `cmd/connectorgen/asana_api_surface_test.go` **against this connector's real
   bundle** and **watch it fail**. Paste the verbatim failure into `TDD-LEDGER.md` and set
   `tdd.red_confirmed` / `tdd.red_failure` in `RUN-STATE.json`. Check first for a pre-existing
   surface test (`ls cmd/connectorgen/ | grep '^asana_'`) — gong carried two, and a targeted
   `-run` missed the second.
2. **GREEN** — author the bundle to satisfy it.
3. **REFACTOR** — docs, catalogs, operation endpoint ledger resync.
4. Gates, then no-mistakes.

`check_red_observed.py` refuses to let this connector proceed to implementation until the red
failure is real observed output.

## Safety notes

- Do not loosen `connectorgen validate`, the connector boundary gate, `certify`, or
  `TestEveryImplementedCommandPassesRuntimePreflight` to make this pass.
- Nothing is marked `implemented` unless its command genuinely runs; every block names a dependency.
- Run the WHOLE `cmd/connectorgen` package plus `internal/cli`, never just a targeted `-run`.
- Regenerating docs rewrites ~1,034 files of pre-existing `main` drift (finding F4) — revert every
  non-asana path.
- Inspect the website catalog diff **by object, not by line**.
- No credential or token-derived value is ever emitted.
