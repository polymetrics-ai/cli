# Verification checklist — Issue #4390

## Completed checks

- Completed — JSON parsing: `jq empty internal/connectors/defs/asana/{enabled_connector_contract,missing-foundation}.json internal/connectors/defs/asana/sources/asana-source-lane-matrix.json`.
- Completed — deliberate red: `TestAsanaSourceLaneArtifactsProjectTheTrackAMatrix` hides `sync_transport.json`, removes an applicable direct-read backlink, and promotes a mapped-unproven ETL mapping in memory. Each produces the expected validation error before its checked-in green assertion.
- Completed — focused green: `go test ./internal/connectors/defs/asana -run TestAsanaSourceLaneArtifactsProjectTheTrackAMatrix -count=1`.
- Completed — full connector-local green: `go test ./internal/connectors/defs/asana -count=1`.
- Completed — definition validation: `go run ./cmd/connectorgen validate internal/connectors/defs/asana` reports `1 connector(s) checked, 0 findings`.
- Completed — enabled-contract/source checks: focused `enabledcontract_final` tests, retained Asana source-lock/import tests, and `go run ./cmd/connectorgen declaration-admission internal/connectors/defs --json` are green; admission reports no findings.
- Completed — non-writing generated-surface check: `go run ./cmd/connectorgen surface-reconcile internal/connectors/defs/asana --check --json` exits 0 with no pending reclassification.
- Completed — changed-path review and `git diff --check`.

## Carried-forward and new evidence counts

- Carried forward: 7 foundation entries, preserving their original IDs, states, affected lanes, reasons, source references, and artifact-role inventory. Their added exact matrix projections contain 12, 22, 1, 1, 260, 2, and 447 lane cells respectively.
- Newly added: 1 connector-local gap entry for the 52 `etl:mapped_unproven` scope/fanout cells. The pre-existing incremental-event limitation remains in its carried-forward foundation entry with 11 ETL plus 11 sync-not-applicable cells.
- No shared runtime foundation is requested or implemented.

## Recorded non-writing baseline results

- `go run ./cmd/connectorgen source-import asana --check` exits 1 with pre-existing derived CLI drift (`writes=0 cli=297`).
- `go run ./cmd/connectorgen source-import asana --read-projection-only --check` exits 1 with pre-existing descriptor/projection drift (`writes=0 cli=0`).
- `go run ./cmd/connectorgen source-materialize asana --check` correctly refuses the retained v3 lock: only a v4 lock with materialization block is supported.
- The unchanged `TestRetainedAsanaMutationDispositionsCoverEveryDeferredSourceOperation` remains red on pre-existing generated-surface coverage findings. It does not read `missing-foundation.json`; the task changes none of its source mutation-dispositions/descriptor/writes/operations/API/CLI inputs, so this evidence is retained rather than suppressed.

## Runtime/foundation boundary

- No shared runtime foundation is implemented in this task.
- Any present connector-local gap record must be exact source-matrix evidence with a typed reason and Atlas lookup; it cannot become an execution selector.
