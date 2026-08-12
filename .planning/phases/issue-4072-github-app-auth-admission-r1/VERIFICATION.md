---
phase: issue-4072-github-app-auth-admission-r1
verified: 2026-08-12T00:00:00Z
status: focused_green_passed
score: 5/5 must-haves verified
---

# Issue #4072: GitHub App auth admission - Verification Plan

**Phase goal:** GitHub App installation-token minting is admitted by the same
declared shared/process-local rate boundary as ordinary GitHub REST requests.

## Observable Truths

| # | Truth | Focused evidence | Status |
|---|---|---|---|
| 1 | Missing shared coordinator refuses before token transport | local recording transport + typed error | pass |
| 2 | Lost shared coordinator refuses before token transport | local fake coordinator + typed error | pass |
| 3 | One granted token POST has one Decide and one Finish | local recording coordinator/transport | pass |
| 4 | Admission matches declared token route while sending actual path | captured declaration and request path | pass |
| 5 | Credentials/tokens do not enter coordinator evidence | captured reservation/observation/error inspection | pass |

## Required Focused Checks

| Command | Purpose | Status |
|---|---|---|
| `go test ./internal/connectors/hooks/github -run '^TestGitHubAppAuthRateAdmission' -count=1` | causal GREEN proof using secret-blind recording transport | pass |
| `go test ./internal/connectors/engine ./internal/connectors/hooks/github -run 'TestGitHubAppAuthRateAdmission|TestAuthenticatorGithubApp|TestRequireSharedGitHubWriteHook|TestGitHubDeclaredRateLimits' -count=1` | narrow behavior/regression matrix | pass |
| `go test ./internal/connectors/hooks/github -count=1` | complete GitHub hook package regression check | pass |

## Deferred by Firstmate Shared Validation Gate

Focused race coverage, vet/lint, generator/docs/help/website parity checks,
issue guard, `verify-work`, code review, no-mistakes, full CI, push, PR, and
parent-route decision are deferred. This is not a passing verification report.
