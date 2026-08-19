# Sync transport disposition — increment 001

## Verdict: path (b), foundation gap

No selected connector receives `sync_transport.json` in this increment. That omission is an explicit, recoverable declaration rather than a missing record: each of Docker Hub, GitLab, Jira, Vercel, Notion, Stripe, Bitbucket, CircleCI, Sentry, and Asana has a `sync_transport` rejection with `reason: foundation-gap` and `recoverable: true` in `REJECTION-LIST.json`.

The pinned provider descriptions define HTTP resources, not a safe warehouse transport contract. They do not establish a source delivery guarantee, source-to-destination field mapping, destination acknowledgement, or the typed apply action for every sync mode. Those facts must not be inferred.

## Runtime evidence

- `internal/synctransport/definition_composition.go:145-168` refuses any source or destination whose exact executor factory is not registered or whose conformance evidence is not accepted.
- `internal/app/issue_label_warehouse_transport.go:54-103` supplies the only declarative API source factory and accepts the exact `declarative_stream_source` evidence constant from lines 19-24. Reusing that GitHub-specific evidence for another connector would be a false certification claim.
- `internal/app/issue_label_warehouse_transport.go:322-368` makes the only declarative API destination a closed GitHub issue-label contract: it requires an `issues` stream and apply, replace, and cleanup actions with precise bindings. The ten selected `writes.json` files have zero `transport_binding` actions.
- `internal/connectors/certify/stages_transport_internal_test.go:89` proves the runtime fails closed before provider I/O when a declared source executor is not registered.

## Smallest safe recovery

Extend #4093 with a connector-neutral source `DefinitionFactory` that carries per-bundle (not GitHub-reused) conformance evidence, then add a closed typed destination `DefinitionFactory` for a named, connector-owned action with explicit source bindings, acknowledgement, and per-mode apply strategies. The destination must remain a typed action adapter; a generic HTTP write executor is prohibited. Once one exists, its `sync_transport.json` can be derived from that action and the pinned source streams, validated, and certified independently.
