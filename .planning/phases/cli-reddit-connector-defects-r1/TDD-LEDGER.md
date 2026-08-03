# TDD Ledger — Reddit connector three-defect fix (PR #3677)

Manual-GSD fallback: `scripts/gsd prompt programming-loop` is absent from the repo-local command
registry (`scripts/gsd: unknown GSD command: programming-loop`), so this ledger records the manual
GSD/TDD loop per `PLAN.md`.

## Red/green slices

| Slice | Red evidence | Green evidence | Status |
|---|---|---|---|
| Defect 1a — User-Agent declaration site | Red: `TestRedditUserAgentDeclaredAsTemplatedHeader` against the base reddit bundle (`9c977b208`) fails with `base.user_agent = "polymetrics-go-cli", want empty (the field is not interpolated; use base.headers)`. | Green: `base.user_agent` is gone and `base.headers["User-Agent"]` holds `go:ai.polymetrics.cli:v1 (by /u/{{ config.reddit_username }})`. Test passes. | Green |
| Defect 1b — User-Agent on the wire | Red: `TestRedditUserAgentSentOnReadRequests` against the base bundle fails with `request 0 User-Agent = "polymetrics-go-cli", want "go:ai.polymetrics.cli:v1 (by /u/polymetrics_test_bot)"`. | Green: both the initial page and the `after` cursor follow-up carry the resolved conforming header; 2 records over 2 requests. Test passes. | Green |
| Defect 1c — fail-closed without `reddit_username` | Red: `TestRedditUserAgentFailsClosedWithoutUsername` against the base bundle fails — the base bundle has no templated header, so `Read` proceeds and hits the test server 5 times (retries), ending in `Read error = reddit stream=posts page=0: http 500 ...` instead of a header-resolution error naming `reddit_username`. | Green: header interpolation fails closed before any request is sent; the error names both `User-Agent` and `reddit_username`, and the test server is never called. Test passes. | Green |
| Defect 2 — api_surface honesty | Red: `api_surface.json` listed `GET /r/{subreddit}/comments` as ordinary covered surface with no note that it is undocumented, and carried no entry for the real documented endpoint. Verified against the `cli-reddit-rdtcli-official-api-parity-r1` report §4.3 (the route is absent from Reddit's 202 documented endpoints). | Green: `scope` carries a dated, cited caveat naming the undocumented route and both blockers (two-element top-level Listing array vs. `records.path`'s dotted-key-only traversal; N+1 fan-out against the 100 QPM budget), plus an explicit `excluded` entry for `[/r/subreddit]/comments/{article}`. `out_of_scope` count moves 5 -> 6 and `docs.md` matches. | Green (documentation-only; no behavior change, so no Go test) |
| Defect 3 — token expiry documented | Red: `spec.json`, `docs.md`, and the 401 hint were silent about the 1-hour bearer-token lifetime, so an operator hitting a mid-sync 401 had no way to know refresh was the cause. | Green: the expiry is stated in `access_token`'s description, in `docs.md`, and in the 401 hint (`reddit bearer tokens expire 1 hour after issuance and this connector does not refresh them`). The 429 hint now names `X-Ratelimit-Used`, `X-Ratelimit-Remaining`, and `X-Ratelimit-Reset`. | Green (documentation-only; no behavior change, so no Go test) |
| Engine precedence assumption | Red: the fix depends on `base.headers` beating `base.user_agent`; an unverified assumption would silently ship the wrong header. | Green: verified in-tree — `Requester.applyHeaders` (`internal/connectors/connsdk/http.go:623`) sets `User-Agent` from `r.UserAgent` at lines 629-631, then applies `r.DefaultHeaders` with `Header.Set` at lines 638-640, so `DefaultHeaders` always wins. `TestRedditUserAgentSentOnReadRequests` pins this on the wire rather than trusting the reading. | Green |

## Actual evidence

```bash
# --- RED: base reddit bundle (streams.json + spec.json from 9c977b208) + current test ---
git checkout 9c977b208 -- internal/connectors/defs/reddit/streams.json \
                          internal/connectors/defs/reddit/spec.json
grep -n 'user_agent' internal/connectors/defs/reddit/streams.json
#   4:    "user_agent": "polymetrics-go-cli",

go test ./internal/connectors/engine/... -run 'TestRedditUserAgent' -count=1
# --- FAIL: TestRedditUserAgentDeclaredAsTemplatedHeader (0.01s)
#     reddit_user_agent_test.go:41: base.user_agent = "polymetrics-go-cli", want empty (the field is not interpolated; use base.headers)
# --- FAIL: TestRedditUserAgentSentOnReadRequests (0.14s)
#     reddit_user_agent_test.go:103: request 0 User-Agent = "polymetrics-go-cli", want "go:ai.polymetrics.cli:v1 (by /u/polymetrics_test_bot)"
# --- FAIL: TestRedditUserAgentFailsClosedWithoutUsername (7.57s)
#     reddit_user_agent_test.go:117: unexpected request to /r/golang/new without reddit_username configured   (x5, retries)
#     reddit_user_agent_test.go:135: Read error = reddit stream=posts page=0: http 500 for http://127.0.0.1:62896/r/golang/new: Internal Server Error, want it to name the User-Agent header and reddit_username
# FAIL	polymetrics.ai/internal/connectors/engine	9.115s

git checkout HEAD -- internal/connectors/defs/reddit/streams.json \
                     internal/connectors/defs/reddit/spec.json   # restored; worktree clean

# --- GREEN: branch head ---
go test ./internal/connectors/engine/... -run 'TestRedditUserAgent' -count=1 -v
# === RUN   TestRedditUserAgentDeclaredAsTemplatedHeader
# --- PASS: TestRedditUserAgentDeclaredAsTemplatedHeader (0.01s)
# === RUN   TestRedditUserAgentSentOnReadRequests
# --- PASS: TestRedditUserAgentSentOnReadRequests (0.16s)
# === RUN   TestRedditUserAgentFailsClosedWithoutUsername
# --- PASS: TestRedditUserAgentFailsClosedWithoutUsername (0.00s)
# PASS
# ok  	polymetrics.ai/internal/connectors/engine	1.883s

gofmt -l internal/connectors/engine/reddit_user_agent_test.go
# no output

go vet ./internal/connectors/engine/...
# exit 0

go test ./internal/connectors/engine/... -count=1
# ok  	polymetrics.ai/internal/connectors/engine	11.386s

go build ./cmd/pm
# exit 0
```

## Notes

- The red run was produced by checking out only the two base reddit definition files into the
  worktree and restoring them immediately afterwards; `git status --porcelain` was verified empty
  after restore, and HEAD stayed at `5ae5c62b0`. No production code was ever weakened on the branch.
- The `TestRedditUserAgentFailsClosedWithoutUsername` red run takes ~7.5s because the base bundle
  lets the request through and the engine retries the synthetic 500 five times. That is the point of
  the test: on the fixed bundle it fails closed in 0.00s, before any request leaves the process.
- Defects 2 and 3 are documentation-honesty fixes with no runtime behavior change, so they carry
  prose red/green evidence rather than Go tests. Only defect 1 changes what goes on the wire, and it
  is the one pinned by tests.
- The `v1` version segment in the User-Agent is a hardcoded literal, not a real build version. No
  `{{ version }}` template namespace exists in the engine and adding one is a shared-runtime change
  outside this connector-scoped slice. Recorded as an open item.
- No secrets requested, printed, summarized, or stored. The test uses the synthetic literal
  `synthetic-conformance-secret` against an `httptest` server; no live Reddit call was made and no
  credential was read.
