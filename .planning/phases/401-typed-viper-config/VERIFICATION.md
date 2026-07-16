# Issue 401 Verification Checklist — Typed Viper Configuration

**Issue:** #401
**Branch:** `feat/401-typed-viper-config`
**Base:** `feat/cli-architecture-v2`

## Required gates

- [ ] `gofmt -w cmd internal`
- [ ] `go test ./internal/config/... -count=1`
- [ ] `go test ./internal/cli/ -run 'Golden|Config' -count=1`
- [ ] `go test ./internal/cli/ -run Certify -count=1`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`
- [ ] `git diff --check origin/feat/cli-architecture-v2...HEAD`
- [ ] `git diff origin/feat/cli-architecture-v2...HEAD -- go.mod go.sum`

## Focused TDD gates

- [ ] Red: `go test ./internal/config/... -count=1`
- [ ] Red: `go test ./internal/cli/ -run Config -count=1`
- [ ] Green: `go test ./internal/config/... -count=1`
- [ ] Green: `go test ./internal/cli/ -run 'Golden|Config' -count=1`
- [ ] Certify re-entrancy: `go test ./internal/cli/ -run Certify -count=1`

## Dependency gate

- [ ] Direct dependency added exactly: `github.com/spf13/viper v1.21.0`.
- [ ] No additional direct modules beyond Viper.
- [ ] Resolved transitives recorded (`afero`, `cast`, `fsnotify`, `gotenv`, `toml`, `mapstructure`, `sourcegraph/conc`, and Viper-resolved support modules).
- [ ] No frontend dependency changes.
- [ ] `make tidy-check` remains green inside `make verify`; no tidy-check weakening.
- [ ] `git diff origin/feat/cli-architecture-v2...HEAD -- go.mod go.sum` recorded exactly.

## Config behavior checklist

- [ ] `internal/config.Load` uses `viper.New()` inside load path only.
- [ ] No package-level Viper singleton.
- [ ] No `AutomaticEnv`.
- [ ] No `WatchConfig` / `OnConfigChange`.
- [ ] Missing `.polymetrics/config.yaml` is non-error.
- [ ] Malformed config file returns typed config load error.
- [ ] CLI maps malformed config to validation exit 3 through `writeError`.
- [ ] Config file path is derived from invocation project root.
- [ ] Explicit `POLYMETRICS_*` env bindings with `PM_*` aliases/fallback.
- [ ] Bound `--root` and `--json` flags outrank env/file in `internal/config` where supplied.
- [ ] Existing CLI behavior/golden transcripts preserved; `cli.Run` signature unchanged.
- [ ] Existing scattered `os.Getenv` readers not migrated in this issue.
- [ ] User-named credential env vars and certify credsfile env refs not treated as app config.
- [ ] Invocation isolation/no state leak covered by tests.

## CLI help / docs / website parity

Applies: yes, config behavior and docs are CLI-visible.

- [ ] Runtime help checked: `pm help <config-topic>` or chosen topic.
- [ ] Command help checked: `pm runtime --help` and/or changed command help.
- [ ] Bare namespace behavior checked: `pm runtime` and one unchanged namespace command.
- [ ] `docs/cli/**` updated or marked with exact exemption.
- [ ] `website/**` updated or marked with exact exemption.
- [ ] Generated help/manual artifacts updated or marked with exact exemption.
- [ ] Docs mention config keys, defaults, precedence, aliases, file format, and security boundaries.
- [ ] Docs do not include secret values and do not route user-named credential env vars through app config.
- [ ] Docs/website grep parity recorded.

## Review route

- [ ] Open non-draft stacked PR to `feat/cli-architecture-v2` with `Refs #401` and `Refs #397`.
- [ ] Do not post `@claude review` because repository Claude workflow is `disabled_manually` for this run.
- [ ] Do not request Copilot because quota is exhausted for this blocker window.
- [ ] Record review coverage as human/parent-PR fallback pending; no approval claims.

## Full `make verify` result

Pending.
