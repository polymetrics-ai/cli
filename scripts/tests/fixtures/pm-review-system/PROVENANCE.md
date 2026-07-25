# PM Review Corpus Provenance

Frozen before treatment implementation on 2026-07-23.

## Visible historical seeds

- `h-71c6e8a1`, `h-84bd29f3`, `h-a10f54d9`: F6–F8 as present at
  `0665ad7aad1ec083f4bb0572a88ac1a38f417a35`.
- `a-27e19c44`, `a-c0915bd7`: two findings accepted by the captain at PR #495 source
  `fc7167990c92292625493f05b495c70e2c7ce886` for this follow-up.

## Preimplementation opaque families

A read-only scout, before detector implementation, proposed independent families for dependency
integrity, transitive route purity, stable candidate continuity, stale exact-version evidence,
schema enum drift, missing targets, coverage, context overflow, and correction-cap transitions.
The independent plan checker added untrusted path/identity and threshold-boundary cases. Opaque IDs,
renamed gates, arbitrary hashes, and paired clean/metamorphic controls reduce dependence on PR #495
literal strings.

## Blinding boundary

`inputs.json` contains detector-visible cases. `oracle.json` is passed only to scoring/test
comparison after each detector subprocess exits; the detector command receives only `inputs.json`
and one opaque `case_id`. `corpus-manifest.json` freezes both hashes before GREEN.

This is process-level fixture blinding, not a secret benchmark: all files become public after commit.
It measures deterministic preflight behavior only. It does not establish hosted-model recall,
token/cost improvement, or prospective production performance.

## Captain correction corpus (frozen 2026-07-24)

`correction-inputs.json` and separate `correction-oracle.json` were frozen before bidirectional
impact-graph or counterfactual-lab treatment implementation. The detector receives only the inputs
and opaque case id; scoring receives the oracle afterward. The corpus covers upstream-only leaf
references, both-direction script/package impact, authority neighborhoods, generator chains, Go
importer/tests, unknown variants, cycles/bounds, missing targets, unrelated controls, candidate/
outside writes, forbidden effect categories, secret capture, resource limits, cleanup/identity,
inconclusive hypotheses, clean experiment/static controls, and contract migration.

`correction-corpus-manifest.json` freezes content hashes/sizes. Real disposable-repository and OS
sandbox integration tests remain necessary because the detector corpus cannot prove parser,
subprocess, filesystem, sandbox, process, or cleanup enforcement.
