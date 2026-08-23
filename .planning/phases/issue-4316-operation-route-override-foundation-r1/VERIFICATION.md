# Verification — Issue 4316

## Checklist

- [ ] Resolver carries operation declaration source trace, base, version, and route without caller-controlled URLs.
- [ ] Missing routing returns an explicit missing-foundation diagnostic before any provider I/O.
- [ ] Conflicting declaration bases return the same diagnostic before provider I/O.
- [ ] Direct read, direct write, binary download, binary upload, ETL, and reverse ETL all compose their provider request through the shared resolver.
- [ ] Five real Help Scout v3 direct-read operations pass their behavioral URL and pagination assertions.
- [ ] Source trace, canonical mapping, generated surface, fixtures, and conformance evidence are current.
- [ ] No edit to `internal/connectors/defs/github/rate_limits.json`.
- [ ] CLI/manual/website applicability documented and generated output checked.
- [ ] Targeted package tests, individual verification gates, build, GSD verification, and code review evidence recorded.
