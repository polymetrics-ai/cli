# Plan — Issue #187 Freshchat sensitive/admin policy

Refs #187, #180.

## Scope

Add explicit typed-confirmation policy for Freshchat sensitive/admin/destructive reverse ETL commands:

- extend the declarative write confirmation vocabulary beyond `destructive` so metadata can distinguish `admin` and `sensitive` actions;
- tag Freshchat agent-management writes as `admin`, externally visible message/report writes as `sensitive`, and delete actions as `destructive`;
- add validation/runner coverage proving those commands carry confirmation challenges into reverse ETL plans;
- update Freshchat docs/website/generated data to show the extra confirmation gates.

## Out of scope

- Running any Freshchat write, reverse ETL execution, or credentialed command.
- Adding new write actions or expanding scope beyond official Freshchat endpoints already accounted for.
- Changing approval-token mechanics or weakening existing destructive confirmations.

## Required skills

gsd-core, golang-how-to, golang-cli, golang-testing, golang-error-handling, golang-security, golang-safety, golang-structs-interfaces, golang-documentation.

## TDD slices

1. Red: Freshchat sensitive/admin/destructive write actions should declare the expected confirmation challenge (`admin`, `sensitive`, or `destructive`).
2. Red: commandrunner should carry those Freshchat confirmation challenges into write command plans.
3. Green: extend schema/metadata/docs and rerun validation.

## Verification

Focused:

```bash
go test ./cmd/connectorgen -run 'TestValidate_CLISurface|TestFreshchatSensitiveAdminWritesRequireTypedConfirmation'
go test ./internal/connectors/commandrunner -run TestFreshchatSensitiveAdminWriteCommandsCarryConfirmationChallenges
go run ./cmd/connectorgen validate internal/connectors/defs
```

Full handoff:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```
