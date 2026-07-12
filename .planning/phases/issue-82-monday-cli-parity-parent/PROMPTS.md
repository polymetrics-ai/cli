# Prompts — issue #82 Monday CLI parity parent

## GSD adapter evidence

- `scripts/gsd doctor` — passed (node, repo root, official docs, command registry, Pi settings/extension/skill/prompt all ok).
- `scripts/gsd verify-pi` — passed.
- `scripts/gsd list --json` — passed; 69 commands listed.
- `scripts/gsd prompt plan-phase issue-82-monday-cli-parity --skip-research` — generated this parent planning prompt.
- `scripts/gsd prompt programming-loop init --phase issue-82-monday-cli-parity --dry-run` — unavailable (`unknown GSD command: programming-loop`). Manual GSD programming-loop fallback is recorded in `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, and `RUN-STATE.json`; TDD remains mandatory.

## Required reading completed

- `AGENTS.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md`
- `.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md`
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `.agents/agentic-delivery/workflows/coderabbit-review-loop.md`
- `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`

## Required skills loaded

`gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-graphql`, `golang-documentation`, `golang-spf13-cobra`, `golang-lint`.
