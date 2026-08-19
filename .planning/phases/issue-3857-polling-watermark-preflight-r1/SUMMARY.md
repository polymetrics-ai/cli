---
coverage:
  - id: D1
    description: Closed polling watermark and target-apply declarations load defensively from their separate optional bundle file.
    verification:
      - kind: unit
        ref: internal/connectors/engine/polling_definition_test.go
        status: pass
    human_judgment: false
  - id: D2
    description: The runtime preflight resolves only exact registered executors with immutable corpus evidence and refuses unsafe declarations before source I/O.
    verification:
      - kind: unit
        ref: internal/connectors/engine/polling_preflight_test.go
        status: pass
    human_judgment: false
  - id: D3
    description: Every declared polling mode receives eligibility only through the real preflight gate.
    verification:
      - kind: unit
        ref: "internal/connectors/engine/polling_preflight_test.go: TestPollingModeEligibilitySweepsEveryImplementedPollingModeThroughRuntimePreflight"
        status: pass
    human_judgment: false
---

# Summary — #3857 polling-watermark preflight

The shared connector engine now accepts an optional, separate
`polling_watermark.json` declaration. Its descriptor is closed and data-only;
it cannot turn a polling scan into CDC or invent a REST command.

`PollingPreflight` validates the implementation declaration, catalog object,
registered source/apply executors, immutable #3856 conformance evidence, and
mode-to-apply strategy before returning an immutable resolution. It has no
read or DML method, so #3858/#3859 remain the owners of source execution and
target apply. `PollingModeEligibilityOf` invokes that same runtime gate per
declared mode rather than reproducing its rules.

Happy, sad, and corpus-derived edge tests use individually justified guarded
fakes: the issue forbids live database calls and owns neither a driver nor the
future executors. The tests assert source-read, target-prepare, emitted-record,
and empty-page counters, as well as exact refusal strings; they cannot pass on
a no-op. The immutable corpus itself also runs unfiltered.

No connector-specific bundle was declared, and `defs.FS` remains unchanged
until an engine-specific lane owns both a real declaration and its embed glob.
The existing changefeed/CDC and commandrunner REST paths were not changed.
