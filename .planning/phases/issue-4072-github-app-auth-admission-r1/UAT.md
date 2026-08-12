---
phase: "issue-4072-github-app-auth-admission-r1"
status: passed
mode: inline_manual_auto
verified_at: 2026-08-12
---

# Verify-work evidence — Issue #4072

This is a Go-only transport foundation with no end-user interactive handoff.
`scripts/gsd prompt verify-work issue-4072-github-app-auth-admission-r1 --auto`
was resolved. The named issue phase is not a numeric GSD roadmap phase, and the
canonical single-worker contract prohibits spawning a verifier, so this records
the permitted inline automatic fallback.

| Deliverable | Acceptance evidence | Result |
|---|---|---|
| Missing coordinator fails closed before runtime completion | `TestGitHubAppAuthRateAdmissionRequireSharedRefusesBeforeTokenSend` observes typed `shared_coordinator_unavailable` and zero recording-transport sends | pass |
| Lost coordinator fails closed before runtime completion | `TestGitHubAppAuthRateAdmissionLostCoordinatorRefusesBeforeTokenSend` observes the same typed error and zero sends | pass |
| Granted token mint has one physical lifecycle | `TestGitHubAppAuthRateAdmissionGrantingCoordinatorUsesSingleLifecycle` observes one reservation, one token POST, and one finish | pass |
| Declared route is distinct from actual request path | The granting test captures declared `POST /app/installations/{installation_id}/access_tokens` and the escaped installation path sent on the wire | pass |
| Secret boundary is preserved | Secret-blind fakes record only method/path and coordination facts; the privacy assertions reject JWT, private key, and installation-token leakage | pass |
| Existing GitHub behavior remains intact | Bounded engine/GitHub tests cover bearer auth, write-hook admission, ordinary declared requests, process-local admission, and GraphQL exclusion | pass |

## Commands observed

```text
GOMAXPROCS=2 go test -p 1 ./internal/connectors/engine ./internal/connectors/hooks/github -run '^(TestGitHubAppAuthRateAdmission|TestAuthenticatorGithubApp|TestRequireSharedGitHubWriteHook|TestGitHubWriteHook|TestGitHubDeclaredRateLimits|TestGitHubRateLimitAdmission)' -count=1
GOMAXPROCS=2 go test -p 1 -race ./internal/connectors/engine ./internal/connectors/hooks/github -run '^(TestGitHubAppAuthRateAdmission|TestAuthenticatorGithubApp|TestRequireSharedGitHubWriteHook|TestGitHubWriteHook|TestGitHubDeclaredRateLimits|TestGitHubRateLimitAdmission)' -count=1
go vet ./internal/connectors/engine ./internal/connectors/hooks/github
```

All commands passed. They use local fakes only; no GitHub credentials or
provider mutations were used. Repository-wide validation and delivery actions
remain deferred to Firstmate's shared #4071 gate.
