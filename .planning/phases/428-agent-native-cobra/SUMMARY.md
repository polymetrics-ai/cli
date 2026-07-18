# Phase 428 Summary

Status: implementation complete, fully verified, and pushed; no PR created.

## Identity

- Session: `issue-428-pi-openai-codex-gpt-5.6-sol-high-20260718T124925Z`
- Model/thinking: `openai-codex/gpt-5.6-sol`, `high`
- Branch: `refactor/428-agent-native-cobra`
- Exact start: `235233f7cfde4a24612be6b0f95fb37a412d388a`
- Verification end: `20260718T131634Z`
- Parent: #397; umbrella: #407; draft parent PR #438

## Delivered

- Registered native Cobra nodes for `agent`, `agent plan`, `agent image`, `agent image build|pull|ensure`, and hidden positional `agent help`.
- Declared repeated `StringArray --request` with bare `true`, unknown-flag tolerance, no-file completion seam, and exact spaced/assigned/repeated-last-wins behavior.
- Removed `agent` from the legacy wrappers and removed only the agent handler's `parseFlags` call.
- Preserved bare/text/JSON/long/short/positional help; deterministic plan text/JSON; invalid-action categories; assigned globals/config; action-tail `--help`/`-h`; and continuation after literal `--`.
- Added a context-aware injected image runtime and fake-only coverage for build, pull, ensure-present, ensure-pull, exact command arguments/paths/output, failures, and invocation order.
- Added bounded validation for agent request controls, project/build root, Podman binary, and image references before external lookup/execution.
- Kept worker/RLM behavior, canonical help/docs/website/goldens, dependencies, connector definitions, other namespaces, and Phase 15/19 scope unchanged.

## TDD and verification

The test-only RED failed exactly on the intentionally missing `runAgentImageAction` and `newRootCmdWithAgentImageRuntime` seams before production edits. Focused GREEN passed in `4.408s`; expanded/final focused runs passed in `4.480s`/`4.386s`. Agent/router/golden passed in `11.408s`; standalone golden in `5.816s` and `6.054s`; full CLI in `235.686s`; focused race in `1.751s`.

A 25-case base/head differential matched exact exit/stdout/stderr for help, plan flag forms, unknown/positional input, assigned globals, trailing help, literal separators, and missing/invalid image actions. Built help routes were byte-identical (`450` bytes), plan output deterministic, invalid action exit `2`, and invalid assigned boolean/unsafe request exit `3`.

Runtime dependency-free config/RLM/worker tests passed without optional services. Temporary CLI docs generation diff, temporary/tracked connector docs validation, website generation (11 pages), golden, dependency, connector-def, and scope guards were clean. gofmt, vet, build, full repository tests (`real 345.240s`; CLI `238.990s`, certify `341.079s`), and `make verify` (`real 25.853s`; smoke OK, lint `0 issues`, 547 connectors / 0 findings) passed.

## Workflow and safety

GSD doctor/list/plan-phase passed. Programming-loop is absent, so the recorded manual universal-loop fallback enforced strict TDD. Verify-work and code-review prompts generated (7137/6003 bytes) and were executed inline; local review was clean. No external review was requested.

No secret, credentialed connector, optional service, Podman/Docker command, image build/pull/publish, Temporal/PostgreSQL/Dragonfly operation, dependency, worker behavior change, PR, or merge occurred. All image behavior used injected fakes/temp roots. Invalid-action differential used executable lookup only, not execution. The required `make verify` local sample smoke followed reverse ETL plan → preview → approval → run without external writes.

## Delivery

Commits pushed:

- `a37719a7` — planning checkpoint
- `6cdaea60` — RED test checkpoint
- `4e1b20f3` — native implementation checkpoint
- `75e0d732` — validation hardening checkpoint
- final verification/handoff artifact commit

Residual risk: the agent-only pre-Cobra normalizer intentionally preserves legacy ignored help/separator tokens until Phase 19 deliberately changes focused help. Tests and the exact differential pin that boundary.
