# ADR — Flow Engine (Phase 0)

## ADR-001: JSON as the canonical manifest format (YAML as authoring sugar)

**Status:** Accepted

**Context:** Flow manifests need to be both human-editable and machine-parseable. Go's stdlib
has `encoding/json` but no YAML parser. The project's go.mod does not include a YAML library.

**Decision:** Manifests are stored and validated as JSON. YAML is not supported in Phase 0.
Users author manifests as JSON directly. If YAML support is required later, it is added only
after confirming a YAML library already exists in go.mod (e.g., pulled in by another dep), or
by implementing a minimal YAML-to-JSON pre-processor. A new dependency will not be added.

**Consequences:** Slightly more verbose authoring experience in Phase 0. Eliminates any
dependency risk. YAML support is a backwards-compatible future add.

---

## ADR-002: File-backed checkpoint store (not reusing internal/state.JSONStore)

**Status:** Accepted

**Context:** `internal/state.JSONStore` is a generic key-value store, but it is tightly coupled
to the `app.App` state struct and path conventions. Importing it into `internal/flow` would
create a coupling that makes `internal/flow` harder to test independently.

**Decision:** Implement `FileCheckpointStore` in `internal/flow/checkpoint.go` using the same
pattern (atomic write via rename) but as a standalone implementation. The interface
`CheckpointStore` is defined in the same package so tests can stub it.

**Consequences:** Small amount of duplicated logic (read/write JSON file). The trade-off of
testability and decoupling is worth it. If a shared utility emerges, it can be extracted later.

---

## ADR-003: FileLock for lease (not runtime.Module.Leases)

**Status:** Accepted

**Context:** `internal/runtime.Module` has a `Leases` interface backed optionally by
DragonflyDB. Phase 0 must work dependency-free.

**Decision:** Use `internal/state.FileLock` directly in the engine. Add stale-lock recovery
(PID check). When `internal/runtimecheck` reports Dragonfly is available, the engine can be
upgraded to use `runtime.Module.Leases` in a future phase without changing the interface.

**Consequences:** File locks do not survive NFS/shared filesystems reliably. Acceptable for a
local-first tool. Documented in RUNBOOK.md.

---

## ADR-004: AppAdapter interface in internal/flow

**Status:** Accepted

**Context:** `internal/flow/engine.go` needs to call `app.App.ETLRun` and `app.App.QuerySQL`,
but importing `*app.App` directly would make unit tests require a fully initialized App (file
system, vault, connections, etc.).

**Decision:** Define `AppAdapter` interface in `internal/flow` with exactly the two methods
needed. Production code wires a thin adapter struct. Tests use `stubAppAdapter`.

**Consequences:** One extra indirection. The interface is small (2 methods) so it remains easy
to maintain.

---

## ADR-005: Sequential execution only in Phase 0

**Status:** Accepted

**Context:** Steps with no DAG dependency between them could execute concurrently for
performance. However, concurrent execution requires bounded worker pools, context propagation,
and more complex error handling.

**Decision:** Phase 0 executes steps sequentially in topological order. Parallelism is
explicitly deferred to a later iteration when there is a demonstrated need and benchmark data.

**Consequences:** No performance regression since the feature is new. Parallelism can be added
as an opt-in `--parallel` flag in a future phase.

---

## Design direction note

Not applicable — this is a CLI/backend phase with no visual UI.
