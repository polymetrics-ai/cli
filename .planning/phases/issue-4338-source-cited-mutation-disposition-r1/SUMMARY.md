---
coverage:
  - id: D1
    description: Per-operation source-cited non-executable mutation disposition
    verification:
      - kind: integration
        ref: "cmd/connectorgen/sourceprojection_test.go: Asana, Jira, Sentry, and Vercel behavioural projection/coverage fixtures"
        status: pass
    human_judgment: false
  - id: D2
    description: Closed mutation safety boundary
    verification:
      - kind: unit
        ref: "cmd/connectorgen/sourceprojection_test.go: complete-action, implemented-claim, read-only, and GraphQL controls"
        status: pass
    human_judgment: false
  - id: D3
    description: Existing connector projection remains unchanged
    verification:
      - kind: integration
        ref: "cmd/connectorgen/sourceprojection_test.go: byte-identical GitHub projection control"
        status: pass
    human_judgment: false
---

# Summary — issue 4338 source-cited mutation disposition

## Delivered

Added a connector-agnostic, per-source-operation disposition for a mutation
whose locked provider source exists but whose declaration has no complete
executable action. The disposition is a cited runtime gap, never an action,
contract, or CLI command.

The source-import path reads a connector-owned disposition document when one is
present, verifies the source ID, method, path, and provider citation, and emits
the gap in the source descriptor. Projection and executable-coverage validation
then accept the retained gap only while no complete action and no implemented
command claim exist.

## Safety properties

- A complete action is rejected from the disposition path and remains
  `implemented`.
- An `implemented` command claim, even when its action is incomplete, is
  rejected rather than being silently downgraded.
- A read-only foundation cannot satisfy a mutating POST, PUT, PATCH, DELETE, or
  GraphQL mutation.
- An HTTP POST GraphQL query cannot be treated as a mutation just because its
  transport method is POST.
- No connector definition was changed; the model contains no connector-name or
  volume-based exception.

## Consumer evidence

Behavioural source-projection and executable-coverage tests separately cover:

- Asana absent-action mutation;
- Jira incomplete action-contract mutation;
- Sentry SCIM PATCH and dashboard POST mutations; and
- a 159-operation Vercel-sized batch spanning bulk redirect PATCH, bulk restore
  POST, nested file-write POST, and destructive-cache POST forms.

The Vercel handoff is not a connector-wide waiver: batch 1 must author a
complete action for every Vercel operation intended to execute. This foundation
can record only an individually cited operation that batch 1 intentionally
leaves non-executable.

## Validation

See `VERIFICATION.md` and `TDD-LEDGER.md` for red/green and full command
evidence. The inline GSD fallback is recorded because the direct-PR adapter
could not supply the isolated Pi worker runtime.
