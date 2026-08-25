# Batch 1 source-operation mapping manifest

This document defines the machine-readable [Batch 1 mapping manifest](batch1-source-operation-mapping-manifest.json).
It is the independent, source-lock denominator for the ten Batch 1 providers;
it is not a connector `*-declaration-admission.json` sidecar, a generated
surface, or authorization to send a provider request.

## Task Delivery Header

- Issue: Refs #4325 — Batch 1 connector repair and declaration mapping.
- Base branch: `main`.
- Merges into: `main`.
- Delivery: Mapping manifest committed on `fm/cli-batch1-repair-r1`; connector
  materialization remains a separately verified follow-up on this branch.
- Working branch: `fm/cli-batch1-repair-r1`.
- Task: Give every locked Batch 1 source operation an independent identity,
  canonical target, and stable CLI path without treating policy or provenance
  as an implementation gap.
- Verification: Validate the JSON schema/invariants, count every source-lock
  row exactly once, check path uniqueness, and run `git diff --check`.

| Acceptance criterion | Evidence | Observable assertion |
| --- | --- | --- |
| The source denominator is independent | live | The manifest reads the ten `*-operation-source-lock.json` inventories, so removing a declaration does not change its 4,341 source records. |
| Every source row has one mapping | live | The JSON invariant check asserts 4,341 unique `record_key` and source-operation identities, with a nonempty CLI path on every record. |
| Policy cannot masquerade as a foundation | live | The validation check rejects policy/provenance values as `missing_implementation.component`; every deferred row carries one concrete component. |
| Promotion keeps its command identity | live | The stored path is the path exercised before and after a component is supplied; the regression plan below requires no path remapping. |

Required skills: `golang-documentation`, `golang-cli`, and `golang-security`.
This is documentation/mapping work, so no implementation GSD lifecycle is
claimed for the manifest itself.

## Counted result

| Provider | Runnable | Declarable now | Genuine gap | Source operations |
| --- | ---: | ---: | ---: | ---: |
| Docker Hub | 4 | 0 | 50 | 54 |
| Notion | 42 | 0 | 7 | 49 |
| Stripe | 8 | 0 | 581 | 589 |
| Bitbucket | 50 | 136 | 111 | 297 |
| GitLab | 0 | 917 | 835 | 1,752 |
| CircleCI | 16 | 75 | 20 | 111 |
| Sentry | 3 | 144 | 76 | 223 |
| Vercel | 0 | 261 | 139 | 400 |
| Asana | 82 | 130 | 37 | 249 |
| Jira | 562 | 3 | 52 | 617 |
| **Total** | **767** | **1,666** | **1,908** | **4,341** |

The split is the independently measured source accounting in
`.planning/phases/issue-4325-batch1-connectors-repair-r1/TDD-LEDGER.md:379-398`.
The 1,908 genuine gaps preserve the source-ledger evidence from
[the deferred-foundation matrix](batch1-deferred-foundation-matrix.md), while
the record inventory itself comes directly from the ten current source locks.
The manifest never has a zero denominator.

## Exact JSON contract

`schema_version` is `batch1-source-operation-mapping-manifest/v1`. The JSON
contains `source_basis`, `record_schema`, `invariants`,
`per_connector_counts`, and `records`. `records` is sorted by provider then
independent source operation identity.

Every record has this shape:

| Field | Contract |
| --- | --- |
| `record_key` | Unique `<provider>:<source.operation_id>` key. It is based on the source lock, not a sidecar-provided identifier. |
| `provider` | One of the ten Batch 1 provider names. |
| `source` | Source-lock operation identity, optional provider operation ID, protocol, exact method/path, source URL/location, lock path, SHA-256, and byte count. SHA/bytes are provenance only. |
| `lane` | Source-ledger parity lane for the operation. |
| `canonical_target` | `rest:<lowercase-method>:<exact-provider-path>` plus the same endpoint as structured method/path. |
| `intended_cli_path` | A nonempty, unique provider command path with the source of that mapping and its current availability. `manifest_reservation` is an exact future command projection, not an executable generic HTTP surface. |
| `declaration_state` | `implemented` only for the 767 already runnable rows; otherwise `deferred`. |
| `mapping_state` | `runnable`, `declarable`, or `deferred`. `declarable` means the current source shape needs connector-owned materialization, not a shared provider/importer fix. |
| `missing_implementation` | Absent only when `declaration_state` is `implemented`; otherwise present exactly once, with one concrete component, source-cited foundation/evidence, and a machine-readable `projection_prerequisite`. |

