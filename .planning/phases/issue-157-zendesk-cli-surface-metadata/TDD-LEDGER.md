# TDD Ledger: Zendesk CLI Surface Metadata

## Red evidence

- `test -d internal/connectors/defs/zendesk` failed with exit code 1 because the umbrella Zendesk bundle is absent.
- Added `TestBundleLoadEmbeddedZendeskCLISurface`.
- `go test ./internal/connectors/engine -run TestBundleLoadEmbeddedZendeskCLISurface -count=1` failed as expected:

```text
Load(defs.FS, zendesk): load bundle zendesk: missing required file metadata.json
```

## Planned validation

- `go run ./cmd/connectorgen validate internal/connectors/defs` should remain clean after adding Zendesk metadata.

## Green evidence

- Parsed the official Zendesk OAS from `https://developer.zendesk.com/zendesk/oas.yaml`; operation count matched the baseline: 617 operations across 429 paths, methods GET=320, PUT=89, POST=110, DELETE=85, PATCH=13.
- `api_surface.json` lists each operation exactly once as a blocked-by-default operation-ledger row. Candidate model counts: `direct_read=282`, `binary_read=37`, `sensitive_reverse_etl=210`, `destructive_action=85`, `deprecated=3`.
- `cli_surface.json` contains a lean, runtime-embedded command inventory with 5 planned/docs-only category rows across 3 groups, while per-operation detail lives in non-runtime `api_surface.json`; no implemented `raw_api` or `direct_write` command is present.
- `go test ./internal/connectors/engine -run TestBundleLoadEmbeddedZendeskCLISurface -count=1` passed.
- `go test ./cmd/connectorgen -run 'CLISurface|Surface' -count=1` passed.
- `go test ./internal/connectors/engine -run 'CLISurface|Zendesk' -count=1` passed.
- `go run ./cmd/connectorgen validate internal/connectors/defs --json` passed: 548 connectors checked, 0 findings, 0 warnings.

## Refactor notes

- Keep generated helper scripts, if any, in temporary paths unless they become durable project tooling.
- Do not widen `directReadOutputPolicies` or operation schemas unless #161/#162 requires it and has its own red tests.
