# Monday documented-operation parity — plan

> **Generated mechanically** from `data/cli-top50-fixed-schema-sweep-r1/MASTER-PLAN.json`
> (`monday`, landing order 17). The operation count below was derived in the sweep-wide
> artifact pass, not re-reasoned here. **Per-connector findings are NOT pre-planned** — if this
> connector surprises you during implementation, STOP and record it rather than forcing it into
> this shape.

Part of `cli-top50-fixed-schema-sweep-r1`. One connector, one PR.

## Operation surface, derived before authoring

- **Artifact**: `https://api.monday.com/v2/get_schema?format=sdl`
- **Kind**: graphql_sdl
- **Retrieved**: 2026-08-07, 542501 bytes
- **Documented operations: 292**
- **By method**: n/a
- **Read / write split**: 96 read, 196 write
- **Deprecated (still counted)**: 2

### Against the provider-artifact ledger

| | |
| --- | ---: |
| Ledger recorded | 280 |
| Re-derived | 292 |
| Delta | 12 |

**Finding: the ledger is stale.** The live artifact disagrees; see the note below.

**How it was counted.** Counted from the LIVE production schema fetched directly and unauthenticated from monday's own introspection-SDL endpoint https://api.monday.com/v2/get_schema?format=sdl (200 OK, 542501 bytes, retrieved 2026-08-07), not from the given developer.monday.com HTML reference page. That page's left-nav (enumerated via developer.monday.com/sitemap.xml, 160 unique /api-reference/reference/<slug> pages) is organized by ENTITY/TOPIC (e.g. 'boards', 'items', 'columns', each bundling its queries + mutations + supporting-type docs on one page), so it has no 1:1 page-per-operation structure and cannot be mechanically counted as N operations - the live SDL is the only reliable enumerable source. Parsed the SDL with a brace/paren/bracket-depth-aware scanner (handles multi-line argument lists and nested default-value object/list literals) to find exactly one top-level 'type Query { }' block (96 fields) and one 'type Mutation { }' block (196 fields); the schema{} declaration only lists query+mutation (no Subscription type exists in this API). Explicitly excluded directive invocations like @deprecated(reason:"...") from field detection - an initial naive pass ('identifier immediately followed by (') produced 2 spurious 'field' entries literally named 'deprecated' by matching the @deprecated(...) directive call itself; this was caught by a duplicate-name sanity check and fixed by excluding any identifier immediately preceded by '@'. operations_total = 96 + 196 = 292.

## Hazards

- **Path-collapse hazard**: Every monday.com GraphQL operation is POST https://api.monday.com/v2 (schema introspection is a separate GET-style /v2/get_schema endpoint but all real reads/writes go through the single POST /v2 endpoint). api_surface rows must key on the operation NAME (e.g. 'Query.boards', 'Mutation.create_item'), not the request path, or the whole 292-operation surface collapses to one row. Rows in `api_surface.json` must be keyed so each operation stays distinct, or the whole surface collapses into a handful of rows. This has already bitten this programme once (DynamoDB's `X-Amz-Target`).

## Webhook inventory (input for `cli-webhook-surface-sweep-r1`)

- **Webhook events: 21** — excluded from the operation total per the counting policy.
- **Webhook management endpoints: 3** — these stay **in scope** and are counted in the
  total. Create/list/update/delete of a webhook *subscription* is an ordinary REST operation; only
  webhook *events* are deferred.

## Required scopes — all five must be covered

Every documented operation must land as an ETL stream, a reverse-ETL write, a direct read, a direct
write, or a binary transfer, **and** be individually reachable as its own `pm monday <command>`.
Every `api_surface.json` row must carry exactly one of `executable`,
`blocked-with-named-dependency`, or `unsupported-with-source-citation` — **never blank, and never
the legacy `excluded` category**.

## Issues

Parent **#82** (old generation); children are expected at **#83–#88** (old-generation pattern) — **CONFIRM with `gh-axi` before using them in a PR body**, this is inferred from the pattern, not verified.

Use `Closes` only for what this PR genuinely completes and `Refs` for the rest.

## TDD sequence — the red test is NOT generated, by design

1. **RED** — write `cmd/connectorgen/monday_api_surface_test.go` **against this connector's real
   bundle** and **watch it fail**. Paste the verbatim failure into `TDD-LEDGER.md` and set
   `tdd.red_confirmed` / `tdd.red_failure` in `RUN-STATE.json`. Check first for a pre-existing
   surface test (`ls cmd/connectorgen/ | grep '^monday_'`) — gong carried two, and a targeted
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
  non-monday path.
- Inspect the website catalog diff **by object, not by line**.
- No credential or token-derived value is ever emitted.
