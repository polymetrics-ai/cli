# Correction specification

## Outcome

Close four release blockers without reopening the completed 17-row hardening matrix:

1. `start --issue N` bootstraps a canonical schema-2 JSON plan when the plan is genuinely absent.
2. A `verification` AgentSession selects immutable plan command IDs and the host executes the exact commands.
3. Verification is explicitly trusted-local: reviewed test code runs with the current OS user's authority.
4. A GitHub draft-to-ready transition accepts an exact non-draft observation whose second-resolution revision equals the prepared revision.

## Plan bootstrap contract

The host reads the authoritative open parent issue, bounded sub-issue facts, repository/default-branch
coordinates, authenticated maintainer identity, current non-default parent branch, and exact parent head.
A `planning` AgentSession (`openai-codex/gpt-5.6-sol/xhigh`) receives read-only repository access and
one typed `host_inspect` submission capability. It proposes semantic children without numeric issue IDs
or host-owned authority fields. The host validates the proposal, constructs canonical orchestration
markers, creates or reconciles child issues, inserts the returned issue numbers, validates the final
schema-2 plan, and atomically publishes `.planning/shepherd/issue-N.json` without overwriting an existing
manual plan. Existing ordinary sub-issues are planning evidence; this correction does not silently adopt
an issue that lacks Shepherd's canonical marker.

## Verification contract

A `verification` AgentSession receives only repository read access and `host_verify({id})`. The ID must
be the next immutable command in the validated plan. Implementation and correction AgentSessions receive
the same ID-only capability and may rerun declared commands while editing so they can perform RED→GREEN;
they cannot alter a command tuple. The host, not agent prose, resolves the executable, argv, cwd, timeout,
output bound, cancellation, diagnostic redaction, and authoritative result. The independent verification
stage observes every required command in order; the first failure stops the sequence and routes the child
to correction. No generic shell, arbitrary argv, environment, GitHub write, HTTP write, or SQL write
authority is exposed.

## Trusted-local authority decision

The user explicitly accepts execution of repository test code with the local user's process authority.
This is not a sandbox claim. Secrets remain excluded; default-branch mutation remains forbidden; the
parent merge stays human-owned; and destructive/auth/dependency/quality-gate changes remain gated.

## Ready-transition invariant

Exact repository, PR number, open state, branch, and head SHA are mandatory. `draft:false` with
`appliedRevision >= expectedRevision` proves readiness. Equality is never accepted while the PR is draft.
