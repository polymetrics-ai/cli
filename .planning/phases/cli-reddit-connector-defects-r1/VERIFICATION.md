# Verification — Reddit connector three-defect fix (PR #3677)

## Required commands

| Command | Result | Notes |
|---|---|---|
| `scripts/gsd doctor` | Pass | Adapter healthy; node v24.13.1, 69 commands registered; official docs, commands registry, upstream lock, and Pi resources all `ok`. |
| `scripts/gsd prompt programming-loop init --phase cli-reddit-connector-defects-r1 --dry-run` | Fallback | Registry returned `scripts/gsd: unknown GSD command: programming-loop`; explicit manual-GSD fallback recorded in `PLAN.md` and `TDD-LEDGER.md`. |
| `gofmt -l internal/connectors/engine/reddit_user_agent_test.go` | Pass | No output; the one new Go file is formatted. |
| `go vet ./internal/connectors/engine/...` | Pass | Exit 0. |
| `go test ./internal/connectors/engine/... -count=1` | Pass | `ok polymetrics.ai/internal/connectors/engine 11.386s`. |
| `go build ./cmd/pm` | Pass | Exit 0. |

## TDD gate evidence

| Check | Result | Notes |
|---|---|---|
| Base-bundle red run (reddit `streams.json` + `spec.json` from `9c977b208`, current test) | Red as expected | All three `TestRedditUserAgent*` cases fail. Declaration case reports `base.user_agent = "polymetrics-go-cli"`; wire case reports `request 0 User-Agent = "polymetrics-go-cli"`; fail-closed case reaches the test server and ends in an HTTP 500 instead of a header-resolution error. |
| Branch-head green run | Pass | All three cases pass; `ok ... 1.883s`. |
| Worktree restored after the red run | Clean | `git status --porcelain` empty; HEAD `5ae5c62b0`; `base.headers["User-Agent"]` confirmed present again in `streams.json`. |

## Defect-by-defect closure

| Defect | Closed by | Pinned by |
|---|---|---|
| 1 — non-conforming `user_agent: "polymetrics-go-cli"` | Templated `base.headers["User-Agent"]` = `go:ai.polymetrics.cli:v1 (by /u/{{ config.reddit_username }})`, driven by the new required `reddit_username` config field. | `TestRedditUserAgentDeclaredAsTemplatedHeader`, `TestRedditUserAgentSentOnReadRequests`, `TestRedditUserAgentFailsClosedWithoutUsername`. |
| 2 — `api_surface.json` claimed documented coverage for `GET /r/{subreddit}/comments` | Dated, cited scope caveat plus an explicit `excluded` entry for the real documented route `[/r/subreddit]/comments/{article}`; `out_of_scope` count 5 -> 6, mirrored in `docs.md`. | Documentation-only; no runtime behavior change. Prose red/green in `TDD-LEDGER.md`. |
| 3 — undocumented 1-hour `access_token` expiry | Expiry stated in `spec.json`'s `access_token` description, in `docs.md`, and in the 401 hint; 429 hint now names `X-Ratelimit-Used` / `-Remaining` / `-Reset`. | Documentation-only; no runtime behavior change. Prose red/green in `TDD-LEDGER.md`. |

## Engine-constraint evidence (assumptions verified, not inferred)

| Claim | Verified how | Result |
|---|---|---|
| `base.user_agent` is never interpolated | `base.user_agent` is copied to `connsdk.Requester.UserAgent` verbatim; only `base.headers` values go through `InterpolateHeader`. | Confirmed. Leaving the template on `base.user_agent` would ship literal `{{ config.reddit_username }}` on the wire. |
| `base.headers` beats `base.user_agent` | `Requester.applyHeaders`, `internal/connectors/connsdk/http.go:623` — sets `User-Agent` from `r.UserAgent` at lines 629-631, then applies `r.DefaultHeaders` via `Header.Set` at lines 638-640. | Confirmed; `DefaultHeaders` always wins. Additionally pinned on the wire by `TestRedditUserAgentSentOnReadRequests`. |
| Documented per-article comments route is unreachable | Response is a two-element top-level Listing array (`[post_listing, comment_listing]`); `records.path` is dotted-key traversal only (`connsdk.RecordsAt`), with no array-index selection. Per-article scope also implies an N+1 fan-out against Reddit's 100 QPM budget. | Confirmed; migration deferred, honesty caveat shipped instead. |

## Breaking-change evidence

