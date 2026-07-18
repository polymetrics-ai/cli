# Phase 427 Prompts

## Kickoff snapshot

Task: Implement polymetrics-ai/cli#427 as the seventh serialized Phase 9 unit under #407/#397 from exact parent HEAD `ab847da28ebf78e5732ac1edcde8e39f92dc2656`, using isolated branch `refactor/427-docs-native-cobra`, with no PR or external review.

Identity: session `issue-427-pi-openai-codex-gpt-5.6-sol-high-20260718T112639Z`; explicit model `openai-codex/gpt-5.6-sol`; explicit thinking `high`.

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd sources programming-loop
scripts/gsd prompt plan-phase 427 --skip-research
scripts/gsd prompt programming-loop init --phase 427 --dry-run
```

Doctor/list passed and the plan prompt generated/executed inline. Programming-loop is absent (`scripts/gsd: unknown GSD command: programming-loop`, exit 1), so the manual GSD universal-runtime-loop fallback is active.

Execution decision: `local_critical_path` — assigned isolated serialized namespace worktree; central router scope collides with later units; this session has no subagent tool; user restricted the invocation to #427 and prohibited PR/external review.

Downstream artifact: all six phase files plus `internal/cli/cobra_router.go`, `cobra_router_test.go`, `cli.go`, and `docs_cli_test.go`.

Verification result: pass — exact RED preceded production edits; focused/router/golden/full CLI, gofmt, vet, full repository tests, build, final `make verify`, built-binary help/output/error/config/global forms, temp docs generation/validation, website generation, and generated/golden/dependency/scope checks all passed.

## Verification snapshot

```bash
scripts/gsd prompt verify-work 427 > /tmp/gsd-verify-work-427.prompt
```

Prompt generation passed (7133 bytes) and was executed locally through the recorded manual loop. No Claude, Copilot, PR, or other external review was requested.

Execution decision: `local_critical_path` — local verification of one bounded already-isolated diff.
