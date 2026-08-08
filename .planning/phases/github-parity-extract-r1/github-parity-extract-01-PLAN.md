---
phase: github-parity-extract-r1
plan: "01"
type: tdd
wave: 1
depends_on: []
files_modified:
  - internal/connectors/defs/github/rate_limits.json
  - internal/connectors/defs/github/spec.json
  - internal/connectors/defs/github/docs.md
  - internal/connectors/defs/defs.go
  - internal/connectors/engine/github_rate_limits_test.go
  - internal/connectors/engine/bundle_test.go
  - .planning/phases/github-parity-extract-r1/TDD-LEDGER.md
autonomous: true
requirements: []
must_haves:
  truths:
    - "D-05: GitHub uses the existing Runtime.RequesterFor admission/observation path; this plan adds no second limiter."
    - "D-06: Provider-cited policy selection uses selector.auth_types and an opaque, non-secret coordination scope, never a credential."
    - "D-07: The declared budget is deliberately exceeded through pm and independently checked against GitHub headroom in the wave-3 live proof."
  artifacts:
    - path: "internal/connectors/defs/github/rate_limits.json"
      provides: "Provider-cited declared GitHub rate policy."
    - path: "internal/connectors/engine/github_rate_limits_test.go"
      provides: "Direct requester attachment and scope coverage."
  key_links:
    - from: "internal/connectors/defs/github/rate_limits.json"
      to: "internal/connectors/engine/rate_limit_runtime.go"
      via: "embedded bundle policy resolved by Runtime.RequesterFor"
      pattern: "selector.auth_types"
---

<objective>
Declare GitHub's real provider-cited rate budgets through the existing engine and prove a declared
GitHub policy attaches to the existing requester for the intended auth flows without using a raw
credential as scope.

Purpose: make the live sweep exercise the existing limiter, not a parallel mechanism.
Output: embedded GitHub `rate_limits.json`, documented non-secret scope inputs, and red/green tests.
</objective>

<must_haves>
<truths>
- D-05: GitHub uses the existing `Runtime.RequesterFor` admission/observation path; this plan adds no second limiter.
- D-06: Provider-cited policy selection uses `selector.auth_types` and an opaque, non-secret coordination scope, never a credential.
- D-07: The declared budget is deliberately exceeded through `pm` and independently checked against GitHub headroom in the wave-3 live proof.
</truths>
</must_haves>

<threat_model>
| Threat | Mitigation | Verification |
|---|---|---|
| Rate key leaks a credential | Scope only non-secret spec fields and derive opaque key through `CoordinationIdentity.RateScopeKey`. | Tests reject secret scope fields and assert no declaration references `token`. |
| Auth policy silently does not attach | Exercise `Runtime.RequesterFor` for every declared GitHub auth-type branch. | Test requires admission/observer on matching requests and absent attachment for non-matches. |
| A provider limit is invented or stale | Cite GitHub's official REST rate-limit document with `retrieved_at: 2026-08-08`. | Loader/fixture test asserts exact source and budgets. |
</threat_model>

<feature>
  <name>GitHub declared rate-limit policy</name>
  <files>internal/connectors/engine/github_rate_limits_test.go, internal/connectors/defs/github/rate_limits.json, internal/connectors/defs/github/spec.json, internal/connectors/defs/defs.go</files>
  <behavior>
    Before implementation, loading the embedded GitHub bundle must fail the new test because it
    has no declared policy. After implementation, matching authenticated/user, GitHub App
    installation, and public flows acquire the existing requester admission/observation hooks with
    their provider-cited request budgets; unknown or unmatched flows do not manufacture a scope.
  </behavior>
  <implementation>
    Add only declaration/config/embed wiring. Do not add a second limiter, requester path, registry,
    or credential-derived key. Add exact non-secret identity fields required by the policy and
    generated docs, then regenerate only GitHub-derived artifacts.
  </implementation>
</feature>

