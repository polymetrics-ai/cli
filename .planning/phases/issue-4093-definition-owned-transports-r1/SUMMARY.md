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

The acceptance matrix in VERIFICATION.md records stateful proof for valid
registration, zero-side-effect refusals, evidence denial before source I/O,
and the mandated Docker/Colima PostgreSQL integration run.
