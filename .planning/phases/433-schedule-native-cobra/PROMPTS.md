# Phase 433 Prompts

## Kickoff snapshot

Task: Implement polymetrics-ai/cli#433 as the next serialized Phase 9 unit under #407/#397 from exact parent HEAD `ab1c79eede67fa87e1c6b808d6ddba0b27fcf00d`, using isolated branch `refactor/433-schedule-native-cobra`, Sol/high explicit, no credentials/services/dependencies/unrelated changes/real scheduler effects/PR/review.

Identity: session `issue-433-pi-sol-high-20260719T044819Z`; explicit model profile `Sol`; thinking `high`; start `20260719T044819Z` UTC.

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd prompt plan-phase 433 --skip-research
scripts/gsd prompt programming-loop init --phase 433 --dry-run
```

Doctor/list passed and the plan prompt was generated. `programming-loop` is absent from the adapter registry, so the manual GSD universal-runtime-loop fallback enforces plan-before-production and strict RED → GREEN → refactor.

Execution decision: `local_critical_path` — assigned isolated serialized namespace worktree; central router scope collides with siblings; this session exposes no subagent tool; the user restricted delivery to #433 implementation/commit/push and prohibited PR/review/services/dependencies/real scheduler effects.

Required skills: `gsd-core`; `golang-how-to`; `golang-cli`; `golang-testing`; `golang-error-handling`; `golang-security`; `golang-safety`; `golang-context`; `golang-documentation`; `golang-spf13-cobra`.

Safety prompt: use only temporary schedule roots, redirected temporary crontab files, fixed clocks, and injected fake backends. Never request, print, summarize, store, or log secrets. Never execute `crontab`, `launchctl`, `systemctl`, Temporal, or another scheduler. Preserve current create/list/install/remove, cron/backend selection, context, payload, cleanup, output, and exit semantics. Do not implement Phase 11's interactive schedule wizard.

Contract clarification: the current issue/repository exposes `remove`, not `uninstall`, and has no `schedule run` or `schedule history`. Preserve that contract: focused tests cover `uninstall`, `run`, and `history` as invalid action heads with zero backend calls rather than adding out-of-scope commands.

Downstream artifact: complete. Test-only RED preceded the native schedule Cobra tree, typed handlers, invocation-private operand capture, schedule-only normalization, and injected runtime seam. Only the schedule wrapper and schedule `parseFlags` call sites were removed. Current `remove` behavior remains the uninstall operation; out-of-scope `uninstall`, `run`, and `history` action heads remain effect-free usage errors.

Verification route: `scripts/gsd prompt verify-work 433` generated 106 lines and was executed inline under the manual universal loop.

Verification result: pass at implementation head `7b20f9fe`. Two exact-start differentials matched 248/248 cases. Focused/repeated/race/router/golden/full CLI and schedule packages, runtime help, generated docs/website, gofmt, vet, full repository tests, build, scope/dependency guards, and `make verify` pass. No real scheduler, service, live credential, dependency, PR, or review was used.
