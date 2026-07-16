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
