# Local Codex Review Loop

This is the canonical code-review implementation for current and forward PM parent orchestration.
Run it after exact-head verification and before independent Shepherd trajectory validation. It does
not request GitHub-hosted AI review and does not replace final human authority.

## Inputs

The parent orchestrator supplies:

- repository and isolated working directory;
- pull request or review target;
- exact base branch and exact base SHA;
- exact head branch and exact head SHA;
- issue scope, allowed paths, acceptance criteria, and human gates;
- `max_correction_rounds` (default 4) and `rounds_by_range` for this exact review lineage;
- completed verification commands and their results.

A branch name, mutable PR ref, prior review, or session memory is not an exact identity.

## Fresh-context reviewer

1. Confirm the candidate worktree and remote head equal the supplied exact head SHA. Confirm the
   comparison base equals the supplied exact base SHA. Stop on drift.
2. Spawn a fresh-context local Codex reviewer using the read-only `pm-reviewer` role (Sol/xhigh) or
   the runtime's equivalent read-only Codex context. The reviewer must not inherit implementation
   reasoning as authority.
3. Give the reviewer read-only tools. `bash` is allowed only for non-mutating commands such as
   `git status`, `git rev-parse`, `git diff`, `git log`, tests explicitly assigned to review, and
   read-only `gh-axi` inspection. No edit/write, commit, push, PR mutation, or merge is allowed.
4. Review the exact `base...head` range for correctness, security, safety, regressions, test
   adequacy, evidence truthfulness, write-scope violations, machine contracts, and human gates.
5. Return `CLEAN_NO_ACTIONABLE_FINDINGS` or findings with severity, file/line evidence, impact,
   and the smallest safe correction. List residual risk separately from actionable findings.

## Disposition and correction

The parent orchestrator owns disposition. For every actionable finding, record one of `accepted`,
`accepted_with_modification`, `declined`, `duplicate`, `deferred`, or `needs_human`, with a reason
and follow-up reference where applicable.

Accepted corrections return to the isolated implementation worker, then repeat affected tests and
exact-head verification. Every changed head requires a fresh-context re-review against the new
exact head. Increment `rounds_by_range` for the exact base/candidate lineage. When it exceeds
`max_correction_rounds` (default 4), mark the range blocked with outstanding findings and stop for a
human; never continue indefinitely or reset the count through a replacement PR.

Review is clean only when no actionable finding remains and every prior finding has a disposition.
Local Codex review is review evidence, not merge approval.

## Independent Shepherd gate

After local Codex review is clean at the exact head, run
`.agents/agentic-delivery/workflows/shepherd-validator.md` in an independent context against the
same exact identities and durable trajectory. Shepherd validates orchestration order and evidence;
it does not replace code review. Integration requires both a clean local Codex review and a
Shepherd `PROCEED` verdict for the relevant review transition.

Any head change after either gate invalidates both exact-head results and requires verification,
fresh-context local Codex re-review, and Shepherd validation again.

## Review coverage record

Record for every candidate range:

- exact base branch and SHA;
- exact head branch and SHA;
- reviewer runtime/model and fresh-context identity;
- local Codex status: `pending`, `clean`, `comments_addressed`, or `blocked`;
- findings and disposition artifact;
- Shepherd status, verdict, score, and evidence artifact;
- CI status and residual human gates.

## Prohibited PM coverage routes

Do not request or count Claude or GitHub Copilot as required, fallback, or substitute review
coverage for this PM route. Historical records and legacy bot-review documents remain truthful but
are not inputs to current PM orchestration.

The parent PR into `main`, dependency approval, auth scope changes, secrets, destructive actions,
production deploys, and quality-gate reductions remain human-only decisions.
