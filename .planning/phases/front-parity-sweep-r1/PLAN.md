# Front documented-operation parity — plan

> **Generated mechanically** from `data/cli-top50-fixed-schema-sweep-r1/MASTER-PLAN.json`
> (`front`, landing order 15). The operation count below was derived in the sweep-wide
> artifact pass, not re-reasoned here. **Per-connector findings are NOT pre-planned** — if this
> connector surprises you during implementation, STOP and record it rather than forcing it into
> this shape.

Part of `cli-top50-fixed-schema-sweep-r1`. One connector, one PR.

## Operation surface, derived before authoring

- **Artifact**: `https://raw.githubusercontent.com/frontapp/front-api-specs/main/core-api/core-api.json`
- **Kind**: openapi, OAS `3.0.0`
- **Retrieved**: 2026-08-07, 500400 bytes
- **Documented operations: 244**
- **By method**: DELETE 27, GET 123, PATCH 23, POST 70, PUT 1
- **Read / write split**: 123 read, 121 write
- **Deprecated (still counted)**: 11

### Against the provider-artifact ledger

| | |
| --- | ---: |
| Ledger recorded | 255 |
| Re-derived | 244 |
| Delta | -11 |

**Finding: the ledger is stale.** The live artifact disagrees; see the note below.

**How it was counted.** Located via GitHub API tree listing of frontapp/front-api-specs (2 spec files found: core-api/core-api.json and channel-api/channel-api.json; README.md confirms these are two distinct products -- Core API for 'backend integrations', Channel API for 'building partner channels'). Counted operations_total=244 from https://raw.githubusercontent.com/frontapp/front-api-specs/main/core-api/core-api.json (HTTP 200, 500400 bytes, openapi=3.0.0, info.version=1.0.0, 147 paths), per this task's own hint to find 'the main' spec and per this repo's internal/connectors/defs/front/api_surface.json and metadata.json, which explicitly scope the 'front' connector to 'Front Core REST API' only (docs_url points at Core API docs) and make no reference to Channel API. IMPORTANT RECONCILIATION FINDING: channel-api.json alone has 11 operations, and 244 + 11 = 255 EXACTLY matches ledger_total. However this naive sum double-counts: PATCH /channels/{channel_id} (operationId update-channel, summary 'Update Channel') is documented VERBATIM IDENTICALLY in BOTH spec files (same operationId, same summary, same request-body schema ref, same server https://api2.frontapp.com). The true deduplicated union of unique (METHOD,path) pairs across both files is 254, not 255. So the ledger's 255 looks like it was built by summing raw per-file operation counts across both products without (a) recognizing Channel API is a distinct out-of-scope product for this connector, or (b) deduplicating the one endpoint shared by both specs.

## Hazards

- **Spec is fragmented** across multiple files; confirm the assembled surface matches the provider's published total before authoring.

## Webhook inventory (input for `cli-webhook-surface-sweep-r1`)

- **Webhook events: 0** — excluded from the operation total per the counting policy.
- **Webhook management endpoints: 0** — these stay **in scope** and are counted in the
  total. Create/list/update/delete of a webhook *subscription* is an ordinary REST operation; only
  webhook *events* are deferred.

## Required scopes — all five must be covered

Every documented operation must land as an ETL stream, a reverse-ETL write, a direct read, a direct
write, or a binary transfer, **and** be individually reachable as its own `pm front <command>`.
Every `api_surface.json` row must carry exactly one of `executable`,
`blocked-with-named-dependency`, or `unsupported-with-source-citation` — **never blank, and never
the legacy `excluded` category**.

## Issues

Parent **#188** (old generation); children are expected at **#189–#194** (old-generation pattern) — **CONFIRM with `gh-axi` before using them in a PR body**, this is inferred from the pattern, not verified.

Use `Closes` only for what this PR genuinely completes and `Refs` for the rest.

## TDD sequence — the red test is NOT generated, by design

1. **RED** — write `cmd/connectorgen/front_api_surface_test.go` **against this connector's real
   bundle** and **watch it fail**. Paste the verbatim failure into `TDD-LEDGER.md` and set
   `tdd.red_confirmed` / `tdd.red_failure` in `RUN-STATE.json`. Check first for a pre-existing
   surface test (`ls cmd/connectorgen/ | grep '^front_'`) — gong carried two, and a targeted
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
  non-front path.
- Inspect the website catalog diff **by object, not by line**.
- No credential or token-derived value is ever emitted.
