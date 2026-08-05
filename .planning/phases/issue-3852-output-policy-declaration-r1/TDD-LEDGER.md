# TDD LEDGER — issue #3852 output-policy declaration

| ID | Enforcement | RED evidence | GREEN evidence | Refactor/verification |
|---|---|---|---|---|
| R1 | A bundle may declare direct-write `json`, and the runtime keeps that decoded response complete. | Planned: `TestBundleLoadAcceptsRuntimeSupportedNonRedactingDirectWriteJSONPolicy` must fail before the enum change because `cli_surface.schema.json` excludes `json`, after `validateOperationDirectWriteOutputPolicy` and `operationDirectWriteResponseBody` establish runtime support. | Pending. | Retain the test and record focused engine output. |
| R2 | The schema enum cannot diverge from the direct-read/write runtime policy sets. | Planned after R1: test initially cannot compile until enumerable registries exist. | Pending: parse the schema enum and compare it to the runtime union plus retained binary compatibility policy. | Focused command-runner test and all-bundle validation. |
| R3 | Existing narrow policies remain schema-valid and authoring guidance selects the non-redacting policy deliberately. | N/A: compatibility is covered by R2's full enum equality. | Pending documentation update and schema set comparison. | Docs grep/check and `connectorgen validate`. |

## CLI/docs/website parity

- Runtime help / bare namespace behavior: not applicable; no command behavior changes.
- `docs/cli/**`, website docs, generated manual/help fixtures, completions: not applicable; no public
  CLI command, flag, or help-topic changes.
- Connector authoring documentation: required; `docs/migration/conventions.md` is updated and
  checked.
