# Issue #397 Wave 1 Parent Synchronization Plan

Status: active
Owner: Wave 1 integration worker
Branch: `fm/cli-architecture-v2-wave1-parent-sync-r1`
Stacked base: `feat/cli-architecture-v2`
Parent PR: #438 (draft; human-only)
Scope: synchronize current `main` into the published parent through an ordinary merge and publish a draft stacked PR. Issue #408 is excluded.

## Pinned pre-merge truth

- `origin/main`: `873cd7b251f70c4a35a607a0d4e86051ea0fbd15`
- `origin/feat/cli-architecture-v2`: `21d195aff0c7bd60b3bf54f14b1ce165cec9e03f`
- PR #438 head: `21d195aff0c7bd60b3bf54f14b1ce165cec9e03f`
- merge base: `74ab381eb8236305170ffd44d5aed74f8d0d2936`
- `origin/main` ancestor of parent: no
- #462 human approval: https://github.com/polymetrics-ai/cli/pull/468#issuecomment-5054325561 (approved exact PR #468 head `a4ca9813d44c79bbbc0a8c499b74253f38265aa8`)
- #419: `deferred_by_human`; no implementation or dependency approval
- #425–#436 exact-range review/process waiver: pending and out of scope

## GSD and skills

- `scripts/gsd doctor`: pass
- `scripts/gsd list`: 69 commands
- `scripts/gsd sources` without a command correctly reported that a command/prompt is required; sources for `plan-phase`, `execute-phase`, `verify-work`, and `code-review` resolve to the pinned registry/lock/official command docs.
- `scripts/gsd prompt plan-phase 397 --skip-research`: generated 10,536 bytes.
- `scripts/gsd prompt programming-loop init --phase 397-wave1-parent-sync-r1 --dry-run`: unavailable (`unknown GSD command: programming-loop`).
- Manual lifecycle: PLAN → RED → GREEN → REFACTOR → VERIFY → REVIEW → INTEGRATE.
- Loaded: `cli-architecture-v2-delivery` from in-flight commit `2ba47dc3036a2b063bd051d66d60b37ccea96bf6` without absorbing its files or commits; `gsd-core`; `golang-how-to`; `golang-cli`; `golang-testing`; `golang-error-handling`; `golang-safety`; `golang-security`; `golang-spf13-cobra`; `golang-spf13-viper`; `golang-documentation`.

## Write scope

Allowed:

- merge-conflict resolutions from `origin/main`;
- focused combined-behavior regression tests;
- this Wave 1 phase directory;
- narrow Wave 1 synchronization append/update in issue #397 parent orchestration artifacts.

Forbidden:

- #408 dashboards or other downstream phases;
- connector migration or new product behavior beyond preserving current main;
- files from the separate delivery-skill PR (skill/routing/Makefile validation changes);
- parent PR #438 edits/merge/readiness changes;
- Claude/Copilot requests, credentials, live connector calls, external mutations, deployments, or dependencies.

## Conflict inventory from merge-tree

Expected content conflicts:

1. `go.mod` / `go.sum`: union parent CLI Architecture v2 dependencies with current main's Gong requirements; resolve through module graph, then prove `go mod verify` and `go mod tidy -diff`.
2. `internal/cli/cli.go`: preserve Cobra/native namespace/config/event/telemetry routing and dynamic connector dispatch while retaining current Gong connector CLI behavior.
3. `internal/connectors/connectors.go`: preserve parent instrumentation/requester contracts and current main's Gong/declarative connector capability behavior.
4. `internal/connectors/connsdk/http.go`: preserve parent logging/telemetry/events/retry contracts and current main's Gong payload/query/multipart behavior.

Auto-merged related files, especially `internal/app/app.go` and `internal/connectors/connsdk/http_test.go`, must be inspected because successful textual merge is not semantic proof.

## Lifecycle

### PLAN

Pin Git/GitHub truth, inspect base/parent/main histories and every conflict, record safety and review constraints.

### RED

Run conflict-focused tests against the unresolved/naive merge where practical. Add combined regression tests where side-specific existing tests cannot prove both behaviors survive. The intended regression matrix includes:

- Gong dynamic connector routes through the Cobra root without losing native namespaces, config, progress/events, or certify re-entrancy;
- Gong metadata/operation/direct-read/write surfaces remain discoverable;
- connsdk JSON array, multipart, query, retry, telemetry, and redaction behavior coexist;
- golden stdout/stderr/JSON/exit codes and bare/help behavior;
- reverse plan → preview → approval → execute safety.

### GREEN

Resolve every conflict manually; do not choose an entire side blindly. Complete one ordinary `git merge --no-ff origin/main` merge commit. Do not rebase, reset, stash, or rewrite.

### REFACTOR

Keep the resolution minimal, gofmt touched Go, remove conflict markers, and preserve generated/module consistency without broad regeneration.

### VERIFY

Run focused conflict tests and representative CLI contracts, then all required gates:

```bash
gofmt -w cmd internal
git diff --exit-code -- cmd internal
git diff --check
go vet ./...
go test -timeout 20m ./...
go build ./cmd/pm
go mod verify
go mod tidy -diff
make verify
```

No credentialed or runtime-backed checks.

### REVIEW

Run a fresh-context read-only Codex/reviewer against exact base `21d195aff0c7bd60b3bf54f14b1ce165cec9e03f` and exact task head. Fix/disposition every finding and rerun after any head change. Then run read-only trajectory/Shepherd validation against exact evidence.

### INTEGRATE

Commit all evidence, push normally, open one draft stacked PR to `feat/cli-architecture-v2` with `Refs #397`, inspect checks/reviews, wait for required checks green, and do not merge or mark ready.

## Execution outcome

- Merge commit: `c545c3740c71b889fd2f1f64cec5491003f7b654` (`origin/main` is the second parent and an ancestor).
- Reconciliation commit after the full-test RED: `2a2e964b17144939b0a42f297de0d2b1c87383e1`.
- Conflict decisions:
  - modules: retained parent CLI/TUI/OTel dependencies and main Gong dependency graph;
  - CLI: retained parent Cobra shell/config/events/telemetry and delegated unknown connector paths plus connector help to Gong;
  - connector runtime: retained both requester/events/metrics policy and parent logger/telemetry policy;
  - HTTP: retained JSON/query requests, Gong multipart semantics, parent retry body cleanup, and retry metrics/log redaction.
- Full required commands and focused races are green. Representative help, bare namespace, native command, Gong command, and invalid-action contracts are green.
- Original synchronization review/Shepherd passed at exact head `3fd63fbe0f526873fa3adb8a75fa5f20342d52a6`; draft PR #495 is open on the identical convention-compliant branch head.
- Captain then authorized an additive PM-orchestrator correction in PR #495. Its separate evidence lives in `.planning/phases/397-pm-orchestrator-extension/`; prior exact-head review/check results are historical and final REVIEW/INTEGRATE gates must be rerun after the extension.
- PR #438 remains unchanged. #408 remains excluded.

## Checkpoints

1. plan-only checkpoint;
2. ordinary merge + conflict regressions green;
3. evidence/verification checkpoint;
4. review-fix checkpoint if needed;
5. final exact-head evidence and draft PR.
