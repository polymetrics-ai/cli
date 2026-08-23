# Issue #4319 — TDD ledger

## Red

- `go test -timeout 20m ./cmd/connectorgen -run '^TestOperationEvidence' -count=1` failed before production implementation with undefined projector classes, gaps, artifact, and fixed-100 validator symbols. The exact compiler failure is recorded in `traces/red-operation-evidence.txt`.
- The new behavioral command tests cover a complete real GitHub definition row; each missing source/canonical/runtime/CLI/website/fixture/conformance evidence kind; a provider-evidenced unavailable surface; deterministic duplicate rollup and byte stability; and all 100 one-at-a-time aggregate regressions.

## Green

- `go test -timeout 20m ./cmd/connectorgen -run '^(TestOperationEvidenceProjectsGitHubAcrossEveryEvidenceSurface|TestOperationEvidenceReportsEachMissingEvidenceKind|TestOperationEvidenceRecordsProviderEvidencedAbsence|TestOperationEvidenceIsByteStableAndRollsUpDuplicates|TestOperationEvidenceCheckRunsFixed100Gate)$' -count=1` passed in 10.329s.
- `go test -timeout 20m ./cmd/connectorgen -run '^TestOperationEvidenceFixed100RejectsEveryRegression$' -count=1` passed in 4.131s. It mutates every one of the independent 100 source IDs and verifies the gate names the regression.
- `make connectorgen-operation-evidence` passed: the generated 1,525-row GitHub-source artifact is byte-current, has five deduplicated gap rollups, and the fixed-100 gate passed.
- After review refinements, `go test -timeout 20m ./cmd/connectorgen -run '^TestOperationEvidence' -count=1` passed in 13.176s. This includes duplicate source-row deduplication and a real GraphQL acronym field (`createIpAllowListEntry`) matched to its fixed operation document.
- After merging v3 source locks from `origin/main` at `cf493b834`, the same targeted behavioral suite passed in 14.161s; `make connectorgen-operation-evidence` and the full `make verify` (including lint) passed. The 1,525-row artifact and fixed-100 cohort were byte-identical, proving that the v3 representation did not change GitHub's projected evidence.

## Refactor / review

- GraphQL matches operation identity rather than its shared `POST /graphql` transport endpoint; this prevented one source field from falsely claiming every GraphQL command as evidence.
- No parser/schema change in `sourceimport.go` was made. The read-only projector consumes source-lock v2/v3 and certification provenance through their existing checked-in interfaces.
