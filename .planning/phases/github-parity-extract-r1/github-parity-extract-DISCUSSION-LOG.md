# Phase github-parity-extract: r1 — Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution
> agents. Decisions are captured in `github-parity-extract-CONTEXT.md`.

**Date:** 2026-08-08  
**Phase:** github-parity-extract-r1  
**Mode:** `--auto`; the captain's explicit orders supplied the decisions.  
**Areas discussed:** live proof, rate limits, destructive-write semantics, generated help,
connector scope.

---

## Live proof

| Option | Description | Selected |
|---|---|---|
| Fixture/reachability sample | Reuse static fixtures or a subset of commands. | |
| Full live harness | Execute every implemented GitHub command and require asserted returned data or a concrete untestable reason. | ✓ |

**Choice:** Full live harness.  
**Notes:** Captain order makes `proven / untestable-with-reason / failed` the only
acceptable terminal accounting. Writes remain in a dedicated private test repository.

---

## Rate limits

| Option | Description | Selected |
|---|---|---|
| New response-driven limiter | Add a separate GitHub mechanism. | |
| Existing runtime limiter | Declare GitHub policy in the existing schema and prove the existing requester wrapper stops before GitHub. | ✓ |

**Choice:** Existing runtime limiter.  
**Notes:** The captain corrected the earlier direction: `Runtime.requesterFor` already
performs admission and observation for all request paths. The proof must show remaining
GitHub headroom after the local limiter stops the sweep.

---

## Destructive writes and help

| Option | Description | Selected |
|---|---|---|
| Retain a nonexistent flag | Continue to document `--allow-destructive`. | |
| Derive real point-of-use help | Remove false notes and expose the real confirmation and grant flow. | ✓ |

**Choice:** Derive real point-of-use help.  
**Notes:** Confirmation is caller-supplied intent acknowledgement, not human approval.
GitHub-only test coverage prevents the eight notes from returning; other connectors are counted,
not changed in this lane.

---

## Connector and PR scope

| Option | Description | Selected |
|---|---|---|
| Sweep-wide cleanup | Change unrelated connector definitions while touching the generator. | |
| GitHub-only PR | Limit code and generated artifact deltas to GitHub; report other debt. | ✓ |

**Choice:** GitHub-only PR.  
**Notes:** The paused sweep rebases after this PR lands. Inline GSD execution is the documented
fallback because this session has no compatible isolated GSD workers and delegation is not authorised.

---

## Remaining documented-operation gaps

| Option | Description | Selected |
|---|---|---|
| Treat blocked rows as a final exclusion | Preserve the 98 classified rows without an executable contract. | |
| Close executable foundations in this lane | Add bounded status/text response and declared request-body support, then regenerate GitHub declarations. | ✓ |

**Choice:** Close executable foundations in this lane.
**Notes:** The captain's later order governs: no final PR or completion claim while these
self-fixable GitHub gaps remain. The first slice is deliberately narrow: status-only reads,
text responses, and GitHub Markdown's declared raw-text POST body. OAuth application endpoints
will use a separately declared Basic-auth contract or remain fail-closed; the normal GitHub bearer
credential is not a substitute for the provider's documented authentication requirement.

## the agent's Discretion

- Select safe, existing harness seams and exact rate budgets based on provider-cited policy and
  the live header contract.

## Deferred Ideas

- Other connector phantom-flag remediation is deferred to the relevant parity lane.

---

## OneOf request-body contracts

| Option | Description | Selected |
|---|---|---|
| Flatten a union into one permissive command | Let one command accept incompatible documented request shapes. | |
| Separate named actions | Model each disjoint documented request arm as its own preflightable write contract. | ✓ |

**Choice:** Separate named actions. The existing `covered_by.writes` array is the intended
one-provider-endpoint-to-many-write-contract relationship, so no shared engine extension is needed.
The four bulk-attestation delete-request contracts remain destructive caller-supplied intent
acknowledgements; campaign, ProjectsV2, and Codespaces contracts remain approval-only creates.
