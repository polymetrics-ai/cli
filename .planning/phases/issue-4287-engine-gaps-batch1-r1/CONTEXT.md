# Context: Issue #4287

The issue body, the batch-1 evidence brief, and the current connector canon
lock this foundation increment. The launcher requires autonomous execution, so
the manual discussion conclusion adopts those sources without further product
questions.

## Phase Boundary

Add five definition-selected engine capabilities: per-operation REST
pagination, a typed response-less HEAD status operation, a closed set of
additional JSON-family write content types, bounded text export, and a live
secret-response write path that redacts output and routes the secret to the OS
keychain. This is shared engine work, not a Docker Hub definition change.

## Implementation Decisions

- Per-operation pagination takes priority. Its declaration is source-derived,
  supersedes the bundle-global direct-read paginator only for its operation,
  and rejects source/declaration disagreement before any request.
- A response-less HEAD operation uses a separate typed executor and output
  contract; it cannot select the JSON direct-read executor by method alone.
- Extra write media types use a finite JSON-family allowlist. Unknown media
  types are rejected during declaration/preflight before transport I/O.
- Text export is a separate bounded contract. It always requires an explicit
  positive byte cap and rejects a missing cap before opening a request.
- Sensitive operations retain their existing non-inline input and explicit
  approval rules. A secret-typed response is never emitted, logged, or
  materialized; only the supported OS keychain destination may receive it.
- Existing bundle-global pagination and existing connectors retain their
  behavior. No connector-name branches or generic raw HTTP/write escape hatches
  are permitted.

## Canonical References

- `AGENTS.md` — delivery, connector-boundary, direct-read, secrets, and local
  verification rules.
- `docs/connector-canon/IMPLEMENTATION-PROCEDURE.md` §2 and §6 — foundation
  check and real runtime preflight requirements.
- `docs/migration/conventions.md` §2.9 — derived command parameter and paging
  rules.
- `docs/architecture/connector-architecture-v2-design.md` — declarative engine
  boundaries.
- `internal/connectors/defs/dockerhub/sources/dockerhub-declaration-disposition.json`
  on `fm/cli-top100-declaration-batch-r1` — batch proof target; owned by the
  connector-sweep lane and not edited here.
- `/Users/karthiksivadas/karthik-agent-workspace/data/ENGINE-GAPS-BATCH1.md` —
  authoritative gap evidence supplied by the launch brief.

## Manual-GSD Fallback

`scripts/gsd doctor`, all five required `sources` commands, the five generated
GSD prompts, and `go run ./cmd/agentcontractgen check` were run. The adapter
cannot resolve issue `4287` as a roadmap phase (`phase_found: false`), and this
task forbids role spawning, so the lifecycle is carried out inline in this
issue-scoped phase directory. This is a recorded fallback, not a waiver.
