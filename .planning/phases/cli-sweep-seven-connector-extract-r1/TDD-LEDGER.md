# TDD ledger — seven connector extraction r1

## Plan checkpoint

- **GSD:** `discuss-phase --auto` and `plan-phase --tdd` prompts generated through the project
  adapter; execution is inline/manual for the documented single-worker reason in `PLAN.md`.
- **Green:** pending — import exactly the seven authored bundle deltas and regenerate derived
  surfaces/docs/data.
- **Refactor:** pending — review generated diff for allowlist compliance and generator drift.

## Required red command

```bash
go test -timeout 20m ./cmd/connectorgen -run 'Test(Chatwoot|Gmail|Greenhouse|HelpScout|Jira|LeverHiring|WorkdayRest)'
```

The red result must be recorded verbatim enough to distinguish an assertion failure from a missing
test, compilation failure, or unrelated baseline failure. No production bundle input may change
before that capture.

## Red — observed before any bundle input changed

The seven acceptance test files were imported from `c28bc75a3`; the bundle inputs remained the
current-main versions. The required command exited 1 with assertions (not a compile or unrelated
baseline failure):

```text
--- FAIL: TestChatwootAPISurfaceOperationLedger
    operation_ledger_version = 0, want 1
--- FAIL: TestGmailAPISurfaceOperationLedger
    operation_ledger_version = 0, want 1 (the v2 provenance ledger is required)
    34 legacy excluded row(s) remain, want 0
    covered(45)+blocked(0) = 45, want 79
--- FAIL: TestGreenhouseAPISurfaceOperationLedger
    operation_ledger_version = 0, want 1
    endpoints = 129, want 138 documented operations
    covered(126)+blocked(0) = 126, want 138
--- FAIL: TestHelpScoutAPISurfaceOperationLedger
    operation_ledger_version = 0, want 1
    endpoints = 8, want 144 documented operations
    covered(4)+blocked(0) = 4, want 144
--- FAIL: TestJiraDocumentedSurfaceIsComplete
    operation_ledger_version = 0, want 1
    documented endpoints = 15, want 617
    covered rows = 3, want 590
    blocked rows = 0, want 27
--- FAIL: TestLeverHiringAPISurfaceOperationLedger
    operation_ledger_version = 0, want 1
    endpoints = 11, want 106 documented operations
    covered(5)+blocked(0) = 5, want 106
--- FAIL: TestWorkdayRESTDocumentedSurfaceIsComplete
    operation_ledger_version = 0, want 1
    documented endpoints = 0, want 907
    covered(3)+blocked(0) = 3, want 4
FAIL	polymetrics.ai/cmd/connectorgen
```

**Red conclusion:** current-main bundle inputs cannot satisfy the source-complete operation ledgers;
the target-specific tests precisely fail on the intended missing surface/disposition work.

## Foundation gate — captain-authorized shared support

After importing the seven bundle inputs, `go run ./cmd/connectorgen surface-sync
internal/connectors/defs` failed before generation:

```text
connectorgen surface-sync: runtime operation endpoint ledger: load jira: load bundle jira:
api_surface.json: /endpoints/1/covered_by/writes: additional property not allowed
```

A confirming validation run reported the same issue for Jira and Workday. The source commit adds
plural `covered_by.writes` support to the engine bundle type, API-surface schema, and validator.

The captain then explicitly authorized this bounded foundation in the extraction because the two
connectors require multiple named write contracts on a single documented endpoint. The red slice is
the focused plural-write validator test, followed by implementation only in the four paths named in
`PLAN.md`; its own labelled commit will precede the connector bundle commit. The all-bundle validator
must report all 551 checked with zero findings before surface generation resumes.

### Red — focused plural-write validator test

Before the engine type was changed, the focused test command exited 1 because the proposed plural
field did not exist; this is the expected compile-time red state, not a baseline failure:

```text
# polymetrics.ai/cmd/connectorgen [polymetrics.ai/cmd/connectorgen.test]
cmd/connectorgen/validate_surface_test.go:53:40: unknown field Writes in struct literal of type engine.SurfaceCoverage
cmd/connectorgen/validate_surface_test.go:77:40: unknown field Writes in struct literal of type engine.SurfaceCoverage
FAIL	polymetrics.ai/cmd/connectorgen [build failed]
```

### Green — focused plural-write validator test

After adding only plural write-target handling to the engine type, API-surface schema, and validator,
the focused test command passed:

```text
ok  	polymetrics.ai/cmd/connectorgen	0.745s
```

### Green — all-bundle compatibility check

With the seven imported bundle inputs present, the real validator loaded every bundle and found no
regression in the existing connector corpus:

```text
connectorgen validate: 551 connector(s) checked, 0 findings
```

## Green — seven connector extraction and generated artifacts

- The seven connector directories were compared byte-for-byte with `c28bc75a3`: 457 source files,
  zero missing, extra, or changed files. No `github` or `zendesk-support` source/test file entered
  the branch.
- `go run ./cmd/connectorgen surface-sync internal/connectors/defs` regenerated the runtime endpoint
  ledger (`updated 2060 endpoint(s)`) with no hand-authored command-field correction. Its subsequent
  `--check` scanned 551 connectors with zero fields filled or corrected.
- The endpoint-ledger comparison against `origin/main` changed only `chatwoot`, `jira`,
  `lever-hiring`, and `workday-rest`, which is confined to the seven-connector allowlist.
- The focused seven acceptance tests passed:

  ```text
  ok  	polymetrics.ai/cmd/connectorgen	1.201s
  ```

- The generated connector manuals, catalog, website connector data, and CLI golden transcripts were
  regenerated by their documented generators. The full `internal/cli` package passed after the
  transcript generator updated the root-help listings:

  ```text
  ok  	polymetrics.ai/internal/cli	635.099s
  ```

## Green — executable surface proof

The compiled `/tmp/pm-cli-sweep-seven` binary was invoked once for every `availability: implemented`
command, and every command's own `NAME` header matched its connector and path:

```text
real-binary help reachability: 1984/1984 implemented commands reached their own NAME line
```

The seven documented help-topic and bare-namespace entry points also passed. `pm help <connector>`
correctly resolves the connector manual (`pm connectors inspect <connector>`); bare `pm <connector>`
correctly renders that connector's group help.
