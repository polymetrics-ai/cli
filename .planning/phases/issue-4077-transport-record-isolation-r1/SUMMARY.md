---
phase: issue-4077-transport-record-isolation-r1
plan: 01
subsystem: transport
tags: [synctransport, record-isolation, json-raw-message, tdd]
requires:
  - phase: transport-parent-3862
    provides: closed source/destination dispatch seam
provides:
  - source and downstream records no longer share RawMessage or string-map storage
  - unrecognized record values fail before a transport boundary
affects: [transport, warehouse-stage, destination-executor]
tech-stack:
  added: []
  patterns: [closed explicit record-value cloning with contextual boundary errors]
key-files:
  created: [.planning/phases/issue-4077-transport-record-isolation-r1/SUMMARY.md]
  modified: [internal/synctransport/types.go, internal/synctransport/orchestrator.go, internal/synctransport/transport_test.go]
key-decisions:
  - "Clone json.RawMessage before []byte and map[string]string explicitly."
  - "Reject every unrecognized record value instead of silently forwarding a possible alias."
requirements-completed: []
coverage:
  - id: D1
    description: RawMessage and string-map source storage stays independent across both downstream boundaries.
    verification:
      - kind: unit
        ref: internal/synctransport/transport_test.go focused mutation regressions
        status: pass
      - kind: other
        ref: no-mistakes targeted transport behavior transcript
        status: pass
    human_judgment: false
  - id: D2
    description: Unsupported mutable record values fail before crossing a boundary.
    verification:
      - kind: unit
        ref: internal/synctransport/transport_test.go unsupported-value regressions
        status: pass
    human_judgment: false
completed: 2026-08-13
---

# #4077 — Transport record-isolation correction summary

**Closed Transport record copies now isolate `json.RawMessage` and `map[string]string`, while rejecting
unrecognized values before staging or destination application.**

## TDD record

1. **Plan:** `a34ac4bb1` — established the GSD/TDD lifecycle and exact accepted-parent reproduction.
2. **RED:** `8b0c2cc57` — committed the failing aliasing and unsafe-forwarding regression.
3. **GREEN:** `3b250b874` — added explicit copies and contextual fail-closed propagation.

## Verification

- Focused normal/race regressions, full normal/race Transport tests, and the application’s canonical
  all-seven-mode selector passed.
- `go vet ./...`, real `pm` build, GSD evidence gate, and split repository gates passed except for the
  documented pre-existing generated-worktree inventory failures.
- no-mistakes run `01KZWMAV3JEKZ9GFK5REF0K2RV` passed with no findings or correction loops.

## Scope and delivery

No connector adapter, credentials, polling, registry, public CLI, warehouse payload format, or mode
semantics changed. The local no-mistakes path intentionally skipped push/PR/CI; a human-selected
stacked-draft delivery exception is required before any remote action.

## Deviations from plan

None. The documented inline/manual GSD fallback was used because the named issue phase is absent from the
numeric roadmap and compatible lifecycle role spawning is forbidden by the repository contract.
