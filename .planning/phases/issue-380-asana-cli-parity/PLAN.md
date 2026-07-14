# Asana CLI Parity — Parent Execution Plan

Issue: #380  
Branch: `feat/380-asana-cli-parity`  
PR base: `main`  
Runtime: governed Go Shepherd + GSD Core/Pi in Podman

## Goal

Bring the existing declarative Asana connector to safe CLI parity without inventing a raw API
escape hatch. The current baseline is 12 streams, 13 typed actions, 250 inventoried REST endpoints,
26 covered endpoints, 224 explicit exclusions, and no `cli_surface.json`.

## Execution graph

1. **CLI surface metadata** — define the stable command vocabulary and validate all references.
2. **Help/docs renderer** — after (1), implement namespace/topic/leaf help plus docs and website parity.
3. **Stream runner** — after (1), prove stream commands use the generic runner with fixture tests.
4. **Operation ledger** — classify every official REST operation exactly once with evidence.
5. **Direct reads** — after (4), allow-list bounded, redacted fixed reads only.
6. **Attachment downloads** — after (4), add fixed-operation bounded downloads confined to a safe path.
7. **Reverse ETL policy** — after (4), map product-safe typed actions and block/gate risky classes.
8. **Parent convergence** — integrate reviewed sub-PRs, rerun full gates, and stop at human review.

Slices (2) and (3) may run concurrently only in isolated worktrees with disjoint scopes. Slices
(5), (6), and (7) may run concurrently after the ledger lands, subject to the same rule. The parent
orchestrator owns shared artifacts, branch integration, review coverage, and final verification.

## Required lifecycle per slice

1. Read the issue, parent issue, AGENTS instructions, migration handoff/conventions, GSD adapter,
   issue contract, CLI parity reference, and connector rollout checklist.
2. Load `gsd-programming-loop`, `golang-how-to`, and the Go CLI/testing/error/security/safety/design/
   interface/documentation skills.
3. Create an isolated branch/worktree from the current parent head.
4. Update slice PLAN, PROMPTS, TDD-LEDGER, VERIFICATION, RUN-STATE, and worker handoff.
5. Capture a failing behavior test or explicit docs-only exemption before production changes.
6. Implement the smallest declarative-first change; use hooks/native Go only with written proof.
7. Run targeted gates, connector validation, CLI parity checks, and the declared broader gates.
8. Push a green slice, open a stacked PR to the parent branch, obtain review coverage, disposition
   findings, and integrate only when policy permits.

## Non-goals

- No credentialed or live Asana calls.
- No generic HTTP, shell, SQL, or arbitrary API executor.
- No new dependencies without explicit human approval.
- No execution of reverse ETL, destructive/admin operations, or attachment downloads in production.
- No deletion of legacy connector code before the architecture-v2 human-gated cutover.
- No merge of the parent PR to `main` by an agent.

## Parent verification

```bash
go run ./cmd/connectorgen validate internal/connectors/defs --json
go test ./internal/connectors/conformance -run 'TestConformance/asana' -count=1
go test ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/cli -count=1
go run ./cmd/pm help asana
go run ./cmd/pm asana
go run ./cmd/pm asana --help
rg -n 'asana' docs/cli website
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
git diff --check
```

## Stop conditions

Stop at `human_gate` for dependencies, auth/security scope changes, secret access, credentialed live
checks, destructive external actions, quality-gate reductions, reverse ETL execution, or parent PR
merge. A timeout, partial test run, stale head, or missing review is not a pass.

