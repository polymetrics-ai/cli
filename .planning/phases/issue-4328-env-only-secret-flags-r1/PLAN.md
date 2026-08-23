# Issue #4328: declaration-owned env-only secret flags

## Task Delivery Header

- Issue: Refs #4328 — fix(connectorgen): keep OpenAPI write secrets out of reverse-ETL CLI flags
- Base branch: `main`
- Merges into: `main`
- Delivery: Pull request open against `main`, with the local full `make verify` gate green and the API-reported PR base read back as `main`.
- Working branch: `fm/cli-env-only-secret-flag-generalization-r1`
- Task: Generalize `env_only` generation so every declaration-owned request secret is environment-only independent of protocol, intent, flag type, or mapping depth; preserve established GraphQL behavior; prove CircleCI webhook signing-secret is protected; and report the full connector-definition blast-radius count.
- Verification: Red then green behavioral tests through the actual `connectorgen validate` path against definitions, a deterministic all-definition sweep, GitHub source/descriptor byte-and-SHA measurements, `go test -timeout 20m ./cmd/connectorgen/...`, `make verify`, `git diff --check`, and a direct PR base API read-back.

## Evidence table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A declared REST write secret is generated `env_only`. | live | A real definition passes the validator only when its resulting CLI flag has `env_only: true`; unchanged logic leaves it false. |
| A declared GraphQL mutation secret remains `env_only`. | live | Existing real GraphQL declaration validates with `env_only: true` after the generalization. |
| A non-secret remains an ordinary CLI flag. | live | A real non-sensitive declaration's resulting flag remains `env_only: false`. |
| A secret outside a single top-level body mapping becomes `env_only`. | live | A real nested/path/query/header declaration with declared sensitivity validates only with `env_only: true`. |
| CircleCI webhook signing-secret is not emitted as an ordinary flag. | live | CircleCI's generated surface has the relevant `--signing-secret` flag marked `env_only: true`. |
| The repository-wide blast radius is known. | live | A deterministic all-definition test/sweep reports the count that would have been unprotected under the old GraphQL/top-level/json/required/direct-write predicate. |
| GitHub parity stays byte-identical. | live | Measured source and descriptor byte counts plus SHA-256 exactly match the launch brief's values. |

## Scope and exclusions

- Production ownership: `cmd/connectorgen/validate.go` plus the adjacent `cmd/connectorgen/sourceprojection.go` materializer and their tests. `sourceprojection.go` is required because validation alone cannot change a generated flag; `cmd/connectorgen/sourceimport.go` is an active lane and must not change.
- Connector definitions are read for validation and sweep evidence only. Do not edit `internal/connectors/defs/github/rate_limits.json`.
- No raw secret values, credential material, reverse-ETL execution, generic HTTP write, generic SQL write, or generic shell surface is added.
- CLI docs/help/website parity: no user-facing flag spelling, help wording, command, manual, or website surface changes. The existing `env_only` semantics change how the already generated declaration is safely supplied, so the affected generated HubSpot skill must be refreshed to describe those three flags as `env-only`; runtime help, manual, and website changes remain not applicable.

## TDD slices

1. **Discovery and red proof.** Locate the declaration-owned request-sensitivity representation (the request-side analogue to `OutputSecretFields`) and map the old predicate's candidates across the real definition tree. Add real-definition tests first for REST, GraphQL, non-secret, non-top-level mapping, and CircleCI; run the focused package test to record a failure under the current predicate.
2. **Minimal generalization.** Replace the protocol/shape gate with the declaration's sensitivity result while retaining `env_only` assignment and all current behavior. Do not alter importer code.
3. **Blast-radius and parity proof.** Run the all-definition sweep to report newly protected flags and separately measure GitHub source/descriptor bytes and SHA-256. Confirm no forbidden GitHub rate-limit file changed.
4. **Full verification and review.** Run `make verify` without parallel full suites, complete `verify-work` and code review inline, disposition findings, commit/push, open the direct PR, and read the GitHub API base back.

## Lifecycle record

- Completed before planning: `scripts/gsd doctor`; `scripts/gsd sources` for `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`; all five generated prompts; `go run ./cmd/agentcontractgen check`.
- Manual GSD fallback: execute the generated workflow inline because the direct-PR task disallows role spawning; retain full discussion, plan, red/green, verification, and review evidence in this phase directory.
- Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-graphql`, and `golang-lint`.

## Commit checkpoints

1. Planning artifacts.
2. Red real-definition regression evidence.
3. Green generalized validator plus blast-radius evidence.
4. Verification/review evidence and fixes, then direct PR.
