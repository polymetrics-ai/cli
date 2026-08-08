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

## HTTP-phase and local-host isolation

An opt-in `net/http/httptrace` probe of a fresh PM GitHub `create_issue` run
recorded the following ordered phases, with every recorded phase error-free:

```
ConnectStart
ConnectDone
TLSHandshakeStart
TLSHandshakeDone
WroteHeaders
WroteRequest
EOF
```

There was no `GotFirstResponseByte`. The failure is therefore post-write and
pre-response, not DNS, TCP connect, or TLS verification.

The same freshly built `pm` binary completed the normal reverse
plan → preview → approved run against a local loopback HTTP fixture using the
GitHub bundle's `base_url` override. The fixture received exactly one
`POST /repos/local/fixture/issues` with a 44-byte body and returned 201. This
proves PM can perform a connector write; the live failure is GitHub-path
specific rather than a general inability to POST.

A local TLS fixture was also attempted. PM correctly rejected its self-signed
certificate as an unknown authority; no TLS verification bypass or trust-store
change was made merely to complete a diagnostic. That result is expected and
does not alter the live trace, where GitHub TLS completed successfully.

Proxy and VPN observations, recorded without values or service names:

- Upper- and lower-case `HTTP_PROXY`, `HTTPS_PROXY`, `ALL_PROXY`, and
  `NO_PROXY` were all unset for the curl shell.
- Go's `http.ProxyFromEnvironment` returned no proxy for both GitHub HTTPS and
  loopback HTTP, matching what PM inherits.
- `scutil --nc list` reported one connected VPN service (and the host has nine
  `utun` interfaces). The VPN is active, but it is not represented as an HTTP
  proxy and the trace does not indicate a TLS failure.

The remaining plausible environmental interaction is a VPN/middlebox or
GitHub-edge response reset specific to PM/Go's successfully written request.
No network setting or no-duplicate-write safeguard was changed.

## Credential and authentication validation

The captain asked whether the failing write could instead be using a different,
anonymous, malformed, or non-writing credential. The original disposable
live-test vault had already been deliberately removed, so its historical
ciphertext cannot be reopened. To test the same construction path without
retaining a secret, a fresh disposable project added the GitHub credential from
the current `gh auth token` source, with the same private repository scope.

Observable results, with no token, digest, length, header value, or body
logged:

- `pm credentials test github-live-write --json` returned
  `CredentialTest` with `status: ok`. Its GitHub check is a GET of the private
  repository, so this is authenticated traffic; an anonymous lookup could not
  have read that private repository.
- Credential metadata recorded the `token` secret field, no `app_id`, no
  enabled `public_access`, and no `auth_type` override. GitHub's declared auth
  rules select bearer first when `secrets.token` is present, before the app and
  public branches. The reconstructed live path therefore resolved **bearer
  token auth**, not GitHub App or anonymous/public auth.
- An internal comparison computed SHA-256 and byte-length equality only. Both
  matched the `gh auth token` source; the stored token had neither a trailing
  newline nor surrounding whitespace. Neither digest nor length was emitted.
- A read-only authenticated repository permission response reported
  `admin: true`, `push: true`, and `pull: true` for the dedicated private
  repository. This is direct evidence of write permission for the same source
  credential, even if an OAuth-scope header does not advertise a classic
  `repo` label (for example, a fine-grained token).

This removes credential selection, validity, storage corruption/whitespace,
and repository write permission as explanations for the observed PM EOF. It
does not claim byte-for-byte access to the already-deleted historical vault;
that limitation is explicit.

## Strict-versus-ordinary transport matrix — strict transport isolated

The remaining hypothesis was that the write-only looking symptom actually
belonged to the transport configuration selected for a strict, non-replayable
request. An opt-in live diagnostic used the production `connsdk.Requester` and
the exact `noReplayClient` settings. For the safe GET, the test explicitly
applied the strict cloned transport and one-use `Close` request because the
production code intentionally does **not** apply strict write behavior to a
safe GET automatically.

| Request | Ordinary transport | Strict transport |
| --- | --- | --- |
| Authenticated private-repository GET | HTTP 200 | transport EOF |
| `POST /markdown` with `{"text":"hi"}` | HTTP 200 | transport EOF |
| `POST /graphql` with read-only `viewer { login }` query | HTTP 200 | transport EOF |

The POST endpoints are GitHub render/query operations and create no repository
or account state. This is decisive: a strict GET has no request body and no
mutation semantics, so `GetBody`, issue payload construction, authorization,
and the `/issues` endpoint cannot explain its EOF.

### Red: forced HTTP/1 and TLS ALPN disagreed

A GET-only flag matrix held `DisableRetries` constant so requester-managed
retries could not hide a first-attempt error. It showed:

