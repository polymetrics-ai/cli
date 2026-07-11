# Twenty S6 CLI surface + help/manual/website parity TDD ledger (#283)

Status: REVIEW_FIX_F1_ACCEPTED_METADATA_APPROVAL_CORRECTED. Manual GSD fallback because `scripts/gsd prompt programming-loop init --phase twenty-s6-cli-surface --dry-run` is unavailable (`scripts/gsd: unknown GSD command: programming-loop`).

Loaded skills: `gsd-core`; fallback Go/docs/web skills `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-documentation`, `golang-spf13-cobra`, `frontend-design`, `web-design-guidelines`, `vercel-react-best-practices`; `caveman` for compact handoff. Repo-local `.pi/skills/go-implementation/SKILL.md` and `.pi/skills/ts-website/SKILL.md` missing (`ENOENT`). Skill rule anchors: CLI stdout/stderr + help behavior; testing rules 1/3/5; security threat model questions 1-3 + hardcoded secret rule; safety rules 2/4/6/10; error rules 1/2/7/9; docs concision/no invented context.

## Red evidence captured before production edits

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

## TDD ledger

| # | Red / validation-first gate | Green implementation | Refactor / notes | Status |
|---|---|---|---|---|
| 1 | `cli_surface.json` absent; Python target count would fail. | Added 168-command Twenty CLI surface. | Generated from existing streams/writes/API metadata; engine test also rejects generic `--data/--payload/--raw/--record/--records` flags. | GREEN |
| 2 | `pm help twenty` failed with `help topic "twenty" not found`; focused Go test red. | `writeManual` resolves connector manuals for connector topics. | No credential resolution. | GREEN |
| 3 | `pm twenty` failed with unknown command / missing connector command path depending surface state; focused Go test red. | Bare connector namespace renders manual and exits 0. | Invalid actions still blocked as errors. | GREEN |
| 4 | `pm twenty --help` failed via static manual lookup; focused Go test red. | Top-level connector `--help` renders manual and exits 0. | Reuses connector manual renderer. | GREEN |
| 5 | Docs/manual/website lacked Twenty Command Surface. | Regenerated docs/connectors/manual/skill and website data. | Robust website diff hash unchanged after second generation. | GREEN |
| 6 | Local gates incomplete. | Ran jq/Python/connectorgen/conformance/focused tests/vet/build/help/docs/website/full test gates. | `make verify` skipped due reverse ETL smoke target. | GREEN |
| 7 | Review F1: generated manual/skill claimed `create_*` requires no approval. | Corrected Twenty metadata source, Twenty manual/skill, and connector catalog approval wording to require plan/preview/approval/execute for every create/update/batch/delete action; deletes additionally require `--confirm destructive`. | Safety wording fix only; no live writes. | GREEN |

## Green evidence highlights

```text
twenty cli surface ok 168 Counter({'list': 28, 'get': 28, 'create': 28, 'update': 28, 'batch': 28, 'delete': 28})
go test ./internal/cli -run 'TestTwentyConnector' -count=1 -> ok
go test ./internal/connectors/defs ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/cli ./cmd/connectorgen -count=1 -> ok
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedTwentyCLISurfaceCount -count=1 -> ok
go test ./internal/connectors ./internal/connectors/engine ./internal/cli -run 'Test.*Connector|TestTwentyConnector|TestDocsGenerateAndValidateConnectorDocs|TestBundleLoadEmbeddedTwentyCLISurfaceCount' -count=1 -> ok
go test ./internal/cli -run 'TestTwentyConnector|TestDocsGenerateAndValidateConnectorDocs' -count=1 -> ok
go test -timeout 20m ./... -> ok
```

## Safety ledger

- Reverse ETL execution: NOT RUN (`make verify` skipped because `smoke-no-build` executes `pm reverse run`).
- Live credentials: NOT USED.
- Destructive external actions: NOT EXECUTED; only command metadata/docs/help.
- Generic HTTP/raw write tools: NOT EXPOSED.
- New dependencies: NONE; `go.mod`/`go.sum` unchanged.

- Review fix F2: refreshed connector catalog/help count parity for Twenty (552 total / 548 declarative; catalog rows now show Twenty with 28 streams and 112 writes).

### Review fix F3 numeric scalar CLI flags

Claude local review on head `46f49175` found an important non-blocking gap: Twenty create/update commands surfaced string, boolean, and string-array scalar fields, but silently omitted write-schema `number` scalar fields such as `position` and PDL/count metrics. Plan: add a typed `number` CLI flag kind, coerce it to JSON numbers in commandrunner, expose Twenty numeric scalar write fields as `number` flags (not raw JSON), update generated docs/website artifacts, and rerun focused gates.

#### F3 red tests (numeric scalar flags)
- `go test ./internal/connectors/commandrunner -run TestBuildWriteCommandCoercesNumberFlag -count=1` failed as expected: `flag --estimate has unsupported type "number"`.
- `go test ./internal/connectors/engine -run TestBundleLoadEmbeddedTwentyCLISurfaceCount -count=1` failed as expected: Twenty `attachments create` missing `--position` number flag mapped to `record.position`.

#### F3 green implementation (numeric scalar flags)
- Added typed `number` CLI flag support in `cli_surface.schema.json` and commandrunner coercion (`strconv.ParseFloat`).
- Added Twenty numeric scalar flags for create/update write-schema `number` fields (`position`, PDL/count metrics) without adding generic JSON/raw write flags.
- `go test ./internal/connectors/commandrunner -run TestBuildWriteCommandCoercesNumberFlag -count=1` passed.
- `go test ./internal/connectors/engine -run TestBundleLoadEmbeddedTwentyCLISurfaceCount -count=1` passed.
- Rejects non-finite values (`NaN`/`Inf`) for `number` flags so plans only carry JSON-safe finite numbers.


#### F4 red test (create/update example validity)
- Added regression coverage in `TestBundleLoadEmbeddedTwentyCLISurfaceCount` requiring implemented Twenty create/update examples to include a non-id mutable scalar flag when scalar flags exist.
- Red check failed as expected before the example fix: `attachments create` example `pm twenty attachments create --preview --json` omitted a mutable record flag.

#### F4 green implementation (valid create/update examples)
- Updated implemented Twenty create/update examples with a mutable typed scalar flag where one is exposed (for example `--name fixture` or `--position 1`).
- Removed examples and added notes for workspace-members create/update, whose write schema exposes nested object/array fields only and no scalar CLI flags.
- `go test ./internal/connectors/engine -run TestBundleLoadEmbeddedTwentyCLISurfaceCount -count=1` now passes with regression coverage for example validity.
