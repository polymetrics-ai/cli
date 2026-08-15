# Verification — Issue #4176

## Checklist

- [x] Worktree isolation and integration-base ancestry proven before edits.
- [x] Task delivery header and live/fake evidence table created before production edits.
- [x] GSD lifecycle commands resolved; inline fallback recorded.
- [x] Red tests recorded: no `schedule_fire` stage/command on the dispatch base.
- [x] Green implementation and focused package tests pass.
- [x] Shipped CLI construction-path proof passes: `TestCertifyCLISingleConnectorPassExitsZero` (30.602s).
- [x] Affected CLI package passes: `go test -timeout 20m ./internal/cli -count=1` (492.716s).
- [x] Static, build, derived, and workflow checks pass.
- [x] Inline verify-work and code review complete.
- [ ] Pull request open and API-reported base is `integration/4015-mvp-flat-r1`.

## CLI parity disposition

No command, flag, help text, manual, generated artifact, or website page is intentionally changed. The verification record will show that the existing help/docs remain applicable and that report output is covered by the certification tests.

## Automated evidence

- `go test -timeout 20m ./internal/connectors/certify -run 'Test(FullCertificationStageSetIsStrictSuperset|GlueStagesScheduleFire|GlueStagesAgainstSample)' -count=1 -v` passed after the red run showed no `schedule_fire` stage or command.
- `go test -timeout 20m ./internal/connectors/certify -count=1` passed in 13.801s.
- `go test -timeout 20m ./internal/cli -run '^TestCertifyCLISingleConnectorPassExitsZero$' -count=1 -v` passed in 30.602s; the whole `internal/cli` package passed in 492.716s.
- `gofmt`, `git diff --check`, `go vet ./internal/connectors/certify ./internal/cli`, `go build ./cmd/pm`, `go run ./cmd/agentcontractgen check`, and `scripts/verify-gsd-workflow` passed.
- `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connector-boundary`, and `make release-workflow-check` passed.
- Help/docs parity: `pm help connectors`, `pm connectors`, and `pm connectors certify --help` exit successfully; docs and website already document `connectors certify --full` and `schedule fire`, and no user-facing surface changed.
