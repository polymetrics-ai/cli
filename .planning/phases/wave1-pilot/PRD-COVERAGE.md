# PRD Coverage

Phase: wave1-pilot

- Passed: yes
- Frontend/UI work detected: yes

Use `- <artifact key>: not-applicable - <reason>` only when an artifact truly does not apply.

| Artifact | Status | References | Notes |
| --- | --- | --- | --- |
| PRD | present | docs/plans/universal-programming-loop-prd.md |  |
| SPEC | present | .planning/phases/wave1-pilot/SPEC.md |  |
| PLAN | present | .planning/phases/wave1-pilot/EVAL-PLAN.md<br>.planning/phases/wave1-pilot/PLAN.md<br>.planning/phases/wave1-pilot/TEST-PLAN.md |  |
| Test plan | present | .planning/phases/wave1-pilot/TEST-PLAN.md |  |
| Design direction | not_applicable |  | Backend-only phase (declarative connector bundles, Go hooks, planning docs); the "frontend detected" flag comes from the out-of-scope docs website under website/, unchanged this phase. |
| Architecture notes | present | docs/architecture/repo-profile.json |  |
| ADR | present | docs/adr/0001-connectors-as-data.md |  |
| API contract | present | .planning/phases/wave1-pilot/API-CONTRACT.md |  |
| Data model | present | .planning/phases/wave1-pilot/DATA-MODEL.md |  |
| Threat model | present | .planning/phases/wave1-pilot/THREAT-MODEL.md |  |
| Observability plan | present | .planning/phases/wave1-pilot/OBSERVABILITY.md |  |
| Rollback/runbook | present | .planning/phases/wave1-pilot/RUNBOOK.md |  |
| Eval plan | present | .planning/phases/wave1-pilot/EVAL-PLAN.md |  |
| Release notes | not_applicable |  | Pre-release internal milestone; pilots add unregistered bundles only (registry flip is wave6), so there is no user-facing behavior change to note; release notes deferred to wave6 convergence. |
| Postmortem template | not_applicable |  | Global template exists at docs/plans/POSTMORTEM-TEMPLATE.md (added in wave0); no phase-local copy needed. |
