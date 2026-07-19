# Phase 435 Prompts

## Kickoff snapshot

Task: Implement polymetrics-ai/cli#435 as the next serialized Phase 9 unit under #407/#397 from exact parent HEAD `14c02d295065c3bf33c65eaac5f8d36642798f81`, using isolated branch `refactor/435-worker-native-cobra`, Sol/high explicit, no dependencies/credentials/services/Temporal/Podman/PR/review.

Identity: session `issue-435-pi-sol-high-20260719T064417Z`; explicit model profile `Sol`; thinking `high`; start `20260719T064417Z` UTC.

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd prompt plan-phase 435 --skip-research
scripts/gsd prompt programming-loop init --phase 435 --dry-run
```

Doctor/list passed and the plan prompt was generated. `programming-loop` is absent from the adapter registry, so the manual GSD universal-runtime-loop fallback enforces six artifacts before production and strict RED → GREEN → refactor.

Execution decision: `local_critical_path` — assigned isolated serialized namespace worktree; central router scope collides with siblings; this session exposes no subagent tool; the user restricted delivery to #435 implementation/commit/push and prohibited PR/review/dependencies/credentials/services/Temporal/Podman.

Required skills: `gsd-core`; `golang-how-to`; `golang-cli`; `golang-testing`; `golang-error-handling`; `golang-security`; `golang-safety`; `golang-context`; `golang-concurrency`; `golang-documentation`; `golang-spf13-cobra`.

Safety prompt: inject invocation-local fake status/probe and serve seams. Never dial Temporal, start a Temporal worker or listener, invoke Podman, access a database, or start runtime services. Assert no fake starts on bare/help/invalid/malformed-config paths. Never print config canaries. Keep the worker hidden and typed to the RLM Temporal workflow only; never add a generic runner.

Downstream artifact: RED and GREEN/refactor complete. Native hidden worker owns status/serve/help with invocation-local fakes, strict action parsing, contextual help, config/cancellation/nondisclosure coverage, and no generic runner. Full default-only gates and handoff remain pending.

Verification result: focused/repeated/race/worker-fake/router/golden/full CLI, exact-start differential, runtime help, and docs/website checks pass; `RUN-STATE.json` remains non-terminal with `verificationPassed=false` until full `make verify` exits 0.
