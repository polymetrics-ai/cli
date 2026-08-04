# YouTube Analytics parity-resume specification

## Goal

Rehydrate the preserved YouTube Analytics bundle delta onto current `origin/main`, then make every documented YouTube Analytics and YouTube Reporting API operation either genuinely reachable through the declared command surface or explicitly blocked with its named foundation.

## Constraints

- Target ownership is limited to `internal/connectors/defs/youtube-analytics/` plus generated connector/manual/website artifacts and this phase record.
- Rebase/recovery must start from preserved commit `fc727ac88`; current main is the runtime authority.
- Every request field needs provider-owned citation evidence using the convention that lands before final validation; do not invent a competing shared schema.
- `media.download` must use the now-wired `binary_download` executor; `reports.query` remains blocked on typed provider-query foundation issue #2985, not `provider_search`.
- Do not represent a mutation with `rest_write`; this connector's Reporting API management mutations remain unavailable until a supported runtime contract exists.
- No credentialed calls, no reverse-ETL execution, no new dependency, no shared engine/schema/conventions edits.

## Acceptance criteria

1. The provider references support the documented-operation total and every ledger row has one honest disposition.
2. Implemented operations pass the real commandrunner preflight and representative `pm youtube-analytics` invocations reach their declared executor without credentials.
3. Each request field has a provider citation or a tier-5 deferral with evidence; citations include source location, evidence type, confidence, and requiredness rationale in the repository convention.
4. Generated CLI/manual/website surfaces are synchronized only after the connector definition validates.
5. Focused contract gates and the no-mistakes delivery pipeline complete on the feature branch.
