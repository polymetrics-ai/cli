Refs #4328

## Intent

Generalize declaration-owned `env_only` protection so request secrets cannot be emitted as ordinary CLI flags based on a GraphQL-specific predicate.

## What changed

- Detect request secrets from schema `x-secret` markers, including secret-bearing nested JSON fields, and exact request-side sensitive-policy declarations.
- Require `env_only` for matching REST, GraphQL, write-record, parameter, nested-body, and source-projected flags without protocol, intent, type, or mapping-depth gates.
- Preserve source `x-secret` metadata and generate `env_only` for CircleCI-shaped webhook signing fields.
- Correct HubSpot's three missed REST secret flags and refresh their generated manuals/skills plus the certification subject.

## Blast radius

The full sweep found **3** previously unprotected flags across **552** connector definitions:

1. `hubspot reverse delete-oauth-v1-refresh-tokens-token-archive --token`
2. `hubspot reverse post-oauth-v1-token-create --client-secret`
3. `hubspot reverse post-oauth-v1-token-create --refresh-token`

## TDD and GSD evidence

- Red evidence is recorded in `.planning/phases/issue-4328-env-only-secret-flags-r1/TDD-LEDGER.md`.
- Green coverage exercises actual GitHub and HubSpot definitions, REST/nested JSON/non-secret cases, and CircleCI webhook projection.
- The direct-PR brief forbids role spawning, so the mandatory GSD lifecycle ran inline; its plan, ledger, verification, and review records are in `.planning/phases/issue-4328-env-only-secret-flags-r1/`.
- Required Go skills: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-graphql`, and `golang-lint`.

## Generated parity and safety

- Regenerated `docs/connectors/hubspot/{MANUAL,SKILL}.md`, `docs/skills/pm-hubspot/SKILL.md`, and `internal/connectors/certifications/current-subject.json` through repository-owned generators.
- GitHub source lock measured **3,420,025** bytes / `281b1cfcc67eb63e19ef83daf06197bf3d3b23db0b6bc9b73e02fc18ee278fb6`; descriptor measured **43,354,021** bytes / `d1978c0c6fd0eb66e9fcd4d78d637864a6e486f558aaad1e51550bc43758b899`.
- `cmd/connectorgen/sourceimport.go` and `internal/connectors/defs/github/rate_limits.json` are untouched.

## Verification

- `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestValidate_CLISurfaceEnvOnlyFlagRequiresDeclaredSecretGraphQLContract|TestValidate_CLISurfaceEnvOnlyFlagAcceptsDeclaredRESTSecretRegardlessOfFlagShape|TestValidate_RealGitHubSecretCommandRequiresEnvOnly|TestValidate_RealHubSpotRequestSecretsRequireEnvOnly|TestSourceProjectionMarksDeclaredCircleCIWebhookSecretsEnvOnly)$'` — pass
- `go test -count=1 -timeout 20m ./cmd/connectorgen/...` — pass
- `make verify` — pass, including lint, full test suite, docs/smoke, surface synchronization, GitHub parity, certification, boundary, canon, and release checks
- `git diff --check` — pass

## Automated review

Route: `claude_auto`, pending GitHub's automatic review after this non-draft PR opens. No fallback route requested.
