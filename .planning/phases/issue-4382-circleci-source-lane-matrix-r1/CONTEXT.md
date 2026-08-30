# CircleCI source-lane matrix R1 — context

## Delivery contract

- Issue: #4382
- Base: `fm/cli-top100-declaration-batch-r1` at `dc481bac1a8b78d60ac0b4a2b0dfd1a9068ce8db`
- Work branch: `fix/4382-circleci-track-a-r1`
- Target integration: Batch R1 parent branch, then its normal path to `main`
- Scope: CircleCI source-lock facts, a source-lane matrix, focused local validation, and planning evidence only.

## Source authority

`internal/connectors/defs/circleci/sources/circleci-operation-source-lock.json` is the authority.
It pins `https://circleci.com/api/v2/openapi.json` at SHA-256
`61c6ce11e8de509948aa3d53dcd0169913f52de20920b130b6a85dea41d66d07`,
621321 bytes, and 111 retained REST operation rows.

The matrix must retain every source row, with the original source operation ID, method,
path, and cited source location. Existing `api_surface.json`, `streams.json`, and
`writes.json` are artifacts that may backlink to a matrix cell; they are not source-ID
authorities and are not changed by this phase.

## Decisions fixed by issue evidence

- Materialize exactly seven cells per source operation: `direct_read`, `direct_write`,
  `binary_download`, `binary_upload`, `etl`, `reverse_etl`, and `sync_transport`.
- Retain every source row even when present importer/certification artifacts do not
  represent it.
- There are 61 retained GET rows and 50 retained mutation rows. Each mutation gets an
  independent `direct_write` and `reverse_etl` map-only disposition; this neither
  asserts executability nor creates a provider transport.
- Only nine retained GET rows contain a source-documented cursor continuation suitable
  for a map-only ETL/sync candidate: `getProjectWorkflowMetrics`,
  `getProjectWorkflowRuns`, `getProjectWorkflowJobMetrics`, `listPipelines`,
  `listWorkflowsByPipelineId`, `listPipelinesForProject`, `listMyPipelines`,
  `listSchedulesForProject`, and `listWorkflowJobs`.
- A response token without an operation-level request continuation remains a source
  mapping gap, not an inferred ETL capability.
- Preserve source-declared webhook `signing-secret` field facts without storing a
  secret or generating shell/credential surfaces. Preserve `project-slug` path facts
  without converting them into a runtime identity contract.

## Foundation Atlas result

The Foundation Atlas was consulted before naming any gap. `source.projection-admission.v1`,
`transport.sync-contract.v1`, and `warehouse.reverse-etl.v1` exist. This delivery invokes
none of their runtime paths. Because no runtime behavior is claimed, all prospective
runtime cells remain `mapped_unproven`; no `missing_foundation` reason and no new
foundation request is warranted. Any absence is recorded as a source/mapping evidence
gap only.

## GSD workflow record

`scripts/gsd doctor` passed and prompts for `discuss-phase`, `plan-phase --tdd`,
`execute-phase`, `verify-work`, and `code-review` were read. The canonical GSD workflow
normally uses specialized runtime subagents, which are unavailable to this Pi-local
session and not authorized by the task. The required inline/manual fallback is used:
this context, PLAN, TDD ledger, verification report, and review record are maintained
in the phase directory.
