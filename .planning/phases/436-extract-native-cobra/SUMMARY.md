# Phase 436 Summary

Status: native implementation green; broad/final verification pending.

## Identity

- Session: `issue-436-pi-sol-high-20260719T074902Z`
- Model/thinking profile: `Sol`, `high`
- Branch: `refactor/436-extract-native-cobra`
- Exact start: `eec03373dcc581c7f5c3331fe63287519b317f53`
- Parent: #397; umbrella: #407; draft parent PR #438

## Delivered

Native Cobra owns hidden extract, positional help, and all nine current flags while preserving dependency-free routing, globals, ignored operands/unknowns, literal separators, text/JSON output, errors, and hidden discovery. An invocation-local query/analyzer runtime seam makes effect routing testable without optional services. Direct/topic/positional/trailing help now uses a dedicated canonical manual.

Extract validates RLM input/output as bare table names before analyzer effects. Shared RLM warehouse reads and atomic writes use the existing `os.Root`-backed local filesystem scope, preventing traversal, external input-link reads, and external output-temp link writes while safely replacing a final link without changing its external target. Only extract's legacy registration and `parseFlags` call were removed.

## Workflow

GSD doctor/list and plan-phase prompt generation passed. `programming-loop` is absent from the adapter registry, so the manual universal-loop fallback is active. All six issue-local artifacts precede RED tests and production edits. Execution decision is `local_critical_path` because this is the assigned serialized isolated unit, central router writes collide, and no subagent tool is exposed.

## Safety

No external files/services, credentials, dependencies, model, Temporal, Podman, worker, listener, database service, broad paths, generic write tools, destructive/admin actions, production operations, PR, or review. Tests will use only temporary roots, synthetic records, external sentinel files created by the tests, injected fakes, and existing hermetic local execution.

## TDD and verification

RED preceded production edits: native extract symbols were absent, and traversal input, external input final-link, and external output-temp final-link effects all reproduced under temporary roots. GREEN focused/repeated/race extract/RLM/safety and router/golden checks pass. Exact-start differential matched 8/8 preserved cases; 5/5 intentional contextual-help routes pass. Full CLI/repository/vet/build/`make verify` remain pending.
