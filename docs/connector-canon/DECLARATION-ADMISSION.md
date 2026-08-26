# Source-declaration admission

`connectorgen declaration-admission` is the deterministic source-completeness
certificate. It reads three required repository-level artifacts and performs
no provider I/O:

- `internal/connectors/defs/declaration_admission_inventory.json` is the
  independently reviewed completeness denominator. Each entry selects one
  exact operation ID from a connector-owned
  `<connector>/sources/*-operation-source-lock.json` and assigns its compact
  source ID. Deleting mutable rows or changing adjacent counts cannot shrink
  this inventory; the v2 catalogs do not accept expected-count fields.
- `internal/connectors/defs/declaration_admission_sources.json` is the compact
  source cohort used by authoring admission and shipped deferred-target
  preflight. Every row must equal the selected lock operation's provider URL,
  exact document location, protocol, provider operation ID, method, and path.
- `internal/connectors/defs/declaration_admissions.json` contains the separate
  canonical declaration for each inventoried source identity.

The checker does not scan connector-local sidecars. Adding a sidecar without
adding its operation to the required source cohort cannot make the global gate
pass or silently expand the certified cohort.

The checker resolves only connector-owned reviewed lock files and reads no
provider or retained artifact bytes. A path outside the selected connector's
`sources/` directory, a nonexistent locked operation, an unrelated URL, or a
semantic location/endpoint alias fails closed. The reviewed lock already owns
the retained content identity; declaration admission neither fetches that
content nor recomputes its byte count or hash. Source retention verification
remains a separate certificate.

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
no dot segments, and stable ordering for bounded single-valued non-credential
query keys. The
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

Public connector dispatch validates required, unknown, typed, enum, bounded,
and `env_only` inputs after help handling and before opening the app or resolving
a credential. A valid invocation still reaches the existing credential-bound
runtime boundary; admission and credential-free input validation do not make a
runtime or live usability claim.

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

`unsupported_with_provider_evidence` is a third, non-executable disposition.
It is valid in every admission lane when the provider documents the operation
but the CLI cannot support its semantics. The operation remains in the
inventory and command discovery surface with its exact source target and a
bounded reason, while declaring no stream, write, operation executor, or
missing foundation. Commandrunner returns typed
`system/provider_evidenced_unsupported`; it cannot be counted as implemented
or deferred runtime usability.

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

Admission requires a reviewed connector-owned source-lock operation selection,
but it does not fetch/read retained source bytes, recompute a hash, infer a
request body or typed schema, resolve credentials, inspect a provider response,
or claim live proof. Declaration completeness and runtime/live usability are
distinct certificates:

```bash
go run ./cmd/connectorgen declaration-admission
go run ./cmd/connectorgen surface-sync --check
go test -timeout 20m -run '^TestEveryImplementedCommandPassesRuntimePreflight$' ./internal/connectors/commandrunner
go run ./cmd/connectorgen certification-matrix --check
```
