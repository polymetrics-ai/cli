# TDD Ledger: Batch R1 vNext source-lock cutover

## Planned evidence

| Slice | Red characterization | Green contract | Refactor/verification |
| --- | --- | --- | --- |
| Runtime dependency | Reference connectors require embedded authoring/certification/admission material. | GitHub, GitLab, and Asana load and reach credential/approval preflight from execution JSON alone. | Audit embedded/runtime reads and run focused plus fleet tests. |
| Connector-local invalidity | Global ledgers or one bundle error can suppress the fleet. | Malformed required execution JSON rejects that connector without hiding healthy connectors. | Assert stable typed diagnostics and deterministic discovery. |
| Canonical rendering | Existing source-lock paths do not own a single canonical all-lane projection. | One vNext model renders byte-stable existing execution JSON through shared schema refs. | Check every rendered file and reject stale output. |
| Lane semantics | Retention/certification state can hide documented commands and source operations cannot express all lanes canonically. | Direct, binary, ETL, reverse ETL, sync, and explicit-empty lanes are surfaced without provider switches. | Run the same all-lane contract for every Batch R1 connector. |

## Actual evidence

### 2026-09-01 — planning checkpoint

- Red: pending characterization test commit.
- Green: pending implementation.
- Baseline: clean isolated worktree at `9e96c14946c485bec026fdce211ca4098ef31b16`, matching `origin/fm/cli-top100-declaration-batch-r1` before edits.
- Manual GSD fallback: lifecycle prompts are executed inline in this single-owner worktree; required sources and architecture inputs were read before this plan.
