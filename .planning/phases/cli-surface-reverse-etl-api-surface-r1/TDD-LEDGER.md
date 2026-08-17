# TDD ledger — reverse-ETL API-surface derivation r1

## Planned red/green evidence

| Slice | Red evidence | Green evidence | Status |
| --- | --- | --- | --- |
| Repository endpoint-summary invariant | New all-bundle test finds the pre-existing 214 GitHub commands with empty API surface and parseable summary. | The same test reports zero violations after regeneration, while listing exactly the 14 intentionally nonmatching aliases. | planned |
| Generic synchronization | Minimal implemented reverse-ETL fixture with `POST /widgets/{id}` summary does not receive `api_surface`; a human-summary alias remains empty. | The address fixture gets exactly one matching endpoint, including a trailing action annotation variant; the alias remains untouched. | planned |
| Generated GitHub recovery | Base sweep: 228 empty API surfaces, split 214 parseable and 14 aliases. | Regenerated sweep: all 214 recoverable rows have one endpoint; only the 14 aliases remain empty. | planned |
| Status-accounting preservation | Base buckets: `1466 + 25 + 50 + 29 + 1 = 1571`. | After metadata regeneration the same buckets sum to 1,571; each bucket delta is zero. | planned |

## Constraints

- No hand-edited generated artifacts; every generated change comes from its owning generator.
- No connector-specific identifiers, allowlists, status reclassification, or test/gate weakening.
- The red run is performed before the production fix against the existing 214 failures; it proves the invariant detects the actual defect without mutating a generated artifact by hand.
- No credentials or provider writes are needed.
