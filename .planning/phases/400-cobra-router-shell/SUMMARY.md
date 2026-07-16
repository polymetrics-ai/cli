# Issue 400 Summary — Cobra Router Shell

Status: PR #440 review-fix implemented locally; exact gates passed; push pending.

## Scope delivered

- Added a Cobra router shell around the existing CLI while preserving `cli.Run(args, stdout, stderr) int`.
- Built a fresh `newRootCmd` per invocation.
- Registered `DisableFlagParsing` wrappers for existing top-level handlers; legacy parsers/handlers remain beneath wrappers.
- Kept `extract` and `worker` hidden.
- Preserved dynamic `pm <connector> <path...>` fallback and arbitrary flag passthrough.
- Added `mapCobraErr` so Cobra/pflag-style errors still flow into `writeError`; `writeError` remains the sole exit-code authority.
- Added focused internal Cobra shell tests, including legacy fallback help interception for unknown/dynamic connector commands.
- Review-fix: added root persistent `--root` / `--json` definitions on each fresh root without changing `parseGlobal` ownership.
- Review-fix: marked legacy/root-fallback errors so `mapCobraErr` bypasses plain legacy messages containing `unknown flag` / `unknown command` while genuine Cobra parse errors still map to usage.
- Review-fix: expanded tests for all top-level wrappers including `init`, per-root flag state isolation, genuine Cobra parse mapping, legacy bypass, and deterministic dynamic connector passthrough with late globals.
- Added approved dependency `github.com/spf13/cobra v1.10.2`.

## TDD evidence

Red before implementation:

- `go list -deps ./internal/cli | grep '^github.com/spf13/cobra$'` -> no output, exit 1.
- `go test ./internal/cli/ -run TestCobraRouterShell -count=1` -> setup failure: missing `github.com/spf13/cobra` module.

Green/refactor:

- `go test ./internal/cli/ -run TestCobraRouterShell -count=1` -> pass.
- `go test ./internal/cli/ -run Golden -count=1` -> pass, byte-identical; no golden fixture updates.
- `go test ./internal/cli/ -run Certify -count=1` -> pass.
- `go test ./internal/cli/ -count=1` -> pass.
- Post-refactor `go test ./internal/cli/ -run 'TestCobraRouterShell|Golden' -count=1` -> pass.
- Post-refactor `go test ./internal/cli/ -run Certify -count=1` -> pass.
- Review-fix red: `go test ./internal/cli/ -run TestCobraRouterShell -count=1` failed because root persistent flags were missing and legacy handler errors were categorized as usage.
- Review-fix green: `go test ./internal/cli/ -run TestCobraRouterShell -count=1` -> `ok  	polymetrics.ai/internal/cli	2.026s`.

## Verification

Passed locally for review-fix:

- `gofmt -w cmd internal` -> pass.
- `go test ./internal/cli/ -run 'TestCobraRouterShell|Golden' -count=1` -> `ok  	polymetrics.ai/internal/cli	7.724s`.
- `go test ./internal/cli/ -run Certify -count=1` -> `ok  	polymetrics.ai/internal/cli	92.156s`.
- `go test ./internal/cli/ -count=1` -> `ok  	polymetrics.ai/internal/cli	155.648s`.
- `go vet ./...` -> pass, no output.
- `go test ./...` -> pass; `internal/cli 162.510s`, `internal/connectors/certify 347.398s`.
- `go build ./cmd/pm` -> pass, no output.
- `make verify` -> pass; ended `connectorgen validate: 547 connector(s) checked, 0 findings`.
- `git diff --check origin/feat/cli-architecture-v2...HEAD` -> pass, no output.
- `git diff origin/feat/cli-architecture-v2...HEAD -- go.mod go.sum` -> recorded expected dependency delta only.

## Dependency delta

- Direct: `github.com/spf13/cobra v1.10.2`.
- Indirect go.mod: `github.com/spf13/pflag v1.0.9`, `github.com/inconshreveable/mousetrap v1.1.0`.
- go.sum includes additional Cobra module-metadata `go.mod` checksums only; no additional direct dependency.

## CLI parity

- `/tmp/pm-400 help connectors` -> exit 0.
- `/tmp/pm-400 connectors` -> exit 0.
- `/tmp/pm-400 docs --help` -> exit 0.
- `/tmp/pm-400 worker --help --json` -> preserved current hidden-command golden behavior, exit 1.
- Review-fix spot checks: `/tmp/pm-400 help connectors`, `/tmp/pm-400 connectors`, and `/tmp/pm-400 docs --help` all exit 0 with unchanged byte counts.
- `docs/cli/**`, `website/**`, and generated help/manual artifacts not updated because no help text/command/flag output changed and golden/docs-diff tests stayed green.

## Review-fix dispositions

Accepted pm-reviewer findings:

- MEDIUM: `mapCobraErr` string-matched legacy handler/root-fallback plain errors containing `unknown flag` or `unknown command`. Implemented `cobraLegacyError` marking/bypass; `writeError` taxonomy unchanged; regression tests cover legacy bypass and genuine Cobra parse usage mapping.
- LOW/MEDIUM: root command missed ADR-required persistent `--root` / `--json` definitions. Implemented persistent flags on each fresh root with defaults from parsed invocation state; `parseGlobal` remains semantic owner; `DisableFlagParsing` unchanged; tests cover state isolation.

Residual gaps closed in scope: all-wrapper `DisableFlagParsing`/visibility assertions including `init`; deterministic dynamic connector passthrough with arbitrary connector flags plus late globals.

## Review route

Claude workflow is `disabled_manually`; Copilot quota already exhausted for this blocker window. Do not post `@claude review`; do not request Copilot. Record human/parent-PR fallback pending; no approval claims.

## Human gates

No secrets, credentialed checks, dependency deviations, generic write tools, reverse ETL execution, quality-gate reduction, or merge to `main`.
