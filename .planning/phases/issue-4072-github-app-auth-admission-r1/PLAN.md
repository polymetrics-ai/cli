---
phase: issue-4072-github-app-auth-admission-r1
plan: "01"
type: tdd
wave: 1
depends_on: []
files_modified:
  - internal/connectors/engine/read.go
  - internal/connectors/engine/auth.go
  - internal/connectors/engine/hooks.go
  - internal/connectors/hooks/github/hooks.go
  - internal/connectors/hooks/github/hooks_test.go
  - internal/connectors/hooks/github/github_app_rate_admission_integration_test.go
  - .planning/phases/issue-4072-github-app-auth-admission-r1/TDD-LEDGER.md
autonomous: true
requirements:
  - ISSUE-4072
---

<objective>
Gate GitHub App installation-token minting through the engine's declared
rate-admission path.

Purpose: Close the pre-runtime network bypass that lets a GitHub App token POST
reach a provider before `require_shared` can refuse a missing/lost coordinator.

Output: A narrow engine-owned custom-auth request capability, wired to the
existing declared GitHub REST policy. Keep #3754's physical-route Requester
admission boundary, add deterministic local RED/GREEN evidence, and prove the
real shared coordinator tightens the budget across two test processes without
secret-bearing coordination data.
</objective>

<context>
@.planning/phases/issue-4072-github-app-auth-admission-r1/CONTEXT.md
@/Users/karthiksivadas/karthik-agent-workspace/data/cli-rate-3754-exhaustion-audit-r1/report.md
@internal/connectors/engine/read.go
@internal/connectors/engine/auth.go
@internal/connectors/engine/hooks.go
@internal/connectors/engine/rate_limit_runtime.go
@internal/connectors/connsdk/http.go
@internal/connectors/hooks/github/hooks.go
@internal/connectors/defs/github/rate_limits.json
</context>

<feature>
  <name>GitHub App token exchange uses declared shared rate admission</name>
  <files>internal/connectors/engine/read.go, internal/connectors/engine/auth.go, internal/connectors/engine/hooks.go, internal/connectors/hooks/github/hooks.go, internal/connectors/hooks/github/hooks_test.go, internal/connectors/hooks/github/github_app_rate_admission_integration_test.go</files>
  <behavior>
    Cases:
    - GitHub App plus `require_shared` and no coordinator: `NewRuntime` returns
      `*coordination.SharedRateLimitUnavailableError`; recording transport
      receives zero token POSTs.
    - An unreachable shared registry produces the same typed refusal and zero sends.
    - Two processes sharing a real Dragonfly one-request budget admit one
      physical `POST /app/installations/<escaped-id>/access_tokens` and block
      the other process, proving the shared state tightened.
    - The hook supplies the declaration path, but #3754 admits the actual
      escaped installation path at the Requester send boundary; JWT/key/token
      material is absent from coordination keys, errors, and test diagnostics.
    - Existing bearer auth, declared REST requests, process-local admission,
      GitHub write-hook admission, and `POST /graphql` exclusion retain their
      behavior.
  </behavior>
  <implementation>
    Build `rateLimitResolver` plus a base requester before custom auth. Add a
    private/narrow engine capability that resolves one declared method/path to a
    requester and performs one JSON request using an actual path and hook-owned
    headers/body. Deliver it only to auth hooks that opt into the capability;
    do not expose `BudgetCoordinator` or `Runtime`. The GitHub hook replaces
    `http.DefaultClient.Do` with that capability. Allow a non-CLI runtime HTTP
    client injection only if needed to make the local recording transport
    deterministic; keep it engine-owned and default-compatible.
  </implementation>
</feature>

