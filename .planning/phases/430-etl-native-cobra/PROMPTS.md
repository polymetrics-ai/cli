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

Downstream artifact: focused test-only RED checkpoint and ETL native Cobra command/typed handler adaptation are complete; broad refactor/parity verification and finalized six issue-local phase artifacts remain downstream.

Verification result: initial focused RED failed before production edits at `internal/cli/etl_cli_test.go:22:9: undefined: newETLCobraCommand`, as required. Focused GREEN passed in `13.396s`, and the broader ETL/router focused suite passed in `27.999s`. Local review then added a test-first invalid-action trailing-help case; it failed because Cobra rendered `CommandManual` and exited 0. `RUN-STATE.json` remains non-terminal with `verificationPassed=false` until the correction and complete declared gates pass.
