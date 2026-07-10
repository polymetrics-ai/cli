# TDD Ledger — Issue #187

## GSD note

`scripts/gsd prompt programming-loop ...` is unavailable (`unknown GSD command: programming-loop`), so this slice uses the manual GSD/TDD fallback.

## Red target

- Freshchat write metadata must mark admin, sensitive, and destructive operations with typed confirmation challenges.
- Commandrunner must carry those challenges into connector command write plans.

Expected initial failure: only Freshchat delete actions currently declare `confirm: destructive`; admin/sensitive writes have no confirmation challenge and the schema only permits `destructive`.

## Green target

Extend confirmation vocabulary, add Freshchat confirmations, and update docs/generated data.

## Verification ledger

Red evidence:

```bash
gofmt -w cmd/connectorgen/freshchat_api_surface_test.go internal/connectors/commandrunner/runner_test.go
go test ./cmd/connectorgen -run TestFreshchatSensitiveAdminWritesRequireTypedConfirmation
```

Initial failure: Freshchat sensitive/admin writes did not declare the expected confirmation challenge (`extract_report` first reported empty confirmation; other sensitive/admin writes were also missing).

Green focused gates:

```bash
go test ./cmd/connectorgen -run 'TestValidate_CLISurface|TestFreshchatSensitiveAdminWritesRequireTypedConfirmation'
go test ./internal/connectors/commandrunner -run TestFreshchatSensitiveAdminWriteCommandsCarryConfirmationChallenges
go run ./cmd/connectorgen validate internal/connectors/defs
```

Results: pass; connectorgen reported `547 connector(s) checked, 0 findings`.

Full gates pass:

```bash
cd website && pnpm run gen:website-data
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```
