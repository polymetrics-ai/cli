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
