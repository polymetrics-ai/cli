# TDD ledger — Issue #3898: direct-read page context

Every behaviour below started RED with retained failing evidence, then became
GREEN only after the same focused command passed.

**These tests assert RETURNED RECORD COUNTS against a known-larger fixture, never
exit status.** That is the whole point: the defect exited 0 while discarding
97.9% of a collection, so an exit-status assertion would have passed against it.

## Red evidence, verbatim

Captured from `go test ./internal/connectors/engine/ -run 'DirectRead...'`
before any implementation existed:

```
--- FAIL: TestOperationDirectReadReturnsEveryRecordForRootArray
    records = 30, want 120 (provider default page is 30 — a short result must never be reported as a complete one)
--- FAIL: TestOperationDirectReadReturnsEveryRecordForCursorEnvelope
    results = 30, want 120
--- FAIL: TestDirectReadReturnsEveryRecordForNestedCursor
    logs = 30, want 120
--- FAIL: TestOperationDirectReadReturnsEveryRecordForOffsetLimit
    results = 30, want 120
```

The fixture holds 120 records and serves 30 when the client sends no page-size
parameter — the exact shape that produced the live GitHub finding.

| ID | Guarantee | Red assertion | Status |
| --- | --- | --- | --- |
| P1 | Declared page, not provider default | A `page_number` read returns the provider's default 30 instead of the declared 100. | GREEN |
| P2 | Collection reachable by page number | Following `next_number` does not reach all 120 records. | GREEN |
| P3 | Cursor strategies hand back a token | A `cursor` read reports no `next_cursor`, or claims an addressable page number it does not have. | GREEN |
| P4 | Legacy executor parity | The non-operation `DirectRead` path truncates where the operation path does not. | GREEN |
| P5 | offset_limit is addressable | `--page 2` on an `offset_limit` strategy does not return the second window. | GREEN |
| P6 | Single object unaffected | A one-object read issues more than one request, or is reported as an incomplete collection. | GREEN |
| P7 | Undeclared paging is admitted | A bundle declaring no strategy reports `complete: true` it cannot prove. | GREEN |
| P8 | Refusal, never a quiet page one | Asking a cursor strategy for `--page 3` silently returns page one instead of erroring. | GREEN |
| P9 | Cursor onto a non-final page | Following a cursor onto a page that still has successors faults on a nil loop-guard map. | GREEN |

P9 exists because an earlier test passed while hiding a real panic: its followed
page happened to be the last one. Reverting the fix reproduces
`panic: assignment to entry in nil map` in `tokenPathCursor.Next`, which is the
retained red evidence for that row.

## Captain validation regression: limited reverse plan

The captain-required private-repository validation staged one row from the
three-row sample table. `pm reverse plan` succeeded, but `pm reverse run`
rejected the untouched source before any GitHub request. The local regression
below reproduces the same condition without credentials or a network write:

```
--- FAIL: TestLimitedReversePlanPreviewsAndRunsItsExactApprovedSlice (1.43s)
    reverse_confirmation_test.go:192: PreviewReversePlan() error = reverse plan source rows or payload files changed before preview
FAIL
```

The red test plans `Limit: 1` from a two-record warehouse fixture, then
requires preview and run to stage exactly one record. Before the correction,
preview and execution instead read `RecordCount + 1`, hashing a second,
unapproved record as if it were drift. The pre-existing changed-row rejection
remains a separate green regression.

Green after changing both preview and run to read `max(1, RecordCount)`:

```sh
go test -timeout 20m ./internal/app -run '^(TestLimitedReversePlanPreviewsAndRunsItsExactApprovedSlice|TestRunReverseETLRejectsPlanHashMismatchWhenRowsChange)$' -count=1
# ok   polymetrics.ai/internal/app
```

## Live GitHub write-target investigation

The captain-authorized live `create_issue` dispatch displayed
`Post "https://api.github.com/.../issues%22: EOF` before a private-repository
mutation. A local regression was added to assert the full generic reverse
workflow sends exactly `POST /repos/acme/widgets/issues`. It is GREEN on the
current implementation:

```sh
go test -timeout 20m ./internal/app -run '^TestGitHubCreateIssueReversePlanUsesDeclaredEndpoint$' -count=1
# ok   polymetrics.ai/internal/app
```

The user-required trace then settled the apparent malformed target before any
claim about a live discrepancy:

- The exact `pm` argv contained the expected `reverse run`, plan ID, `--root`,
  redacted approval argument, and `--json`; no argument had a literal quote.
- The stored connection and reverse-plan records had no values containing a
  literal quote in any path/config field inspected (including the GitHub owner
  and repository fields).
- Immediately before `client.Do(req)`, the resolved input path and endpoint
  were exactly `POST /repos/karthik-sivadas/<private-test-repo>/issues`, with
  no quote in the base URL or target.

Fixture and live execution share `Requester.doWithBody`. They diverge only
after `internal/connectors/connsdk/http.go` calls `client.Do(req)`: the fixture
returns its created response, while the live transport supplied Go's standard
quoted URL error form (`Post "https://.../issues": EOF`). The misleading
`%22` was introduced afterward by the error renderer, not by the request:

1. `internal/connectors/engine/errors.go:51` sends the error text through
   `safety.RedactErrorText`.
