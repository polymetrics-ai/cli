# Conventions

**Analysis Date:** 2026-07-08
**Generated via:** Upstream `/gsd:map-codebase` workflow shape, issue #122 prompt.

## Go Conventions

- Go-only CLI monolith for core runtime.
- Use `gofmt` on `cmd` and `internal` when Go code changes.
- Prefer table-driven tests for engine, conformance, certify, and CLI behavior.
- Use contextual errors and safety redaction; never include secrets in errors/logs.
- Do not add dependencies without human approval.

## CLI Conventions

- Prefer `--json` for machine-readable agent interactions.
- Commands must validate untrusted arguments and avoid broad path traversal or control characters.
- Reverse ETL must follow plan, preview, approval, execute.
- Do not expose generic shell, generic HTTP write, generic SQL write, or unrestricted raw API tools.

## Connector Authoring Conventions

Source of truth: `docs/migration/conventions.md`.

- Tier 1: declarative bundle only where possible.
- Tier 2: hooks for justified custom auth/stream/record/write/check behavior.
- Tier 3: native component split for non-HTTP or full custom connectors.
- Connector names use bare names, not `source-*` / `destination-*` slugs.
- `api_surface.json` must map each documented operation to exactly one covered surface or typed exclusion.
- Fixtures should represent real wire shapes and sanitized values.
- Parity deviations must be ledgered with typed blockers, not silent approximations.

## Multi-Technology Surface Conventions

Connector parity must classify all documented surfaces, not just REST endpoints:

- REST/HTTP JSON operations.
- GraphQL operations.
- XML/SOAP/XML feeds.
- CSV/TSV/NDJSON/report export operations.
- Binary uploads/downloads.
- File/object storage operations.
- SQL/database/CDC operations.
- Queue/event/webhook/audit-log operations.
- Admin/destructive/elevated-scope operations.

Each operation gets one primary classification; aliases and duplicate docs entries are cross-references.

## Issue-First Delivery Conventions

Source of truth: `.agents/agentic-delivery/contracts/issue-agent-contract.md`.

- Read issue and required context first.
- Keep one primary issue per PR.
- Create/update plan, TDD ledger, and verification checklist before production edits.
- Commit coherent checkpoints and push active issue branches, never `main`.
- Use Conventional Commit PR titles and `Closes #N` for completed main-targeted work.
- Run automated review routing and disposition actionable findings.

## Safety Conventions

- No secrets are needed for planning work.
- Do not run credentialed connector checks in issue #122.
- Do not run real reverse ETL execution in issue #122.
- New dependencies, auth scope changes, destructive external actions, production deploys, quality gate reductions, and `main` merge are human-gated.

---
*Convention analysis: 2026-07-08*
