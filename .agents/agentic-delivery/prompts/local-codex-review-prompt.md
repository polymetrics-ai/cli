# Fresh-context local Codex exact-head review

Independent read-only review. This unchanged `pm-review-system.py render` prompt binds exact base,
exact head/tree, packet, and slices. Follow `local-codex-review-loop.md` and
`pm-review-packet-template.md`.

- Verify identity; drift is `blocked`. Review every assignment/invariant. Build impact first across
  upstream, downstream, lateral, and temporal paths; inspect relevant history and siblings.
- Findings are unlimited. Actionable means introduced, activated, worsened, or relied on by this
  candidate; put unrelated pre-existing issues in residual risk.
- Record hypotheses (`id`, `claim`, `strongest_alternative`, `falsifier`, `evidence_paths`) and
  disconfirming evidence. Experiment only if static evidence is insufficient.
- Never change the primary candidate. Use only `scripts/pm-review-lab.py`. Declared bounded dummy
  writes, edits, caches, databases, and local services are valid. Primary-project modification,
  commit, push, and deployment are forbidden. Network, install, host credentials, and live
  connectors require exact captain approval. Denial, unsafe effect, limit, residue, cleanup/identity
  failure, or inconclusive evidence blocks.
- Review correctness, security, regressions, tests, evidence, contracts, and secrets. Return one
  `polymetrics.ai/pm-review-packet-response/v4` JSON object with exact assignments; invariant,
  hypothesis, behavior, and experiment evidence; unreviewed files; findings; residual risk; context;
  wall time; and only available telemetry. Never invent telemetry.
- Each finding has severity, category, path/line evidence, impact, and smallest safe correction.
  `clean` requires no finding, overflow, truncation, unreviewed assignment, or blocked evidence.

One-token-per-byte bounds input; response reserve is separate. Head changes invalidate evidence.
Statuses: `pending`, `findings_correction_required`, `clean`, `comments_addressed`, `blocked`.
Disposition:
`finding_disposition_values: [accepted, accepted_with_modification, declined, duplicate, deferred, needs_human]`.
No integration approval; same-identity Shepherd and final human authority remain.
