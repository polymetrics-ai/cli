# Review — GitLab Track A source-to-seven-lane matrix

## Verdict

Pass for the scoped Track A mapping delivery. The change set is connector-local source retention, a source-lane matrix, one reconciliation test, and issue-scoped planning evidence only. It adds no executor, importer, generator, command, stream, write definition, or shared Foundation Atlas implementation.

## Reviewed invariants

- Exact transferred source bytes are bound by Git blob SHA-1 and size; retained rendered-reference artifacts are additionally checked by SHA-256 and byte count.
- The matrix preserves all 1,754 retained source rows and only classifies the three crosswalk-only identities as `not_source_row` boundary evidence.
- Each row has precisely the seven required cells. GET, POST/PUT/PATCH/DELETE, binary, explicit pagination, and hook-registration decisions are recomputed from locked source facts in the test.
- Direct-write and reverse-ETL cells remain independent and every DELETE is retained as a mutation candidate.
- No applicable cell is promoted to `implemented`; the three `/hooks` source-registration cells remain `missing_foundation` with a typed, source-listed Atlas lookup.
- `rest.path_bridge` and the two malformed required-path facts remain source-visible mapping restrictions, never foundation claims and never row-removal justifications.

## Verification outcome

- Focused test and edge variants: pass.
- JSON syntax and package vet: pass.
- Canonical agent-contract check: pass.
- `source-import --check --read-projection-only` and connector validation both stop at the pre-existing `path_bridge` parser acceptance gap. This delivery records it but intentionally does not change shared code.

## Remaining review boundary

The shared source-lock parser must later accept and preserve the established `rest.path_bridge` field before its generic source-import/projection checks can pass. That is a mapping-control repair outside this issue's connector-local scope.
