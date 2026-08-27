# Issue #4359 — TDD ledger

## Red

`go test -timeout 20m ./internal/connectors/engine -run '^TestResolveImplementedCommandBindingProvesEveryCircleCICompositeProjectSlugBinding$' -count=1`

Failed at compile time as intended before the foundation existed: the test named the
missing `CommandEndpointCompositeProviderPathIdentity` proof and the absent
`APISurface.CompositeProviderPathIdentity` declaration/configuration types. The
fixture enumerates all eleven CircleCI source bindings and their existing runtime
paths; it does not import Batch 1's `cli_surface.json`.

The red fixture exercised `ResolveImplementedCommandBinding` with the actual
CircleCI source identity and existing runtime templates. It reproduced the
eleven binding failures against the immutable Batch-1 declaration shape before
production implementation existed.

## Green

`go test -timeout 20m ./internal/connectors/engine -run 'CircleCICompositeProviderPathIdentity|TestResolveImplementedCommandBinding(ProvesEveryCircleCICompositeProjectSlugBinding|RejectsUnprovenCircleCICompositeProjectSlugSubstitutions)$' -count=1`

Passed. The closed manifest admits exactly the six ETL stream and five
reverse-ETL write bindings, emits `composite_provider_path_identity`, and is
loaded from the embedded, source-cited CircleCI
`composite_provider_path_identity.json`. The negative
matrix refuses missing/reordered/partial/extra configuration, unretained source
identity, reordered declarations, partial/reordered/extra/absolute/query/route
transport, method drift, cross-connector configuration, and direct-read,
direct-write, binary-download, and binary-upload lanes.

## Refactor / review

Pending final independent review. The implementation remains a closed
equivalence proof; no request interpolation, execution, parser, method, body,
or credential logic changed.
