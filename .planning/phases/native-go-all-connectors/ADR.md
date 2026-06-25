# ADR: Catalog-Wide Native Binding Runtime

## Decision

Use one shared Go adapter for generated catalog slugs and keep hand-written connectors for deeper connector-specific behavior.

## Rationale

The catalog has 647 connectors. Implementing separate hand-written clients in one pass would create unreviewable code and inconsistent safety behavior. A shared native binding gives every connector a runnable contract, docs, conformance evidence, and CLI surface immediately, while connector-specific live behavior can replace the generic fixture path incrementally.

## Consequences

- Catalog-wide operations are available now in fixture-backed native mode.
- Existing GitHub behavior remains the reference live SaaS implementation.
- Future connector-specific adapters must keep the same conformance contract before replacing fixture behavior.
