# Context — Issue #4086 PostgreSQL fence split

## Locked delivery decision

This is a mechanical file-layout change from comparison base
`integration/4015-mvp-flat-r1` at `5a457970b3bc15343e5ba6b7b4acf48994b63add`.
No behavior, generated capability output, public connector surface, declaration
signature, or test semantics/assertions change. File splitting may change only
Go test registration order; that difference is accepted as inert. The proof
uses Go's shuffle facility (`-shuffle=on`, reproduced deterministically as
`-shuffle=N`) at the comparison base and original move head
`6e31ac1abfc4a46fd1dbbef3ec54086da85b682e`, for both affected packages and
seeds `408601` and `408602`. It must not be normalized through order-prefixed
files or declaration reordering.

## Registration-order proof

Each numeric `-shuffle=N` run is Go's reproducible shuffled mode. The recorded
commands were `go test -count=1 -timeout 20m -shuffle=<seed> <package>`.

| Revision | Seed | `./internal/connectors/database` | `./internal/connectors/native/postgres` |
| --- | --- | --- | --- |
| Base `5a457970b3bc15343e5ba6b7b4acf48994b63add` | `408601` | `ok 6.452s` | `ok 0.881s` |
| Base `5a457970b3bc15343e5ba6b7b4acf48994b63add` | `408602` | `ok 6.229s` | `ok 0.878s` |
| Original move head `6e31ac1abfc4a46fd1dbbef3ec54086da85b682e` | `408601` | `ok 6.416s` | `ok 0.875s` |
| Original move head `6e31ac1abfc4a46fd1dbbef3ec54086da85b682e` | `408602` | `ok 6.652s` | `ok 0.870s` |

## Ownership after the split

| Future lane | Owned test file(s) |
| --- | --- |
| Source | `internal/connectors/native/postgres/source_test.go`; `internal/connectors/database/source_read_plan_test.go` |
| Target | `internal/connectors/database/target_admission_test.go` |
| Mapping | `internal/connectors/database/mapping_definition_test.go` |
| CDC | `internal/connectors/native/postgres/cdc_capability_fence_test.go` |

`internal/connectors/native/postgres/capability_surface_test.go` is the stable
connector capability fence. It is deliberately not assigned to an execution
lane. Shared test-only scaffolding is placed in
`internal/connectors/database/test_helpers_test.go` so no production declaration
or connector-specific literal moves into the connector-neutral database package.

## Scope fence

- Move existing declarations only, within their present Go package.
- Keep PostgreSQL-specific definition literals under `internal/connectors/defs/postgres/`.
- Do not modify generated artifacts or run a generator that would alter them.
- The required command/docs parity is comparison-only: output must be identical,
  so no help/manual/website source is expected to change.

## GSD execution note

Issue #4086 is not a numbered `.planning/ROADMAP.md` phase. The required
`discuss-phase → plan-phase --tdd → execute-phase → verify-work → code-review`
path was resolved with `scripts/gsd`; it is executed inline because this task
forbids role spawning and the issue is not initializable by the numbered phase
workflow. This directory is the manual GSD record.
