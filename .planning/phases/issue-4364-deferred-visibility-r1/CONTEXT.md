# Context — issue 4364 deferred visibility bridge

## Decision record

- **Issue:** Refs #4364 — Batch R1 typed deferred visibility bridge.
- **Base:** `origin/fm/cli-top100-declaration-batch-r1` at
  `687eb1ded6b42cc456f8cc3c1e97f0a84fd042a8`.
- **Delivery:** candidate-only branch `codex/4364-deferred-visibility-r1`;
  no parent integration or main merge.
- **Authority:** captain-approved mapping/certification-admission work only.
  Runtime foundations, executors, transport, credential handling, source
  imports, materialization, and provider I/O are out of scope.

## Source and ownership facts

- CodeGraph is absent at the repository root, so targeted repository
  inspection is the applicable discovery method.
- The frozen Batch R1 cohort is ten connector-local source locks and 4,341
  primary source identities. The source-lane matrices are the authoritative
  lane evidence, not a new competing provider-fact source.
- The matrices use two established wire forms: `source_operations` with a
  `lanes` object, and `operations` with a `cells` array. Both must normalize
  to the same seven lanes without connector or operation allow-lists.
- GitLab additionally records two source-cited rendered-reference binary rows
  through a connector-owned supplemental source lock. They are not part of the
  4,341 primary denominator, but must remain visible and source-bound rather
  than silently dropped.
- `mapped_unproven` is a source-backed mapping status. It is not a missing
  runtime implementation claim. Its generic authoring prerequisite is a
  field-complete source-bound declaration under the existing
  `source.projection-admission.v1` Atlas capability.
- `missing_foundation` is a concrete non-executing gap. Existing connector
  ledgers or matrix `foundation_atlas` records name its foundation/capability;
  this issue will surface, validate, and report those records without creating
  or implementing them.

## Design choice

Add a check-only `connectorgen deferred-visibility <cohort> --check [--json]`
authoring command. It reads only local source locks, source-lane matrices,
connector-local missing-foundation ledgers when present, and the Foundation
Atlas catalog. Its deterministic report contains only source evidence and
deferred discovery facts:

- connector, exact source-operation ID, source-lock path, source citation,
  provider method/path/operation ID, and the source-fact fragment;
- one exact seven-lane matrix cell and its underlying disposition;
- a stable bridge reason (`deferred_visibility.mapped_unproven.v1` or
  `deferred_visibility.missing_foundation.v1`), plus the original cited reason;
- the required source-projection authoring capability for mapped-unproven
  cells, or a named connector foundation/Atlas reference for concrete gaps;
- explicit `mapping_only: true`, `runtime_claim: "none"`, and zero executable
  declarations.

It must not invoke source import/materialization/projection, runtime bundle
loading, credentials, transport, a provider, or any executor. It neither
writes an artifact nor changes a lock/matrix.

## Manual GSD fallback

`scripts/gsd doctor` and the official `discuss-phase`, `plan-phase --tdd`,
`execute-phase`, `verify-work`, and `code-review` prompts were resolved.
The repository adapter is Pi-project-local, while this lane has no compatible
interactive Pi/subagent runtime. The documented inline/manual fallback is used:
this context, plan, TDD ledger, verification record, and inline review carry
the same evidence rather than creating a parallel workflow.
