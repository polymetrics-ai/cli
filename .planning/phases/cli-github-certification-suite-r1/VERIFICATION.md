# Verification checklist — GitHub certification suite r1

- [ ] Fresh red test for exhaustive surface-derived sweep generation is recorded.
- [ ] Green command test proves 1,571 unique commands each receive exactly one status/reason and totals reconcile.
- [ ] Generated GitHub sweep passes its deterministic `--check` command twice.
- [ ] Product defects include `releases assets view`; provider refusals are a separate non-pass type.
- [ ] Scratch post-schema assertion sabotage demonstrably makes its own certified operation red; source is restored and green rerun recorded.
- [ ] Any live run is serial, resumable, uses only the disposable identity through an environment variable, asserts produced values, and leaves no owned resource residue.
- [ ] No accepted evidence is emitted while #4198 is open; the blocked reason is recorded.
- [ ] `go test -timeout 20m ./cmd/connectorgen` passes as the required consumer package.
- [ ] Changed-package and `internal/cli` tests, `go vet`, `gofmt`, generator checks, docs checks, lint, boundary, and release checks are run individually and results are recorded before push.
- [ ] Inline code review has no unresolved actionable finding.
- [ ] Opened PR base is read back from the API and equals `integration/4015-mvp-flat-r1`.
