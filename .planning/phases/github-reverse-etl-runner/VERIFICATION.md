# Verification

## Passed

- `jq . internal/connectors/defs/github/cli_surface.json`
- `go run ./cmd/connectorgen validate internal/connectors/defs --json`
- `go test ./internal/connectors/commandrunner ./internal/app ./internal/cli -run 'TestBuildWriteCommand|TestRunReverseETLCommandRemainsNonExecutableInGenericRunner|TestRunReverseETLRejectsMissingWriteAndUnsupportedFlagMapping|TestGitHubCommandSurfacePlansReverseETLCommand|TestGitHubCommandWriteUsesReversePlanApproval|TestRunReverseETLRejectsApprovalTokenReplay|TestRunReverseETLRejectsPlanHashMismatchWhenRowsChange'`
- `go test ./cmd/connectorgen ./internal/connectors/engine`
- `go build ./cmd/pm`
- `PATH=/Users/karthiksivadas/.cache/codex-runtimes/codex-primary-runtime/dependencies/node/bin:/Users/karthiksivadas/.cache/codex-runtimes/codex-primary-runtime/dependencies/bin:$PATH /Users/karthiksivadas/.cache/codex-runtimes/codex-primary-runtime/dependencies/bin/pnpm typecheck`
- `go run ./cmd/pm docs validate --dir docs/cli --connectors-dir docs/connectors --website-dir website/content/docs`
- `git diff --check`

## Notes

- Full `internal/cli` focused runs are slow in this environment because runtime failure-path tests
  wait on local service connection timeouts, but the targeted GitHub command-write tests pass.
- Website data was regenerated with bundled Node because the system `npm` CLI is too old for its own
  runtime.
