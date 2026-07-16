# Issue 401 TDD Ledger — Typed Viper Configuration

**Issue:** #401
**Parent:** #397
**Branch:** `feat/401-typed-viper-config`
**Sub-PR base:** `feat/cli-architecture-v2`

## Loaded skills

- `gsd-core` — `.pi/skills/gsd-core/SKILL.md`.
- `caveman` — `.agents/skills/caveman/SKILL.md`.
- `golang-how-to` — routing table: Viper config layering -> `golang-spf13-viper` + `golang-spf13-cobra`; CLI behavior -> `golang-cli`; tests -> `golang-testing`; env/filesystem/security -> `golang-security` + `golang-safety`.
- `golang-cli` — config layering, exit-code preservation, stdout/stderr discipline, injected writer tests.
- `golang-testing` — Best Practices #1 named subtests, #3 independent tests, #5 observable behavior.
- `golang-error-handling` — Best Practices #2 wrapping, #5 chain inspection, #7 single handling rule.
- `golang-security` — explicit env allowlist, no unbounded env ingestion, no secret values in docs/tests.
- `golang-documentation` — config docs/manual/website parity.
- `golang-spf13-viper` — `viper.New()` isolation, explicit `BindEnv`, optional config file, mapstructure tags, no `AutomaticEnv`/`WatchConfig`.
- `golang-spf13-cobra` — fresh command tree and current global flag binding.
- `golang-structs-interfaces` — typed config structs with explicit tags.
- `golang-safety` — no package singleton, no state leakage, safe zero/default values.
- Website-awareness only: `vercel-react-best-practices`, `vercel-composition-patterns`; no React component work planned.

Missing repo-local stack skill: `.pi/skills/go-implementation/SKILL.md` returned `ENOENT`; `.pi/skills/ts-website/SKILL.md` returned `ENOENT`. User-required Go skills above were loaded.

## GSD evidence

```bash
scripts/gsd doctor
```

Result: pass (`ok` for node, repo root, official docs, command registry, upstream lock, Pi settings/extension/skill/prompt, commands=69).

```bash
scripts/gsd prompt plan-phase 401 --skip-research >/tmp/gsd-plan-phase-401.prompt && wc -c /tmp/gsd-plan-phase-401.prompt
```

Result: pass (`10668 /tmp/gsd-plan-phase-401.prompt`).

```bash
scripts/gsd prompt programming-loop init --phase 401 --dry-run >/tmp/gsd-programming-loop-401.prompt 2>/tmp/gsd-programming-loop-401.err
```

Result: adapter gap, exit 1:

```text
scripts/gsd: unknown GSD command: programming-loop
```

Manual fallback: `.pi/prompts/pm-gsd-loop.md` loaded and followed.

## Dependency decision evidence

```bash
go list -m -versions github.com/spf13/viper
```

Result includes latest stable v1:

```text
github.com/spf13/viper ... v1.20.1 v1.21.0
```

Selected `github.com/spf13/viper v1.21.0` per ADR 0002 / task dependency gate.

## Cycle 0 — planning setup

Status: complete.

Artifacts created before production edits:

- `.planning/phases/401-typed-viper-config/PLAN.md`
- `.planning/phases/401-typed-viper-config/TDD-LEDGER.md`
- `.planning/phases/401-typed-viper-config/VERIFICATION.md`
- `.planning/phases/401-typed-viper-config/SUMMARY.md`
- `.planning/phases/401-typed-viper-config/PROMPTS.md`
- `.planning/phases/401-typed-viper-config/RUN-STATE.json`

## Cycle 1 — red tests before implementation

Planned red commands:

```bash
go test ./internal/config/... -count=1
go test ./internal/cli/ -run Config -count=1
```

Expected red:

- `internal/config` package/tests absent before implementation.
- CLI malformed-config validation tests fail until `cli.Run` loads config and maps load errors to validation exit 3.

Actual red evidence captured before production implementation:

```bash
go test ./internal/config/... -count=1
```

Result: fail, exit 1.

```text
# polymetrics.ai/internal/config [polymetrics.ai/internal/config.test]
internal/config/config_test.go:20:18: undefined: Config
internal/config/config_test.go:46:14: undefined: Load
internal/config/config_test.go:46:19: undefined: Options
internal/config/config_test.go:88:180: undefined: Config
internal/config/config_test.go:89:160: undefined: Config
internal/config/config_test.go:90:158: undefined: Config
internal/config/config_test.go:91:204: undefined: Config
internal/config/config_test.go:92: undefined: Config
internal/config/config_test.go:93:244: undefined: Config
internal/config/config_test.go:94: undefined: Config
internal/config/config_test.go:94: too many errors
FAIL	polymetrics.ai/internal/config [build failed]
FAIL
```

