# TDD ledger: Issue #4328

## Planned evidence

| Slice | Red | Green | Refactor/verification |
| --- | --- | --- | --- |
| Declaration-owned sensitivity | Real-definition tests demonstrate a REST secret, a non-top-level secret, and CircleCI's signing-secret are ordinary flags under the old narrow predicate. | The same validation path sets `env_only` for every declared request secret without protocol or mapping-shape checks. | Sweep all definitions and record the number newly protected by the generalized rule. |
| Existing behavior | Existing GraphQL mutation secret behavior is recorded before the change. | GraphQL remains `env_only`; a real non-secret remains false. | Focused package test covers every required row in one observable table. |
| Repository safety | Baseline GitHub source/descriptor measurements are captured before production edits. | Post-change measurements match exactly; forbidden rate-limit file is absent from the diff. | Full `make verify`, review, and PR-base API read-back complete delivery. |

## Actual evidence

### 2026-08-23 — planning checkpoint

- Red: pending. No production source or test edit has occurred.
- Green: pending.
- Manual GSD fallback: inline execution is required because the task contract disallows role spawning. The resolved prompts and loaded skills are recorded in `PLAN.md`.

### 2026-08-23 — red regression tests

- Red: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestValidate_CLISurfaceEnvOnlyFlagAcceptsDeclaredRESTSecretRegardlessOfFlagShape|TestSourceProjectionMarksDeclaredCircleCIWebhookSecretsEnvOnly)$'` failed as required.
- Observable gap: the existing validator rejected the declared CircleCI REST signing-secret because it required a GraphQL mutation's JSON/top-level/required shape; source projection emitted both declared secret fields with `env_only = false` and discarded `x-secret` from the projected write schema.
- Green: pending the declaration-driven projection and validator generalization.

### 2026-08-23 — green declaration-driven protection

- Green: `gofmt -w cmd/connectorgen/main_test.go cmd/connectorgen/sourceprojection.go cmd/connectorgen/sourceprojection_test.go cmd/connectorgen/validate.go && go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestValidate_CLISurfaceEnvOnlyFlagRequiresDeclaredSecretGraphQLContract|TestValidate_CLISurfaceEnvOnlyFlagAcceptsDeclaredRESTSecretRegardlessOfFlagShape|TestValidate_RealGitHubSecretCommandRequiresEnvOnly|TestValidate_RealHubSpotRequestSecretsRequireEnvOnly|TestSourceProjectionMarksDeclaredCircleCIWebhookSecretsEnvOnly)$'` exited `0` (`ok polymetrics.ai/cmd/connectorgen 1.608s`).
- The real GitHub GraphQL mutation remains protected; real HubSpot REST request secrets, including path and query mappings, are now protected; the REST nested-body and ordinary non-secret cases are covered through the validation path; and the CircleCI webhook projection retains `x-secret` and emits `env_only` for `signing-secret` and `callback-token`.
- The corrected validator sweep on the uncorrected definition surfaces found **3** newly required flags: HubSpot `reverse delete-oauth-v1-refresh-tokens-token-archive --token`, and `reverse post-oauth-v1-token-create --client-secret` plus `--refresh-token`. They are now marked `env_only` in the owned generated surface.
- Post-correction corpus evidence: `go run ./cmd/connectorgen validate internal/connectors/defs --json > /tmp/cli-4328-connectorgen-validate.json && jq -e 'if type == "array" then length == 0 else (.findings // [] | length == 0) end' /tmp/cli-4328-connectorgen-validate.json && jq -r 'if type == "array" then length else (.findings // [] | length) end' /tmp/cli-4328-connectorgen-validate.json` exited `0`, reporting `true` and `0` findings. The final full-gate invocation measured the complete corpus as `connectorgen validate: 552 connector(s) checked, 0 findings`.

### 2026-08-23 — final green gate

- Refactor verification: source projection preserves only its existing bare-string behavior while adding the declaration-owned secret marker; `go run ./cmd/connectorgen surface-sync --check` reported `552 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)`.
- Generated artifacts: `./pm docs generate --dir docs/cli`, `./pm skills generate --dir docs/skills`, and `go run ./cmd/connectorgen certification-subject` refreshed exactly the HubSpot manual/skill surfaces and certification subject that derive from the corrected flags.
- Full Green: `make verify` exited `0`, including the complete `go test -timeout 20m ./...`, docs validation, smoke, `golangci-lint` (`0 issues`), validation, surface synchronization, GitHub parity artifacts, certification subject/matrix/candidates/sweep, connector boundary, connector canon, release tooling, and installed GitHub certification archive proof.

### 2026-08-23 — PR #4334 base reconciliation

- Reason: after the PR opened, `main` advanced to `e338cd301` through #4327. The website generated-data gate therefore compared this branch's pre-merge output with the new bundle corpus.
- Green: merged `origin/main` with `git merge --no-edit origin/main` (merge commit, no rebase), then ran `cd website && pnpm run gen:website-data`. The repository generator refreshed the two derived catalog payloads; no file was hand-edited.
- Post-merge regression: `GOFLAGS=-p=3 go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestValidate_CLISurfaceEnvOnlyFlagRequiresDeclaredSecretGraphQLContract|TestValidate_CLISurfaceEnvOnlyFlagAcceptsDeclaredRESTSecretRegardlessOfFlagShape|TestValidate_RealGitHubSecretCommandRequiresEnvOnly|TestValidate_RealHubSpotRequestSecretsRequireEnvOnly|TestSourceProjectionMarksDeclaredCircleCIWebhookSecretsEnvOnly)$'` exited `0` (`ok polymetrics.ai/cmd/connectorgen 1.681s`).
- Rebuild: `GOFLAGS=-p=3 go build ./cmd/pm` exited `0`.
