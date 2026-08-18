# Code review — Issue #4015 GitHub declared-command parity

## Mode

Inline/manual GSD `code-review` fallback. The generated prompt was resolved and executed inline
because the repository's canonical delivery contract requires one active worker and forbids spawning
a reviewer role for this lane.

## Scope reviewed

- Exact 50-command inventory, aliases, availability transitions, flags, and fixed endpoint bindings.
- GraphQL supplemental-operation preservation and generated Node projections used by issue/PR reads.
- Reverse-ETL record schemas, opaque provider identifiers, approval metadata, and preflight behavior.
- Binary representation selection, cross-host redirect policy, credential stripping, byte caps,
  destination confinement, and archive-extraction refusal.
- Generated certification ledgers, manuals, skill, website data, help parity, and provider cleanup.

## Findings and dispositions

No critical, warning, or informational finding remains.

- **Accepted and fixed — release assets returned metadata rather than bytes.** A real 51-byte asset
  first produced 1,798 bytes of GitHub JSON because the operation had no fixed representation.
  `application/octet-stream` is now declaration-owned and regression-tested.
- **Accepted and fixed — provider redirect was undeclared.** The byte response redirects to
  `release-assets.githubusercontent.com`; the release operation permits that exact host. Actions
  artifacts use provider-generated signed storage hosts, so that fixed operation opts into the
  shared any-host redirect policy, whose transport strips credentials before the hop.
- **Accepted and fixed — provider identifiers lost exact spelling.** Live autolink deletion exposed
  JSON-number scientific notation. Autolink and workflow identifiers are opaque strings, protected
  by a provider-contract test.
- The raw API, local git/config/browser/SSH/extension workflows, binary upload, cryptographic
  verification, and multi-operation composites remain non-executable by design. Their retained
  declarations do not expose a generic escape hatch or claim provider absence.

## Review evidence

- Full tests for `connsdk`, `engine`, `commandrunner`, and `cmd/connectorgen` pass.
- The 1,571-command no-credential reachability sweep reports 1,571 reachable and zero unreachable.
- Live release download matched the fixture byte-for-byte; release/tag cleanup independently returned 404.
- `go vet ./...`, `go build ./cmd/pm`, generated-artifact checks, website gates, and `git diff --check` pass.

## PR #4236 gap review

Inline/manual GSD `code-review` was repeated for the GraphQL inventory correction. No finding
remains.

- The correction does not weaken the source lock or change its 305-root denominator.
- Duplicate transport bindings fail explicitly, the source-generated prefixes are counted
  separately, and the only permitted supplemental fixed document is exactly `github.repo.list`.
- Production surface counting remains unchanged at all 306 executable bindings.
- The direct-read expectation change is supported by an exact base-to-head declaration diff (eight
  singular-to-plural migrations and 23 added plural targets), not an unexplained count bump.
- `make lint`, focused/full certification tests, source artifact checks, runtime preflight, workflow
  evidence, vet, and whitespace validation all pass.
