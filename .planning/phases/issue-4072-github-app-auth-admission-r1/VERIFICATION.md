---
phase: issue-4072-github-app-auth-admission-r1
verified: 2026-08-12T00:00:00Z
status: focused_verification_complete
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
| `GOMAXPROCS=2 go test -p 1 ./internal/connectors/engine ./internal/connectors/hooks/github -run '^(TestGitHubAppAuthRateAdmission|TestAuthenticatorGithubApp|TestRequireSharedGitHubWriteHook|TestGitHubWriteHook|TestGitHubDeclaredRateLimits|TestGitHubRateLimitAdmission)' -count=1` | bounded behavior and regression matrix | pass |
| `GOMAXPROCS=2 go test -p 1 -race ./internal/connectors/engine ./internal/connectors/hooks/github -run '^(TestGitHubAppAuthRateAdmission|TestAuthenticatorGithubApp|TestRequireSharedGitHubWriteHook|TestGitHubWriteHook|TestGitHubDeclaredRateLimits|TestGitHubRateLimitAdmission)' -count=1` | bounded concurrency coverage for the changed engine and hook paths | pass |
| `go vet ./internal/connectors/engine ./internal/connectors/hooks/github` | package-scoped static analysis | pass |

## Completed GSD Evidence

`scripts/gsd prompt verify-work issue-4072-github-app-auth-admission-r1 --auto`
was resolved before this check. The named issue phase is outside the numeric
roadmap, and the single-worker contract forbids role spawning, so verify-work
is recorded as an inline automatic fallback in `UAT.md`. The completed inline
deep review is recorded in `REVIEW.md`.

## Deferred by Firstmate Shared Validation Gate

Repository-wide lint, generator/docs/help/website parity checks, issue guard,
no-mistakes, full CI, push, PR, and parent-route decision remain deferred to
Firstmate's shared #4071 validation gate. This is a focused verification report,
not a substitute for those release gates.
