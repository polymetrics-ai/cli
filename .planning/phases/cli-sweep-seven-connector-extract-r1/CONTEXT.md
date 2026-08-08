# Context — seven connector extraction r1

## Locked scope

Extract the completed bundle inputs at source commit `c28bc75a3` into the branch based on current
`main`, then regenerate every derived artifact locally. The fixed connector allowlist is:

`workday-rest`, `jira`, `help-scout`, `greenhouse`, `chatwoot`, `gmail`, and `lever-hiring`.

Do not import `github`, `zendesk-support`, another connector directory, connector-engine code, or
unrelated sweep support. `github` is deliberately excluded because its live parity lane has a newer
1,147-command surface; `zendesk-support` already has a separately owned live lane.

## Integration classification

This is a captain-directed extraction/integration phase, not seven new connector-authoring lanes:
the source commit already contains completed bundle inputs and connector-specific acceptance tests.
Its single target is the immutable source commit plus the seven-path allowlist. The phase may carry
the seven connector-local `cmd/connectorgen/*_api_surface_test.go` acceptance tests and generated
artifacts induced by those bundles. If a shared runtime, schema, command-runner, or generator source
change is required, stop and split it rather than absorbing it.

## Safety and truthfulness

- No credential is read, requested, printed, stored, or used.
- No live connector call is made. Runtime verification is help/preflight-only.
- The PR handoff must say all seven are **implemented, not certified, and never live-tested**; this
  worker holds no credentials for them.
- The real binary, not only `cli_surface.json` or commandrunner preflight, must resolve every
  `availability: implemented` command. The assertion is the exact help `NAME` line because a bare
  connector namespace can exit zero while showing group help.

