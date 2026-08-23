# Verification — issue #4342 binary upload CLI and certification foundation

## Planned gates

- [ ] `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/engine -run 'BinaryUpload|Write' -count=1`
- [ ] `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/commandrunner -run 'BinaryUpload|WriteCommand' -count=1`
- [ ] `GOFLAGS='-p=3' go test -timeout 20m ./internal/cli -run 'BinaryUpload|Connector.*Upload' -count=1`
- [ ] `GOFLAGS='-p=3' go test -timeout 20m ./internal/connectors/certify -run 'Binary.*Upload|Report' -count=1`
- [ ] `GOFLAGS='-p=3' go test -timeout 20m ./cmd/connectorgen -run 'Certification.*Upload|OperationEvidence' -count=1`
- [ ] `GOFLAGS='-p=3' go run ./cmd/connectorgen validate internal/connectors/defs`
- [ ] `GOFLAGS='-p=3' go run ./cmd/connectorgen surface-sync --check`
- [ ] `GOFLAGS='-p=3' go run ./cmd/connectorgen operation-evidence --check`
- [ ] Generator/docs parity checks located from `make verify` and updated artifacts.
- [ ] `GOFLAGS='-p=3' go vet ./...`
- [ ] `GOFLAGS='-p=3' go build ./cmd/pm`
- [ ] GSD `verify-work` and `code-review` prompts generated and executed inline with dispositions recorded.

## Constraint

No credentialed provider call is authorized for this task. Tests must use the existing declaration-bound fixture/provider doubles and assert actual byte transfer through that real application path. A missing live candidate is `not_live`, not a passing transfer assertion.
