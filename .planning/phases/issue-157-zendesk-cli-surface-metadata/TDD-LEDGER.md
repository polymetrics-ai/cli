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

## Planned green evidence

- Official OAS operation count matches the baseline or any drift is documented in `VERIFICATION.md`.
- `api_surface.json` lists each operation exactly once and uses blocked-by-default operation rows only until later implementation lanes cover streams/direct reads/writes/binary.
- `cli_surface.json` validates and contains no implemented raw API/direct-write command.
- `engine.Load(defs.FS, "zendesk")` loads `CLISurface` from embedded runtime metadata.

## Refactor notes

- Keep generated helper scripts, if any, in temporary paths unless they become durable project tooling.
- Do not widen `directReadOutputPolicies` or operation schemas unless #161/#162 requires it and has its own red tests.