2. Before this fix, `internal/safety/safety.go` matched URLs with
   ``https?://[^\s]+``; that match swallowed Go's closing `"` delimiter.
3. `redactURL` then called `parsed.String()`, which percent-encoded the
   swallowed delimiter to `%22`.

The new safety regression first failed with the exact effect, including a
query-bearing variant so the secret-redaction contract remains covered:

```
--- FAIL: TestRedactErrorTextPreservesQuotedHTTPTransportURLDelimiter/without_query
    RedactErrorText() = "send request: Post \"https://api.example.test/v1/items%22: EOF",
        want "send request: Post \"https://api.example.test/v1/items\": EOF"
--- FAIL: TestRedactErrorTextPreservesQuotedHTTPTransportURLDelimiter/with_query
    RedactErrorText() = "send request: Post \"https://api.example.test/v1/items: EOF",
        want "send request: Post \"https://api.example.test/v1/items\": EOF"
FAIL
```

Green after excluding literal `"` from `httpURLPattern`:

```sh
go test -timeout 20m ./internal/safety -count=1
# ok   polymetrics.ai/internal/safety
```

Therefore this is neither an outbound malformed endpoint nor harness quoting.
The remaining live outcome is a plain transport `EOF`; it is recorded without
substituting another client for `pm`.

## PM-only EOF isolation

After the formatter correction, the same-machine controls against the one
captain-authorized private repository settled the network/body question:

| Caller | Request shape | Result |
| --- | --- | --- |
| curl | POST, HTTP/1.1, JSON body (43 bytes) | HTTP 201 |
| `gh api` | equivalent POST | created issue #2 |
| curl | POST, HTTP/1.1, `Connection: close`, JSON body (44 bytes) | HTTP 201 |
| curl | PM's HTTP/1.1, `Connection: close`, User-Agent, API-version, and header-name shape | HTTP 201 |
| `pm reverse run` | fresh plan → preview → approved one-row `create_issue` | transport EOF; no issue created |

The temporary PM probe emitted metadata only:

```
method=POST
target_path=/repos/karthik-sivadas/<private-test-repo>/issues
content_length=44
body_bytes_read=44
header_names=[Accept Authorization Content-Type User-Agent X-Github-Api-Version]
transport_error=true
```

So PM's declared Content-Length equals the bytes its Go transport read; the
classic length/body mismatch hypothesis is disproved. The independently listed
repository contained exactly the four curl/`gh` diagnostic issues and no PM
issue, so this is not merely a lost response after a successful mutation.

The differential is now PM's transport path: `writeRequester` makes this
non-idempotent action no-retry, and `noReplayClient` uses a fresh one-use Go
HTTP/1.1 connection. Curl succeeded with the same HTTP version and connection
close semantics, so it is not a generic GitHub outage or a simple
`Connection: close` incompatibility. A VPN/middlebox interaction specific to
Go's transport remains possible, but it is a PM-path defect until disproved.

No prior live PM write success is recorded. The original GitHub live report
performed zero mutations and left all 196 writes untested; the captain's order
states that this reverse-ETL half "could never be tested before." This is the
first recorded end-to-end PM write exercise. Six PM `create_issue` runs in the
dedicated repository have all failed with transport EOF.

## Stale pooled-connection replay hypothesis — disproved

This hypothesis was tested before changing production transport behavior. It
does not explain the live EOF:

- `Requester.do` constructs the JSON body as `bytes.NewReader(payload)`.
  Although `requestBody.Reader` is declared as `io.Reader`, the concrete
  `*bytes.Reader` reaches `http.NewRequest`, which populates `Request.GetBody`.
- `disableTransportReplay` clears `GetBody` only afterward for a strict
  non-idempotent write. That is deliberate: restoring it would permit an
  automatic replay that can duplicate an external mutation.
- `noReplayClient` clones the transport with keep-alives disabled, and the
  strict request sets `Close=true`; this path has no pooled idle connection to
  reuse.

The new test uses a primed idle connection with a deterministic pre-write
failure on its next `POST`, the transport-level equivalent of a stale server
close. It proves both sides of the contract:

```sh
go test -timeout 20m ./internal/connectors/connsdk \
  -run '^(TestRequesterKeepsJSONBodyReplayableBeforeNoReplayPolicy|TestRequesterReplaysReplayableJSONPostAfterStaleIdleWriteFailure|TestRequesterStrictMutationAvoidsStaleIdleConnectionReplay)$' \
  -count=1
# ok   polymetrics.ai/internal/connectors/connsdk
```

The ordinary request retained `GetBody`, retried onto a fresh dial, and reached
the handler exactly once. The strict mutation instead opened a fresh connection
before its first send; the injected stale-connection failure never fired, and
the handler again saw exactly one dispatch. This is a green falsification, not
a red regression: the proposed failure is absent in the current code, so no
production change was made or forced.

## Red/green commands

```sh
go test ./internal/connectors/engine/ -run 'DirectRead' -count=1
go test ./internal/connectors/commandrunner/ -count=1
go test ./internal/cli/ -run 'DirectRead|ConnectorCommand|Manual|Help|Limits' -count=1
```

No live provider call is made by any test in this ledger; every fixture is a
local `httptest` server with fabricated record identifiers.
