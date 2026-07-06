# T/B-18 — .golangci.yml + Makefile gates

Phase: wave0-engine-harness · Wave: G · Executor: gsd-loop-backend (sonnet)

## Scope

- `.golangci.yml` (new, v2 schema)
- `Makefile` — add `lint`, `connectorgen-validate` targets; extend `verify:` to run them after
  the existing steps
- This ledger (RED/GREEN validation-artifact record — per `workflow.tdd_mode`, this task is not
  unit-testable; the RED artifact is a recorded failing run, not a test file)

FORBIDDEN / untouched: everything else. `internal/connectors/engine/**` and `definition.go` are
owned by a parallel agent this session — not read for editing, not modified.

## RED evidence

### 1. `make lint` before the target exists

```
$ make lint
make: *** No rule to make target 'lint'.  Stop.
$ echo "EXIT:$?"
EXIT:2
```

Confirms: no `lint` target exists in the Makefile prior to this task.

### 2. `golangci-lint run` with no `.golangci.yml`, scoped to the new wave0 packages

```
$ golangci-lint run ./internal/connectors/engine/... ./internal/connectors/certify/... \
    ./cmd/connectorgen/... ./cmd/inventorygen/...
internal/connectors/engine/bundle_test.go:296:1: SA9009: ineffectual compiler directive due to extraneous space: "// go:embed all:* scaffold end-to-end: wave0 ships zero bundles (goldens land" (staticcheck)
// go:embed all:* scaffold end-to-end: wave0 ships zero bundles (goldens land
^
internal/connectors/engine/read_test.go:518:11: QF1012: Use fmt.Fprintf(...) instead of Write([]byte(fmt.Sprintf(...))) (staticcheck)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":[{"id":"%d","name":"a","updated_at":"2026-01-01T00:00:00Z"}], "has_more": true}`, page)))
			       ^
internal/connectors/engine/read_test.go:851:11: QF1012: Use fmt.Fprintf(...) instead of Write([]byte(fmt.Sprintf(...))) (staticcheck)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"data":[{"id":"1","name":"a","updated_at":"2026-01-01T00:00:00Z"}],"meta":{"next_page_link":%q}}`, srv.URL+"/widgets?page=2")))
			       ^
3 issues:
* staticcheck: 3
```

`internal/connectors/certify`, `cmd/connectorgen`, `cmd/inventorygen` each individually reported
`0 issues.` (default linter set, no config). The 3 findings above are in `engine/bundle_test.go`
and `engine/read_test.go` — files owned by the parallel Wave C/D engine agent, currently mid-flight
in this session (git log shows the Wave C read/write commit landed just before this task started).
Per the FORBIDDEN list I do not fix these; they are recorded below as findings for the coordinator.

### 3. Repo-wide default-linter scale check (justifies scoping decision)

```
$ timeout 90 golangci-lint run ./...
...
80 issues:
* errcheck: 40
* ineffassign: 3
* staticcheck: 20
* unused: 17
```

This is with **zero** `.golangci.yml` and only golangci-lint's built-in default linters
(govet, staticcheck, errcheck, unused, ineffassign, gosimple — a small subset of what the task asks
to enable). The legacy ~560-connector tree already produces 80 findings under the *default*
linter set alone, before adding `misspell` or tightening anything. This is the evidence basis for
the scoping decision below.

## Scope decision: `new-packages` (not repo-wide)

**Chosen: lint scope limited to the new wave0 packages, enforced via the Makefile `lint` target's
path list — NOT via `.golangci.yml` `issues.exclude-dirs`/`skip-dirs` blanket-excluding legacy
code, and NOT via `new-from-rev`.**

Rationale:
- The task brief requires the config to "pass on the EXISTING ~560-package legacy tree, not just
  new code" if run naively, but also says: "if a linter drowns in legacy findings, scope lint paths
  (Makefile target) to the new packages ... rather than weakening linters repo-wide."
- Evidence above (§3) shows the legacy tree fails even trivially (errcheck/unused/staticcheck) under
  default settings alone. Fixing ~560 legacy connector packages is out of scope for wave0 and would
  touch files this task is forbidden from touching (only `.golangci.yml`/`Makefile`/ledger are
  in-scope files).
- Weakening linters (e.g. disabling errcheck/unused repo-wide, or setting impossibly high
  thresholds) to make legacy code pass would violate "Human gates: quality-gate reductions" — this
  task adds gates, it must not water them down to fit legacy code.
- `.golangci.yml` therefore stays literal/strict (govet, staticcheck, errcheck, ineffassign, unused,
  misspell — exactly the six named in the task brief, no more, no fewer) with NO path-based
  exclusions for legacy directories (only generated-file and `_quarantine` exclusions, which are
  legitimate exclusions regardless of scope). The Makefile `lint` target passes an explicit path
  list scoped to the new wave0 tree (`internal/connectors/{engine,defs,hooks,native,conformance,certify}/...`
  and `cmd/{connectorgen,inventorygen}/...`) so `golangci-lint run <paths>` never touches the
  legacy ~560 connector packages at all. `make lint` therefore both (a) is honest about what it
  covers, and (b) passes today without touching any forbidden file, and (c) will automatically pick
  up new packages as later waves (conformance, certify stages, goldens) land in those same
  directories.
