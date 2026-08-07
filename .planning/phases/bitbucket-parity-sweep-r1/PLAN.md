# Bitbucket documented-operation parity — plan

> **Generated mechanically** from `data/cli-top50-fixed-schema-sweep-r1/MASTER-PLAN.json`
> (`bitbucket`, landing order 21). The operation count below was derived in the sweep-wide
> artifact pass, not re-reasoned here. **Per-connector findings are NOT pre-planned** — if this
> connector surprises you during implementation, STOP and record it rather than forcing it into
> this shape.

Part of `cli-top50-fixed-schema-sweep-r1`. One connector, one PR.

## Operation surface, derived before authoring

- **Artifact**: `https://developer.atlassian.com/cloud/bitbucket/swagger.v3.json`
- **Kind**: openapi, OAS `3.0.0`
- **Retrieved**: 2026-08-07, 1485564 bytes
- **Documented operations: 331**
- **By method**: DELETE 54, GET 179, POST 50, PUT 48
- **Read / write split**: 181 read, 150 write
- **Deprecated (still counted)**: 52

### Against the provider-artifact ledger

| | |
| --- | ---: |
| Ledger recorded | 331 |
| Re-derived | 331 |
| Delta | 0 |

The ledger's figure **reconciles exactly** with the live artifact.

**How it was counted.** Fetched swagger.v3.json (openapi=3.0.0, info.version='2.0', x-revision=1eced60796b8, 193 path items, 1,485,564 bytes) with curl, parsed with json.load (no $ref path items, clean). Counted one operation per real HTTP method key under each path item (excluding sibling 'parameters'). Result: 331 operations across GET=179, POST=50, PUT=48, DELETE=54; all 331 (method,path) pairs unique. This equals ledger_total=331 exactly.

## Hazards

- None recorded during derivation beyond the standard bar.

## Webhook inventory (input for `cli-webhook-surface-sweep-r1`)

- **Webhook events: 0** — excluded from the operation total per the counting policy.
- **Webhook management endpoints: 12** — these stay **in scope** and are counted in the
  total. Create/list/update/delete of a webhook *subscription* is an ordinary REST operation; only
  webhook *events* are deferred.

## Required scopes — all five must be covered

Every documented operation must land as an ETL stream, a reverse-ETL write, a direct read, a direct
write, or a binary transfer, **and** be individually reachable as its own `pm bitbucket <command>`.
Every `api_surface.json` row must carry exactly one of `executable`,
`blocked-with-named-dependency`, or `unsupported-with-source-citation` — **never blank, and never
the legacy `excluded` category**.

## Issues

Parent **#79** (old generation); children are expected at **#80–#85** (old-generation pattern) — **CONFIRM with `gh-axi` before using them in a PR body**, this is inferred from the pattern, not verified.

Use `Closes` only for what this PR genuinely completes and `Refs` for the rest.

## TDD sequence — the red test is NOT generated, by design

1. **RED** — write `cmd/connectorgen/bitbucket_api_surface_test.go` **against this connector's real
   bundle** and **watch it fail**. Paste the verbatim failure into `TDD-LEDGER.md` and set
   `tdd.red_confirmed` / `tdd.red_failure` in `RUN-STATE.json`. Check first for a pre-existing
   surface test (`ls cmd/connectorgen/ | grep '^bitbucket_'`) — gong carried two, and a targeted
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
  non-bitbucket path.
- Inspect the website catalog diff **by object, not by line**.
- No credential or token-derived value is ever emitted.
