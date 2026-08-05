# Phase: cli-engine-secret-input-r1

## GSD execution record

The repo-local GSD adapter is healthy, but `scripts/gsd prompt programming-loop` is not a
registered adapter command. This phase therefore follows the required programming-loop lifecycle
manually and inline. The fallback does not waive planning, strict red-first work, verification, or
the phase evidence files.

**Required skills used:** `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`,
`golang-design-patterns`, `golang-structs-interfaces`, `no-mistakes`, and
`gsd-programming-loop`. The CLI help/docs/website parity reference was also read.

**Orchestration decision:** `local_critical_path`. This is a small shared CLI/runner/definition
boundary; splitting it would create overlapping ownership of the same dispatch path. The current
runtime instruction also forbids proactive delegation.

## Scope

Implement the typed, non-inline secret-input boundary for the nine secret-sensitive Zendesk
Support operations. The boundary accepts source *references* only (environment-variable names or
one explicitly selected stdin target), validates them against the operation's declared typed input
shape, and keeps resolved values transient and out of diagnostics.

The affected operation identities are:

- `zendesk-support.validate_token`
- `zendesk-support.create_oauth_client`
- `zendesk-support.update_client`
- `zendesk-support.client_generate_secret`
- `zendesk-support.create_oauth_token`
- `zendesk-support.set_user_password`
- `zendesk-support.change_own_password`
- `zendesk-support.create_or_update_user`
- `zendesk-support.create_token_for_grant_type`

## Storage-seam ruling

The firstmate decision `secret-storage-seam-collision` settles the prior collision: connector
mechanism foundations owns persistence. This phase must not retain or restore the legacy
`SecretStore` seam and must not modify shared storage. It may define the narrow command-input
requirements a later binding needs, but it defers persistence implementation until that lane lands
on `main` and this branch rebases onto it.

## Non-negotiable safety properties

- No secret value is accepted from a command argument. The command-line surface carries only a
  target field and an environment-variable name, or a target field naming stdin input.
- Secret values never enter plan metadata, JSON output, logs, error text, or an argv-derived
  `Flags` map.
- Failure paths use fixed, field-oriented errors and redact wrapped text before it can surface.
- The existing encrypted credential vault remains the only eventual persistence mechanism; this
  phase adds no store, file, config field, or plaintext fallback.

## Tasks

- [ ] Task: secret-input-parse type: behavior — add a reusable parser for the existing
  `--from-env field=ENV` and `--value-stdin field` source-reference conventions. It rejects inline
  values, duplicate/conflicting sources, undeclared field paths, missing environment values, and
  more than one stdin target without ever echoing a resolved value.
- [ ] Task: secret-input-typing type: behavior — validate a parsed secret target against a declared
  string body-field path and the operation's `env_or_stdin` policy before a request body is built.
  The implementation stays provider-neutral; Zendesk definitions supply the allowed paths.
- [ ] Task: secret-input-leak-safety type: behavior — prove a sentinel secret is absent from argv
  representations, plan/preview output, logs, and returned errors when parsing, validation, or an
  operation fails. Mutation-check the redaction assertion.
- [ ] Task: zendesk-nine-op-metadata type: behavior — after the prerequisite `rest_write` and
  schema-deriver work lands, expose the typed sources only on the nine Zendesk operations and make
  all existing help/manual/website generated surfaces agree.
- [ ] Task: storage-seam-binding type: behavior — after mechanism foundations lands, bind transient
  values through its approved encrypted credential seam. This task is deliberately deferred by the
  ruling above; no alternate persistence is permitted.

## Explicit exclusions

- Configuration-time constraint validation (owned by its separate lane).
- Changes to vault, browser-auth, key protection, or any credential persistence implementation.
- Generic raw body, HTTP, SQL, shell, file, or secret-value command interfaces.
- Provider behavior outside the nine named Zendesk operations.

## Rebase gate

Before final integration, rebase onto `main` after the rest-write executor, schema-deriver union
fix, and connector mechanism foundations have landed. Re-read the landed storage API rather than
assuming the superseded refresh-token seam still exists. If the landed seam cannot satisfy the
narrow command-input persistence requirements, report that finding; do not work around it.
