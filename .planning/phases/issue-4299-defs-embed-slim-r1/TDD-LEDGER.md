# TDD ledger — issue #4299 definition embed slim

## Manual-GSD record

The inline `discuss-phase` and `plan-phase --tdd` prompts were run after the captain
resolved Option A. No GSD role was spawned because the repository contract forbids it.

## Planned red/green evidence

| Contract | Red evidence | Green assertion | Status |
| --- | --- | --- | --- |
| Source locks are opt-in | A whole `*/sources/*` embed exposes non-GitHub source-lock paths in the real embedded inventory | Only `github/sources/github-operation-source-lock.json` remains in `defs.FS` | Planned |
| Inert artifacts never ship | A real inventory finds `api_surface.json` or a fixture/source artifact and fails with its path | Every embedded artifact class is allowlisted and sorted deterministically | Planned |
| GitHub exception is raw and exact | The pre-existing suite does not compare the source file to embedded bytes | Byte-for-byte equality and SHA-256 equality protect exact raw content and source provenance | Planned |
| Installed certification is self-contained | No release-like archive check proves the GraphQL lock survives outside a checkout | Extracted artifact preserves the GitHub full-certification schema boundary without a source checkout | Planned |
| Binary growth is bounded | No deterministic package report attributes embedded content or refuses a size breach | Budget report is stable, attributed, and rejects a breach before release package publication | Planned |

## Red evidence

Red: `go test ./internal/connectors/defs -run TestProductionEmbedDeclaresOnlyGithubSourceLockException -count=1`

```text
--- FAIL: TestProductionEmbedDeclaresOnlyGithubSourceLockException (0.00s)
    defs_test.go:82: source locks must be explicitly allowlisted, directive = "operation_endpoint_ledger.json */metadata.json */changefeed.json */polling_watermark.json */sync_transport.json */spec.json */streams.json */writes.json */schemas/* */sources/* */docs.md */operations.json */cli_surface.json */certification.json */rate_limits.json */database.json"
FAIL
FAIL    polymetrics.ai/internal/connectors/defs    0.870s
FAIL
```

This fails before the production directive changes and demonstrates that the
wildcard automatically admits source-lock paths instead of naming the one
installed-certification exception.

## Green evidence

Green: `go test -timeout 20m ./internal/connectors/defs ./internal/connectors/certify -count=1`

```text
ok      polymetrics.ai/internal/connectors/defs       1.860s
ok      polymetrics.ai/internal/connectors/certify    10.043s
```

The production directive now names
`github/sources/github-operation-source-lock.json` literally. `Inventory()`
walks the actual embedded filesystem, assigns stable artifact classes and byte
totals, and rejects API-surface manifests, fixtures, non-exempt source locks,
and unknown artifacts. The tests also compare the exception byte-for-byte to
the committed source lock, pin its SHA-256, and compile GitHub's full GraphQL
inventory after changing to a temporary directory outside the checkout.

Green: `scripts/tests/release-size-budget.sh`

```text
release size budget guard passed
```

The release guard reports archive and installed-binary bytes in a fixed order,
rejects an over-budget binary and malformed flags, and has a quiet mode so the
existing release-subject machine output remains a pure subject list.

## Refactor/review guard

The final review must prove that the exception stays a literal path, no compression/minification or checkout-root support enters the diff, and no connector definition changed.
