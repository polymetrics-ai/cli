# Issue #4072 Focused GREEN Summary

**Issue:** #4072 — `fix(engine): gate GitHub App token minting through shared rate admission`

**Parent:** #3754
**Branch:** `fix/4072-github-app-auth-rate-admission`
**Recovered base:** `da8a8ff07aaf00e5c7965cd4d1d3c7252017d785`
**Correction ledger:** fresh 0/5
**Canonical private finish-plan snapshot SHA256:**
`939f14f61defd993f8ad0335a5aeb617d97083c9f73a6a75259d0e312ae8f408`

## Result

GitHub App installation-token minting now goes through an engine-owned,
declared-route JSON capability. `newRuntime` constructs the base requester and
rate resolver before custom authentication; the GitHub hook sends its actual
installation path through declared `POST /app/installations/{installation_id}/access_tokens`.
The raw `http.DefaultClient.Do` bypass is removed.

The capability does not expose a coordinator, a runtime, an arbitrary URL, or a generic user-facing HTTP writer.
Per-request JWT headers are copied only into a requester clone, not retained in runtime or coordination state.

## TDD Record

- `d0777bcc7` — manual inline GSD discuss/context/plan checkpoint.
- `9a44c9163` — causal RED: no coordinator still sent a physical token POST.
- `3f20bf7ba` — expanded RED: no/lost/grant lifecycle and privacy cases.
- `3f83bf3af` — GREEN: token exchange uses declaration-aware admission.

The causal RED failed for the intended behavior: it observed one physical `http.DefaultClient` token send before `NewRuntime` returned a shared coordinator refusal.
GREEN passes the no-coordinator, lost-coordinator, granting-coordinator, declared-vs-actual-path, secret-boundary, process-local, GitHub write-hook, bearer/ordinary request, and GraphQL-exclusion coverage.

## Focused Verification

- `go test ./internal/connectors/hooks/github -run '^TestGitHubAppAuthRateAdmission' -count=1`
- `go test ./internal/connectors/engine ./internal/connectors/hooks/github -run 'TestGitHubAppAuthRateAdmission|TestAuthenticatorGithubApp|TestRequireSharedGitHubWriteHook|TestGitHubDeclaredRateLimits' -count=1`
- `go test ./internal/connectors/hooks/github -count=1`

All passed on 2026-08-12. Only local fakes were used; no provider credentials or provider mutation occurred.

Resumed bounded verification also passed with `GOMAXPROCS=2`, one package at a
time: the focused engine/GitHub behavior matrix, the same matrix under
`-race`, and `go vet ./internal/connectors/engine ./internal/connectors/hooks/github`.
Inline automatic verify-work and inline deep code-review evidence are recorded
in `UAT.md` and `REVIEW.md`.

## Deferred Gate

This focused checkpoint deliberately stops before repository-wide lint,
generators, docs/help/website parity, issue guard, no-mistakes, push, PR
creation, CI, or merge. Firstmate owns the shared #4071 validation-gate release
and the eventual parent PR route.
