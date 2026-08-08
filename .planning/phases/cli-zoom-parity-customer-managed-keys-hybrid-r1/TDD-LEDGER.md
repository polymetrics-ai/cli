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

## RED operation-origin/auth foundation — captured before foundation code

The provider's Key Connector archival request has a distinct customer-hosted origin and a distinct
JWKS-signed bearer credential. Reusing the bundle-wide Zoom OAuth transport for that request
would make a configuration mistake capable of sending the ordinary Zoom bearer to the customer
host. The focused test declares the two fields in `operations.json` and exercises the real
preview/approval/execute path; it is RED because the current operation schema has no such
operation-scoped origin/auth contract.

```text
$ go test -count=1 -timeout 20m -run TestOperationDirectWriteUsesDeclaredOperationOriginAndAuth ./internal/connectors/engine
--- FAIL: TestOperationDirectWriteUsesDeclaredOperationOriginAndAuth (0.00s)
    direct_write_test.go:223: Load declared per-operation origin/auth bundle: load bundle acme: operations.json: /operations/0/rest/auth: additional property not allowed
FAIL
FAIL    polymetrics.ai/internal/connectors/engine    0.717s
FAIL
```

The test uses loopback servers and synthetic credentials only. It asserts that the ordinary API
server receives zero requests and the operation-scoped server receives exactly the declared bearer
request. It is committed and pushed before extending the engine/meta-schema.

## GREEN operation-origin/auth foundation

Commit 833a2d9d4 adds a deliberately narrow, paired rest.base_url plus rest.auth override for
rest_write only:

- The operation loader rejects either field alone and rejects both on non-write operations.
- Preview binds its request URL to the same declared origin later used by execution.
- Execution clones only the operation's transport fields before creating the runtime, retaining the
  existing request shaping, rate-limit selection, approval, no-retry, redirect, and redaction
  controls.
- The focused loopback test proves the ordinary API server sees zero requests while the
  operation-scoped server sees exactly one request with its declared bearer authentication.
- connectorgen validate statically checks the operation-scoped URL/auth templates against
  spec.json just as it does the bundle-wide transport fields.

~~~text
$ go test -count=1 -timeout 20m -run TestOperationDirectWriteUsesDeclaredOperationOriginAndAuth ./internal/connectors/engine
ok  polymetrics.ai/internal/connectors/engine  0.750s

$ go test -count=1 -timeout 20m ./internal/connectors/engine
ok  polymetrics.ai/internal/connectors/engine  4.843s

$ go test -count=1 -timeout 20m ./cmd/connectorgen
ok  polymetrics.ai/cmd/connectorgen  10.938s

$ go vet ./internal/connectors/engine ./cmd/connectorgen
~~~

This foundation is separate from the Zoom declaration and is pushed in 833a2d9d4. It is available
to any future declarative customer-hosted rest_write that needs a distinct origin and credential;
no other connector bundle is changed by this foundation commit.

## RED direct-write endpoint-ledger foundation — captured before production code

The Customer Managed Keys Hybrid declaration reaches the real direct-write preflight, but the
endpoint ledger reconciler only knows `direct_read` operation rows. It therefore refuses the
provider's `sensitive_reverse_etl` row instead of replacing it with a runtime-proven executable
coverage binding. The focused fixture uses a declared `rest_write` operation, its matching
approval-gated direct-write command, and the real `commandrunner.Preflight` path; it contains no
provider credential or key material.

```text
$ go test -count=1 -timeout 20m -run TestRunSurfaceReconcileCoversSensitiveDirectWriteWithRuntimePreflight ./cmd/connectorgen
--- FAIL: TestRunSurfaceReconcileCoversSensitiveDirectWriteWithRuntimePreflight (0.02s)
    surfacereconcile_test.go:60: stats = {Scanned:1 Covered:0 Blocked:0 Unchanged:0 Refused:1}, want one runtime-covered direct write
FAIL
FAIL	polymetrics.ai/cmd/connectorgen	0.754s
FAIL
```

The RED checkpoint commits only this test and GSD/TDD evidence. It leaves the pending Zoom
declaration uncommitted, so the foundation remains independently reviewable.

## GREEN direct-write endpoint-ledger foundation

Commit `410eb1bb7` adds the reusable direct-write endpoint coverage contract:

- `api_surface.json` can record `covered_by.direct_write` or `covered_by.direct_writes`; both
  resolve only to implemented direct-write commands and only on mutation methods.
- `surface-reconcile` promotes the eligible sensitive/admin/destructive operation models only
  after the exact command passes the real runtime preflight. It retains a typed blocked reason
  when no candidate passes and refuses all other operation models.
- Direct-write command preflight now carries its operation, HTTP method, path, and output policy
  to the engine. A command cannot claim a different endpoint from its declared `rest_write`
  operation.
- The runtime accepts a generated direct-write ledger row after reconciliation, while its embedded
  fallback remains derived solely from the shipped operation declaration.
- Static validation, conformance, and certification inventory accounting treat direct-write
  coverage as executable mutation surface and enforce `capabilities.write`.

This foundation unblocks every declarative connector with a safely promotable `rest_write`
operation that is currently held in `sensitive_reverse_etl`, `admin_reverse_etl`, or
`destructive_action` ledger state; no existing connector ledger was reclassified in this commit.
It is separate from Zoom JSON authoring.

```text
$ go test -count=1 -timeout 20m ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/connectors/conformance ./internal/connectors/certify
ok  polymetrics.ai/cmd/connectorgen  14.463s
ok  polymetrics.ai/internal/connectors/engine  5.791s
ok  polymetrics.ai/internal/connectors/commandrunner  9.483s
ok  polymetrics.ai/internal/connectors/conformance  20.228s
ok  polymetrics.ai/internal/connectors/certify  13.560s
```

## GREEN connector — pending

The connector declaration will make the existing RED Zoom surface test green: one exact POST,
one endpoint-ledger row, an approval-gated typed command, required body flags, operation-scoped
customer-hosted JWT selection, and redacted response/error outputs. Generated docs/catalog changes
will be regenerated and mechanically scoped to Zoom.
