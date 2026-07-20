# Phase 436 Prompts

## Kickoff snapshot

Task: Implement polymetrics-ai/cli#436 as the next serialized Phase 9 unit under #407/#397 from exact parent HEAD `eec03373dcc581c7f5c3331fe63287519b317f53`, using isolated branch `refactor/436-extract-native-cobra`, Sol/high explicit, no external files/services/credentials/dependencies/PR/review.

Identity: session `issue-436-pi-sol-high-20260719T074902Z`; explicit model profile `Sol`; thinking `high`; start `20260719T074902Z` UTC.

```bash
scripts/gsd doctor
scripts/gsd list
scripts/gsd prompt plan-phase 436 --skip-research
scripts/gsd prompt programming-loop 436
```

Doctor/list and plan prompt generation passed. `programming-loop` is absent from the adapter registry, so the manual GSD universal-runtime-loop fallback enforces six artifacts before production and strict RED → GREEN → refactor.

Execution decision: `local_critical_path` — assigned isolated serialized namespace worktree; central router scope collides with siblings; this session exposes no subagent tool; the user restricted delivery to #436 implementation/commit/push and prohibited PR/review/external files/services/credentials/dependencies.

Required skills: `gsd-core`; `golang-how-to`; `golang-cli`; `golang-testing`; `golang-error-handling`; `golang-security`; `golang-safety`; `golang-project-layout`; `golang-documentation`; `golang-spf13-cobra`.

Safety prompt: keep extract hidden and dependency-free. Inject local fakes for query/analyzer behavior. Use only temporary project/warehouse roots and synthetic records. Reject broad input/output names and effect-time symlink escapes; preserve external sentinel files. Never call a model, Temporal, Podman, worker, network listener, database service, credentialed connector, generic shell/HTTP/SQL write tool, destructive/admin action, production operation, or reverse ETL. Remove only extract's parser registration/call.

Downstream artifact: hidden native extract with typed current flags, injected query/analyzer seams, bounded table validation, rooted RLM warehouse input/output effects, canonical manual/docs/website/golden parity, and extract-only legacy/parser removal.

Verification result: correction complete at implementation head `28393538`. All four findings from `/tmp/pm-397-review-436.log` were accepted. The held narrow project-rooted RLM warehouse scope, parser-owned literal/malformed/help-like tails, exact-bare detection, and complete pre-factory RLM table validation pass deterministic RED→GREEN, focused/repeated/race/adversarial tests, exact-base 11/11, full CLI/RLM/safety, runtime help/output/hidden checks, docs/manual/website drift checks, gofmt/vet/build/module/scope guards, and final `make verify` (CLI `433.304s`, certify `336.725s`, RLM `1.780s`). No external user files/services, dependencies, credentials, PR, or review.
