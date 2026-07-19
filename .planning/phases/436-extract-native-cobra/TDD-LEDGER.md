# Phase 436 TDD Ledger

Issue: #436 — nativize hidden extract command.
Invocation: `issue-436-pi-sol-high-20260719T074902Z`
Model/thinking profile: `Sol` / `high`
Starting HEAD: `eec03373dcc581c7f5c3331fe63287519b317f53`

## GSD and skills

Doctor/list passed; `plan-phase 436 --skip-research` generated and is executed inline. The adapter lacks `programming-loop`, so the recorded manual universal-runtime-loop fallback is active.

Loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-project-layout`, `golang-documentation`, `golang-spf13-cobra`.

| Step | Kind | Command / evidence | Status |
|---:|---|---|---|
| 0 | Planning | Create PLAN/TDD-LEDGER/VERIFICATION/PROMPTS/RUN-STATE/SUMMARY with identity and exact start before tests or production edits | Complete |
| 1 | RED | `go test ./internal/cli ./internal/rlm -run 'TestExtract|TestDeterministicRunRejectsWarehouseInputPathEscape|TestDeterministicRunRejectsExternalInputFinalLink|TestWriteOutTable' -count=1` | Failed as required before production edits: CLI lacks `newExtractCobraCommand`, `extractCommandRuntime`, and `newRootCmdWithExtractRuntime`; RLM traversal input, external input final link, and external temporary output final link all succeeded |
| 2 | GREEN | Hidden native extract command, typed flags/runtime, bounded table checks, rooted RLM warehouse I/O, and extract-only legacy/parser removal | Pass: focused CLI/RLM `1.752s`/`0.227s`; existing extract request test updated only for intentional bare-help behavior |
| 3 | Refactor | Focused/repeated/race extract/RLM/safety, router/golden/full CLI, exact-start differential, and parity checks | Pass: final extract ×10 `43.880s`, extract race `50.303s`, RLM/safety ×10/race green, router/golden green, final full CLI `429.304s`, exact-start 8/8 preserved plus 5/5 intentional help |
| 4 | Full gate | gofmt, vet, full tests, build, `make verify` default-only | Final pass: `make verify` exit 0; CLI `433.681s`, certify `337.079s`, docs/smoke/lint/connectors green |
| 5 | Containment correction RED/GREEN | Reject a warehouse-directory symlink escaping the selected extract project root before analyzer effects | RED failed in `0.569s`; GREEN passes in `0.560s`, extract ×10 `43.880s`, extract race `50.303s` after project-root validation before analyzer construction |
| 6 | Delivery | Re-run affected/full gates; finalize six artifacts, scope/dependency checks, commit/push; no PR/review | Prior implementation complete at `748f6bdb`; correction reopened |
| 7 | Accepted review correction plan | Accept all four findings; reset terminal verification; plan held project-root effect scope, parser-ownership repairs, unknown-only usage behavior, and full RLM table validation | Complete before correction tests/production edits |
| 8 | Correction RED | `go test ./internal/cli -run 'TestExtract(HeldProjectRootRejectsWarehouseReplacementBeforeAnalyzerEffects|RejectsUnsafeTablesBeforeAnalyzer|LiteralUnknownOperandsGlobalsAndOutputCompatibility)$' -count=1` | Failed as required before production edits: directory replacement re-rooted deterministic input/output externally; literal/assigned/malformed tails rendered `CommandManual`; unknown-only rendered bare help; both dot-table cases reached the fake analyzer |
| 9 | Correction GREEN/refactor | Narrow held warehouse scope, exact-bare state, tail ownership normalization, and full RLM table validation | GREEN: focused extract/router/golden `11.271s`; RLM/safety green; extract ×10 `43.889s`; extract race `50.339s`; RLM/safety repeated/race green; exact-base 11/11 and adversarial 37 matched + 17 intentional help / 0 mismatches |
| 10 | Correction delivery | Finalize artifacts with exact evidence; commit/push; no PR/review | Pending |

## RED contract

- `extract` remains hidden but is a native Cobra command with all current flags: request, sql, limit, provider, model, llm-base-url, in, out, and spec-name.
- Repeated values remain last-wins; assigned, space, and bare forms retain legacy meaning; trailing operands/unknown tails do not become new effects.
- Bare, long/short, topic, positional, trailing, text, and JSON help is contextual and side-effect free while root discovery remains hidden.
- Literal `--`, malformed/legal unknown command heads, invalid actions/operands, and later valid-looking tokens fail closed without query/analyzer/file effects.
- Global root/json/plain/no-input/progress placement and assigned forms retain current invocation semantics.
- Simple-query and RLM routes preserve exact output and error categories through injected dependency-free fakes; no model, Temporal, Podman, worker, service, credential, or network call occurs.
- Extract RLM input/output table names cannot escape the selected warehouse root. Rooted input opens reject final links escaping the root. Rooted output temporary creation and atomic final replacement do not follow an external final link; external sentinel files remain unchanged. Valid local names and in-root paths still work.
- Only extract's legacy registration/parser call is removed. Dynamic connector and other namespace parser behavior remains.

## RED evidence

The complete test-only contract preceded production edits. Focused CLI compilation failed on the intentionally absent native constructor/runtime seam. Independently, all three expected RLM safety failures reproduced: `../outside` input escaped the warehouse, an input final symlink reached an external file, and `scores.ndjson.tmp` followed and changed an external target. The final-output symlink replacement control already passed, proving atomic rename itself replaces rather than follows the final link.

No external user file or service was used: all roots, symlinks, records, and sentinels were created under `t.TempDir`.

## GREEN / refactor evidence

Native Cobra now owns hidden extract and its nine current local flags. An invocation-local runtime isolates query and analyzer effects; production retains the old project-open gate, heuristic/optional classifier, read-only query engine, and typed agent analyzer. Extract validates both table names before analyzer construction. Shared RLM input opens and output temp/create/rename effects now run through a held `os.Root`-backed local filesystem scope; deterministic reads remain streaming while agent staging retains its required byte copy.

Focused, repeated, and race suites pass. The original input traversal, external input final-link, and external output-temp final-link tests are green; atomic replacement preserves the external final-target sentinel and replaces the link. Router/golden/docs focus passes after the one reviewed extract-help fixture change and generated `docs/cli/extract.md`. An exact-start built-binary differential matched 8/8 preserved parser/output cases; 5/5 intentional bare/topic/positional/trailing help routes pass.

## Post-GREEN containment correction

The first implementation roots individual RLM table effects, but extract still needs to prove its fixed warehouse directory is within the selected project root before treating that directory as the held root. Temp-only RED linked `.polymetrics/warehouse` to an external directory and failed in `0.569s`: extract constructed and ran the injected analyzer, returning an ExtractResult instead of a validation error. The external directory remained empty because the fake analyzer has no file effect. GREEN validates the fixed warehouse directory against the selected project root before analyzer construction and RLM request creation. The focused test passes in `0.560s`; all extract tests repeated ×10 pass in `43.880s`; extract race passes in `50.303s`.

## Final verification evidence

- Final full CLI passed in `429.304s`; final full `make verify` passed with CLI `433.681s`, certify `337.079s`, docs validation, ordered temp-root smoke, lint 0, and 547 connector bundles / 0 findings.
- RLM/safety repeated ×10 and focused race passed after the correction. Exact-start binaries matched 8/8 preserved parser/output cases; 5/5 intentional contextual-help routes passed.
- Runtime topic/bare/direct/positional/trailing help, JSON ExtractResult, usage error, and hidden root discovery checks passed.
- Generated CLI manual/golden and website docs-data checks passed with no unexpected delta. Website TypeScript typecheck could not run because `tsc`/existing `node_modules` is absent; no install was attempted because dependencies/external acquisition were prohibited.
- `go mod tidy -diff`, `go mod verify`, dependency/scope/parser-removal checks, gofmt, diff check, vet, full tests, and build passed.

## Accepted correction RED contract

- A held project-root handle must exist before the analyzer factory can mutate `.polymetrics/warehouse`; all analyzer input/output effects remain rooted to that handle and cannot follow the replacement into an external directory.
- Literal `--` and assigned/malformed unknown tokens retain legacy parser ownership; later `--help` or positional `help` cannot be rediscovered as Cobra help, while intentional direct and trailing help remains unchanged.
- Exact bare `extract` alone renders help. `--unknown`, `--unknown=ignored`, `--help=true`, and `--=help` remain usage errors with no effects.
- Both `--in=.` and `--out=.` fail validation before analyzer factory construction using the complete RLM bare-table contract.

## Correction RED evidence

The focused test exited 1 before production edits. The controlled analyzer-factory hook renamed the in-project warehouse and replaced it with an external symlink; the deterministic analyzer then read one external row and atomically replaced the external output sentinel, returning a successful `ExtractResult`. Literal `-- --help`, assigned unknown plus positional `help`, and malformed assigned unknown plus positional `help` all returned `CommandManual` instead of the base `ExtractResult`. `--unknown`, `--unknown=ignored`, `--help=true`, and `--=help` returned bare help instead of a usage error. Both `--in=.` and `--out=.` constructed and ran the fake analyzer. All files and symlinks were owned by `t.TempDir`; no external user file or service was touched.

## Correction GREEN/refactor evidence

`rlm.OpenProjectWarehouse` now opens the selected project root before analyzer construction and passes an RLM-only `WarehouseScope` through `RunRequest`; the scope exposes no generic file API, and deterministic, fixture, and agent input/output helpers reuse it for the whole run. The controlled mutation test now fails closed without changing the external sentinel. Extract records exact bare invocation before pflag discards unknowns, normalizes parser-owned help-like tails to inert legacy tokens, preserves direct/trailing help, and applies exported RLM bare-table validation after identifier validation. Focused extract/router/golden, full RLM/safety, package vet, build, module tidy/verify, repeated/race, 11-case exact-base differential, and 54-case adversarial differential all pass.

## Evidence log

Do not backfill. Append exact commands, failures, durations, commits, and gate results after execution.
