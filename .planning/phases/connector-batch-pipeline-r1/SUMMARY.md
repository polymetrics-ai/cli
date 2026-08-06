# Summary — connector batch pipeline r1

Implemented the batch control plane in `cmd/connectorgen`, authored the first
evidence-quality-first batch, and recorded its individual and full gates.

- Added `connectorgen batch plan` for deterministic ledger-to-manifest intake.
- Added `connectorgen batch gate` for failure-isolated per-bundle validation,
  surface-sync drift detection, real runtime preflight, and operation-split
  reporting.
- Generated a five-candidate first manifest: DocuSeal, DefiLlama, Docker Hub,
  Flexmail, and Alpaca Broker API (190 surveyed operations).
- Materialized the cited public artifacts into v2 provenance-backed operation
  ledgers, generated `operations.json` and reachable `cli_surface.json`, and
  retained the reviewed synthetic fixtures and regenerated documentation
  surfaces.
- Individually validated and runtime-preflight-gated every candidate, then ran
  the clean batch gate: five included, zero dropped, 203 declared operations,
  split 39 executable / 27 provider-blocked / 137 excluded.
- Rebased on current `origin/main` at `de5ebb55b` and re-materialized/re-gated
  after #3870 and #3871. The executable total truthfully remains 39: direct
  reads still have only redacting runner policies, and the cited DocuSeal
  document endpoints are JSON contracts rather than `rest.multipart` inputs.
- Documented the drop procedure, operator commands, selection evidence, and a
  conservative 30-connector operating estimate in
  `docs/migration/connector-batch-pipeline.md`.
- Added the Top-50 size-tier batch map from the provider-surface audit. It
  preserves the 46 measured-provider total of 13,761 operations, treats the
  five invalid inventory foundations and four uncountable providers as explicit
  gates, and proposes 27 appropriately sized authoring batches rather than
  pretending every provider fits a 30-connector batch.

Final scoped repository validation passed on `origin/main` `de5ebb55b`. No new
Top-50 connector is authored until its planned batch obtains an eligible
provider-artifact manifest; no shared schema, engine, or command-runner source
was edited.
