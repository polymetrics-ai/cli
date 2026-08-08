# TDD Ledger — Zoom Customer Managed Keys Hybrid documented-operation parity, R1

## RED — captured before production declarations

The RED checkpoint contains only tests, synthetic fixture/evidence, and planning artifacts. It
fails against the pre-CMK Zoom bundle and current direct-write engine because:

- the sole provider endpoint remains a blocked Zoom ledger row, so executable coverage is still
  `22`, local contracts still `1820`, no Customer Managed Keys Hybrid command is reachable, and
  commandrunner reports the planned path as unknown;
- `json_redacted` direct-write output currently retains a generic secret-shaped field and an
  operation-declared sensitive response field;
- the plan sample for a direct-write command still retains declared sensitive request input.

The test was run before any production declaration or engine change:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/app ./internal/connectors/defs/zoom/...
--- FAIL: TestOperationDirectWritePreviewsApprovesAndExecutesSingleFormRequest (0.00s)
    direct_write_test.go:127: json_redacted direct-write result exposed a generic token-shaped field
--- FAIL: TestOperationDirectWriteJSONRedactedErrorsHideDeclaredRequestAndResponseFields (0.00s)
    direct_write_test.go:236: json_redacted direct-write error exposed sensitive request or response content
--- FAIL: TestSecretSensitiveDirectWriteRequiresTypedConfirmation (0.00s)
    direct_write_test.go:260: secret-sensitive direct-write target = engine.DestructiveTarget{Connector:"acme", Operation:"acme.decrypt", Method:"POST", MutationClass:"secret", Destructive:false, Confirmation:""}, want typed destructive confirmation
FAIL
FAIL    polymetrics.ai/internal/connectors/engine    5.762s
--- FAIL: TestBuildOperationDirectWriteCommandUsesTypedInputsAndPlanLifecycle (0.00s)
    runner_test.go:1811: direct-write preview record id = "t3_abc", want declared redaction
FAIL
FAIL    polymetrics.ai/internal/connectors/commandrunner    8.832s
--- FAIL: TestDirectWriteCommandPlanPreviewApprovalAndExecute (1.17s)
    rest_write_command_test.go:247: plan preview sample = []connectors.Record{connectors.Record{"dir":1, "id":"t3_abc"}}, want declared direct-write redaction
FAIL
FAIL    polymetrics.ai/internal/app    219.937s
--- FAIL: TestProviderInventoryLedgerIsComplete (0.03s)
    command_surface_test.go:154: executable rows = 22, want 23
    command_surface_test.go:157: operations awaiting Zoom-local contracts = 1820, want 1819
--- FAIL: TestCoveredStreamsHaveReachableCommands (0.03s)
    command_surface_test.go:253: reachable direct_write operation commands = 0, want 1
--- FAIL: TestCustomerManagedKeysHybridCommandExecutesWithFixture (0.03s)
    command_surface_test.go:1186: BuildWriteCommand without encrypt-context = connector command "customer-managed-keys-hybrid archival-key decrypt" is blocked: unknown command, want required typed flag rejection
FAIL
FAIL    polymetrics.ai/internal/connectors/defs/zoom    3.646s
FAIL
```

No `operations.json`, `streams.json`, `spec.json`, `cli_surface.json`, `api_surface.json`,
metadata, docs, website catalog, or generated endpoint ledger changes in this checkpoint. The
foundation/connector GREEN commits follow this red commit.

## GREEN foundation — completed before Zoom authoring

The foundation is intentionally separate from Zoom authoring. It changes only the reusable
direct-write engine/command surface path:

- `json_redacted` now redacts generic secret-shaped response fields and every declared
  `sensitive_policy.redact_fields` value before result output.
- A `json_redacted` provider error preserves status/URL context but treats the body as redacted;
  declared request literals are additionally removed before error persistence.
- A direct-write plan sample applies its command-declared redaction fields while retaining the
  private typed execution record needed to bind preview and execution.
- `secret_sensitive` (or `mutation_class=secret`) plus
  `sensitive_policy.approval_mode=typed_confirmation` now maps to the existing closed destructive
  confirmation grant. There is no new prompt or approval vocabulary.

The foundation tests are green while the Zoom surface intentionally remains red until the next
connector declaration commit:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine ./internal/connectors/commandrunner
ok  polymetrics.ai/internal/connectors/engine  4.801s
ok  polymetrics.ai/internal/connectors/commandrunner  7.491s

$ go test -count=1 -timeout 20m -run 'TestDirectWriteCommand(PlanPreviewApprovalAndExecute|FailureRedactsDeclaredOutputPolicyContent)$' ./internal/app
ok  polymetrics.ai/internal/app  3.241s
```

This reusable foundation is committed separately from Zoom JSON authoring in `0987a58bc` and is
pushed before the connector declaration begins.

## GREEN connector — pending

The connector declaration will make the existing RED Zoom surface test green: one exact POST,
one endpoint-ledger row, an approval-gated typed command, required body flags, customer-hosted
JWT profile selection, and redacted response/error outputs. Generated docs/catalog changes will
be regenerated and mechanically scoped to Zoom.
