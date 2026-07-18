# Phase 429 Prompts

## Kickoff snapshot

Task: Implement polymetrics-ai/cli#429 as the ninth serialized Phase 9 unit under #407/#397 from exact parent HEAD `0f1ec1e89cdae761e9da06ab9906fcc641b38e0a`, using isolated branch `refactor/429-credentials-native-cobra`, with Sol/high explicit, no PR, and no external review.

Identity: session `issue-429-pi-openai-codex-gpt-5.6-sol-high-20260718T143346Z`; explicit model `openai-codex/gpt-5.6-sol`; explicit thinking `high`; start `20260718T143346Z` UTC.

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd prompt plan-phase 429 --skip-research
scripts/gsd prompt programming-loop init --phase 429 --dry-run
```

Doctor/list passed and the plan-phase prompt was generated for inline execution. Programming-loop is absent (`scripts/gsd: unknown GSD command: programming-loop`, exit 1), so the manual GSD universal-runtime-loop fallback enforces plan-before-production and strict RED → GREEN → refactor.

Execution decision: `local_critical_path` — assigned isolated serialized namespace worktree; central router scope collides with later units; this session has no subagent tool; user restricted the invocation to #429 and prohibited PR/external review.

Required skills: `gsd-core`; `golang-how-to`; `golang-cli`; `golang-testing`; `golang-error-handling`; `golang-security`; `golang-safety`; `golang-documentation`; `golang-spf13-cobra`.

Safety prompt: Never request, read, print, summarize, store, or log real secret values. Tests may use opaque synthetic redaction fixtures only and must not include fixture content in diagnostics. Do not add interactive secret entry, dependencies, credentialed checks, external services, or unrelated changes. Preserve env/stdin-only intake, path containment, error/output contracts, action boundaries, and legacy help/literal compatibility.

Downstream artifact: test-only RED checkpoint, `internal/cli/credentials_cli.go`, credentials router/legacy-handler adaptation, focused security tests, and six issue-local phase artifacts. Full parity verification remains pending.

Verification result: focused GREEN. Credentials/router passed in `25.475s`; focused race subset in `111.267s`; golden in `5.513s`; exact preserved start differential 28/28. Full repository/parity gates pending.
