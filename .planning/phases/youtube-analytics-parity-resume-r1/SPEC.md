# YouTube Analytics parity-resume specification

## Goal

Rehydrate the preserved YouTube Analytics bundle delta onto current `origin/main`, then make every documented YouTube Analytics and YouTube Reporting API operation either genuinely reachable through the declared command surface or explicitly blocked with its named foundation.

## Constraints

- Target ownership is limited to `internal/connectors/defs/youtube-analytics/`, generated connector/manual/website artifacts, this phase record, and the focused shared `buildInitialQuery` repair required for connector-owned typed query templates.
- Rebase/recovery must start from preserved commit `fc727ac88`; current main is the runtime authority.
- Every request field needs provider-owned citation evidence using the convention that lands before final validation; do not invent a competing shared schema.
- `media.download` must use the now-wired `binary_download` executor; `reports.query` remains planned solely for typed provider-query foundation issue #2985, not `provider_search`.
- All seven documented mutations (`jobs.create`, `jobs.delete`, `groups.insert`, `groups.update`, `groups.delete`, `groupItems.insert`, and `groupItems.delete`) remain typed, approval-gated `reverse_etl` commands backed by their specific `writes.json` actions; none uses `rest_write`.
- No credentialed calls, no reverse-ETL execution, no new dependency, and no shared schema/conventions edits; the only shared-engine change is the focused `buildInitialQuery` repair with regression coverage.

## Acceptance criteria

1. The provider references support the documented-operation total and every ledger row has one honest disposition.
2. Implemented operations pass the real commandrunner preflight and representative `pm youtube-analytics` invocations reach their declared executor without credentials.
3. Each request field has a provider citation or a tier-5 deferral with evidence; citations include source location, evidence type, confidence, and requiredness rationale in the repository convention.
4. Generated CLI/manual/website surfaces are synchronized only after the connector definition validates.
5. Focused contract gates and the no-mistakes delivery pipeline complete on the feature branch.
