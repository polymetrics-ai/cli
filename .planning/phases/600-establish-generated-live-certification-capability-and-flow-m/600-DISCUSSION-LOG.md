# Phase 600: Generated live certification capability and flow matrices - Discussion Log

> **Audit trail only.** Decisions are captured in `600-CONTEXT.md`.

**Date:** 2026-08-10  
**Phase:** 600-establish-generated-live-certification-capability-and-flow-m  
**Mode:** `scripts/gsd prompt discuss-phase 600 --auto` executed inline

---

## Runtime-derived function inventory

| Option | Description | Selected |
|---|---|---|
| Copied static list | Maintain a separate manually updated taxonomy | |
| Source/runtime discovery | Derive kinds and implementation paths from executable code | ✓ |

**Auto-selected choice:** Source/runtime discovery. The user explicitly
requires new kinds not to be silently omitted.

## Evidence and certification rule

| Option | Description | Selected |
|---|---|---|
| Filename/status inference | Treat existing certification-named files as proof | |
| Strict accepted evidence | Require a typed live-evidence record and keep missing proof false | ✓ |

**Auto-selected choice:** Strict accepted evidence. Existing bundle contracts
and Employment Hero fixtures are not accepted live certification artifacts.

## Pair-scoped end-to-end flows

| Option | Description | Selected |
|---|---|---|
| Endpoint-only status | Infer flow certification from individual connectors | |
| Source/destination pair cells | Require pair-level applicability and independently-read-back proof | ✓ |

**Auto-selected choice:** Source/destination pair cells. The captain's
extension explicitly requires non-compositional flow evidence.

## Captain clarification: warehouse mediation and delivery guarantees

| Option | Description | Selected |
|---|---|---|
| Direct endpoint hop | Treat API-to-API as source directly to destination | |
| Warehouse-mediated round trip | Require source → local Parquet warehouse → destination, including same-connector pairs | ✓ |

**Captain-selected choice:** Every flow uses the local warehouse mediator and
is live only after a real `pm` round trip with independent warehouse and
destination readbacks. A working flow separately reports whether it is
resumable, receipt-backed, checkpointed, replay-identifiable, and backed by a
provider idempotency key; false guarantees require named limitations.

## Delivery shape

| Option | Description | Selected |
|---|---|---|
| Hand-maintained report | Edit certification status directly | |
| Generated artifacts and check | Commit deterministic output and fail on drift | ✓ |

**Auto-selected choice:** Generated artifacts and check, with capability and
flow layers committed independently.

## Captain corrections: whole workflow, sync modes, and narrow PR boundary

| Decision | Selected handling |
|---|---|
| Connector status | Require all applicable capability, ETL, reverse ETL, flow-authoring, schedule, sync-mode, and pair-flow cells; leave the connector reachable when false. |
| Sync scoreboard | Derive all modes from synccontract and score each against the four warehouse-facing primitives, with named non-applicability. |
| Reachability analysis | Do not widen this PR from the direct unsupported-operation check to a deep call-graph proof. |
| Delivery shepherd | Do not add a pre-commit shepherd gate under .agents in this PR. |

**Captain-selected boundary:** The generated artifact, strict proof writer,
matrix validation, visible quality signal, and honest red baseline ship here.
Deep reachability analysis and a pre-commit gate are separate issues.

## Inline GSD fallback

The project adapter and generated prompt are available, but the canonical
single-worker contract and this worker lane forbid spawning GSD roles. The
discuss, plan, execute, verification, and review lifecycle is therefore run
inline with the same required artifacts and TDD evidence.

## Deferred Ideas

None.
