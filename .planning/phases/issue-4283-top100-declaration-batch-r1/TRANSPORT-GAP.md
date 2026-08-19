# Sync transport disposition — increment 001

## 2026-08-19 complete-map supersession: definition-owned contract

The earlier runtime evidence remains historical context, but the complete map
is now assessed against the definition-owned contract published by PR #4286 in
`docs/sync-transport-definition.md`, not the former GitHub-shaped descriptor.
For each selected connector, the map carries two distinct recoverable
declaration-pending records:
`sync-transport-source-definition-<connector>` and
`sync-transport-destination-definition-<connector>`.

- Source evidence: `docs/sync-transport-definition.md:15-38` requires an
  exact executor, positive eligible-stream allowlist, modes, delivery facts,
  and connector-owned conformance evidence. None of the ten bundles has a
  `sync_transport.json` declaring those facts.
- Destination evidence: `docs/sync-transport-definition.md:39-75` additionally
  requires eligible typed actions, durable acknowledgement, one closed
  apply-strategy per mode, and connector-owned conformance evidence. None is
  present in the ten bundles.
- Smallest recovery: declare each direction only after its actual executor,
  evidence, stream/action allowlists, delivery guarantees, typed bindings and
  acknowledgement exist. The source documents must never be used to infer any
  of those runtime facts.

The exact source-operation dispositions and the full six-class summary are
recorded in `COMPLETE-PARITY-MAP.md` and in each connector's
`sources/<connector>-declaration-disposition.json`.

## Historical verdict: path (b), superseded by declaration-pending

No selected connector receives `sync_transport.json` in this increment. That omission is an explicit, recoverable declaration rather than a missing record: each of Docker Hub, GitLab, Jira, Vercel, Notion, Stripe, Bitbucket, CircleCI, Sentry, and Asana now has a `sync_transport` disposition with `reason: declaration-pending` and `recoverable: true` in `REJECTION-LIST.json`. The preceding foundation-gap label is historical and is not a foundation-lane request.

The pinned provider descriptions define HTTP resources, not a safe warehouse transport contract. They do not establish a source delivery guarantee, source-to-destination field mapping, destination acknowledgement, or the typed apply action for every sync mode. Those facts must not be inferred.

## Runtime evidence

- `internal/synctransport/definition_composition.go:145-168` refuses any source or destination whose exact executor factory is not registered or whose conformance evidence is not accepted.
- `internal/app/issue_label_warehouse_transport.go:54-103` supplies the only declarative API source factory and accepts the exact `declarative_stream_source` evidence constant from lines 19-24. Reusing that GitHub-specific evidence for another connector would be a false certification claim.
- `internal/app/issue_label_warehouse_transport.go:322-368` makes the only declarative API destination a closed GitHub issue-label contract: it requires an `issues` stream and apply, replace, and cleanup actions with precise bindings. The ten selected `writes.json` files have zero `transport_binding` actions.
- `internal/connectors/certify/stages_transport_internal_test.go:89` proves the runtime fails closed before provider I/O when a declared source executor is not registered.

## Smallest safe recovery

Extend #4093 with a connector-neutral source `DefinitionFactory` that carries per-bundle (not GitHub-reused) conformance evidence, then add a closed typed destination `DefinitionFactory` for a named, connector-owned action with explicit source bindings, acknowledgement, and per-mode apply strategies. The destination must remain a typed action adapter; a generic HTTP write executor is prohibited. Once one exists, its `sync_transport.json` can be derived from that action and the pinned source streams, validated, and certified independently.
