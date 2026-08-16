# Research — PostgreSQL history-mode truthfulness repair

## Production path

`app.Open` loads PostgreSQL's bundle, composes definition-owned factories, and
registers their existing native polling source and managed-target destination.
`synctransport.Registry.Preflight` validates the outer source/destination mode
lists and the destination apply strategy before looking up either executor.
Therefore no registry code change is needed: adding the existing mode to both
outer lists and its existing strategy mapping makes preflight resolve the
already registered executors.

## Existing implementation

`postgresPollingTransportSource` declares all six modes and a target that
supports `PollingApplyStrategyDedupeHistory`. The managed target's history
writer already implements the close-and-insert transaction, order fence,
validity layout, durable receipt, and tombstone close. The existing
`TestPostgresManagedTargetIncrementalDedupeHistoryLive` proves those inner
semantics but constructs the adapter directly; it does not meet this task's
binary-entry-point proof bar.

## Required new proof

`internal/cli/postgres_transport_binary_integration_test.go` already builds a
fresh `pm`, creates real source/target PostgreSQL databases, adds credentials
via the CLI, saves a connection, produces a real approved run, and queries the
target separately. Extend that proof for a history connection rather than
introducing a fake target or a second harness.

No new package or dependency is required.
