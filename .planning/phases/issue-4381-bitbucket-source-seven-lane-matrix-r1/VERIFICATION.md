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

## Semantic-source repair — 2026-08-31

- [x] `go test -timeout 20m ./internal/connectors/defs/bitbucket -run '^TestBitbucketSourceLaneMatrix' -count=1` — focused semantic, red/green, edge, backlink, count, binary, webhook, and source-boundary checks passed.
- [x] `go test -timeout 20m ./internal/connectors/defs/bitbucket -count=1` — complete Bitbucket package passed.
- [x] `go test -race -timeout 20m ./internal/connectors/defs/bitbucket -run '^TestBitbucketSourceLaneMatrix' -count=1` — passed.
- [x] `go vet ./internal/connectors/defs/bitbucket` — passed.
- [x] `jq empty internal/connectors/defs/bitbucket/sources/bitbucket-source-lane-matrix.json` and `jq empty internal/connectors/defs/bitbucket/sources/bitbucket-operation-source-lock.json` — passed; the lock was read only.
- [x] `go run ./cmd/agentcontractgen check` — passed: canonical contract and registered projections are current.
- [x] `git diff --check` — passed.
- [x] Inline GSD verify/review fallback — the generated `verify-work` and `code-review` prompts were inspected; the changed source selector, matrix backlinks, parser failure paths, preservation query, and scope boundary were manually reviewed with no finding. The isolated runner may not spawn the prompt’s reviewer role.
- [x] Connector-local preservation query — 297/297 binary-download, binary-upload, sync-transport, and event-cursor records are unchanged; the only newly declared ETL IDs are `searchAccount`, `searchTeam`, and `searchWorkspace`; direct-read/direct-write dispositions are unchanged.
- [x] `GOFLAGS='-p=3' go test -timeout 30m ./...` — started as an additional broad check, remained CPU-active in unrelated `connectorgen`, application, CLI, and boundary packages after 20 minutes, then was deliberately interrupted to release the shared machine. This is neither a Bitbucket failure nor a substitute for the passed complete Bitbucket package test; it remains a follow-up CI check rather than a scoped merge blocker.

No credential, provider request, generator write mode, reverse-ETL execution, runtime/transport/certification change, PR creation, or merge occurred in this repair.