```bash
go test ./internal/cli/ -run Config -count=1
```

Result: fail, exit 1.

```text
--- FAIL: TestConfigMalformedFileExitsValidation (0.00s)
    config_test.go:25: exit code = 0, want 3
        stdout={
          "api_version": "polymetrics.ai/v1",
          "commit": "none",
          "date": "unknown",
          "kind": "Version",
          "version": "dev"
        }

        stderr=
FAIL
FAIL	polymetrics.ai/internal/cli	1.052s
FAIL
```

## Cycle 2 — green implementation evidence

Implemented `internal/config` with Viper instance-mode load, explicit env allowlist, typed structs, missing-file handling, malformed-file `LoadError`, bound global flag values, and minimal CLI validation integration. Added Viper v1.21.0 dependency.

```bash
gofmt -w internal/config internal/cli/cli.go internal/config/config_test.go
go mod tidy
go test ./internal/config/... -count=1
go test ./internal/cli/ -run 'Golden|Config' -count=1
```

Result:

```text
ok  	polymetrics.ai/internal/config	0.473s
ok  	polymetrics.ai/internal/cli	6.887s
```

Certify re-entrancy focused gate:

```bash
go test ./internal/cli/ -run Certify -count=1
```

Result:

```text
ok  	polymetrics.ai/internal/cli	91.181s
```

Dependency delta observed after `go get github.com/spf13/viper@v1.21.0 && go mod tidy`:

