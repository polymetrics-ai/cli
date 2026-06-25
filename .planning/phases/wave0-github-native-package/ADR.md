# ADR — Per-System Connector Packages + Self-Registration

## Status
Accepted (Wave 0).

## Context
The platform targets 626 unified connectors, each a real Go package named by bare system slug
(`github`, `postgres`). The current flat `package connectors` + hand-switch `NewRegistry()` does
not scale: every new connector would edit a central switch and pollute one package's namespace.

## Decision
1. **One package per system** under `internal/connectors/<bare-name>/`, importing the shared
   `connectors` types/interface and the `connsdk` toolkit.
2. **Self-registration**: `connectors.RegisterFactory(name, func() Connector)` called from each
   connector package's `init()`. A generated `registry_gen.go` blank-imports every connector
   package; `NewRegistry()` consumes the registered factories. Adding a connector = add one blank
   import line (later automated by codegen), never a switch edit.
3. GitHub is migrated first as the **reference** package.

## Consequences
- (+) Scales to hundreds of packages; clean namespaces; copy-the-template authoring.
- (+) `NewRegistry()` becomes data-driven; built-ins (Sample/File/Warehouse/Outbox) stay explicit.
- (−) Compile time/binary size grow with package count — mitigated later with family build tags.
- Import-cycle avoided: `connectors` only blank-imports connector packages (init side-effects),
  connector packages depend on `connectors`; `connsdk` is a leaf.

## Alternatives rejected
- Keep flat package (namespace collisions at scale).
- Code-generate one giant switch (still central churn; worse diffs).
