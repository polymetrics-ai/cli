# Phase Summary

Phase: `cli-schema-deriver-union-arms-r1`
Scope ruling: `bulk-batch-split-scope` (2026-08-05)

## Delivered

- Added the union-aware record-schema promotion guard in
  `internal/connectors/engine/record_schema_promotion.go`. It expands root
  `oneOf` and `anyOf` arms before measuring named fields and rejects both a
  root union (which needs separate named actions) and a closed schema admitting
  only `{}`.
- Routed the guard through real commandrunner preflight, the promotion gate the
  CLI actually uses. Connectorgen intentionally does not restate that runtime
  rule, so native/Tier-3 actions retain their own executable semantics.
- Promoted five OAS-cited Zendesk actions with fixture-backed execution:
  `update_permission_policy`, `bulk_set_agent_attribute_values_job`,
  `create_many_users`, `create_or_update_many_users`, and
  `request_user_create`.
- Enforced the OAS 100-user maxima for the two user-array actions. The selected
  `UserMergeInput` arm permits the OAS id/email/external-id identifier variants
  without inventing an external-id-only requirement.
- Regenerated connector manuals/catalog data and website connector data.
- Removed the CodeQL-reported provider-schema length arithmetic from required
  field merging while preserving its stable de-duplication and empty-array JSON
  contract; the focused engine and runtime-preflight checks are green.

## Explicitly deferred, not faked

| Source operation | Intended actions | Missing shared capability | Location | Why it remains planned |
| --- | --- | --- | --- | --- |
| `tickets_update_many` | `tickets_update_many_bulk`, `tickets_update_many_batch` | Multi-write endpoint coverage and record-array query rendering for `ids` | `engine.SurfaceCoverage` in `internal/connectors/engine/bundle.go`; write query resolution/interpolation in `internal/connectors/engine/write.go`, `read.go`, and `interpolate.go` | One endpoint can currently bind only one `covered_by.write`; CLI string-array coercion produces `[]string` while `join` accepts `[]any`, and optional record queries do not omit absent record values. |
| `update_many_users` | `update_many_users_bulk`, `update_many_users_batch` | Multi-write endpoint coverage and record-array query rendering for `ids` / `external_ids` | Same shared locations | The OAS is typed, but neither a second endpoint action nor a truthful array-selector query can be expressed today. |

These two paths deliberately remain `availability=planned`; no ambiguous union
command or known-runtime-failing second command was created.

## Regression and sweep evidence

- The initial static fixture was red before the guard (`findings=[]`); its
  replacement real-runtime preflight regression is green after it.
- `go test ./internal/connectors/commandrunner -run
  TestEveryImplementedCommandPassesRuntimePreflight -count=1` passed after the
  guard was installed. No other declarative implemented command was found with
  the hollow-schema shape.

## Adjacent sweep result, intentionally not changed

`amazon-sqs` has two empty-record, zero-flag commands: `queue delete` and
`queue purge`. They are Tier-3 native SQS actions, not declaratively derived
schemas: `native/amazon-sqs/writer.go` targets the credential's required
`queue_url` and uses its native SigV4/XML executor. They are therefore not the
Zendesk hollow-command defect and are outside this task's connector scope.

Final local-gate evidence is recorded in `VERIFICATION.md`.
