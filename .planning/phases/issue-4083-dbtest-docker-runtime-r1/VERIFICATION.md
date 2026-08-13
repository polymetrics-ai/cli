# Verification — Issue #4083

Status: planned.

## Required evidence

- [x] Foundation owner issue #4083 created rather than expanding #3976's
  PostgreSQL connector lane with shared test infrastructure.
- [x] GSD adapter and command sources resolved; inline/manual lifecycle
  fallback recorded because role spawning is forbidden.
- [x] Required skills loaded.
- [ ] Red test evidence recorded.
- [ ] Green focused unit evidence recorded.
- [ ] dbtest and MySQL non-live package regression evidence recorded.
- [ ] Docker live proof recorded, or direct local daemon unavailability stated.
- [ ] Podman live proof recorded, or direct local daemon unavailability stated.
- [ ] README/AGENTS guidance checked for exact environment and safety wording.
- [ ] Inline GSD verify-work and code-review evidence recorded.
- [ ] no-mistakes PR/CI evidence recorded; no merge performed.

## External-endpoint verdict

Rejected for this issue. An external PostgreSQL server is neither a Docker nor
Podman test target and would weaken the harness's image/version, settings,
extension, storage, cleanliness, ownership, and cleanup guarantees. It needs
its own opt-in test contract if ever required.
