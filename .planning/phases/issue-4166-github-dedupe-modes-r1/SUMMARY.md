---
coverage:
  - id: D1
    description: Certification mode cells use real transport-role admission.
    verification:
      - kind: unit
        ref: cmd/connectorgen TestCertificationSyncModeReadRequiresDeclaredSourceTransportMode
        status: pass
      - kind: other
        ref: certification-matrix --all repeat SHA-256 comparison
        status: pass
    human_judgment: false
  - id: D2
    description: GitHub current dedupe and history execute through production transport dispatch.
    verification:
      - kind: integration
        ref: internal/app TestGithubContractDedupeModesMaterializeCurrentAndHistoryRows
        status: pass
      - kind: other
        ref: fresh pm against retained private GitHub PR
        status: pass
    human_judgment: false
---

# Summary: GitHub dedupe modes r1

The certification generator now intersects coarse capabilities with the
definition-owned source or destination transport modes. GitHub admits the two
implemented source modes into the local warehouse, whose destination declares
the matching dedupe and history apply strategies.

The bounded declarative source re-emits provider pages for those two replay-safe
modes, so a changed record reaches the identity-aware warehouse materializer
instead of producing a false invalid-checkpoint recovery. Current dedupe keeps
one primary-key row; history emits stable source versions with closed/open SCD2
intervals.

The direct-PR live proof used retained private repository
`karthik-sivadas/pm-truth-github-dedupe-modes-build-r1`, pull request #1. No
credential material is recorded here.
