# Verification — #4352 source-bound read execution foundation

## F1 source/local request-input closure checklist — 2026-08-27

- [x] Operation-only `query.rogue` red recorded; source-import check and
  validate reject it after repair.
- [x] Operation-plus-CLI `--rogue -> query.rogue` red recorded; source-import
  check and validate reject it after repair.
- [x] Header/body class closure, source paging placement, and valid Asana
  optional-filter preservation have focused regressions.
- [x] Affected Go packages and full `cmd/connectorgen` pass: generator
  `212.521s`, engine `13.647s`, commandrunner `21.764s`, App `291.638s`, CLI
  `532.727s`, and Asana definitions `5.150s`.
- [ ] Source import, validate, surface sync, operation evidence, declaration
  admission, runtime preflight, canon, docs/website validation, and an actual
  fresh 212-command credential-boundary census pass.
- [ ] Exact pushed SHA/base are recorded; a separate fresh audit is pending.

### Current-head audit ledger (AUDIT-001–006)

`003.msg` fetched `origin/main` at
`1324c52bab0b224ed8958858af7676b8b8e191b4`; it is already the merge base of
the repair head, so no no-op merge was created. `004.msg` requires the frozen
audit be accounted for rather than treating F1 as a standalone readiness claim:

- [x] **001:** source projection excludes raw paging from caller parameters;
  fresh source-import and surface-sync are clean.
- [x] **002:** App/CLI origin-before-credential coverage remains green in the
  full App and CLI suites.
- [x] **003:** App/direct stream binding regression remains green in the full
  engine suite.
- [x] **004:** the existing action/delete lane is covered by the fresh Asana
  definition suite and 212-command credential-boundary census.
- [x] **005:** operation evidence reports fixed-100 passed.
- [x] **006:** docs validation, website generation, and all 34 website script
  tests pass; no generated file drift remains.

The F1 repair adds the independent descriptor-to-local input closure; it does
not narrow or waive any of the original six findings. The next authority is a
fresh, separate audit of the exact pushed SHA.

## Repair r4 checklist

- [x] AUDIT-001 paging red/green: projection, engine/direct read, CLI/App parity, and generated help.
- [x] AUDIT-002 ordering red/green: CLI/App credential/auth/requester spies.
- [x] AUDIT-003 direct `Read`/`ReadWithOutcome` binding red/green.
- [x] AUDIT-004 Asana 21-mutation/delete census red/green.
- [x] AUDIT-005 both fixed-100 isolation red/green; full `go test -timeout 20m ./cmd/connectorgen`.
- [x] AUDIT-006 generated docs/manual/help/website semantic red/green.
- [x] `source-import --check`, validation, surface sync, operation evidence, runtime-preflight/canon, docs/website, and credential-free binary census. Declaration-admission is owned by open PR #4351 and is not present on this branch; the source-bound admission covered here is verified by source import and origin preflight.
- [ ] Final fresh independent exact-SHA Codex audit after code SHA is pushed.

### AUDIT-005 result

- [x] Intended red and complete green recorded in `TDD-LEDGER.md`; focused fixed-100 isolation command passed in `6.906s`.

Status: verified locally, pending PR automation and human gate.

## r4 follow-up final local run

- `go test -timeout 20m -count=1 ./cmd/connectorgen` passed in `182.924s`;
  `./internal/connectors/engine`, `./internal/connectors/commandrunner`,
  `./internal/connectors/defs/asana`, and `./internal/cli` passed (the CLI run
  completed in `420.427s`).
- `go vet` over all affected packages, `make lint`, `make tidy-check`,
  `make docs-check-no-build`, `go run ./cmd/agentcontractgen check`, and
  `make connector-canon-check` passed. `npm --prefix website run test:scripts`
  passed 34 tests. Website `typecheck` could not run because this worktree has
  no `tsc` executable; this is an environment prerequisite, not a product
  test result.
- Source import (249 Asana operations), full validation (553 connectors),
  surface-sync (zero corrections), operation evidence (1,774 rows/5 rollups;
  fixed-100 passed), runtime preflight, connector boundary (clean), and all
  certification checks passed after final regeneration.
- The fresh built `pm` censused all 212 implemented Asana commands in a new
  credential-free project: every command exited 1 with `missing --credential`.
- [ ] The required independent exact-SHA audit remains pending until the
  normal fast-forward repair commit is pushed to #4356.

## Fresh-audit F6 / AUDIT-002 follow-up

- [x] Red evidence recorded from the exact-SHA independent audit: persisted
  source-bound `base_url` configuration was not part of the early preflight.
- [x] Green: `TestSourceBoundOriginRejectsPersistedCredentialConfigBeforeVault`
  proves a persisted public credential configuration rejects the origin before
  vault/authentication/provider work.
- [x] `go test -timeout 20m -count=1 ./internal/app` passed (267.522s).
- [x] `go test -timeout 20m -count=1 ./internal/cli` passed (413.315s).
- [x] `go vet ./internal/app ./internal/cli`, `go build -o ./pm ./cmd/pm`, and
  `make lint` passed (0 lint issues).
- [ ] Rerun the required independent audit at the next pushed SHA.

