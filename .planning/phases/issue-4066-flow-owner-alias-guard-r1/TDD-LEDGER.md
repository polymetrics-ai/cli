# #4066 TDD Ledger

**Starting head:** `c5b91917e3f5c07a010db2bdf58348cbc73cb9d5`  
**Collision correction head:** `bda85b778f89f4320760b8d83826ac9d393b0220`  
**State:** GREEN verified with focused race coverage

| Slice | RED contract | GREEN contract | Result |
|---|---|---|---|
| 1. Omitted flow owner aliases | A real two-owner Parquet/DuckDB flow using either quoted or unquoted generated `records__<connection-id>` succeeds before the policy exists. | Both aliases return `*warehouse.AmbiguousTableError` for `records`, include the flow `connection` remedy, and leave no successful result or checkpoint. | GREEN 2026-08-11 |
| 2. Generic containment | `pm query run --sql` can currently use the generated alias. | The same generic command still returns the selected owner while only the flow origin is denied. | GREEN 2026-08-11 |
| 3. Regression fence | Existing flow coverage proves selected, bare ambiguous, `_unattributed`, `SELECT 1`, and action-source behavior. | One focused race run stays green after the structural policy is added. | GREEN 2026-08-11 |
| 4. Generated-alias collision | A unique real `records__<connection-id>` table is registered as a bare view before the unscoped-flow policy can resolve the identifier to `records`. | Quoted and unquoted omitted-flow queries fail with the typed `records` ambiguity and remedy; generic quoted and unquoted queries return the real table. | GREEN 2026-08-11 |

## Red command

`go test -timeout 20m ./internal/cli -run '^TestFlowOmittedConnectionRejectsGeneratedOwnerAliases$' -count=1`

Expected before production code: exit 1 because the unscoped flow can read the
generated alias.

## Recorded RED

**2026-08-11:** the red command exited 1 at the exact production head
`c5b91917e3f5c07a010db2bdf58348cbc73cb9d5`. Both the unquoted and quoted
flow subtests failed at `require.Error`: the generated owner alias returned a
successful flow result instead of a typed ambiguity. The generic CLI query
subtest retained the selected `globex` row and `SELECT 1` remained successful.
The non-secret command output is in `traces/red-flow-owner-alias-guard.txt`.

## Collision RED command

`go test -timeout 20m ./internal/cli -run '^TestFlowGeneratedOwnerAliasCollision$' -count=1`

Expected before the collision fix: exit 1 because the omitted-connection flow
uses the colliding real bare view instead of returning the `records` ambiguity.

## Recorded collision RED

**2026-08-11:** the collision RED command exited 1 at exact production head
`bda85b778f89f4320760b8d83826ac9d393b0220`. Both omitted-flow forms failed at
`require.Error` because the real colliding table returned a successful result.
Both generic controls failed because the generated alias view was registered
before the same-named real bare view, producing DuckDB's duplicate-view error.
The non-secret command output is in
`traces/red-flow-owner-alias-collision.txt`.

## Green command

`go test -race -timeout 20m ./internal/cli -run '^(TestFlowOmittedConnectionRejectsGeneratedOwnerAliases|TestFlowSourceConnectionSelectorsReadOnlyOwningRows|TestFlowSourceConnectionSelectorRefusesOmissionAndAcceptsUnattributed|TestFlowActionSourceReadsAllSelectedConnectionRows)$' -count=1`

**2026-08-11:** PASS. The new quoted and unquoted flow aliases return the
base-table ambiguity and existing manifest remedy, while the identical generic
CLI alias returns the `globex` row. The focused matrix also preserves explicit
flow selection, bare ambiguity, `_unattributed`, `SELECT 1`, and action-source
behavior under the race detector. Output is recorded in
`traces/green-flow-owner-alias-guard.txt`.

## Collision GREEN command

`go test -race -timeout 20m ./internal/cli -run '^(TestFlowOmittedConnectionRejectsGeneratedOwnerAliases|TestFlowGeneratedOwnerAliasCollision|TestFlowSourceConnectionSelectorsReadOnlyOwningRows|TestFlowSourceConnectionSelectorRefusesOmissionAndAcceptsUnattributed|TestFlowActionSourceReadsAllSelectedConnectionRows)$' -count=1`

**2026-08-11:** PASS. The collision policy preserves a real catalog table for
generic quoted and unquoted queries, rejects those same identifiers in omitted
flows as a typed `records` ambiguity with the flow remedy, and retains the
non-colliding generic alias plus the existing scope regression fence. Output is
recorded in `traces/green-flow-owner-alias-collision.txt`.
