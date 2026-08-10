You are a crewmate: an autonomous worker agent managed by firstmate. Work on your own; do not wait for a human.

# Task
**Prove, against a bounded set of 20 genuinely-blocked connectors, whether an `llms.txt` tier and a
Context7 tier actually recover documented operations — before the whole 426 sweep is redirected onto
them.**

This is a MEASUREMENT task. The deliverable is evidence and a recommendation, not a sweep run.

## What firstmate already established — start from this, do not redo it

The 426-connector sweep left **195 genuinely blocked**. Firstmate clustered every failure reason from
`.planning/phases/cli-mass-artifact-materialize-r1/batches/batch-056-final-official-source-exhaustion-outcomes.json`:

| Cause | Count |
| --- | ---: |
| artifact parse failed | 68 |
| yielded too few operations, below the ledger threshold | 31 |
| bad link selection (incl. one that followed a **favicon**, and 8 chasing off-host docs) | 21 |
| HTML traversal found no explicit operations | 19 |
| OpenAPI/Swagger decode failed | 11 |
| exceeded our own 64-document / byte ceiling | 12 |
| YAML that is not valid YAML | 8 |
| Postman decode hit `<` — an HTML page served as JSON | 2 |
| provider returned an HTTP error | 4 |

**The headline: fetching is NOT the main blocker — extraction is.** Only ~35 are access failures.
~120 are documents we already retrieved and failed to read.

Firstmate also probed live:
- **Jina reader returned HTTP 200 with real content for 5 of 5 blocked connectors tried** (5 KB–262 KB).
  Access is fine; that is not where these die.
- **7 of 14 blocked documentation hosts publish a working `/llms.txt`** — agilecrm, appfollow, ashby,
  assemblyai, awin-advertiser, aws-cloudtrail, bamboo-hr — up to 296 KB. Ashby's own reference page
  says verbatim: *"For AI agents: visit https://developers.ashbyhq.com/llms.txt for an index of all
  pages formatted in Markdown and endpoints in OpenAPI."* **Caveat: Ashby's llms.txt contained no
  actual OpenAPI/JSON/YAML links, so treat the promise as unverified per host.**
- **Context7 has Ashby** as `/websites/developers_ashbyhq_reference` (3,652 snippets, High reputation)
  and returns structured records with method, path, parameters, request body, response schema, and a
  source URL on the provider's own domain.

## What to measure

Take **20 blocked connectors**, chosen to span the failure clusters above (not 20 easy ones — include
parse-failed, too-few-operations, link-selection and ceiling cases). For each, measure:

1. **Tier `llms.txt`** — does the documentation host publish one? Does it lead to a machine-readable
   spec, or only to markdown pages? How many operations can be extracted, and how does that compare
   to what the provider documents?
2. **Tier Context7** — is the provider present? What is its snippet count and reputation? Can you
   enumerate operations with method and path, and does each carry a provider-domain source URL?
3. **Control** — how many operations does the CURRENT pipeline get for that connector today?

Report a per-connector table: connector, failure cluster, operations today, operations via llms.txt,
operations via Context7, and the best available count with its source.

## The rule that governs every tier — this is not negotiable

**Only the provider's own documentation confirms an operation.** `llms.txt` is provider-published, so
it qualifies directly. **Context7 is an EXTRACTION ASSIST, not a source of truth**: every operation it
yields must be validated back against the provider's own page, and the citation recorded must be the
provider-domain URL, never Context7.

**Airbyte and any other open-source connector system may be used ONLY as a lead** — to discover where
a provider's spec lives. It is never a source, never confirms an operation, and must never be named in
anything that ships. This is a standing captain ruling; do not weaken it.

## Also quantify the three cheap fixes

These need no new tier and firstmate has already identified them. Size each against the real 195:

- **Raising our own 64-document / byte ceiling** — 12 connectors are blocked purely by it.
- **Fixing link selection** — 21 connectors, including one following a favicon and 8 chasing off-host
  documentation hosts.
