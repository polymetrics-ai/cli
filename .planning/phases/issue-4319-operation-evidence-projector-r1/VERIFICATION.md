# Issue #4319 — verification

## Result

Passed locally on the final staged tree.

## Evidence

- `go test -timeout 20m ./cmd/connectorgen -run '^TestOperationEvidence' -count=1` — passed in 13.176s.
- `go run ./cmd/connectorgen operation-evidence --write-fixed-100` — generated a byte-stable 1,525-row artifact and fixed cohort.
- `make connectorgen-operation-evidence` — passed; artifact current, five rollups, fixed-100 passed.
- `make verify` — passed in full, including formatting, tidy, vet, `go test -timeout 20m ./...`, build, docs, smoke, lint, connector generation/validation, boundary, and release tooling.

## Current-main revalidation

- Merged `origin/main` at `cf493b83455aca3dc38164cee01520f5be5803cf`, which adds the v3 multi-document source-lock model, without rebasing the published branch.
- `go test -timeout 20m ./cmd/connectorgen -run '^TestOperationEvidence' -count=1` — passed in 14.161s against the merged tree.
- `make connectorgen-operation-evidence` — passed against the v3 locks: the checked-in evidence artifact remains byte-current at 1,525 rows and five rollups, and the fixed-100 gate passes. No projector artifact changed because v3 preserves the effective GitHub operation evidence consumed by this projector.
- `make verify` — passed in full after the merge, including lint and the final release-installed GitHub certification check.
- Frozen GitHub source lock: 3,420,025 bytes; SHA-256 `281b1cfcc67eb63e19ef83daf06197bf3d3b23db0b6bc9b73e02fc18ee278fb6`.
- Frozen GitHub descriptor: 43,354,021 bytes; SHA-256 `d1978c0c6fd0eb66e9fcd4d78d637864a6e486f558aaad1e51550bc43758b899`.

## Scope and dependency

The projector consumes the v2/v3 source-lock interface read-only and needs no
parser or schema change. The current-main validation proves the v3
multi-document/provenance representation remains compatible with the GitHub
evidence projection.
