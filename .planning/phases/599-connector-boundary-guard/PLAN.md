# Issue 599 — Connector Definition Boundary Guard

## GSD setup

- Issue: https://github.com/polymetrics-ai/cli/issues/599
- Branch: `fm/cli-connector-boundary-guard-r1`
- GSD preflight: `scripts/gsd doctor` and `scripts/gsd list` passed on 2026-07-28.
- GSD prompt path: `scripts/gsd prompt programming-loop init --phase issue-599 --dry-run` was attempted first, but the repo-local command registry returned `unknown GSD command: programming-loop`; manual-GSD fallback is recorded here per `.agents/agentic-delivery/references/gsd-pi-adapter.md`.
- Planning prompt fallback: `scripts/gsd prompt plan-phase issue-599 --skip-research` was generated and applied inline.
- Orchestration decision, plan cycle: `read_only_spawned` — one read-only scout inspected command/test/CI conventions; implementation stays local in this isolated worktree because this is one focused shared-code slice and mutating subagent isolation is unnecessary.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-spf13-cobra` (CLI command-shape review; `connectorgen` itself is stdlib, not Cobra)
- `golang-testing`
- `golang-security`
- `golang-error-handling`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-safety`
- `golang-documentation`
- `golang-continuous-integration`
- `golang-lint`
- CLI help/docs/website parity reference

## Scope boundaries

### In scope

1. Add stdlib-only `internal/connectors/boundary` scanner.
2. Add `connectorgen boundary [repo-root] [--json] [--base <ref>]`.
3. Build a connector lexicon from `internal/connectors/defs/*/metadata.json`, connector directory names, display names, CLI-surface roots, and legacy source/destination aliases.
4. Parse shared production Go with `go/parser`/`go/ast`; classify imports, string literals, switch/case/comparison literals, provider-prefixed policy constants, certification contracts, and shared Go examples.
5. Deterministically classify allowed locations: defs, native/hook implementations, generated outputs, tests/fixtures, documentation/migration archives.
6. Add a narrow exception ledger for current-main residue with rule, connector, exact path, exact match, reason, migration issue URL, owner, expiry, and `max_matches`.
7. Fail on expired, stale, broadened, or malformed exceptions. `approved_by` text is not read as approval.
8. Add synthetic fixture tests for disallowed provider branches/helper placement attempts and allowed defs/native/hooks/generated/test/docs paths.
9. Add stable JSON/human output, exit behavior tests, baseline test, Makefile target, documentation, and standalone CI workflow.

### Out of scope

- No migration of GitHub/Twenty/WhatsApp/Gong behavior.
- No branch-protection or repository-settings mutation.
- No update to existing connector issues/PRs.
- No new dependencies.
- No credentialed connector checks or live provider calls.

## Implementation plan

### Slice 1 — Red tests and scanner API

- Create `internal/connectors/boundary` package with public `Scan(root string, opts Options) (Report, error)` API and report/findings/lexicon types.
- Add tests with temporary synthetic repos:
  - shared production Go `case "gong"` / `if connector == "github"` fails;
  - helper-file placement in shared runtime still fails;
  - defs/native/hooks/generated/test/docs locations are allowed or non-blocking;
  - JSON report shape is stable;
  - findings sort by path, line, rule, connector, match.

### Slice 2 — Scanner implementation and exceptions

- Implement deterministic path classification and whole-tree scanner.
- Implement AST scanning for imports and string literals.
- Implement provider-prefixed policy detection and certification/docs-output classifiers.
- Implement exact exception ledger loading and validation.
- Add baseline exception ledger for current residue only.
- Confirm Gong has no shared-code exception.

### Slice 3 — CLI, docs, Makefile, CI

- Wire `connectorgen boundary` in `cmd/connectorgen` with stdout/stderr and exit-code conventions.
- Add `--json` and `--base <ref>`.
- Add `make connector-boundary` and include it in `make verify`/`verify-parallel`.
- Add `.github/workflows/connector-boundary.yml` named/job `connector-boundary` with read-only permissions and no path filters.
- Add focused developer docs in `docs/migration/connector-boundary-guard.md`.

### Slice 4 — Verification and commit

- Run focused tests.
- Run required verification commands.
- Update `TDD-LEDGER.md`, `VERIFICATION.md`, `SUMMARY.md`, and `RUN-STATE.json` with actual evidence.
- Commit the focused green slice.

## CLI/help/docs parity checklist

- `connectorgen boundary --help` renders usage.
- `connectorgen boundary . --json` has machine-readable output and documented exit behavior.
- `docs/migration/connector-boundary-guard.md` documents runbook and exception ledger rules.
- `docs/cli/**` and website docs are not applicable in this guard-only slice because `connectorgen` is developer tooling, not a generated public `pm` user command; standalone workflow and Makefile target are the discovery surfaces.

## Verification checklist

- `go test ./internal/connectors/boundary ./cmd/connectorgen`
- `go run ./cmd/connectorgen boundary . --json`
- `go run ./cmd/connectorgen validate internal/connectors/defs --json`
- `make connector-boundary`
- `make verify`
- `git diff --check`

## CI repair checkpoint — 2026-07-28

- Failing check: `branch-name` rejects the active PR branch `fm/cli-connector-boundary-guard-r1`.
- Root cause: `.github/workflows/conventions.yml` enforces Conventional Commit-style branch prefixes but does not recognize the `fm/` branch family used by this issue branch.
- Planned fix: update only the branch-name workflow convention so this existing issue branch family passes without changing connector behavior, dependencies, credentials, branch protection, repository settings, or reverse ETL behavior.
- Focused verification: run the workflow shell logic locally with `HEAD_REF=fm/cli-connector-boundary-guard-r1`, plus negative and existing-positive branch examples.

## Commit checkpoint plan

One implementation commit after planning, red tests, implementation, docs, and focused verification are green. Push/PR is deferred to the no-mistakes handoff per the worker brief.
