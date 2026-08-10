---
phase: 600
command: gsd-code-review 600
mode: inline_manual_fallback
depth: standard
status: clean
---

# Phase 600 code review

The repo-local Pi adapter does not provide the upstream review worker and this
lane forbids role spawning, so a standard review was performed inline.

## Review focus

- The proof writer sanitizes completed live exchanges before it opens an
  evidence output file; response bodies, headers, query values, and prepared
  credentials are fingerprints only.
- Accepted evidence and generated matrices validate first in check mode, then
  perform their byte-level source cross-check. Proofless and malformed records
  cannot create a certified cell.
- All status aggregation inputs are required: capability, workflow, sync-mode,
  and every applicable pair flow. Community-build status only changes
  inspection output, never command reachability.
- Rebase integration preserved current main's GraphQL write executor and
  PostgreSQL/MySQL change-capture discovery; regenerated artifacts reflect the
  updated code without inventing fixture or live proof.
- The GraphQL direct-read branch is explicitly annotated as an executable
  `graphql_query`; the paired GitHub `local_git` operation has no executor
  annotation and remains unimplemented.

## Finding disposition

| Severity | Finding | Disposition |
|---|---|---|
| warning | Provider and record-name values were checked for presence but not safety before a proof write. | Fixed: safe-identifier validation and TestCertificationRejectsUnsafeEvidenceIdentifiers. |
| warning | The source-derived detector omitted the engine's real GraphQL direct-read branch, producing a false `implemented=false` for GitHub `graphql_query`. | Fixed: the direct-read executor annotation names `graphql_query`, and the paired regression proves `local_git` remains false. |

No remaining critical, warning, or info findings were identified. The direct
unsupported-operation implementation check remains intentionally shallow: the
captain explicitly deferred deeper call-graph reachability proof to a separate
issue. No pre-commit shepherd gate was added.

## Evidence

- Focused and full connectorgen tests passed.
- Full CLI, engine, and generated-status package tests passed.
- Lint, vet, builds, generator validation, boundary, release, docs, smoke, and
  artifact-drift gates passed; exact commands are retained in VERIFICATION.md.
