# Twenty S6 CLI surface + help/manual/website parity verification (#283)

Status: GREEN_LOCAL_GATES_PASSED_COMMITTED_GSD_VERIFIED. Manual GSD fallback active because `scripts/gsd prompt programming-loop init --phase twenty-s6-cli-surface --dry-run` returned `scripts/gsd: unknown GSD command: programming-loop`.

## Pre-production red evidence

```text
pm_exists=no
cli_surface_exists=no

go run ./cmd/pm help docs -> status=0
go run ./cmd/pm help connectors -> status=0
go run ./cmd/pm help twenty -> status=1; stderr: error: help topic "twenty" not found
go run ./cmd/pm twenty -> status=1; stderr: error: unknown command "twenty"; exit status 2
go run ./cmd/pm twenty --help -> status=1; stderr: error: help topic "twenty" not found
go run ./cmd/pm connectors -> status=0; rendered connectors help
```

## Red test evidence after CLI test addition

```text
go test ./internal/cli -run 'TestTwentyConnector' -count=1
--- FAIL: TestTwentyConnectorHelpCommandsRenderManualWithoutCredentials
    help_twenty: Run([help twenty]) code = 1 stderr = error: help topic "twenty" not found
    twenty: Run([twenty]) code = 2 stderr = error: unknown command "twenty"
    twenty_--help: Run([twenty --help]) code = 1 stderr = error: help topic "twenty" not found
FAIL	polymetrics.ai/internal/cli
```

## Green local gates

```text
jq . internal/connectors/defs/twenty/cli_surface.json
=> jq ok

python3 count/intent gate
=> twenty cli surface ok 168 Counter({'list': 28, 'get': 28, 'create': 28, 'update': 28, 'batch': 28, 'delete': 28})

go run ./cmd/connectorgen validate internal/connectors/defs --json
=> {
=>   "findings": [],
=>   "warnings": [],
=>   "connectors_checked": 548
=> }

go test ./internal/connectors/conformance -run 'TestConformance/twenty' -count=1
=> ok  	polymetrics.ai/internal/connectors/conformance	1.230s

go test ./internal/connectors/defs ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/cli ./cmd/connectorgen -count=1
=> ok  	polymetrics.ai/internal/connectors/defs	1.562s
=> ok  	polymetrics.ai/internal/connectors/engine	1.323s
=> ok  	polymetrics.ai/internal/connectors/commandrunner	0.332s
=> ok  	polymetrics.ai/internal/cli	161.980s
=> ok  	polymetrics.ai/cmd/connectorgen	5.582s

go test ./internal/connectors/engine -run TestBundleLoadEmbeddedTwentyCLISurfaceCount -count=1
=> ok  	polymetrics.ai/internal/connectors/engine	0.430s

go test ./internal/connectors ./internal/connectors/engine ./internal/cli -run 'Test.*Connector|TestTwentyConnector|TestDocsGenerateAndValidateConnectorDocs|TestBundleLoadEmbeddedTwentyCLISurfaceCount' -count=1
=> ok  	polymetrics.ai/internal/connectors	0.342s
=> ok  	polymetrics.ai/internal/connectors/engine	0.382s
=> ok  	polymetrics.ai/internal/cli	83.897s

go test ./internal/cli -run 'TestTwentyConnector|TestDocsGenerateAndValidateConnectorDocs' -count=1
=> ok  	polymetrics.ai/internal/cli	4.173s

git diff --check
=> passed (no output)

go vet ./...
=> passed (no output)

go build ./cmd/pm
=> passed (no output; built ./pm)

gofmt -l cmd internal
=> passed (no output)

./pm docs validate --connectors-dir docs/connectors
=> Validated connector docs in docs/connectors

go test -timeout 20m ./...
=> ok; full package suite passed after final Go changes, including internal/connectors/certify in 352.219s and internal/cli in 154.815s
```

## Runtime help/docs parity checks

```text
./pm connectors >/tmp/pm-connectors-help.txt
./pm help twenty >/tmp/pm-help-twenty.txt
./pm twenty >/tmp/pm-twenty-help.txt
./pm twenty --help >/tmp/pm-twenty-flag-help.txt
=> all exit 0
=> wc -l /tmp/pm-connectors-help.txt /tmp/pm-help-twenty.txt /tmp/pm-twenty-help.txt /tmp/pm-twenty-flag-help.txt
=>      138 /tmp/pm-connectors-help.txt
=>      745 /tmp/pm-help-twenty.txt
=>      745 /tmp/pm-twenty-help.txt
=>      745 /tmp/pm-twenty-flag-help.txt
=>     2373 total

rg -n "Twenty CRM|companies list|delete_companies|Command Surface" docs/cli docs/connectors/twenty internal/connectors/defs/twenty website/data website/lib | wc -l
=> 47
```

