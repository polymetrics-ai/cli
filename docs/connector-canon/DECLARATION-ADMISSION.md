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
surface endpoint.

An admitted `implemented` declaration needs the appropriate existing runtime
binding. An admitted `deferred` declaration instead names one foundation gap,
and its discoverable command carries the same `foundation_gap` metadata. The
runtime then refuses that command with a typed missing-foundation error before
provider I/O. A connector with no runnable operations is complete when every
source operation is deferred this way.

Admission does **not** require retained source bytes, a hash, a request body,
or a typed schema. Those belong to source-lock/import, materialization, and
runtime contracts. It also does not certify runtime reachability, credentials,
fixtures, or live provider behavior. Keep those certificates separate:

```bash
go run ./cmd/connectorgen declaration-admission
go run ./cmd/connectorgen surface-sync --check
go test -timeout 20m -run '^TestEveryImplementedCommandPassesRuntimePreflight$' ./internal/connectors/commandrunner
go run ./cmd/connectorgen certification-matrix --check
```