| Check | Result | Notes |
|---|---|---|
| `reddit_username` added to `spec.json` `required[]` | Intentional break | `reddit` is `release_stage=ga`, so adding a required config field is a breaking config change regardless of current user count. |
| Commit / PR type | `fix(connectors)!:` | Commit `9b25e9a69` `fix(connectors)!: require reddit_username for a conforming reddit User-Agent`. The `!` marker makes release-please and `CHANGELOG.md` report the break honestly. |
| Captain decision | Recorded | Keep the field required as committed (no existing reddit connector users), but mark it breaking. This closes the prior run's `breaking-required-config-unflagged` ask-user finding. |

## CLI/help/docs/website parity evidence

| Check | Result | Notes |
|---|---|---|
| `pm help <topic>`, `pm <namespace>`, `pm <command> --help`, `docs/cli/**` | Not applicable | No `pm` command, subcommand, flag, output, or help-topic change; this slice edits a connector bundle and its generated documentation only. |
| Connector surface (`pm connectors inspect reddit --json`) | Updated | Spec gains the required `reddit_username` field and the revised `access_token` description. |
| `docs/connectors/catalog/all-connectors.json` | Updated, reddit-only | `git diff --numstat` reports `7 2` (7 insertions, 2 deletions), all inside the reddit entry: `required[]` gains `reddit_username`, the `access_token` description is updated, and the `reddit_username` property is added. Verified no other connector entry is touched. |
| `docs/connectors/reddit/MANUAL.md`, `docs/connectors/reddit/SKILL.md` | Updated | Reddit-scoped only. |
| `website/data/connectors.generated.json`, `website/lib/connectors.catalog.data.generated.json` | Updated, reddit-only | One line each. |
| Full `pm docs generate` deliberately not run | Recorded | It regenerates all ~550 connectors due to pre-existing unrelated drift; that churn is kept out of this connector-scoped PR. Only the reddit entry was synced by hand, and the diff scope was verified. |
| `docs.md` markup | Fixed | Commit `5ae5c62b0` drops `**not**` bold markup that the connector docs renderer does not support and that surfaced as literal asterisks on the website. |

## Repository gate evidence

| Check | Result | Notes |
|---|---|---|
| `scripts/verify-gsd-workflow` before this phase | Fail, reproduced locally | `GSD_BASE_REF=9c977b208 scripts/verify-gsd-workflow` -> `verify-gsd-workflow: cmd/internal changed, but no GSD planning evidence changed.` exit 1. This is the `gsd-workflow-evidence` CI failure on PR #3677: the branch changed `internal/connectors/defs/reddit/**` and added `internal/connectors/engine/reddit_user_agent_test.go` but shipped no `.planning/**` evidence. |
| `scripts/verify-gsd-workflow` after this phase | Pass | This phase directory supplies `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, and `RUN-STATE.json`, each recording the explicit manual-GSD fallback plus red/green evidence, satisfying both the path check and the GSD/TDD content check. |

## Open items for the human merge gate

- **Hardcoded version segment.** The User-Agent's version is the literal `v1`, not a real build
  version. No `{{ version }}` template namespace exists in the engine; adding one is a shared-runtime
  change outside this connector-scoped slice. The header conforms to Reddit's required shape today,
  but the version will not track releases until that namespace exists.
- **Undocumented comments route retained.** `GET /r/{subreddit}/comments` is live and is what the
  legacy connector always called, but it is not documented by Reddit and could change without
  notice. Migrating to the documented per-article tree needs array-index support in `records.path`
  and a fan-out budget answer; both are deferred.
- **No token refresh.** `access_token` stays caller-supplied and expires after 1 hour. Refresh
  support needs a new engine auth mode. Scheduled or long-running reddit syncs will fail with 401
  until that lands; the limit is now documented in three places rather than silent.
- **Reddit commercial-terms / data-retention decisions remain open** and are untouched by this
  slice; all three defects fixed here are independent of that thread.
- **Prior-run classification accepted.** The 4th finding from the earlier review round,
  `user-agent-conformance-unenforced` (info/no-op), is accepted as classified; no action taken.
- Merge stays human-gated; never pushed to `main`.

## Safety notes

- No secrets requested, printed, summarized, or stored. The new test uses the synthetic literal
  `synthetic-conformance-secret` against a local `httptest` server.
- No live Reddit calls, no credentialed connector checks, no reverse ETL execution.
- No dependencies added.
- No branch protection or repository settings mutation.
- `reddit_username` is explicitly documented as **not** a secret — it is a public identity component
  Reddit's own API rules require in the User-Agent — so it is a plain config field rather than an
  `x-secret` one.
