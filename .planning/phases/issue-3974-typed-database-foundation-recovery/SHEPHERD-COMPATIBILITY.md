# #3995 Shepherd-compatible verdict — Issue #3974

## Bounded result

`SHEPHERD-COMPATIBILITY.json` uses the exact version-1 #3995 verdict fields:
`schema_version`, `connector`, `transition`, `decision`, and `failures`.
It is **not** an automatic #3995 approval. The #3995 implementation at
`ef66991350a5de137ab17e492acf2640feecf5c3` is not an ancestor of either this
recovery child or `origin/feat/3972-postgres-parity`, so its evaluator is not
part of the needed ancestry and is deliberately neither imported nor replayed.

The bounded compatibility decision is `RETRY` for the future
`integrate_sub_pr` transition. Current generated evidence says PostgreSQL is
`COMMUNITY BUILD, UNCERTIFIED` and `capability_complete=false`; #3974 must not
change that because this is F1 policy/admission work, not a certification,
write, query, or CDC promotion. The #3995 decision therefore blocks a future
in-scope certification transition but does not pretend to reject the currently
unmerged foundation PR.

This compatibility record is read-only: it uses the committed/generated
certification artifacts after their #4026 synchronization and records no
credential, live proof, or provider interaction. It does not consume a new
correction loop because no integrated #3995 evaluator emitted a retry and no
source defect is being repaired.
