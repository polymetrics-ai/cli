# Context — seven connector extraction r1

## Locked scope

Extract the completed bundle inputs at source commit `c28bc75a3` into the branch based on current
`main`, then regenerate every derived artifact locally. The fixed connector allowlist is:

`workday-rest`, `jira`, `help-scout`, `greenhouse`, `chatwoot`, `gmail`, and `lever-hiring`.

Do not import `github`, `zendesk-support`, another connector directory, or unrelated sweep support.
`github` is deliberately excluded because its live parity lane has a newer 1,147-command surface;
`zendesk-support` already has a separately owned live lane. The captain explicitly authorized one
bounded shared foundation: plural `covered_by.writes` support in the engine type, API-surface schema,
and `connectorgen` validator, because Jira and Workday need several named write contracts over one
provider endpoint. It must be its own clearly labelled commit and must retain current-main's unrelated
engine behavior (including REST operation parameters).

## Integration classification

This is a captain-directed extraction/integration phase, not seven new connector-authoring lanes:
the source commit already contains completed bundle inputs and connector-specific acceptance tests.
Its single target is the immutable source commit plus the seven-path allowlist. The phase may carry
the seven connector-local `cmd/connectorgen/*_api_surface_test.go` acceptance tests and generated
artifacts induced by those bundles. Captain authorization extends that allowlist only to the small
plural-write foundation and its focused validator tests; it does not authorize any `github` or
`zendesk-support` tests or unrelated source-branch engine/generator changes.

## Safety and truthfulness

- No credential is read, requested, printed, stored, or used.
- No live connector call is made. Runtime verification is help/preflight-only.
- The PR handoff must say all seven are **implemented, not certified, and never live-tested**; this
  worker holds no credentials for them.
- The real binary, not only `cli_surface.json` or commandrunner preflight, must resolve every
  `availability: implemented` command. The assertion is the exact help `NAME` line because a bare
  connector namespace can exit zero while showing group help.

`cli_surface.json` is a connector-local command contract, not a wholly generated file: command
identity, grouping, summaries, examples, and risk decisions cannot be recreated by `surface-sync`.
It is therefore imported atomically from the fixed source commit, never hand-merged. Immediately
afterward `surface-sync` regenerates its derivable fields (`api_surface`, flag `maps_to`, output
policy, and response cap) from `operations.json`; the root endpoint ledger, manuals, and website
data are generated without importing their source-branch copies.
