# Native Implementation Status Policy

Status: required policy for connector work

## Problem

The current generic Go binding made every catalog slug runnable in fixture mode, but the status label made it look like every connector was fully live-native. That is misleading.

This policy prevents that from happening again.

## Definitions

| Term | Definition |
| --- | --- |
| Catalog connector | Connector metadata imported from the public connector catalog. |
| Fixture binding | Go code that can run safe local fixture operations without touching the external system. |
| Runtime family port | Shared Go runtime that can execute a class of upstream connectors. |
| Live connector port | Connector-specific native Go implementation that talks to the real external system. |
| Enabled connector | Live connector port that passed conformance and is safe to use by default. |

## Status Values

Use these exact values in future code:

- `catalog_imported`
- `native_fixture_binding`
- `runtime_family_ported`
- `connector_live_ported`
- `live_conformance_passed`
- `enabled`
- `unsupported_deprecated`

## Capability Rules

`metadata=true` may be set for all imported catalog entries.

`check`, `catalog`, `read`, `write`, `query`, `etl`, and `reverse_etl` must include a mode:

- `fixture`
- `live`
- `unsupported`

Example:

```json
{
  "read": {
    "fixture": true,
    "live": false,
    "unsupported_reason": "Declarative HTTP runtime has not imported this connector manifest yet."
  }
}
```

Do not expose a boolean-only capability for new native port work, because it hides the difference between fixture support and live support.

## Enablement Rule

A connector can be marked `enabled` only when all required live operations pass:

- source: live `check`, live `catalog`, live `read` for at least one representative stream, state checkpoint/resume, docs, redaction.
- destination: live `check`, live `ValidateWrite`, live `DryRunWrite` when supported, live `Write`, receipt, idempotency, docs, redaction.
- query connector: live safe query validation and execution.
- CDC connector: setup validation, CDC event capture, resume, delete semantics, and retention warning tests.

## CLI Display Rule

`pm connectors inspect <slug>` must show:

- status stage.
- fixture capabilities.
- live capabilities.
- missing live gates.
- upstream implementation source.
- next implementation task.

Agents must treat `native_fixture_binding` as non-live. They can use it for docs, schema planning, fixture tests, and generated code validation only.

## Migration From Current State

Next implementation phase must:

1. Rename current generic catalog connector status from `enabled` to `native_fixture_binding`.
2. Preserve fixture commands.
3. Add live status fields.
4. Keep `github` as `connector_live_ported` or `live_conformance_passed` depending on full stream parity.
5. Mark only connectors with real live code and tests as `enabled`.

## Test Requirements

Add tests that fail if:

- a fixture-only connector is marked `enabled`.
- a connector has live capability without live conformance evidence.
- a secret field appears in manual, JSON, logs, errors, previews, skills, or benchmark files.
- a reverse ETL live write bypasses plan/preview/approval.
