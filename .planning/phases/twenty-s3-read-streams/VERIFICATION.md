# Twenty S3 read streams — verification

Issue: #280
Head before evidence commit: `935656af`

## Worker-reported green gates

```bash
go test ./internal/connectors/engine -run TestTwentyBundleReadStreams -count=1
go test ./internal/connectors/conformance -run 'TestConformance/twenty' -count=1
gofmt -w cmd internal
go vet ./...
go build ./cmd/pm
go test ./...
go run ./cmd/connectorgen validate
```

Worker result: pass; `connectorgen validate` reported `548 connector(s) checked, 0 findings`.

## Orchestrator pre-PR checks

```bash
jq . internal/connectors/defs/twenty/api_surface.json
jq . internal/connectors/defs/twenty/fixtures/streams/attachments/page_1.json
jq . internal/connectors/defs/twenty/fixtures/streams/attachments/page_2.json
jq . internal/connectors/defs/twenty/streams.json
jq '{endpoints:(.endpoints|length), covered_streams:([.endpoints[]|select(.covered_by.stream?) ]|length), excluded:([.endpoints[]|select(.excluded?)]|length)}' internal/connectors/defs/twenty/api_surface.json
jq '.streams|length' internal/connectors/defs/twenty/streams.json
```

Result: JSON parse pass; `api_surface` has 56 read rows (`covered_streams=28`, `excluded=28`); `streams.json` has 28 streams.

## Independent VERIFY still required

After sub-PR open, run from the S3 worktree at the latest pushed head:

```bash
make connectorgen-validate
go test ./internal/connectors/engine -run TestTwentyBundleReadStreams -count=1
go test ./internal/connectors/conformance -run 'TestConformance/twenty' -count=1
go test ./...
go build ./cmd/pm
scripts/verify-gsd-workflow b4895064
```

Do not run credentialed connector checks. Do not execute reverse ETL.
