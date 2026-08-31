# GitLab #4394 semantic POST mapping repair — verification checklist

## Required checks

- [x] Exact base recorded: `687eb1ded6b42cc456f8cc3c1e97f0a84fd042a8`; GitHub remote verified.
- [x] Baseline: `GOCACHE=/private/tmp/gocache-gitlab-4394-r4 go run ./cmd/connectorgen validate --connector gitlab` returned `1 connector(s) checked, 0 findings` before source edits.
- [x] Red test recorded against the uncorrected matrix reason.
- [x] `GOCACHE=/private/tmp/gocache-gitlab-4394-r4 go test -count=1 ./internal/connectors/defs/gitlab -run '^TestGitLabSourceLaneMatrixRetainsEveryLockedOperationAndLane$'` passes (2.943s); the same focused test passes under `-race` (27.094s).
- [x] Connector-shape source/contract checks pass without credentials or provider I/O: `go test -count=1 ./internal/connectors/defs/gitlab` (6.027s) and `go test -count=1 ./cmd/connectorgen -run '^TestGitLabEnabledContractReconcilesSourceLock$'` (1.547s).
- [x] `GOCACHE=/private/tmp/gocache-gitlab-4394-r4 go run ./cmd/connectorgen validate --connector gitlab` passes after the correction with `1 connector(s) checked, 0 findings`.
- [x] `jq empty internal/connectors/defs/gitlab/sources/gitlab-source-lane-matrix.json` passes.
- [x] `GOCACHE=/private/tmp/gocache-gitlab-4394-r4 go vet ./internal/connectors/defs/gitlab` passes.
- [x] `GOCACHE=/private/tmp/gocache-gitlab-4394-r4 go run ./cmd/agentcontractgen check` passes.
- [x] `git diff --check` passes before commit; the commit range check is repeated after commit.
- [x] Diff has only the bounded GitLab matrix/test/planning evidence paths; explicit diff checks show the source lock, descriptor, runtime artifacts, and Foundation Atlas unchanged.
- [ ] Fresh remote check and candidate-only push SHA recorded; no parent/main integration.

## Explicit non-claims

- No Conan POST request is executable: the retained source prose requires a
  JSON body but does not retain a typed request schema or media contract.
- No status-only semantic POST direct-read is executable: source mapping is
  preserved, but this repair creates neither a bodyless POST contract nor a
  direct-read operation/CLI declaration. Existing legacy method-partition
  write/reverse declarations are outside this slice.
- No ETL, reverse ETL, sync, binary, certification, or live-provider proof is
  added by this mapping-only repair.
