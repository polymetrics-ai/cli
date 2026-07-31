# AWS CloudTrail parity wave04 verification checklist

Required final gates from task:

- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/aws-cloudtrail --json` -> 1 connector checked, 0 findings.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/aws-cloudtrail' -count=1` -> pass.
- [x] Focused connector/native tests: `go test ./internal/connectors/native/aws-cloudtrail ./internal/connectors/hooks/aws-cloudtrail -count=1` -> pass.
- [x] Focused CLI tests: `go test ./internal/cli -run 'TestGoldenTranscripts|TestGoldenDocsGenerateMatchesTrackedCLIManuals|TestAWSCloudTrailCommandSurfaceHelpScopes|TestRootHelpListsDynamicConnectorCommands' -count=1` -> pass.
- [x] `go build ./cmd/pm` -> pass.
- [x] `make connector-boundary` -> clean.
- [x] `make verify` -> pass.
- [x] `git diff --check` -> pass.

Additional parity/documentation checks:

- [x] `go run ./cmd/pm help docs` read before docs generation.
- [x] Regenerated connector docs/skills for `aws-cloudtrail`, connector catalog docs, CLI golden transcripts, and website connector data.
- [x] `go run ./cmd/pm connectors inspect aws-cloudtrail --json` contains 19 streams, 31 write actions, config/secret metadata, and no secret values.
- [x] `go run ./cmd/pm aws-cloudtrail --help`, `go run ./cmd/pm aws-cloudtrail read describe-trails --help`, and `go run ./cmd/pm aws-cloudtrail events lookup --help` render connector command help successfully.
- [x] Issues #3142-#3149 received a single idempotent captain-policy addendum with actual 60/19/10/31/0/0/0 counts.
- [x] Optional website unit test attempted with `cd website && npm run test:unit -- tests/api/connector-data.test.ts`; local repo lacks installed `vitest` binary (`sh: vitest: command not found`), so website data parity relies on generator output plus Go docs/catalog checks.

No live provider calls or credentialed checks are part of this verification.
