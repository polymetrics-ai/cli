# TDD Ledger: Batch R1 vNext source-lock cutover

## Planned evidence

| Slice | Red characterization | Green contract | Refactor/verification |
| --- | --- | --- | --- |
| Runtime dependency | Reference connectors require embedded authoring/certification/admission material. | GitHub, GitLab, and Asana load and reach credential/approval preflight from execution JSON alone. | Audit embedded/runtime reads and run focused plus fleet tests. |
| Connector-local invalidity | Global ledgers or one bundle error can suppress the fleet. | Malformed required execution JSON rejects that connector without hiding healthy connectors. | Assert stable typed diagnostics and deterministic discovery. |
| Canonical rendering | Existing source-lock paths do not own a single canonical all-lane projection. | One vNext model renders byte-stable existing execution JSON through shared schema refs. | Check every rendered file and reject stale output. |
| Lane semantics | Retention/certification state can hide documented commands and source operations cannot express all lanes canonically. | Direct, binary, ETL, reverse ETL, sync, and explicit-empty lanes are surfaced without provider switches. | Run the same all-lane contract for every Batch R1 connector. |

## Actual evidence

### 2026-09-02 — G0 direct-parent delivery amendment

- Authority and route: #4325 comment `5500153864` mandates the issue lifecycle and one-writer, ordinary fast-forward delivery to `fm/cli-top100-declaration-batch-r1`; #4294 comment `5500165004` is the authoritative routing correction. The attempted PR-body replacement failed before mutation on deprecated `projectCards`, so no PR-body mutation is claimed.
- Immutable baseline: `git fetch origin fm/cli-top100-declaration-batch-r1`; `git rev-parse HEAD`; and `git rev-parse origin/fm/cli-top100-declaration-batch-r1` each returned `d260b725ce6f53403961d7af1ef48ea6651cdd66`. `git merge-base --is-ancestor HEAD origin/fm/cli-top100-declaration-batch-r1` succeeded, and `git merge-base origin/main origin/fm/cli-top100-declaration-batch-r1` returned `813f457a925f7ee3fe3bea101a43e445992c8552`. The fixed #4325 denominator is 4,341 primary retained source operations.
- Local-state inspection: `git ls-tree -r --name-only HEAD -- internal/connectors/certifications`, `git ls-files --stage -- internal/connectors/certifications`, and `git log --all --full-history -- internal/connectors/certifications/.fingerprint-salt` returned no tracked item. The opaque untracked `.fingerprint-salt` resides below the legacy certification tree deleted by `0b214b79eeb871238ce8454cd7b896e71e2746a7`; it is local-only, unowned by the parent branch, not ignored, unstaged, unread, and unmodified. Its disposition is preserve in place and exclude from G0/N1, not recover, delete, certify, or admit.
- Green condition: the G0 evidence-only change is reviewed with `git diff --check`, committed without the local residue, normally pushed to the existing parent branch, and its remote SHA is read back before N1 starts. No production or test behavior changes belong to G0.

### 2026-09-02 — #4423 N1 executable proof baseline (planned)

