# TDD ledger — CLI PM Broker profile/context foundation

## Red plan

Capture fail-first evidence before production edits:

- `go test ./internal/pmbroker` should fail because the new PM Broker domain/context package does not exist yet.
- `go test ./internal/config -run TestLoadBroker` should fail before safe broker config keys exist.
- `go test ./internal/cli -run 'TestPMBroker(Context|Organizations|Workspaces|Environments)'` should fail before CLI commands/help are wired.

Required behaviors to constrain:

1. PM Broker immutable IDs validate contract-compatible `org_`, `wks_`, `env_`, and `bpf_` shapes while display names remain user-facing metadata only.
2. Versioned safe user state stores contexts with Organization, Workspace, Environment, BrokerProfile, and runtime mode; it rejects ambiguous/mismatched tuples and does not carry secret fields.
3. Context resolution precedence: explicit `--context` > future approval-bound requirement placeholder > project-required context > active user context > synthesized legacy-local context.
4. Runtime modes: `remote`, `local`, and `hybrid`; `hybrid` requires a policy binding; production defaults to `remote`; production writes and scheduled production jobs cannot use local fallback.
5. Incompatible broker contract version seam returns status `426`, code `incompatible_contract_version`, supported version `1.0`, and safe correlation metadata.
6. CLI commands manage/list cached metadata and render JSON/human output without network calls or secrets.
7. Help/manual/docs mention unsupported live provider operations and future fake-client integration TODO rather than duplicating a broker client.

## Red evidence

Pending: add tests and record exact failures.

## Green evidence

Pending implementation.

## Refactor/hardening evidence

Pending verification.

## Skills

`gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-documentation`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-safety`, `golang-naming`, `golang-spf13-cobra`, `golang-spf13-viper`.