<tasks>
  <task type="tdd">
    <name>RED: prove GitHub App token minting bypasses shared admission at the recovered base</name>
    <read_first>
      - `internal/connectors/engine/read.go`
      - `internal/connectors/engine/auth.go`
      - `internal/connectors/engine/hooks.go`
      - `internal/connectors/engine/rate_limit_runtime.go`
      - `internal/connectors/connsdk/http.go`
      - `internal/connectors/hooks/github/hooks.go`
      - `internal/connectors/hooks/github/hooks_test.go`
      - `internal/connectors/defs/github/rate_limits.json`
    </read_first>
    <action>
      Add focused local tests that load the real GitHub bundle, select
      `github_app`, copy the app-installation policy with
      `coordination=require_shared`, and use a generated test key plus an
      injected recording transport. Assert absent and unreachable shared
      registries return `*coordination.SharedRateLimitUnavailableError` before
      runtime construction completes and issue zero token POSTs. Add physical
      route/privacy assertions. Do not alter production source in this task.
    </action>
    <acceptance_criteria>
      - The focused test fails at the recovered base because the token POST
        reaches the recording transport before the resolver's shared refusal.
      - Failure is behavioral, not a compile/import/syntax failure.
      - The RED commit contains only test/planning evidence for #4072.
    </acceptance_criteria>
    <verify>`go test ./internal/connectors/engine ./internal/connectors/hooks/github -run 'TestGitHubAppAuthRateAdmission' -count=1` exits non-zero for the intended premature token send.</verify>
  </task>
  <task type="tdd">
    <name>GREEN: route the GitHub App token POST through declared rate admission</name>
    <read_first>
      - `internal/connectors/connectors.go`
      - `internal/connectors/engine/read.go`
      - `internal/connectors/engine/auth.go`
      - `internal/connectors/engine/hooks.go`
      - `internal/connectors/engine/rate_limit_runtime.go`
      - `internal/connectors/connsdk/http.go`
      - `internal/connectors/hooks/github/hooks.go`
      - `internal/connectors/engine/github_app_auth_admission_test.go`
    </read_first>
    <action>
      Construct the resolver and base requester before custom auth; supply a
      narrow declared-route JSON request capability to the GitHub App hook;
      resolve `POST /app/installations/{installation_id}/access_tokens` before
      sending its interpolated actual path. Preserve existing hook contracts
      where they do not opt into the capability, preserve headers/body/auth
      semantics, and route all token requests through `connsdk.Requester` so
      admission/finish behavior stays centralized.
    </action>
    <acceptance_criteria>
      - Missing/lost shared coordinator causes the typed refusal and zero
        physical token sends.
      - Granting coordinator observes exactly one decision and one finish for
        the single successful token send.
      - No source retains `http.DefaultClient.Do` in the GitHub App exchange.
      - The GitHub rate policy and GraphQL exclusion are unchanged.
    </acceptance_criteria>
    <verify>`go test ./internal/connectors/engine ./internal/connectors/hooks/github -run 'TestGitHubAppAuthRateAdmission|TestAuthenticatorGithubApp|TestRequireSharedGitHubWriteHook|TestGitHubDeclaredRateLimits' -count=1` exits 0.</verify>
  </task>
  <task type="standard">
    <name>Record focused evidence and pause at the Firstmate validation gate</name>
    <read_first>
      - `.planning/phases/issue-4072-github-app-auth-admission-r1/TDD-LEDGER.md`
      - `.planning/phases/issue-4072-github-app-auth-admission-r1/VERIFICATION.md`
      - `.planning/phases/issue-4072-github-app-auth-admission-r1/RUN-STATE.md`
    </read_first>
    <action>
      Record actual RED failure, GREEN command and result, commits, scope
      checks, and the private finish-plan snapshot hash. Do not start broad
      suites, race sweeps, no-mistakes, push, PR creation, CI, or merge. Append
      the required paused state after the focused GREEN commit and clean-tree
      check.
    </action>
    <acceptance_criteria>
      - Fresh 0/5 ledger contains causal RED and GREEN evidence.
      - Phase artifacts identify #4072, exact recovered base, and snapshot
        SHA256 `939f14f61defd993f8ad0335a5eb617d97083c9f73a6a75259d0e312ae8f408`.
      - Work stops at Firstmate's shared validation gate.
    </acceptance_criteria>
  </task>
</tasks>

<threat_model>
## Threat Model

| Threat | Mitigation | Verification |
|---|---|---|
| Token POST bypasses `require_shared` | Resolve declaration-aware requester before custom auth | no/lost coordinator zero-send tests |
| Hook gains generic egress or coordinator authority | Narrow request capability only; no coordinator/runtime export | source/API review |
| Secret reaches coordination evidence | Use config-derived non-secret scope only and opaque reservation data | fake coordinator inspects captured batch/observation |
| GraphQL accounting changes accidentally | Do not edit policy declaration; retain existing exclusion test | focused GitHub policy test |
</threat_model>

<verification>
Focused now:

1. `go test ./internal/connectors/engine ./internal/connectors/hooks/github -run 'TestGitHubAppAuthRateAdmission' -count=1` — causal RED, then GREEN.
2. `go test ./internal/connectors/engine ./internal/connectors/hooks/github -run 'TestGitHubAppAuthRateAdmission|TestAuthenticatorGithubApp|TestRequireSharedGitHubWriteHook|TestGitHubDeclaredRateLimits' -count=1` — focused regression matrix after GREEN.

Deferred until Firstmate releases the shared validation lane: focused race
coverage, vet/lint, generator/docs/help/website checks, issue guard, GSD
verify-work, GSD code review, no-mistakes without `--yes`, push/PR topology.
</verification>

<success_criteria>
- RED and GREEN commits occur in order on `fix/4072-github-app-auth-rate-admission` from `da8a8ff07aaf00e5c7965cd4d1d3c7252017d785`.
- Token minting uses one declared rate-admitted requester send, without raw coordinator exposure.
- No/lost shared coordinator prevents token transport; granting coordinator sees one Decide/Finish per send.
- Existing GitHub admission behavior remains intact and no credential material enters coordination evidence.
- The focused-stage handoff pauses cleanly for Firstmate, without no-mistakes, push, PR, merge, or CI.
</success_criteria>
