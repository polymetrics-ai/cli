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
The initial accounting was superseded by the read/auth follow-up: 32 operations are
now exercised (4 worked, 28 failed), and 22 mutation operations remain untestable
with `Named dependency:` reasons. `VERIFICATION.md` carries the complete row-level
matrix; the unresolved external work is an account-permission result, not an
unverified command route.

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

## Live read/auth coverage follow-up (2026-08-08)

With the write-scoped token held by the captain, the supplied read credential was
used immediately rather than idling. The rebuilt binary attempted every documented
read operation (28, including the three status-only HEAD checks) and all three
authentication exchanges through their normal plan → preview → approval lifecycle.
The exact redacted row-level outcomes are recorded in `VERIFICATION.md`:

- read/auth vector: **31/31 attempted** — 4 worked and 27 failed explicitly;
- including the earlier approved repository-create request: **32 exercised** — 4
  worked and 28 failed; **22** account/SCIM/repository writes remain untestable;
- the SCIM-only credential contains `scim_bearer_token` and no `docker_pat`. Its
  seven command paths returned explicit HTTP 401 outcomes (one canonical schema URN
  first hit `path variable id contains invalid character ':'`; a URL-safe sentinel
  then received HTTP 401). This is an empirical non-silent outcome only; it does not
  change the SCIM classification.
- a fresh `access-tokens list` diagnostic returned a concrete redacted
  `http 403 for https://hub.docker.com/v2/access-tokens` error, confirming that the
  permission rejections surface cleanly rather than failing silently.

The three auth actions deliberately used invalid **non-secret** fixtures, never a
real credential in a raw CLI argument. Their plan and preview stages passed; each
live exchange returned explicit HTTP 401. This proves the action routes and approval
flow, not a successful credential exchange.

### CLI/doc parity disposition

No command, flag, help text, output shape, generated manual, or website catalog
contract changes: this corrects the bundle-internal execution target behind existing
commands. Runtime help/manual/website regeneration is intentionally not applicable;
the rebuilt binary's existing command help remains rechecked as part of live reachability.

## Namespace override gap — live write E2E prerequisite (2026-08-08)

Captain-directed live coverage proved a second definition-owned behavior defect before the
write-scoped token was used: an accepted `--config namespace=library` was silently discarded.
`streams.json` (the `repositories`, `tags`, `repository_detail`, and `tag_detail` ETL streams)
and the base health check all interpolated `config.docker_username`, so they issued requests for
the authentication identity (`polymetrics`) rather than the requested target namespace. The
result was a misleading Docker Hub 404, not an operator error.

**Decision.** `docker_username` remains the Docker Hub login identity used only by the custom
authentication hook. Introduce a distinct, required `namespace` configuration property for the
target Hub namespace and interpolate it in every stream path and the health check. The declarative
engine supports static schema defaults only; it has no safe cross-key "namespace defaults to
docker_username" facility. The schema declares `namespace` required and the stream/check template
fails locally before issuing HTTP when it is absent, avoiding another accepted-then-discarded
configuration value. The credential store intentionally validates only flat-map constraints, not
JSON Schema `required`; changing that project-wide behavior is not necessary to repair this
connector-local route and is outside this slice. No shared engine or auth-hook behavior changes.

**GSD inline fallback.** Before this slice, I ran `scripts/gsd doctor`, resolved `discuss-phase`,
`plan-phase --gaps --tdd`, `execute-phase --gaps-only`, `verify-work`, and `code-review` with
`scripts/gsd sources`, and inspected each generated prompt. The parent-worker contract forbids
role spawning, so this is recorded as an inline manual execution of the same lifecycle. Required
skills loaded through the project routing file: `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`,
`golang-structs-interfaces`, and `golang-documentation`.

### TDD gap slice

1. **Red:** add `internal/connectors/defs/dockerhub/namespace_override_test.go`. Through the real
   `engine.Read` and `engine.Check` paths against an `httptest` server, pass distinct
   `docker_username=auth-identity` and `namespace=target-namespace` configuration values. Assert
   each of the four stream request paths plus the base check route use `target-namespace`, never
   `auth-identity`. Run the focused test and commit its verbatim failure before touching production
   JSON.
2. **Green:** define `namespace` as a required Docker Hub configuration property, replace all four
   stream path templates and the health-check template with `config.namespace`, regenerate only
   derivable artifacts, and update generated connector documentation if its generator reports a
   Docker Hub change.
3. **Verification:** run the focused package test, Docker Hub validation, surface-sync check,
   executable command preflight, scoped build/help checks, and the rebuilt binary against the live
   account. Re-add the local credential via `--from-env` with both `docker_username=polymetrics`
   and `namespace=polymetrics`, never logging the PAT. Create a unique private repository, prove
   privacy before image content, push a minimal OCI image through Docker's registry protocol, then
   read tags and repository detail through `pm`. Do not delete that repository; clean up only
   uniquely created test resources that are safe to remove.

### CLI/doc parity disposition

The command surface itself is unchanged; this corrects a connector configuration contract behind
existing commands. The required `namespace` property changes generated configuration documentation,
so inspect the Docker Hub manual/skill output and website catalog generator result. Recheck
`pm dockerhub`, the affected command help, `pm help dockerhub`, and docs/website generator scope.
No page, `per_page`, or `limit` flags are authored; stream pagination remains declaration-derived.

