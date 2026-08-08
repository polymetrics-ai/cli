# PersistIQ artifact materialization pilot - Discussion Log

> **Audit trail only.** Decisions are captured in CONTEXT.md. This autonomous
> pilot had no unresolved product-choice prompt; the latest captain scope
> change supplied the locked choices.

**Date:** 2026-08-08
**Phase:** persistiq-artifact-materialize-pilot-r1
**Areas discussed:** pilot boundary, operation mapping, static gates, certification

## Pilot boundary

| Option | Description | Selected |
|--------|-------------|----------|
| Continue bulk 420/392 sweep | Fetch and materialize the broader pool | |
| PersistIQ only | Run one timed connector end to end, then stop | ✓ |
| Choose another connector | Replace the named pilot | |

**Decision:** PersistIQ only; no second connector.

## Operation mapping

| Decision | Selected |
|----------|----------|
| Preserve the five repository buckets and report zero for absent kinds | ✓ |
| Collapse all reads and writes into two HTTP-method buckets | |
| Invent a bucket for unsupported operations | |

**Decision:** 11 ETL, 1 direct_read, 7 reverse_etl, 2 direct_write, 0
binary_download, 0 unclassified before artifact reconciliation.

## Gates and certification

| Decision | Selected |
|----------|----------|
| Static/no-network gates plus real binary reachability; withhold certification | ✓ |
| Credentialed provider exercise | |
| Treat validation as certification | |

**Decision:** Certification remains explicitly withheld.

## Captain ruling: complete operation inventory

The captain replaced the prior fail-closed materialization policy before the
rerun. Every documented artifact operation is now mapped into the bundle.
Missing runtime foundations are represented as `not_implemented` commands with
named dependencies, while existing source-surface operations absent from the
artifact remain present and are marked
`present-in-surface-absent-from-artifact`. An implemented command still must
pass the real runtime preflight; this ruling does not authorize an executable
claim the runtime cannot honor.

## GSD execution mode

The requested pilot is not represented as a phase in `ROADMAP.md`, and the
available Codex runtime cannot run the interactive Pi worker sequence. The
resolved GSD commands were checked with `scripts/gsd sources`; this phase uses
the documented inline/manual fallback. Red/Green evidence, verification, and
the fallback reason are recorded in this phase directory.

## Deferred Ideas

Bulk materialization, the 99-record feasibility plan, and all other connectors
are deferred by the captain's scope change.
