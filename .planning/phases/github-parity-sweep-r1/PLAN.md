# Github documented-operation parity — plan

> **Generated mechanically** from `data/cli-top50-fixed-schema-sweep-r1/MASTER-PLAN.json`
> (`github`, landing order 27). The operation count below was derived in the sweep-wide
> artifact pass, not re-reasoned here. **Per-connector findings are NOT pre-planned** — if this
> connector surprises you during implementation, STOP and record it rather than forcing it into
> this shape.

Part of `cli-top50-fixed-schema-sweep-r1`. One connector, one PR.

## Operation surface, derived before authoring

- **Artifact**: `https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.json`
- **Kind**: openapi, OAS `3.0.3`
- **Retrieved**: 2026-08-07, 12920264 bytes
- **Documented operations: 1220**
- **By method**: DELETE 187, GET 636, HEAD 0, OPTIONS 0, PATCH 70, POST 193, PUT 134, TRACE 0
- **Read / write split**: 641 read, 579 write
- **Deprecated (still counted)**: 37

### Against the provider-artifact ledger

| | |
| --- | ---: |
| Ledger recorded | 1220 |
| Re-derived | 1220 |
| Delta | 0 |

The ledger's figure **reconciles exactly** with the live artifact.

**How it was counted.** Fetched with curl, HTTP 200 (12,920,264 bytes, ~12.3 MiB). Parsed with json.load (no memory issues). Counted one operation per (METHOD, path) pair under top-level `paths` ONLY; `x-webhooks` (the other relevant top-level key) is a disjoint sibling object, so there is no double-counting risk between the two. Confirmed no `/graphql` path exists anywhere in `paths`, and info.description literally reads "GitHub's v3 REST API" -- this artifact is REST-only, satisfying rule 7 (GraphQL is not merged in). POST reads vs writes: keyword pass (same regex list as jira/zendesk) then manual description review of every hit, PLUS two extra manual passes over the write bucket looking for (a) compute-only operationId/summary stems (evaluate/analyse/render/check/export/computed/etc.) and (b) get/fetch/retrieve-prefixed operationIds -- pass (b) found nothing beyond what pass (a) already caught. Net corrections: credentials/revoke (matched 'list' via its summary 'Revoke a list of credentials' but its description says 'Submit a list of credentials to be revoked' -> write), issues/approve-suggestion and issues/dismiss-suggestion (both matched 'suggest' only via the 'suggestions' resource name in the path; both explicitly transition suggestion state per description -> write); apps/check-token, markdown/render, markdown/render-raw (none matched any keyword; all three are non-mutating per description and response codes [200/304/422/404, no 201/202 resource-creation signal] -> read).

## Hazards

- None recorded during derivation beyond the standard bar.

## Webhook inventory (input for `cli-webhook-surface-sweep-r1`)

- **Webhook events: 270** — excluded from the operation total per the counting policy.
- **Webhook management endpoints: 28** — these stay **in scope** and are counted in the
  total. Create/list/update/delete of a webhook *subscription* is an ordinary REST operation; only
  webhook *events* are deferred.

## Required scopes — all five must be covered

Every documented operation must land as an ETL stream, a reverse-ETL write, a direct read, a direct
write, or a binary transfer, **and** be individually reachable as its own `pm github <command>`.
Every `api_surface.json` row must carry exactly one of `executable`,
`blocked-with-named-dependency`, or `unsupported-with-source-citation` — **never blank, and never
the legacy `excluded` category**.

## Issues

Parent **#2989** (new generation); children are expected at **#2990–#2996** (new-generation pattern) — **CONFIRM with `gh-axi` before using them in a PR body**, this is inferred from the pattern, not verified.

Use `Closes` only for what this PR genuinely completes and `Refs` for the rest.

## TDD sequence — the red test is NOT generated, by design

1. **RED** — write `cmd/connectorgen/github_api_surface_test.go` **against this connector's real
   bundle** and **watch it fail**. Paste the verbatim failure into `TDD-LEDGER.md` and set
   `tdd.red_confirmed` / `tdd.red_failure` in `RUN-STATE.json`. Check first for a pre-existing
   surface test (`ls cmd/connectorgen/ | grep '^github_'`) — gong carried two, and a targeted
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
  non-github path.
- Inspect the website catalog diff **by object, not by line**.
- No credential or token-derived value is ever emitted.