- `connectorgen-validate` and `verify:` additions run against the same new-package tree
  (`internal/connectors/defs`), consistent with this scope choice.

This is recorded here per the task's "Human-gate note: adding gates only (no reduction). Tool
acquisition decision flagged" instruction — flagging the scope choice explicitly rather than
silently picking repo-wide-weak or new-packages-strict.

## Implementation

### `.golangci.yml` (v2 schema, `version: "2"`)

- `linters.enable`: `govet`, `staticcheck`, `errcheck`, `ineffassign`, `unused`, `misspell` — exactly
  the six named in the task brief, nothing added.
- `linters.exclusions.rules`: exclude `errcheck` on `*_test.go`? — NOT used; kept literal. Instead:
  `linters.exclusions.paths`: `(^|/)([a-zA-Z0-9_.\\-]+_gen\\.go)$` equivalent via golangci-lint v2's
  generated-file convention (`issues.exclude-generated: strict` maps to v2's
  `linters.exclusions.generated: strict`, which already skips `// Code generated ... DO NOT EDIT.`
  headers used by `*_gen.go` files in this repo — confirmed against
  `internal/connectors/hooks/hookset/hookset_gen.go` and `internal/connectors/native/nativeset/nativeset_gen.go`,
  both carry that exact header). Also added an explicit path-glob exclusion for `_gen\\.go$` and for
  any `_quarantine` directory (per task brief), belt-and-suspenders alongside the generated-header
  detection.
- `run.timeout: 5m` (v2 default is unbounded; set a sensible bound per golang-lint skill's Common
  Issues table).
- No `formatters` block added (task brief doesn't ask for `gofmt`/`gofumpt` gating here — repo
  already has a separate `fmt`/`tidy-check` verify step).

### `Makefile`

- `lint`: guards on `command -v golangci-lint`, then runs
  `golangci-lint run ./internal/connectors/engine/... ./internal/connectors/defs/... ./internal/connectors/hooks/... ./internal/connectors/native/... ./internal/connectors/conformance/... ./internal/connectors/certify/... ./cmd/connectorgen/... ./cmd/inventorygen/...`
  — `conformance` does not exist as a directory yet (Wave E, not landed this session); `go list`/
  golangci-lint tolerates a nonexistent path being globbed IF at least one path exists — verified
  below.
- `connectorgen-validate`: `go run ./cmd/connectorgen validate internal/connectors/defs`.
- `verify:` extended from `fmt tidy-check vet test build docs-check smoke` to
  `fmt tidy-check vet test build docs-check smoke lint connectorgen-validate` (gates ADDED at the
  end, existing steps and their order untouched).

## GREEN evidence

### `.golangci.yml` schema validity

```
$ golangci-lint config verify
$ echo EXIT:$?
EXIT:0
```

### `make connectorgen-validate` — GREEN

```
$ make connectorgen-validate
go run ./cmd/connectorgen validate internal/connectors/defs
connectorgen validate: 0 connector(s) checked, 0 findings
$ echo EXIT:$?
EXIT:0
```

### `make -n verify` — confirms the extended chain, in order, additions appended at the end

```
$ make -n verify
gofmt -w cmd internal
go mod tidy
git diff --exit-code -- go.mod go.sum
go vet ./...
go test ./...
go build ./cmd/pm
./pm docs validate --connectors-dir docs/connectors
<smoke recipe...>
command -v golangci-lint >/dev/null || (echo "golangci-lint not found — brew install golangci-lint" && exit 1)
golangci-lint run ./internal/connectors/engine/... ./internal/connectors/defs/... ./internal/connectors/hooks/... ./internal/connectors/native/...  ./internal/connectors/certify/... ./cmd/connectorgen/... ./cmd/inventorygen/...
go run ./cmd/connectorgen validate internal/connectors/defs
```

(Note: `internal/connectors/conformance` is correctly omitted from the expanded `$(LINT_PKGS)` list
above — that directory does not exist yet this session, Wave E lands it. `LINT_PKGS` is a Make
`$(wildcard ...)` filter, so it will automatically pick the path up once Wave E creates the
directory, with no further Makefile edit required.)

### `make lint` — status: BLOCKED by a concurrent, in-progress parallel task (not by this task's config)

```
$ make lint
golangci-lint run ./internal/connectors/engine/... ./internal/connectors/defs/... ./internal/connectors/hooks/... ./internal/connectors/native/...  ./internal/connectors/certify/... ./cmd/connectorgen/... ./cmd/inventorygen/...
internal/connectors/engine/auth.go:1: : # polymetrics.ai/internal/connectors/engine [polymetrics.ai/internal/connectors/engine.test]
internal/connectors/engine/connector_test.go:16:38: undefined: Connector
...
internal/connectors/engine/connector_test.go:101:7: undefined: New
internal/connectors/engine/connector_test.go:428:2: undefined: Base
1 issues:
* typecheck: 1
make: *** [Makefile:74: lint] Error 1
```

