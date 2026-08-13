# Issue #4072 Focused GREEN Summary

**Issue:** #4072 — `fix(engine): gate GitHub App token minting through shared rate admission`

**Parent:** #3754
**Branch:** `fix/4072-github-app-auth-rate-admission`
**Recovered base:** `da8a8ff07aaf00e5c7965cd4d1d3c7252017d785`
**Correction ledger:** 1/5 (configured-linter correction resolved; generated
artifact synchronization does not consume correction 2/5)
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

## Generated capability-matrix synchronization

At clean `31be96f…`, the canonical
`go run ./cmd/connectorgen certification-matrix --check` was the causal
inherited RED: it failed because `capability-matrix.json` had stale generated
source locations. The identical tracked artifact and check failure exist at
recovered base `da8a8ff…`; no #4072 auth behavior caused it.

GREEN ran only the canonical generator and its follow-up check. It reports
556 connectors with no capability-complete/certified status change. The
resulting artifact SHA-256 is
`e63b906cb640b8fb4fc8fd46c1076b77b7dbced7889919d60527f9b4335d520a`;
the diff is exactly six `discovery_source` line updates for the
`capability:*` entries. Removing `discovery_source` fields yields the same
base/generated semantic SHA-256
`bc5d14758c26755d83a9dc4dcbb715da31d95f67de38e352bc652b752c0819bc`.

#4026/#4034 is cited only as the established generator precedent. Its
PostgreSQL ancestry is not imported; the generator source hash is unchanged
at `bba4dea056e18ecc1231fe68cea8321dfc8a53d2b6ce58ac32ca02fa816a7bbf`.
