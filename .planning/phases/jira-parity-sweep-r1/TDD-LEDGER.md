# TDD ledger — jira parity sweep

Strict red-first. The test is written against the REAL shipped bundle, run, and its failure recorded
verbatim before any production edit.

## Cycle 1 — fifteen wildcard rows standing for 617 operations

**Red:** `cmd/connectorgen/jira_documented_surface_test.go` written and run against the bundle as
shipped, 2026-08-07.

```
$ go test ./cmd/connectorgen/ -run TestJiraDocumentedSurfaceIsComplete
--- FAIL: TestJiraDocumentedSurfaceIsComplete (0.00s)
    jira_documented_surface_test.go:225: operation_ledger_version = 0, want 1
    jira_documented_surface_test.go:301: 10 malformed row(s): DELETE /rest/api/3/issue*, /comment*,
        /worklog*, /attachment*, /project*, /filter*, /workflow* and related delete actions,
        GET /rest/api/3/attachment/content/{id}, /attachment/thumbnail/{id}, ...
    jira_documented_surface_test.go:311: 12 legacy excluded row(s) remain, want 0
    jira_documented_surface_test.go:314: documented endpoints = 15, want 617
    jira_documented_surface_test.go:317: covered(3)+blocked(0) = 3, want 15 — every row needs a disposition
    jira_documented_surface_test.go:321: byMethod = map[DELETE:1 GET:11 POST:2 PUT:1], want map[DELETE:89 GET:276 POST:134 PUT:118]
    jira_documented_surface_test.go:328: covered rows = 3, want 592
    jira_documented_surface_test.go:331: blocked rows = 0, want 25
    jira_documented_surface_test.go:334: blocked rows by named dependency class = map[], want map[dynamic_key_map:5 empty_contract:3 raw_binary_body:3 scalar_body:2 unbounded_body:12]
    jira_documented_surface_test.go:356: GET /rest/api/3/universal_avatar/view/type/{type}: avatar image read is not covered
    jira_documented_surface_test.go:364: expected "POST /rest/api/3/universal_avatar/type/{type}/owner/{entityId}" — the avatar UPLOAD, whose raw image body must never be modelled as a download
    jira_documented_surface_test.go:377: expected read-shaped POST "POST /rest/api/3/app/field/context/configuration/list"
    ... 24 read-shaped POSTs in total are absent from the surface ...
    jira_documented_surface_test.go:396: POST /rest/api/3/issue/properties: expected a covered write
FAIL
FAIL	polymetrics.ai/cmd/connectorgen	0.243s
```

The complete, unedited output is in `RUN-STATE.json.red_failure`; the block above is elided only
where a single assertion repeats per row. `24` is the measured count of read-shaped-POST failures:
`go test ... 2>&1 | grep -c 'expected read-shaped POST'` → `24`.

**What red proves.** jira's gap is the *opposite* of zendesk-support's. zendesk shipped a complete,
correctly counted 631-row inventory that was unreachable; jira ships **15 rows**, twelve of which are
comma-joined or wildcard "and similar" families standing for 602 endpoints, and `cli_surface.json`,
`writes.json` and `operations.json` do not exist at all. `byMethod = map[DELETE:1 GET:11 POST:2
PUT:1]` is the whole story in one line: one DELETE row for all 89 documented deletes.

Both the inventory assertions and the reachability assertions fail, which is what makes this a
restructure rather than a promotion — and it is exactly the check the handoff said to run first
("check what fraction of its rows are already blocked-by-blanket-sentence before deriving anything").
jira's answer: **none are blocked; twelve are `excluded` by wildcard**, which is a different failure
with the same effect, and `excluded` is not one of the three dispositions this sweep accepts
(finding 18).

**Green:** slice 2 — `tools/gen_jira.py` rewrites `api_surface.json` to 617 one-operation rows and
generates `cli_surface.json`, `operations.json` and `writes.json`. Recorded in `SUMMARY.md` with the
measured command count.

## Cycle 2 — reachability, measured by running the binary

**Red:** authoring a command is not evidence that it routes. gmail returned `unknown command` for all
79 operations while the records claimed it worked, and github's first probe passed 749 commands it
had never verified because it checked only the exit code — `pm <connector> <nonsense> --help` renders
the group help and **exits 0** (finding 30).

**Green:** `tools/probe_reachability.sh` asserts the rendered `NAME` line reads
`pm jira <path> - …` for every implemented and partial command. Result recorded in `VERIFICATION.md`.

## Shared gates held throughout

`connectorgen validate` 551/0 · `surface-sync --check` clean · the **whole** `cmd/connectorgen`
package (finding 5) · `TestEveryImplementedCommandPassesRuntimePreflight` · `go test -timeout 20m
./internal/cli/` (finding 36). jira's own surface test stays red until slice 2 lands, which is
correct and honest; a shared gate breaking is not.
