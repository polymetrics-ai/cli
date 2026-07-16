# Issue 401 Verification Checklist — Typed Viper Configuration

**Issue:** #401
**Branch:** `feat/401-typed-viper-config`
**Base:** `feat/cli-architecture-v2`

## Required gates

- [x] `gofmt -w cmd internal` — pass, no output.
- [x] `go test ./internal/config/... -count=1` — pass: `ok  	polymetrics.ai/internal/config	0.228s`.
- [x] `go test ./internal/cli/ -run 'Golden|Config' -count=1` — pass: `ok  	polymetrics.ai/internal/cli	6.812s`.
- [x] `go test ./internal/cli/ -run Certify -count=1` — pass: `ok  	polymetrics.ai/internal/cli	91.270s`.
- [x] `go vet ./...` — pass, no output.
- [x] `go test ./...` — pass; notable packages: `polymetrics.ai/internal/cli 158.655s`, `polymetrics.ai/internal/connectors/certify 343.692s`.
- [x] `go build ./cmd/pm` — pass, no output.
- [x] `make verify` — pass; ended `connectorgen validate: 547 connector(s) checked, 0 findings`.
- [x] `git diff --check origin/feat/cli-architecture-v2...HEAD` — pass, no output.
- [x] `git diff origin/feat/cli-architecture-v2...HEAD -- go.mod go.sum` — pass/recorded approved Viper delta.

## Focused TDD gates

- [x] Red: `go test ./internal/config/... -count=1` — fail: `undefined: Config`, `undefined: Load`, `undefined: Options`.
- [x] Red: `go test ./internal/cli/ -run Config -count=1` — fail: malformed config ignored, `version` exited 0 instead of validation exit 3.
- [x] Green: `go test ./internal/config/... -count=1` — pass.
- [x] Green: `go test ./internal/cli/ -run 'Golden|Config' -count=1` — pass.
- [x] Certify re-entrancy: `go test ./internal/cli/ -run Certify -count=1` — pass.

## Dependency gate

- [x] Direct dependency added exactly: `github.com/spf13/viper v1.21.0`.
- [x] No additional direct modules beyond Viper.
- [x] Resolved transitives recorded (`fsnotify`, `go-toml/v2`, `locafero`, `sourcegraph/conc`, `afero`, `cast`, pflag upgrade, `gotenv`, `go.yaml.in/yaml/v3`; existing `go-viper/mapstructure/v2` retained).
- [x] No frontend dependency changes.
- [x] `make tidy-check` remains green inside `make verify`; no tidy-check weakening.
- [x] `git diff origin/feat/cli-architecture-v2...HEAD -- go.mod go.sum` recorded exactly.

## Config behavior checklist

- [x] `internal/config.Load` uses `viper.New()` inside load path only.
- [x] No package-level Viper singleton.
- [x] No `AutomaticEnv`.
- [x] No `WatchConfig` / `OnConfigChange`.
- [x] Missing `.polymetrics/config.yaml` is non-error.
- [x] Malformed config file returns typed config load error.
- [x] CLI maps malformed config to validation exit 3 through `writeError`.
- [x] Config file path is derived from invocation project root.
- [x] Explicit `POLYMETRICS_*` env bindings with `PM_*` aliases/fallback.
- [x] Bound `--root` and `--json` flags outrank env/file in `internal/config` where supplied.
- [x] Existing CLI behavior/golden transcripts preserved; `cli.Run` signature unchanged.
- [x] Existing scattered `os.Getenv` readers not migrated in this issue.
- [x] User-named credential env vars and certify credsfile env refs documented as not app config.
- [x] Invocation isolation/no state leak covered by tests.

## CLI help / docs / website parity

Applies: yes, config behavior and docs are CLI-visible.

- [x] Runtime help checked: `/tmp/pm-401 help config` — exit 0; stdout 4203 bytes; stderr 0 bytes.
- [x] Command help checked: `/tmp/pm-401 runtime --help` — exit 0; stdout 470 bytes; stderr 0 bytes; `/tmp/pm-401 config --help` — exit 0; stdout 4203 bytes; stderr 0 bytes.
- [x] Bare namespace behavior checked: `/tmp/pm-401 runtime` — exit 0; stdout 470 bytes; stderr 0 bytes.
- [x] `docs/cli/**` updated: `docs/cli/config.md` generated from embedded `configHelp`.
- [x] `website/**` updated: `website/content/docs/cli-reference.mdx` plus regenerated `website/lib/docs.generated.ts`.
- [x] Generated help/manual artifacts updated: `docs/cli/config.md`; website docs generated data regenerated with `node website/scripts/gen-docs-data.mjs`.
- [x] Docs mention config keys, defaults, precedence, aliases, file format, and security boundaries.
- [x] Docs do not include secret values and do not route user-named credential env vars through app config.
- [x] Docs/website grep parity recorded: `rg -n "POLYMETRICS_ROOT|PM_ROOT|runtime.postgres_url|rlm.llm.model|PM_LLM_API_KEY" docs/cli website/content/docs/cli-reference.mdx website/lib/docs.generated.ts` — exit 0, 11 lines.

