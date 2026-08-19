# VERIFICATION — issue #3584 HubSpot / Bitbucket forward remediation

## Planned local gates

Worker-owned ledger gate:

```bash
go test ./.planning/phases/connector-guardrail-remediation-r1/workers/issue-3584
```

Issue-required focused gates:

```bash
go test ./internal/cli ./internal/app ./internal/connectors/bundleregistry
go test ./internal/connectors/boundary ./cmd/connectorgen
```

Broader gates when feasible:

```bash
gofmt -w cmd internal
go vet ./...
go build ./cmd/pm
make verify
```

## CLI help / docs / website parity

Applies: no runtime CLI behavior/help/manual/website change is planned. This worker audits and records prior shared CLI/app/manifest/runtime changes. If a forward CLI behavior change becomes necessary, apply `.agents/agentic-delivery/references/cli-help-docs-website-parity.md` and add runtime help/docs/website checks before PR.

## Current results

| Command | Result | Notes |
| --- | --- | --- |
| `pwd -P` / `git rev-parse --show-toplevel` / `git status --short --branch` | pass | Worker isolated in assigned cwd and branch. |
| `scripts/gsd doctor` | pass | Adapter healthy. |
| `scripts/gsd prompt programming-loop init --phase connector-guardrail-remediation-r1 --dry-run` | fallback | `unknown GSD command: programming-loop`; manual GSD loop active. |
| `scripts/gsd prompt execute-phase connector-guardrail-remediation-r1 --dry-run` | pass | Prompt rendered. |
| worker ledger test red | pass (expected fail) | `go test ./.planning/phases/connector-guardrail-remediation-r1/workers/issue-3584` failed before ledger existed: `read remediation-ledger.json: open remediation-ledger.json: no such file or directory`; exit 1. |
| worker ledger test green | pass | `go test ./.planning/phases/connector-guardrail-remediation-r1/workers/issue-3584` → `ok`. |
| first focused attempt | infra fail | `go test ./internal/cli ./internal/app ./internal/connectors/bundleregistry` initially failed with `/opt/homebrew/.../link: mapping output file failed: no space left on device`; several app tests failed to create `/tmp/...`: no space. Freed local Go build cache with `go clean -cache -testcache`, then reran. |
| issue focused tests | pass | `go test ./internal/cli ./internal/app ./internal/connectors/bundleregistry` → pass (`internal/cli` 454.086s; app/bundleregistry cached). `go test ./internal/connectors/boundary ./cmd/connectorgen` → pass (`boundary` 53.629s, `connectorgen` 10.387s). |
| broader gates | pass | `gofmt -w cmd internal`; `go vet ./...`; `go build ./cmd/pm`; `make verify` all passed. |
| no-mistakes doctor | pass | `no-mistakes doctor` reported daemon running and pi gate validation runnable. Scoped run pending after commit/PR if feasible. |
| sub-PR | pending | Target base `fix/3579-connector-path-ownership-guardrails`. |
| automated review | pending | Route: Claude automatic expected on non-draft sub-PR by trusted author; no manual `@claude` spam. |

## Safety verification

- No secrets requested, printed, summarized, or stored.
- No credentialed connector checks.
- No reverse ETL execution.
- No new dependencies.
- No push to `main`.
- No parent PR merge.
