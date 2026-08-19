# TDD Ledger — Issue #3773 api_surface provenance

## Red / green evidence plan

| Slice | Red evidence to capture | Green evidence required |
| --- | --- | --- |
| #3785 structural v2 contract | A v2 sample containing root `artifacts` and endpoint `provenance` is rejected by the current closed `api_surface` schema / strict decoder. | A complete v2 ledger decodes into typed artifacts/provenance; v1 remains accepted; unknown fields and classifier violations remain rejected. |
| #3787 semantic validation | Once structural v2 loading works, a v2 endpoint can initially conform despite missing citation, duplicate artifact, unknown artifact, non-HTTPS URL, or invalid date. | Shared validator reports endpoint index/method/path and conformance reports a failing `surface_complete` result; v1 is unchanged. |
| #3789 generator enforcement | The real `connectorgen validate` path initially does not surface a malformed v2 provenance relation. | It emits connector/file/endpoint-specific findings from the shared validator; a valid v2 fixture is clean and v1 is not promoted. |
| #3791 certification evidence | Existing certification report has no provenance summary and cannot distinguish v1 from valid/invalid v2. | Deterministic fixture tests assert `legacy_unverified`, `complete`, and actionable `invalid` evidence output without capability changes. |

## Safety assertions retained by tests

- `covered_by` accepts only declared stream/write/implemented-direct-read targets and remains the
  sole capability/executor binding.
- Provenance is not a classifier and cannot satisfy, replace, or add a `covered_by` target.
- Blocked and destructive rows remain blocked/destructive after v2 evidence is added.
- V1 compatibility avoids orphaning the existing fleet during the separate evidence migration.

## Commands to record

```sh
go test ./internal/connectors/engine -run 'TestBundleLoad.*APISurface|Test.*Provenance' -count=1
go test ./internal/connectors/conformance -run 'Test.*Surface|Test.*Provenance' -count=1
go test ./cmd/connectorgen -run 'Test.*Provenance|TestValidate' -count=1
go test ./internal/connectors/certify -run 'Test.*Surface.*Provenance|Test.*Provenance' -count=1
```

## Recorded evidence

- **#3785 red → green:** before the schema/model change,
  `go test ./internal/connectors/engine -run '^TestBundleLoadAPISurfaceV2ProvenanceContract$' -count=1`
  rejected the v2 fixture at `//artifacts: additional property not allowed`. The same focused test
  is green after adding the closed `artifacts` and endpoint `provenance` shapes and typed loader
  model; its table retains the v1, unknown-key, and non-classifier cases.
- **#3787 red → green:** before conformance integration, a v2 covered endpoint without
  provenance passed `checkSurfaceComplete`, while a v2 sensitive blocked row without a legacy
  `operation.source_url` failed. The focused conformance tests are green once the shared validator
  runs before unchanged `covered_by` resolution and v2 endpoint provenance supplies only the
  blocked-row citation.
- **#3789 red → green:** the consumer test was introduced before `ruleSurfaceProvenance` and the
  shared-validator call existed. It now proves complete, missing-citation, unknown-artifact,
  duplicate-artifact, non-HTTPS-citation, malformed-date, and v1 paths through the real
  `validateDir` entry point.
- **#3791 red → green:** report fixture tests initially failed because `SurfaceResult` had neither
  provenance evidence nor a raw-inventory helper; text rendering initially omitted surface
  evidence. The retained tests now prove `complete`, `invalid`, and `legacy_unverified` results in
  the report, text, and JSON paths.
