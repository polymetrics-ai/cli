# Phase 425 Prompts

## Kickoff snapshot

Task: Implement polymetrics-ai/cli#425 as the fifth serialized Phase 9 unit under #407/#397 from exact parent HEAD `479a62f930e7c8a9a51ba0b3deb088bf3aad3ecc`, with no PR or external review request.

Identity: session `issue-425-pi-openai-codex-gpt-5.6-sol-high-20260718T095316Z`; explicit model `openai-codex/gpt-5.6-sol`; explicit thinking `high`; branch `refactor/425-version-native-cobra`.

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd prompt plan-phase 425 --skip-research --model=openai-codex/gpt-5.6-sol --thinking=high
scripts/gsd prompt programming-loop init --phase 425 --dry-run --model=openai-codex/gpt-5.6-sol --thinking=high
```

Doctor/list passed; plan prompt generated. Programming-loop is absent (`scripts/gsd: unknown GSD command: programming-loop`, exit 1), so the manual GSD universal-loop fallback was used.

Execution decision: `local_critical_path` — assigned isolated serial namespace worktree; central router scope collides with later namespace units; no subagent tool is exposed; user restricted work to #425.

Downstream artifact: all six phase files plus `internal/cli/cobra_router.go`, `cobra_router_test.go`, `version.go`, and `version_cli_test.go`.

Verification result: pass — exact RED captured before production edits; focused and full CLI tests, full repository tests, gofmt, vet, build, `make verify`, built-binary parity, docs/website/golden checks, dependency/scope checks, and branch pushes completed without external review or PR creation.

## Verification/review snapshot

```bash
scripts/gsd prompt verify-work 425 --model=openai-codex/gpt-5.6-sol --thinking=high >/tmp/gsd-verify-work-425.prompt
scripts/gsd prompt code-review 425 --model=openai-codex/gpt-5.6-sol --thinking=high >/tmp/gsd-code-review-425.prompt
```

Both prompts generated (7292 and 6158 bytes). They were executed locally through the recorded manual loop: all declared gates passed and scoped diff review found no actionable issue. No Claude, Copilot, PR, or other external review was requested.

Execution decision: `local_critical_path` — local verification/review of one bounded already-isolated diff.
