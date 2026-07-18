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

Downstream artifact: pending all six phase files plus focused tests and the smallest docs/Cobra handler changes.

Verification result: pending strict RED before production edits, focused/router/golden/full CLI, docs and website generation diffs, gofmt, vet, full repository tests, build, and `make verify`.

## Verification snapshot

Pending `scripts/gsd prompt verify-work 427` after GREEN implementation.