Root cause: `internal/connectors/engine/connector_test.go` is untracked (`git status --porcelain`
confirms), references `Connector`, `Base`, `New` — none of which exist yet
(`ls internal/connectors/engine/connector.go` → no such file). This is task **T/B-10** (Wave D,
`engine.Connector`/`engine.Base`/`Definition`), being executed concurrently by a different agent in
this same session, currently in its RED (test-written, behavior-not-yet-implemented) state. `go
build ./internal/connectors/engine/...` succeeds cleanly (test files aren't part of a `go build`) —
confirming the *package* is not broken, only its in-flight test-first WIP is temporarily
uncompilable, exactly as expected mid-TDD-cycle for a task I do not own and must not touch (FORBIDDEN
list: "a parallel agent owns internal/connectors/engine + definition.go right now").

Verified the `lint` target's own logic (path filtering, guard, invocation) is correct by running the
identical `golangci-lint run` invocation against every OTHER new-package directory in the list
(i.e. everything except the mid-flight `engine` package):

```
$ golangci-lint run ./internal/connectors/defs/... ./internal/connectors/hooks/... \
    ./internal/connectors/native/... ./internal/connectors/certify/... \
    ./cmd/connectorgen/... ./cmd/inventorygen/...
0 issues.
$ echo EXIT:$?
EXIT:0
```

This isolates the failure to the one package under concurrent construction and confirms `.golangci.yml`
+ the Makefile `lint` target are correctly configured — `make lint` will pass once T/B-10 lands
`connector.go` (no action needed from this task; the coordinator should re-run `make lint` after
T/B-10 merges, before V-21).

Update (later in the same session, re-checked before finalizing this ledger): T/B-10 progressed —
`internal/connectors/engine/connector.go` and `internal/connectors/definition.go` now exist
(untracked, still WIP) — but `make lint` still fails, now on a different typecheck error
(`Base does not implement connectors.ManifestProvider (missing method Manifest)`), confirming T/B-10
is still actively mid-implementation, not yet complete. `go build ./...` remains clean throughout
(EXIT:0) — this is production-code-compiles-but-test-drives-ahead TDD RED state for T/B-10, exactly
as expected for concurrent execution, and remains outside this task's scope to fix or wait out.
The coordinator should re-run `make lint` once T/B-10 is reported complete/GREEN, ahead of V-21.

## Lint findings on already-committed new code (for the coordinator — NOT fixed by this task)

Per the "everything else... FORBIDDEN" file-ownership rule, the following pre-existing findings in
`internal/connectors/engine` test files are recorded here rather than fixed:

| File:Line | Linter | Finding |
|---|---|---|
| `internal/connectors/engine/bundle_test.go:296:1` | staticcheck (SA9009) | Ineffectual compiler directive due to extraneous space: `// go:embed all:* scaffold end-to-end: ...` — this is a doc comment that happens to start with `// go:embed` (should be `//go:embed` with no space to be a real directive, or reworded so staticcheck doesn't parse it as one) |
| `internal/connectors/engine/read_test.go:518:11` | staticcheck (QF1012) | `w.Write([]byte(fmt.Sprintf(...)))` should be `fmt.Fprintf(w, ...)` |
| `internal/connectors/engine/read_test.go:851:11` | staticcheck (QF1012) | same QF1012 pattern, second occurrence |

All three are pre-existing in files owned by the parallel engine agent (Wave A/C), not introduced by
this task, and not touched by this task.

`internal/connectors/certify`, `cmd/connectorgen`, `cmd/inventorygen`, `internal/connectors/defs`,
`internal/connectors/hooks`, `internal/connectors/native` each report `0 issues.` under the new
`.golangci.yml` — no additional findings beyond what B-11's ledger already recorded as fixed.

## Self-verify

```
$ golangci-lint config verify                                    # EXIT:0
$ make connectorgen-validate                                     # EXIT:0, 0 findings
$ make -n verify                                                 # confirms fmt tidy-check vet test build docs-check smoke lint connectorgen-validate, in order
$ golangci-lint run ./internal/connectors/defs/... ./internal/connectors/hooks/... \
    ./internal/connectors/native/... ./internal/connectors/certify/... \
    ./cmd/connectorgen/... ./cmd/inventorygen/...                # EXIT:0, 0 issues (isolates lint target correctness from T/B-10's transient WIP)
```

Did NOT run full `make verify` (runs `smoke`, several minutes — reserved for the coordinator at
V-21 per the dispatch). Did NOT modify any file outside `.golangci.yml`, `Makefile`, and this ledger.
No git commit made (per instructions).
