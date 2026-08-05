# TDD LEDGER — issue #3852 output-policy declaration

| ID | Enforcement | RED evidence | GREEN evidence | Refactor/verification |
|---|---|---|---|---|
| R1 | A bundle may declare direct-write `json`, and the runtime keeps that decoded response complete. | Executed 2026-08-06: `go test ./internal/connectors/engine -run '^TestBundleLoadAcceptsRuntimeSupportedNonRedactingDirectWriteJSONPolicy$' -count=1` failed at `bundle_test.go:763`: `cli_surface.json: /commands/0/output_policy: value not in enum [repository_contents_file_metadata repository_contents_directory json_redacted clinical_json_redacted binary_file_bounded]`. The same test first passed `validateOperationDirectWriteOutputPolicy("json")` and verified `operationDirectWriteResponseBody("json", ...)` preserved the complete decoded object. | The same focused command passed after the schema enum gained `json` and `none`. | Retained in `bundle_test.go`; focused engine and combined engine/commandrunner tests passed. |
| R2 | The schema enum cannot diverge from the direct-read/write runtime policy sets. | The RED test exposed one direction of the drift; the pre-change runtime sets could not be enumerated for a bidirectional comparison. | `TestCLISurfaceOutputPolicyEnumMatchesRuntimePolicySets` parses the schema and compares it to enumerable direct-read/direct-write maps plus the retained `binary_file_bounded` compatibility value. Its focused command-runner run passed. | `go run ./cmd/connectorgen surface-sync --check` scanned 550 connectors with 0 corrections; `go run ./cmd/connectorgen validate` checked 550 connectors with 0 findings. |
| R3 | Existing narrow policies remain schema-valid and authoring guidance selects the non-redacting policy deliberately. | N/A: compatibility is covered by R2's full enum equality. | `docs/migration/conventions.md` now directs complete direct-write output to `json` and deliberately empty output to `none`, while preserving existing specialized values as non-default compatibility contracts. | `make docs-check-no-build`, `make connector-boundary`, and the no-schema-drift guard all passed. |

## CLI/docs/website parity

- Runtime help / bare namespace behavior: not applicable; no command behavior changes.
- `docs/cli/**`, website docs, generated manual/help fixtures, completions: not applicable; no public
  CLI command, flag, or help-topic changes.
- Connector authoring documentation: required; `docs/migration/conventions.md` is updated and
  checked.
