# TDD ledger — reverse-ETL API-surface derivation r1

## Planned red/green evidence

| Slice | Red evidence | Green evidence | Status |
| --- | --- | --- | --- |
| Repository endpoint-summary invariant | The initial broad predicate failed on 248 rows: 214 GitHub raw endpoints and 34 Workday strings with sentence punctuation. The canonical endpoint join distinguishes the 214 declared addresses from the 34 nonmatching strings. | The canonical all-bundle invariant passes after generated synchronization; GitHub's 14 human-summary aliases do not match. | green |
| Generic synchronization | `TestSyncBundleDerivesAPISurfaceFromEndpointSummary` failed: the reverse-ETL address fixture filled zero API-surface fields. | The address fixture gets exactly one matching endpoint, including a trailing action annotation variant; the alias remains untouched. | green |
| Generated GitHub recovery | Base sweep: 228 empty API surfaces, split 214 canonical endpoint summaries and 14 aliases. | Regenerated sweep: all 214 recoverable rows have one endpoint; only the 14 aliases remain empty. The source-bound rule does not invent 34 Workday endpoints from punctuated prose. | green |
| Status-accounting preservation | Base buckets: `1466 + 25 + 50 + 29 + 1 = 1571`. | After metadata regeneration the same buckets remain `1466 + 25 + 50 + 29 + 1 = 1571`; every bucket delta is zero. | green |

## Constraints

- No hand-edited generated artifacts; every generated change comes from its owning generator.
- No connector-specific identifiers, allowlists, status reclassification, or test/gate weakening.
- The broad red run was performed before the production fix against the existing 248 strings. Its 34 Workday false candidates proved that the invariant must join the declared API endpoint inventory rather than normalize endpoint-like prose; the resulting canonical predicate detects the 214 actual GitHub defects without mutating a generated artifact by hand.
- No credentials or provider writes are needed.
