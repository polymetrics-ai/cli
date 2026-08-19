# PLAN — sync-contract durability defects

## Goal

Repair the two crash-consistency defects that shipped in #3882: state reloads must retain legacy
mode normalization, and a newly created warehouse path must have every new parent entry persisted
before its checkpoint can be acknowledged.

## Required skills and GSD evidence

- Loaded: `golang-how-to`, `golang-testing`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, and `golang-safety`.
- Passed: `scripts/gsd doctor`; sources resolved for `discuss-phase`, `plan-phase`,
  `execute-phase`, `verify-work`, and `code-review`; `go run ./cmd/agentcontractgen check`.
- Manual inline fallback: this remediation is not a numbered roadmap phase. Generated prompt
  guidance is recorded in the context, plan, ledger, and verification artifacts; no GSD role was
  spawned under the single-worker contract.

## TDD sequence

1. Add the exact persisted-legacy-state sequence test and run it before production edits, saving
   the failure. Audit every `a.store.Load()` result assigned to `a.state`; record each call site
   and verdict in the verification checklist.
2. Add an observation-based directory sync test for a warehouse root whose entire parent chain is
   absent; run it before production edits, saving the failure.
3. Route the audited reload assignments through the existing normalizer and implement only the
   minimum parent-chain sync helper needed to establish a known durable boundary without
   inferring which `MkdirAll` components were new.
4. Run both red tests green, then focused app and durability tests, formatting, vet, build, and
   all separate `make verify` gates required by `AGENTS.md`.
5. Generate and execute `verify-work` and `code-review` prompts inline. Record their evidence and
   the known limitation of observing sync calls rather than causing a real crash.

## Commit checkpoints

- Planning/TDD evidence checkpoint.
- Focused red-test evidence checkpoint.
- Green implementation plus focused verification checkpoint.
- Review/verification-fix checkpoint if needed.

## Non-applicable parity

No command, help, documentation, website, connector bundle, or generated manual changes. CLI
help/manual/website parity is not applicable.
