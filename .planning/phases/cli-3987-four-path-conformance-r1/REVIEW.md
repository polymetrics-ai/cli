# Code review — #3987 four-path warehouse conformance

## Method

Inline/manual review followed `scripts/gsd prompt code-review cli-3987-four-path-conformance-r1 --auto`. A compatible isolated reviewer is unavailable and the delivery contract forbids role spawning, so the changed test code and planning evidence were inspected directly after all focused and broader gates passed.

## Findings

No blocking, warning, or informational findings.

- The app matrix is derived from `certificationcatalog.FlowKinds` and fails if the generated catalog gains, loses, or changes a role without a corresponding explicit contract.
- The matrix uses actual registered GitHub/PostgreSQL descriptors for direction identity and persists source-bound connection selection before calling production dispatch logic.
- The shared transport proof records stage/reopen/apply/read-back/checkpoint order and asserts that a destination receives the exact sealed receipt and reopened workset bound to the source connection.
- The mode proof treats `incremental_dedupe_history` as executable current behavior and gives the sole non-pass `change_capture` result a concrete refusal reason, excluding it from the pass count.
- No credentials, live writes, connector definition changes, or changes to the protected certification roll-up were introduced.