- Scope: restore truthful executable proof without a runtime behavior change. The fixed denominator/base above remains immutable.
- Red: preserve the stale `defs` compile/API failure, demonstrate that each named reference/Atlas proof selector cannot silently match zero tests, and add a reference-lock characterization that fails on the frozen architectural defect.
- Green: repair only test/API drift and proof selectors; make the reference-lock characterization read authoring inputs, render the complete expected set in memory, and fail a frozen architectural defect without source-lock re-pin, execution-manifest rewrite, generated connector change, runtime change, credential use, or provider I/O.
- Required evidence: focused `defs`, `connectorgen`, `engine`, and named proof commands with `-count=1 -timeout 20m`; explicit test-count assertion; `git diff --check`; and a fresh exact-SHA review range. The green commit carries `Refs #4423`.
- Actual RED: `go test -count=1 -timeout 20m ./internal/connectors/defs -run '^TestRuntimeEmbedContainsExecutionJSONOnly$'` exited 1 because `defs_test.go` referenced removed `engine.Bundle` fields `Docs`, `Surface`, and `Fixtures`. The two named Atlas selectors in `./cmd/connectorgen` and `./internal/connectors/engine` both exited 0 with `no tests to run`; the explicit `go test -list '^<name>$' | grep -Fx '<name>'` test-count assertion exited 1 for each. These are retained as the honest baseline, not a pushed failing commit.
- Actual GREEN: only `defs_test.go`, `vnext_lock_test.go`, and `vnext_execution_bundle_test.go` changed. The repaired defs test asserts current execution fields; the Foundation Atlas names now select real test functions; the reference characterization reads the three committed source locks, renders each complete execution set twice in memory, byte-compares it with committed execution JSON, and proves the closed-set comparator rejects an unrendered artifact. No runtime, source lock, generated execution JSON, credential, or provider-I/O path changed.
- Green commands: the three exact named proof commands passed with `-count=1 -timeout 20m`; each `go test -list '^<name>$' | awk` assertion counted exactly one selected top-level test. `go test -count=1 -timeout 20m ./cmd/connectorgen ./internal/connectors/defs` passed. `go test -count=1 -timeout 20m ./internal/connectors/engine` is not claimed green: `TestOperationRoutesFailClosedBeforeProviderIO` fails on its route-diagnostic expectation, while the N1 diff in that package is only the named-test rename. The named engine proof remains green; this unrelated broader-suite failure is a remaining gate for later work.
- Go-task skill record: the repository-required route was loaded before the N1 test work and re-read before review/push after Captain policy inbox `006`: `go-engineering` with `references/fundamentals.md`, `references/production.md`, and `references/agentic-etl.md`; `golang-how-to`; `golang-design-patterns`; `golang-structs-interfaces`; `golang-error-handling`; `golang-security`; `golang-safety`; `golang-testing`; `golang-context`; `golang-concurrency`; and `golang-cli`. No substitution was needed. N1 adds no production interface, error, credential, context, goroutine, or CLI behavior; the review applies those skills as test-only guardrails for concrete fields, checked errors, bounded local file reads, no secret/provider path, no mutable global state, and no concurrent work.

### 2026-09-01 — inherited ledger reconciliation

- The inherited `Red: pending` / `Green: pending` record contradicts the branch handoff's claimed reference-cohort green checkpoint. It is retained as history, not accepted as evidence for this continuation.
- Baseline: clean isolated continuation branch `fm/cli-batch1-vnext-cutover-r2` at `0b214b79eeb871238ce8454cd7b896e71e2746a7`, with that SHA proven reachable from `origin/fm/cli-top100-declaration-batch-r1`.
- Manual GSD fallback: the adapter resolves every required lifecycle command and the canonical contract check passes, but its generated commands cannot execute this pre-existing named phase because `.planning/ROADMAP.md` is absent. The inline artifacts in this directory carry the required lifecycle evidence.

### 2026-09-01 — cleanup slice A: native fixture bypass

- Red target: a native connector supplied only `config.mode=fixture` must not report a successful check or emit canned records; it must continue through normal credential/config validation before provider I/O.
- Red command: `go test -timeout 20m ./internal/connectors/native/alpha-vantage -run '^TestFixtureModeNoLongerBypassesCredentialBoundary$'`.
- Green contract: the same invocation returns the connector's missing-credential error with no provider request, and the production implementation contains no fixture-mode branch.
- Follow-on residual proof: scan production native/hook/engine/connectorgen code, definitions, generated docs/skills, and website sources for a fixture, importer, certification, retention, compatibility, feature-flag, or second-executor execution/admission path. Any retained mention must be connector-local provider provenance only and is recorded by path and reason in `VERIFICATION.md`.

### 2026-09-01 — cohort migration template

- Red: before each named connector migration, `lock-render <connector> --check` fails against the newly written lock or the connector lacks a usable declaration/credential-boundary witness.
- Green: `lock-render <connector>` produces byte-identical execution JSON; all seven lanes are explicit; malformed execution JSON rejects locally; and an isolated, credential-free command reaches the ordinary missing-credential or approval boundary without provider I/O.
- Connector sequence: Bitbucket, CircleCI, Docker Hub, Jira, Notion, Sentry, Stripe, Vercel. Each receives its own red/green entry, review record, commit SHA, and normal push evidence.
