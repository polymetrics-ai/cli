# VERIFICATION — issue 3581 target-scope core validator

## Required local gates

```bash
gofmt -w cmd internal
go test ./internal/connectors/boundary ./cmd/connectorgen
go run ./cmd/connectorgen ownership . --help
go build ./cmd/pm
```

## Broader gates if feasible

```bash
go vet ./...
make verify
```

## CLI help / docs / website parity

Applies: yes, this adds `connectorgen ownership` developer CLI surface.

Planned checks:

- `go run ./cmd/connectorgen ownership . --help`
- `go run ./cmd/connectorgen ownership . --json` with controlled fixture/working tree path where feasible
- docs under `docs/migration/**` updated with usage or explicit narrow exemption
- `docs/cli/**`, `website/**`, generated pm manual artifacts: likely not applicable because `connectorgen` is a developer generator command, not the shipped `pm` command surface; record final decision after implementation.

## Current evidence

- Sub-PR: https://github.com/polymetrics-ai/cli/pull/3590
- Latest remote checks observed: 9 passed, 0 failed, 3 skipped, 4 pending.
- Automated review: no Claude/Copilot review records observed yet on sub-PR; route pending `claude_auto` / parent fallback.
- `pwd -P`: `/Users/karthiksivadas/.treehouse/cli-83d592/5/worker-3581-core`
- `git rev-parse --show-toplevel`: `/Users/karthiksivadas/.treehouse/cli-83d592/5/worker-3581-core`
- `git status --short --branch`: `## fix/3581-target-scope-core-validator`
- `scripts/gsd doctor`: pass
- `scripts/gsd list`: pass
- `scripts/gsd prompt execute-phase connector-guardrail-remediation-r1 --dry-run`: pass / prompt generated
- `scripts/gsd prompt programming-loop init --phase connector-guardrail-remediation-r1 --dry-run`: fail with `scripts/gsd: unknown GSD command: programming-loop`; manual-GSD fallback active

## Final verification log

```bash
gofmt -w cmd internal
go test ./internal/connectors/boundary ./cmd/connectorgen
go run ./cmd/connectorgen ownership . --help
go build ./cmd/pm
```

Result: pass.

```text
ok  	polymetrics.ai/internal/connectors/boundary	53.875s
ok  	polymetrics.ai/cmd/connectorgen	10.227s
usage:
  connectorgen ownership [repo-root] [--json] [--base <ref>] [--scope-file <path>]
  connectorgen ownership [repo-root] [--json] --changed-path <path> [--changed-path <path>...] [--scope-file <path>]
...
```

```bash
go run ./cmd/connectorgen ownership . --changed-path internal/connectors/defs/github/metadata.json --json
```

Result: pass; returned `kind: ConnectorOwnershipReport`, `outcome: clean`, `target_connector: github`, and stable empty arrays for `findings` and `warnings`.

```bash
go vet ./...
```

Result: pass.

```bash
make connector-boundary
```

Result: pass; returned `outcome: clean`.

`make verify` not run: the target includes local smoke with reverse ETL execution; user hard rules for this sub-issue prohibit reverse ETL execution. Required focused gates and `go vet ./...` passed instead.
