# Issue 400 Summary — Cobra Router Shell

Status: implementation green locally; stacked PR pending.

## Scope delivered

- Added a Cobra router shell around the existing CLI while preserving `cli.Run(args, stdout, stderr) int`.
- Built a fresh `newRootCmd` per invocation.
- Registered `DisableFlagParsing` wrappers for existing top-level handlers; legacy parsers/handlers remain beneath wrappers.
- Kept `extract` and `worker` hidden.
- Preserved dynamic `pm <connector> <path...>` fallback and arbitrary flag passthrough.
- Added `mapCobraErr` so Cobra/pflag-style errors still flow into `writeError`; `writeError` remains the sole exit-code authority.
- Added focused internal Cobra shell tests.
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

## Verification

Passed locally:

- `gofmt -w cmd internal`
- `go vet ./...`
- `go test ./...`
- `go build ./cmd/pm`
- `make verify`

Pending after implementation commit:

- `git diff --check origin/feat/cli-architecture-v2...HEAD`
- `git diff origin/feat/cli-architecture-v2...HEAD -- go.mod go.sum`

## Dependency delta

- Direct: `github.com/spf13/cobra v1.10.2`.
- Indirect go.mod: `github.com/spf13/pflag v1.0.9`, `github.com/inconshreveable/mousetrap v1.1.0`.
- go.sum includes additional Cobra module-metadata `go.mod` checksums only; no additional direct dependency.

## CLI parity

- `/tmp/pm-400 help connectors` -> exit 0.
- `/tmp/pm-400 connectors` -> exit 0.
- `/tmp/pm-400 docs --help` -> exit 0.
- `/tmp/pm-400 worker --help --json` -> preserved current hidden-command golden behavior, exit 1.
- `docs/cli/**`, `website/**`, and generated help/manual artifacts not updated because no help text/command/flag surface changed and golden/docs-diff tests stayed green.

## Review route

Claude workflow is `disabled_manually`; Copilot quota already exhausted for this blocker window. Do not post `@claude review`; do not request Copilot. Record human/parent-PR fallback pending; no approval claims.

## Human gates

No secrets, credentialed checks, dependency deviations, generic write tools, reverse ETL execution, quality-gate reduction, or merge to `main`.
