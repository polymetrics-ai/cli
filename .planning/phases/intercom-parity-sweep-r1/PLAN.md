# Intercom documented-operation parity — plan

> **Generated mechanically** from `data/cli-top50-fixed-schema-sweep-r1/MASTER-PLAN.json`
> (`intercom`, landing order 20). The operation count below was derived in the sweep-wide
> artifact pass, not re-reasoned here. **Per-connector findings are NOT pre-planned** — if this
> connector surprises you during implementation, STOP and record it rather than forcing it into
> this shape.

Part of `cli-top50-fixed-schema-sweep-r1`. One connector, one PR.

## Operation surface, derived before authoring

- **Artifact**: `https://developers.intercom.com/_bundle/docs/references/@2.16/rest-api/api.intercom.io.yaml`
- **Kind**: openapi, OAS `3.0.1`
- **Retrieved**: 2026-08-07, 1362253 bytes
- **Documented operations: 231**
- **By method**: DELETE 31, GET 108, PATCH 1, POST 68, PUT 23
- **Read / write split**: 114 read, 117 write
- **Deprecated (still counted)**: 0

### Against the provider-artifact ledger

| | |
| --- | ---: |
| Ledger recorded | 324 |
| Re-derived | 231 |
| Delta | -93 |

**Undetermined** — a defensible count could not be established from a current artifact.

**How it was counted.** Fetched api.intercom.io.yaml@2.16 (info.version=2.16, openapi=3.0.1, 164 path items, 1,362,253 bytes) with curl --compressed, parsed with pyyaml (yaml.safe_load succeeded, no external $ref path items, no unresolved refs, 0 operations with 'callbacks'). Counted one operation per real HTTP method key (get/put/post/delete/patch/head/options/trace) under each path item, excluding the shared 'parameters' sibling key. Result: 231 operations across GET=108, POST=68, PUT=23, DELETE=31, PATCH=1; all 231 (method,path) pairs are unique (0 duplicates). 0 operations flagged 'deprecated: true'. No top-level 'webhooks' object (n/a for OAS 3.0.1 regardless). To sanity-check the large gap vs ledger_total=324, fetched every available historical version of the SAME bundle (@2.10 through @2.16) and re-ran the identical parser: 2.10=108, 2.11=108, 2.12=118, 2.13=127, 2.14=150, 2.15=162, 2.16=231 -- a clean monotonic growth curve, and @2.17..@2.21/@unstable/@latest all 404 (2.16 is confirmed the current max). No available snapshot of this specific artifact -- past or present -- reaches 324, so operations_total is reported as the artifact-verified 231, not adjusted toward the ledger.

## Hazards

- None recorded during derivation beyond the standard bar.

## Webhook inventory (input for `cli-webhook-surface-sweep-r1`)

- **Webhook events: 0** — excluded from the operation total per the counting policy.
- **Webhook management endpoints: 0** — these stay **in scope** and are counted in the
  total. Create/list/update/delete of a webhook *subscription* is an ordinary REST operation; only
  webhook *events* are deferred.

## Required scopes — all five must be covered

Every documented operation must land as an ETL stream, a reverse-ETL write, a direct read, a direct
write, or a binary transfer, **and** be individually reachable as its own `pm intercom <command>`.
Every `api_surface.json` row must carry exactly one of `executable`,
`blocked-with-named-dependency`, or `unsupported-with-source-citation` — **never blank, and never
the legacy `excluded` category**.

## Issues

Parent **#164** (old generation); children are expected at **#165–#170** (old-generation pattern) — **CONFIRM with `gh-axi` before using them in a PR body**, this is inferred from the pattern, not verified.

Use `Closes` only for what this PR genuinely completes and `Refs` for the rest.

## TDD sequence — the red test is NOT generated, by design

1. **RED** — write `cmd/connectorgen/intercom_api_surface_test.go` **against this connector's real
   bundle** and **watch it fail**. Paste the verbatim failure into `TDD-LEDGER.md` and set
   `tdd.red_confirmed` / `tdd.red_failure` in `RUN-STATE.json`. Check first for a pre-existing
   surface test (`ls cmd/connectorgen/ | grep '^intercom_'`) — gong carried two, and a targeted
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
  non-intercom path.
- Inspect the website catalog diff **by object, not by line**.
- No credential or token-derived value is ever emitted.
