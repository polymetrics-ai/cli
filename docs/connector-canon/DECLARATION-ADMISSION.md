# Source-declaration admission

`connectorgen declaration-admission` is the source-completeness certificate.
It reads optional
`sources/<connector>-declaration-admission.json` sidecars and performs no
provider I/O. A sidecar separates provider source operations from their one
canonical declaration so the checker can reject an omitted, duplicate, stale,
uncited, lane-changing, or destructive-metadata-free operation.

Every source operation records its provider URL, raw provider operation ID,
exact document location, method, source path, and optional source base path.
Every declaration references exactly one source ID and names the resulting
canonical endpoint, exactly one lane (`etl`, `reverse_etl`, `direct_read`,
`direct_write`, `binary_download`, or `binary_upload`), and a discoverable
`cli_surface.json` command path. The command must cite the same canonical API
surface endpoint. There is no command-projection omission state: when a
provider operation cannot run, its deterministic command still exists and is
`deferred`.

An admitted `implemented` declaration needs the appropriate existing runtime
binding and must pass the real no-I/O `commandrunner.Preflight` path. An
admitted `deferred` declaration instead names one foundation gap, and its
discoverable command carries the same `foundation_gap` metadata. Before the
runtime returns typed `system/missing_foundation`, that same preflight path
resolves exactly one non-excluded, non-policy API-surface target and verifies
the declared absence. A connector with no runnable operations is complete when
every source operation is deferred this way.

A deferred declaration is a claim about a missing *implementation component*,
not a policy label. Its admission-sidecar `foundation_gap` records a bounded
component (for example `typed_write_action`, `typed_record_schema`,
`source_importer`, or `runtime_executor`) plus evidence naming the absent
piece and exactly one method/path target. The command surface carries the same
gap ID, reason, typed component/evidence, and target for its runtime refusal.
Free-form evidence, an unknown target, an excluded row, and a policy-only
`disallowed` row cannot produce `missing_foundation`. Method, operation lane,
destructive/risk marker,
`blocked_by_default`, confirmation or approval policy, source retention/hash,
and certification state are not foundation components. They cannot hide a
cited source operation or its discoverable command.

`deferred` is endpoint-specific missing-foundation state, not a classification
for operation kinds. In particular, an implemented delete remains implemented
when its declared delete action and runtime binding exist; GitHub's `label
delete` is the admission/runtime regression control. A missing action contract
for a specific endpoint must be named as that endpoint's foundation gap rather
than treating deletes or destructive operations as generically deferred.
Every `DELETE` declaration carries explicit `kind=delete` metadata. A
non-`DELETE` target classified as `destructive_action` also carries explicit
destructive metadata, while a non-destructive target cannot be labelled delete
merely to change its admission semantics.

Admission does **not** require retained source bytes, a hash, a request body,
or a typed schema. Those belong to source-lock/import, materialization, and
runtime contracts. Reusing no-I/O runtime preflight prevents resolver drift; it
does not certify credential-bound reachability, fixtures, provider responses,
or live behavior. Keep those certificates separate:

```bash
go run ./cmd/connectorgen declaration-admission
go run ./cmd/connectorgen surface-sync --check
go test -timeout 20m -run '^TestEveryImplementedCommandPassesRuntimePreflight$' ./internal/connectors/commandrunner
go run ./cmd/connectorgen certification-matrix --check
```
