# Source-declaration admission

`connectorgen declaration-admission` is the deterministic source-completeness
certificate. It reads two required repository-level catalogs and performs no
provider I/O:

- `internal/connectors/defs/declaration_admission_sources.json` is the
  independent source cohort and completeness denominator. Its nonzero expected
  connector and operation counts make a missing catalog, deleted row, or
  zero-work run fail.
- `internal/connectors/defs/declaration_admissions.json` contains the separate
  canonical declaration for each source identity. Its expected declaration
  count must match the rows present.

The checker does not scan connector-local sidecars. Adding a sidecar without
adding its operation to the required source cohort cannot make the global gate
pass or silently expand the certified cohort.

Each source row records a stable source ID, protocol, provider HTTPS URL,
exact document location, optional raw provider operation ID, method/base/path,
one binding identity, and the provider operation's `none`, `delete`, or
`destructive` semantic. A provider format may have no raw operation ID; the
stable source ID, exact citation, endpoint, and binding remain authoritative.
Source-row uniqueness is provenance-only: canonical URL, document location,
protocol, raw operation ID, and canonical provider endpoint/operation identity.
The authored URL must already equal its canonical form; admission never rewrites
evidence. Canonical citations use public HTTPS, lowercase unambiguous DNS hosts
without a trailing dot, no explicit default `:443`, normalized path escapes,
and stable ordering for bounded single-valued non-credential query keys. The
same canonicalizer and identity key are used by authoring admission and the
compact production ledger. Changing a URL spelling or runtime binding therefore
cannot disguise a duplicate provider row. Binding uniqueness is checked
separately; one command, stream, write action, or operation cannot claim two
source rows. Together those checks prevent GraphQL operations or other actions
sharing a transport endpoint from borrowing each other's implementation.

Each declaration references one source ID and repeats its exact binding and
canonical endpoint. It names exactly one lane (`etl`, `reverse_etl`,
`direct_read`, `direct_write`, `binary_download`, or `binary_upload`) and one
discoverable `cli_surface.json` command. Source-owned destructive semantics
determine whether `delete` or `destructive` metadata is required; surface or
declaration self-labeling cannot change that semantic.

An admitted `implemented` row must resolve through the engine's shared runtime
binding resolver, match the source binding and endpoint, retain the source
destructive semantic, and pass the real no-I/O `commandrunner.Preflight`. The
admission checker does not copy lane-specific runtime rules. Canonical provider
identity is kept separate from physical transport. A difference is accepted
only through a named, closed equivalence proof: a declared/configured base
path, placeholder-position identity, a registered hook's fixed transport,
GraphQL operation-to-`POST /graphql`, a declaration-owned query, a known
provider `.json` suffix, or an operation annotation. A method change or an
unrelated path is never replaced by the command surface.

An admitted `deferred` row instead names one missing implementation component
with closed evidence. Its discoverable command carries the same gap and an
exact source target: source ID, optional raw operation ID, binding identity,
destructive semantic, method, and path. Before returning typed
`system/missing_foundation`, runtime preflight verifies that identity against
the admitted source ledger and rejects stale, excluded, policy-only,
duplicated, or operation-swapped targets. Commandrunner first promotes the
exact row through its real implemented preflight; a runnable command is stale
and cannot reach `missing_foundation`. The compact source ledger is embedded in
`defs.FS` and inventoried as `runtime_declaration_target_ledger`, so this
remains true in the shipped binary even though the full `api_surface.json` is
intentionally not embedded. A complete connector may have zero runnable
operations when every source row is explicitly deferred.

A foundation gap names a missing implementation component such as
`typed_write_action`, `typed_record_schema`, `source_importer`, or
`runtime_executor`. Components are executor-specific: for example,
`typed_response_descriptor` applies only to REST or binary operations and
`source_importer` only to a missing direct-operation declaration. Provider
idempotency policy is not a missing executor. A method, lane, risk, approval policy,
`blocked_by_default`, retained artifact/hash, or live-certification state is
not a foundation component and cannot hide a source operation. Deferred state
also does not apply to an operation class: an existing implemented delete
remains implemented. GitHub `label delete` is the admission/runtime control.

Admission requires neither retained source bytes nor a hash, request body,
typed schema, credentials, provider response, or live proof. Declaration
completeness and runtime/live usability are distinct certificates:

```bash
go run ./cmd/connectorgen declaration-admission
go run ./cmd/connectorgen surface-sync --check
go test -timeout 20m -run '^TestEveryImplementedCommandPassesRuntimePreflight$' ./internal/connectors/commandrunner
go run ./cmd/connectorgen certification-matrix --check
```
