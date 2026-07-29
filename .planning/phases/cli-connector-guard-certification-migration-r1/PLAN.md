# Connector Guard Issue C — Definition-Owned Certification Contracts

## GSD setup

- Source task: Issue C from `/Users/karthiksivadas/karthik-agent-workspace/data/cli-connector-guard-rollout-migration-r1/report.md`.
- Branch: `fm/cli-connector-guard-certification-migration-r1`.
- GSD preflight: `scripts/gsd doctor` and `scripts/gsd list` passed in this isolated worktree.
- GSD prompt path: `scripts/gsd prompt programming-loop init --phase cli-connector-guard-certification-migration-r1 --dry-run` was attempted first, but the repo-local command registry returned `unknown GSD command: programming-loop`; manual-GSD fallback is recorded here per `.agents/agentic-delivery/references/gsd-pi-adapter.md`.
- Planning prompt fallback: `scripts/gsd prompt plan-phase cli-connector-guard-certification-migration-r1 --skip-research` was generated at `/tmp/gsd-plan-cert-r1.md` and applied inline.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-testing`
- `golang-safety`
- `golang-code-style`
- `golang-error-handling`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-security`
- CLI/help/docs/website parity reference, marked not applicable to public `pm` help because this slice changes certify internals and connector bundle metadata only.

## Scope boundaries

### In scope

1. Add optional typed `certification.json` metadata to connector bundles.
2. Add GitHub certification metadata under `internal/connectors/defs/github/` for:
   - default stream;
   - live-unavailable classifiers;
   - source credential defaults;
   - direct-read candidates;
   - binary candidates;
   - write pairings;
   - write record schemas loaded from definition-owned writes.
3. Make GitHub certification behavior load contracts from `defs/github`, with no second shared GitHub source of truth.
4. Preserve ledger and sweeper safety behavior, including safe/no-op unknown connector behavior.
5. Remove obsolete Connector Boundary Guard `provider_certify_contract` exceptions after the shared certify GitHub literals are gone, without weakening scanner rules.
6. Add focused validation/regression tests for parsing, malformed metadata, GitHub behavior, and unknown connectors.

### Out of scope

- No live or credentialed connector checks.
- No PM Broker changes.
- No Homebrew/release workflow changes.
- No generated docs/native-hook/public-rollout issues D onward.
- No raw generic write tools.
- No reverse ETL execution; existing plan → preview → approval → execute gates must remain unchanged.
- No new dependencies.

## Implementation plan

### Slice 1 — Definition parsing red/green

- Add `engine.CertificationSpec` and optional `Bundle.Certification` loaded from `certification.json`.
- Include `*/certification.json` in the embedded defs FS.
- Add meta-schema validation and strict decode behavior.
- Red tests:
  - valid synthetic bundle parses certification metadata;
  - malformed/unknown-field certification metadata fails validation/load;
  - missing certification metadata is safe and optional.

### Slice 2 — Certification profile loader

- Add provider-neutral certify helpers that read `DefinitionProvider.Definition()` / engine bundle metadata for certification contracts.
- Load source credential defaults, default stream, live-unavailable classifiers, direct-read candidates, binary candidates, pairings, and write schemas through that helper.
- Keep unknown connector behavior empty/no-op and do not invent cleanup/write behavior.

### Slice 3 — GitHub metadata migration

- Add `internal/connectors/defs/github/certification.json` with the current GitHub contracts.
- Remove hard-coded GitHub maps/switches/helpers from shared certify code.
- Load write record schemas directly from `defs/github/writes.json` actions instead of `builtinWriteSchemas`.
- Preserve ledger/sweeper cleanup selection and failure behavior.

### Slice 4 — Boundary exceptions and verification

- Delete only the now-obsolete GitHub `provider_certify_contract` exception rows.
- Do not relax boundary scanner rules.
- Run focused tests, boundary guard, defs validation, and `make verify` before commit.
- Update `TDD-LEDGER.md`, `VERIFICATION.md`, `SUMMARY.md`, and `RUN-STATE.json` with actual evidence.

## CLI/help/docs parity checklist

- Public `pm` command help/docs/website updates: not applicable; no CLI command, flag, help text, or public docs surface changes in this focused certify metadata migration.
- Developer validation is through Go tests, `connectorgen validate`, connector boundary guard, and `make verify`.

## Verification checklist

- `go test ./internal/connectors/engine ./internal/connectors/certify ./cmd/connectorgen ./internal/connectors/boundary`
- `go run ./cmd/connectorgen validate internal/connectors/defs/github --json`
- `go run ./cmd/connectorgen validate internal/connectors/defs --json`
- `go run ./cmd/connectorgen boundary . --json`
- `make connector-boundary`
- `make verify`
- `git diff --check`

## Commit checkpoint plan

One focused implementation commit after planning, tests, implementation, verification, and status artifacts are green. Push/PR is deferred to the firstmate/no-mistakes handoff.
