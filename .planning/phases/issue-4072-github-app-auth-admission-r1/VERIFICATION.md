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
| 2 | Lost shared coordinator refuses before token transport | local fake coordinator + typed error | pending |
| 3 | One granted token POST has one Decide and one Finish | local recording coordinator/transport | pending |
| 4 | Admission matches declared token route while sending actual path | captured declaration and request path | pending |
| 5 | Credentials/tokens do not enter coordinator evidence | captured reservation/observation/error inspection | pending |

## Required Focused Checks

| Command | Purpose | Status |
|---|---|---|
| `go test ./internal/connectors/engine ./internal/connectors/hooks/github -run 'TestGitHubAppAuthRateAdmission' -count=1` | causal RED then implementation proof | pending |
| `go test ./internal/connectors/engine ./internal/connectors/hooks/github -run 'TestGitHubAppAuthRateAdmission|TestAuthenticatorGithubApp|TestRequireSharedGitHubWriteHook|TestGitHubDeclaredRateLimits' -count=1` | narrow behavior/regression matrix | pending |

## Deferred by Firstmate Shared Validation Gate

Focused race coverage, vet/lint, generator/docs/help/website parity checks,
issue guard, `verify-work`, code review, no-mistakes, full CI, push, PR, and
parent-route decision are deferred. This is not a passing verification report.
