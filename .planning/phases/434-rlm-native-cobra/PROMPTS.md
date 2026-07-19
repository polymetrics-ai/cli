# Phase 434 Prompts

## Kickoff snapshot

Task: Implement polymetrics-ai/cli#434 as the next serialized Phase 9 unit under #407/#397 from exact parent HEAD `2ac457a163cbd7bc9a3708da88b03d375ec5e952`, using isolated branch `refactor/434-rlm-native-cobra`, Sol/high explicit, no dependencies/credentials/services/model/Temporal/Podman/PR/review.

Identity: session `issue-434-pi-sol-high-20260719T053630Z`; explicit model profile `Sol`; thinking `high`; start `20260719T053630Z` UTC.

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd prompt plan-phase 434 --skip-research
scripts/gsd prompt programming-loop init --phase issue-434 --dry-run
```

Doctor/list passed and the plan prompt was generated. `programming-loop` is absent from the adapter registry, so the manual GSD universal-runtime-loop fallback enforces plan-before-production and strict RED → GREEN → refactor.

Execution decision: `local_critical_path` — assigned isolated serialized namespace worktree; central router scope collides with siblings; this session exposes no subagent tool; the user restricted delivery to #434 implementation/commit/push and prohibited PR/review/dependencies/credentials/services/model/Temporal/Podman calls.

Required skills: `gsd-core`; `golang-how-to`; `golang-cli`; `golang-testing`; `golang-error-handling`; `golang-security`; `golang-safety`; `golang-context`; `golang-concurrency`; `golang-documentation`; `golang-spf13-cobra`.

Safety prompt: use only temporary spec/warehouse paths and injected analyzer/factory fakes or existing hermetic fake runner paths. Never request, print, summarize, store, or log secrets or request contents. Never call a model, Temporal, Podman, worker service, or another external service. Preserve dependency-free deterministic/fixture behavior and optional agent configuration. Do not expose a generic runner and do not implement Phase 16's RLM viewer.

Downstream artifact: complete. Focused test-only RED preceded the native RLM run/help tree, all six typed flags, RLM-only normalization, typed handler, and injected analyzer factory. Only the RLM wrapper/dispatcher/`parseFlags` call were removed. Focused/repeated/race/analyzer/worker-fake/router/golden/full CLI and 24/24 exact-start differential gates pass.

Verification route: `scripts/gsd prompt verify-work 434` generated 106 lines and was executed inline under the manual universal loop.

Verification result: pass at implementation head `633f1e21`. Exact-start differential matched 24/24 cases. Focused/repeated/race/router/golden/full CLI, RLM, worker-fake, runtime help, generated docs/website, gofmt, vet, full repository tests, build, scope/dependency guards, and `make verify` pass. No model, Temporal, Podman, worker service, optional service, live credential, dependency, generic runner, PR, or review was used.
