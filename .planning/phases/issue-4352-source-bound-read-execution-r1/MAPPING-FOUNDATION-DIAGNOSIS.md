# Mapping-foundation diagnosis — Asana source-cited mutations

Recorded 2026-08-26 under Captain authorization (`012.msg`).

## User-visible contract

Every documented, locked Asana provider operation must remain discoverable and
source-mapped. A source operation may be execution-deferred only when its
provider contract requires a named missing foundation; it must not disappear,
be silently treated as complete, or be turned into an invented executor.

## Symptom and trigger

`go run ./cmd/connectorgen validate internal/connectors/defs` reports 90
Asana source-projection findings:

- 25 exact source operations have no declaration-owned action (23 `DELETE`
  routes plus the two access-request decision `POST`s).
- 65 exact source operations retain an `implemented` reverse-ETL command but
  their pinned provider request needs `cli-request-schema-foundation-r1` for a
  dynamic/complex body or non-scalar query serialization.

The complete operation/route inventory is
`MUTATION-GAP-INVENTORY.md` in this phase directory.

## Masking condition and boundary

The current `sourceNonExecutableMutationDisposition` correctly represents an
absent, non-executable operation. It deliberately rejects an operation whose
matching CLI route is already `implemented`, even when that operation's
source-request contract is incomplete. The aggregate reverse-ETL availability
claim therefore masks the operation-granular missing foundation and makes the
validator report the 65 source gaps.

The smallest counterfactual is the existing fixture test
`TestSourceProjectionSourceCitedNonExecutableMutationDispositionRejectsImplementedIncompleteActionClaim`:
the same exact mutation citation is accepted when no command claims
implementation, but is rejected solely once the matching command is marked
`implemented`. The guard is
`sourceProjectionMutationClaimsImplementedAction` in
`sourceProjectionApplyNonExecutableMutationDispositions`
(`cmd/connectorgen/sourceprojection.go`).

For comparison, the fully covered `asana.rest.getAccessRequests` read retains
an exact source operation, `GET /access_requests`, required typed input, and
the normal credential boundary; it needs neither a deferred disposition nor a
route escape hatch. The desired mutation representation must have the same
source identity/citation fidelity while truthfully recording its narrower
typed-contract foundation.

## Repair boundary

Add an operation-granular source-cited *partial coverage* disposition for only
an existing, implemented mutation action whose provider contract is incomplete.
It must not apply to a source-complete action, an absent action, a read, or an
unsupported provider operation. The existing non-executable disposition remains
the representation for the 25 absent actions. Neither representation changes
the command/action files or makes provider I/O.