## Website generation/idempotence

```text
(cd website && pnpm run gen:website-data)
=> Wrote 11 docs pages to lib/docs.generated.ts.
=> Wrote 548 connectors to data/connectors.generated.json; 334 icons copied.
=> Connectors with write actions: 224
=> Wrote 548 connectors to lib/connectors.catalog.generated.ts and lib/connectors.catalog.data.generated.json (7110 KB).
=> Categories: {"api":544,"queue":1,"database":2,"accounting":1}
=> Capabilities: {"check":548,"read":548,"write":224,"query":0,"cdc":0,"dynamicSchema":4}
=> Featured: 42
=> Wrote 548 connectors to lib/connectors.generated.ts (71 famous first, 477 alphabetical).

(cd website && pnpm run gen:website-data && git diff --exit-code -- website/data website/lib website/public)
=> passed in shell with the same generator output above.

Robust root-relative idempotence check:
=> website diff before=f9e24526e3b32d2560bf206e3420b8dab106622b5dfc00d229ac2df16685c861 after=f9e24526e3b32d2560bf206e3420b8dab106622b5dfc00d229ac2df16685c861
```

## GSD workflow check

```text
scripts/verify-gsd-workflow 62b8b46c
=> verify-gsd-workflow: implementation changes have GSD/TDD evidence against 62b8b46c
=> Implementation files changed:
=> internal/cli/cli.go
=> internal/cli/cli_test.go
=> internal/cli/docs.go
=> internal/connectors/defs/twenty/cli_surface.json
=> internal/connectors/defs/twenty/docs.md
=> internal/connectors/engine/bundle_test.go
=> internal/connectors/guide.go
=> Evidence files changed:
=> .planning/phases/twenty-s6-cli-surface/PLAN.md
=> .planning/phases/twenty-s6-cli-surface/RUN-STATE.json
=> .planning/phases/twenty-s6-cli-surface/SUMMARY.md
=> .planning/phases/twenty-s6-cli-surface/TDD-LEDGER.md
=> .planning/phases/twenty-s6-cli-surface/VERIFICATION.md
```

## make verify decision

`make verify` was NOT RUN because the Makefile `verify` target includes `smoke`, and `smoke-no-build` executes `./pm reverse run ...`; issue instructions forbade running `make verify` if it executes reverse run. Equivalent gates run separately: fmt/vet/test/build/docs validate/connectorgen validate plus focused command/help checks.

## Safety result

- Live credentials: not used.
- Credentialed connector checks: not run.
- Reverse ETL execution: not run.
- Destructive external actions: not executed.
- New dependencies: none; `go.mod`/`go.sum` unchanged.

## Review fix F1 verification

F1 accepted: generated Twenty MANUAL/SKILL claimed `create_<object>` actions require no approval. This violated the reverse ETL safety contract.

Fix: updated `internal/connectors/defs/twenty/metadata.json` risk.approval, regenerated/restored Twenty manual/skill wording, and updated the generated connector catalog risk metadata so every create/update/batch/delete action requires plan/preview/approval/execute; delete actions additionally require typed `--confirm destructive`.

Commands:

```bash
jq . internal/connectors/defs/twenty/metadata.json >/dev/null
go run ./cmd/connectorgen validate internal/connectors/defs --json | jq '{findings,warnings,connectors_checked}'
go test ./internal/cli -run 'TestTwentyConnector|TestDocsGenerateAndValidateConnectorDocs' -count=1
./pm docs validate --connectors-dir docs/connectors
(cd website && pnpm run gen:website-data >/tmp/s6-gen2.log && git diff --exit-code -- website/data website/lib website/public)
scripts/verify-gsd-workflow 62b8b46c
```

Results: all passed; website generated data remained idempotent; no live credentials or reverse ETL execution.

### Review fix F2 catalog/help parity

