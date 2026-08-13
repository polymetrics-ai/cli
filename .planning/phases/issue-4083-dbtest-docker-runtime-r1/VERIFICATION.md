# Verification — Issue #4083

Status: planned.

## Required evidence

- [x] Foundation owner issue #4083 created rather than expanding #3976's
  PostgreSQL connector lane with shared test infrastructure.
- [x] GSD adapter and command sources resolved; inline/manual lifecycle
  fallback recorded because role spawning is forbidden.
- [x] Required skills loaded.
- [x] Red test evidence recorded: focused dbtest selector exited 1 against the
  hard-wired Podman baseline because `New` accepted no explicit runtime.
- [x] Green focused unit evidence recorded: explicit runtime values, endpoint
  pinning, shared unsafe-endpoint refusals, Docker identity/capacity proof, and
  Docker parser rejection cases pass.
- [x] dbtest and MySQL non-live package regression evidence recorded.
- [x] Docker live proof unavailable: the explicitly supplied direct Docker
  socket did not report a reachable daemon identity/image-store path; the
  precise tagged command and failure are in `TDD-LEDGER.md`.
- [x] Podman live proof unavailable: client installed, but no direct local
  Podman socket was present in the checked local paths; no global default was
  queried.
- [x] README/AGENTS guidance checked for exact runtime/environment and safety
  wording.
- [ ] Inline GSD verify-work and code-review evidence recorded.
- [ ] no-mistakes PR/CI evidence recorded; no merge performed.

## External-endpoint verdict

Rejected for this issue. An external PostgreSQL server is neither a Docker nor
Podman test target and would weaken the harness's image/version, settings,
extension, storage, cleanliness, ownership, and cleanup guarantees. It needs
its own opt-in test contract if ever required.