<tasks>
  <task type="tdd">
    <name>RED — specify GitHub declaration and scope behavior</name>
    <read_first>
      - internal/connectors/connsdk/rate_limits.go
      - internal/connectors/engine/rate_limit_runtime.go
      - internal/connectors/engine/rate_limit_runtime_test.go
      - internal/connectors/defs/github/spec.json
      - internal/connectors/defs/defs.go
    </read_first>
    <action>
      Add a GitHub-focused engine test that loads `defs.FS` and requires
      `RateLimits.State == declared`, provider source URL
      `https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api`,
      `retrieved_at == 2026-08-08`, auth-type selector coverage, a non-secret account/IP or
      installation scope, and requester `Admission` plus `Observer` for matching paths. Run it
      before creating `rate_limits.json` or changing the embed directive; record the genuine RED
      output in `TDD-LEDGER.md`.
    </action>
    <acceptance_criteria>
      - `go test ./internal/connectors/engine/ -run TestGitHubDeclaredRateLimits` fails because GitHub has no embedded declared policy.
      - The failure is behavioral (missing declaration/attachment), not a compile or syntax error.
      - `TDD-LEDGER.md` contains the command and RED result before implementation.
    </acceptance_criteria>
  </task>
  <task type="tdd">
    <name>GREEN — declare and embed provider-cited GitHub policies</name>
    <read_first>
      - internal/connectors/engine/rate_limits.go
      - internal/connectors/coordination_identity.go
      - internal/connectors/defs/github/spec.json
      - internal/connectors/defs/github/docs.md
      - internal/connectors/defs/defs.go
      - .planning/phases/github-parity-extract-r1/github-parity-extract-RESEARCH.md
    </read_first>
    <action>
      Add `internal/connectors/defs/github/rate_limits.json` in schema version 1 with `state`
      `declared`, the official source and date, separate `selector.auth_types` policies for the
      documented authenticated-user, unauthenticated, and GitHub App installation REST budgets, and
      only representable request budgets. Add the smallest explicit non-secret scope inputs needed
      to avoid using `token` or `owner` as an assumed authenticated identity. Add
      `*/rate_limits.json` to the production embed directive. Update GitHub definition-owned docs;
      regenerate rather than hand-edit help/catalog artifacts.
    </action>
    <acceptance_criteria>
      - The RED test passes and shows matching `RequesterFor` values have both `Admission` and `Observer`.
      - `go run ./cmd/connectorgen validate internal/connectors/defs` exits 0.
      - `go run ./cmd/connectorgen surface-sync --check` exits 0.
      - A declaration validator test proves no scope subject_config is secret and no source URL carries credentials or query parameters.
      - No new limiter implementation or raw-credential registry key appears in the diff.
    </acceptance_criteria>
  </task>
  <task type="tdd">
    <name>REFACTOR — keep auth, docs, and policy ownership singular</name>
    <read_first>
      - internal/connectors/engine/github_rate_limits_test.go
      - internal/connectors/defs/github/rate_limits.json
      - internal/connectors/defs/github/docs.md
    </read_first>
    <action>
      Remove only duplication exposed by the GREEN implementation, retain explicit provider and
      scope names, rerun targeted tests, and update the ledger with the refactor result or `none`.
    </action>
    <acceptance_criteria>
      - Targeted engine tests remain green.
      - Any refactor commit preserves the passing behavior and records it in `TDD-LEDGER.md`.
    </acceptance_criteria>
  </task>
</tasks>

<verification>
- `go test -timeout 20m ./internal/connectors/engine/ -run 'GitHubDeclaredRateLimits|RateLimit'`
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go run ./cmd/connectorgen surface-sync --check`
- `go build ./cmd/pm`
</verification>

<success_criteria>
- GitHub has an embedded declared, provider-cited policy selected by auth type.
- Every policy scope is explicitly non-secret and derived through the existing coordination identity.
- RED and GREEN evidence is committed before the live sweep begins.
</success_criteria>
