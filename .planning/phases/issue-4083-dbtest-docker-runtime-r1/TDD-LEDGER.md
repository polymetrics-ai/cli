# TDD Ledger — Issue #4083

| Stage | Check | Expected result | Evidence |
| --- | --- | --- | --- |
| CI remediation red | `GOTOOLCHAIN=go1.25.12 go run golang.org/x/vuln/cmd/govulncheck@latest ./...` | Fails because the pinned standard library contains reachable vulnerabilities fixed in Go 1.25.13. | Exited 1 with GO-2026-6218, GO-2026-6091, GO-2026-6090, GO-2026-6089, GO-2026-6088, GO-2026-5972, and GO-2026-5026, all found in Go 1.25.12 and fixed in Go 1.25.13. |
| CI remediation green | `GOTOOLCHAIN=go1.25.13 go run golang.org/x/vuln/cmd/govulncheck@latest ./...` | Passes with no reachable vulnerabilities. | `go version` reported `go1.25.13 darwin/arm64`; the scan exited 0 with `No vulnerabilities found` and `Your code is affected by 0 vulnerabilities.` |
| Red | `TestNewRequiresAnExplicitContainerRuntime` | Fails because `New` silently accepts the hard-wired Podman default instead of requiring an explicit runtime. | `go test -timeout 20m -run '^TestNewRequiresAnExplicitContainerRuntime$' ./internal/connectors/native/dbtest` exited 1: `New() accepted a database test configuration without an explicit Docker or Podman runtime`. |
| Green | Focused runtime-selection/endpoint/target tests | Passes for explicit Docker and Podman selections while unknown/unsafe input remains refused. | `go test -timeout 20m -run '^(TestNewRequiresAnExplicitContainerRuntime\|TestNewAcceptsExplicitDockerRuntime\|TestNewRejectsUnknownContainerRuntime\|TestRuntimeCommandPinsTheConfiguredEndpoint\|TestNewRejectsUnsafeEndpoints\|TestStartUsesDockerTargetIdentityAndCapacity\|TestParseDockerTargetIdentity)$' ./internal/connectors/native/dbtest` passed. |
| Regression | dbtest and MySQL non-live package tests | Pass without a container daemon. | `go test -timeout 20m ./internal/connectors/native/dbtest`, `go test -timeout 20m ./internal/connectors/native/mysql`, and the tagged MySQL selector without its opt-in all passed; the tagged selector visibly skipped before startup. |
| Enabled incomplete config | Tagged MySQL proof with opt-in but no runtime variables | Fails clearly instead of reporting a skipped pass. | `POLYMETRICS_DATABASE_INTEGRATION=1 go test -tags=databaseintegration -count=1 -timeout 20m -run '^TestMySQLContainerHarness$' -v ./internal/connectors/native/mysql` exited 1 and named both required variables, `docker`, `podman`, and the direct Unix socket form. |
| Live Docker | Tagged MySQL proof with a direct explicit Docker socket | Passes only if actually observed. | Not proved: `POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_CONTAINER_RUNTIME=docker POLYMETRICS_CONTAINER_ENDPOINT=unix:///Users/karthiksivadas/.docker/run/docker.sock go test -tags=databaseintegration -count=1 -timeout 20m -run '^TestMySQLContainerHarness$' -v ./internal/connectors/native/mysql` exited 1 because that configured local socket did not report a daemon identity or locally measurable image-store path. |
| Live Podman | Tagged MySQL proof with a direct explicit Podman socket | Passes only if actually observed. | Not proved: the `podman` client binary exists, but `/Users/karthiksivadas/.local/share/containers/podman/machine/podman.sock`, `/run/user/<uid>/podman/podman.sock`, and `/var/run/podman/podman.sock` were absent, so there was no explicit local endpoint that could be tested without reading a global default. |
| Refactor | gofmt, vet, diff check, scoped review | No formatting, safety, or scope drift. | `gofmt -w` on changed Go files, `go vet ./internal/connectors/native/dbtest ./internal/connectors/native/mysql`, and `git diff --check` passed; inline review pending final gate sweep. |

## Red rationale

The protected behavior is not merely an alternate binary name. An explicit
runtime must select its own client command while every invocation remains
pinned to the caller-provided local Unix endpoint. The red test therefore
tests the safety contract and the missing Docker dispatch together.
