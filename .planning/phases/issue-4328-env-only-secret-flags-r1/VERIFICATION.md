# Verification: Issue #4328

## Required checks

- [x] Focused red real-definition validator regression test, recorded in `TDD-LEDGER.md`.
- [x] Focused green targeted validator/projection tests.
- [x] Deterministic all-definition env-only blast-radius sweep with the precise newly protected count recorded below.
- [x] GitHub source lock measured as bytes plus SHA-256 and matching `3420025` / `281b1cfcc67eb63e19ef83daf06197bf3d3b23db0b6bc9b73e02fc18ee278fb6`.
- [x] GitHub descriptor measured as bytes plus SHA-256 and matching `43354021` / `d1978c0c6fd0eb66e9fcd4d78d637864a6e486f558aaad1e51550bc43758b899`.
- [x] `git diff --check`.
- [x] Full `make verify` locally, including lint, with no package timeout budget above 20 minutes.
- [x] Inline `verify-work` and code-review records completed.
- [ ] PR opened with `Refs #4328`; GitHub API confirms its base is exactly `main`.

## Measurements and results

- **Blast radius:** The corrected rule found **3** currently unprotected flags across all **552** connector definitions: HubSpot `reverse delete-oauth-v1-refresh-tokens-token-archive --token`; and `reverse post-oauth-v1-token-create --client-secret` and `--refresh-token`. All three are corrected in this change. The post-correction full corpus validation returned `0` findings.
- **GitHub source lock, measured:** `3420025` bytes, SHA-256 `281b1cfcc67eb63e19ef83daf06197bf3d3b23db0b6bc9b73e02fc18ee278fb6`.
- **GitHub descriptor, measured:** `43354021` bytes, SHA-256 `d1978c0c6fd0eb66e9fcd4d78d637864a6e486f558aaad1e51550bc43758b899`.
- **Forbidden-path check:** `git diff --name-only -- internal/connectors/defs/github/rate_limits.json` returned no paths.
- **Focused green result:** `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestValidate_CLISurfaceEnvOnlyFlagRequiresDeclaredSecretGraphQLContract|TestValidate_CLISurfaceEnvOnlyFlagAcceptsDeclaredRESTSecretRegardlessOfFlagShape|TestValidate_RealGitHubSecretCommandRequiresEnvOnly|TestValidate_RealHubSpotRequestSecretsRequireEnvOnly|TestSourceProjectionMarksDeclaredCircleCIWebhookSecretsEnvOnly)$'` exited `0` (`ok polymetrics.ai/cmd/connectorgen 1.608s`).
- **Package suite:** `go test -count=1 -timeout 20m ./cmd/connectorgen/...` exited `0` (`ok polymetrics.ai/cmd/connectorgen 364.107s`).
- **Generated skill parity repair:** The first full `make verify` run reached the full test target and failed only at `internal/cli` `TestSkillsGenerateMatchesTrackedSkills`: the three corrected HubSpot flags render as `env-only`. Refreshed `docs/skills/pm-hubspot/SKILL.md` from `go run ./cmd/pm skills generate --dir <temporary-dir>`; `diff -qr` found that one intended generated file. `go test -count=1 -timeout 20m ./internal/cli -run '^TestSkillsGenerateMatchesTrackedSkills$'` then exited `0` (`ok polymetrics.ai/internal/cli 4.467s`). A complete clean `make verify` rerun remains required before PR creation.
- **Final full gate:** `make verify` exited `0`. It reported `connectorgen validate: 552 connector(s) checked, 0 findings`, `connectorgen surface-sync: 552 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)`, `golangci-lint` `0 issues`, clean connector boundary, and `installed GitHub certification archive proof passed`.
- **Diff integrity:** `git diff --check` exited `0` after all generated artifacts were refreshed.
- **Final focused regression:** `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestValidate_CLISurfaceEnvOnlyFlagRequiresDeclaredSecretGraphQLContract|TestValidate_CLISurfaceEnvOnlyFlagAcceptsDeclaredRESTSecretRegardlessOfFlagShape|TestValidate_RealGitHubSecretCommandRequiresEnvOnly|TestValidate_RealHubSpotRequestSecretsRequireEnvOnly|TestSourceProjectionMarksDeclaredCircleCIWebhookSecretsEnvOnly)$'` exited `0` (`ok polymetrics.ai/cmd/connectorgen 1.764s`) after adding the enclosing JSON-field secret case.
