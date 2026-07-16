# Verification — Issue 399 Golden Transcript Safety Net

## Required commands

```bash
scripts/gsd doctor
scripts/gsd prompt plan-phase 399 --skip-research >/tmp/gsd-plan-phase-399.prompt
test -s /tmp/gsd-plan-phase-399.prompt
scripts/gsd prompt programming-loop init --phase 399 --dry-run >/tmp/gsd-programming-loop-399.prompt
```

Expected programming-loop result: command exits non-zero with `scripts/gsd: unknown GSD command: programming-loop`; record Pi-local `.pi/prompts/pm-gsd-loop.md` fallback instead of skipping TDD/planning.

## Red/absent evidence commands

```bash
go test ./internal/cli/ -run Golden -count=1
rg -n "Golden|golden transcript|docs generation diff|docs-generate-diff" internal/cli
```

## Phase gate

```bash
go test ./internal/cli/ -run Golden
```

## Full local gates

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
git diff -- go.mod
```

## CLI parity checks

Because #399 adds tests for existing CLI behavior and should not change runtime help/manual/website content, parity is verified as no behavior/docs drift plus spot checks:

```bash
go build -o /tmp/pm-399 ./cmd/pm
/tmp/pm-399 help docs
/tmp/pm-399 connectors
/tmp/pm-399 docs --help
rg -n "docs|connectors" docs/cli website
```

Mark website docs as not applicable if no CLI-visible docs behavior changes occur.

## Review-fix requested commands

```bash
gofmt -w internal/cli
go test ./internal/cli/ -run Golden -count=1
go test ./internal/cli/ -count=1
make verify
git diff --check
git diff -- go.mod
```

Review-fix scope notes:

- Do not change production dispatcher behavior.
- `pm connectors help <name>` golden records current legacy namespace-help interception only; behavior/docs cleanup is deferred to #417.
- Docs-generation diff intentionally compares generated top-level CLI manuals against `docs/cli/**`; connector manuals are generated to a temp directory only to avoid repository writes during the test.
- Claude workflow is `disabled_manually`; do not post `@claude review`; coverage route remains `parent_pr_fallback` pending/blocked.

## Expected results

- GSD doctor exits 0.
- Plan-phase prompt is non-empty.
- Programming-loop shell command unavailable and fallback recorded.
- Golden suite initially absent before edits, then present and green.
- Docs generation test writes only to a temp directory and diffs cleanly against `docs/cli/**`.
- No ANSI escapes appear in pinned stdout/stderr.
- No credentialed checks run.
- `git diff -- go.mod` has no output.
- Sub-PR targets `feat/cli-architecture-v2`, not `main`.

## Results log

### Red / absent evidence

```text
ok  	polymetrics.ai/internal/cli	0.525s [no tests to run]
go_test_status=0
rg_status=1
```

### Docs-diff red after adding test

`TestGoldenDocsGenerateMatchesTrackedCLIManuals` initially failed because `docs/cli/connectors.md` contained a stale `GITHUB CERTIFICATION` block that `pm docs generate` does not emit.

### Targeted green

```bash
gofmt -w cmd internal && go test ./internal/cli/ -run Golden -count=1
```

```text
ok  	polymetrics.ai/internal/cli	6.316s
```

### Focused package

```bash
go test ./internal/cli/ -count=1
```

```text
ok  	polymetrics.ai/internal/cli	154.101s
```

### Full gates

```bash
go vet ./...
go test ./...
go build ./cmd/pm
make verify
git diff -- go.mod
git diff --check
```

Results:

- `go vet ./...`: pass (no output).
- `go test ./...`: pass; slowest packages included `internal/connectors/certify` 364.209s and `internal/cli` 163.133s.
- `go build ./cmd/pm`: pass (no output).
- `make verify`: pass; finished through fmt, tidy-check, vet, test, build, docs-check, smoke, lint, and `connectorgen validate` (`connectorgen validate: 547 connector(s) checked, 0 findings`).
- `git diff -- go.mod`: pass (no output).
- `git diff -- go.sum`: pass (no output).
- `git diff --check`: pass (no output).

### CLI parity spot checks

```bash
./pm help docs >/tmp/pm-help-docs-399.out
./pm connectors >/tmp/pm-connectors-399.out
./pm docs --help >/tmp/pm-docs-help-399.out
rg -n "pm docs|pm connectors" docs/cli website | head -40 >/tmp/pm-docs-website-grep-399.out
```

Results:

- `./pm help docs`: pass; output begins `NAME\n  pm docs - generate CLI documentation`.
- `./pm connectors`: pass; bare namespace prints contextual manual and exits 0.
- `./pm docs --help`: pass; output begins `NAME\n  pm docs - generate CLI documentation`.
- docs/website grep: pass; CLI docs and website generated docs reference `pm docs` / `pm connectors`.
- Website docs: no source `website/content/**` update needed because #399 adds tests and aligns tracked generated CLI markdown to existing runtime help; no new CLI command/flag/output behavior introduced.

### Review-fix disposition log

Dispositions completed:

- MEDIUM `pm connectors help <name>` golden ambiguity: accepted with modification; renamed/annotated as known legacy namespace-help interception and deferred dispatcher/help-tree change to #417.
- LOW connector-manual recursive docs comparison: declined for #399 scope; added clarifying test comment that connector manuals generate to temp dir to avoid repository writes while comparison remains `docs/cli/**`.
- LOW RUN-STATE allowed paths: accepted; included `docs/cli/**` / `docs/cli/connectors.md` in scope evidence.

Review-fix gates:

```text
gofmt -w internal/cli
go test ./internal/cli/ -run Golden -count=1
ok  	polymetrics.ai/internal/cli	6.257s

go test ./internal/cli/ -count=1
ok  	polymetrics.ai/internal/cli	155.838s

make verify
PASS: completed fmt, tidy-check, vet, full tests, build, docs validation, smoke, lint, and connectorgen validate: 547 connector(s) checked, 0 findings.

git diff --check
(no output)

git diff -- go.mod
(no output)
```

### Commit-range whitespace correction — 2026-07-16

Coordinator found the previous `git diff --check` claim was worktree-only evidence and did not
cover committed PR-range whitespace. Correct red evidence before the fix:

```text
git diff --check origin/feat/cli-architecture-v2...HEAD
.planning/phases/399-golden-transcript-safety-net/PLAN.md:3: trailing whitespace.
.planning/phases/399-golden-transcript-safety-net/PLAN.md:4: trailing whitespace.
.planning/phases/399-golden-transcript-safety-net/PLAN.md:5: trailing whitespace.
.planning/phases/399-golden-transcript-safety-net/PLAN.md:6: trailing whitespace.
.planning/phases/399-golden-transcript-safety-net/PLAN.md:7: trailing whitespace.
.planning/phases/399-golden-transcript-safety-net/PROMPTS.md:7: trailing whitespace.
.planning/phases/399-golden-transcript-safety-net/PROMPTS.md:8: trailing whitespace.
.planning/phases/399-golden-transcript-safety-net/PROMPTS.md:9: trailing whitespace.
.planning/phases/399-golden-transcript-safety-net/PROMPTS.md:48: trailing whitespace.
status=2
```

Correction scope: remove trailing whitespace / whitespace-only blank-line defects from all files
changed by PR #439. No production behavior change, dependency change, review-request change, or PR
merge.

Requested post-fix gates:

```bash
git diff --check origin/feat/cli-architecture-v2...HEAD
git diff --name-only origin/feat/cli-architecture-v2...HEAD
go test ./internal/cli/ -run Golden -count=1
git diff -- go.mod
```

Results: pending until the fix is committed so the commit-range gate includes the correction.

### PR / CI / review status

- Branch pushed: `test/399-golden-transcript-safety-net`.
- Sub-PR opened: https://github.com/polymetrics-ai/cli/pull/439.
- Pre-review-fix head SHA: `d7ffbb1ee01b709a3470f62976cba65c2c586921`; review-fix commit SHA is recorded in worker handoff after commit/push.
- GitHub checks observed before review fix: branch-name, pr-title, require-linked-issue, gsd-workflow-evidence, Dependency Review, govulncheck, CodeQL, and verify were running/passing per PR view; review-fix CI will run after push.
- Claude review route: workflow `Claude Code Review` is `disabled_manually`; no Claude approval claimed. Parent-PR fallback coverage remains pending/blocked.
