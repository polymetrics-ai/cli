# TDD ledger: cli-reverse-etl-authoring-r1

Every promotion in this phase was preceded by a failing test. "Verified by
execution" below means the command or action was actually run and the resulting
HTTP request inspected — never that a definition file was read and looked right.

Red evidence: `.planning/phases/cli-reverse-etl-authoring-r1/traces/red-run.txt`
(captured at commit `cc3c821d5`, before any definition was authored).

## Group 1 — ledger reconciliation (222 operations, 117 promoted / 105 deferred)

| | |
|---|---|
| **RED** | `TestReverseETLLedgerReconciles` (both bundles). Predecessor form failed with `60 reverse-ETL operations are still blocked by an unimplemented typed write action` (asana) and `162 ...` (zendesk-support). |
| **GREEN** | 117 rows rebound to `covered_by.write` (zendesk-support 57, asana 60); the remaining 105 zendesk rows re-blocked under two precise cited reasons — 98 whose pinned source declares no request body, 7 whose declared body is an unbounded/bulk/`oneOf` region. See PLAN.md for why each bucket could not be derived and what would unblock it. |
| **VERIFIED** | Two independent assertions, because either alone is defeatable. (a) The MEASURED number of `api_surface.json` rows bound to a write action is pinned (`reverseETLBoundEndpoints` = 84 zendesk-support / 73 asana, both including the pre-existing writes), so dropping a `covered_by` block and re-blocking it under a third reason string fails. (b) The arithmetic `promotedReverseETLOperations + noContract + noShape == originalBlockedReverseETL` (57+98+7=162, 60+0+0=60) still holds, so rewording a reason cannot hide the shortfall. `go test ./internal/connectors/defs/asana/ ./internal/connectors/defs/zendesk-support/` → ok. |

## Group 2 — write-action execution (117 new actions)

| | |
|---|---|
| **RED** | `TestReverseETLWriteActionsExecute`. Before authoring, zendesk-support had 0 implemented reverse-ETL commands and asana 60 unbound operations; every new action subtest failed at `write action %q has no cli_surface command`. |
| **GREEN** | 57 zendesk-support + 60 asana actions authored with bounded record schemas, a required request envelope, redaction declarations, sanitized fixtures, and generated command surfaces. |
| **VERIFIED** | Each action is executed by the real `engine.Write` against an `httptest` capture server, and the observed method, path AND body are compared to its sanitized fixture. Every one of the 117 fixtures declares a non-empty `expect.body` covering the exact JSON body `engine.Write` constructs (every record field minus `path_fields`), so a mis-derived envelope key, wrong nesting level, or wrong scalar type fails the subtest rather than passing on a method/URL match alone. Additionally `conformance`'s own `write_request_shape` check re-runs the same execution independently: `go test ./internal/connectors/conformance/ -run 'TestConformance/(asana\|zendesk-support)$'` → ok. |

## Group 3 — plan → preview → explicit approval → execute

| | |
|---|---|
| **RED** | Same subtest. `commandrunner.Preflight` returned `BlockedCommandError` for every target command; `BuildWriteCommand` refused with `implemented reverse ETL commands must reference write action`. |
| **GREEN** | Promoted commands carry `write`, `risk`, `approval`, `redact_fields`, and flags covering every required record field. |
| **VERIFIED — by running the CLI, not the test harness** | Against a local capture server: <br>`pm zendesk-support operations create_brand --name "Acme Support" --subdomain acmesupport` → plan `rplan_434a…` + approval token.<br>`--preview` → resolved `POST http://…/api/v2/brands`, **0 requests reached the server**.<br>`--approve <token>` → `Reverse ETL run rrun_deec… completed: succeeded=1 failed=0`; server received exactly `POST /api/v2/brands {"brand":{"name":"Acme Support","subdomain":"acmesupport"}}`.<br>Replaying the token → `reverse plan … was already executed`.<br>Asana equivalent: `pm asana teams create-team --name Platform` → server received `POST /teams {"data":{"name":"Platform"}}`. |

## Group 4 — destructive operations stay blocked

