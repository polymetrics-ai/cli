# AWS CloudTrail parity wave04 verification checklist

No live provider calls or credentialed checks are part of this verification.

## Initial 3dd65b20d verification

The initial commit verified a 60/19/10/31 implementation with shared promoted-native support. That shared support was later ruled out of scope by the user, so the current corrective head preserves 60/19/0/0 with 41 blocked/planned operations by keeping resource-detail reads inside connector-local discovery/fan-out while shared direct/write forwarding remains blocked.

## Scope-corrected final surface

Final implemented/blocked counts:

- 60 official CloudTrail API actions inventoried exactly once.
- 19 implemented ETL/read streams.
- 0 implemented direct/provider query commands.
- 0 implemented reverse-ETL write/admin actions.
- 41 blocked/planned operations: 10 direct/provider query + 31 write/admin.
- 0 binary, 0 CDC, 0 excluded.

## Scope correction 2026-08-01

Verification after restoring shared command/direct/write files, keeping a bundle-backed manifest wrapper, and reclassifying dependent CloudTrail commands:

- [x] The final head keeps only promoted-native bundle-backed `Manifest()` forwarding for catalog/inspect truthfulness; CloudTrail command-surface, operation-direct-read, write-validation, and dry-run forwarding remain blocked/planned.
- [x] `go test ./internal/connectors/native/aws-cloudtrail ./internal/connectors/hooks/aws-cloudtrail -count=1` -> pass.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/aws-cloudtrail' -count=1` -> pass.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` -> pass, 549 connectors checked, 0 findings.
- [x] `go build ./cmd/pm` -> pass.
- [x] `go run ./cmd/pm connectors catalog --json` -> AWS CloudTrail reports read-only, 19 streams, 0 write actions.
- [x] `go run ./cmd/pm connectors inspect aws-cloudtrail --json` -> runtime metadata reports `write=false`; manifest reports 19 streams and 0 write actions from the bundle-backed promoted-native wrapper.
- [x] `go run ./cmd/pm aws-cloudtrail --help` -> fails with `help topic "aws-cloudtrail" not found`, which is the truthful reduced help surface because `cli_surface.json` is removed.
- [x] `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test ./internal/cli -run TestGoldenTranscripts -count=1` -> pass and refreshes root help to remove `pm aws-cloudtrail`.
- [x] `./pm docs generate --dir docs/cli --connectors-dir docs/connectors` plus connector docs/catalog truthfulness edits -> CloudTrail generated catalog/docs no longer claim 31 writes.
- [x] `cd website && node scripts/gen-connector-bundles.mjs && node scripts/gen-connector-catalog.mjs` -> AWS CloudTrail website data reports 19 streams, 0 write actions, `cli_surface: null`.
- [x] Posted scope-correction addendum with marker `pm-aws-cloudtrail-wave04-r1-scope-correction` to issues #3142-#3149.
- [x] `make connector-boundary` -> clean.
- [x] `make verify` -> pass.
- [x] `git diff --check` -> pass.
- [x] `git diff --cached --check` -> pass.

Blocked shared-runtime dependencies documented by this scope correction:

1. Focused connector-dir validation depends on the reverted `cmd/connectorgen` shared enhancement; the final supported local gate is whole-defs validation.
2. Runtime-visible CloudTrail dynamic commands, direct/provider queries, and typed reverse-ETL write/admin actions depend on separate promoted-native command-surface, operation-direct-read, write-validation, and dry-run forwarding. They are now blocked/planned in the ledger and generated surfaces rather than claimed executable.
