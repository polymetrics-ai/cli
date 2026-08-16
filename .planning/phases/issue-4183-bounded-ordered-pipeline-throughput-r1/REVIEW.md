# Inline code review — bounded ordered pipeline throughput R1

## Scope reviewed

- Capability declaration/load/validation: `internal/connectors/**`.
- Application admission and persisted target policy: `internal/app/**`.
- Shared Arrow producer/consumer lifetime, cancellation, credit, ordering, and publication barrier: `internal/synctransport/**`.
- CLI parser/help/plan-preview output and checked-in docs/generated artifacts.

## Findings and dispositions

1. **Accepted and fixed — initial queue capacity admitted one extra source callback record.** A `depth-1` buffered queue plus one active consumer allowed page N+2 to start while N was applying and N+1 was queued at configured depth two. The source record existed before its callback blocked, so this exceeded the requested batch bound. The controller now reserves the active consumer and synchronous callback positions and uses `depth-2` queued slots. The deterministic test blocks COPY 1, proves page 2 overlaps, and proves page 3 cannot begin until capacity advances.
2. **Accepted and fixed — an explicit pipeline flag could have been ignored on a declared but non-Arrow route.** Application dispatch now rejects every explicit depth unless the exact transformed full-overwrite Arrow route is admitted as well as both declarations. The refusal happens before source/destination I/O.
3. **Accepted and fixed — changed leaf help paths attempted execution.** `pm etl run --help` and `pm connections create --help` now render their contextual manuals and return success; regression tests cover both paths.
4. **Reviewed, intended scope boundary — `target-copy-workers` is saved policy, not a second execution lane.** The field is validated against the declarative target pool ceiling and shown in plan/preview. It does not claim or cause concurrent COPY work. A second COPY lane is explicitly deferred and must be separately implemented and measured.

## Final assessment

No unresolved correctness, secret-handling, connector-boundary, ordering, checkpoint, or unbounded-retention issue remained after the fixes above. Focused race, package, production-binary, generated-artifact, and repository checks are recorded in `VERIFICATION.md`; the unmeasured 5 GB proof remains explicitly pending by captain decision.
