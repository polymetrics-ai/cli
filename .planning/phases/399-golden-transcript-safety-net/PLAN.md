# Issue 399 Plan — Golden Transcript Safety Net

**Issue:** [#399](https://github.com/polymetrics-ai/cli/issues/399)  
**Parent:** [#397](https://github.com/polymetrics-ai/cli/issues/397)  
**Parent PR:** [#438](https://github.com/polymetrics-ai/cli/pull/438) (`feat/cli-architecture-v2` → `main`, draft)  
**Worker branch:** `test/399-golden-transcript-safety-net`  
**Mode:** spawned bounded mutating worker / stacked sub-PR  
**GSD command path:** `scripts/gsd doctor`; `scripts/gsd prompt plan-phase 399 --skip-research`; `scripts/gsd prompt programming-loop init --phase 399 --dry-run` failed with `scripts/gsd: unknown GSD command: programming-loop`, so the Pi-local `.pi/prompts/pm-gsd-loop.md` contract is the recorded manual GSD programming-loop fallback.

## Objective

Capture the legacy CLI contract before the Cobra/Viper strangler migration by adding a golden transcript test harness and a docs generation diff test. The safety net pins exit code, stdout, and stderr for representative invocations while preserving the current dispatcher unchanged.

## Scope

Allowed writes:

- Issue-local GSD artifacts under `.planning/phases/399-golden-transcript-safety-net/`.
- Golden transcript harness/fixtures under `internal/cli/**`.
- Docs-generate-diff test for `docs/cli/**`.
- Minimal tracked `docs/cli/**` drift fix only when required for the docs-generate-diff test to pass.

Forbidden / out of scope:

- No production dispatcher changes.
- No connector bundle changes.
- No shared parent orchestration artifacts, parent PR body, `.planning/PROJECT.md`, `.planning/ROADMAP.md`, or `.planning/STATE.md` edits.
- No `go.mod` / dependency changes.
- No credentialed connector checks or reverse ETL execution.

## Required skills / references loaded

Skills loaded and applied:

- `golang-how-to` — Always-on Go skill routing for Go implementation/test tasks.
- `golang-cli` — Required for `pm` CLI commands, flags, output, exit codes, stdout/stderr discipline, and command tests.
- `golang-testing` — Required for table/golden tests and deterministic fixtures.
- `golang-error-handling` — Required for stable error taxonomy and stderr behavior.
- `golang-documentation` — Required for generated docs/manual parity tests.
- `golang-security` — Loaded because tests handle command args, filesystem temp output, and external-ish CLI IO boundaries.
- `golang-safety` — Loaded for review-fix hardening checks around fixture/test safety.
- `golang-lint` — Loaded for review disposition/gate interpretation.
- `caveman` — Loaded for compact worker handoff prose.
- `gsd-core` — Repo-local GSD/Pi adapter guidance.

Rule references: `.agents/agentic-delivery/references/required-skills-routing.md` sections **Always-on Go skill routing**, **CLI and command behavior**, **Documentation for Go behavior**, and **Reviews and hardening**; `.agents/agentic-delivery/references/cli-help-docs-website-parity.md` parity checklist.

## Slice plan

### Slice 0 — Plan/TDD setup

1. Confirm branch, parent PR, issue #399, parent #397, and write scope.
2. Create issue-local `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `RUN-STATE.json`, `SUMMARY.md`, and `PROMPTS.md` before test harness edits.
3. Record GSD command output and programming-loop fallback.
4. Commit/push the planning checkpoint if clean and useful.

### Slice 1 — Red/absent evidence

1. Run `go test ./internal/cli/ -run Golden -count=1` before adding golden tests.
2. Grep for existing golden transcript/docs-diff tests to prove absence.
3. Record exact evidence in `TDD-LEDGER.md`.

Expected red/absent result: no `Golden` tests execute and no docs-generate-diff test exists.

### Slice 2 — Golden transcript harness

1. Add a table-driven `TestGoldenTranscripts` in `internal/cli`.
2. Cover about 80 invocations across:
   - root/help/man/manual paths;
   - bare namespace help and `--help`/`-h` variants;
   - JSON manual/envelope paths;
   - flag forms (`--flag=value`, `--flag value`, repeated flags, bare boolean-style flags, `--root` both forms, late `--json`);
   - error categories and stable exit codes;
   - dynamic connector dispatch failures;
   - hidden/internal commands such as `extract` and `worker`.
3. Pin exit code, stdout, and stderr for every case. Keep expected output generated in-process as Go fixtures, not broad generated docs rewrites.
4. Ensure no ANSI escapes in captured output.

### Slice 3 — Docs generation diff test

1. Add a test that runs CLI docs generation into `t.TempDir()`.
2. Compare generated files and contents against tracked `docs/cli/**`.
3. Fail with a useful diff summary when docs drift.
4. If existing tracked docs drift from generator output, apply the minimal `docs/cli/**` generated-doc fix and record it in the TDD ledger.

### Slice 4 — Verification, parity, PR

1. Run required gates:
   - `go test ./internal/cli/ -run Golden`
   - `gofmt -w cmd internal`
   - `go vet ./...`
   - `go test ./...`
   - `go build ./cmd/pm`
   - `make verify`
   - `git diff -- go.mod`
2. Run CLI parity spot checks without credentials:
   - `go build -o /tmp/pm-399 ./cmd/pm`
   - `/tmp/pm-399 help docs`
   - `/tmp/pm-399 connectors`
   - `/tmp/pm-399 docs --help` or equivalent current dispatcher help behavior
   - docs/website grep or mark website not applicable because this issue adds tests only and does not change CLI-visible docs.
3. Update phase artifacts with real results.
4. Commit and push green slice to `origin/test/399-golden-transcript-safety-net`.
5. Open stacked sub-PR to `feat/cli-architecture-v2` with `Refs #399` and `Refs #397`.
6. Record automated review route/status; do not merge.

## Spawn decision for this cycle

`spawned`: parent issue #397 assigned this isolated worker directory, branch, and issue-scoped write scope for #399. This worker does not spawn subagents.

## Review-fix cycle — 2026-07-16

Sidecar review dispositions:

1. **MEDIUM accepted with modification** — rename/annotate the `pm connectors help <name>` golden so it records a known legacy namespace-help interception, not an endorsed connector-manual alias contract. Do not change dispatcher behavior. Defer behavior/docs cleanup to help-tree issue #417.
2. **LOW declined for scope** — keep docs-generation diff scoped to `docs/cli/**`; add a concise test comment that connector manuals are generated to a temp dir only to avoid repository writes.
3. **LOW accepted** — make `RUN-STATE.json` allowed-path evidence include `docs/cli/**` / `docs/cli/connectors.md` because the original slice changed that tracked generated CLI doc.

Review-fix plan:

1. Update issue-local phase artifacts with dispositions and scope notes before test/fixture edits.
2. Rename the ambiguous golden case name in `internal/cli/golden_transcript_test.go` and `internal/cli/testdata/golden_transcripts.json`; keep args/exit/stdout/stderr unchanged.
3. Add a clarifying comment in the docs-generation diff test; do not expand comparison beyond `docs/cli/**`.
4. Run requested gates: `gofmt -w internal/cli`, `go test ./internal/cli/ -run Golden -count=1`, `go test ./internal/cli/ -count=1`, `make verify`, `git diff --check`, and `git diff -- go.mod`.
5. Commit/push one coherent review-fix commit and update PR #439 body with dispositions. Do not post `@claude review`; coverage route remains `parent_pr_fallback` pending/blocked.

Review-fix execution decision: `local_critical_path` — existing spawned worker applies bounded review fixes locally; no subagents spawned by worker role.

## Human gates

- No secrets, credential prompts, or credentialed connector checks.
- No dependencies or `go.mod` edits.
- No production dispatcher changes.
- No generic shell/HTTP/SQL write tools.
- No reverse ETL execution.
- No merge to `main` or parent PR merge.
