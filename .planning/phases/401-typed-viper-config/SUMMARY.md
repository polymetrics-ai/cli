# Issue 401 Summary — Typed Viper Configuration

Status: planning artifacts created; red tests pending.

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

Focused gates passed; broader gates pending.

## Review route

Claude workflow remains `disabled_manually`; Copilot quota exhausted for this blocker window. Do not post `@claude review`; do not request Copilot. Human/parent fallback pending.

## Human gates

No secrets, credentialed checks, dependency deviations, generic write tools, reverse ETL execution, quality-gate reduction, or merge to `main`.
