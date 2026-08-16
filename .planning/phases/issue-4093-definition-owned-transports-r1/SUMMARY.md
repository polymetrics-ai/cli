# Summary — #4093 definition-owned production transports

`sync_transport.json` is now an optional version-1 bundle declaration that is
schema-checked, strictly decoded, validated with the existing closed transport
types, and deep-cloned into each connector definition projection.

Production composition accepts exact executor-reference factories, validates
every discovered declaration before registering anything, and admits only
factory-owned conformance evidence. GitHub now declares its issue-label source
and destination roles in `defs/github`; PostgreSQL declares its bounded native
snapshot source in `defs/postgres`. Connector-local factory providers keep the
App composition root free of native connector imports. The old Go-authored
descriptors and the GitHub registry wrapper are removed.

The bundle embed inventory includes `sync_transport.json`, so these definitions
are present in `defs.FS` and therefore in the production `pm` binary, rather
than only on the source filesystem.

`pm connectors inspect github --json` now visibly reports both roles as
`declared`; its help, CLI manual, generated connector documentation, and the
website agent guide and generated website data explain that inspection remains
metadata-only.

The acceptance matrix in VERIFICATION.md records stateful proof for valid
registration, zero-side-effect refusals, evidence denial before source I/O,
and the mandated Docker/Colima PostgreSQL integration run.

## R2 continuation

The residual scalable-registration proof now loads a throwaway second connector
from its own `sync_transport.json`, supplies only definition-selected named
test hooks, and drives it through the real App composition, transport registry,
and generic orchestrator. One record is read, staged, planned, applied,
read-back, and committed without a connector-name branch or a production App,
orchestrator, or dispatch edit for that connector.

Transient connection-owned worksets retain their durable manifest and Parquet
through ordinary Open, so recovery and certification can inspect the real
execution evidence. The generic optional exact-receipt retirement contract is
still available to stages that choose eager disposal. Connection-owned cleanup
instead occurs boundedly before the next generic source read: it matches the
candidate checkpoint to the persisted committed checkpoint, deletes only that
receipt's derived manifest/WAL/Parquet, and retains foreign, malformed, active,
and uncommitted worksets.
