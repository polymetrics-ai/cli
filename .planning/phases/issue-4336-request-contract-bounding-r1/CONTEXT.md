# Context — issue 4336 request-contract bounding

## Task Delivery Header

- Issue: Refs #4336 — represent valid provider request contracts under finite
  PM execution envelopes.
- Base branch: `main` at required launch base `8127de418`.
- Merges into: `main`.
- Delivery: direct pull request open against `main`, with the GitHub API-reported
  base verified after opening; the parent merge remains human-gated.
- Working branch: `fm/cli-request-contract-bounding-design-r1`.
- Task: Implement the accepted bounding policy from
  `data/cli-request-contract-bounding-design-r1/report.md` without provider
  exceptions, schema falsification, truncation, or removal of finite resource
  limits.
- Verification: behavioral real-importer and runtime tests, deliberate bounding
  sabotage proof, focused package/CLI tests, build/vet, generator gates, and
  inline GSD verify/code-review evidence. Test commands use `GOFLAGS=-p=3` and
  no more than one heavy suite at a time.

## Decided policy

The provider schema and PM resource policy are different contracts. The source
descriptor retains the provider schema unchanged and records a separate,
versioned execution envelope. Missing optional provider maxima on common scalar
or collection inputs are represented with that PM envelope, not called malformed
and not emitted as thousands of gaps. Semantic/serialization ambiguity stays a
source-traced merge-blocking gap. Malformed paths and retained-media budget
overruns remain quarantined operation-local gaps under the separate #4339 work;
this branch does not duplicate that open PR.

Generation must prove an executable input has a finite envelope. Runtime must
enforce the same effective bound at the actual encoded/structured boundary
before authentication or network I/O. Rejection names PM and the measured unit;
values are never truncated or coerced.

## Compatibility boundary

- Existing path/query default: 4 KiB exact encoded bytes; hard ceiling 64 KiB.
- Existing named JSON flag cap: 1 MiB before decode.
- Existing projection defaults (8 KiB string code points, 256 items/properties)
  stay unchanged in this first provenance slice so already-working commands do
  not silently tighten or broaden.
- The proposed unbounded-header 4 KiB default is not enabled until the shipped
  header census proves it non-breaking; current bounded headers and the 16 KiB
  hard ceiling remain unchanged.
- Exact numeric parsing may newly accept finite values outside `int64`/`float64`,
  but the original lexeme must reach the wire unchanged and remain byte-bounded.

## GSD execution note

The official adapter prompts for `discuss-phase`, `plan-phase --tdd`,
`execute-phase`, `verify-work`, and `code-review` are executed inline. This task
is outside the numeric roadmap and the repository's single-worker contract
forbids spawning GSD roles in this lane.
