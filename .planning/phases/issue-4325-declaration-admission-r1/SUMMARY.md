# Summary — issue 4325 declaration-admission foundation

This phase adds a provider-I/O-free declaration-completeness certificate:
`connectorgen declaration-admission`. A versioned optional sidecar lists cited
provider source operations, their canonical lane/endpoint/CLI command, and
either an existing runtime binding or a named deferred foundation. It is
deliberately separate from source retention, runtime preflight, and live proof.

Deferred commands are now explicit `cli_surface.json` records with a named
`foundation_gap`; they remain discoverable and commandrunner rejects them with a
typed missing-foundation refusal before execution. No connector definition was
converted in this shared foundation phase.
