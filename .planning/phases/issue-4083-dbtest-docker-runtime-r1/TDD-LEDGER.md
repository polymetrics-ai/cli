# TDD Ledger — Issue #4083

| Stage | Check | Expected result | Evidence |
| --- | --- | --- | --- |
| Red | `TestNewRequiresAnExplicitContainerRuntime` | Fails because `New` silently accepts the hard-wired Podman default instead of requiring an explicit runtime. | `go test -timeout 20m -run '^TestNewRequiresAnExplicitContainerRuntime$' ./internal/connectors/native/dbtest` exited 1: `New() accepted a database test configuration without an explicit Docker or Podman runtime`. |
| Green | Same focused test after the runtime abstraction | Passes for explicit Docker and Podman selections while unknown/unsafe input remains refused. | Pending. |
| Regression | dbtest and MySQL non-live package tests | Pass without a container daemon. | Pending. |
| Live Docker | Tagged MySQL proof with a direct explicit Docker socket | Passes only if actually observed. | Pending availability check. |
| Live Podman | Tagged MySQL proof with a direct explicit Podman socket | Passes only if actually observed. | Pending availability check. |
| Refactor | gofmt, vet, diff check, scoped review | No formatting, safety, or scope drift. | Pending. |

## Red rationale

The protected behavior is not merely an alternate binary name. An explicit
runtime must select its own client command while every invocation remains
pinned to the caller-provided local Unix endpoint. The red test therefore
tests the safety contract and the missing Docker dispatch together.
