# Phase 428 Prompts

## Accepted High correction kickoff snapshot

Task: From exact correction start `746b2a98b01ba1e119974e31569fc8deb06cd897`, prevent leading invalid agent/image action heads from allowing Cobra to discover or execute later image build/pull/ensure actions. Add fake-runtime RED first; preserve exact actions, agent help, and literal/action-tail compatibility; run bounded local gates only.

Identity: session `issue-428-review-fix-pi-openai-codex-gpt-5.6-sol-high-20260718T132841Z`; explicit model `openai-codex/gpt-5.6-sol`; explicit thinking `high`.

GSD doctor/list passed. `scripts/gsd prompt programming-loop init --phase 428-agent-native-cobra --dry-run` again failed because the registry has no `programming-loop` command, so the manual universal-runtime-loop fallback remains active.

Execution decision: `local_critical_path` — urgent accepted finding in the isolated serialized router scope; no subagent tool, container/services, PR, or external review.

Downstream artifact: `internal/cli/agent_cli_test.go`, `internal/cli/cobra_router.go`, and finalized issue-local phase artifacts. Planning, RED, and implementation checkpoints were pushed to the active branch.

Verification result: pass. Focused RED failed as expected before production edits (`0.587s`). GREEN makes all 30 fake-runtime cases return usage with zero lookups/files/runs. Focused agent/router (`4.446s`), race (`1.679s`), repeated boundary (`0.582s`), 35/35 exact base differential, full CLI (`234.335s`), gofmt, vet, build, diff, scope, and dependency gates pass. No container/service or external action ran.

## Kickoff snapshot

Task: Implement polymetrics-ai/cli#428 as the eighth serialized Phase 9 unit under #407/#397 from exact parent HEAD `235233f7cfde4a24612be6b0f95fb37a412d388a`, using isolated branch `refactor/428-agent-native-cobra`, with no PR or external review.

Identity: session `issue-428-pi-openai-codex-gpt-5.6-sol-high-20260718T124925Z`; explicit model `openai-codex/gpt-5.6-sol`; explicit thinking `high`; verification end `20260718T131634Z`.

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd sources programming-loop
scripts/gsd prompt plan-phase 428 --skip-research
scripts/gsd prompt programming-loop init --phase 428 --dry-run
```

Doctor/list passed and the plan prompt generated 10668 bytes for inline execution. Programming-loop is absent (`scripts/gsd: unknown GSD command: programming-loop`, exit 1), so the manual GSD universal-runtime-loop fallback enforced plan-before-production and strict RED → GREEN → refactor.

Execution decision: `local_critical_path` — assigned isolated serialized namespace worktree; central router scope collides with later units; this session has no subagent tool; user restricted the invocation to #428 and prohibited PR/external review.

Downstream artifact: `internal/cli/cobra_router.go`, `cobra_router_test.go`, `cli.go`, `agent_image_cli.go`, `agent_cli_test.go`, and all six phase files.

Verification result: pass — exact RED preceded production edits; focused/router/golden/full CLI and repository tests, focused race, 25-case exact legacy differential, built help/plan/error/global checks, runtime dependency-free tests, docs/website/generated parity, gofmt, vet, build, `make verify`, and scope/dependency guards passed. No Podman/Docker or image operation ran.

## Verification and local review snapshot

```bash
scripts/gsd prompt verify-work 428 > /tmp/gsd-verify-work-428.prompt
scripts/gsd prompt code-review 428 > /tmp/gsd-code-review-428.prompt
```

Prompt generation passed (7137 and 6003 bytes). Both were executed inline under the manual universal loop because no subagent tool is available. Local diff/error/security/safety review found no actionable issue. No Claude, Copilot, PR, reviewer subagent, or other external review was requested.

Execution decision: `local_critical_path` — local verification/review of one bounded already-isolated diff; external review prohibited by the user.
