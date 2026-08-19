# #4069 correction 2 / 5 — flow manual generator parity plan

- Canonical finish-plan SHA-256:
  `939f14f61defd993f8ad0335a5aeb617d97083c9f73a6a75259d0e312ae8f408`
- Exact failing head: `678e294568a8a010a460ecb05fe11a42e1eb40f2`
- Exact GitHub run: `31590254616` (`Verify`)
- Causal test: `internal/cli.TestGoldenDocsGenerateMatchesTrackedCLIManuals`
- Failure: `generated docs drift for flow.md`

The tracked manual contains this approved fail-closed wording:

```text
  A case-equivalent spelling whose owner cannot be decided also fails closed;
  set "connection" to a known healthy owner rather than relying on an
  unscoped query.
```

At the failing head, `internal/cli/docs.go`'s authoritative `flowHelp` omitted
those lines, so `pm docs generate` could not reproduce the tracked manual.
Correction 2 adds the wording to `flowHelp`, invokes the checked-in manual and
golden transcript generators, runs the existing website data generator to
prove dependent output, and preserves all runtime behavior, delivery topology,
and prior commits.
