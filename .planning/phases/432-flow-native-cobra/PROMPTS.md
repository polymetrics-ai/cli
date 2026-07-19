# Phase 432 Prompts

## Kickoff snapshot

Task: Implement polymetrics-ai/cli#432 as the next serialized Phase 9 unit under #407/#397 from exact parent HEAD `ec12c1729e0aaf233a853eff5c6291885f910b15`, using isolated branch `refactor/432-flow-native-cobra`, Sol/high explicit, no credentials/services/dependencies/unrelated changes/action writes/PR/review.

Identity: session `issue-432-pi-openai-codex-gpt-5.6-sol-high-20260719T034344Z`; explicit model `openai-codex/gpt-5.6-sol`; thinking `high`; start `20260719T034344Z` UTC.

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd prompt plan-phase 432 --skip-research
scripts/gsd prompt programming-loop init --phase 432 --dry-run
```

Doctor/list passed and the plan prompt was generated. `programming-loop` is absent from the adapter registry, so the manual GSD universal-runtime-loop fallback enforces plan-before-production and strict RED → GREEN → refactor.

Execution decision: `local_critical_path` — assigned isolated serialized namespace worktree; central router scope collides with siblings; this session exposes no subagent tool; the user restricted delivery to #432 implementation/commit/push and prohibited PR/review/services/dependencies/external writes.

Required skills: `gsd-core`; `golang-how-to`; `golang-cli`; `golang-testing`; `golang-error-handling`; `golang-security`; `golang-safety`; `golang-context`; `golang-concurrency`; `golang-documentation`; `golang-spf13-cobra`.

Safety prompt: use only temporary manifests, temporary project roots, dependency-free local execution, and in-memory fakes. Never request, print, summarize, store, or log secret/approval values. Do not contact external connectors or optional runtime services. Do not execute action steps, reverse ETL, generic HTTP writes, or generic SQL writes. Preserve current directory, manifest/DAG, cancellation, events, telemetry, checkpoint, ledger, output, and exit semantics. Do not implement Phase 10 dashboards, Phase 11 flow create, or Phase 19 help/man churn.

Downstream artifact: complete. Test-only RED preceded the native flow Cobra tree and typed handlers. A 200-case exact-start differential then exposed 20 bounded pflag gaps; an eight-case correction RED preceded invocation-private operand capture and flow-only normalization. Only the flow wrapper/parser was removed. The 200/200 differential, focused/repeated/race/router/golden/full CLI and flow/event/telemetry packages, runtime help, generated docs/website, gofmt, vet, full repository tests, build, scope/dependency guards, and `make verify` pass.

Verification route: `scripts/gsd prompt verify-work 432` generated 106 lines and was executed inline under the manual universal loop.

Verification result: pass at implementation head `e61cae17`. No action flow step, external write/service, live credential, dependency, PR, or review was used; no public docs/golden artifact changed.
