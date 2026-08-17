# TDD ledger — reverse-ETL API-surface derivation r1

## Planned red/green evidence

| Slice | Red evidence | Green evidence | Status |
| --- | --- | --- | --- |
| Repository endpoint-summary invariant | `go test -timeout 20m ./cmd/connectorgen -run '^TestImplementedEndpointSummaryAlwaysHasAPISurface$' -count=1` failed on 248 rows: 214 GitHub and 34 Workday REST endpoint summaries. | The invariant passes after generated synchronization; GitHub's 14 human-summary aliases do not match. | green |
| Generic synchronization | `TestSyncBundleDerivesAPISurfaceFromEndpointSummary` failed: the reverse-ETL address fixture filled zero API-surface fields. | The address fixture gets exactly one matching endpoint, including a trailing action annotation variant; the alias remains untouched. | green |
| Generated GitHub recovery | Base sweep: 228 empty API surfaces, split 214 parseable and 14 aliases. | Regenerated sweep: all 214 recoverable rows have one endpoint; only the 14 aliases remain empty. The generic invariant also corrects 34 Workday rows. | green |
| Status-accounting preservation | Base buckets: `1466 + 25 + 50 + 29 + 1 = 1571`. | After metadata regeneration the same buckets remain `1466 + 25 + 50 + 29 + 1 = 1571`; every bucket delta is zero. | green |

## Constraints

- No hand-edited generated artifacts; every generated change comes from its owning generator.
- No connector-specific identifiers, allowlists, status reclassification, or test/gate weakening.
- The red run was performed before the production fix against the existing 248 failures; it proves the invariant detects the actual defect without mutating a generated artifact by hand.
- No credentials or provider writes are needed.