Reviewer found generated connector catalog/help artifacts still stale after F1. Fixed by updating runtime/static help counts (`552` total / `548` declarative), adding Twenty to `docs/connectors/README.md` and `docs/connectors/catalog/all-connectors.md`, and replacing the Twenty entry in `docs/connectors/catalog/all-connectors.json` with the generated 28-stream / 112-write metadata. Re-ran stale-text grep, connectorgen validate, focused CLI docs tests, `pm docs validate`, website idempotence, and `scripts/verify-gsd-workflow 62b8b46c`; all passed.

### Review fix F3 numeric scalar CLI flags

Claude local review on head `46f49175` found an important non-blocking gap: Twenty create/update commands surfaced string, boolean, and string-array scalar fields, but silently omitted write-schema `number` scalar fields such as `position` and PDL/count metrics. Plan: add a typed `number` CLI flag kind, coerce it to JSON numbers in commandrunner, expose Twenty numeric scalar write fields as `number` flags (not raw JSON), update generated docs/website artifacts, and rerun focused gates.

### Review fix F3 numeric scalar flags

Fixed the `claude_local` important finding by adding a typed `number` CLI flag kind/coercion and exposing Twenty create/update write-schema `number` scalar fields as typed flags. Updated Twenty generated manual/skill and website generated data. Verification passed:

- `go run ./cmd/connectorgen validate internal/connectors/defs --json | jq '{findings,warnings,connectors_checked}'` → `findings: []`, `warnings: []`, `connectors_checked: 548`.
- `go test ./internal/connectors/commandrunner ./internal/connectors/engine ./internal/cli -run 'TestBuildWriteCommandCoercesNumberFlag|TestBundleLoadEmbeddedTwentyCLISurfaceCount|TestTwentyConnector|TestDocsGenerateAndValidateConnectorDocs|TestConnectorCatalog' -count=1` → all `ok`.
- `go build ./cmd/pm` → passed.
- `./pm docs validate --connectors-dir docs/connectors` → passed.
- `cd website && pnpm run gen:website-data` followed by `git diff --exit-code -- website/data website/lib website/public` → idempotent.

- `go test ./...` → passed after F3 numeric flag changes.
- `scripts/verify-gsd-workflow 62b8b46c` → passed after F3.
- F3 finite-number guard added for `number` flags and covered by `TestBuildWriteCommandCoercesNumberFlag`.

### Review fix F4 create/update example validity

Fixed the `pm-reviewer` blocking example finding by updating Twenty create/update examples to include a mutable typed scalar flag where possible, and by removing examples plus adding notes for workspace-members create/update where only nested object/array record fields are exposed. Regenerated Twenty manual/skill and website generated data.

F4 verification passed:
- `go run ./cmd/connectorgen validate internal/connectors/defs --json | jq '{findings,warnings,connectors_checked}'` → `findings: []`, `warnings: []`, `connectors_checked: 548`.
- `go test ./internal/connectors/commandrunner ./internal/connectors/engine ./internal/cli -run 'TestBuildWriteCommandCoercesNumberFlag|TestBundleLoadEmbeddedTwentyCLISurfaceCount|TestTwentyConnector|TestDocsGenerateAndValidateConnectorDocs|TestConnectorCatalog' -count=1` → all `ok`.
- `go build ./cmd/pm` → passed.
- `./pm docs validate --connectors-dir docs/connectors` → passed.
- Website generation idempotence (`pnpm run gen:website-data` + `git diff --exit-code -- website/data website/lib website/public`) → passed.
- `scripts/verify-gsd-workflow 62b8b46c` → passed.

### Review fix F5 generated catalog parity

Fixed verifier's remaining generated catalog diff by updating `docs/connectors/catalog/all-connectors.md` and `.json` to match `go run ./cmd/pm docs generate` catalog output for the GitHub generated counts/metadata. Verification passed:

- `go run ./cmd/connectorgen validate internal/connectors/defs --json | jq '{findings,warnings,connectors_checked}'` → `findings: []`, `warnings: []`, `connectors_checked: 548`.
- `go test ./internal/connectors/engine ./internal/cli -run 'TestBundleLoadEmbeddedTwentyCLISurfaceCount|TestDocsGenerateAndValidateConnectorDocs|TestConnectorCatalog' -count=1` → all `ok`.
- `./pm docs validate --connectors-dir docs/connectors` → passed.
- Temp docs generation catalog compare (`diff -q docs/connectors/catalog/all-connectors.{md,json} $tmp/connectors/catalog/...`) → matched.
- `scripts/verify-gsd-workflow 62b8b46c` → passed.
