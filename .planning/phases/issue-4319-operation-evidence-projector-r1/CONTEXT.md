# Issue #4319 — operation evidence projector context

## Task Delivery Header

- Issue: Refs #4319 — feat(connectorgen): project operation evidence and enforce fixed-100 gate
- Base branch: main
- Merges into: main
- Delivery: Pull request open against `main` with the required local verification green and API-reported base confirmed.
- Working branch: fm/cli-operation-evidence-projector-r1
- Task: Build a provider-neutral, deterministic operation evidence projector that joins official source operations to source traces, declaration/runtime/CLI/website/proof surfaces and reports all six capability classifications. Add an executable fixed-100 gate that rejects any regression in the fixed reference cohort. Missing evidence is explicit; genuine absence requires provider-owned evidence.
- Verification: Behavioral Go tests exercise connector definitions through the command path; the projector is checked for deterministic bytes and explicit gaps; `go test -timeout 20m ./cmd/connectorgen`; all repository verification entry points including `make verify`; and opened PR base is read through the GitHub API.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| An official operation joins every required evidence surface | live | A checked-in connector definition and its source lock project to a row whose trace, canonical mapping, reachability, CLI, website, proof, and classification fields are observable in the emitted artifact. |
| Each missing surface is a specific gap | live | A real bundle fixture is made incomplete one surface at a time and the command emits the corresponding named gap rather than omitting its source operation. |
| Provider-established absence is recorded, not asserted | live | A fixture source lock with the provider's `not_enumerable` evidence projects a documented absence with its trace. |
| Rollups and generated artifacts are deterministic | live | Two command invocations against the same real definition tree emit byte-identical artifacts and one deduplicated missing-foundation rollup. |
| The fixed-100 gate is executable | live | A reference cohort passes, then each selected required row is mutated in turn and the command rejects the exact regression. |

## Discussion record

- The GSD adapter is healthy and its command sources resolve. Compatible isolated GSD worker runtime is unavailable in this session and repository policy forbids role spawning for this task, so `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` are executed inline with their artifacts recorded here.
- The task-provided reports `data/cli-batch-dispatch-foundation-design-r1/report.md` and `data/cli-multidoc-source-lock-foundation-design-r1/report.md` are absent from `6410fe59c`, all fetched refs, and the concurrent source-lock branch's committed tree. The available Batch 8–10 report at `origin/fm/cli-map-batch8910-r1:.planning/phases/issue-4292-parity-batches-8-10-r1/SOURCE-SURFACE-REPORT.md` confirms the external provider-operation inventories. This lane will consume the current source-lock interface only and will not edit `cmd/connectorgen/sourceimport.go` or its schema.
- Scope is the separate shared-foundation issue #4319. No individual connector definitions, GitHub rate-limit declarations, or source-lock parser/schema are in scope.
