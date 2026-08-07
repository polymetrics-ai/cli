# Square documented-operation parity — plan

> **Generated mechanically** from `data/cli-top50-fixed-schema-sweep-r1/MASTER-PLAN.json`
> (`square`, landing order 22). The operation count below was derived in the sweep-wide
> artifact pass, not re-reasoned here. **Per-connector findings are NOT pre-planned** — if this
> connector surprises you during implementation, STOP and record it rather than forcing it into
> this shape.

Part of `cli-top50-fixed-schema-sweep-r1`. One connector, one PR.

## Operation surface, derived before authoring

- **Artifact**: `https://raw.githubusercontent.com/square/connect-api-specification/master/api.json`
- **Kind**: openapi, OAS `3.0.0`
- **Retrieved**: 2026-08-07, 3269568 bytes
- **Documented operations: 334**
- **By method**: DELETE 27, GET 121, HEAD 0, OPTIONS 0, PATCH 0, POST 150, PUT 36, TRACE 0
- **Read / write split**: 158 read, 176 write
- **Deprecated (still counted)**: 26

### Against the provider-artifact ledger

| | |
| --- | ---: |
| Ledger recorded | 334 |
| Re-derived | 334 |
| Delta | 0 |

The ledger's figure **reconciles exactly** with the live artifact.

**How it was counted.** Counted every unique METHOD+path under `paths` in the Fern-generated OpenAPI 3.0.0 doc (150 POST, 121 GET, 36 PUT, 27 DELETE = 334, exactly matching the ledger). Read/write split for POST used explicit /search paths plus Square's own `_READ`/`_WRITE` OAuth security scopes as an authoritative secondary signal: every Retrieve*/BatchRetrieve*/BulkRetrieve*/Calculate* operation carries only a `_READ` scope (or none), confirming non-mutating semantics; CreateSubscription-style operations that also require `_READ` scopes for referenced resources were kept as writes because their primary action (create/update) still mutates state.

## Hazards

- None recorded during derivation beyond the standard bar.

## Webhook inventory (input for `cli-webhook-surface-sweep-r1`)

- **Webhook events: 0** — excluded from the operation total per the counting policy.
- **Webhook management endpoints: 8** — these stay **in scope** and are counted in the
  total. Create/list/update/delete of a webhook *subscription* is an ordinary REST operation; only
  webhook *events* are deferred.

## Required scopes — all five must be covered

Every documented operation must land as an ETL stream, a reverse-ETL write, a direct read, a direct
write, or a binary transfer, **and** be individually reachable as its own `pm square <command>`.
Every `api_surface.json` row must carry exactly one of `executable`,
`blocked-with-named-dependency`, or `unsupported-with-source-citation` — **never blank, and never
the legacy `excluded` category**.

## Issues

Parent **#3191** (new generation); children are expected at **#3192–#3198** (new-generation pattern) — **CONFIRM with `gh-axi` before using them in a PR body**, this is inferred from the pattern, not verified.

Use `Closes` only for what this PR genuinely completes and `Refs` for the rest.

## TDD sequence — the red test is NOT generated, by design

1. **RED** — write `cmd/connectorgen/square_api_surface_test.go` **against this connector's real
   bundle** and **watch it fail**. Paste the verbatim failure into `TDD-LEDGER.md` and set
   `tdd.red_confirmed` / `tdd.red_failure` in `RUN-STATE.json`. Check first for a pre-existing
   surface test (`ls cmd/connectorgen/ | grep '^square_'`) — gong carried two, and a targeted
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
  non-square path.
- Inspect the website catalog diff **by object, not by line**.
- No credential or token-derived value is ever emitted.
