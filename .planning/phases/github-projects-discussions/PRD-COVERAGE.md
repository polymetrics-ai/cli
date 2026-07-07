# PRD Coverage

Phase: github-projects-discussions

- Passed: yes
- Frontend/UI work detected: no

Use `- <artifact key>: not-applicable - <reason>` only when an artifact truly does not apply.

| Artifact | Status | References | Notes |
| --- | --- | --- | --- |
| PRD | present | docs/plans/universal-programming-loop-prd.md |  |
| SPEC | present | .planning/phases/github-projects-discussions/SPEC.md |  |
| PLAN | present | .planning/phases/github-projects-discussions/EVAL-PLAN.md<br>.planning/phases/github-projects-discussions/PLAN.md<br>.planning/phases/github-projects-discussions/TEST-PLAN.md |  |
| Test plan | present | .planning/phases/github-projects-discussions/TEST-PLAN.md |  |
| Design direction | not-applicable | docs/design/generated/github-projects-discussions-direction-a-critique.md | CLI-only connector parity slice with no UI design changes. Generated placeholder critique marks all sections TBD. |
| Architecture notes | present | docs/architecture/repo-profile.json |  |
| ADR | present | docs/adr/0001-connectors-as-data.md |  |
| API contract | present | .planning/phases/github-projects-discussions/API-CONTRACT.md |  |
| Data model | present | .planning/phases/github-projects-discussions/API-CONTRACT.md, internal/connectors/defs/github/streams.json, internal/connectors/defs/github/schemas/{projects,project_items,discussions,discussion}.json, fixtures/streams/{projects,project_items,discussions,discussion}/*.json | GraphQL variable resolution model in API contract; stream schemas define record shapes; fixtures provide typed read_query parameter examples. |
| Threat model | present | .planning/phases/github-projects-discussions/THREAT-MODEL.md |  |
| Observability plan | present | .planning/phases/github-projects-discussions/OBSERVABILITY.md |  |
| Rollback/runbook | present | .planning/phases/github-projects-discussions/ROLLBACK-RUNBOOK.md |  |
| Eval plan | present | .planning/phases/github-projects-discussions/EVAL-PLAN.md |  |
| Release notes | present | .planning/phases/github-projects-discussions/RELEASE-NOTES.md |  |
| Postmortem template | present | .planning/phases/github-projects-discussions/POSTMORTEM-TEMPLATE.md |  |
