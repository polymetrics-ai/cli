# TDD Ledger — Front parity resume

## Recovery boundary

The historical implementation was recovered as commits `a7df96e51` and `a8bf86569` onto current
`origin/main`; this is not newly authored production code. Any new bundle behavior in this phase
must have fresh red validation evidence before its production edit.

## Active red-evidence checklist

| Slice | Red command / observation | Expected evidence | Status |
| --- | --- | --- | --- |
| runtime reachability | `go test ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight` | exposes any post-#3712 invalid implemented Front command | green after promotion |
| Front conformance | `go test ./internal/connectors/conformance/... -run 'TestConformance/front'` | exposes replay/schema/write-shape regressions | green after promotion |
| binary promotion | focused Front validation + commandrunner preflight before edits | current `planned` binary rows demonstrate the gap to close | green: five `binary_download` commands are implemented |
| citation convention | field matrix audit against provider docs | every request field cited or tier-5 deferred | research complete: 933 declared fields cited; 15 body-absence findings recorded |
| CLI registration | `go test ./internal/cli/...` | generated help must include the recovered Front namespace | green: root goldens regenerated and the full suite passed |
| conditional provider contracts | static provider/OpenAPI comparison of analytics, draft/template, link, conversation, channel, call, and merge requests | current flattened schemas and fixtures omit discriminator/allOf/cross-field obligations | remediated connector-locally with fixed variants, provider-valid fixtures, and draft-07 contract sidecar; shared composition enforcement pending |
| provider-schema contradiction review | compare the executable and composed Front contracts with the current provider-owned Core and Channel OpenAPI schemas | `variable_mappings` invents `minItems: 1`; `parent_external_call_id` admits null; custom-channel composition admits Twilio-only secrets | remediated in the action, sidecar, static-template fixture, CLI help, and cited field matrix; shared composition enforcement remains pending |
| human-intent batching | inspect `create_message`, `create_message_reply`, and `mark_message_seen` action metadata | absent `batchable` defaults true | remediated: all three actions declare `batchable: false` |
| recipient and destination constraints | compare Front recipient/destination prose with flat, composed, CLI, and research surfaces; enumerate schema and CLI constraints with property-map-aware keyword accounting | `create_message` unconditionally requires `to`; `create_conversation.teammate_ids` invents `minItems: 1` across four surfaces; the first audit omitted three `minProperties` rules and misclassified a field named `pattern` | remediated with an inclusive recipient `anyOf`, permissive flat fields, cc-only fixture, removal of all four unsupported minimum copies, and an 844-occurrence audit with the three provider-backed `minProperties` rules cited; shared composition enforcement remains pending |

## Green evidence

| Slice | Command | Result | Evidence |
| --- | --- | --- | --- |
| binary promotion (red) | `jq -e '[.commands[] | select(.availability == "planned")] | length == 0' internal/connectors/defs/front/cli_surface.json` | expected fail | printed `false`, exit 1: five documented binary commands remain planned despite #3712 runtime wiring. |
| connector validation semantics | `go run ./cmd/connectorgen validate front` | expected fail | current tool treats `front` as an unavailable root (`read root: open .: no such file or directory`); use the documented defs-root gate after edits. |
| runtime preflight baseline | `go test ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight -count=1` | pass | all currently implemented commands—including Front's 249—pass real runtime preflight; planned binaries are outside the sweep. |
| Front conformance baseline | `go test ./internal/connectors/conformance/... -run 'TestConformance/front' -count=1` | pass | recovered bundle's fixture replay and existing writes are healthy on current main. |
| binary promotion (green) | `go run ./cmd/connectorgen surface-sync --check && go run ./cmd/connectorgen validate internal/connectors/defs/front && go test ./internal/connectors/conformance/... -run 'TestConformance/front' -count=1 && go test ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight -count=1` | pass | Surface synchronization reported zero drift; Front had zero validator findings; focused conformance and current-main runtime preflight both passed after five binary promotions. |
| CLI registration (red) | `go test ./internal/cli/... -count=1` | expected fail | Root golden transcripts did not contain the newly recovered `pm front` dynamic namespace; no implementation error or runtime blocker was reported. |
| CLI registration (green) | `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test ./internal/cli -run TestGoldenTranscripts -count=1` | pass | Regenerated exactly nine root transcript entries to include Front; the focused golden suite passed in 34.432s. A full focused CLI rerun remains a final gate. |
| CLI registration (final green) | `go test ./internal/cli/... -count=1` | pass | Full focused CLI suite passed in 418.708s after the deterministic golden regeneration. |
| generated docs / website | `go run ./cmd/pm docs generate --dir docs/cli && go run ./cmd/pm docs validate --connectors-dir docs/connectors && cd website && pnpm run gen:website-data` | pass | Front manuals and catalog data regenerated. The generator's unrelated Amazon SQS text drift was explicitly excluded from this connector-owned slice. |
| built CLI smoke | built `pm`: `pm help front`, `pm front`, `pm front --help`, binary command help, and `pm connectors inspect front --json` | pass | All expected Front command surfaces render; an isolated temporary project reaches the no-credential boundary without a provider request. |
| conditional-contract review | `jq empty` on changed Front JSON, `connectorgen validate`, `surface-sync --check`, focused Front conformance, and implemented-command runtime preflight | pass | Zero Front findings and zero surface drift; conformance passed in 1.661s and runtime preflight passed in 1.362s. Shared composition integration remains an explicitly recorded external dependency. |
| provider-schema correction | `git diff --check`; focused `jq` contract assertions; Front `connectorgen validate`; `surface-sync --check`; `go test ./internal/connectors/conformance/... -run 'TestConformance/front' -count=1` | pass | Four contract assertions passed; Front validation reported zero findings, surface sync reported zero drift, and the static-template fixture passed conformance in 1.626s. |
