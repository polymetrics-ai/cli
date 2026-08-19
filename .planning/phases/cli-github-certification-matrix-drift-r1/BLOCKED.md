# Blocked: GitHub certification-matrix drift is absent on current `origin/main`

The task required a RED reproduction before editing generated JSON. At
`origin/main` `51dd6d468e4a40ece70c36efb81df4fdede8a8b6`, the canonical command
already passes:

```text
go run ./cmd/connectorgen certification-matrix --check
certification shards are current: connectors=3 capability_complete=0 certified=0
```

The GitHub-only canonical writer completed twice without changing
`internal/connectors/defs/github/certification-matrix.json`. Focused tests
also prove that an intentionally stale shard is rejected and that a scoped
GitHub generation does not rewrite other connector output or shared status.

`cmd/connectorgen/certificationmatrix.go` is the authority: it loads the
GitHub declaration bundle through `engine.Load`, scopes the generated runtime
endpoint ledger entry, derives capability/workflow data, serializes the
connector shard deterministically, then byte-compares it in `--check` mode.
No source or generator defect was found.

No generated JSON was edited and no pull request should be opened for a
no-op refresh. Re-dispatch only with a remote commit SHA that actually fails
the canonical check.
