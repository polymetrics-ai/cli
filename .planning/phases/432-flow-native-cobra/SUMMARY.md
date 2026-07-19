# Phase 432 Summary

Status: planning initialized before tests or production edits.

## Identity

- Session: `issue-432-pi-openai-codex-gpt-5.6-sol-high-20260719T034344Z`
- Model/thinking: `openai-codex/gpt-5.6-sol`, `high`
- Branch: `refactor/432-flow-native-cobra`
- Exact start: `ec12c1729e0aaf233a853eff5c6291885f910b15`
- Parent: #397; umbrella: #407; draft parent PR #438

## Plan

Nativize only the current flow namespace and flags while preserving flow directory defaults, manifest/DAG behavior, named runs, cancellation, deterministic events/telemetry/checkpoints/ledger/output, exact error taxonomy, global booleans, and legacy help/literal/unknown/operand semantics. Remove only the flow parser; retain dynamic connector parsing. Phase 10 dashboards, Phase 11 create wizard, and Phase 19 focused help/man work are excluded.

## Workflow

GSD doctor/list passed and plan-phase generated. The adapter lacks `programming-loop`, so the manual universal-loop fallback is active. All six issue-local artifacts were created with exact identity/start before tests or production edits. Execution decision is `local_critical_path` for this assigned serialized isolated unit; no subagent tool is exposed.

## Safety

Temporary manifests/roots and fakes only. No action flow execution, reverse ETL, external write, credentialed check, optional service, dependency, unrelated change, PR, or review.

## TDD and verification

Pending RED → GREEN → refactor and final gates.