- Direct: `github.com/spf13/viper v1.21.0`.
- Indirect additions/updates: `github.com/fsnotify/fsnotify v1.9.0`, `github.com/pelletier/go-toml/v2 v2.2.4`, `github.com/sagikazarmark/locafero v0.11.0`, `github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8`, `github.com/spf13/afero v1.15.0`, `github.com/spf13/cast v1.10.0`, `github.com/spf13/pflag v1.0.10` (upgrade from Cobra's v1.0.9), `github.com/subosito/gotenv v1.6.0`, `go.yaml.in/yaml/v3 v3.0.4`.
- No additional direct module beyond Viper; no frontend dependency changes.

## Cycle 3 — refactor / docs parity evidence

Added `pm help config` docs topic, generated `docs/cli/config.md`, updated `website/content/docs/cli-reference.mdx`, and regenerated `website/lib/docs.generated.ts` with `node website/scripts/gen-docs-data.mjs`.

Parity spot checks after `go build -o /tmp/pm-401 ./cmd/pm`:

```text
/tmp/pm-401 help config      -> exit 0, stdout 4203 bytes, stderr 0 bytes
/tmp/pm-401 runtime          -> exit 0, stdout 470 bytes, stderr 0 bytes
/tmp/pm-401 runtime --help   -> exit 0, stdout 470 bytes, stderr 0 bytes
/tmp/pm-401 config --help    -> exit 0, stdout 4203 bytes, stderr 0 bytes
```

Docs/website grep:

```bash
rg -n "POLYMETRICS_ROOT|PM_ROOT|runtime.postgres_url|rlm.llm.model|PM_LLM_API_KEY" docs/cli website/content/docs/cli-reference.mdx website/lib/docs.generated.ts
```

Result: exit 0, 11 lines.

Refactor safety checks:

- `cli.Run(args, stdout, stderr) int` signature preserved.
- Existing `os.Getenv` readers left for #402.
- No `AutomaticEnv`, `WatchConfig`, or package-level Viper singleton.
- `viper.New()` appears only in `internal/config.Load`.

## Cycle 4 — final verification evidence

```bash
gofmt -w cmd internal
go test ./internal/config/... -count=1
go test ./internal/cli/ -run 'Golden|Config' -count=1
go test ./internal/cli/ -run Certify -count=1
```

Result:

```text
ok  	polymetrics.ai/internal/config	0.228s
ok  	polymetrics.ai/internal/cli	6.812s
ok  	polymetrics.ai/internal/cli	91.270s
```

```bash
go vet ./...
go test ./...
go build ./cmd/pm
```

Result: pass. `go vet` and `go build` produced no output; `go test ./...` passed with notable packages `polymetrics.ai/internal/cli 158.655s` and `polymetrics.ai/internal/connectors/certify 343.692s`.

```bash
make verify
```

Result: pass; final line:

```text
connectorgen validate: 547 connector(s) checked, 0 findings
```

```bash
git diff --check origin/feat/cli-architecture-v2...HEAD
git diff origin/feat/cli-architecture-v2...HEAD -- go.mod go.sum
```

Result: diff check pass/no output; dependency diff recorded approved Viper delta only.

## Cycle 5 — review-fix plan / accepted pm-reviewer finding

Status: planned before review-fix production edits.

Finding disposition: Accepted. CLI must honor `POLYMETRICS_ROOT`/`PM_ROOT` for discovery before loading `.polymetrics/config.yaml`, honor `POLYMETRICS_JSON`/`PM_JSON` for malformed-config error rendering, and use loaded `Config.Root`/`Config.JSON` for invocation after successful load.

GSD refresh evidence:

```bash
scripts/gsd doctor
```

Result: pass (`ok` for node, repo root, official docs, commands registry, upstream lock, Pi settings/extension/skill/prompt, commands=69).

```bash
scripts/gsd prompt programming-loop init --phase 401 --dry-run >/tmp/gsd-programming-loop-401-reviewfix.prompt 2>/tmp/gsd-programming-loop-401-reviewfix.err
```

Result: adapter gap remains, exit 1:

```text
scripts/gsd: unknown GSD command: programming-loop
```

Manual fallback remains `.pi/prompts/pm-gsd-loop.md`.

Skills refreshed for this review-fix cycle: `gsd-core`, `caveman`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-documentation`, `golang-spf13-viper`, `golang-spf13-cobra`, `golang-safety`, and `golang-lint`. Missing repo-local stack skill remains `.pi/skills/go-implementation/SKILL.md` (`ENOENT`), with required Go skills loaded per `.agents/agentic-delivery/references/required-skills-routing.md` Always-on Go skill routing and CLI/Viper rules.

Planned red tests before production edits:

```bash
go test ./internal/config/... -run 'Bootstrap|Discovery|ConfigFileRoot' -count=1
go test ./internal/cli/ -run Config -count=1
```

Expected red: env/alias root discovery, env/alias JSON malformed-error rendering, config-file `json`/`root` invocation, and CLI invocation isolation fail because current CLI discards `Config` and uses flag-only root/json.

Actual red evidence captured before review-fix production edits:

```bash
gofmt -w internal/config/config_test.go internal/cli/config_test.go
go test ./internal/config/... -run 'Bootstrap|Discovery|ConfigFileRoot' -count=1
```

Result: fail, exit 1.

```text
# polymetrics.ai/internal/config [polymetrics.ai/internal/config.test]
internal/config/config_test.go:246:22: undefined: ResolveBootstrap
FAIL	polymetrics.ai/internal/config [build failed]
FAIL
```

```bash
go test ./internal/cli/ -run Config -count=1
```

Result: fail, exit 1.

```text
--- FAIL: TestConfigRootEnvControlsDiscoveryAndInvocationRoot (0.00s)
    config_test.go:62: env root was not initialized: stat /var/folders/tk/bmp_tx0976s4rkh1phvrpjlw0000gn/T/TestConfigRootEnvControlsDiscoveryAndInvocationRoot985015084/002/.polymetrics/config.yaml: no such file or directory
--- FAIL: TestConfigPMRootAliasControlsDiscoveryAndInvocationRoot (0.00s)
    config_test.go:81: PM_ROOT alias root was not initialized: stat /var/folders/tk/bmp_tx0976s4rkh1phvrpjlw0000gn/T/TestConfigPMRootAliasControlsDiscoveryAndInvocationRoot346832289/002/.polymetrics/config.yaml: no such file or directory
--- FAIL: TestConfigJSONEnvRendersMalformedConfigAsJSON (0.00s)
    config_test.go:110: exit code = 0, want 3
        stdout=pm dev
        commit: none
        built: unknown

        stderr=
--- FAIL: TestConfigPMJSONAliasRendersMalformedConfigAsJSON (0.00s)
    config_test.go:127: exit code = 0, want 3
        stdout=pm dev
        commit: none
        built: unknown

        stderr=
--- FAIL: TestConfigFileJSONControlsInvocationOutput (0.00s)
    config_test.go:155: stdout = pm dev
        commit: none
        built: unknown
        , want JSON version envelope from config file json
--- FAIL: TestConfigFileRootControlsInvocationWithoutRelocatingDiscovery (0.00s)
    config_test.go:174: stdout = Initialized Polymetrics project at /var/folders/tk/bmp_tx0976s4rkh1phvrpjlw0000gn/T/TestConfigFileRootControlsInvocationWithoutRelocatingDiscovery3262617763/001/.polymetrics
        , want JSON init result for file root
--- FAIL: TestConfigInvocationIsolation (0.00s)
    config_test.go:188: json root stdout = pm dev
        commit: none
        built: unknown
        , want JSON version envelope
FAIL
FAIL	polymetrics.ai/internal/cli	1.061s
FAIL
```

Test-plan correction before final green: `TestConfigFileRootControlsInvocationWithoutRelocatingDiscovery` now uses cwd/default discovery instead of explicit `--root` so it validates file `root` behavior without violating flags > file precedence.

## Cycle 6 — review-fix green implementation evidence

Implemented shared `internal/config.ResolveBootstrap`, reused the same env-binding table for bootstrap and Viper `BindEnv`, moved config-file discovery to the bootstrap root, and changed `cli.Run` to use loaded `Config.Root`/`Config.JSON` for command invocation after successful load. No `AutomaticEnv`, `WatchConfig`, package-level Viper singleton, new dependency, or #402 env-reader migration added.

Focused green gates:

```bash
gofmt -w internal/config/config.go internal/config/config_test.go internal/cli/cli.go internal/cli/config_test.go internal/cli/docs.go
go test ./internal/config/... -count=1
go test ./internal/cli/ -run 'Golden|Config' -count=1
```

Result:

```text
ok  	polymetrics.ai/internal/config	0.507s
ok  	polymetrics.ai/internal/cli	7.054s
```

```bash
go test ./internal/cli/ -run Certify -count=1
```

Result:

```text
ok  	polymetrics.ai/internal/cli	103.737s
```

Docs parity update: clarified that `--json`, `POLYMETRICS_JSON=true`, or `PM_JSON=true` selects JSON error envelopes for malformed config, and that `PM_ROOT` can select discovery root before file load. Regenerated website data with `node website/scripts/gen-docs-data.mjs`.

## Cycle 7 — review-fix full verification evidence

```bash
gofmt -w cmd internal
go test ./internal/config/... -count=1
go test ./internal/cli/ -run 'Golden|Config' -count=1
go test ./internal/cli/ -run Certify -count=1
go vet ./...
go test ./...
go build ./cmd/pm
make verify
git diff --check origin/feat/cli-architecture-v2...HEAD
git diff origin/feat/cli-architecture-v2...HEAD -- go.mod go.sum
```

Results:

```text
go test ./internal/config/... -count=1 -> ok  	polymetrics.ai/internal/config	0.331s
go test ./internal/cli/ -run 'Golden|Config' -count=1 -> ok  	polymetrics.ai/internal/cli	10.587s
go test ./internal/cli/ -run Certify -count=1 -> ok  	polymetrics.ai/internal/cli	120.088s
go vet ./... -> pass, no output
go test ./... -> pass; internal/cli 197.284s; internal/config 1.086s; internal/connectors/certify 385.345s
go build ./cmd/pm -> pass, no output
make verify -> pass; final line: connectorgen validate: 547 connector(s) checked, 0 findings
git diff --check origin/feat/cli-architecture-v2...HEAD -> pass, no output
git diff origin/feat/cli-architecture-v2...HEAD -- go.mod go.sum -> approved Viper delta only; no new review-fix dependency changes
```

Config help/docs/website parity:

```bash
go build -o /tmp/pm-401-reviewfix ./cmd/pm
/tmp/pm-401-reviewfix help config
/tmp/pm-401-reviewfix runtime
/tmp/pm-401-reviewfix runtime --help
/tmp/pm-401-reviewfix config --help
rg -n "POLYMETRICS_ROOT|PM_ROOT|POLYMETRICS_JSON|PM_JSON|malformed|relocate" docs/cli/config.md website/content/docs/cli-reference.mdx website/lib/docs.generated.ts
```

Result:

```text
/tmp/pm-401-reviewfix help config -> exit 0, stdout 4375 bytes, stderr 0 bytes
/tmp/pm-401-reviewfix runtime -> exit 0, stdout 470 bytes, stderr 0 bytes
/tmp/pm-401-reviewfix runtime --help -> exit 0, stdout 470 bytes, stderr 0 bytes
/tmp/pm-401-reviewfix config --help -> exit 0, stdout 4375 bytes, stderr 0 bytes
rg -> exit 0, 11 matches
```
