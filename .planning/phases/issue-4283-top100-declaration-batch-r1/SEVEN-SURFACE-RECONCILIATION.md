# Seven-surface reconciliation — PR #4294

This is the authoritative ten-connector reconciliation for the relaunch on
2026-08-20. The machine-readable counts, typed-action set hashes, fixtures and
exact source mappings are in `SEVEN-SURFACE-RECONCILIATION.json`.

| Connector | Documented operations | CLI implemented | Typed actions | Static destination proof |
|---|---:|---:|---:|---|
| Docker Hub | 54 | 45 | 20 | None: no fixture-proven target identity mapping |
| Notion | 49 | 45 | 24 | `views` → `update_view` |
| Stripe | 589 | 8 | 3 | `customers` → `update_customer` |
| Bitbucket | 331 | 3 | 54 | None: fixture lacks required target workspace identity |
| GitLab | 1,755 | 4 | 0 | Not applicable |
| CircleCI | 111 | 2 | 7 | `schedules` → `update_schedule` |
| Sentry | 223 | 0 | 0 | Not applicable |
| Vercel | 400 | 2 | 18 | `projects` → `update_project` |
| Asana | 249 | 82 | 73 | None: closed mapper cannot construct nested action data |
| Jira | 617 | 584 | 292 | None: representative typed input violates mapper identifier grammar |

The JSON ledger now contains one named eligibility disposition for every one
of the 491 typed actions: four are `eligible_bound_fixture_mapping` and 487
are eligible but pending the exact source-identity, nested-object, or
action-scoped mapping dependency named alongside that action. The action-set
hash remains an anti-drift selector for each connector. These actions are not
excluded for risk, privilege, destructiveness, or the lack of live credentials.

The current generic destination can model one action mapping per destination
declaration. It resolves that mapping from source executor and stream, not the
selected action. `action-scoped-source-binding` is therefore the precise
remaining foundation dependency for multi-action destination coverage. The
application/CLI dispatch integration remains upstream in #4304; no row in this
ledger claims provider-live reverse-ETL deployment.

## Documented-operation command reachability boundary

The source crosswalk-to-current-surface join records 3,366 of 4,378 documented
operations without a declared command binding: Docker Hub 11, Notion 5, Stripe
581, Bitbucket 134, GitLab 1,751, CircleCI 95, Sentry 220, Vercel 378, Asana
164, and Jira 27. This count excludes the documented surface-only variants
that are not in each pinned provider source denominator.

Captain rejected a disabled-command placeholder: an endpoint is reachable only
when the installed CLI executes its fixed provider operation with typed inputs
and a credential. `EXECUTABLE-OPERATION-CAPABILITY-AUDIT.json` reclassifies all
3,366 rows by their next executable capability: 1,389 fixed REST reads, 1,828
fixed REST writes, 120 bounded binary transfers, 10 status-kind registrations,
and 19 provider contracts with no bounded typed payload. The 224 typed actions
with nested schemas are a downstream #4305 structured-body consumer.

`EXECUTABLE-OPERATION-FOUNDATION-DESIGN.md` defines the closed implementation
slices: hash-matched source import; fixed REST contract materialization;
typed declared headers; #4305 structured bodies; bounded binary/multipart,
status, and text paths; #4304 App/CLI destination dispatch and action-scoped
mapping; then connector-local command/binary proof. No caller-selected URL,
verb, header, body, arbitrary JSON, or generic provider transport is proposed.

## Captain hard pre-merge gate — blocked

No connector or batch may be called merge-ready while the gate in
`SEVEN-SURFACE-RECONCILIATION.json#pre_merge_gate` is blocked. For every
provider-defined operation, the canonical ledger must prove a two-way trace
between the pinned provider source, the ledger row, the generated installed
CLI command, and generated website representation. Aggregate counts and a
placeholder command are not evidence of completion.

| Surface | Ledger name | Per-operation requirement |
| --- | --- | --- |
| ETL | `etl` | Extraction semantics, preserved stream/request/output evidence |
| Reverse ETL | `reverse_etl` | Exact typed action/source mapping, strategy, acknowledgement, delivery, approval |
| Direct read | `direct_read` | Fixed typed read, pagination/output semantics, generated command |
| Direct write | `direct_write` | Fixed typed mutation, confirmation/approval, preserved result |
| Binary download | `binary_read` | Fixed media/output target, byte and redirect bounds, file manifest/output facts |
| Binary upload | `binary_write` | Fixed multipart/binary schema, bounded bytes, approval and acknowledgement |

Only a provider-evidenced absence of a semantic capability permits `N/A` for a
surface. Scope, cost, destructive behavior, safety, rarity, and missing
credentials never do. Each applicable surface needs runtime fixture/conformance
evidence, output-preservation proof, and exact CLI/help/manual/website drift
checks derived from the same ledger; silent omissions fail the gate.

Zoom, Twenty, and Gong additionally require authorized provider-live
certification in their own lanes. This PR does not include those connectors and
does not authorize credentials; all other connectors remain live-certification
`pending` unless accepted credentialed evidence exists. The current PR remains
paused on F0, F2/F4, #4305, and the final #4304 dispatch heads.

### Enforcement is deferred, not complete

This planning gate is necessary but cannot certify a merge. After the named
foundations publish, the integration branch must add an executable,
CI-suitable repository validator for the fixed 100. Its schema-backed machine
output must include per-connector and aggregate pass/fail verdicts, operation
counts, surface verdicts, CLI/website/runtime-evidence verdicts, and structured
failures. The validator must have negative tests for a missing source hash,
missing or unreachable command, CLI/website drift, absent fixture/conformance
evidence, surface misclassification, binary-direction collapse, a disabled
callable operation, and omission of non-secret output. No planning JSON can be
used as a substitute for that executable validator or its passing CI result.
