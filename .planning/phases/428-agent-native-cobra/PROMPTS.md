# Phase 428 Prompts

## Kickoff snapshot

Task: Implement polymetrics-ai/cli#428 as the eighth serialized Phase 9 unit under #407/#397 from exact parent HEAD `235233f7cfde4a24612be6b0f95fb37a412d388a`, using isolated branch `refactor/428-agent-native-cobra`, with no PR or external review.

Identity: session `issue-428-pi-openai-codex-gpt-5.6-sol-high-20260718T124925Z`; explicit model `openai-codex/gpt-5.6-sol`; explicit thinking `high`.

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd sources programming-loop
scripts/gsd prompt plan-phase 428 --skip-research
scripts/gsd prompt programming-loop init --phase 428 --dry-run
```

Doctor/list passed and the plan prompt generated 10668 bytes for inline execution. Programming-loop is absent (`scripts/gsd: unknown GSD command: programming-loop`, exit 1), so the manual GSD universal-runtime-loop fallback is active.

Execution decision: `local_critical_path` — assigned isolated serialized namespace worktree; central router scope collides with later units; this session has no subagent tool; user restricted the invocation to #428 and prohibited PR/external review.

Downstream artifact: pending native agent production/tests plus all six phase files.

Verification result: pending.