## Current-main integration #4351

- [x] Red provenance: before integration this branch's merge-base with
  `origin/main` was `b33983927d863032dac8220949990506e812937d`, not the
  authorized #4351 main SHA `1324c52bab0b224ed8958858af7676b8b8e191b4`.
- [x] Green: normal merge `114a4e3f4` has current `origin/main`
  `1324c52bab0b224ed8958858af7676b8b8e191b4` as a parent. Four true conflicts
  were compositionally resolved: #4351 admission/deferred metadata is retained
  alongside #4356 source-operation identity and origin preflight; the
  certification subject was regenerated. A merge-introduced redundant CLI
  config parse was removed.
- [x] The full `cmd/connectorgen` integration red (`246/4`, not `243/4`
  endpoint aliases) was corrected as the expected #4356 source-bound surface;
  the complete package passed in 165.432s. Engine (11.401s), commandrunner
  (21.479s), App (269.692s), CLI (426.131s), Asana defs (5.067s), and defs
  (2.013s) passed.
- [x] Source import verified 249 Asana operations; validation checked 553
  connectors with zero findings; surface sync had zero corrections; operation
  evidence is current at 1,774 rows / five rollups with fixed-100 passed.
  Declaration admission, runtime preflight, canon, connector boundary clean
  (324 files / 553 connectors), agent-contract, docs, website-data generation,
  website scripts (34/34), certification subject/matrix/candidates/sweep,
  affected `go vet`, tidy, and lint all passed.
- [x] Rebuilt `pm`, initialized one fresh credential-free project, and ran all
  212 implemented Asana commands serially with only non-secret fixture values
  for required flags: `212/212` exited 1 at `missing --credential`; zero
  failures and no credential/provider use.
- [ ] Website typecheck was attempted but not runnable (`tsc: command not
  found`). Aggregate `go test ./...` and `make verify` remain CI-owned due the
  per-command runner limit; they are not recorded as local passes.

## Final post-rebase repair run

- `go test -timeout 20m -count=1 ./cmd/connectorgen` passed after the V3
  operation-evidence regeneration.
- `go test -timeout 20m -count=1 ./internal/connectors/engine`,
  `./internal/connectors/commandrunner`, and `./internal/connectors/defs/asana`
  passed. Focused checks cover fan-out route substitution, configured-origin
  rejection before auth/I/O, direct-vs-stream proof, missing foundation, and
  direct-write/reverse-ETL/delete regression.
- `go run ./cmd/connectorgen source-import asana --read-projection-only --check`,
  `validate internal/connectors/defs`, `surface-sync internal/connectors/defs
  --check`, and `operation-evidence --check` passed. The last reports 1,774
  rows and 5 rollups, including all 249 Asana source rows.
- `./pm docs validate --connectors-dir docs/connectors`, the selected
  `TestGoldenTranscripts`/`TestSkillsGenerateMatchesTrackedSkills`, and
  `npm --prefix website run gen:website-data` passed. `git diff --check` passed.
- Scoped `go vet` for `cmd/connectorgen`, engine, commandrunner, and Asana
  passed; a freshly built binary in `/tmp/pm-source-bound-read-KcTH0F` returned
  `missing --credential` (exit 1) for `access-requests get-access-requests`
  with only `--target fixture-target`, before any provider I/O.

## Focused red/green proof

- Red: `go test ./cmd/connectorgen -run '^TestSourceProjectionMaterializesSourceBoundGETReadsWithoutInventingETL$' -count=1` failed before production changes because `get_access_requests` carried no source binding.
- Green generator coverage: `go test ./cmd/connectorgen -run '^(TestSourceProjectionMaterializesSourceBoundGETReadsWithoutInventingETL|TestSourceProjectionLeavesIncompleteReadAsNamedFoundation|TestSourceProjectionSourceCitedMutationDispositionLeavesExistingProjectionByteIdentical|TestSourceProjectionRequiresExplicitReadOnlyNonMutationDeclaration|TestSourceProjectionRequiresReachableRESTReadOrConcreteGap)$' -count=1` passed.
- Green engine coverage: `go test ./internal/connectors/engine -run '^TestPreflightSourceBound' -count=1` passed.
- Green runner coverage: `go test ./internal/connectors/commandrunner -run '^(TestRunSourceBoundOperationDirectReadRejectsBeforeDispatch|TestRunSourceBoundReadMissingFoundationRefusesBeforeDispatch)$' -count=1` passed.
- Green Asana controls: `go test ./internal/connectors/defs/asana -run '^(TestSourceBoundReadControlsReachEnginePreflight|TestReverseETLLedgerReconciles|TestDestructiveOperationsStayBlocked|TestReverseETLWriteActionsExecute)$' -count=1` passed.

## Regression and repository gates

