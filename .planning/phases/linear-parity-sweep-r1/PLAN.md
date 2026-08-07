# Linear documented-operation parity — plan

> **Generated mechanically** from `data/cli-top50-fixed-schema-sweep-r1/MASTER-PLAN.json`
> (`linear`, landing order 28). The operation count below was derived in the sweep-wide
> artifact pass, not re-reasoned here. **Per-connector findings are NOT pre-planned** — if this
> connector surprises you during implementation, STOP and record it rather than forcing it into
> this shape.

Part of `cli-top50-fixed-schema-sweep-r1`. One connector, one PR.

## Operation surface, derived before authoring

- **Artifact**: `https://raw.githubusercontent.com/linear/linear/master/packages/sdk/src/schema.graphql`
- **Kind**: graphql_sdl
- **Retrieved**: 2026-08-07, 1270042 bytes
- **Documented operations: 538**
- **By method**: n/a
- **Read / write split**: 165 read, 373 write
- **Deprecated (still counted)**: 17

### Against the provider-artifact ledger

| | |
| --- | ---: |
| Ledger recorded | unknown |
| Re-derived | 538 |
| Delta | unknown |

**Undetermined** — a defensible count could not be established from a current artifact.

**How it was counted.** No ledger value was supplied for linear (recorded UNKNOWN in the task). Counted from the official public SDL schema at https://raw.githubusercontent.com/linear/linear/master/packages/sdk/src/schema.graphql (200 OK, 1270042 bytes / 50383 lines, retrieved 2026-08-07) rather than the given linear.app/developers/graphql landing page, which is prose/tutorial content with no operation enumeration of its own. This raw file is Linear's own machine-generated schema dump, refreshed by Linear's CI one day before retrieval. Parsed with the same brace/paren/bracket-depth-aware scanner used for monday (see monday.derivation_note for the exact @deprecated-directive false-positive bug caught and fixed during development, before it was applied here). Exactly one top-level 'type Query {}' (165 fields), one 'type Mutation {}' (373 fields), and one 'type Subscription {}' (81 fields) exist; no 'extend type Query/Mutation' fragments were found anywhere in the file. operations_total = 165 + 373 = 538, per the stated policy (reads=Query, writes=Mutation only; Subscription is neither and is excluded from the total, see surprises).

## Hazards

- **Path-collapse hazard**: Every Linear GraphQL operation is POST https://api.linear.app/graphql selected by the request body's operation document. api_surface rows must key on the operation NAME (e.g. 'Query.issues', 'Mutation.issueCreate'), not the request path, or the whole 538-operation surface collapses to one row. Rows in `api_surface.json` must be keyed so each operation stays distinct, or the whole surface collapses into a handful of rows. This has already bitten this programme once (DynamoDB's `X-Amz-Target`).

## Webhook inventory (input for `cli-webhook-surface-sweep-r1`)

- **Webhook events: 22** — excluded from the operation total per the counting policy.
- **Webhook management endpoints: 8** — these stay **in scope** and are counted in the
  total. Create/list/update/delete of a webhook *subscription* is an ordinary REST operation; only
  webhook *events* are deferred.

## Required scopes — all five must be covered

Every documented operation must land as an ETL stream, a reverse-ETL write, a direct read, a direct
write, or a binary transfer, **and** be individually reachable as its own `pm linear <command>`.
Every `api_surface.json` row must carry exactly one of `executable`,
`blocked-with-named-dependency`, or `unsupported-with-source-citation` — **never blank, and never
the legacy `excluded` category**.

## Issues

Parent **#80** (old generation); children are expected at **#81–#86** (old-generation pattern) — **CONFIRM with `gh-axi` before using them in a PR body**, this is inferred from the pattern, not verified.

Use `Closes` only for what this PR genuinely completes and `Refs` for the rest.

## TDD sequence — the red test is NOT generated, by design

1. **RED** — write `cmd/connectorgen/linear_api_surface_test.go` **against this connector's real
   bundle** and **watch it fail**. Paste the verbatim failure into `TDD-LEDGER.md` and set
   `tdd.red_confirmed` / `tdd.red_failure` in `RUN-STATE.json`. Check first for a pre-existing
   surface test (`ls cmd/connectorgen/ | grep '^linear_'`) — gong carried two, and a targeted
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
  non-linear path.
- Inspect the website catalog diff **by object, not by line**.
- No credential or token-derived value is ever emitted.
