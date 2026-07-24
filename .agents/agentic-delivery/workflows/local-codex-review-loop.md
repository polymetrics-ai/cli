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
- completed verification commands and their results;
- the active review-system contract at
  `.agents/agentic-delivery/contracts/pm-review-system.json`.

A branch name, mutable PR ref, prior review, or session memory is not an exact identity.

## Deterministic review compilation

1. The parent orchestrator confirms the fetched remote head and candidate worktree equal the supplied exact head SHA and records that precondition without delegating network access. Confirm the comparison base is the candidate merge base. Stop on drift.
2. Run `scripts/pm-review-system.py compile --scope <validated-per-run-scope>` for that exact base/head/tree. It must return `ready`
   before model review. Treat its changed-path assignment, active reference closure, authority
   inventory, and versioned practical impact graph as review inputs. The graph indexes its declared
   universe before traversal, seeds canonical roots plus every changed file, and follows a typed
   bidirectional upstream/downstream/lateral/temporal relation policy with edge provenance and
   `active|inactive|unknown` certainty. Missing/unresolved edges, unsafe paths, incomplete impact,
   authoritative-state disagreement, or any graph/index/traversal/packet bound block review. This is
   file/package impact, not a symbol-level call/data-flow claim.
3. The compiler emits paths and metadata only. It must not copy file contents, environment values,
   or credentials into packet artifacts.
4. Complete impact discovery before packetization. For a small coherent range, use one combined
   diff packet only when all configured file/line/domain limits pass; otherwise split architecture,
   authority, and implementation packets. Always assign complete impact files/edge ids to bounded
   impact packets. If discovery or a packet cannot fit without truncation, stop as blocked.

## Fresh-context packet review and synthesis

1. Spawn a fresh-context local Codex reviewer for each compiled packet using the read-only
   `pm-reviewer` role (Sol/xhigh) or the runtime's equivalent. Packet reviewers are analytical
   inputs; the parent orchestrator remains the only lifecycle and disposition owner.
2. Keep the canonical candidate and review source read-only. `bash` is allowed only for non-mutating
   local identity, diff, log, and assigned test inspection; packet reviewers have no network access. Temporary hypothesis
   changes are allowed only through `scripts/pm-review-lab.py` in a private disposable exact-head
   copy. No candidate edit/write, generic shell, network, commit, push, PR mutation, install,
   credentialed/live call, deployment, destructive external effect, or merge is allowed.
3. Build an impact model before judging lines. Trace all four directions, inspect history and
   divergent siblings when relevant, state falsifiable claim/alternative hypotheses, seek
   disconfirming evidence, and use the smallest discriminating lab experiment only when static
   evidence is insufficient. An unavailable sandbox, denial, timeout/bound, cleanup failure,
   candidate drift, or inconclusive experiment blocks clean review.
4. Each v3 response follows `pm-review-packet-template.md`: exact base/head/tree; changed, closure,
   authority, impact-file, impact-edge, edge-context-file, invariant, and behavior coverage; experiment/no-experiment
   evidence; unreviewed files; context overflow/truncation; and findings. Finding count is unlimited.
   Missing token/cost/latency data stays explicitly null.
5. Preserve raw responses and lab evidence outside the tracked worktree. Run `scripts/pm-review-system.py synthesize`
   to produce one PM-owned result. Missing responses/coverage, stale identities, any unreviewed file,
   or overflow/truncation cannot synthesize clean.
6. Review correctness, security, safety, regressions, test adequacy, evidence truthfulness,
   write-scope violations, machine contracts, and human gates. Return findings with severity,
   file/line evidence, impact, and smallest safe correction. List residual risk separately.
7. The synthesized result is `clean`, `findings_correction_required`, or `blocked`. Only complete
   clean packet responses with zero findings produce `clean`.

## Disposition and correction

The parent orchestrator owns disposition. Use this exact machine vocabulary:

`finding_disposition_values: [accepted, accepted_with_modification, declined, duplicate, deferred, needs_human]`

For every actionable finding, record one value with a reason and follow-up reference where applicable.

Accepted corrections return to the isolated implementation worker, then repeat affected tests and
exact-head verification. Every changed head invalidates the prior manifest, packet responses,
synthesis, and Shepherd result; compile fresh packets and run fresh-context review against the new
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
- compiler manifest identity, active closure/authority findings, typed practical impact graph
  counts/bounds/provenance, packet selection, and exact changed/impact coverage;
- packet ids, reviewer runtime/model/fresh-context identities, raw-response and hypothesis-lab
  evidence paths/hashes, observable behavior/experiment outcomes, safety/cleanup proofs, and any
  unavailable token/cost/latency fields;
- local Codex synthesized status: `pending`, `findings_correction_required`, `clean`, `comments_addressed`, or `blocked`;
- findings and disposition artifact;
- Shepherd status, verdict, score, and evidence artifact;
- CI status and residual human gates;
- measured fixture/replay scope separately from prospective review outcomes, without claiming
  unmeasured improvement.

## Prohibited PM coverage routes

Do not request or count Claude or GitHub Copilot as required, fallback, or substitute review
coverage for this PM route. Historical records and legacy bot-review documents remain truthful but
are not inputs to current PM orchestration.

The parent PR into `main`, dependency approval, auth scope changes, secrets, destructive actions,
production deploys, and quality-gate reductions remain human-only decisions.
