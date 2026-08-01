# AWS CloudTrail parity wave04 verification checklist

No live provider calls or credentialed checks are part of this verification.

## Initial 3dd65b20d verification

The initial commit verified a 60/19/10/31 implementation with shared promoted-native support. That shared support was later ruled out of scope by the user, so the current corrective head narrows executable reads further to 60/9/0/0 with 51 blocked/planned operations.

## Scope-corrected final surface

Final implemented/blocked counts:

- 60 official CloudTrail API actions inventoried exactly once.
- 9 implemented ETL/read streams.
- 0 implemented direct/provider query commands.
- 0 implemented reverse-ETL write/admin actions.
- 51 blocked/planned operations: 10 parameterized read + 10 direct/provider query + 31 write/admin.
- 0 binary, 0 CDC, 0 excluded.

## Scope correction 2026-08-01

Verification after restoring the six shared files and reclassifying dependent CloudTrail commands:

- [x] `git diff --exit-code HEAD^ -- cmd/connectorgen/main.go cmd/connectorgen/validate.go internal/cli/cli_test.go internal/connectors/bundleregistry/registry.go internal/connectors/engine/connector.go internal/connectors/native/nativeset/promoted.go` -> pass; the six shared files match pre-task contents.
- [x] `go test ./internal/connectors/native/aws-cloudtrail ./internal/connectors/hooks/aws-cloudtrail -count=1` -> pass.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/aws-cloudtrail' -count=1` -> pass.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` -> pass, 549 connectors checked, 0 findings.
- [x] `go build ./cmd/pm` -> pass.
- [x] `go run ./cmd/pm connectors catalog --json` -> AWS CloudTrail reports read-only, 9 streams, 0 write actions.
- [x] `go run ./cmd/pm connectors inspect aws-cloudtrail --json` -> runtime metadata reports `write=false`; manifest remains 0/0 because shared Manifest forwarding is reverted.
- [x] `go run ./cmd/pm aws-cloudtrail --help` -> fails with `help topic "aws-cloudtrail" not found`, which is the truthful reduced help surface because `cli_surface.json` is removed.
- [x] `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test ./internal/cli -run TestGoldenTranscripts -count=1` -> pass and refreshes root help to remove `pm aws-cloudtrail`.
- [x] `./pm docs generate --dir docs/cli --connectors-dir docs/connectors` plus connector docs/catalog truthfulness edits -> CloudTrail generated catalog/docs no longer claim 31 writes.
- [x] `cd website && node scripts/gen-connector-bundles.mjs && node scripts/gen-connector-catalog.mjs` -> AWS CloudTrail website data reports 9 streams, 0 write actions, `cli_surface: null`.
- [x] Posted scope-correction addendum with marker `pm-aws-cloudtrail-wave04-r1-scope-correction` to issues #3142-#3149.
- [x] `make connector-boundary` -> clean.
- [x] `make verify` -> pass.
- [x] `git diff --check` -> pass.
- [x] `git diff --cached --check` -> pass.

Blocked shared-runtime dependencies documented by this scope correction:

1. Focused connector-dir validation depends on the reverted `cmd/connectorgen` shared enhancement; the final supported local gate is whole-defs validation.
2. Runtime-visible CloudTrail dynamic commands, direct/provider queries, and typed reverse-ETL write/admin actions depend on the reverted `internal/connectors/native/nativeset/promoted.go` and `internal/connectors/engine/connector.go` shared forwarding. They are now blocked/planned in the ledger and generated surfaces rather than claimed executable.