`missing_implementation.component` is closed to
`source_descriptor`, `typed_input_schema`, `action_binding`,
`cli_projection`, `executor_binding`, `source_citation`, or
`provider_contract`. It must never be `delete`, `destructive_action`,
`reverse_etl`, risk, confirmation, approval, a source hash, retained bytes,
or live certification. Those remain visible source/policy attributes but are
not a reason to omit the operation or its command path.

For a source operation without a current command projection, the path is a
stable `api op-<sha256>` reservation computed from the provider and source
operation ID. This is a typed-operation namespace reservation only: it does
not expose a generic HTTP, shell, or write command. A later definition must
use the same path unless the manifest mapping itself is explicitly reviewed
and updated.

## Mapping and promotion invariant

Every source-backed operation has a canonical target and discoverable CLI path
now. Its state determines the expected behavior, never whether it is counted.

| State | Required observable behavior |
| --- | --- |
| `implemented` / `runnable` | In an initialized isolated project with no credential, `pm <provider> <path>` resolves and reaches `error: missing --credential`; it must not be `unknown command`. |
| `deferred` / `declarable` | The same path resolves to a typed, source-cited `missing_foundation` refusal before credential or provider I/O. The exact `missing_implementation.component` names what still has to be materialized. |
| `deferred` / `deferred` | The same path resolves to the concrete source/importer/schema/contract gap recorded in the manifest, rather than disappearing from the CLI surface. |
| Promoted | Once the named component exists and real preflight accepts the contract, the same path and arguments reach the no-credential boundary. Promotion never changes the operation key or command path merely to obtain a passing count. |

Current main lacks a trustworthy deferred-command projection contract; the
frozen declaration-admission review recorded DA-001 through DA-007. Each
nonimplemented manifest row therefore carries
`projection_prerequisite.kind: runtime_deferred_command_projection`. This is
an explicit, machine-readable prerequisite rather than an absent mapping and
does not depend on PR #4351 being merged.

## Materialization order

The 1,666 `declarable` rows are the immediate connector-mapping queue. Their
manifest records already provide one source citation, lane, canonical target,
path, and a concrete connector-owned component. Materialize a bounded
connector slice by consuming those fields into the provider definition, then
prove the matching command behavior from the table above. The 1,908 genuine
gaps use the same mapping and path but cannot become executable until their
recorded component is present.

The existing [Stripe accounts-delete handoff](batch1-deferred-foundation-matrix.md#mapping-lane-handoff-stripe-accounts-delete)
is an example: its `stripe.rest.DeleteAccountsAccount` record retains a path
and a source descriptor prerequisite; DELETE/destructive policy is not the
prerequisite.

## Regression plan

For each materialized provider slice, add or extend tests that assert all
three transitions rather than only a declaration count:

1. Read the provider source lock and manifest together; reject a missing,
   duplicate, mismatched, or empty-path record. Assert the expected nonzero
   provider denominator and the aggregate `4,341 / 767 / 1,666 / 1,908`
   split.
2. For every `implemented` record, run the built binary in its own
   credential-free project and assert the exact credential-boundary result;
   `unknown command` is a reachability failure.
3. For every deferred record that the runtime can project, assert the typed
   `missing_foundation` result before credential/provider I/O and compare its
   component, source ID, and path with the manifest.
4. After a record's named component is added, rerun its original command and
   arguments and assert the credential boundary without changing the manifest
   command path.

These checks are source-to-runtime evidence. They do not fetch provider bytes,
read a credential, send a provider request, or replace source identity with a
declaration-authored denominator.