## Review-fix cycle checklist

Finding disposition: Accepted. Required behavior: bootstrap `root`/`json` from flags > `POLYMETRICS_*` > `PM_*` > defaults before config-file discovery; use loaded `Config.Root`/`Config.JSON` for command invocation after successful load; preserve file-root non-relocation, malformed-file validation category, JSON envelope, stdout/stderr discipline, Viper instance scope, and #402 deferral.

Planned red gates:

- [x] `go test ./internal/config/... -run 'Bootstrap|Discovery|ConfigFileRoot' -count=1` — failed before implementation: `undefined: ResolveBootstrap`.
- [x] `go test ./internal/cli/ -run Config -count=1` — failed before implementation: env root/alias ignored, env JSON malformed errors rendered as human success, config-file JSON/root invocation ignored.

Review-fix required gates:

- [x] `gofmt -w cmd internal` — pass, no output.
- [x] `go test ./internal/config/... -count=1` — pass: `ok  	polymetrics.ai/internal/config	0.331s`.
- [x] `go test ./internal/cli/ -run 'Golden|Config' -count=1` — pass: `ok  	polymetrics.ai/internal/cli	10.587s`.
- [x] `go test ./internal/cli/ -run Certify -count=1` — pass: `ok  	polymetrics.ai/internal/cli	120.088s`.
- [x] `go vet ./...` — pass, no output.
- [x] `go test ./...` — pass; notable packages: `polymetrics.ai/internal/cli 197.284s`, `polymetrics.ai/internal/config 1.086s`, `polymetrics.ai/internal/connectors/certify 385.345s`.
- [x] `go build ./cmd/pm` — pass, no output.
- [x] `make verify` — pass; ended `connectorgen validate: 547 connector(s) checked, 0 findings`; smoke path `/var/folders/tk/bmp_tx0976s4rkh1phvrpjlw0000gn/T/tmp.sjzBMl0mWK`.
- [x] `git diff --check origin/feat/cli-architecture-v2...HEAD` — pass, no output.
- [x] `git diff origin/feat/cli-architecture-v2...HEAD -- go.mod go.sum` — pass/recorded approved Viper delta only; no new dependency changes in review-fix commit.

Review-fix parity gates:

- [x] `/tmp/pm-401-reviewfix help config` — exit 0; stdout 4375 bytes; stderr 0 bytes.
- [x] Bare namespace behavior: `/tmp/pm-401-reviewfix runtime` — exit 0; stdout 470 bytes; stderr 0 bytes.
- [x] Command help: `/tmp/pm-401-reviewfix runtime --help` — exit 0; stdout 470 bytes; stderr 0 bytes; `/tmp/pm-401-reviewfix config --help` — exit 0; stdout 4375 bytes; stderr 0 bytes.
- [x] Docs/website grep for root/json env aliases and file-root non-relocation — pass: 11 matches across `docs/cli/config.md`, `website/content/docs/cli-reference.mdx`, and `website/lib/docs.generated.ts`.
- [x] Generated website data regenerated through existing generator: `node website/scripts/gen-docs-data.mjs` -> `Wrote 11 docs pages to lib/docs.generated.ts`.

## Review route

- [x] Open non-draft stacked PR to `feat/cli-architecture-v2` with `Refs #401` and `Refs #397`: PR #441.
- [x] Do not post `@claude review` because repository Claude workflow is `disabled_manually` for this run.
- [x] Do not request Copilot because quota is exhausted for this blocker window.
- [x] Record review coverage as human/parent-PR fallback pending; no approval claims.
- [x] Update PR body with pm-reviewer finding disposition and review-fix evidence after gates pass — REST patch via `gh api repos/polymetrics-ai/cli/pulls/441 -X PATCH`; `gh pr edit` was blocked by GitHub Projects classic GraphQL deprecation.

## Full `make verify` result

Pass. Final lines:

```text
./pm docs validate --connectors-dir docs/connectors
Validated connector docs in docs/connectors
smoke ok: /var/folders/tk/bmp_tx0976s4rkh1phvrpjlw0000gn/T/tmp.yANpw9EndF
golangci-lint run ./internal/connectors/engine/... ./internal/connectors/defs/... ./internal/connectors/hooks/... ./internal/connectors/native/... ./internal/connectors/conformance/... ./internal/connectors/certify/... ./cmd/connectorgen/...
0 issues.
go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 547 connector(s) checked, 0 findings
```
