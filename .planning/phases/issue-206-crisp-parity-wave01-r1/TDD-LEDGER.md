# TDD LEDGER — Crisp provider parity Wave 1

## Red / pre-change evidence

- Passed: inventory-only commit `5e212f869` JSON-validates with 234 endpoint rows, each carrying a cited blocked operation disposition.
- Passed (red): `go test ./internal/connectors/defs/crisp -run TestCrispWaveOneParityContract -count=1` failed before bundle creation with `missing required file metadata.json`.
- Deferred: pre-bundle `pm crisp` command execution was not useful because the dynamic connector surface did not yet exist; post-bundle command-resolution checks are the authoritative reachability evidence.

## Green / implementation evidence

- Passed: the connector-local parity test now verifies 21 documented GET rows against 21 read-only streams and individually named CLI commands with matching `{method,path}` entries.
- Passed: HTTP Basic identifier/key configuration and a fixture-only safe connection check are declared.
- Passed: 213 unimplemented provider rows retain cited, explicit blocked dispositions; this bundle declares no output redaction metadata.
- Passed: `TestCrispListCommandPreservesFixtureContent` replays the Crisp list fixture through a local HTTP server and `commandrunner.Run`; its synthetic `content` field reaches the emitted record unchanged after the #3868/#3872 rebase.

## Refactor / verification evidence

- Passed: `go run ./cmd/connectorgen validate internal/connectors/defs` and `go run ./cmd/connectorgen surface-sync --check` report zero findings/drift.
- Passed: `go test` for Crisp, commandrunner, engine, conformance, connectors, and `internal/cli`; targeted `go vet` also passes.
- Passed: `go build -o ./pm ./cmd/pm`, `pm help crisp`, `pm crisp`, and all 21 `pm crisp <command> --help` command-resolution checks pass without credentials or provider traffic.
- Passed: `gofmt`, `git diff --check`, fresh isolated generated-manual comparison, docs validation, website tests/typecheck, and the scoped Make verification gates pass.
