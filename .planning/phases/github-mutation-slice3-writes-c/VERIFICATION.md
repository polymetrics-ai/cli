# GitHub mutation slice 3 verification

## Planned checks

- `go run ./cmd/connectorgen certification-matrix --check`
- `go test -timeout 20m ./cmd/connectorgen -run '^TestCertification' -count=1`
- `git diff --check`
- Repository verification sub-gates listed in `AGENTS.md`, scoped to the evidence-only change.
- `gh-axi` API read-back of the opened PR base: `integration/4015-mvp-flat-r1`.

## Live acceptance rule

No mutation is marked certified unless independent provider read-back proves both the requested effect and subsequent direct-provider cleanup. Non-certified commands retain their individual specified outcome bucket.
