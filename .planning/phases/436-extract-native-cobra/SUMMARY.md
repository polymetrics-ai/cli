# Phase 436 Summary

Status: complete, verified, and ready for terminal artifact commit/push.

## Identity

- Session: `issue-436-pi-sol-high-20260719T074902Z`
- Model/thinking profile: `Sol`, `high`
- Branch: `refactor/436-extract-native-cobra`
- Exact start: `eec03373dcc581c7f5c3331fe63287519b317f53`
- Implementation head: `15417f956f428c73159b3bb8824eb55cf3d44d36`
- Parent: #397; umbrella: #407; draft parent PR #438

## Delivered

Native Cobra owns hidden extract, positional help, and all nine current flags while preserving dependency-free routing, globals, ignored operands/unknowns, literal separators, text/JSON output, errors, and hidden discovery. An invocation-local query/analyzer runtime seam makes effect routing testable without optional services. Direct/topic/positional/trailing help now uses a dedicated canonical manual.

Extract validates RLM input/output as bare table names and verifies `<root>/.polymetrics/warehouse` resolves beneath the selected project root before analyzer construction. Shared RLM warehouse reads and atomic writes use the existing `os.Root`-backed local filesystem scope, preventing traversal, external input-link reads, and external output-temp link writes while safely replacing a final link without changing its external target. Only extract's legacy registration and `parseFlags` call were removed.

The canonical extract manual, `docs/cli/extract.md`, reviewed golden help fixture, website agent guide/CLI reference/architecture text, and generated website docs data are aligned. Extract remains hidden from root discovery.

## Workflow

GSD doctor/list and plan-phase prompt generation passed. `programming-loop` is absent from the adapter registry, so the manual universal-loop fallback was used. All six issue-local artifacts preceded the first RED. A post-GREEN self-inspection found the warehouse-directory-root gap; its plan and failing test also preceded the correction production edit. Execution decisions remained `local_critical_path` because this was the assigned isolated serialized unit, router writes collide, and no subagent tool was exposed.

## TDD and verification

Initial RED failed on absent native extract symbols and reproduced input traversal, external input final-link, and external output-temp final-link effects. Native GREEN passed focused/repeated/race extract/RLM/safety tests. The containment-correction RED failed in `0.569s` because the fake analyzer ran through an external warehouse-directory link; GREEN passed in `0.560s`, repeated extract ×10 in `43.880s`, and extract race in `50.303s`.

Final full CLI passed in `429.304s`. Exact-start built binaries matched 8/8 preserved parser/output cases, and 5/5 intentional contextual-help routes passed. Runtime help/output/error/hidden checks, golden/manual generation, website docs-data generation, RLM/safety repeated/race, module/scope/parser-removal checks, gofmt, vet, build, and full repository tests passed. Final `make verify` exited 0 with CLI `433.681s`, certify `337.079s`, docs validation, ordered temp-root smoke, lint 0, and connector validation 547/0.

Website TypeScript typecheck could not run because `tsc` and an existing `node_modules` tree are absent. No install was attempted because the task prohibits dependency acquisition; dependency-free website docs generation passed with no unexpected delta.

## Safety

No external user file/service, model, Temporal, Podman, worker, listener, database service, credential, credentialed connector check, dependency, generic shell/HTTP/SQL write tool, destructive/admin action, or production operation was used. Tests used only owned temporary roots, synthetic records, symlinks, sentinels, and injected/hermetic fakes. `make verify` ran its established local temporary-root smoke in the required reverse plan → preview → approval → run order.

## Worker handoff

- Sub-issue: #436
- Parent issue: #397; umbrella #407
- Branch: `refactor/436-extract-native-cobra`
- Base: `feat/cli-architecture-v2`
- Sub-PR/review: not created or run per user instruction
- Implementation head: `15417f956f428c73159b3bb8824eb55cf3d44d36`
- Review/integration coverage: intentionally pending parent-orchestrator handling; do not infer approval
