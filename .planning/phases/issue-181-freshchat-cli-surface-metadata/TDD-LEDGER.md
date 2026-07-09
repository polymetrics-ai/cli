# TDD Ledger — Issue #181 Freshchat CLI surface metadata

## Red target

Test name to add before production metadata:

```bash
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedFreshchatCLISurface
```

Expected red result before `internal/connectors/defs/freshchat/cli_surface.json` exists:

- Freshchat bundle loads from embedded defs.
- `b.CLISurface == nil` or has no commands.
- Test fails with message requiring Freshchat `cli_surface.json`.

## Green target

After adding `cli_surface.json`:

```bash
gofmt -w internal/connectors/engine/bundle_test.go
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedFreshchatCLISurface
go run ./cmd/connectorgen validate internal/connectors/defs
```

Expected green result:

- Freshchat CLISurface is non-nil.
- Usage is `pm freshchat <command> [flags]`.
- At least one implemented ETL command maps to `users`.
- At least one implemented reverse-ETL command maps to `create_user`.
- Full connector defs validation has zero findings.

## Refactor notes

Keep test assertions focused on metadata existence and safety-critical mappings. Avoid snapshotting the entire command file to reduce churn.

## Safety evidence

Validation must reject secret-shaped literals if accidentally introduced into `cli_surface.json`; use `go run ./cmd/connectorgen validate internal/connectors/defs` as the safety gate.
