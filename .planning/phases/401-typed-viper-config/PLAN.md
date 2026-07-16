# Issue 401 Plan — Typed Viper Configuration

**Issue:** [#401](https://github.com/polymetrics-ai/cli/issues/401)
**Parent:** [#397](https://github.com/polymetrics-ai/cli/issues/397)
**Parent PR:** [#438](https://github.com/polymetrics-ai/cli/pull/438) (`feat/cli-architecture-v2` → `main`, draft)
**Worker branch:** `feat/401-typed-viper-config`
**Sub-PR base:** `feat/cli-architecture-v2`
**Parent dependency integrated:** #400 via PR #440, parent commit `8900db141cc289b65491365d2ebcab490af57789`
**Mode:** spawned bounded mutating worker / stacked sub-PR
**GSD command path:** `scripts/gsd doctor`; `scripts/gsd prompt plan-phase 401 --skip-research`; `scripts/gsd prompt programming-loop init --phase 401 --dry-run` failed with `scripts/gsd: unknown GSD command: programming-loop`, so `.pi/prompts/pm-gsd-loop.md` is the recorded manual GSD programming-loop fallback.

## Objective

Add an invocation-scoped `internal/config` package that loads typed app configuration through a fresh Viper instance with explicit flag/env/file/default precedence, reads `.polymetrics/config.yaml` from the invocation root, and integrates the CLI only far enough to validate malformed config through the existing validation exit-code funnel.

## Final correction cycle — PR #441 trace SHA and CLI docs caveat findings

Dispositions: Accepted with modification for trace SHA artifacts; accepted for CLI manual/docs caveat.

Required action for this final correction:

1. Stop treating predecessor commits as the final PR head in issue-local trace artifacts. Keep predecessor SHAs only as implementation/evidence checkpoints.
2. Record final-head source as `PR #441 headRefOid / git rev-parse HEAD at handoff`; the exact final SHA is intentionally not committed into trace artifacts and will be posted externally by the parent orchestrator after this last push.
3. Update `PROMPTS.md` downstream artifact/result so it is current and does not leave the final correction result unresolved.
4. Update `internal/cli/docs.go` and regenerated `docs/cli/config.md` so `root`/`json` are CLI-effective now while runtime/RLM/schedule command consumption remains on legacy readers until #402.
5. Keep website wording aligned; regenerate generated docs data with existing generators only.
6. Run the user-requested gates, commit/push one coherent final correction, update PR #441 body via API if needed, and do not request Claude/Copilot.

Final correction slice plan:

- FC0 plan gate: refresh issue-local GSD plan/TDD/verification artifacts before production docs edits; record `local_critical_path` because worker scope has no subagent tool.
- FC1 validation evidence: grep issue-local trace artifacts for stale predecessor-as-final-head claims and grep CLI/website config docs for contradictory legacy-reader wording.
- FC2 docs/trace correction: rewrite stale trace claims to predecessor checkpoints plus final-head source, align CLI manual/docs caveat with website wording, regenerate `docs/cli/config.md`, and run `node website/scripts/gen-docs-data.mjs`.
- FC3 verification: run `gofmt -w cmd internal`, `go test ./internal/config/... -count=1`, `go test ./internal/cli/ -run 'Golden|Config' -count=1`, `node website/scripts/gen-docs-data.mjs`, `make verify`, and `git diff --check origin/feat/cli-architecture-v2...HEAD`; also run `go vet ./...` and `go build ./cmd/pm` per local gate policy.
- FC4 delivery: commit/push the final correction; do not add a follow-up SHA-recording commit; PR body may record the exact post-push SHA externally.

## Final re-review cycle — PR #441 website caveat finding

Disposition: Accepted.

Finding: website config section lists runtime/RLM/schedule typed keys without the CLI manual's caveat that those commands still use legacy env readers until #402, potentially implying current command consumption.

Required action for this final review fix:

1. Add a concise caveat to `website/content/docs/cli-reference.mdx` that root/json config is CLI-effective now, while runtime/RLM/schedule reader migration remains owned by #402.
2. Regenerate `website/lib/docs.generated.ts` with `node website/scripts/gen-docs-data.mjs` only.
3. Run requested gates plus available website package-script checks without adding dependencies.
4. Update issue-local `SUMMARY.md`, `VERIFICATION.md`, `RUN-STATE.json`, `TDD-LEDGER.md`, and `PROMPTS.md` with gate evidence and predecessor checkpoints only; exact final head is sourced at handoff from `PR #441 headRefOid / git rev-parse HEAD`; update PR #441 body externally if needed; do not request Claude/Copilot.

Final review-fix slice plan:

- FRF0 plan gate: refresh issue-local GSD artifacts before production docs edit; record `local_critical_path` because worker scope has no subagent tool.
- FRF1 validation evidence: record that `website/content/docs/cli-reference.mdx` lacks the CLI manual caveat before the docs edit.
- FRF2 docs/data: add only the website caveat and regenerate `website/lib/docs.generated.ts` with the existing docs generator.
- FRF3 verification: run `node website/scripts/gen-docs-data.mjs`, `git diff --check origin/feat/cli-architecture-v2...HEAD`, `go test ./internal/config/... -count=1`, `go test ./internal/cli/ -run 'Golden|Config' -count=1`, `make verify`, plus existing website scripts that can run without dependency installs; record exact results.
- FRF4 delivery: commit/push to `feat/401-typed-viper-config`, update PR #441 body with the accepted disposition and a predecessor checkpoint if needed, and return compact handoff. Later corrections must not treat that checkpoint as the final PR head.

## Review-fix cycle — PR #441 pm-reviewer finding

Disposition: Accepted.

Finding: docs promise `POLYMETRICS_ROOT`/`PM_ROOT` choose config discovery root and `POLYMETRICS_JSON`/`PM_JSON` control invocation behavior, but CLI currently computes config path from flag-only root, calls `config.Load` only for validation, discards `Config`, and renders malformed-config errors with flag-only `jsonOut`.

Required behavior for this review fix:

1. Add one `internal/config` bootstrap resolver for `root` and `json` using bound flags > `POLYMETRICS_*` primary env > `PM_*` alias > defaults before config-file discovery.
2. Use the bootstrap root for `<effective-root>/.polymetrics/config.yaml`; explicit `--root` overrides env; file `root` does not relocate discovery for the same load.
3. Use bootstrap JSON for malformed-config error rendering; explicit `--json` overrides env; preserve the one JSON envelope on stdout and human diagnostics on stderr.
4. After successful load, invoke commands with resolved `Config.Root` and `Config.JSON` so flags > env > file > default precedence affects command behavior, not validation only.
5. Keep #402 env-reader migration out of scope and keep Viper instance-scoped with no package globals, `AutomaticEnv`, or `WatchConfig`.
6. Update config help/docs/website only as needed to clarify malformed-file JSON rendering and file-root non-relocation.

Review-fix slice plan:

- RF0 plan gate: update `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `SUMMARY.md`, and `RUN-STATE.json`; record accepted finding and `local_critical_path` decision.
- RF1 red tests: add failing `internal/config` bootstrap/discovery tests and failing `internal/cli` tests for `POLYMETRICS_ROOT`, `PM_ROOT`, `--root` override, `POLYMETRICS_JSON`/`PM_JSON` malformed-error JSON, explicit `--json` override, config-file `json`/`root` invocation, and isolation.
- RF2 green implementation: expose the small bootstrap resolver inside `internal/config`; have `Load` and CLI share it; invoke Cobra with loaded `Config.Root`/`Config.JSON`; preserve malformed-load validation classification and JSON envelope/stdout-stderr contract.
- RF3 docs parity: update `config` help/manual/website and regenerate existing website data only through `node website/scripts/gen-docs-data.mjs` if docs change.
- RF4 verification: run the user-required focused gates, full gates, diff checks, and config help/docs/website parity checks; commit and push the review-fix slice; update PR #441 body with disposition and evidence.

## Scope

Allowed writes:

- New `internal/config/**` typed configuration package and focused tests.
- Minimal `internal/cli/**` integration to load/bind config and map malformed config to validation exit 3 without changing `cli.Run(args, stdout, stderr) int` or legacy command behavior.
- `go.mod` / `go.sum` only for ADR-0002 approved Viper current stable v1 line and resolved expected transitives.
- Config documentation in `docs/cli/**` and matching website docs/data under existing website conventions.
- Issue-local `.planning/phases/401-typed-viper-config/**` artifacts.

Forbidden / out of scope:

- No parent/shared orchestration artifacts, parent PR body, `.planning/PROJECT.md`, `.planning/ROADMAP.md`, `.planning/STATE.md`, or orchestration ledger edits.
- No migration of scattered `os.Getenv` call sites; #402 owns consumption of the new config.
- No user-named credential env vars, certify credsfile env refs, or raw secret values in typed config docs/examples.
- No frontend dependency changes, TUI/env-migration/telemetry work, broad generated-file rewrites, connector bundle edits, generic write tools, reverse ETL execution, credentialed checks, quality gate reductions, or merges to `main`.

## Dependency decision

Selected exact approved dependency: `github.com/spf13/viper v1.21.0`.

Rationale: `go list -m -versions github.com/spf13/viper` shows `v1.21.0` as the latest stable v1 release. ADR 0002 explicitly approves Viper for Phase 3; the user task allows Viper v1 current stable and resolved transitives (`afero`, `cast`, `fsnotify`, `gotenv`, `toml`, `mapstructure`, `sourcegraph/conc`, and Viper-resolved support modules). Stop if `go get` introduces an additional direct module, a major-version deviation, frontend dependency changes, or unrelated graph changes.

## Required skills / references loaded

Skills loaded and applied:

- `gsd-core` — repo-local GSD/Pi adapter workflow.
- `caveman` — compact handoff prose only.
- `golang-how-to` — CLI config task routes to `golang-spf13-viper`, `golang-spf13-cobra`, `golang-cli`, `golang-testing`, `golang-security`, `golang-error-handling`, `golang-safety`, `golang-structs-interfaces`, and `golang-documentation`.
- `golang-cli` — config layering, exit-code preservation, stdout/stderr discipline, injected writer tests.
- `golang-testing` — named table tests, red/green evidence, invocation isolation tests.
- `golang-error-handling` — lower-case wrapped errors, `errors.As`, single existing CLI error funnel.
- `golang-security` — explicit env bindings, no unbounded env ingestion, no secret values in docs/tests.
- `golang-documentation` — CLI/manual/website parity and configuration docs.
- `golang-spf13-viper` — `viper.New()` per load, explicit `BindEnv`, optional config file, `Unmarshal`, no `AutomaticEnv`/`WatchConfig`.
- `golang-spf13-cobra` — fresh command tree, persistent flag binding, no shared Cobra/Viper state.
- `golang-structs-interfaces` — typed config struct design and mapstructure tags.
- `golang-safety` — zero-value/default safety, no state leak across invocations.
- `vercel-react-best-practices` / `vercel-composition-patterns` loaded only for website-doc awareness; no React component changes are planned.

Rule references to cite in PR/handoff:

- `.agents/agentic-delivery/references/required-skills-routing.md`: **Always-on Go skill routing** and **CLI and command behavior** require `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-documentation`; Viper/Cobra code additionally requires `golang-spf13-viper` and `golang-spf13-cobra`.
- `golang-spf13-viper`: precedence pipeline; ConfigFileNotFoundError is non-fatal; `viper.New()` for test isolation; explicit binding and mapstructure tags.
- `golang-spf13-cobra`: fresh command tree per run/test and injected writers.
- `golang-testing`: Best Practices #1 named subtests, #3 independent tests, #5 observable behavior.
- `golang-security`: Security Thinking Model #1-#3 for env/filesystem boundaries; avoid secrets in config docs/examples.
- `golang-safety`: Best Practices #10 useful zero values and #11 no ambient shared mutable state.
- `golang-error-handling`: Best Practices #2 wrapping, #5 chain inspection, #7 single handling rule.

Missing repo-local stack skill: `.pi/skills/go-implementation/SKILL.md` returned `ENOENT`; `.pi/skills/ts-website/SKILL.md` also returned `ENOENT`. Recorded as missing repo-local skill artifacts, not a blocker because user-required Go skills above were loaded.

## Config key model

`internal/config.Config` will cover invocation/app keys without consuming them in existing call sites yet:

- `root` / `json`: current Cobra global flags (`--root`, `--json`) bound when a flag set is supplied. Config file root does not relocate the config-file discovery root for the same invocation.
- `version`, `project`, `warehouse.connector`, `warehouse.path`: current `.polymetrics/config.yaml` writer keys/defaults.
- `runtime.postgres_url`, `runtime.dragonfly_addr`, `runtime.temporal_addr`: current runtime config-shaped env readers.
- `rlm.image`, `rlm.podman_bin`, `rlm.fake_runner`, `rlm.embedded_worker`: current RLM agent/runtime toggles and Podman image/bin config.
- `rlm.llm.provider`, `rlm.llm.base_url`, `rlm.llm.model`: non-secret LLM client config. Secret API-key env vars remain outside config-file examples and are documented as env-only security boundaries.
- `schedule.crontab_file`: current local crontab redirection seam (`PM_CRONTAB_FILE`) with `POLYMETRICS_CRONTAB_FILE` primary binding.

Every non-secret key gets explicit `POLYMETRICS_*` primary env binding with documented `PM_*` legacy alias fallback; no `AutomaticEnv`.

## Slice plan

### Slice 0 — Plan/TDD setup

1. Confirm branch `feat/401-typed-viper-config` is based on `origin/feat/cli-architecture-v2` at `8900db141cc289b65491365d2ebcab490af57789`.
2. Create issue-local `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `SUMMARY.md`, `PROMPTS.md`, and `RUN-STATE.json` before production edits.
3. Record GSD command output and `programming-loop` fallback.
4. Commit/push the planning checkpoint.

### Slice 1 — Red tests before production edits

1. Add failing `internal/config` table tests for:
   - default values;
   - missing config file is non-error;
   - file > default for every file-backed key;
   - explicit `POLYMETRICS_*` env > file for every env-backed key;
   - `PM_*` alias fallback when primary env is absent;
   - bound `--root` and `--json` flags > env/file;
   - malformed YAML returns a typed config load error;
   - invocation isolation/no state leak across roots/env/flag sets;
   - unbound env vars do not affect config.
2. Add failing CLI-focused tests for malformed `.polymetrics/config.yaml` -> validation exit 3 / JSON error category `validation`, and missing file preserving success for root manual.
3. Run red commands and record exact output:
   - `go test ./internal/config/... -count=1`
   - `go test ./internal/cli/ -run Config -count=1`

### Slice 2 — Minimal green implementation

1. Add `github.com/spf13/viper@v1.21.0` and run `go mod tidy`; inspect dependency delta.
2. Implement `internal/config`:
   - `viper.New()` inside `Load` only;
   - `SetConfigFile(filepath.Join(invocationRoot, ".polymetrics", "config.yaml"))`;
   - missing file non-error;
   - malformed/unmarshal errors wrapped in typed load error;
   - explicit defaults, `BindEnv` with primary + alias names, and `BindPFlag` for current Cobra global flags;
   - no `AutomaticEnv`, no `WatchConfig`, no package singleton.
3. Integrate in `cli.Run` after `parseGlobal`/fresh root command creation: load config for validation/binding only; map config load errors to `validationErrorf` through `writeError`. Keep legacy `root`/`jsonOut` behavior from `parseGlobal` so goldens/certify stay byte-identical.
4. Keep existing scattered `os.Getenv` readers unchanged.

### Slice 3 — Docs and parity

1. Update embedded manual/docs map with config topic or runtime/security sections as minimally required.
2. Regenerate/update `docs/cli/**` only for changed embedded manual pages, preserving generated format.
3. Update website docs under existing conventions (`website/content/docs/cli-reference.mdx`; generated website data only if existing convention requires it).
4. Document keys, defaults, precedence, aliases, file format, and secret boundaries; no secret values.
5. Verify runtime help/bare namespace/command help and docs/website grep parity.

### Slice 4 — Verification, PR, review route

Run required gates:

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

Additional parity:

- Build local binary and run `pm help <topic>` before unfamiliar commands.
- Check `pm help config` (or documented topic), `pm runtime --help`, `pm runtime`, and a bare namespace command not otherwise changed.
- Grep `docs/cli/**` and `website/**` for config keys/aliases.
- Record generated help/manual artifact status; regenerate only if parity requires.

Commit/push coherent plan/red/green/refactor checkpoints. Open non-draft stacked PR to `feat/cli-architecture-v2` with Conventional Commit title, `Refs #401`, `Refs #397`, dependency delta, GSD/TDD/skills/parity/gate evidence. Claude workflow remains `disabled_manually`; Copilot quota exhausted. Do not post `@claude review`; do not request Copilot. Record human/parent fallback as delegated with no approval claim.

## Spawn decision for this cycle

Initial implementation cycle: `spawned` — parent #397 assigned this isolated worker directory, branch, issue #401, and bounded write scope. This worker does not spawn subagents.

Review-fix cycle: `local_critical_path` — same isolated worktree and branch; no subagent tool in worker scope; issue #401 review fix is within the bounded `internal/config`, `internal/cli`, docs, and issue-local planning scope.

Final re-review caveat cycle: `local_critical_path` — same isolated worktree and branch; no subagent tool in worker scope; docs-only website caveat plus generated website data and issue-local trace updates are on the critical path for PR #441.

## Human gates

- No secrets, credential prompts, credentialed connector checks, or secret values in docs/tests.
- No dependency deviation beyond ADR 0002 Viper current stable v1 and resolved transitives.
- No additional direct modules, major-version deviation, frontend dependencies, or unrelated graph changes.
- No generic shell/HTTP/SQL write tools.
- No reverse ETL execution outside existing test/smoke gates.
- No quality gate reductions.
- No merge to `main` or parent PR merge.
