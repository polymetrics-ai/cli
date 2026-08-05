# Phase: cli-reverse-etl-authoring-r1

Author typed reverse-ETL write definitions that unblock the already-supported
reverse-ETL operations on `zendesk-support` and `asana`. This is authoring, not
engine work: the declarative write engine (`internal/connectors/engine/write.go`)
already executes `writes.json` actions, and `internal/connectors/commandrunner`
already resolves `cli_surface.json` reverse-ETL commands through
plan -> preview -> explicit approval -> execute.

## Required skills used

`golang-how-to`, `golang-testing`, `golang-cli`, `golang-security`,
`golang-safety`, plus `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
for the generated help/manual/website parity pass.

## Inventory (counted on the live bundles, not on the audit narrative)

| Reason string (verbatim, from `api_surface.json`) | zendesk-support | asana |
|---|---:|---:|
| `Blocked by default until a connector-local typed reverse-ETL action implements plan, preview, explicit approval, and execute.` | **162** | – |
| `planned reverse-ETL write action; blocked until a named action has a bounded record schema, redaction, sanitized fixture, and plan -> preview -> explicit approval -> execute evidence.` | – | **60** |

Promotable total: **222**. Method mix contains **zero DELETE** rows
(zendesk PUT 68 / POST 80 / PATCH 14; asana POST 46 / PUT 14).

### Sibling reverse-ETL reason variants found, and why each stays blocked

| Connector | n | Reason variant | Disposition |
|---|---:|---|---|
| zendesk-support | 88 | `In scope, but blocked by default until represented by a typed action with confirm:"destructive"...` | **Stays blocked** — destructive (77 DELETE, 8 PUT, 3 POST). The shared gate now exists; connector-local typed action, command, and fixture authoring remains. |
| zendesk-support | 9 | `...typed action implements non-inline credential input where required, credential redaction...` | **Stays blocked** — OAuth client/token creation and password set/change. Promoting means accepting a secret as a CLI flag value, i.e. inline credential input, which the reason forbids and `AGENTS.md` forbids. |
| zendesk-support | 5 | `File or multipart upload needs a bounded file-transfer foundation...` | **Stays blocked** — needs the file-transfer foundation; an engine concern outside this lane's file boundary. |
| asana | 34 | `planned destructive reverse-ETL action...` | **Stays blocked** — destructive (17 DELETE, 17 POST). |
| asana | 17 | `planned admin/elevated reverse-ETL action...` | **Stays blocked** — all 17 carry `destructive: true` in `operations.json`. |
| asana | 1 | `planned destructive attachment operation...` | **Stays blocked** — destructive. |
| asana | 1 | `planned bounded attachment upload...` | **Stays blocked** — needs the file-upload foundation. |
| asana | 1 | `Asana /batch is a generic subrequest wrapper...` | **Stays blocked** — deliberate refusal of raw method/path/body passthrough. |

### Reconciliation against the audit's 276

276 = 162 (zendesk typed reverse-ETL rows) + 114 (every asana mutation row that
is not a changefeed surface: 60 + 34 + 17 + 1 destructive attachment + 1 upload
+ 1 `/batch`). The audit's extra **54 over 222** is exactly asana's
destructive/admin/upload/refused set. Those must stay blocked, so the honest
promotable figure is 222, not 276.

## Derivation source

Bounded record schemas are derived from the same pinned OpenAPI documents the
bundles already cite in `operation.source_url`:

- zendesk-support: `https://developer.zendesk.com/zendesk/oas.yaml` (Support API, 434 paths)
- asana: `https://raw.githubusercontent.com/Asana/openapi/master/defs/asana_oas.yaml` (Asana, 175 paths)

All 222 blocked endpoints resolve in those documents (162/162 and 60/60), so no
schema is invented.

Schema shape rules (matching the existing `asana/create_task` precedent):

- root object, `additionalProperties: false`, path parameters as required strings
- the provider's request envelope (`data` for asana, the singular resource key
  for zendesk) closed with an explicit property list and listed in the root
  `required` array, exactly as `create_task` requires `data` and `create_ticket`
  requires `ticket`. An optional envelope would leave the provider-required
  fields declared inside it unenforced, so a flagless invocation would validate,
  stage a preview, burn a single-use approval token, and only then be rejected
  by the provider. `connectorgen`'s `mappedRecordPathSatisfies` accepts a flag
  mapped *below* a required region, so the envelope is required and its leaves
  carry the flags.
