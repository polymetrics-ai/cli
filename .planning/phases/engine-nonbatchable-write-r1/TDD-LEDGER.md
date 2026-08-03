# TDD LEDGER — engine primitive: non-batchable write actions

Red-before-green, manual GSD loop (adapter has no `programming-loop` command; fallback recorded in
`PLAN.md`).

## Red-first expectations

| # | Area | Red evidence before production edit | Green evidence |
| --- | --- | --- | --- |
| R1 | Schema accepts the field | `writes.json` declaring `"batchable": false` fails to load — `strictDecode` rejects the unknown field | loader accepts `true`, `false`, and absent; a non-boolean is rejected by the meta-schema |
| R2 | Safe default | no way to ask an action whether it is batchable | `IsBatchable()` is true for absent, true for explicit `true`, false for explicit `false`, and **true for the Go zero value** `WriteAction{}` |
| R3 | Existing bundles unchanged | — | every action across all embedded bundles reports batchable; no bundle file changed |
| R4 | Manifest propagation | `connectors.WriteActionSpec` cannot express batchability, so the app cannot see it | a bundle's `batchable: false` reaches `connectors.ManifestOf(c).WriteActions[i].IsBatchable() == false` |
| R5 | Definition propagation | `pm connectors describe`-shaped `WriteActionInfo` cannot express it | `Definition().WriteActions[i]` carries the same value; existing actions omit the JSON key entirely |
| R6 | No pointer aliasing | a shared `*bool` would let a manifest consumer mutate the loaded bundle | mutating the returned manifest's `*Batchable` does not change the bundle's action |
| R7 | **Bulk plan refused** | `PlanReverseETL` creates a plan and mints an approval token for a `batchable: false` action | `PlanReverseETL` returns `*NonBatchableActionError`; `ListReversePlans()` is empty; no approval token exists |
| R8 | Refusal is actionable | a bare error would name neither the action nor the alternative | error text contains the action name, the connector name, the source table, the reason, and the `pm <connector> <command>` to run instead |
| R9 | Refusal is programmatically detectable | callers can only string-match | `errors.As(err, &*NonBatchableActionError)` succeeds and exposes the fields |
| R10 | Execute-time re-check | a plan created before the declaration (or hand-written into `state.json`) executes in bulk | `RunReverseETL` refuses the SourceTable-mode plan even with a valid approval token, mirroring the `confirmationChallengeForPlan` live-manifest precedent |
| R11 | **Individually executable** | — | the same `batchable: false` action plans as a connector command, previews, approves, and executes, and the real HTTP request arrives at the destination server |
| R12 | Batchable actions unaffected | — | a `batchable: true` action and an absent-field action both plan and execute in bulk unchanged; all pre-existing app/engine tests stay green |
| R13 | Help/manual parity | manifest guide output cannot mention batchability | a non-batchable action renders a bulk-reverse-ETL line in `pm help <connector>` / manual output; connectors with no non-batchable action render byte-identically to before |

## Notes

- R7 asserts **no plan persisted**, not just an error return. Refusing after minting an approval
  token would leave an approvable artifact behind.
- R11 is the half that makes the primitive worth having; it is proven by observing the request on an
  `httptest` server, not by reading the schema.
- R13's "byte-identically" clause is what keeps the generated website catalog and golden transcripts
  from churning, since no shipped bundle declares the field.
