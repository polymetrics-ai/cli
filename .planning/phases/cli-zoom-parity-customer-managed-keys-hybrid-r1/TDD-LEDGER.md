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

## GREEN connector — completed

Commit `cec675503` declares the sole Customer Managed Keys Hybrid operation and reconciles its one
existing blocked ledger row to `covered_by.direct_write`:

- `zoom.decrypt_customer_managed_key_archival` is a single non-batchable `rest_write` POST with
  exactly `encrypt_context` and `key_id` in its closed JSON body schema. It uses `json_redacted`,
  `mutation_class=secret`, and the established plan → preview → approval → typed-confirmation
  lifecycle.
- The command is `customer-managed-keys-hybrid archival-key decrypt`, with exactly the two
  required declared body flags and no page, per-page, limit, cursor, or other invented input.
- Its paired operation transport derives the host from `key_connector_base_url` and bearer from
  `key_connector_jwt`; the normal Zoom OAuth credential is not selected for that origin.
- The Zoom fixture is entirely synthetic and asserts exact POST/path/body/auth plus redaction of
  the returned key ID, plaintext key, and generic token-shaped field.

The declaration and generated surface are green:

```text
$ go run ./cmd/connectorgen surface-sync --check
connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)

$ go run ./cmd/connectorgen validate internal/connectors/defs/zoom
connectorgen validate: 1 connector(s) checked, 0 findings

$ go test -count=1 -timeout 20m ./internal/connectors/defs/zoom/...
ok  polymetrics.ai/internal/connectors/defs/zoom
```

Generated Zoom documentation and the Zoom-only website catalog projection are in separately
pushed commit `f86e6a480`. The shared docs indexes were restored whole because their generator
output also contained unrelated Gorgias stale-doc changes; no generated file was hand-merged.

## RED/GREEN review hardening — inherited header isolation

Manual code review found that a paired operation origin/auth override still inherited bundle-wide
HTTP headers. A header such as `Authorization` or an API-key header can carry an ordinary API
credential, so sending it to a customer-hosted origin would violate the transport boundary.

The focused RED test was committed and pushed before its production fix in `5c9518918`:

```text
$ go test -count=1 -timeout 20m -run TestOperationDirectWriteUsesDeclaredOperationOriginAndAuth ./internal/connectors/engine
--- FAIL: TestOperationDirectWriteUsesDeclaredOperationOriginAndAuth (0.00s)
    direct_write_test.go:233: key connector request inherited an ordinary API secret header
FAIL
FAIL    polymetrics.ai/internal/connectors/engine
```

Commit `dfa221bcd` clears inherited bundle headers whenever a `rest_write` selects its paired
operation-scoped origin/auth transport. The exact focused test and full engine package are green:

```text
$ go test -count=1 -timeout 20m -run TestOperationDirectWriteUsesDeclaredOperationOriginAndAuth ./internal/connectors/engine
ok  polymetrics.ai/internal/connectors/engine

$ go test -count=1 -timeout 20m ./internal/connectors/engine
ok  polymetrics.ai/internal/connectors/engine
```

## Final GREEN verification

```text
$ go test -count=1 -timeout 20m ./cmd/connectorgen ./internal/connectors/defs/zoom/... ./internal/connectors/commandrunner ./internal/connectors/conformance ./internal/connectors/certify
ok  polymetrics.ai/cmd/connectorgen
ok  polymetrics.ai/internal/connectors/defs/zoom
ok  polymetrics.ai/internal/connectors/commandrunner
ok  polymetrics.ai/internal/connectors/conformance
ok  polymetrics.ai/internal/connectors/certify

$ go test -count=1 -timeout 20m ./internal/app
ok  polymetrics.ai/internal/app

$ go test -count=1 -timeout 20m ./internal/cli
ok  polymetrics.ai/internal/cli

$ go vet ./...
exit 0

$ make lint
0 issues.

$ go run ./cmd/connectorgen surface-reconcile --check --notes-contains provider_module=customer-managed-keys-hybrid
connectorgen surface-reconcile: 551 connector(s) scanned; covered=0 blocked=0 unchanged=0 refused=0
```

The final built binary passed `pm help zoom`, bare `pm zoom`, bare Customer Managed Keys Hybrid,
and the exact command help route. It then performed plan, no-network preview, approval, typed
confirmation, and one declared POST against an isolated loopback Key Connector. The script asserted
the returned synthetic key material and generic token field were redacted before output; no
synthetic credential or approval token was emitted.
