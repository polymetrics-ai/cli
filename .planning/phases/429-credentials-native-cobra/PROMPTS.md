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

Downstream artifact: test-only initial and correction RED checkpoints, `internal/cli/credentials_cli.go`, credentials router/legacy-handler adaptation, focused security tests, and finalized six issue-local phase artifacts.

Verification result: pass. Initial focused GREEN passed. Local review then exposed a post-action name-discovery bypass; correction RED failed 8/10 cases before the fix, and corrected focused/repeated/race tests pass. Final full CLI (`275.269s`), preserved differential 28/28, built help/docs parity, website generation, gofmt, vet, build, full repository tests, and `make verify` pass. No real secret, credentialed external check, service, dependency, PR, or external review occurred.

## Verification and local review snapshot

```bash
scripts/gsd prompt verify-work 429 > /tmp/gsd-verify-work-429.prompt
scripts/gsd prompt code-review 429 > /tmp/gsd-code-review-429.prompt
```

Prompt generation passed (7161 and 6027 bytes). Both were executed inline under the manual universal loop. The first local security review produced the action-name boundary correction described in PLAN/TDD-LEDGER; the post-fix diff/error/security/safety review found no remaining actionable issue. External review was prohibited.
