# PLAN — issue 3581 target-scope core validator

## Scope

Sub-issue: #3581 — target-scope contract and core changed-path ownership validator.
Parent: #3579 / draft parent PR #3580.
Branch: `fix/3581-target-scope-core-validator`.
Base branch: `fix/3579-connector-path-ownership-guardrails`.
Worker directory: `/Users/karthiksivadas/.treehouse/cli-83d592/5/worker-3581-core`.
Spawn decision: `spawned`.

Allowed write scope:

- `cmd/connectorgen/**`
- `internal/connectors/boundary/**`
- focused tests under those packages
- minimal validator usage docs under `docs/migration/**`
- this worker artifact directory

## GSD mode

- `scripts/gsd doctor`: passed in this worker directory.
- `scripts/gsd list`: passed; adapter exposes 69 commands.
- `scripts/gsd prompt execute-phase connector-guardrail-remediation-r1 --dry-run`: used as available GSD command path.
- `scripts/gsd prompt programming-loop init --phase connector-guardrail-remediation-r1 --dry-run`: unavailable with `scripts/gsd: unknown GSD command: programming-loop`.
- Fallback: manual GSD universal runtime loop, same as parent phase plan. TDD, verification, no-mistakes, and review gates are not weakened.

## Required skills loaded

- `gsd-core` (`.pi/skills/gsd-core/SKILL.md`)
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-documentation`
- `golang-lint`
- `no-mistakes` for scoped validation if feasible

Note: `.pi/skills/go-implementation/SKILL.md` is not present in this checkout; loaded the repository-routed Go skills above from `.agents/agentic-delivery/references/required-skills-routing.md`.

## Loaded routing rules

- Required-skills routing: Always-on Go skill routing; CLI and command behavior; Connector runtime and architecture; Reviews and hardening; Documentation for Go behavior.
- CLI parity reference: runtime help, machine-readable output docs/help, docs/website/manual parity checked or exempted in PR.
- Skill rule highlights used: golang-testing Best Practices 1/3/5; golang-error-handling Best Practices 1/2/3/7/9; golang-security Security Thinking Model 1-3; golang-safety Best Practices 1/2/4/6; golang-design-patterns Best Practices 3/5/9/10/20/21; golang-structs-interfaces principles Keep Interfaces Small / Accept Interfaces Return Structs / safe assertions; golang-documentation Writing Principles; golang-lint suppression rules 1-4.

## Implementation slices

### Slice A — red tests (done)

Added failing tests before production edits for:

1. scope contract requires exactly one connector slug;
2. validator auto-detects target connector from changed paths without labels/scope markers;
3. shared runtime/tooling paths are rejected;
4. unrelated connector definitions are rejected;
5. unrelated generated connector docs/website/manual paths are rejected;
6. target defs/fixtures/docs and narrow shared indexes/goldens are allowed;
7. connector-lane exception/config edits cannot silently weaken the gate;
8. CLI supports machine-readable output and help.

Red evidence captured in `TDD-LEDGER.md`.

### Slice B — boundary ownership package (done)

Implemented reusable changed-path validator in `internal/connectors/boundary/**`:

- machine-readable scope contract with exactly one slug;
- inference from changed paths;
- closed allowed path classes;
- structured report/findings/warnings for JSON output;
- reject validator/config path edits in connector implementation lanes.

### Slice C — connectorgen CLI (done)

Added `go run ./cmd/connectorgen ownership ...` command:

- usage/help;
- path/base changed-path collection for local/CI use;
- scope-file and inferred-scope behavior;
- `--json` machine-readable output;
- exit `0` clean, `1` findings, `2` invalid invocation/config.

### Slice D — minimal docs (done)

Updated `docs/migration/connector-boundary-guard.md` with ownership command usage and scope-file contract.

## Verification plan

Required focused verification:

```bash
gofmt -w cmd internal
go test ./internal/connectors/boundary ./cmd/connectorgen
go run ./cmd/connectorgen ownership . --help
go build ./cmd/pm
```

Additional gates if feasible after coherent green slice:

```bash
go vet ./...
make verify
```

## Commit / PR plan

- Commit coherent green slice(s) to `fix/3581-target-scope-core-validator`.
- Push branch to origin; never push `main`.
- Open sub-PR against `fix/3579-connector-path-ownership-guardrails` with Conventional Commit title.
- PR body must contain exactly the issue references:
  - `Refs #3581`
  - `Refs #3579`
  plus GSD/TDD/skills/red-green/verification/no-mistakes/review evidence.

## Human gates

Stop for secrets, new dependencies, branch protection/ruleset/auth scope blockers, destructive external actions, quality gate weakening, no-mistakes daemon error, ask-user no-mistakes finding, or push/merge to `main` risk.