- `go test -timeout 20m ./internal/connectors/engine -count=1` passed.
- `go test -timeout 20m ./internal/connectors/commandrunner -count=1` passed, including `TestEveryImplementedCommandPassesRuntimePreflight`.
- `go test -timeout 20m ./internal/connectors/defs/asana -count=1` passed.
- `go vet ./...`, `go build ./cmd/pm`, `make tidy-check`, `make docs-check-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make lint`, `make connector-runtime-preflight`, `make smoke-no-build`, `make connector-canon-check`, and `npm --prefix website run typecheck` passed.
- `go run ./cmd/connectorgen validate internal/connectors/defs` reported 553 connectors and 0 findings; `go run ./cmd/connectorgen surface-sync internal/connectors/defs --check` reported no drift.
- `git diff --check` passed after the final generated-doc refresh.

`connector-boundary` and `release-installed-github-certification.sh` exceeded the local per-command runner limit. Their child processes were terminated after confirming they were this task's duplicate validation attempts; they are not recorded as passing and remain CI/PR checks.

## Final F1/F2 and mutation-mapping repair evidence

- The source-lock audit's serialized broad-generator gate exited PASS. On this
  exact repair head, `go test -timeout 20m -count=1 ./cmd/connectorgen` passed
  in 150.671s.
- Focused source-projection tests passed with the retained Asana source lock:
  the retained import rejects source identity, method/path, typed-input, and
  workspace pagination drift; partial coverage preserves an implemented
  incomplete action but rejects a complete action, an unrelated foundation,
  and a non-mutating operation.
- `go run ./cmd/connectorgen source-import asana --read-projection-only --check`
  verified 249 operations; full `validate internal/connectors/defs` reported
  553 connectors and 0 findings; full `surface-sync ... --check` scanned 553
  connectors with zero fill/correction drift.
- On this head, `go test -timeout 20m -count=1` passed for
  `internal/connectors/{commandrunner,engine}`, the focused Asana
  source-bound/reverse-ETL/delete controls, the selected generated root-help
  transcripts, and the generated Asana skill check. `pm docs validate`,
  `go vet ./cmd/connectorgen`, `go run ./cmd/agentcontractgen check`, and
  `git diff --check` passed.
- The source-cited Asana mutation disposition inventory is exact and retained:
  21 absent actions are non-executable; 65 implemented reverse-ETL actions
  retain `cli-request-schema-foundation-r1`; 4 implemented delete actions
  retain `source-path-parameter-alias-foundation-r1`. No action/command was
  downgraded or invented.

## Credential boundary

Built a fresh temporary `pm`, initialised a fresh temporary project, and invoked `pm asana access-requests get-access-requests --root <temp> --json` without a credential. It returned `error: missing --credential` (exit 1). No provider credential was configured and no provider request was made.

## CLI/docs/website parity

- Checked `pm --help`, `pm asana --help`, and the generated Asana manual/skill output.
- `docs/connectors/asana/{MANUAL,SKILL}.md`, `docs/skills/pm-asana/SKILL.md`, the connector catalog, and website generated connector data were refreshed.
- Top-level `docs/cli/**` did not need a hand-authored command-tree change; broad unrelated generated pages were intentionally not included.

## R5-pass current-main reconciliation — 2026-08-27

- [x] Reconciled `main` `b9b2478b3b2451d632d28b9aa138a170ad835110` with the existing #4356 repair. The only real overlaps were source-import and source-projection behavior/tests; both partial mutation coverage and write-disabled artifact projection are retained.
- [x] Regenerated source import, surface, docs/manual, skills, and website data. Check mode verified 249 Asana operations; validation reported 553 connectors/zero findings; surface sync reported zero corrections; operation evidence reported 1,774 rows/five rollups and green fixed-100 checks.
- [x] Passed focused conflict tests (19.500s), full changed packages: `cmd/connectorgen` (177.850s), engine (9.928s), commandrunner (22.072s), Asana defs (5.582s), App (267.112s), and CLI (456.149s). Scoped vet, tidy, lint, agent-contract, declaration-admission, runtime-preflight, canon, connector-boundary, docs validation, website scripts (34/34), and certification subject/matrix/candidates/sweep all passed.
- [x] Built `pm` and ran every implemented Asana command in its own initialized credential-free project. `212/212` reached `missing --credential` (exit 1): 106 direct reads, 12 ETL, 94 reverse ETL; zero unknown/blocked/provider results.
- [ ] Website typecheck remains unavailable locally because `tsc` is absent. Aggregate `go test ./...` and `make verify` remain CI-owned due this runner's per-command limit.

## R6 merge-blocker repair

- [x] `go test -timeout 20m -count=1 ./internal/connectors/commandrunner`
  passed (22.353s), including schema/encoding/OpenAPI-sibling/batch structured
  unavailable-disposition and arbitrary-prose refusal controls.
- [x] `go test -timeout 20m -count=1 ./internal/connectors/engine` passed
  (11.504s), including both adapter methods with zero auth/requester calls on
  source-bound ETL drift.
- [x] Rebuilt `pm` and confirmed the four real Asana unavailable examples
  return exact declared schema, encoding+schema, OpenAPI-sibling, and generic
  batch dispositions at exit 7 before any credential/provider action.
- [x] Source import check (249), validate (553/zero), surface-sync (zero),
  runtime preflight, canon, and scoped vet passed. `make lint` was attempted
  but could not start because another process held its parallel lint lock; it
  is not claimed as a pass.
