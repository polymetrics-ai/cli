# Issue 401 Summary — Typed Viper Configuration

Status: final re-review website caveat fixed in commit `10938836cf2a846e03e2c284ce2ddeeec7c4f193` from starting head `77e8fe559b6bab458ed19cb30d3fdc6aa6778f56`; local gates passed; push/PR body update pending; human/parent review fallback remains pending.

## Final re-review disposition

- Finding: website config section lists runtime/RLM/schedule typed keys without the CLI manual caveat that those command readers still use legacy env readers until #402.
- Disposition: Accepted.
- Fix: added a concise website caveat distinguishing `root`/`json` as CLI-effective now from runtime/RLM/schedule migration owned by #402, then regenerated `website/lib/docs.generated.ts`.
- Pre-edit validation: `rg -n "Current command behavior|#402|legacy readers|env-reader migration" website/content/docs/cli-reference.mdx` exited 1 with no output; `docs/cli/config.md` already carried the caveat.
- Scope guard: no Go code behavior change, no dependency change, no frontend dependency install, no parent/shared edit, no Claude/Copilot request.

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

## Final re-review verification

Passed locally:

- `gofmt -w cmd internal` -> pass, no output.
- `node website/scripts/gen-docs-data.mjs` -> `Wrote 11 docs pages to lib/docs.generated.ts.`
- `git diff --check origin/feat/cli-architecture-v2...HEAD` -> pass, no output.
- `go test ./internal/config/... -count=1` -> `ok  	polymetrics.ai/internal/config	0.498s`.
- `go test ./internal/cli/ -run 'Golden|Config' -count=1` -> `ok  	polymetrics.ai/internal/cli	6.817s`.
- `go vet ./...` -> pass, no output.
- `go build ./cmd/pm` -> pass, no output.
- `make verify` -> pass; smoke path `/var/folders/tk/bmp_tx0976s4rkh1phvrpjlw0000gn/T/tmp.k3AV9VUv6j`; ended `connectorgen validate: 547 connector(s) checked, 0 findings`.
- Caveat grep -> pass across `website/content/docs/cli-reference.mdx` and `website/lib/docs.generated.ts`.

Website package-script checks without dependency install:

- `npm --prefix website run gen:docs` -> pass; `Wrote 11 docs pages to lib/docs.generated.ts.`
- `npm --prefix website run typecheck` -> failed, exit 127: `sh: tsc: command not found`.
- `npm --prefix website run test:unit` -> failed, exit 127: `sh: vitest: command not found`.
- `npm --prefix website run build` -> failed, exit 127: `sh: next: command not found`.
- `npm --prefix website run lint` -> failed, exit 127: `sh: next: command not found`.
- `npm --prefix website run test:e2e` -> failed, exit 1: `Error: Cannot find module '@playwright/test'`.
- `npm --prefix website run test` -> failed, exit 127: `sh: vitest: command not found`.

Website script failures are dependency-availability blockers only: `website/node_modules` is absent, and this task forbids adding/installing frontend dependencies.

## Review route

Sub-PR: https://github.com/polymetrics-ai/cli/pull/441 (non-draft, base `feat/cli-architecture-v2`).

PR body updated with pm-reviewer disposition and evidence via GitHub REST patch after `gh pr edit` hit the Projects classic GraphQL deprecation. Claude workflow remains `disabled_manually`; Copilot quota exhausted for this blocker window. Did not post `@claude review`; did not request Copilot. Human/parent fallback pending.

## Human gates

No secrets, credentialed checks, dependency deviations, generic write tools, reverse ETL execution, quality-gate reduction, or merge to `main`.
