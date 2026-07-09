# PRD Coverage

Phase: github-graphql-engine

- Passed: yes
- Frontend/UI work detected: no

| Artifact | Status | References | Notes |
| --- | --- | --- | --- |
| PRD | present | docs/plans/universal-programming-loop-prd.md | Connector architecture v2 requires full capability and GraphQL/XML scalable primitives. |
| SPEC | not-applicable | .planning/phases/github-graphql-engine/PLAN.md | Phase is a narrow engine slice; PLAN carries the executable spec. |
| PLAN | present | .planning/phases/github-graphql-engine/PLAN.md | Issue #39 scoped implementation plan. |
| Test plan | present | .planning/phases/github-graphql-engine/TEST-PLAN.md | Red/green test targets recorded. |
| Design direction | not-applicable | N/A | Backend engine change; no UI surface in this phase. |
| Architecture notes | present | docs/architecture/repo-profile.json | Go CLI monolith, connector engine package. |
| ADR | present | docs/adr/0001-connectors-as-data.md | Connectors as JSON bundle data. |
| API contract | present | internal/connectors/engine/schema/streams.schema.json; internal/connectors/engine/schema/writes.schema.json | Adds `graphql` request metadata and `body_type: graphql`. |
| Data model | present | internal/connectors/engine/bundle.go | Adds `GraphQLRequestSpec` to streams and writes. |
| Threat model | present | .planning/phases/github-graphql-engine/PLAN.md | Fixed bundle documents only; no raw GraphQL escape hatch. |
| Observability plan | not-applicable | N/A | No new runtime observability surface; existing engine errors carry connector/stream/action context. |
| Rollback/runbook | present | .planning/phases/github-graphql-engine/VERIFICATION.md | Revert #39 commit or remove GraphQL metadata fields before use. |
| Eval plan | not-applicable | N/A | Deterministic Go engine tests validate behavior. |
| Release notes | present | .planning/phases/github-graphql-engine/SUMMARY.md | Summarizes engine capability. |
| Postmortem template | not-applicable | N/A | No incident/production deploy in this phase. |
