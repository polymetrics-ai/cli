---
coverage:
  - id: D1
    description: The 70 reviewed Zoom direct-read operations and commands are declared together.
    verification:
      - kind: unit
        ref: internal/connectors/defs/zoom/command_surface_test.go:TestReviewedDirectReadSalvageCohort
        status: pass
      - kind: other
        ref: go run ./cmd/connectorgen validate internal/connectors/defs/zoom
        status: pass
    human_judgment: false
  - id: D2
    description: Every available sanitized direct-read fixture executes against the real operation runner.
    verification:
      - kind: unit
        ref: internal/connectors/defs/zoom/command_surface_test.go:TestReviewedDirectReadFixturesExecute
        status: pass
    human_judgment: false
  - id: D3
    description: Every direct-read command is binary-reachable and stops before network without a credential.
    verification:
      - kind: other
        ref: seven 10-command pm binary sweeps
        status: pass
    human_judgment: false
  - id: D4
    description: The wave remains explicitly pending central certification scope.
    verification:
      - kind: other
        ref: go run ./cmd/agentcontractgen certification-gate --connector zoom --transition integrate_sub_pr
        status: pass
    human_judgment: false
---

# Zoom direct-read salvage — Wave #4267 summary

Wave #4267 selectively salvages 70 reviewed Zoom `rest_read` definitions, command paths, endpoint dispositions, 52 sanitized direct-read fixtures, and their generated projections from PR #3951. It does not rebase or adopt that PR; it starts from post-foundation `main` and changes only this fresh wave's defined artifacts.

The change also regenerates the nine root/help CLI transcript records and Zoom's generated MANUAL/SKILL so runtime help and connector docs describe the enlarged command surface.

The 70 direct-read rows are implemented with fixture proof and each is `fixture_required` in the generated sweep. No live credential was resolved. Certification remains pending because Zoom is intentionally outside the firstmate-owned central scope; the expected scope HALT is not represented as a certification success.

The four SCIM2 reads have a documented bare-origin requirement. Their commands are implemented and fixture-proven with an explicit per-command `base_url` override documented in the connector bundle. Default operation-specific origin selection needs a separately approved foundation feature before live certification, and was not altered here.

The canonical issue-first lifecycle was executed inline: `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` prompts were resolved with `scripts/gsd prompt`. The delivery contract forbids role delegation, so the manual fallback, TDD evidence, verification, and review are all recorded in this phase directory.
