# PLAN — issue 3586 generated/unrelated connector remediation

## Scope contract

- Sub-issue: #3586 — Connector guardrail: generated and unrelated connector remediation.
- Parent issue: #3579.
- Parent PR: #3580 (draft).
- Branch: `fix/3586-generated-unrelated-connector-remediation`.
- Base branch for sub-PR: `fix/3579-connector-path-ownership-guardrails`.
- Spawn decision for this run: `spawned` (one worker issue, one branch, isolated worker directory).
- Write scope: worker artifacts under this directory, audited `docs/connectors/**` manuals/skills, audited `internal/connectors/defs/gong/cli_surface.json`, target `zendesk-support`/`google-ads` generated outputs only if evidence requires, audited generated `website/**` only if evidence requires.

## GSD mode

- `scripts/gsd doctor`: passed in this worker checkout.
- `scripts/gsd prompt programming-loop init --phase connector-guardrail-remediation-r1 --dry-run`: unavailable with `scripts/gsd: unknown GSD command: programming-loop`.
- `scripts/gsd prompt execute-phase connector-guardrail-remediation-r1 --dry-run`: available and used as the reproducible prompt path.
- Fallback: manual GSD universal runtime loop using `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`; TDD, verification, review, and safety gates are not weakened.

## Required context loaded

- `AGENTS.md`.
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`.
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`.
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`.
- `.agents/agentic-delivery/references/required-skills-routing.md`.
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`.
- `.agents/agentic-delivery/contracts/worker-handoff-template.md`.
- `.planning/{config.json,PROJECT.md,ROADMAP.md,STATE.md}`.
- `docs/plans/universal-programming-loop-prd.md` and `docs/prompts/universal-programming-loop-prompts.md`.
- `docs/migration/HANDOFF-CODEX.md`, `docs/migration/conventions.md`, `docs/architecture/connector-architecture-v2-design.md`, and `docs/migration/connector-boundary-guard.md`.
- Audit report sections for PR #3532 and PR #3535 from `/Users/karthiksivadas/karthik-agent-workspace/data/cli-first-eight-connector-guardrail-audit-r1/report.md`.
- `gh-axi issue view 3579 --full`, `gh-axi issue view 3586 --full`, `gh-axi pr view 3580 --full`, and `gh-axi issue view 3581 --full`.

## Skills loaded

- `gsd-core`.
- `caveman` for compact handoff prose only.
- `no-mistakes` for final validation driving.
- `golang-how-to`.
- `golang-cli`.
- `golang-testing`.
- `golang-error-handling`.
- `golang-security`.
- `golang-safety`.
- `golang-documentation`.

No design/React skills loaded because this plan does not edit non-generated `website/**` source.

## Slice plan

### Slice A — worker planning and red evidence

- Create this issue-scoped plan, TDD ledger, and verification checklist.
- Capture red validation that the remediation ledger/proof artifact is absent/incomplete.
- Capture red validation that the target-aware ownership guard is unavailable until #3581 (or its fixture proof is absent), so unrelated docs/defs/generated churn cannot yet be rejected by the current guard.

### Slice B — audit and disposition ledger

- Enumerate every #3532 and #3535 unrelated connector, generated/manual, target generated, and audited website path in an issue-owned remediation ledger.
- Include PR URLs, merge SHAs, path classes, disposition, and evidence.
- Preserve valid shared indexes and target generated outputs with evidence.
- Preserve unrelated connector outputs only when current validation/generator evidence shows the change is a valid safety/generated correction rather than stale output.

### Slice C — guard-proof fixture and green verification

- Add an issue-owned guard fixture/proof listing target connector scope and paths that must be rejected by the new target-aware ownership guard.
- Mark guard executable enforcement as dependent on #3581 integration; this worker does not edit `cmd/connectorgen/**` or shared boundary code.
- Run focused validation: `go test ./internal/connectors/boundary ./cmd/connectorgen` and required `rg` command.
- Run broader local gates when feasible: `gofmt -w cmd internal`, `go vet ./...`, `go build ./cmd/pm`, and `make verify` if time permits.

## Acceptance mapping

- Every listed #3532/#3535 generated/unrelated/manual/website path: covered by `REMEDIATION-LEDGER.md`.
- Remove/regenerate stale outputs: no stale/harmful output evidence found; all audited outputs are preserved with evidence.
- Valid shared indexes/target outputs: preserved and cited with connector validation/generator evidence.
- Guard rejection: `ownership-guard-fixture.json` / `OWNERSHIP-GUARD-FIXTURE.md` provide the issue-owned fixture/proof now; executable enforcement comes from #3581 target-aware guard.
- History safety: forward-only artifacts only; no revert, force-push, or history rewrite.

## Implementation result

- Added issue-owned remediation ledger and ownership guard fixture/proof under this worker directory.
- No production connector, docs, website, or generated files changed because validation showed current state should be preserved.
- Local gates passed; no non-generated website edits were made, so website test command was not required beyond `make verify` repo gates.
