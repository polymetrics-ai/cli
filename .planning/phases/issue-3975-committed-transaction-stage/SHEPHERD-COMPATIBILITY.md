# #3995 Shepherd-compatible verdict — Issue #3975

## Bounded result

`SHEPHERD-COMPATIBILITY.json` uses the exact version-1 #3995 verdict fields:
`schema_version`, `connector`, `transition`, `decision`, and `failures`. It is
**not** an automatic #3995 approval. The #3995 implementation at
`ef66991350a5de137ab17e492acf2640feecf5c3` is an ancestor of neither this
child nor `origin/feat/3972-postgres-parity`; importing or replaying it would
violate #3975's source-agnostic staging scope.

The compatibility decision is `RETRY` for the future `integrate_sub_pr`
transition. #3975 supplies only a reusable private stage and durable receipt
seam; it does not provide PostgreSQL source proof, certification evidence, a
CDC capability change, or LSN acknowledgement. The record therefore blocks a
future certification/integration claim without pretending to reject the
unmerged child implementation itself.

This is read-only local evidence. It uses no credential, provider, database,
or external mutable action, consumes no extra correction round, and does not
override the user-directed hold on push/PR/CI.
