# Execution plan

1. Add a failing validator/preflight regression using a top-level union whose arms
   contain typed fields but whose derived root currently becomes `{}`. Record the
   failing command and output in `TDD-LEDGER.md`.
2. Trace and repair the shared record-schema parser/deriver so union arms are
   expanded before measurement. Gate implementation through the real command
   preflight; do not duplicate the runtime rule in `connectorgen`.
3. Derive the five Zendesk request shapes reachable with the current declarative
   surface, selecting a named closed arm for nested user unions. Preserve the
   documented 100-item user bounds. Keep the two bulk-versus-batch roots planned:
   their intended action names and exact missing shared capabilities are recorded
   in the endpoint ledger rather than faked as ambiguous commands.
4. Add fixture-backed command, source-citation, schema-parity, and bounds tests;
   update the reverse-ETL ledger, regenerate docs/website data, run the repository
   surface synchronizer, then sweep every implemented bundle for a hollow schema.
5. Run focused Go tests and each non-full-suite verification gate, update evidence,
   commit the completed work on `fm/cli-schema-deriver-union-arms-r1`.
6. Respond to the PR CodeQL finding without broadening behavior: remove the
   user-derived length addition used solely for map/slice preallocation in
   `mergeRecordSchemaRequired`, preserve order-preserving deduplication, and
   rerun the focused engine test plus the PR security check.

## Delivery controls

- Manual GSD fallback is in use because `scripts/gsd prompt programming-loop`
  returns `unknown GSD command: programming-loop` even though `scripts/gsd doctor`
  passes; the fallback artifacts live in this directory.
- `local_critical_path`: one worker owns the coupled engine, validator, and Zendesk
  bundle changes in this disposable worktree. Splitting mutations would create
  collision risk without reducing the sequential OAS derivation work.
- Red test must precede production code. The scope ruling requires the completed
  commit to be pushed and visible in a PR before handoff; no no-mistakes pipeline
  run is active, and any later run is driven only through its gates.