| Configuration | Result |
| --- | --- |
| ordinary transport, no requester retry | HTTP 200 |
| `DisableKeepAlives` only | HTTP 200 |
| request `Close` only | HTTP 200 |
| force HTTP/1 only | transport EOF |
| `DisableKeepAlives` + force HTTP/1 | transport EOF |
| full strict configuration | transport EOF |
| full strict configuration with requester retries | transport EOF |

The strict live `httptrace` then proved the mismatch directly: the strict
transport declared HTTP/1-only while its cloned TLS configuration still
advertised `h2`; GitHub selected `h2`, the request received EOF, and no first
response byte arrived. A local HTTP/2-capable TLS fixture reproduced the exact
failure: it logged an HTTP/1.1 request preface on an `h2` connection.

The focused red regressions failed before production code changed:

```text
--- FAIL: TestNoReplayClientKeepsHTTPNegotiationAvailable
    strict transport forced an HTTP protocol instead of preserving normal negotiation
--- FAIL: TestNoReplayClientALPNMatchesConfiguredProtocol
    strict TLS negotiated protocol = "h2", want "http/1.1"
```

### Green: strict transport negotiates honestly

`noReplayClient` now retains the cloned transport's normal protocol negotiation
instead of setting `Transport.Protocols` to HTTP/1 after cloning. This lets the
already-advertised HTTP/2 implementation handle an `h2` ALPN result correctly.
It does **not** restore mutation replay: `DisableKeepAlives`, one-use request
`Close`, cleared `GetBody` and idempotency headers, zero requester retries, and
redirect refusal remain unchanged.

Green local regression command:

```sh
go test -timeout 2m ./internal/connectors/connsdk \
  -run '^(TestNoReplayClientKeepsHTTPNegotiationAvailable|TestNoReplayClientALPNMatchesConfiguredProtocol|TestRequesterDisableRetriesMakesMutationNonReplayable|TestRequesterDisableRetriesPreventsBodylessMutationTransportReplay|TestRequesterStrictMutationAvoidsStaleIdleConnectionReplay)$' \
  -count=1
# ok  polymetrics.ai/internal/connectors/connsdk
```

The safe live matrix is now green in every cell:

| Request | Ordinary transport | Strict transport |
| --- | --- | --- |
| Authenticated private-repository GET | HTTP 200 | HTTP 200 |
| `POST /markdown` with `{"text":"hi"}` | HTTP 200 | HTTP 200 |
| `POST /graphql` with read-only `viewer { login }` query | HTTP 200 | HTTP 200 |

The strict TLS trace now reports negotiated `h2` with no transport EOF. No
repository or account state was created by these probes.

## Red/green commands

```sh
go test ./internal/connectors/engine/ -run 'DirectRead' -count=1
go test ./internal/connectors/commandrunner/ -count=1
go test ./internal/cli/ -run 'DirectRead|ConnectorCommand|Manual|Help|Limits' -count=1
```

No live provider call is made by any test in this ledger; every fixture is a
local `httptest` server with fabricated record identifiers.

## Captain completion audit: legacy cursor flag leakage

The live command help retained `--start-cursor` on Notion direct reads even
though the page contract requires the opaque continuation to flow only through
`--page-cursor`. A focused `surface-sync` regression will begin RED by giving
an implemented direct read both `--start-cursor` and `--page-size`: sync must
remove the cursor flag while retaining the explicit page-size override. This
protects the real rule rather than merely checking help text.

Red before the generator correction:

```text
--- FAIL: TestSyncBundleRemovesLegacyDirectReadCursorFlag
    flags = [... maps_to:query.start_cursor name:start-cursor ...],
        want the legacy raw cursor removed
FAIL
```

Green after `removeLegacyDirectReadCursorFlags` made the surface synchronizer
own this derived contract:

```sh
go test -timeout 3m ./cmd/connectorgen \
  -run '^(TestSyncBundleRemovesLegacyDirectReadCursorFlag|TestParamsImportExcludesProviderPagingParametersByMeaning|TestSyncBundleReportsDivergentFlagMapsTo|TestSyncBundleConsistentBundleIsClean)$' \
  -count=1
# ok   polymetrics.ai/cmd/connectorgen

go run ./cmd/connectorgen surface-sync --check
# connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)
```

The rebuilt binary's `pm notion block children list --help` exposes
`--page-cursor` and the caller-owned `--page-size`, but no `--start-cursor`.
The retained size flag is intentional: it controls the page window, and the
engine uses that exact value when it determines completeness.

The first green run exposed one more executable legacy shape: Gong `logs list`
has no `operation` field, so it returned early from the old synchronization
loop and kept `--cursor`. Its focused regression was deliberately red first:

