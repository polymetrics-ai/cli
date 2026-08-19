---
phase: issue-4072-github-app-auth-admission-r1
verified: 2026-08-14
status: verified
score: 5/5
---

# Issue #4072 verification

| Must-have | Result | Evidence |
|---|---|---|
| Missing coordinator refuses before token transport | pass | typed unavailable error and zero recording-transport sends |
| Unreachable coordinator refuses before token transport | pass | refused local endpoint returns `SharedRateLimitCoordinatorUnreachable`, zero sends |
| Auth token send uses #3754 boundary | pass | GitHub hook calls engine `DoJSON`; engine calls `Requester.Do` with escaped actual path |
| Shared state tightens across processes | pass | real Dragonfly: one mint, one deadline, one physical POST |
| Regression and static checks pass | pass | focused/full package tests, race, vet, build, and required make gates |

## Commands run

| Command | Result |
|---|---|
| `go test -timeout 20m ./internal/connectors/engine ./internal/connectors/hooks/github -run 'TestGitHubAppAuthRateAdmission|TestAuthenticatorGithubApp|TestRequireSharedGitHubWriteHook|TestGitHubDeclaredRateLimits' -count=1` | pass |
| `go test -timeout 20m ./internal/connectors/engine ./internal/connectors/hooks/github` | pass |
| `go test -timeout 20m -race ./internal/connectors/engine ./internal/connectors/hooks/github` | pass |
| `go vet ./internal/connectors/engine ./internal/connectors/hooks/github` | pass |
| `POLYMETRICS_COORDINATION_INTEGRATION=1 go test -timeout 20m -tags=coordinationintegration ./internal/connectors/hooks/github -run '^TestGitHubAppAuthRateAdmissionSharedBudgetAcrossProcesses$' -count=1 -v` | pass |
| `go build ./cmd/pm` | pass |
| `make tidy-check lint docs-check smoke-no-build agent-contract-check connectorgen-validate connectorgen-surface-sync connector-boundary release-workflow-check` (each gate run independently) | pass |

The live test asserts coordinator state through its observable consequence,
not merely lack of an error: two processes contend for one budget and only one
physical installation-token request can reach the fixture.
