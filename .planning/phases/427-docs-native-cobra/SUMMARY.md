# Phase 427 Summary

Status: planned; implementation and verification pending.

## Identity

- Session: `issue-427-pi-openai-codex-gpt-5.6-sol-high-20260718T112639Z`
- Model/thinking: `openai-codex/gpt-5.6-sol`, `high`
- Branch: `refactor/427-docs-native-cobra`
- Exact start: `ab847da28ebf78e5732ac1edcde8e39f92dc2656`
- Parent: #397; umbrella: #407; draft parent PR #438

## Planned delivery

- Register `docs`, `docs generate`, and `docs validate` as native Cobra nodes with declared repeated path flags, legacy bare-flag semantics, unknown-flag tolerance, hidden positional help, and no-file completion seam.
- Preserve canonical manual/help, global/config booleans, error taxonomy, exact output text, generated CLI bytes and connector artifacts, validation behavior, and safe temp-root filesystem checks.
- Remove `docs` from legacy wrappers and remove only its now-unused `parseFlags` call.
- Add focused observable tests for native registration, every current action/flag/output-dir form, help, JSON, unknown flags, invalid actions/errors, assigned globals/config, byte parity, and filesystem containment.
- Keep `cli.Run`, docs map ownership, checked-in docs/website/goldens, dependencies, connector definitions, unrelated namespaces, Phase 14 viewer, and Phase 19 help/man surfaces unchanged unless parity evidence requires a scoped update.

## Workflow and safety

GSD doctor/list/plan-phase passed. Programming-loop is absent (`scripts/gsd: unknown GSD command: programming-loop`), so the recorded manual universal-loop fallback will enforce strict TDD.

No secrets, external services, credentialed checks, dependencies, unrelated namespace changes, external review, PR, or merge.
