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

Captured before production code:

```text
$ go test ./internal/pmbroker ./internal/config ./internal/cli -run 'Test(PMBroker|LoadBroker)' -count=1
exit=1
# polymetrics.ai/internal/pmbroker
internal/pmbroker/domain_test.go:15:13: undefined: RuntimeSelection
internal/pmbroker/domain_test.go:21:18: undefined: RuntimeModeRemote
internal/pmbroker/domain_test.go:22:18: undefined: EnvironmentTypeProduction
internal/pmbroker/domain_test.go:23:18: undefined: RuntimeOperationWrite
FAIL polymetrics.ai/internal/pmbroker [build failed]
# polymetrics.ai/internal/config
internal/config/config_test.go:435:9: cfg.Broker undefined (type Config has no field or method Broker)
FAIL polymetrics.ai/internal/config [build failed]
internal/cli/pmbroker_cli_test.go:17: context create exit = 2, want 0; unknown command "context"
internal/cli/pmbroker_cli_test.go:138: help topic "context" not found
FAIL polymetrics.ai/internal/cli
```


## Green evidence

Captured after implementation:

```text
$ go test ./internal/pmbroker ./internal/config ./internal/cli -run 'TestRuntimeModePolicy|TestContextStateValidationAndResolution|TestStoreRejectsUnknownSecretFields|TestLoadBrokerConfig|TestPMBroker' -count=1
ok  	polymetrics.ai/internal/pmbroker
ok  	polymetrics.ai/internal/config
ok  	polymetrics.ai/internal/cli

$ go test ./internal/cli -count=1
go test ./internal/cli PASS

$ go test ./...
all packages passed, including internal/cli and internal/pmbroker
```

## Refactor/hardening evidence

Reviewer/security-auditor follow-up fixed before commit:

- Cleared PM Broker/config environment bindings in CLI tests to keep them deterministic.
- Validated metadata subcommands before loading user state so invalid actions stay usage errors even with poisoned state.
- Rejected non-canonical context names, runtime modes, and hybrid policy bindings to avoid trim-based ambiguity.
- Hardened user-state load to require a regular file that is not group/world-writable.
- Removed a false website claim that `pm docs validate` checks website docs; current docs parity is covered by golden-doc tests and generator checks.

Final local gates:

```text
$ git diff --check
git diff --check PASS

$ go vet ./...
go vet ./... PASS

$ go build ./cmd/pm
go build ./cmd/pm PASS

$ go run ./cmd/pm docs validate --connectors-dir docs/connectors
Validated connector docs in docs/connectors

$ make verify
make verify PASS
```

## Skills

`gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-documentation`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-safety`, `golang-naming`, `golang-spf13-cobra`, `golang-spf13-viper`.
