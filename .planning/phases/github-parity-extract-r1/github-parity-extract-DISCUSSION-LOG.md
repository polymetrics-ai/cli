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

The body-field audit found an existing but unimplemented `json` CLI flag type. It is used only as
a parsed, schema-validated value for declared `record.*` object/array fields; accepting it as a
literal body, path, query, or header would create a generic API escape hatch and is not selected.

---

## Captain-authorized live lab (2026-08-09)

| Option | Description | Selected |
|---|---|---|
| Reuse historical single-repo skips | Keep the 957 rows unattempted and use a non-PM bootstrap path. | |
| Parameterized PM-only lab | Derive every historical skip into a fail-closed, cohort-scoped PM-only fixture plan. | ✓ |

**Choice:** Parameterized PM-only lab.
**Notes:** The captain explicitly authorized disposable lab resources, but not a bypass around
the connector. The manifest must partition every historical case into retained personal lab,
Free sandbox organization, App/Marketplace, or entitlement cohort. The default-deny boundary must
bind both slug and immutable ID, deny `polymetrics-ai`, and reject ambiguity before `pm` starts.
The retained personal repository stays mutation-ineligible until its proof-program provenance is
recorded in the append-only cleanup ledger. No new no-mistakes run begins before the framework and
the next coherent live-proof increment are committed.

---

## External bootstrap probes (2026-08-09)

| Option | Description | Selected |
|---|---|---|
| Substitute a provider/UI path | Create an organization or App through an untracked path, then point PM at it. | |
| Fixed PM-only probes | Derive the available PM bootstrap surface, probe only safe account reads, and record exact missing surface/credential prerequisites. | ✓ |

**Choice:** Fixed PM-only probes.
**Notes:** The current PM surface has `orgs delete` but no organization-create command; deletion
is not attempted without a run-owned immutable organization. `apps create-from-manifest` requires
a conversion code but no PM command issues one, so no App fixture is attempted. The fixed
`apps get-authenticated` PM read returned a sanitized 401 under the user credential, while the
Marketplace-user subscription read returned a sanitized 200. The successful read proves only that
route; it does not establish an App, listing, plan, or installation fixture.

---

## First reusable personal resource family (2026-08-09)

| Option | Description | Selected |
|---|---|---|
| Exit-status-only label mutation | Treat a successful write command as sufficient evidence. | |
| Read-back asserted label lifecycle | Bind list reads to the lab repository, assert generated ID/color, and delete through the PM confirmation path. | ✓ |

**Choice:** Read-back asserted label lifecycle.
**Notes:** The generated label was absent at baseline, created through PM, read back by immutable
ID, edited with its returned canonical color asserted, then deleted through PM's typed confirmation
and confirmed absent. This is a reusable reversible-resource pattern; no provider output or
credential profile was retained.

---

## Editable issue and comment proof (2026-08-09)

| Option | Description | Selected |
|---|---|---|
| Treat write exit status as proof | Skip the returned-data assertions after issue edit/comment. | |
| Bounded PM-only read-back | Re-supply safe write flags, bind the generated issue by immutable ID, and retry only stale successful PM list observations. | ✓ |

**Choice:** Bounded PM-only read-back.
**Notes:** The first immediate PM list after an accepted generated issue create exposed a
read-after-write visibility delay. It was independently PM-read, PM-closed, and retained before
any edit/comment attempt. The final fresh issue was created, edited, and commented only through
PM plan → preview → approval → execute. Its bounded PM list assertions retained one immutable ID,
matched the caller-declared edit properties in memory, and observed exactly one comment before PM
close and retained-state verification. Record re-supply is required at preview and execution so
the runtime's withheld-field contract remains valid; PM read/provider errors are not retried or
reclassified as visibility delay. No provider payload, comment text, or credential detail is
persisted.

---

## Target rebind and authenticated-list failure (2026-08-10)

| Option | Description | Selected |
|---|---|---|
| Treat PM error as setup friction | Use the externally confirmed repository and omit the failing PM read. | |
| Preserve the PM error as a connector finding | Bind the independently confirmed immutable target, then capture and diagnose the same PM-only failure. | ✓ |

**Choice:** Preserve the failure as a connector finding.
**Notes:** Firstmate independently confirmed the captain-approved target is private, unarchived,
and has immutable repository ID `1327549621`. The current boundary must allow only that exact
owner/repository identity. The historical target remains validation-only evidence and cannot be
re-authorized for fixture traffic. The two earlier `repos list-for-authenticated-user` error
envelopes are a defect candidate: preserve the exact PM call and complete safe envelope, record a
provider status only if the envelope contains one, and determine whether the break is shared
authenticated direct-read plumbing or operation-specific before resuming the cohort.
