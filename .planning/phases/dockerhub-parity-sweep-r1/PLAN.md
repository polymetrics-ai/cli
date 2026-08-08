# Docker Hub documented-operation parity — plan

Pilot of a 22-connector documented-operation-parity programme. One connector, one PR, scoped to
parent issue #3905 and its 9 category sub-issues (#3906-#3914).

## Operation surface, re-verified before authoring

- **Artifact**: `https://docs.docker.com/reference/api/hub/latest.yaml` (Docker Hub API v2-beta,
  OpenAPI 3.0.3)
- **Retrieved**: 2026-08-08, **148,322 bytes**, sha256
  `99d9d53c2d93656a3c66d604885abd153dc5df285abc0ecb13802a3bc53d0756`
- **Documented operations: 54** (35 unique paths)
- **By method**: GET 24, POST 12, DELETE 6, PATCH 5, PUT 4, HEAD 3

### Against the provider-artifact ledger

| | |
| --- | ---: |
| Ledger recorded (`api_surface.json`, pre-phase) | 54 |
| Re-derived from live artifact | 54 |
| Delta | 0 |
| sha256 match | exact |

**How it was counted.** Fetched the artifact fresh via `curl`, hashed it, and independently parsed
every `paths.*.{get,post,put,patch,delete,head}` entry with a Python/PyYAML script — 54 unique
(method, path) pairs, zero internal collisions, exact match against both the recorded count and the
recorded sha256 in the pre-existing `api_surface.json`. This is a rare case where the provider
ledger was already correct; re-derivation proved that rather than assuming it (AGENTS.md: "the
provider ledger has been wrong SIX times in both directions").

## Baseline (before this phase)

`internal/connectors/defs/dockerhub/` was NOT a blank slate: a complete, reasoned
`operation_ledger_version: 2` `api_surface.json` already existed, covering all 54 operations — 4
`covered_by` the existing `repositories`/`repository_detail`/`tags`/`tag_detail` ETL streams, and
**50 rows with `operation.model: "disallowed"`**, each with a `reason` but no `notes`/named
dependency. The structural cause: `spec.json` declared no credential field and `streams.json`
declared `"auth": [{"mode": "none"}]` — the bundle was authored as a deliberately credential-free,
public-registry-read connector, so every account-scoped operation was blocked for want of any
authentication path at all.

## Captain rulings (mid-phase, binding)

1. Zero `unsafe_or_disallowed`/`disallowed` rows. Every not-yet-implementable operation carries a
   real `operation.model` plus a `notes: "named_dependency=..."` marker.
2. The 3 auth-exchange endpoints (`POST /v2/auth/token`, `/v2/users/2fa-login`, `/v2/users/login`):
   implement as real commands with redacted token output, not internal-only blocks.
3. The 3 HEAD existence-check endpoints: a status-only response is still a real action (same logic
   as a write action that returns 204 No Content) — implement them, asserting on status code, not
   payload.
4. The 9 SCIM endpoints, blocked on a distinct `bearerSCIMAuth` scheme: a genuine missing
   foundation, not a danger judgment — but a named dependency is no longer an automatic deferral.
   **Build the missing foundation inside this issue** so the connector ships complete, with the
   foundation in its own commits, separate from connector-authoring commits, and the PR body
   disclosing exactly what the foundation is and which other connectors it unblocks.

Net effect: target is **54 of 54 implemented**, zero blocked rows, two runtime foundations built
inline (a connector-scoped dual-credential auth hook for SCIM, and a shared-runtime HEAD-method
capability for status-only existence checks).

## Foundations built (their own commits, shared-runtime scope disclosed)

1. **`internal/connectors/hooks/dockerhub/` (Tier-2 AuthHook, connector-scoped only)**: Docker Hub's
   API does not accept a PAT directly as a bearer token — it requires exchanging
   `docker_username`+`docker_pat` for a short-lived session JWT via `POST /v2/users/login`
   (non-OAuth2-shaped, so none of the engine's declarative auth modes fit). The hook also
   implements `dualAuth`, routing `/v2/scim/2.0/**` requests to a second, independently-configured
   `scim_bearer_token` credential instead of the session JWT, since Docker Hub's own OpenAPI
   document declares SCIM under a distinct `bearerSCIMAuth` scheme. This foundation is entirely
   connector-local — no shared engine file changes.
2. **HEAD-method direct-read support (shared runtime, affects the mirrored method-allowlist in 6
   files)**: no code path in `internal/connectors/commandrunner/runner.go` previously admitted
   `http.MethodHead` for any executor kind. Added a status-only branch
   (`internal/connectors/engine/direct_read.go`: a HEAD response never carries a body per RFC 9110
   §9.3.2, so the result is `{"status_code": N}`) and mirrored the method-allowlist change across
   every site enforcing the old GET/POST-only rule: `commandrunner/runner.go`, `engine/bundle.go`
   (load-time validation), `engine/operation_endpoint_ledger.go` (both the ledger *deriver* and the
   persisted-ledger *loader* — the loader is what the real `pm` binary reads),
   `cmd/connectorgen/validate.go` (3 sites), `internal/connectors/conformance/static.go`. This is
   purely additive (HEAD was previously rejected everywhere, so no currently-passing command's
   behavior changes) and is now available to any other connector with a documented HEAD-only
   existence-check endpoint — verified fleet-safe by the unchanged pass of
   `TestEveryImplementedCommandPassesRuntimePreflight` (all 551 connectors' implemented commands)
   and the whole `internal/connectors/conformance` suite.

## Required scopes — every operation lands somewhere real

Every documented operation is an ETL stream, a reverse-ETL write, a direct read (including the new
status-only HEAD checks), or a binary download, and is individually reachable as its own
`pm dockerhub <command>`. Every `api_surface.json` row carries exactly one of `covered_by` or
`operation` (blocked + `named_dependency=` note) — final state: 54 `covered_by`, 0 `operation`.

## Issues

Parent **#3905**; children **#3906-#3914** (Repositories & Tags, Personal Access Tokens,
Organization Access Tokens, Audit Logs, Authentication, Groups/Teams, Invites, Organizations &
Settings, SCIM). Single PR scoped to the parent, closing the parent and all 9 children.

## TDD sequence

1. **RED** — `cmd/connectorgen/dockerhub_api_surface_test.go`, written and run against the genuine
   disk-committed pre-phase state (verified via `git status` immediately before running — no
   in-progress bundle edits existed yet, so no stash was needed). Captured failure: 50 rows with
   `operation.model == "disallowed"`, `covered=4`/`blocked=50` against an initial target of
   `covered=39`/`blocked=15`.
2. Bundle authoring proceeded in stages as the captain's rulings arrived; the red test's target
   constants were updated at each stage (39/15 → 51/3 → 54/0) to track the evolving, more complete
   classification — each stage re-run and reconfirmed green before the next.
3. **GREEN** — final state: 54 covered, 0 blocked.
4. Foundation code (hook package + engine/commandrunner/connectorgen/conformance changes) got its
   own dedicated unit tests (engine: 5 new HEAD tests; hooks/dockerhub: 22 tests including 4
   dualAuth/SCIM-specific).
5. Gates, then no-mistakes.

## Safety notes

- Did not loosen `connectorgen validate`, the connector boundary gate, or
  `TestEveryImplementedCommandPassesRuntimePreflight` to make anything pass.
- Nothing is marked `implemented` unless its command genuinely runs — verified by running the real
  built `pm` binary for all 54 `--help` invocations, plus a genuine live loopback HTTP round trip
  proving the new HEAD status-only path (`repository check` against a local test server, HEAD
  request observed server-side, `{"status_code": 200}` returned end-to-end).
- No credential or token-derived value is ever emitted; access-token/SCIM-user create/update
  commands declare `redact_fields` on both the write action and the CLI command.
- Regenerating docs rewrites ~1,027 files of pre-existing `main` drift each pass — reverted every
  non-dockerhub path every time.
- Website catalog diffs inspected by object (Python dict comparison), not by line.

## Live E2E gap closure — reverse-ETL path resolution (2026-08-08)

Captain-authorized live testing reached the connector-command plan/preview gate for
`repository create` without dispatching a mutation. The preview recorded the unsafe
request line `POST https://hub.docker.com/v2/v2/namespaces/{namespace}/repositories`:
the bundle's `writes.json` copied OpenAPI's provider-rooted `/v2/...` paths even
though this bundle's configured engine base URL already ends in `/v2`, and it left
OpenAPI `{parameter}` placeholders where the write engine requires
`{{ record.parameter }}` templates. This is a real executable-surface defect, not
a live-account limitation; no repository was created by the failed attempt.

**Scope:** `internal/connectors/defs/dockerhub/writes.json`, the Docker Hub
definition-owned regression test, the shared strict-write ALPN fix and its isolated
`connsdk` regression test, generated artifacts only if their checks report drift,
and this phase's evidence. The new shared-runtime scope was discovered only after
the corrected request hit Docker Hub's HTTP/2 transport; it is a provider-neutral
one-shot-write safety fix, not a Docker Hub-specific workaround.

**GSD inline fallback:** the project canonical single-worker contract forbids
spawning planner/executor/reviewer roles. The gap workflow is therefore run inline
after `scripts/gsd doctor` and `scripts/gsd sources plan-phase|execute-phase|
verify-work|code-review`, with the generated command prompts recorded in this plan
and the ledger.

### TDD gap slice

1. **Red:** add `internal/connectors/defs/dockerhub/write_paths_test.go`. Load the
   embedded Docker Hub bundle and assert every reverse-ETL action is engine-relative
   (no leading `/v2/`), has no raw OpenAPI `{parameter}` token, and declares every
   record-derived path segment in `path_fields`. Dry-run `create_repository` with
   `base_url=https://hub.docker.com/v2` and assert the fully resolved request is
   exactly `POST https://hub.docker.com/v2/namespaces/polymetrics/<test-repo>`.
   Run the focused test and capture its failure before changing production JSON.
2. **Green:** translate every Docker Hub `writes.json` action from a provider-rooted
   path to its engine-relative equivalent; replace each raw path parameter with the
   matching `{{ record.<field> }}` template; add `namespace` to
   `create_repository.path_fields` so it does not leak into the JSON body.
3. **Verification:** run the focused test, Docker Hub bundle validation, command
   preflight sweep, surface-sync check, connector-boundary gate, scoped tests/build,
   and a rebuilt-binary plan/preview/live execution. Confirm the live-created
   repository is private before Docker image push. Record every live result without
   secret, approval-token, or token-derived values.

**Outcome:** the repaired request and strict-write transport reached the provider,
but the supplied PAT returned HTTP 403 to the single approved private-repository
create. Account listing is empty, so the image-push/readback chain cannot proceed.
The full 54-operation live accounting (23 exercised: 4 worked, 19 failed; 31
untestable with `Named dependency:` reasons) is recorded in `VERIFICATION.md`; this
is an account-permission result, not an unverified command route.

### Strict-write transport TDD follow-up

The first post-path-fix approved execution produced a malformed HTTP/1 response
whose bytes identify an HTTP/2 SETTINGS frame. The plan record and config were
checked mechanically and contain no trailing quote or base URL override. The
provider-neutral reproduction is an `httptest` TLS server advertising HTTP/2:

1. **Red:** `TestRequesterDisableRetriesUsesHTTP1WithHTTP2CapableServer` fails when
   `noReplayClient` pins `Transport.Protocols` but retains a caller TLS ALPN list
   advertising `h2`.
2. **Green:** clone the transport's TLS config and advertise only `http/1.1` for a
   strict one-shot mutation, retaining its no-keepalive/no-redirect/no-replay
   behavior and every unrelated TLS setting.
3. **Verification:** focused connsdk test plus package suite; rebuild `pm`; create
   a new private Docker Hub repository plan/preview/approval/run once, then verify
   the response using documented read operations before progressing through the
   remaining live-safe operation matrix.

### CLI/doc parity disposition

No command, flag, help text, output shape, generated manual, or website catalog
contract changes: this corrects the bundle-internal execution target behind existing
commands. Runtime help/manual/website regeneration is intentionally not applicable;
the rebuilt binary's existing command help remains rechecked as part of live reachability.

## Required skills used

Loaded via `.agents/agentic-delivery/references/required-skills-routing.md`:
`golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`,
`golang-security`, `golang-safety`, `golang-design-patterns`,
`golang-structs-interfaces`, and `golang-troubleshooting`. GSD workflow guidance
used: `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, and
`gsd-code-review`; the canonical single-worker rule required the recorded inline
fallback rather than role spawning.