```text
--- FAIL: TestSyncBundleRemovesLegacyDirectReadCursorWithoutOperation
    flags = [... maps_to:query.cursor name:cursor ...], want the legacy raw cursor removed
FAIL
```

Green after moving the cursor-only cleanup before the operation metadata gate:

```sh
go test -timeout 3m ./cmd/connectorgen \
  -run '^(TestSyncBundleRemovesLegacyDirectReadCursorFlag|TestSyncBundleRemovesLegacyDirectReadCursorWithoutOperation|TestSyncBundleRemovesCursorDescribedAfterButPreservesTimestampBefore|TestParamsImportExcludesProviderPagingParametersByMeaning|TestSyncBundleReportsDivergentFlagMapsTo|TestSyncBundleConsistentBundleIsClean)$' \
  -count=1
# ok   polymetrics.ai/cmd/connectorgen

go run ./cmd/connectorgen surface-sync
# gong: ... corrected ... flag_derived=1 ...
go run ./cmd/connectorgen surface-sync --check
# 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)
```

The rebuilt `pm gong logs list --help` now exposes `--page-cursor` but no raw
`--cursor`. A whole-surface scan found zero named opaque cursor/token mappings
and zero cursor-described `after`/`before` mappings across implemented direct
reads. The `after`/`before` regression additionally preserves a timestamp
`before` filter, so an ordinary provider filter is not lost merely because its
name resembles cursor navigation.

## Post-main-refresh Parquet fixture compatibility gap

**Manual GSD gap fallback:** `scripts/gsd sources plan-phase` and
`scripts/gsd sources execute-phase` resolved the installed lifecycle sources.
The project-local Pi workers are unavailable in this session and the issue
contract forbids substituting spawned roles, so this small CI compatibility gap
is planned and executed inline. It is limited to two tests this branch added
before the Parquet transition; no production warehouse behavior is changed.

**Red:** the GitHub Verify run on refreshed head
`5391499fa8b99dc7f73abcd269f6527328889f1e` and the local focused reproduction
both failed because those fixtures wrote `repo_deletes.jsonl` and
`issue_candidates.jsonl`, which the now-correct legacy-format guard refuses:

```sh
go test -timeout 3m ./internal/app \
  -run '^(TestLimitedReversePlanPreviewsAndRunsItsExactApprovedSlice|TestGitHubCreateIssueReversePlanUsesDeclaredEndpoint)$' \
  -count=1
# FAIL: warehouse tables are stored as Parquet, but 1 table(s) are still JSONL
```

**Green plan:** replace only the hand-written JSONL setup in those two tests
with the existing `seedWarehouseTableRows` helper. That helper uses the real
`warehouse.WriteTable` Parquet writer, so fixtures follow the production format
without adding a second writer or weakening the legacy-JSONL refusal.

**Green:** both exact regressions and the complete affected package now pass:

```sh
go test -timeout 3m ./internal/app \
  -run '^(TestLimitedReversePlanPreviewsAndRunsItsExactApprovedSlice|TestGitHubCreateIssueReversePlanUsesDeclaredEndpoint)$' \
  -count=1
# ok  polymetrics.ai/internal/app  3.470s

go test -timeout 20m ./internal/app -count=1
# ok  polymetrics.ai/internal/app  228.607s

go vet ./internal/app
```

A whole-test-tree scan found no other ordinary fixture that builds a legacy
JSONL warehouse table. The remaining JSONL table references are intentional
legacy-format refusal coverage in `internal/app/warehouse_parquet_test.go`,
`internal/app/warehouse_connection_isolation_test.go`, and
`internal/connectors/warehouse_test.go`; outbox, WAL, and unrelated JSONL
fixtures are not warehouse table fixtures and remain untouched.

## Captain parameter and real-result proof

The final focused CLI regression drives the embedded GitHub bundle through a
local server and asserts the server-observed request and returned record counts:

```sh
go test -timeout 3m ./internal/cli \
  -run '^TestGitHubDirectReadParametersAndPageContextReachWire$' -count=1
# ok   polymetrics.ai/internal/cli
```

It proves an invalid enum names `all|closed|open` and reaches no provider, a
missing `pull_number` reaches no provider, `since` is sent exactly, and two
pages of 100 then 20 returned rows report `complete: false` then `true` while
reaching all 120 fixture rows. A separate connector regression proves the
declared `--per-page 37` reaches the wire and preserves the returned fixture
content:

```sh
go test -timeout 3m ./internal/connectors/defs/crisp \
  -run '^TestCrispListCommandPreservesFixtureContent$' -count=1
# ok   polymetrics.ai/internal/connectors/defs/crisp
```
