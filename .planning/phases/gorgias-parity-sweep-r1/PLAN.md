# Gorgias documented-operation parity — plan

> **Generated mechanically** from `data/cli-top50-fixed-schema-sweep-r1/MASTER-PLAN.json`
> (`gorgias`, landing order 6). The operation count below was derived in the sweep-wide
> artifact pass, not re-reasoned here. **Per-connector findings are NOT pre-planned** — if this
> connector surprises you during implementation, STOP and record it rather than forcing it into
> this shape.

Part of `cli-top50-fixed-schema-sweep-r1`. One connector, one PR.

## Operation surface, derived before authoring

- **Artifact**: `https://dash.readme.com/api/v1/api-registry/1qfhqbgmshn434r (gorgias-rest-api.json, OpenAPI 3.1.0, bound to the STABLE ReadMe docs version 1.5.1)`
- **Kind**: openapi, OAS `3.1.0`
- **Retrieved**: 2026-08-07, 852105 bytes
- **Documented operations: 114**
- **By method**: DELETE 18, GET 46, POST 23, PUT 27
- **Read / write split**: 50 read, 64 write
- **Deprecated (still counted)**: 1

### Against the provider-artifact ledger

| | |
| --- | ---: |
| Ledger recorded | 114 |
| Re-derived | 114 |
| Delta | 0 |

The ledger's figure **reconciles exactly** with the live artifact.

**How it was counted.** Ledger recorded https://developers.gorgias.com/reference as html_reference. That page is genuinely ReadMe.com-hosted; its embedded project JSON lists the ReadMe docs version tree and shows the STABLE version (version_clean 1.5.1, is_stable:true) has one registered OpenAPI definition: filename 'gorgias-rest-api.json', uuid '1qfhqbgmshn434r'. A second registry entry, 'public-rest-api-spec.json' (uuid 2roode5mjcv4gkg), belongs to version 1.2, which is is_stable:false / is_beta:true, so it was excluded as non-current. Fetched the stable spec via ReadMe's public api-registry endpoint https://dash.readme.com/api/v1/api-registry/1qfhqbgmshn434r (HTTP 200, application/json, 852,105 bytes, openapi 3.1.0, 61 path templates). Counted unique (METHOD, path) pairs directly from this spec: 114 with zero internal collisions — an exact match to ledger_total.

## Hazards

- None recorded during derivation beyond the standard bar.

## Webhook inventory (input for `cli-webhook-surface-sweep-r1`)

- **Webhook events: 0** — excluded from the operation total per the counting policy.
- **Webhook management endpoints: 0** — these stay **in scope** and are counted in the
  total. Create/list/update/delete of a webhook *subscription* is an ordinary REST operation; only
  webhook *events* are deferred.

## Required scopes — all five must be covered

Every documented operation must land as an ETL stream, a reverse-ETL write, a direct read, a direct
write, or a binary transfer, **and** be individually reachable as its own `pm gorgias <command>`.
Every `api_surface.json` row must carry exactly one of `executable`,
`blocked-with-named-dependency`, or `unsupported-with-source-citation` — **never blank, and never
the legacy `excluded` category**.

## Issues

Parent **#196** (old generation); children are expected at **#197–#202** (old-generation pattern) — **CONFIRM with `gh-axi` before using them in a PR body**, this is inferred from the pattern, not verified.

Use `Closes` only for what this PR genuinely completes and `Refs` for the rest.

## TDD sequence — the red test is NOT generated, by design

1. **RED** — write `cmd/connectorgen/gorgias_api_surface_test.go` **against this connector's real
   bundle** and **watch it fail**. Paste the verbatim failure into `TDD-LEDGER.md` and set
   `tdd.red_confirmed` / `tdd.red_failure` in `RUN-STATE.json`. Check first for a pre-existing
   surface test (`ls cmd/connectorgen/ | grep '^gorgias_'`) — gong carried two, and a targeted
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
  non-gorgias path.
- Inspect the website catalog diff **by object, not by line**.
- No credential or token-derived value is ever emitted.