- deeply nested free-form regions stay `type: object` with
  `additionalProperties: true`, as `create_task.custom_fields` already does
- every path that ends up in `required` — transitively, including array item
  fields — must have a CLI flag mapped at or below it, so an implemented
  command can always build a valid record from its own flag surface

## Delivered scope, and why it is smaller than the plan

| | zendesk-support | asana | total |
|---|---:|---:|---:|
| Planned (rows carrying the promotable blocked reason) | 162 | 60 | **222** |
| Promoted to a typed write action | 57 | 60 | **117** |
| Deferred — pinned source declares **no request body** | 98 | 0 | **98** |
| Deferred — request body is **unbounded / not flag-representable** | 7 | 0 | **7** |

**Why the 105 deferrals happened.** A typed reverse-ETL action needs a bounded
record schema, and the only permitted derivation source is the document the
connector itself cites. For 98 zendesk operations
`https://developer.zendesk.com/zendesk/oas.yaml` declares no `requestBody` at
all; for 7 more it declares one whose payload is an unbounded, bulk, or `oneOf`
free-form region (`TicketsUpdateRequest`, the `UserUpdateRequest`/`UsersRequest`
`oneOf`, the `create_many`/`create_or_update_many` bulk user payloads, the
routing bulk-job payload, and `UpdatePermissionPolicy`). In both buckets a
bounded schema could only be produced by inventing the payload shape — for
example by reading it back out of the operation's *response* schema. That is
exactly the defect class this repository is closing: shipping an inferred write
contract as `availability: "implemented"` recreates the implemented-but-
unreachable command (the 174 dead commands the audit counted). Two operations
were authored that way in the first pass — `tickets_update_many` and
`update_many_users` shipped with `record_schema.properties == {}` and zero
flags, so the only record they admitted was `{}` and the only request they could
ever send was a bodyless `PUT`. Both were de-promoted back into the unbounded-
shape bucket rather than given an invented payload.

**What would unblock them.** Either the pinned source gains a request body (and
a bounded one) for those operations, or the connector grows a local
request-schema ledger that cites a second reviewable source per operation. Until
one of those exists, each deferred row carries a specific, cited, actionable
reason string in `api_surface.json`, `operations.json`, and `cli_surface.json`
rather than the generic blocked reason, and the counts are pinned by
`TestReverseETLLedgerReconciles` so the shortfall cannot be quietly reworded
away.

## Tasks

1. **RED** — add `reverse_etl_execute_test.go` to each bundle directory asserting
   (a) no endpoint still carries the promotable blocked reason, (b) every write
   action resolves through the real `commandrunner.Preflight` /
   `BuildWriteCommand` / `DryRunWrite`, (c) every write action issues the
   expected HTTP request through the real `engine.Write`, and (d) no unbound
   destructive row was promoted. Capture the failing run.
2. **GREEN zendesk-support** — author an action in `writes.json` for every row
   whose pinned source yields a bounded request body (57 of 162), rebind those
   `api_surface.json` rows to `covered_by.write`, promote the matching
   `cli_surface.json` commands, and add one sanitized
   `fixtures/writes/*.json` per action whose `expect.body` asserts the exact
   JSON body `engine.Write` constructs. Re-block the remaining 105 rows under
   the two cited reasons above, in all three ledgers.
3. **GREEN asana** — the same for 60 of 60.
4. **REFACTOR** — consolidate shared shapes, drop the superseded
   `operations.json` rows for promoted operations, regenerate manuals/skills/
   website catalog.
5. **Gates** — `connectorgen validate` (0 findings), conformance, commandrunner,
   `internal/cli`, boundary, docs-check.

## Human gates

None expected. No dependency additions, migrations, deploys, auth changes,
destructive data actions, or quality-gate reductions. Promoting a destructive
operation would be a human gate; this phase deliberately promotes none.

## File boundary

`internal/connectors/defs/zendesk-support/**`, `internal/connectors/defs/asana/**`,
their fixtures, and artifacts that regenerate deterministically from those two
bundles (`docs/connectors/<name>/`, `website/data/connectors.generated.json`).
Phase artifacts live under `.planning/phases/`, which
`internal/connectors/boundary/ownership.go:isConnectorPlanningArtifactPath`
classifies as lane-local, non-shared evidence that AGENTS.md requires connector
lanes to produce. Nothing under `cmd/connectorgen/`,
`internal/connectors/commandrunner/`, `internal/connectors/engine/`,
`internal/connectors/defs/github/`, or `internal/connectors/defs/gong/` is touched.
