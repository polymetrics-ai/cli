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

## Current-head CI generated-subject recovery

`gh-axi run view 33042859184 --job 98420150790 --log-failed`

Red: current-head CI completed all earlier Verify work, then correctly stopped
at `go run ./cmd/connectorgen certification-subject --check`: the committed
subject's declarations fingerprint predated the new embedded CircleCI
declaration. This is generated-evidence drift, not an engine or route failure.

Green: `go run ./cmd/connectorgen certification-subject` refreshed only
`internal/connectors/certifications/current-subject.json`, changing its
declarations SHA-256 and derived subject fingerprint. The strict subject check
and the certification matrix, candidates, and sweep checks pass afterwards.
No source lock, provider identity, executable command, route, credential, or
six-lane eligibility changed.

## Current-main integration recovery

`git merge --no-ff origin/main -m 'merge: integrate origin/main into composite provider path foundation'`

Red: after `origin/main` advanced to
`2165619ec8f5f9d4141b491b7a5a64bc460d0c71` (#4356), its source-bound-read
foundation changed declaration, source-projection, and CLI-mapping inputs. The
normal merge correctly conflicted only in the generator-owned certification
subject; neither the declaration-owned CircleCI proof nor any CircleCI command
surface conflicted.

Green: the canonical `go run ./cmd/connectorgen certification-subject`
generator resolved that sole artifact, retaining main's new source-projection
and CLI-mapping hashes and deriving the combined declaration SHA-256 and
fingerprint. Merge commit `4d7f57894` contains the unmodified closed CircleCI
proof plus current main. Post-merge full engine, generator, definitions,
commandrunner, App, and CLI suites passed; definition validation, surface sync,
declaration admission, operation evidence, certification projections, Asana
source import, docs, tidy, lint, smoke, contract, GitHub-artifact, connector
canon, and release workflow gates passed. The built binary still returns
`unknown command "circleci"` (exit 2), proving no Batch-1 CLI surface was
invented; the focused eleven-row closed-proof test matrix remains green.
