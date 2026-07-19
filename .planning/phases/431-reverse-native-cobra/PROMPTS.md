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

Original verification result: pass at implementation head `f5aeafb7bb7a6702077382e98acb790d3865073f`. Focused/repeated/race/router/golden/reverse-app/full CLI/repository, 21-case exact-start differential, runtime help, generated docs/website, gofmt, vet, build, scope/dependency guards, and final `make verify` passed.

## Parser compatibility correction snapshot

Task: From exact clean/upstream head `c8f5b9e97a2f71f25cdb362af0055c1c31dc8420`, read `/tmp/pm-397-review-431.log`; add RED differential tests across every reverse action for malformed legacy-accepted unknown `--=x`, `---x`, and representative variants; prove preserved outcomes and no effects; then normalize only malformed unknown tail tokens before pflag. Preserve valid flags, first operands, legal unknown behavior, approval/confirmation ordering, and strict plan → preview → approval → execute.

Identity: session `issue-431-parser-compat-20260719T022304Z`; correction start `20260719T022304Z` UTC.

GSD route: `scripts/gsd doctor` and `scripts/gsd list` passed. `scripts/gsd prompt programming-loop init --phase 431 --dry-run` failed because the adapter registry has no `programming-loop`; manual universal-runtime-loop fallback is active. Execution decision: `local_critical_path` for the bounded correction in this isolated worktree; no subagent tool is exposed.

Skills: `gsd-core`; `golang-how-to`; `golang-cli`; `golang-testing`; `golang-error-handling`; `golang-security`; `golang-safety`; `golang-design-patterns`; `golang-structs-interfaces`; `golang-spf13-cobra`. CLI help/docs parity reviewed: no public-surface change, so docs/website edits are not applicable.

Downstream artifact: complete. Test-only RED reproduced all 50 malformed/action mismatches before production edits. Malformed-only normalization, focused and complete reverse tests, focused race, full CLI, 324-case exact-start differential, no-approval-output/state/outbox checks, and gofmt/vet/build/diff/scope/dependency gates pass. RED `c98e4dad` and GREEN `bbe9bb9c` are pushed; this terminal artifact checkpoint completes delivery on push. No external write/service/dependency/PR/review.

Verification result: pass at implementation head `bbe9bb9c`. Parser differential is 324/324 exact with unchanged state/outbox and no approval output; focused correction `26.289s`, reverse focus `63.513s`, race `295.302s`, full CLI `417.589s`; gofmt, full vet, build, and diff checks pass.
