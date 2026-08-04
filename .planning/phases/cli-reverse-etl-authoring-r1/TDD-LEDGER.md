# TDD ledger: cli-reverse-etl-authoring-r1

Every promotion in this phase was preceded by a failing test. "Verified by
execution" below means the command or action was actually run and the resulting
HTTP request inspected — never that a definition file was read and looked right.

Red evidence: `.planning/phases/cli-reverse-etl-authoring-r1/traces/red-run.txt`
(captured at commit `cc3c821d5`, before any definition was authored).

## Group 1 — ledger reconciliation (222 operations)

| | |
|---|---|
| **RED** | `TestReverseETLLedgerReconciles` (both bundles). Predecessor form failed with `60 reverse-ETL operations are still blocked by an unimplemented typed write action` (asana) and `162 ...` (zendesk-support). |
| **GREEN** | 119 rows rebound to `covered_by.write`; the remaining 103 zendesk rows re-blocked under two precise cited reasons. The assertion is now arithmetic — promoted + deferred must equal the original 162/60 — so rewording a reason cannot hide the shortfall. |
| **VERIFIED** | `go test ./internal/connectors/defs/asana/ ./internal/connectors/defs/zendesk-support/` → ok. |

## Group 2 — write-action execution (119 new actions)

| | |
|---|---|
| **RED** | `TestReverseETLWriteActionsExecute`. Before authoring, zendesk-support had 0 implemented reverse-ETL commands and asana 60 unbound operations; every new action subtest failed at `write action %q has no cli_surface command`. |
| **GREEN** | 59 zendesk-support + 60 asana actions authored with bounded record schemas, redaction declarations, sanitized fixtures, and generated command surfaces. |
| **VERIFIED** | Each action is executed by the real `engine.Write` against an `httptest` capture server, and the observed method/path/body are compared to its sanitized fixture. Additionally `conformance`'s own `write_request_shape` check re-runs the same execution independently: `go test ./internal/connectors/conformance/ -run 'TestConformance/(asana\|zendesk-support)$'` → ok. |

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

## Group 5 — pre-existing defect found by the red test

| | |
|---|---|
| **RED** | `TestReverseETLWriteActionsExecute/update_tag` failed: `BuildWriteCommand("tags update") = asana action=update_tag record=0: //data: required property missing`. The already-shipped `asana tags update` command exposed only `--gid` and `--name` while its record schema declares `color` and `notes`, so a user updating a tag colour could not build a valid record. |
| **GREEN** | Added `--color` and `--notes` flags so the command's flag surface covers its own schema. |
| **VERIFIED** | Subtest passes; `connectorgen validate` still reports 0 findings. |

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