| | |
|---|---|
| **RED** | `TestDestructiveOperationsStayBlocked` pinned the pre-existing baseline (88 zendesk / 36 asana `destructive_action` rows unbound) and required every bound DELETE to declare a typed confirmation challenge. It passed at baseline by construction, so it is a guard rather than a red-to-green step — it fails the moment a promotion crosses the line. |
| **GREEN** | No change: zendesk-support still has 77 unbound DELETE endpoints and 88 blocked `destructive_action` rows; asana still has 19 and 36. |
| **VERIFIED** | Test passes in both bundles after the change; counts re-derived independently from `api_surface.json`. |

## Group 5 — redaction declarations are live, not decorative

| | |
|---|---|
| **RED** | `assertRedactedFieldsHidden` originally probed only top-level dotted lookups and `continue`d when a declaration did not resolve, so a declaration that redacted nothing still passed. Rewritten to fail instead of skip, every bare-leaf `redact_fields` entry in `writes.json` failed: `engine.redactPreviewRecordField` and `writeActionRedactionValues` resolve a declaration as a dotted path anchored at the record ROOT and descend maps only. |
| **GREEN** | `writes.json` declarations are now root-anchored dotted paths that resolve against the real record shape (`author.email`, `brand.signature_template`, `settings.email.*`, `ticket.tpe_voice_comment.*`). Entries whose only real location sits under an array were dropped from `writes.json`, because the engine's resolver has no array-index support and the declaration would be inert; those fields stay declared on the `cli_surface.json` command, where `commandrunner.redactRecordWithFields` matches by normalized field NAME at any depth and does redact them. Every remaining `writes.json` path is populated in the action's sanitized fixture. |
| **VERIFIED** | The helper now asserts two things per action: every `writes.json` path resolves through a map-only walk of the authored record (mirroring `engine.resolveRecordPathValue`), and a full recursive walk of the approval sample shows every declared field name — from the action *and* the command — masked as `***`/`redacted` wherever it appears. |

## Group 6 — pre-existing defects found by the red test

| | |
|---|---|
| **RED (a)** | `TestReverseETLWriteActionsExecute/update_tag` failed: `BuildWriteCommand("tags update") = asana action=update_tag record=0: //data: required property missing`. The already-shipped `asana tags update` command exposed only `--gid` and `--name` while its record schema declares `color` and `notes`, so a user updating a tag colour could not build a valid record. |
| **RED (b)** | Once the fixtures asserted a real body, nine zendesk-support flags failed at `value does not match type [integer]` / `[number]`. Each declared `type: "string_array"` against a numeric array, and `coerceFlagValue` can only ever produce `[]string` for that type, so the flag was guaranteed to fail validation. `connectorgen`'s `cliFlagTypeMatchesSchema` accepts `string_array` against *any* array, which is why nothing caught it. One of them, `reorder_workspaces --ids`, was the command's only body flag, so the command could never build a valid record at all. |
| **GREEN** | (a) Added `--color` and `--notes`. (b) Re-mapped all nine onto an indexed scalar element (`record.ids.0`, `record.ticket.collaborator_ids.0`, …) with `type: "integer"`, which is the idiom these bundles already use for array bodies (`record.tickets.0.subject`, `record.group_memberships.0.group_id`). Separately, four commands whose 24-flag cap had dropped a provider-*required* nested field (`custom_object_field.title`/`.type`, `field.title`/`.type`, `data.resource_subtype`, `data.workspace`) gained those flags, so no required record path is unreachable from the command surface. |
| **VERIFIED** | Subtests pass; `connectorgen validate` still reports 0 findings. |

## Gates

| Gate | Result |
|---|---|
| `go run ./cmd/connectorgen validate internal/connectors/defs` | 550 connectors, **0 findings** |
| `go run ./cmd/connectorgen boundary . --json` | `outcome: clean`, 0 findings, 0 warnings |
| `go test ./internal/connectors/conformance/` | ok |
| `go test ./internal/connectors/commandrunner/` | ok |
| `go test ./internal/connectors/engine/` | ok |
| `go test ./internal/connectors/boundary/` | ok |
| `go test ./internal/cli/` | ok (378s) |
| `go test ./internal/connectors/defs/{asana,zendesk-support}/` | ok |
| `./pm docs validate --connectors-dir docs/connectors` | Validated |
| `go vet ./internal/connectors/...`, `gofmt -l`, `go build ./cmd/pm` | clean |
