# Verification: #477

| Gate | Status | Evidence |
|---|---|---|
| Manual GSD fallback | pass | `scripts/gsd doctor` passed; programming-loop adapter command is unavailable and recorded. |
| Plan-before-code | pass | PLAN, TDD ledger, verification checklist, summary, prompt trace, and run state created before production edits. |
| Focused RED/GREEN | pass | Initial RED: 0 pass/2 file failures (`ERR_MODULE_NOT_FOUND`). Final: 27 total, 26 pass, 0 fail, 1 sandbox skip. |
| Full Shepherd tests | pass | 164 total, 163 pass, 0 fail, 1 sandbox skip; 48877.375042 ms. |
| Strict TypeScript / Pi 0.80.6 | pass | Strict no-emit check passed over all 11 production modules using installed Pi 0.80.6 package/types. |
| Pi extension discovery | pass with documented CLI incompatibility | Required `pi --list-extensions` reports `Unknown option`; offline RPC `get_commands` exits 0 and finds `pm-shepherd` from the explicit extension. |
| Diff whitespace | pass | `git diff --check`. |
| Root Go/static/build | supplemental pass | `go vet ./...`, `go test ./...` (certify 556.599s), and `go build ./cmd/pm` exited 0 before the parent policy change. |
| Full repository verification | superseded/cancelled_by_parent_policy | `make verify` reached its full test step and was intentionally SIGTERM'd by the parent orchestrator (exit 143); parent explicitly replaced this child gate and prohibited retry. This is neither a functional failure nor a claimed pass. |
| Declared child verification equivalent | pass | Focused tests + full Shepherd suite + strict TypeScript + offline Pi RPC smoke + diff check all pass. |
| Live GitHub comments | skipped | No explicitly designated sandbox was supplied; fake transport is the required test boundary. |
| Automated review | not requested | Coordinator explicitly prohibited Claude/Copilot requests for this worker PR. |

No auth-scope change, secret access, new dependency, production mutation, reverse ETL, destructive
action, or default-branch merge occurred. Automated review remains the parent orchestrator's exact
head/range `codex_independent` route and is not represented as Claude, Copilot, or a human decision.

## Exact-head correction verification

Status: passed for corrected implementation checkpoint
`d277d4d58d57eb08c03a3dbbc4b6ea4f2677ec0a`; fresh exact-head xhigh review remains parent-owned.

| Correction gate | Status | Evidence |
|---|---|---|
| Review RED | pass | 39 total, 23 pass, 15 expected failures, 1 sandbox skip; 282.318708 ms, before correction production edits. |
| Final focused tests | pass | 42 total, 41 pass, 0 fail, 1 sandbox skip; 830.242542 ms. |
| Complete Shepherd tests | pass | 179 total, 178 pass, 0 fail, 1 sandbox skip; 52526.971417 ms. |
| Strict owned TypeScript | pass | Both owned modules and tests passed strict no-emit TypeScript 5.9.3 with the Pi 0.80.6 Node type root. |
| Strict production TypeScript | pass | All 11 production Shepherd modules passed after resolving `@earendil-works/pi-coding-agent` from the enclosing global `node_modules`. A preliminary invocation used the package itself as `baseUrl` and produced only the expected TS2307 resolver errors; the corrected pinned invocation exited 0. |
| Offline Pi registration | pass | Pi 0.80.6 RPC `get_commands` exited 0 and returned `pm-shepherd` from the explicit Shepherd extension, with offline mode and discovery disabled. |
| Diff/base/scope | pass | `git diff --check` exited 0; merge-base is immutable base `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`; changed files are only the two owned modules, matching tests/fixture, and issue-local planning artifacts. |
| Live GitHub comments | skipped | No explicitly designated sandbox was supplied. |
| Go/connectors/full repo | not run by policy | Coordinator explicitly limited this correction lane and prohibited Go, connector, and `make verify` gates. |
| Automated review | pending parent route | No Claude/Copilot request was made; the parent orchestrator owns the required fresh exact-head xhigh review. |

The correction introduced no auth-scope change, secret access, dependency, live GitHub mutation,
production/destructive action, merge, or default-branch write.

## Lock-snapshot correction verification

Status: in progress after fresh review of
`f5a4dc68a7b76f708858542a7190ca3d1f375044`.

The declared equivalent remains focused #477 tests, the complete Shepherd suite, strict TypeScript
against pinned Pi 0.80.6, offline Pi RPC registration, and full-range diff/base/owned-scope checks.
No Go, connector, `make verify`, live GitHub comment, Claude/Copilot, or merge action is authorized.
