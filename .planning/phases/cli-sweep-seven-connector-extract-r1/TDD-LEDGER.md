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

- The seven connector directories were imported byte-for-byte from `c28bc75a3`: 457 source files,
  zero missing or extra files, and at import time zero changed files. No `github` or
  `zendesk-support` source/test file entered the branch.

  **The byte-exact claim describes the IMPORT, not the shipped tree.** Inline review then found
  defects in the imported definitions themselves, and 17 files were deliberately corrected after
  import. Those corrections are listed under "Red — inline review findings against the imported
  tree" below; a future audit that diffs the shipped bundles against `c28bc75a3` will find them and
  must read that section rather than treating them as import drift.
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

## Red — inline review findings against the imported tree

Inline review found real defects that the phase's own verification had missed. They are recorded
here as red states because each was a genuine failing condition, not a style note.

### Red — plural write coverage was wired into one consumer of four

The `covered_by.writes` foundation added `engine.SurfaceCoverage.WriteTargets()` and migrated
`cmd/connectorgen/validate.go`, but four other readers of the same rule still read the singular
`.Write` only. jira ships 292 and workday-rest 252 endpoints whose `covered_by` carries **only**
`writes`, so every one of those rows was invisible to them.

Re-running the affected checks against the unmigrated code reproduces the failures:

```text
--- FAIL: TestSurfaceInventoryCountsPluralOnlyWriteCoverage/jira
    Result = "fail" reason="api_surface endpoint 1 is neither covered nor blocked with typed reason"
--- FAIL: TestSurfaceInventoryCountsPluralOnlyWriteCoverage/workday-rest
    Result = "fail" reason="api_surface endpoint 652 is neither covered nor blocked with typed reason"
--- FAIL: TestBatchMaterializePluralOnlyWriteCoverage
    batch materialize exit = 1, want 0; 0 connector(s) materialized, 1 dropped
```

`internal/connectors/conformance` and `internal/connectors/certify` were **not in this phase's
verified package set**, which is exactly why the defect shipped: `TestConformance` sweeps the real
`../defs` tree, so it would have failed on `TestConformance/jira` and `TestConformance/workday-rest`
the moment it was run. Both packages are now part of the recorded gate set.

### Red — twelve help-scout streams carried unbound path placeholders

The imported help-scout `streams.json` declared sub-resource stream paths such as
`/conversations/{conversationId}/threads`. Only `{{ ... }}` is a template; a single-brace `{name}`
is literal text, so those twelve streams would have requested the braces verbatim while the
generated ETL commands already declared matching `--conversation-id`-style flags bound to
`config.*` keys that neither the path nor `spec.json` referenced.

Nothing caught it: `engine.ResolveCheck` scans `{{ }}` matches only, so `connectorgen validate`
reported 551 connectors / 0 findings, and neither `TestEveryImplementedCommandPassesRuntimePreflight`
nor the real-binary NAME sweep issues a request.

### Red — three write-capable connectors published read-only disclosures

help-scout, jira and workday-rest each flipped `capabilities.write` to true while their
`metadata.json` risk block and `docs.md` still stated the connector was read-only with no write
actions — text already published to operators through `pm connectors inspect`, the generated
MANUAL/SKILL files and the website bundle. help-scout's said "read-only, no obviously-safe
reverse-ETL writes" for a connector with 65 write actions including 18 permanent DELETEs.

## Green — post-review corrections and re-run results

### Deliberate post-import corrections (17 files)

These diverge from `c28bc75a3` on purpose; they are fixes to the imported runtime definitions, not
import drift:

- `help-scout/streams.json` — 12 stream paths repointed to `{{ config.<key> }}` interpolation.
- `help-scout/spec.json` — the five scoping config keys those paths and the generated flags
  reference are now declared.
- `help-scout/metadata.json`, `jira/metadata.json`, `workday-rest/metadata.json` — accurate
  `risk.write` and `risk.approval`, and truthful descriptions.
- `help-scout/docs.md`, `jira/docs.md`, `workday-rest/docs.md`, `lever-hiring/docs.md` — corrected
  operation ledgers, write-action sections and known limits.
- `jira/*.json`, `workday-rest/*.json` (10 files) — trailing newline restored to match the
  convention every other bundle follows.

### Green — the four migrated consumers

```text
ok  	polymetrics.ai/internal/connectors/engine
ok  	polymetrics.ai/internal/connectors/conformance
ok  	polymetrics.ai/internal/connectors/certify
ok  	polymetrics.ai/cmd/connectorgen
```

Plural coverage now has direct regression tests rather than incidental coverage:

- `certify.TestSurfaceInventoryCountsPluralOnlyWriteCoverage` pins jira and workday-rest and
  asserts both classification and the write COUNT (292 and 252). github cannot serve this purpose:
  all 231 of its write rows use the singular spelling, so its pinned assertions pass with or
  without plural support.
- `certify.TestSurfaceInventoryPluralOnlyBundlesUseNoSingularWrite` guards that premise — it fails
  if a regeneration ever rewrites those bundles to the singular form and silently drops the
  coverage.
- `connectorgen.TestBatchMaterializePluralOnlyWriteCoverage` drives the real `batch materialize`
  command end to end, exercising `batchSurfaceSplit`, `ensureMaterializedCoverage` and
  `materializeCLISurface` in one pass. Its fixture includes a **two-element** `writes` array over a
  single endpoint — the cardinality the shared foundation exists for, which no shipped bundle
  currently uses.

### Green — the placeholder invariant is now enforced

`engine.ResolveCheckRequestPath` rejects an unbound single-brace placeholder in a declarative read
path — `StreamSpec.Path`, `HTTPBase.URL`, `HTTPBase.Check.Path` and
`FanOutSpec.IDsFrom.Request.Path`. Both `connectorgen validate` and `conformance` call it, so the
rule is enforced once at a shared boundary.

`writes.json` is deliberately exempt: `WriteAction.PathFields` binds `{owner}`/`{repo}`-style
placeholders from the record, and 165 shipped write paths rely on it.

The rule is zero-churn — a sweep of all 551 bundles found 0 stream paths, 0 base URLs, 0 check
paths and 0 fan-out request paths carrying a literal placeholder. Seeded-invalid fixtures
(`stream-path-literal-placeholder`) now cover it in both packages, plus engine unit tests asserting
that `{{ config.x }}`, `{{ fanout.id }}` and plain paths still pass.

### Green — re-run fleet gates on the post-fix tree

```text
connectorgen validate: 551 connector(s) checked, 0 findings
connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)
--- PASS: TestEveryImplementedCommandPassesRuntimePreflight (4.90s)
```

Docs, manuals, catalog and website data were regenerated after the metadata/docs corrections, and
`pm docs validate --connectors-dir docs/connectors` passes.

The 1,984-command real-binary NAME sweep predates these corrections and was not re-run here. None
of the corrections adds, removes or renames a command: the bundle changes are stream path
templates, spec keys, risk/description text and docs prose. The seven implemented/documented counts
are unchanged at 911/911, 584/617, 139/144, 127/138, 100/148, 63/79 and 60/106.
