# Verification — Issue #4090

## Acceptance checklist

- [ ] PostgreSQL's declared source preflights through the real transport registry.
- [ ] `full_overwrite` and `full_append` emit bounded typed records and a valid checkpoint.
- [ ] A missing descriptor refuses before I/O.
- [ ] A wrong executor family refuses before I/O.
- [ ] An unregistered declared executor refuses before I/O.
- [ ] Focused unit and package race checks pass.
- [ ] The PostgreSQL dbtest harness passes against an explicit Docker-or-Podman
      Unix endpoint and its output shows rows, emitted identity/schema, and checkpoint.
- [ ] Required individual repository gates pass.
- [ ] Manual inline `verify-work` and `code-review` records are complete.

## Safety audit

- [ ] No secret value appears in test output, traces, docs, commit, or PR body.
- [ ] No raw SQL/configured SQL path or direct source-to-destination route exists.
- [ ] `internal/connectors/engine/bundle.go` is unchanged.
- [ ] Changed paths remain PostgreSQL-only plus planning evidence.
