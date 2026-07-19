# Phase 431 Prompts

## Kickoff snapshot

Task: Implement polymetrics-ai/cli#431 as the next serialized Phase 9 unit under #407/#397 from exact parent HEAD `0b03361e3ec5082d54c416a31715851f71e845fa`, using isolated branch `refactor/431-reverse-native-cobra`, Sol/high explicit, no credentials/services/dependencies/unrelated changes/PR/review/external writes.

Identity: session `issue-431-pi-openai-codex-gpt-5.6-sol-high-20260719T010451Z`; explicit model `openai-codex/gpt-5.6-sol`; thinking `high`; start `20260719T010451Z` UTC.

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd prompt plan-phase 431 --skip-research
scripts/gsd prompt programming-loop init --phase 431 --dry-run
```

Doctor/list passed and the plan prompt was generated. `programming-loop` is absent from the adapter registry, so the manual GSD universal-runtime-loop fallback enforces plan-before-production and strict RED → GREEN → refactor.

Execution decision: `local_critical_path` — assigned isolated serialized namespace worktree; central router scope collides with sibling units; this session exposes no subagent tool; the user restricted work to #431 and prohibited PR/review/services/dependencies/external writes.

Required skills: `gsd-core`; `golang-how-to`; `golang-cli`; `golang-testing`; `golang-error-handling`; `golang-security`; `golang-safety`; `golang-context`; `golang-documentation`; `golang-spf13-cobra`.

Safety prompt: use only local fakes, built-in local connectors, and temporary roots. Never request, print, summarize, store, or log secret/approval values. Do not contact external connectors or optional runtime services. Do not execute any external write. Preserve strict plan → preview → approval → execute, typed confirmation, single-use approval, exact error taxonomy, output contracts, and legacy help/literal/operand compatibility. The repository's existing ordered reverse smoke may run only inside final `make verify`.

Downstream artifact: focused test-only RED, native reverse Cobra command/typed handler adaptation, private operand ownership, reverse-only parser removal, broad verification, and finalized six issue-local artifacts are complete.

Verification route: `scripts/gsd prompt verify-work 431` generated 106 lines and was executed inline under the manual universal loop.

Verification result: pass at implementation head `f5aeafb7bb7a6702077382e98acb790d3865073f`. Focused/repeated/race/router/golden/reverse-app/full CLI/repository, 21-case exact-start differential, runtime help, generated docs/website, gofmt, vet, build, scope/dependency guards, and final `make verify` pass. No approval value, external write, connector/service, live credential, dependency, unrelated change, PR, or review occurred; final `make verify` used only its established temporary-root ordered reverse smoke.
