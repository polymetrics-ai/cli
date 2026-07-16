# Issue 401 Summary — Typed Viper Configuration

Status: PR #441 review-fix pushed; pm-reviewer finding accepted and fixed; review-fix gates passed; PR body updated; human/parent review fallback remains pending.

## Review-fix disposition

- Finding: CLI promised env-driven `root`/`json` config behavior but only used flag-only bootstrap, validation-only `config.Load`, and flag-only malformed-config JSON rendering.
- Disposition: Accepted.
- Fix: added shared `internal/config.ResolveBootstrap`, use env/alias root for config discovery, use bootstrap JSON for malformed-load errors, then invoke commands with resolved `Config.Root` and `Config.JSON` after successful load.
- Scope guard: kept #402 scattered env-reader migration, new dependencies, parent/shared artifacts, frontend dependency changes, and automated review requests out of scope.

## Scope target

- Add `internal/config` typed invocation config using Viper instance mode.
- Load `.polymetrics/config.yaml` from invocation root; missing file non-error, malformed file validation exit 3 through CLI error funnel.
- Bind current Cobra global flags in config load without changing legacy `cli.Run` behavior.
- Keep scattered `os.Getenv` migrations for #402.
- Document config keys/defaults/precedence/aliases/security boundaries in CLI docs + website.

## GSD / TDD

- `scripts/gsd doctor` passed.
- `scripts/gsd prompt plan-phase 401 --skip-research` generated `/tmp/gsd-plan-phase-401.prompt`.
- `scripts/gsd prompt programming-loop init --phase 401 --dry-run` failed with `scripts/gsd: unknown GSD command: programming-loop`; `.pi/prompts/pm-gsd-loop.md` fallback loaded.
- Plan/TDD/verification/run-state artifacts created before production edits.

## Dependency decision

- Selected `github.com/spf13/viper v1.21.0` (latest stable v1 from `go list -m -versions`).
- Human gate remains active for additional direct modules, major-version deviation, frontend dependency changes, or unrelated graph changes.

## Red / green evidence

Red captured before production edits:

- `go test ./internal/config/... -count=1` -> fails because `Config`, `Load`, and `Options` do not exist.
- `go test ./internal/cli/ -run Config -count=1` -> fails because malformed `.polymetrics/config.yaml` is ignored and `pm version --json` exits 0 instead of validation exit 3.

Green:

- `go test ./internal/config/... -count=1` -> `ok  	polymetrics.ai/internal/config	0.473s`.
- `go test ./internal/cli/ -run 'Golden|Config' -count=1` -> `ok  	polymetrics.ai/internal/cli	6.887s`.
- `go test ./internal/cli/ -run Certify -count=1` -> `ok  	polymetrics.ai/internal/cli	91.181s`.

## Verification

Passed locally:

- `gofmt -w cmd internal`.
- `go test ./internal/config/... -count=1` -> `ok  	polymetrics.ai/internal/config	0.228s`.
- `go test ./internal/cli/ -run 'Golden|Config' -count=1` -> `ok  	polymetrics.ai/internal/cli	6.812s`.
- `go test ./internal/cli/ -run Certify -count=1` -> `ok  	polymetrics.ai/internal/cli	91.270s`.
- `go vet ./...` -> pass, no output.
- `go test ./...` -> pass; `internal/cli 158.655s`, `internal/connectors/certify 343.692s`.
- `go build ./cmd/pm` -> pass, no output.
- `make verify` -> pass; ended `connectorgen validate: 547 connector(s) checked, 0 findings`.
- `git diff --check origin/feat/cli-architecture-v2...HEAD` -> pass, no output.
- `git diff origin/feat/cli-architecture-v2...HEAD -- go.mod go.sum` -> approved Viper delta only.

CLI parity:

- `/tmp/pm-401 help config` -> exit 0; stdout 4203 bytes; stderr 0 bytes.
- `/tmp/pm-401 runtime` -> exit 0; stdout 470 bytes; stderr 0 bytes.
- `/tmp/pm-401 runtime --help` -> exit 0; stdout 470 bytes; stderr 0 bytes.
- `/tmp/pm-401 config --help` -> exit 0; stdout 4203 bytes; stderr 0 bytes.
- Docs/website grep for config keys/aliases -> exit 0, 11 lines.

## Review-fix verification

Red captured before production edits:

- `go test ./internal/config/... -run 'Bootstrap|Discovery|ConfigFileRoot' -count=1` -> failed with `undefined: ResolveBootstrap`.
- `go test ./internal/cli/ -run Config -count=1` -> failed because env root/alias, env JSON malformed errors, config-file JSON/root invocation, and isolation were not honored.

Green / gates:

- `go test ./internal/config/... -count=1` -> `ok  	polymetrics.ai/internal/config	0.331s`.
- `go test ./internal/cli/ -run 'Golden|Config' -count=1` -> `ok  	polymetrics.ai/internal/cli	10.587s`.
- `go test ./internal/cli/ -run Certify -count=1` -> `ok  	polymetrics.ai/internal/cli	120.088s`.
- `go vet ./...` -> pass, no output.
- `go test ./...` -> pass; `internal/cli 197.284s`, `internal/config 1.086s`, `internal/connectors/certify 385.345s`.
- `go build ./cmd/pm` -> pass, no output.
- `make verify` -> pass; ended `connectorgen validate: 547 connector(s) checked, 0 findings`.
- `git diff --check origin/feat/cli-architecture-v2...HEAD` -> pass, no output.
- `git diff origin/feat/cli-architecture-v2...HEAD -- go.mod go.sum` -> approved Viper delta only; no new review-fix dependency changes.

Review-fix CLI parity:

- `/tmp/pm-401-reviewfix help config` -> exit 0; stdout 4375 bytes; stderr 0 bytes.
- `/tmp/pm-401-reviewfix runtime` -> exit 0; stdout 470 bytes; stderr 0 bytes.
- `/tmp/pm-401-reviewfix runtime --help` -> exit 0; stdout 470 bytes; stderr 0 bytes.
- `/tmp/pm-401-reviewfix config --help` -> exit 0; stdout 4375 bytes; stderr 0 bytes.
- Docs/website grep for `POLYMETRICS_ROOT|PM_ROOT|POLYMETRICS_JSON|PM_JSON|malformed|relocate` -> pass, 11 matches.
- `node website/scripts/gen-docs-data.mjs` -> `Wrote 11 docs pages to lib/docs.generated.ts`.

## Review route

Sub-PR: https://github.com/polymetrics-ai/cli/pull/441 (non-draft, base `feat/cli-architecture-v2`).

PR body updated with pm-reviewer disposition and evidence via GitHub REST patch after `gh pr edit` hit the Projects classic GraphQL deprecation. Claude workflow remains `disabled_manually`; Copilot quota exhausted for this blocker window. Did not post `@claude review`; did not request Copilot. Human/parent fallback pending.

## Human gates

No secrets, credentialed checks, dependency deviations, generic write tools, reverse ETL execution, quality-gate reduction, or merge to `main`.
