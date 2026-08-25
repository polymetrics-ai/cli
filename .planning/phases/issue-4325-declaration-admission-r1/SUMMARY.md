# Summary — issue 4325 declaration-admission foundation

This phase adds a provider-I/O-free declaration-completeness certificate:
`connectorgen declaration-admission`. A versioned optional sidecar lists cited
provider source operations, their canonical lane/endpoint/CLI command, and
either an existing runtime binding or a named deferred foundation. It is
deliberately separate from source retention, credential-bound runtime
certification, and live proof. It reuses the real no-I/O command resolver so a
passing admission row cannot drift from runtime preflight semantics.

Deferred commands are explicit `cli_surface.json` records with a typed,
evidenced, exact-target `foundation_gap`; they remain discoverable and
commandrunner rejects them with a typed missing-foundation refusal only after
validating that target. No connector definition is converted in this shared
foundation phase.