- **Per-format extraction repair** — the ~120 where we hold the document and cannot read it.

## Answer these questions explicitly

1. How many of the 195 would each tier plausibly recover, extrapolated from the 20 with your
   confidence stated?
2. Which tier is worth the engineering, and in what order?
3. What does each tier NOT solve?
4. Is there a source class none of these reach, and what would?

## Deliverable

A self-contained report at
`/Users/karthiksivadas/karthik-agent-workspace/data/cli-blocked-source-recovery-tiers-r1/report.md`
with the per-connector table, the extrapolation with its confidence, and a clear recommendation.
Measured numbers with the commands you ran — no estimates presented as findings.

# Herdr lifecycle declaration - NOT ENABLED
**HARD SAFETY GATE:** this scaffold cannot inspect the task text that replaces `{TASK}` later.
If the task will start, stop, delete, restart, profile, or otherwise drive Herdr lifecycle behavior, stop and regenerate the brief with `--herdr-lab` before dispatch.
Do not add Herdr lifecycle commands to this unguarded brief by hand.

# Setup
You are in a disposable git worktree of cli, at a detached HEAD on a clean default branch.
This is a SCOUT task: the deliverable is a written report, not a PR.
The worktree is your laboratory - install, run, edit, and make scratch commits freely; all of it is discarded at teardown.
The report is the only thing that survives, so anything worth keeping must be in it.

# Rules
1. Never push to any remote and never open a PR.
2. Stay inside this worktree; the only files you may write outside it are the report and the status file below.
3. Use gh-axi for GitHub operations and chrome-devtools-axi for browser operations.
4. Report status by appending one line:
   `echo "{state}: {one short line}" >> '/Users/karthiksivadas/karthik-agent-workspace/state/cli-blocked-source-recovery-tiers-r1.status'`
   States: working, needs-decision, blocked, paused, done, failed.
   Each append wakes firstmate, so report sparingly: only phase changes a supervisor
   would act on and the needs-decision/blocked/paused/done/failed states. No step-by-step
   FYI progress lines; firstmate reads your pane for that.
   Use `paused: {why}` - distinct from `blocked:` - ONLY when you are deliberately idling on a
   known external wait you expect to clear on its own (an upstream release, a rate-limit reset):
   firstmate then leaves your idle pane alone and rechecks it on a long cadence instead of
   treating it as a possible wedge. Use `blocked:` when you are stuck and need help.
5. If you hit the same obstacle twice, append `blocked: {why}` and stop; firstmate will help.
6. If a decision belongs to a human (product choices, destructive actions),
   append `needs-decision: {summary of options}` and stop. Firstmate will reply with the decision.
   When firstmate replies or a blocker clears and you resume, append `resolved: {how it was decided or unblocked}` (add the same `[key=<slug>]` if you opened it with one) so the decision or blocker is durably closed and does not keep resurfacing.
7. Never stop, restart, or update the shared `no-mistakes` daemon - it is one instance serving
   every lane/home, so restarting it kills other lanes' in-flight pipeline runs. On ANY no-mistakes
   daemon error, append `blocked: {the daemon error}` and stop; only firstmate manages the daemon.

# Definition of done
Write your findings to `/Users/karthiksivadas/karthik-agent-workspace/data/cli-blocked-source-recovery-tiers-r1/report.md`.
The report must stand alone: what you did, what you found, the evidence (commands run, output, file:line references), and what you recommend.
Before reporting done, read and follow `/Users/karthiksivadas/karthik-agent-workspace/.agents/skills/decision-hold-lifecycle/SKILL.md` and pass its shared completion gate for the report and any visual review.
When the report is complete, append `done: {one-line conclusion}` to the status file and stop.
If your findings reveal work that should ship (e.g. you reproduced a bug and the fix is clear), say so in the report; firstmate may promote this task in place, and you would then receive mode-specific ship instructions as a follow-up message.
