# TDD ledger: Google Analytics Data API parity wave03 r1

## Red/green slices

| Slice | Red evidence | Green evidence | Notes |
| --- | --- | --- | --- |
| Single-bundle validate gate | Baseline command `go run ./cmd/connectorgen validate internal/connectors/defs/google-analytics-data-api` fails because `fixtures/` and `schemas/` are treated as connector dirs. Add focused connectorgen test before tooling fix. | `go test ./cmd/connectorgen -run TestValidate_AcceptsSingleBundleDir -count=1`; required single-bundle validate gate passed. | Required by task gate; tooling now recognizes a root `metadata.json` bundle dir. |
| POST-backed read stream surface | GA `runReport` is an official POST read endpoint, but shared surface checks previously treated any executable POST as a write and forced an untruthful GET workaround. | `go test ./cmd/connectorgen -run TestValidate_APISurfaceAllowsPOSTBackedReadStreamWhenWriteFalse -count=1`; `go test ./internal/connectors/conformance -run TestCheckSurfaceComplete_AllowsPOSTBackedReadStreamWhenWriteFalse -count=1`; GA conformance passed with official POST stream rows. | Validation/conformance now distinguish POST-backed read streams/direct reads from write actions. |
| Stream fixture coverage | Existing conformance has one stream fixture only; add fixture/test coverage for all declared streams. | `go test ./internal/connectors/native/google-analytics-data-api -count=1`; required GA conformance passed. | Fixture-only; no provider calls. |
| Official operation ledger | Existing `api_surface.json` has 5 HOOK rows and no official operation ledger. | `go run ./cmd/connectorgen validate internal/connectors/defs/google-analytics-data-api` passed with 11-method ledger. | Re-audit found 11 current v1beta methods: 4 executable official methods, 7 blocked/planned. |
| Native/direct operation behavior | Existing native connector has no typed direct operation tests and no operation ledger. | `TestOperationDirectReadFixtureCoversImplementedOperations` and `TestOperationDirectReadLiveUsesFixedGETEndpoints` passed. | Implemented fixed GET direct reads only: metadata get, audience-exports list/get. |
| Docs/generated surfaces | Existing MANUAL/SKILL describe quarantine/no auth and stale API coverage. | `make verify`; `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`; generated JSON diff entry checks. | Regenerated selectively; broad docs/skills generator churn was reverted. |

## Baseline commands captured

```bash
# Failed before edits; required command currently validates subdirectories as bundles.
go run ./cmd/connectorgen validate internal/connectors/defs/google-analytics-data-api

# Passed before edits.
go test ./internal/connectors/conformance -run 'TestConformance/google-analytics-data-api' -count=1
```

## Safety evidence to preserve

- All tests use fixture mode or local `httptest`.
- Secret placeholders must not match token scanners.
- No live provider API, credential, write, certification, push, PR, or merge.
- Accidental generator inventory found no `*_gen.go` diffs; broad docs/skills churn was reverted, preserving only GA docs/catalog/website generated entries.