**Green result.** The focused real-engine regression test passed after the template repair; it
also proved an omitted namespace errors locally before HTTP. Docker Hub validation, surface-sync,
the fleet-wide implemented-command preflight, and connector boundary gate passed. `pm docs
generate --dir docs/cli` was run from the rebuilt binary; 1,027 unrelated generated-documentation
changes were restored, while the Docker Hub manual/skill and two parsed-object-verified Docker Hub
website catalog entries remain. No golden transcript changed.

## Required skills used

Loaded via `.agents/agentic-delivery/references/required-skills-routing.md`:
`golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`,
`golang-security`, `golang-safety`, `golang-design-patterns`,
`golang-structs-interfaces`, and `golang-troubleshooting`. GSD workflow guidance
used: `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, and
`gsd-code-review`; the canonical single-worker rule required the recorded inline
fallback rather than role spawning.

## Captain rate-limit gap — plan (2026-08-08)

**Discussion and decision.** The captain's provider facts distinguish two mechanisms:
the documented Registry pull quota (unauthenticated 100, authenticated free 200,
fixed 21,600-second documentation window, paid unlimited) and a separate, unnumbered
Hub API abuse limiter that returns bare HTTP 429. The 54 documented Docker Hub
management operations use `hub.docker.com`; the pull quota is explicitly for
`registry-1.docker.io`. It would be false to attach the pull quota to `hub.docker.com`
just because both products carry the Docker name.

The existing requester/registry foundation already admits every engine send and observes
every response; this slice must **not** build another limiter. Its one missing expression
for Docker's documented asymmetry is an exact host selector. Add that narrow selector
dimension, then declare Registry-only pull policies. The Hub API abuse limiter has no
published numeric budget, so it remains intentionally unbudgeted; the existing generic
Requester 429 path honors a provider `Retry-After` rather than inventing a threshold.

**Planned declared policies.** Both cite Docker's official pull-limit documentation and
use a non-secret scope only: unauthenticated Registry pulls are keyed by a user-supplied
public IP/IPv6-/64 config property, while authenticated personal-free pulls use the
non-secret `docker_username` account identity. `selector.hosts` is exactly
`registry-1.docker.io`; `selector.auth_types` separates unauthenticated and authenticated
traffic; `selector.tiers` limits the 200 policy to `free`. Paid tiers deliberately match
no fixed budget because Docker documents them as unlimited. No Hub API operation matches
either policy.

**Live note, retained rather than normalized away.** The captain-supplied documented
window is 21,600 seconds. A free authenticated-less HEAD probe observed
`ratelimit-limit: 100;w=3600` and `ratelimit-remaining: 100;w=3600` on 2026-08-08.
The declaration remains conservatively documented at 21,600 seconds (which cannot
overshoot Docker); the response parser will retain the leading numeric budget from the
semicolon-form headers without pretending the inline window is a reset timestamp. The
contradiction is a provider/documentation friction point for the pilot, not silently
rewritten evidence.

### TDD gap slice

1. **Red:** add a Docker Hub declaration test and a requester test for Docker's
   `ratelimit-limit: 200;w=21600` / `ratelimit-remaining: 199;w=21600` syntax. Run both
   before adding a production declaration or changing runtime parsing; capture the
   missing-declaration and non-parsed-header failures verbatim.
2. **Green:** add the exact-host selector to the existing declared-policy resolver and
   its schema/overlap validation, parse the standard leading numeric rate-limit value
   before semicolon parameters, add `dockerhub/rate_limits.json`, embed the first
   production declaration, and add non-secret profile/scope config documentation.
3. **Proof:** inject a short fixed-window budget in the Docker Hub bundle test and run
   the real built `pm` through a local HTTP proxy whose requested authority is
   `registry-1.docker.io`; prove the proxy receives exactly the configured number of
   pages and no final over-budget send. Then use Docker's free authenticated Registry
   HEAD probe without printing its transient bearer, and record the provider's remaining
   quota. This proves local admission stops before a Registry pull quota is consumed.
4. **Verification:** focused connsdk/engine/Docker Hub tests, definition validation,
   fleet preflight, surface-sync, connector boundary, rebuilt-binary help/docs checks,
   and inline verify-work/code-review. Required skills remain `golang-how-to`,
   `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`,
   `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and
   `golang-documentation`.

**GSD inline fallback.** `scripts/gsd doctor`, all five `scripts/gsd sources` lookups,
their generated prompts, and `go run ./cmd/agentcontractgen check` were run before this
slice. The canonical parent-worker contract forbids the official role spawning, so
discussion, `plan-phase --tdd`, execution, verification, and review are recorded inline.

### Completion record

The red checkpoint was committed before production changes; the green declaration
and tests were committed and pushed separately. The built binary proof stopped at
100 proxy-observed requests before dispatching its 101st logical request, then the
free Docker Registry HEAD reported full remaining headroom. Final scoped verification,
full `internal/cli` regression coverage, and inline `verify-work`/`code-review` pass;
their commands and outcomes are recorded in `TDD-LEDGER.md`, `VERIFICATION.md`, and
`REVIEW.md`.
