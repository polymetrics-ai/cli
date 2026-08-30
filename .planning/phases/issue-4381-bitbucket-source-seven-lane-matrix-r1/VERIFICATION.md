# Verification checklist — Bitbucket Track A

- [x] `jq empty internal/connectors/defs/bitbucket/sources/bitbucket-source-lane-matrix.json`
- [x] `go test ./internal/connectors/defs/bitbucket -run '^TestBitbucketSourceLaneMatrix' -count=1`
- [x] `go test ./internal/connectors/defs/bitbucket -count=1`
- [x] `go run ./cmd/connectorgen declaration-admission` — passed (`1 connector(s), 1 source operation(s), 0 finding(s)`).
- [x] `go run ./cmd/connectorgen source-import bitbucket --out internal/connectors/defs/bitbucket/sources/bitbucket-operation-descriptor.json --check` — correctly ran in check mode and made no writes; it is blocked by the pre-existing missing `sources/bitbucket-retained-artifacts.json` manifest.
- [x] `go run ./cmd/connectorgen validate --connector bitbucket` — matrix produced no finding; the command remains blocked by the pre-existing missing `sources/bitbucket-operation-descriptor.json`.
- [x] `go run ./cmd/connectorgen surface-sync --check --connector bitbucket` — check-only mode made no writes and is blocked by that same pre-existing missing canonical descriptor.
- [x] `git diff --no-index --check /dev/null <each owned new file>` — passed for all seven owned paths (used because they are intentionally untracked until the scoped commit).
- [x] `git status --short` contains only the owned planning and Bitbucket matrix/test paths.

The whole-definition command `go run ./cmd/connectorgen validate internal/connectors/defs` was also run. It fails on pre-existing non-Bitbucket gaps (missing Batch R1 descriptors, Docker Hub body-schema safety findings, and Sentry source-projection findings); it does not report a Bitbucket matrix finding. This Track A delivery deliberately does not create descriptors, retained artifacts, or generated projections.

No credentialed provider check, reverse-ETL execution, generator write mode, shared runtime change, parent integration, PR opening, or main merge is part of this verification.
