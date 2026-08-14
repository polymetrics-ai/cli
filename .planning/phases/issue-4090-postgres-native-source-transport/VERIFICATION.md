# Verification — Issue #4090

## Acceptance checklist

- [x] PostgreSQL's declared source preflights through the real transport registry.
- [x] `full_overwrite` and `full_append` emit bounded typed records and a valid checkpoint.
- [x] A missing descriptor refuses before I/O.
- [x] A wrong executor family refuses before I/O.
- [x] An unregistered declared executor refuses before I/O.
- [ ] Focused unit and package race checks pass.
- [x] The PostgreSQL dbtest harness passes against an explicit Docker-or-Podman
      Unix endpoint and its output shows rows, emitted identity/schema, and checkpoint.
- [ ] Required individual repository gates pass.
- [ ] Manual inline `verify-work` and `code-review` records are complete.

## Safety audit

- [x] No secret value appears in test output, traces, docs, commit, or PR body.
- [x] No raw SQL/configured SQL path or direct source-to-destination route exists.
- [x] `internal/connectors/engine/bundle.go` is unchanged.
- [x] Changed paths remain PostgreSQL-only plus planning evidence.
