# Phase 427 Summary

Status: implementation complete, fully verified, and pushed; no PR created.

## Identity

- Session: `issue-427-pi-openai-codex-gpt-5.6-sol-high-20260718T112639Z`
- Model/thinking: `openai-codex/gpt-5.6-sol`, `high`
- Branch: `refactor/427-docs-native-cobra`
- Exact start: `ab847da28ebf78e5732ac1edcde8e39f92dc2656`
- Parent: #397; umbrella: #407; draft parent PR #438

## Delivered

- Registered `docs`, `docs generate`, and `docs validate` as native Cobra nodes with typed repeated `StringArray` output-directory flags, legacy bare-flag semantics, unknown-flag tolerance, positional help, and no-file completion seams.
- Preserved bare/text/JSON/short/positional help; spaced, assigned, repeated, bare, default, comma, and space-containing output paths; unknown flags/extra positionals; assigned global/config booleans; error categories; exact output text; generated CLI bytes; and connector docs/catalog/icons validation.
- Removed `docs` from legacy wrappers and removed only its now-unused `parseFlags` call.
- Added focused observable tests for native registration, both actions, every current flag/output-directory form, help, JSON, unknown flags, invalid actions/errors, assigned globals/config, byte parity, default paths, and safe temp-root filesystem containment.
- Kept `cli.Run`, canonical docs-map ownership, checked-in docs, website, goldens, dependencies, connector definitions, unrelated namespaces, Phase 14 viewer, and Phase 19 help/man behavior unchanged.

## TDD and verification

Exact RED (`11.332s`) showed docs still used `DisableFlagParsing` and the native command count was short. Focused GREEN passed in `11.462s`; final focused/default-path tests in `13.710s`; router/golden in `20.158s`; full CLI in `227.224s`; standalone golden in `5.470s`. gofmt, vet, full tests, build, and final `make verify` passed. Full tests included CLI `229.851s` and certify `342.890s`; final verify ended with smoke OK, lint `0 issues`, and `connectorgen validate: 547 connector(s) checked, 0 findings`.

Built-binary parity: `help_bytes=818`, invalid action exit `2`, missing generate dir exit `1`, assigned false JSON/plain error exit `2`; generated CLI docs diff and connector validation clean. Website generator wrote 11 pages with no diff. Dependency, connector-def, docs, website, golden, scope, path-containment, and whitespace guards passed.

## Workflow and safety

GSD doctor/list/plan-phase passed. Programming-loop is absent (`scripts/gsd: unknown GSD command: programming-loop`), so the recorded manual universal-loop fallback enforced strict TDD. Verify-work prompt generation passed (7133 bytes) and local verification found no scoped blocker.

No secrets, external services, credentialed checks, dependencies, unrelated namespace changes, Phase 14/19 churn, external review, PR, or merge. Focused writes stayed in temporary roots. `make verify` used only its existing local temporary smoke and followed reverse ETL plan → preview → approval → run.

Planning, RED, implementation, test-hardening, and final verification checkpoints are pushed to `origin/refactor/427-docs-native-cobra`; the final state-recording commit is the handoff HEAD.
