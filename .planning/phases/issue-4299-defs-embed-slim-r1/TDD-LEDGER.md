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

Pending. Production code has not yet changed. The first red run will be recorded verbatim here before the embed directive is changed.

## Green evidence

Pending implementation.

## Refactor/review guard

The final review must prove that the exception stays a literal path, no compression/minification or checkout-root support enters the diff, and no connector definition changed.
