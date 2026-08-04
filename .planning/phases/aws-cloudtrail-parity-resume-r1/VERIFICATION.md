# Verification checklist: AWS CloudTrail parity resume r1

- [ ] Verify the official AWS Actions page's 60-operation total and record its URL.
- [ ] Capture red native-provider and runtime-help evidence before the delegate.
- [ ] Add a native command-surface delegate and coverage test.
- [ ] Run `go run ./cmd/connectorgen surface-sync`.
- [ ] Run `go run ./cmd/connectorgen surface-sync --check`.
- [ ] Run `go run ./cmd/connectorgen validate aws-cloudtrail`.
- [ ] Run `go test ./internal/connectors/conformance/...` scoped to AWS CloudTrail.
- [ ] Run `go test ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight`.
- [ ] Run the affected native/hooks tests and `go test ./internal/cli/...`.
- [ ] Run `go vet` for changed packages and `go build ./cmd/pm`.
- [ ] Build and run `pm help aws-cloudtrail`, `pm aws-cloudtrail`, `pm aws-cloudtrail --help`, and representative direct-read/write command help or preflight without credentials.
- [ ] Run `cd website && pnpm run gen:website-data`; commit only AWS CloudTrail generated-data changes if any.
- [ ] Run `git diff --check` and verify changed-path compliance.
