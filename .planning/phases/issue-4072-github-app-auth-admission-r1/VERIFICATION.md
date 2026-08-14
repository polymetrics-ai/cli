---
phase: issue-4072-github-app-auth-admission-r1
verified: 2026-08-12T00:00:00Z
status: pending_focused_green
score: 0/5 must-haves verified
---

# Issue #4072: GitHub App auth admission - Verification Plan

**Phase goal:** GitHub App installation-token minting is admitted by the same
declared shared/process-local rate boundary as ordinary GitHub REST requests.

## Observable Truths

| # | Truth | Focused evidence | Status |
|---|---|---|---|
| 1 | Missing shared coordinator refuses before token transport | local recording transport + typed error | pending |
| 2 | Unreachable shared coordinator refuses before token transport | local unavailable registry + typed error | pending |
| 3 | Real coordinator budget tightens across two processes | Dragonfly-backed test: one token POST and one exhausted-budget timeout | pending |
| 4 | Admission matches the physical token route at Requester send | captured fixture request path and shared-budget outcome | pending |
| 5 | Credentials/tokens do not enter coordination evidence | opaque shared-key/error inspection | pending |

## Required Focused Checks

| Command | Purpose | Status |
|---|---|---|
| `go test ./internal/connectors/engine ./internal/connectors/hooks/github -run 'TestGitHubAppAuthRateAdmission' -count=1` | causal RED then implementation proof | pending |
| `go test ./internal/connectors/engine ./internal/connectors/hooks/github -run 'TestGitHubAppAuthRateAdmission|TestAuthenticatorGithubApp|TestRequireSharedGitHubWriteHook|TestGitHubDeclaredRateLimits' -count=1` | narrow behavior/regression matrix | pending |
| `POLYMETRICS_COORDINATION_INTEGRATION=1 go test -tags=coordinationintegration ./internal/connectors/hooks/github -run TestGitHubAppAuthRateAdmissionSharedBudgetAcrossProcesses -count=1` | real-coordinator, two-process budget-tightening proof | pending |

## Deferred by Firstmate Shared Validation Gate

Focused race coverage, vet/lint, generator/docs/help/website parity checks,
issue guard, `verify-work`, code review, no-mistakes, full CI, push, PR, and
parent-route decision are deferred. This is not a passing verification report.
