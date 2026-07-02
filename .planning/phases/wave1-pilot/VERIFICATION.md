# Local Verification

- CI detected: no
- Local harness required: yes

| Check | Status | Command | Notes |
| --- | --- | --- | --- |
| Install or lockfile validation | missing | TBD | No matching command was detected in the repo profile. |
| Format check | configured | gofmt -l cmd internal |  |
| Lint | configured | go vet ./... |  |
| Typecheck or static analysis | configured | go vet ./... |  |
| Unit tests | configured | go test ./... |  |
| Integration tests | configured | make smoke |  |
| E2E or smoke tests | configured | make smoke |  |
| Build | configured | go build ./... |  |
| Dependency vulnerability scan | missing_optional_tool | TBD | Add or configure dependency scanning before production release. |
| Secret scan | configured | git diff --check |  |
| Accessibility check | missing_optional_tool | TBD | Required for user-facing UI work. |
| Load or benchmark | missing_optional_tool | TBD | Required for performance-sensitive paths. |

