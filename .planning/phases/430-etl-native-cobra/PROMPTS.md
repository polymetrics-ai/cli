# Phase 430 Prompts

## Kickoff snapshot

Task: Implement polymetrics-ai/cli#430 as the next serialized Phase 9 unit under #407/#397 from exact parent HEAD `6c94754c58185df5aac53bd97587603c3154b1d5`, using isolated branch `refactor/430-etl-native-cobra`, Sol/high explicit, no PR/review/services/reverse execution.

Identity: session `issue-430-pi-openai-codex-gpt-5.6-sol-high-20260718T225346Z`; explicit model `openai-codex/gpt-5.6-sol`; thinking `high`; start `20260718T225346Z` UTC.

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd prompt plan-phase 430 --skip-research
scripts/gsd prompt programming-loop init --phase 430 --dry-run
```

Doctor/list passed and the plan prompt was generated. Programming-loop is absent (`scripts/gsd: unknown GSD command: programming-loop`), so the manual GSD universal-runtime-loop fallback enforces plan-before-production and strict RED → GREEN → refactor.

Execution decision: `local_critical_path` — assigned isolated serialized namespace worktree; central router scope collides with sibling units; this session exposes no subagent tool; the user restricted work to #430 and prohibited PR/review/services.

Required skills: `gsd-core`; `golang-how-to`; `golang-cli`; `golang-testing`; `golang-error-handling`; `golang-security`; `golang-safety`; `golang-context`; `golang-concurrency`; `golang-documentation`; `golang-spf13-cobra`.

Safety prompt: use only fixture/local temporary connectors and roots. Never request, print, summarize, or store secrets. Do not contact external connectors or optional runtime services; do not execute reverse ETL; do not add dependencies or unrelated changes. Preserve bounded batches, dependency-free default, cancellation, event/telemetry context, output/error taxonomy, and legacy help/literal compatibility.

Downstream artifact: focused test-only RED, ETL native Cobra command/typed handler adaptation, test-first invalid-action help correction, broad verification, and finalized six issue-local phase artifacts are complete.

Verification route: `scripts/gsd prompt verify-work 430` generated 7129 bytes and was executed inline under the manual universal loop.

Verification result: pass. Initial focused RED failed before production edits at `internal/cli/etl_cli_test.go:22:9: undefined: newETLCobraCommand`, as required. Focused GREEN passed, then local review's invalid-action trailing-help test failed before its correction. Focused/repeated/race/router/golden/full CLI/app/repository, 20-case exact-start differential, runtime help, generated docs/website, gofmt, vet, build, scope/dependency guards, and final `make verify` pass. No external connector/service, secret, dependency, standalone reverse operation, PR, or review occurred; final `make verify` used its existing temporary-root local approval-gated smoke.

## Bounded correction snapshot

Task: Correct #430 from exact start `9b0020abde7cd7e007412a0294db6e4cb576f5f3` using `/tmp/pm-397-review-430.log`: preserve the exact legacy first run-ID operand for `etl status` before shared normalization, including `--help`, `-h`, literal `--`, and unknown flag-shaped values followed by a valid ID.

GSD route: doctor/list passed. `scripts/gsd prompt programming-loop 430-etl-status-operand-correction` failed because the registered adapter still lacks that command; manual universal-loop fallback recorded. `scripts/gsd prompt quick "Issue 430 bounded correction ..."` generated 5778 bytes and is executed inline.

Execution decision: `local_critical_path` — coupled status/router/test correction in the assigned isolated worktree; no disjoint worker scope and no subagent tool. Required skills: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-spf13-cobra`.

Safety: temporary local project/run only; no secrets, credentials, optional services, external connectors, dependencies, reverse operation, PR, or review. Preserve all non-status ETL/help behavior and byte-identical docs/manual surfaces.

Downstream artifact: correction planning and focused differential RED are complete before production edits; private capture GREEN, verification, and final delivery remain pending.

Verification result: RED captured. All five differential first-operand cases diverged as expected: help/short/literal/internal-carrier-shaped operands reached the later valid ID, while unknown-flag returned usage instead of the exact base lookup error. `verificationPassed=false` until focused/adversarial/race/base differential/full CLI/repository, gofmt/vet/build/diff, help/docs parity, and `make verify` pass.
