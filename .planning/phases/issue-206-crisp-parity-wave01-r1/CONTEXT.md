# Crisp provider parity — Wave 1 context

## Locked task decisions

- The authoritative inventory is `/Users/karthiksivadas/karthik-agent-workspace/data/cli-crisp-provider-inventory-rebuild-r1/report.md`, read in full on 2026-08-05. It reconstructs 234 provider-owned Crisp REST operations: 105 reads and 129 writes.
- Land that report's `api_surface.json` verbatim as an inventory-only commit before any Crisp implementation.
- Implement only Wave 1: Basic-auth connection checking and the 21 documented GET operations in the Conversations and Conversation collection groups. Do not start later waves.
- Work only in `internal/connectors/defs/crisp/`, Crisp-owned fixtures/tests/CLI metadata/generated docs, and this phase artifact. A shared engine, command runner, or runtime change is a hard stop.
- No provider calls, credentials, or live writes. Use fixture-only evidence.
- The captain policy returns output in full. This wave does not add `redact_fields` or an output-redaction declaration.
- The provider ledger must retain an explicit source citation and a disposition for all 234 operations. The 21 Wave 1 rows become stream-covered; every other row remains explicitly blocked with the report's named reason and source evidence.

## GSD delivery trace and inline fallback

`scripts/gsd doctor`, `scripts/gsd sources discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review` all resolved successfully. The required discuss and TDD-plan prompts were generated with:

```text
scripts/gsd prompt discuss-phase issue-206-crisp-parity-wave01-r1 --auto
scripts/gsd prompt plan-phase issue-206-crisp-parity-wave01-r1 --tdd --skip-research --auto
```

The project has no matching numbered GSD phase and the canonical connector worker contract forbids spawning role agents for this one-connector edit. Per the repository's explicit inline fallback, the worker records the discussion, TDD ledger, execution evidence, verification, and review inline in this directory.

## Issue links

- Parent: #204 — Crisp official API parity
- Ledger prerequisite: #205 — Crisp provider-owned ledger
- Wave implementation: #206 — ETL/read streams and changefeed parity

## Required skills used

`golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-documentation`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, and `golang-lint`.
