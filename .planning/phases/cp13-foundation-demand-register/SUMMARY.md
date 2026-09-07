---
phase: cp13-foundation-demand-register
status: local_verification_complete_review_pending
coverage:
  - id: D1
    description: "Exact source/lane register and prior Atlas lookup"
    requirement: CP13-01
    human_judgment: false
    verification:
      - kind: integration
        ref: "cmd/connectorgen: TestSourceFoundationCoverageExactMembership; exact receipts in VERIFICATION.md and TDD-LEDGER.md"
        status: pass
  - id: D2
    description: "Concrete owner, artifact, example and adopter observations"
    requirement: CP13-02
    human_judgment: false
    verification:
      - kind: integration
        ref: "cmd/connectorgen: TestSourceFoundationExamplesCurrentReferences; exact receipts in VERIFICATION.md and TDD-LEDGER.md"
        status: pass
  - id: D3
    description: "Distinct requirement classifications"
    requirement: CP13-03
    human_judgment: false
    verification:
      - kind: integration
        ref: "cmd/connectorgen: TestSourceFoundationRequirementSharedAndLocalAspects; exact receipts in VERIFICATION.md and TDD-LEDGER.md"
        status: pass
  - id: D4
    description: "Narrow foundation proof admission and lane isolation"
    requirement: CP13-04
    human_judgment: false
    verification:
      - kind: integration
        ref: "cmd/connectorgen: TestSourceFoundationNoLanePromotion; exact receipts in VERIFICATION.md and TDD-LEDGER.md"
        status: pass
  - id: D5
    description: "MIME/auth/body/paging source fit"
    requirement: CP13-05
    human_judgment: false
    verification:
      - kind: integration
        ref: "cmd/connectorgen: TestSourceFoundationFacetSourceSemantics; exact receipts in VERIFICATION.md and TDD-LEDGER.md"
        status: pass
  - id: D6
    description: "Historical receiver and Sentry reconciliation"
    requirement: CP13-06
    human_judgment: false
    verification:
      - kind: integration
        ref: "cmd/connectorgen: TestSourceFoundationObligationsCurrentCorpus; exact receipts in VERIFICATION.md and TDD-LEDGER.md"
        status: pass
  - id: D7
    description: "Existing receiver decision owner preserved"
    requirement: CP13-07
    human_judgment: false
    verification:
      - kind: integration
        ref: "cmd/connectorgen: TestSourceFoundationRequirementDecisionAdmission; exact receipts in VERIFICATION.md and TDD-LEDGER.md"
        status: pass
  - id: D8
    description: "Separate conditional exposure owner preserved"
    requirement: CP13-08
    human_judgment: false
    verification:
      - kind: integration
        ref: "cmd/connectorgen: TestSourceFoundationRequirementDecisionAdmission; exact receipts in VERIFICATION.md and TDD-LEDGER.md"
        status: pass
  - id: D9
    description: "Unresolved work retained beside independent source work"
    requirement: CP13-09
    human_judgment: false
    verification:
      - kind: integration
        ref: "cmd/connectorgen: TestSourceFoundationRegisterCurrentKnownRelation; exact receipts in VERIFICATION.md and TDD-LEDGER.md"
        status: pass
---
# CP13 corrected local implementation summary

The final150/154 correction wave is implemented and locally verified. The authoring demand consumer now binds body/auth reuse to exact mechanisms and source/canonical selectors, derives the narrowly permitted whole-CLI-missing admission from real source/engine observations, requires complete local command accounting, and revalidates input identity/absence/namespace before output. Optional absent/stale proofs preserve unresolved demand. Planned proof file roles are order independent; decision decoder/schema preserve required empty pending conditions while refusing malformed/null/omitted values.

The original148 report accepted CP11/CP12 and left five CP13 findings. Their corrections and all original CP13-01..09 obligations remain subject to a fresh independent judgment. No acceptance, runtime capability or provider certification is inferred from local tests.

Normal and race each cover the same478 unique identities across87parents, including all original263 and later147. Failed normal/interrupted race commands retain their original status; bounded unchanged-parent reuse and complete reruns of affected parents are documented in VERIFICATION.md. Separate CP12 shared consumers pass101normal/101race. Corpus schema79cases, source/demand A/B, saved check, vet/lint/build/docs/agent-contract checks pass. Exact input pins, failures and test identities are in the final candidate packet.

The current register retains U30401/A26/complement30375/K26,25unresolved requirements and one existing Vercel gap,35Atlas examples and0adopters. Four narrow proof observations do not promote any source lane:0implemented. Current generated hashes and the unchanged independent baseline are recorded in VERIFICATION.md. Missing example joins, receiver decisions and separate exposure conditions remain explicit.

The original nine-item and25-family contract-to-evidence tables remain authoritative in PLAN.md and VERIFICATION.md. The final appendix supersedes historical144/147 candidate hashes and test counts, without rewriting their evidence. The sole-owner inline GSD fallback and required skills remain recorded in PLAN.md. Authoring help/schema/docs parity is covered; no PM runtime/App/manual/website consumer changed.

Next: freeze the coherent committed correction and submit it to Firstmate for the final bound independent Astra/xhigh review. Firstmate156 governs carrying0–4 final corrections across the next checkpoint; it does not waive any correction or permit self-acceptance.
