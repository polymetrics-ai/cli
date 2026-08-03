# TDD Ledger — issue #188 Front parity

## Red / baseline evidence

Baseline before production bundle edits:

```bash
go run ./cmd/connectorgen validate internal/connectors/defs/front
```

Result: failed because this tool treats the provided directory as a connector root and therefore scanned `fixtures/` and `schemas/` as sibling connectors (`missing_file` for `metadata.json`). This records that the issue checklist's connector-dir form is not a valid invocation for this repo-local validator. Targeted validation was performed by copying `front/` under a temporary defs root; full-root validation was also run after implementation.

```bash
go test ./internal/connectors/conformance -run 'TestConformance/front' -count=1
```

Result: passed for the old six-stream read-only bundle, proving the red gap is the documented parity contract (no full ledger/writes/direct/binary metadata), not a failing legacy fixture.

## Green evidence

| Cycle | Command | Result | Notes |
| --- | --- | --- | --- |
| plan | `scripts/gsd doctor` | pass | Adapter healthy before planning. |
| plan | `scripts/gsd prompt programming-loop init --phase issue-188-front-parity --dry-run` | unavailable | Adapter lacks `programming-loop`; manual Pi GSD loop fallback recorded. |
| plan | `scripts/gsd prompt quick --full ...` | pass | Prompt trace recorded in `.planning/traces/issue-188-front-parity-wave02-r1-gsd-quick.prompt.md`. |
| red | `go run ./cmd/connectorgen validate internal/connectors/defs/front` | expected fail | Validator directory semantics; see note above. |
| red | `go test ./internal/connectors/conformance -run 'TestConformance/front' -count=1` | pass | Old scope fixture passed. |
| green | targeted temp-root `go run ./cmd/connectorgen validate <tmp-with-front> --json` | pass | 1 connector checked, 0 findings, 0 warnings. |
| green | `go run ./cmd/connectorgen validate` | pass | 549 connectors checked, 0 findings. |
| green | `go test ./internal/connectors/conformance -run 'TestConformance/front' -count=1` | pass | Front conformance passed. |
| green | `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1 -timeout=8m` | pass | Exact issue checklist command passed with extended Go timeout after updating generated golden transcripts. |
| green | `go vet ./internal/connectors/... ./internal/cli/...` | pass | No output. |
| green | `go build ./cmd/pm` | pass | No output. |
| green | `make connector-boundary` | pass | Outcome clean, no findings/warnings. |
| green | `git diff --check` | pass | No whitespace errors. |

## Increment 2 — close CLI-reachability gap (94% → 97.6%)

Red/baseline: `go run ./cmd/connectorgen validate` (0 findings, since the 10 target endpoints
were dispositioned as `excluded`/`blocked` operations, not undispositioned gaps — the "red"
here is the reachability audit finding 239/255 CLI-reachable against the updated 100%-bar
task, not a validator failure). First write attempt surfaced two real red signals fixed in
green:

| Cycle | Command | Result | Notes |
| --- | --- | --- | --- |
| red | `go run ./cmd/connectorgen validate` | fail | `cli_surface_missing_mapping`: new `implemented` write commands lacked flag mappings for every `record_schema` required field (nested array/object requireds like `lines.0.content`). Fixed by loosening `required` to `path_fields` only (matches the pre-existing `update_channel`/`update_conversation`/`create_a_channel` convention: strictly more permissive than Front's own per-field requirements, never stricter — an ACCEPTABLE deviation per `docs/migration/conventions.md` §5's meta-rule), except `add_call_recording` where the extra required fields already had matching flags. |
| red | `go test ./internal/connectors/conformance -run 'TestConformance/front'` | fail | `write_schemas_valid`/`write_request_shape:*`: `record_schema` used `minLength`/`maxLength`/`minItems`, keywords not in the engine's draft-07 subset (`internal/connectors/engine/schema.go`'s `true`-keyed keyword set: `type`/`required`/`properties`/`items`/`enum`/`minProperties`/`additionalProperties`). Fixed by stripping unsupported keywords from `add_call_summary`/`add_call_transcript`/`sync_inbound_message`/`sync_outbound_message`. |
| green | `go run ./cmd/connectorgen validate` | pass | 550 connectors checked (549 + a newer-on-`main` connector), 0 findings. |
| green | `go test ./internal/connectors/conformance -run 'TestConformance/front' -count=1` | pass | All write_request_shape checks pass or correctly Skip (only `add_call_recording`, matching `gong`'s multipart precedent). |
| green | `go build ./...` | pass | |
| green | `go vet ./internal/connectors/...` | pass | |
| red (post-rebase) | `go run ./cmd/connectorgen validate` | fail | A stricter `cli_surface_safety` rule landed on `main` (flags mapping to a `body.*` required-by-operation path must carry `"required": true`) surfaced a pre-existing gap in Front's original `analytics reports create`/`analytics exports create` commands, unrelated to this increment's 10 writes. |
| green (post-rebase) | `go run ./cmd/connectorgen validate` | pass | Fixed by adding `"required": true` to the `start`/`end`/`metrics`/`columns` flags. 550 connectors, 0 findings. |
| green | `pm docs validate --connectors-dir docs/connectors` | pass | |
| green | `pm docs generate` + manual `pm front <command> --help`/`pm connectors inspect front --json` spot checks | pass | Confirmed all 10 new commands individually reachable, 129 write actions surfaced. |
| green | website `pnpm run typecheck` | pass | |
| green | website `pnpm run lint` | pass | 0 errors (13 pre-existing warnings unrelated to Front). |
