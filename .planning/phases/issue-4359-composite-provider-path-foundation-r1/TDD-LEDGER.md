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

## Refactor / CI boundary repair

`go test -timeout 20m ./internal/connectors/engine -run '^TestCompositeProviderPathIdentityOwnsItsConnectorIdentity$' -count=1`

Red: the new declaration-ownership assertion failed to compile because the
identity type did not expose an owning connector. This protected the CI repair
from retaining any provider-specific allow-list in shared engine code.

Green: `CompositeProviderPathIdentity` now reads its connector, citation,
placeholder, ordered components, and source rows only from the declaration.
The provider-neutral engine validates the closed row shape and derives only
the declared inverse transport. It admits no direct/binary lane, hook, route,
base override, query, suffix, annotation, method, binding, or literal-path
variation. `go run ./cmd/connectorgen boundary . --json` reports clean.

No request interpolation, execution, parser, method, body, credential, or
source-lock logic changed. A fresh exact-head independent review remains
required after this repair commit.
