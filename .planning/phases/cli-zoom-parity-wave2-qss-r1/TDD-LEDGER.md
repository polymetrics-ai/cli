# Zoom Wave 2 (qss module) — TDD ledger

## Baseline (before this phase)

`internal/connectors/defs/zoom/` was a 3-stream, read-only bundle (Wave 0/Wave 1 result):
`metadata.json` (`capabilities.write: false`), `spec.json`, `streams.json` (3 streams: `users`,
`meetings`, `webinars`), `api_surface.json` (1,913 rows: 3 `covered_by.stream` rows, 1,839
`operation.status=blocked` implementable_now rows, 17 provider-restricted rows, 54 justified
deprecated exclusions), `docs.md`, `schemas/*.json`, `fixtures/{check.json,streams/**}`,
`command_surface_test.go` (3 passing tests). No `operations.json`, `cli_surface.json` had 3
commands, no `writes.json`.

## RED — command_surface_test.go

Extended the existing `TestProviderInventoryLedgerIsComplete` and
`TestCoveredStreamsHaveReachableCommands` assertions to the Wave 2 target (`covered` 3→6,
`implementable_now` 1839→1836, exact reachable command count 3→6, plus per-command
operation/api_surface/output_policy/required-flag assertions for the 3 new `qss` commands), and
added a new `TestQSSOperationDirectReadCommandsExecuteWithFixtures` test exercising each new
command through `commandrunner.Run` against an `httptest` server and committed fixtures. Committed
fixture files for the new direct reads (`fixtures/direct_reads/list_{meeting_participants,
webinar_participants,session_users}_qos_summary.json`) at this step so the GREEN commit only needs
to touch declaration JSON.

Run against the **unmodified** production JSON (qss endpoints still `operation`-blocked,
`cli_surface.json` still 3 commands, no `operations.json`):

```
=== RUN   TestProviderInventoryLedgerIsComplete
    command_surface_test.go:125: executable stream-backed rows = 3, want 6
    command_surface_test.go:128: operations awaiting Zoom-local contracts = 1839, want 1836
--- FAIL: TestProviderInventoryLedgerIsComplete (0.04s)
=== RUN   TestCoveredStreamsHaveReachableCommands
    command_surface_test.go:216: Zoom cli_surface commands = 3, want exactly 6 (Wave 1 streams + Wave 2 qss operations)
--- FAIL: TestCoveredStreamsHaveReachableCommands (0.03s)
=== RUN   TestQSSOperationDirectReadCommandsExecuteWithFixtures
=== RUN   TestQSSOperationDirectReadCommandsExecuteWithFixtures/meeting_participants
    command_surface_test.go:507: Run("qss meeting-participants list") = connector command "qss meeting-participants list" is blocked: unknown command
--- FAIL: TestQSSOperationDirectReadCommandsExecuteWithFixtures
FAIL
FAIL	polymetrics.ai/internal/connectors/defs/zoom	0.662s
```

Committed as the red checkpoint (`test(connectors): add failing zoom QSS direct-read coverage
(red)`) before any production JSON changed.

## GREEN — production JSON

1. `internal/connectors/defs/zoom/operations.json` created: 3 `rest_read` operations
   (`zoom.list_meeting_participants_qos_summary`, `zoom.list_webinar_participants_qos_summary`,
   `zoom.list_session_users_qos_summary`), `risk: medium`, `approval: none`,
   `output_policy: json_redacted`, `rest.max_bytes: 1048576`.
2. `internal/connectors/defs/zoom/cli_surface.json`: added a `qss` group and 3 commands
   (`qss meeting-participants list`, `qss webinar-participants list`, `qss session-users list`),
   each with one required path-parameter flag. `api_surface`/`output_policy` fields left
   unauthored and filled by `go run ./cmd/connectorgen surface-sync internal/connectors/defs`
   (never hand-authored, per `docs/migration/conventions.md` §2).
3. `internal/connectors/defs/zoom/api_surface.json`: the 3 `qss` endpoint rows' `operation` block
   replaced with `covered_by.direct_read = "<command path>"` (mechanical Python transform,
   `ensure_ascii=False` to avoid unrelated unicode-escaping diff noise elsewhere in the 1,913-row
   file).
4. `command_surface_test.go`'s `TestProviderInventoryLedgerIsComplete` `covered_by` disposition
   check widened from `endpoint.CoveredBy.Stream == ""` to also accept `Write`/`DirectRead`/
   `DirectReads` — the prior check assumed every `covered_by` row was stream-backed, which was
   true before this wave and false after.

First green run surfaced two real, non-obvious findings (not pre-planned):

- **Path-prefix stripping**: `engine/direct_read.go`'s `normalizeDirectReadPathForBaseURL` strips
  the configured `base_url`'s own path component from the declared `/v2/...` operation path before
  the request-layer concatenates `base_url + path`. The execution test's mock `base_url` must
  include a `/v2` suffix (mirroring the real `https://api.zoom.us/v2` default) for this to
  exercise correctly; a bare-host mock `base_url` leaves the full `/v2/...` path unstripped, and
  the wire-level request path the mock server actually receives is always the post-concatenation
  full path (`/v2/...`), not the pre-concatenation stripped one.
- **`next_page_token` redaction**: the shared `json_redacted` output policy's
  `shouldRedactJSONField` redacts any field whose name *contains* `token`, which catches the QSS
  response's own `next_page_token` pagination cursor — not a credential. This is existing shared
  engine behavior, out of this connector-scoped lane to change; documented in `docs.md` and
  accounted for explicitly in the execution test's expected-output construction rather than
  worked around.

Final green run:

```
=== RUN   TestProviderInventoryLedgerIsComplete
--- PASS: TestProviderInventoryLedgerIsComplete (0.04s)
=== RUN   TestCoveredStreamsHaveReachableCommands
--- PASS: TestCoveredStreamsHaveReachableCommands (0.03s)
=== RUN   TestCoveredStreamCommandsExecuteWithFixtures
--- PASS: TestCoveredStreamCommandsExecuteWithFixtures (0.04s)
=== RUN   TestQSSOperationDirectReadCommandsExecuteWithFixtures
--- PASS: TestQSSOperationDirectReadCommandsExecuteWithFixtures (0.10s)
PASS
ok  	polymetrics.ai/internal/connectors/defs/zoom	0.713s
```

## Downstream regression: golden CLI transcripts

`internal/cli`'s `TestGoldenTranscripts` pins the exact stdout of ~80 real CLI invocations,
including the root `pm --help`/`pm help`/manual-page connector catalog listing, which embeds
every connector's `metadata.json` description verbatim. Changing zoom's description (to mention
QSS reads) broke 9 of those golden entries (all root-help variants; verified programmatically that
every changed line in the diff contained the string "zoom" — no unrelated drift). Regenerated with
`POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 go test ./internal/cli/ -run TestGoldenTranscripts`, then
re-ran the full suite green.
