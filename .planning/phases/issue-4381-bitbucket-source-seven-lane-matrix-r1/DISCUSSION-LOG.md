# Discussion log — Bitbucket Track A

No unresolved product choice remains within Track A. The issue and task establish the seven lanes, source-lock denominator, crosswalk-only boundary treatment, no-runtime restriction, and parent-branch delivery target. The only discovery is a typed deferred sync gap for Bitbucket provider webhook delivery; it is retained as a source-cited `missing_foundation` mapping result and is not implemented in this task.

## Semantic-source repair — 2026-08-31

The frozen target exposed two connector-local rigidity defects: method/operation-ID selection for direct/read-versus-mutation classification, and a `paginated_` response-schema naming dependency for ETL. The retained source lock already contains sufficient provider facts to repair both: operation summaries identify the existing bounded-read and mutation vocabulary, and `source_contract.components.schemas` identifies the response `next`/`values` continuation contract. No product or foundation decision is required. The four source-cited webhook subscription rows retain their existing `sync_transport` gap; response pagination does not infer sync transport.
