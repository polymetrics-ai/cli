# Sub-issue #3684 — `batchable` declaration on write actions

> **Filed as [#3684](https://github.com/polymetrics-ai/cli/issues/3684), sub-issue of
> [#3682](https://github.com/polymetrics-ai/cli/issues/3682).** Created under the captain's bounded
> identity exception for issue creation only.

**Title:** `feat(engine): add batchable declaration to write actions`

---

## Objective

Give the engine the vocabulary to express "this write action must never be bulk-automated", as a
single typed boolean on a write action, defaulting to permissive.

## Operations unblocked

**None directly.** This slice ships vocabulary, not enforcement. It is the dependency that both the
enforcement (#3689) slice and the eventual `reddit vote` operation (`reddit` connector,
1 operation) require.

Declaring the field without enforcing it would be worse than nothing — that is precisely the
declared-implemented-but-fails-at-runtime defect class tonight's audit found 174 instances of. So
this slice does not ship alone; it lands in the same PR as enforcement.

## Parent

- Parent issue: #3682
- Branch: `fm/cli-engine-nonbatchable-write-r1`

## Scope

Allowed write scope:

- `internal/connectors/engine/schema/writes.schema.json` — the `batchable` field
- `internal/connectors/engine/bundle.go` — `WriteAction.Batchable` + accessor
- `internal/connectors/engine/connector.go` — propagate to manifest and definition
- `internal/connectors/manifest.go`, `internal/connectors/definition.go` — surface the field
- `internal/connectors/guide.go` — help/manual line for non-batchable actions
- `docs/**`, `website/**` regeneration
- tests under those packages

Do **not** edit any connector bundle under `internal/connectors/defs/**` or any native connector
under `internal/connectors/native/**`.

## Design

`"batchable"` is a boolean on each entry of `writes.json`'s `actions` array, `"default": true` in
the JSON schema. Absent means batchable, which is what all ~existing actions are today, so no bundle
changes and no behavior changes.

Go representation is `Batchable *bool` with `json:"batchable,omitempty"`, plus:

```go
func (a WriteAction) IsBatchable() bool { return a.Batchable == nil || *a.Batchable }
```

A pointer rather than a plain `bool` because Go's zero value for `bool` is `false`, and `false` here
means *non-batchable*. A plain `bool` would make every hand-constructed `WriteAction{...}` literal
and every native connector's `WriteActionSpec{...}` literal silently non-batchable, inverting the
intended default at exactly the sites that never opt in. `nil` → batchable keeps the safe default
for JSON-absent, Go-zero, and test-constructed values alike.

The same `*bool` + `IsBatchable()` pair is mirrored on `connectors.WriteActionSpec` (manifest) and
`connectors.WriteActionInfo` (definition), so consumers that only hold a manifest — which is how the
app resolves write actions today, via `connectors.ManifestOf` — can read it without reaching back
into the engine.

## Verification

- Loader accepts `batchable: true`, `batchable: false`, and an absent field; rejects a non-boolean.
- `IsBatchable()` returns true for absent, true for explicit true, false for explicit false, and
  true for the Go zero value.
- Every existing bundle still loads and every existing action reports batchable.
- `pm connectors inspect <name> --json` and `pm help <connector>` are unchanged for existing
  connectors, and show the non-batchable line for a fixture that declares it.

## Required skills

- `golang-how-to`, `golang-structs-interfaces`, `golang-testing`, `golang-documentation`,
  `golang-safety`
