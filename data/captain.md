# Captain preferences

- 2026-08-10: PGLOGREPL APPROVED - POSTGRES CDC IS UNBLOCKED. The captain explicitly cleared the human-gated `github.com/jackc/pglogrepl` dependency decision that had held PostgreSQL native CDC since 2026-08-06. This is the gate, not a new scope: the work already exists on `origin/fm/cli-postgres-cdc-logical-replication-r1` at `aeafb4ff0` (13 behind / 2 ahead of main at ruling time) with `cdc.go`, `cdc_decode.go`, `cdc_lifecycle.go` and integration tests already written, and `pglogrepl v0.0.0-20260401131349` plus `pgx/v5` already in that branch's go.mod though NOT yet on main. The earlier standing condition still applies and is the acceptance bar: database dependencies are approved CONDITIONAL ON SECURE AND TESTED - maintenance status, CVEs, transitive footprint, licence, and a MEASURED binary-size delta stated in the PR body, to PR #3879's evidence standard. Coordinate with `cli-database-connector-framework-design-r1` so CDC does not invent an abstraction the framework design then contradicts.

- 2026-08-10: PR 3957 (426 SWEEP) MERGE GATE. PR 3957 must NOT merge on the generator capability alone. It merges only when the full 426 are complete: every connector's JSON materialized, its docs generated, and random binary validation run at intervals THROUGHOUT the sweep rather than once at the end, so defects are caught and fixed while the sweep is still running instead of surfacing after everything is authored. Current state at the time of this ruling: 231 materialized and 195 genuinely blocked for lack of a parseable source, so this gate keeps 3957 open until the extraction/recovery work lands too. CONSEQUENCE FIRSTMATE MUST MANAGE: 3957 becomes a long-lived branch over generated artifacts (endpoint ledger, website catalog, golden transcripts), which this repo has repeatedly seen go stale after every rebase and cascade into other branches - so it needs frequent rebases onto main and an inspected diff each time, or the generator capability gets split into its own PR that merges early while the sweep continues separately.

- 2026-08-10: PER-CONNECTOR DOCS ARE PART OF THE SWEEP, AND LUNA RUNS AS A SUB-AGENT UNDER THE TERRA LANE. The 426 sweep was found to update 231 command surfaces while updating ZERO of the newly materialized connectors' `docs.md` and zero `docs/connectors/` pages; `connectorgen` has no documentation subcommand at all, so re-running it can never close that gap. Documentation generation therefore becomes part of each connector's definition of done inside the sweep, not a later repo-wide pass, which is what makes it intrinsic to the generator-backed sweep rather than "documentation work" excluded by the 2026-08-08 Luna scoping rule. Mechanism, chosen by the captain as "whichever is best without corrupting git": the lane stays `gpt-5.6-terra` at `max` and is the SOLE writer of the worktree and the only thing that commits; it fans out concurrent `codex exec --model gpt-5.6-luna -c model_reasoning_effort="max"` SUB-AGENTS that read one connector's retained provider artifact plus its generated `api_surface.json`/`cli_surface.json` and return docs content only. Never run two lanes or two agents against one worktree - shared git index and file tree means interleaved commits, index-lock failures, and silently lost work; many readers and one writer is the only safe shape. Precision is enforced mechanically, not by trusting the model: every command named in `docs.md` must exist in `cli_surface.json`, every `implemented` command must appear in the docs, and docs may not claim availability beyond what the surface declares - a blocking gate, because this programme's dominant defect is declaring capability that does not exist. Existing sub-agent rules still bind: a sub-agent never answers a pipeline gate, never resolves an ask-user finding, never decides a diff is safe, and the lane verifies output against reality rather than accepting a report.

- 2026-08-09: CURRENT THREE-LANE AND MODEL ROUTING (SUPERSEDES every earlier Luna rule for this active sweep). Continue exactly three active CLI lanes: GitHub parity, MySQL, and the generator-backed 426-connector JSON materialization sweep. Docker Hub and Zoom implementation/validation lanes remain paused; their connector JSON may still be materialized inside the authorized generator sweep. Do not start another lane without the captain's explicit permission. Run the current generator-backed connector sweep on native Codex `gpt-5.6-terra` with reasoning effort `max`, not Luna; GitHub, MySQL, no-mistakes steps, investigation, review, and all other work also use Terra max. Preserve Firstmate as the captain's sole coordinator, Herdr as the recorded task-session provider, and Treehouse as isolated project-copy provider. This supersedes the 2026-08-08 entries that reserved Luna for generator-backed connector implementation sweeps.

- 2026-08-08: CONTINUE-ONLY FIRSTMATE/HERDR AND MODEL-ROUTING DECISION. The primary Firstmate is the captain's sole software-work coordinator and may continue only the already-authorized GitHub parity, Zoom parity, Docker Hub parity, connectorgen/provider-artifact generator work (including its bounded three-to-four-connector pilot), and MySQL implementation lanes. Do not start any unrelated lane without the captain's explicit permission. Herdr remains the session provider for those workers while Treehouse remains the isolated project-copy provider: one durable recorded Herdr endpoint per task, project changes only inside each task's isolated copy, and Firstmate supervises/steers rather than editing project code. Use native Codex `gpt-5.6-luna` with reasoning effort `max` ONLY for generator-backed connector configuration implementation sweep workers using the current generated artifacts, including their bounded implementation pilots. Use native Codex `gpt-5.6-terra` with reasoning effort `max` for everything else: non-generator connector work, foundations, planning, investigation, diagnosis, review, and newly started no-mistakes managed steps. This narrowed routing supersedes the broad all-Terra 2026-08-05 pin and every earlier Claude/other-model policy. Preserve existing work when changing models; never restart the shared no-mistakes daemon merely to force configuration uptake because that would interrupt other in-flight runs.

- 2026-08-05: NO REDACTION. This SUPERSEDES the 2026-07-27 and 2026-08-02 redaction entries below and replaces the whole approach. Secrets are protected by ENCRYPTING THE RECORD IN CONFIG - that is, encrypted at rest in the credential/configuration store - and nowhere else. Runtime output is NEVER stripped: live responses, error messages, logs, approval previews, reports and fixtures all keep their complete content. Rationale from the captain: redaction does not mask, it DELETES - the engine removes the field and leaves a `<field>_redacted: true` marker, so an agent reading that record can neither see the value nor infer it, and a human debugging a failure loses the very thing that explains it. That causes agent drift and makes incidents unreadable. Protect the secret where it rests; never blind the operator or the agent at runtime. CONSEQUENCE TO FIX: the engine currently FORCES redaction - `validateSensitivePolicy` in internal/connectors/engine/bundle.go requires any operation marked `secret_sensitive` or mutation_class "secret" to declare at least one `redact_fields` entry. That validator now contradicts this policy and needs its own change.

- 2026-08-05: DO NOT start work on the firstmate repo itself unless the captain explicitly asks for it. A diagnosis found while doing something else is REPORTED, never fixed on my own initiative. The captain's diagnostic question ("why is X still doing Y?") is a request for an answer, NOT authorization to change tooling. Origin: I read "why are we not using gpt-5.6-terra max, it still uses xhigh" as approval, dispatched a firstmate-repo job, and it opened kunchenguid/firstmate#1714 at 00:55 on 2026-08-05 - 451 additions across 15 files including remote-secondmate paths this fleet does not use - for a bug the captain's own ~/.codex/config.toml already worked around. Before ANY such action, check whether upstream already carries the fix or a release includes it, and ASK the captain first. This applies to outward-facing actions generally: opening PRs, posting public comments, closing PRs.

- 2026-07-21: For karthik.page implementation and validation, autonomously fix all security findings that are reversible, non-destructive, and contained to the approved local branch/PR scope; do not interrupt for each finding. Continue to require explicit confirmation for production access or deployment, credentials/secrets, destructive or irreversible actions, public data/retention policy choices, chat/model activation, and other changes that expand external behavior or authority.
- 2026-07-21: For karthik.page, prefer a handcrafted, nerdy Vim/terminal visual identity with code-owned ASCII animation over a conventional portfolio layout or a generic "AI-coded" appearance. Preserve evidence-only content and accessible normal navigation; use generated images as concept references for coded visuals rather than shipping a raster-heavy site.
- 2026-08-05 (SUPERSEDES the 2026-07-25 entry below on parent-branch merges): MERGES TO A `main` BRANCH REMAIN THE CAPTAIN'S EXCLUSIVELY - always ask, every time, per merge. Firstmate now has STANDING FULL AUTHORITY to merge child/sub-PRs into a parent branch or parent draft PR without asking, provided the child is green and clean. Rationale: parent-branch merges are internal stacking moves that only assemble the parent for the captain's own main-merge decision, so gating them on the captain added round trips without adding a real decision. The captain's approval gate is the single main merge at the end.
- 2026-07-25: MERGES TO MAIN ARE THE CAPTAIN'S EXCLUSIVELY. Firstmate must never merge (or fast-forward/land) anything to a `main` branch without the captain's explicit per-merge word or an explicit campaign-scoped merge authorization. The 2026-08-01 50-connector completion campaign has that standing authority for complete, current-code-matched green connector PRs; release-please/release PRs and all other main-bound work still require explicit per-merge approval. Firstmate may merge into feature/integration branches when separately authorized.
- 2026-07-27, amended 2026-08-04: Branch names should follow each repository's committed convention. Where a repository uses the common `<type>/<description>` form, select an accurate conventional type such as `feat`, `fix`, `docs`, `ci`, `test`, `refactor`, `perf`, `build`, or `chore` and a concise descriptive slug. AMENDMENT 2026-08-04: the Firstmate `fm/*` branch namespace is explicitly acceptable and needs no override. The original entry forbade it; the captain confirmed on 2026-08-04 that `fm/*` is fine, so `bin/fm-brief.sh`'s generated `git checkout -b fm/<task-id>` step is correct as-is and must not be overridden when writing a brief. Commit message types and PR titles still follow the repository's convention.
- 2026-07-27: PM binary releases and website deployments are independent products and must not trigger each other. Prioritize the first numbered PM binary release as `v0.1.0`; connectors ship inside the PM binary, with compatible connector fixes delivered through subsequent PM patch releases such as `v0.1.1` unless a separately approved connector packaging model is introduced.
- 2026-07-27: For the PM `v0.1.0` release effort, retain release-documentation polish, superseded-PR notices, valid unrelated cleanup/refactoring already discovered, and fuller connector-release guidance. Parallelize non-overlapping slices rather than dropping them from the overall effort, while keeping the binary release PR focused and independently releasable.
- 2026-07-27: WhatsApp is local-first: an authorized local user or agent must be able to read inbound message content and inspect outbound message content, including approval previews, so it can understand and act on the conversation. Do not redact message bodies from trusted local operational output; redact credentials and administrative secrets, and prevent accidental message-content persistence in logs, errors, telemetry, reports, fixtures, or other non-operational surfaces.
- 2026-08-02: Across connector responses, never redact non-secret business or operational data from trusted local output because local agents need complete data to understand and act. Redact only genuine secrets or credentials, and prefer preventing sensitive response persistence in logs, errors, telemetry, reports, and fixtures over stripping actionable data from the live local response.
- 2026-08-02: Firstmate and all implementation workers must use the dedicated GitHub identity Alfred (`alfred-polymetrics-ai`) for issue/PR creation, comments, implementation commits, branch pushes, and delivery automation. Karthik's ordinary local GitHub CLI/SSH context remains the default human identity and exclusively owns approvals, protected exception decisions, ruleset/branch-protection administration, and final main-merge authority. Use isolated GitHub CLI config and SSH identity injected only by a Firstmate launcher; never use `gh auth switch` as the operating model or share Karthik's credential with workers.
- 2026-07-27: Prioritize PM's unsigned source-build Homebrew formula and Linux archive/native-package release path ahead of Windows distribution. Homebrew must not wait on Apple Developer certificates; defer Windows work when it competes for attention, while retaining already-green Windows foundation work for a later captain-approved merge.
- 2026-07-28: During grill-me decision sessions, explain every recommendation in enough detail to show how it works and include at least one concrete example before asking for the decision.
- 2026-07-29: For the thaalam VPS, Herdr is required as the primary Firstmate task runtime; tmux is recovery fallback only. Do not convert a Herdr setup blocker into a recommendation to stay on tmux; surface and complete the explicit Herdr prerequisite path.
- 2026-07-29: The thaalam end state must support Herdr-backed coding and connector implementation review plus Pi/Codex-driven connector certification using credentials stored in Vaultwarden. Preserve the no-raw-secret-to-agent boundary by routing credential use through a trusted broker/executor unless the captain explicitly decides otherwise after reviewing the risk.
- 2026-07-29: For thaalam and similar managed hosts, package all non-interactive human-required changes into reviewed fixed-scope run-once scripts with exact copy/run commands, published checksums, sanitized verification, and success-only deletion of uploaded or temporary scripts. Never ask the captain to expose passwords or repeatedly paste long command blocks; keep unavoidable account, MFA, recovery, and device-login steps in a separate interactive checklist, while preserving required live configuration and durable non-secret evidence.
- 2026-08-01: For connector parity, include every documented delete or destructive operation under its canonical public command name, backed by typed closed schemas and plan -> preview -> explicit approval -> execute; never hide such operations behind synthetic duplicate commands.
- 2026-08-01: For the authorized 50-connector completion campaign, keep exactly five selected connectors in parallel and refill each completed slot immediately until all 50 merge. Run complete independent review, tests, documentation, lint, push/PR updates, and CI in parallel for those five; serialize rebases and final merges under the campaign-scoped merge authority. Worker and reviewer model pins are superseded by the 2026-08-03 entry below. Every review pass must still inspect the complete branch diff comprehensively, continue beyond the first valid finding, and enumerate all material substantiated issues in one pass.
- 2026-08-01: Before any connector is called review-ready, re-audit its operation ledger against the provider research and authoritative documentation: every documented operation must have an exact, truthful executable, blocked/planned, or justified excluded disposition; all delete and destructive operations must use canonical commands and the approved typed safety flow; validation must fix rather than waive parity gaps.
- 2026-08-02: Connector implementation PRs must enforce target path ownership: keep implementation under the target connector definition plus connector-owned fixtures/tests, target generated docs and necessary shared indexes/goldens, and explicitly approved target native/hook surfaces. Separate generic shared runtime/tooling foundations into their own PRs and reject unrelated connector or broad generated churn from connector lanes; a green content-based connector-boundary check is not sufficient proof. Auto-detect connector implementation diffs so omitting a label cannot skip enforcement, provide local hook feedback, and make required remote CI plus branch/ruleset protection the authoritative merge boundary because client-side hooks are inherently bypassable. When another file is genuinely necessary, the worker must stop and ask through Firstmate with exact paths, reasons, alternatives, and a bounded exception proposal; only the captain's explicit durable approval authorizes it, never an agent-added label, manifest, or exception. GitHub issues/sub-issues and PRs/sub-PRs are the canonical execution record: material updates, blockers, permission requests, decisions, integration, checks, and review dispositions must be mirrored there. Non-bypassable exception approval requires a separate human approver identity or captain-held approval service that worker credentials cannot use.
- 2026-08-03: Reddit is cleared for FULL operation parity including all four writes - vote, comment, subscribe, save. Rationale, verified against Reddit's own docs: self-service OAuth registration closed in late 2025 under Reddit's Responsible Builder Policy, so every credential - free tier included - is issued only after Reddit reviews and approves the applicant's stated use case, and commercial use requires the user's own separate paid agreement. PM never calls Reddit's API; the user does, with keys Reddit vetted for their purpose. The commercial-contract gate is therefore the user's to satisfy, not PM's, and the reddit-commercial-contract-gate hold is resolved on that basis. Reddit's rule that votes must be cast by humans is disclosed in the connector docs rather than used to withhold the operation - the user holds approved credentials and bears responsibility for their use, consistent with the retention ruling. Do not describe the connector as endorsed or sanctioned by Reddit.
- 2026-08-03: Data-retention and deletion-propagation obligations are the USER's responsibility, not the system's. Reddit's 48-hour deletion-propagation guidance for persisted user content is handled by the user operating the tool; PM does not implement automatic deletion propagation. Connector docs must state that plainly so the obligation is visible rather than silently assumed. This resolves the reddit-data-retention-policy hold. It does NOT resolve the separate reddit-commercial-contract-gate (Reddit's terms require written permission and a contract for commercial use of the Data API, and PM is monetized), nor Reddit's rule that votes must be cast by humans, which bears on whether a pm reddit vote command may exist at all.
- 2026-08-03: A connector is REACHABLE as `pm <connector> <command>` if and only if its bundle has a `cli_surface.json`. There is no separate enable switch; `engine.synthesizeCommandSurface` returns nil when `CLISurface` is nil, and the CLI then answers `unknown command`. Verified empirically on 2026-08-03 by building `./cmd/pm` and running it: on `main` all 13 connectors carrying that file were reachable (amazon-sqs, asana, ashby, bahmni, bitbucket, freshchat, github, gong, google-ads, hubspot, stripe, xero, zendesk-support) and on a branch missing six of those files exactly those six answered `unknown command` - a clean 1:1 with no exceptions. The other 537 connectors load and sync but are not usable from the command line. Before accepting any connector as complete, BUILD AND RUN the tool against it rather than inspecting files; several lanes have reported done at states that did not survive that check.
- 2026-08-03: Run the connector campaign as a ROLLING window, not a batch. The moment a connector merges, dispatch the next one immediately - never let a slot sit empty waiting for the rest of a wave to finish. Window size is 10 while token budget is plentiful and 5 otherwise; when stepping down from 10 to 5, let in-flight lanes finish and merge rather than parking them mid-flight, so the reduction costs nothing. Keep a ranked next-up queue ready so a refill needs no deliberation: prefer connectors already nearest the bar, then those with existing unlanded work, then the rest by documented-operation coverage. Before every dispatch, verify the assigned worktree is not already claimed by another live lane and belongs to this home's own clone - on 2026-08-03 four separate misassignments occurred in one hour, including two connectors handed the same worktree simultaneously and one handed a secondmate's clone; workers caught all four, but the assignment is not trustworthy on its own.
- 2026-08-03: The complete operation-parity bar applies to EVERY connector, retroactively to the ones already merged to `main` and forward to every new one; there is no forward-only exemption. A connector is complete only when every documented provider operation maps to an ETL stream, a reverse-ETL write, or a direct read AND is individually reachable as its own `pm <connector> <command>`. An ETL-only connector does not qualify however many streams it has, and being Tier-3 native with a dynamic schema does not exempt the command surface. Coverage files must not claim more than the code delivers; every documented operation carries an executable, blocked-with-named-reason, or justified-excluded disposition with a source citation, and no operation may be left with no disposition at all. Docs and the website catalog must reflect the connector's final state before it merges. Merge each connector as it reaches that bar; keep exactly five connectors in flight at a time and refill a slot only as one lands. Do not run broad all-connector parity audits - work incrementally against issues and sub-issues toward the 50-connector release. Every connector brief must carry this bar as a blocking requirement, because workers default to building streams and writes and stopping there: on 2026-08-03 the whatsapp-web, linkedin-web, and twitter-web lanes all reported done with no `cli_surface.json` at all. See `data/decisions/cli-parity-bar-retroactive-scope-2026-08-03.md`.
- 2026-08-04: A connector command marked `availability: implemented` in `cli_surface.json` can still FAIL at runtime. Marketo declared 28 direct-read commands implemented that all errored with `connector_command_blocked` because the entries omitted the required `api_surface [{method,path}]` array (`commandrunner/runner.go:421`); its true reachable count was 275/327, not the 303/327 the files claimed. Reading files therefore CANNOT verify the bar - only executing can. Before clearing any connector as meeting the parity bar, verify its commands actually run: at minimum check every `implemented` direct-read entry carries `api_surface`, and prefer a live run against a mock provider as Marketo did. Nothing in CI catches this class of defect, so the thirteen connectors already enabled on `main` have unverified implemented counts too.
- 2026-08-04: A connector is NOT 'green' or 'done' merely because its CI passes. CI does not test the parity bar. A connector counts as complete ONLY when every documented operation is executable or blocked for a reason that no amount of our own work would fix - a provider-side restriction, an outbound webhook no client can invoke, or a rule we deliberately refuse such as a raw request passthrough. A connector whose remaining operations wait on OUR foundation work is NOT complete and must never be reported as green; it is 'CI-green, blocked on <named foundation>'. State the blocking foundation by name every time. On 2026-08-04 freshchat, recurly, airtable and marketo were each reported as green while waiting on engine capabilities - that was wrong and the captain corrected it.
- 2026-08-04: BOUNDED EXCEPTION to the Alfred identity rule. No `alfred-polymetrics-ai` credential exists in this environment - verified by a lane that checked gh auth, GH_TOKEN, hosts.yml and the workspace config and found only the captain's own account. Because GitHub authorship cannot be reattributed after creation, lanes were stopping rather than creating issues under the wrong identity. The captain authorised using `gh` under the available identity for ISSUE creation when Alfred is unavailable. This covers issues only; it does not extend to approvals, merges, or branch-protection administration, which remain the captain's exclusively. The underlying credential gap is filed as cli-pipeline-alfred-identity-r1 and should still be fixed.
- 2026-08-04: Every foundation dispatch follows a fixed order: (1) create the GitHub issue and its sub-issues FIRST, so the execution record exists before any code, (2) implement through the GSD workflow, (3) validate through no-mistakes. Do not start implementation before the issue tree exists - GitHub issues and sub-issues are the canonical execution record, and a foundation change without one leaves no trace of why it was built or what it unblocks. This applies to shared-runtime and engine work specifically; connector lanes already inherit their issue tree from the connector campaign.
- 2026-08-04: Foundation and shared-runtime implementation - anything outside a single connector's own bundle, such as connector-engine changes - runs on Opus 5, not Sonnet 5. Sonnet 5 remains the implementation model for connector-local work. This refines the 2026-08-03 model policy rather than replacing it: planning, design, investigation, audit and review still run on Opus 5 everywhere. Additionally, before implementing a shared capability, research established best practice first and put the findings in the brief - do not let the implementer invent a design for something with known prior art.
- 2026-08-03: The fleet runs on Claude, not Codex or Pi/Codex (supersedes every earlier Codex model pin, including the 2026-08-01 connector-campaign pins). Implementation work runs on Sonnet 5 and drives the GSD workflow; planning, design, investigation, audit, and review work runs on Opus 5. no-mistakes review and fix-review use native Claude with Opus 5 at `xhigh`. All work continues to be dispatched and supervised through Firstmate. Implementation was parked at the time of this change while the switch was applied.
- 2026-08-04: Present every main-bound merge to the captain INDIVIDUALLY once it is green, with the complete picture, and wait for their explicit word - never merge on green alone and never batch several merges into one approval request. This holds even when an instruction reads like blanket merge authority: on 2026-08-04 "finish and merge both the foundation to main" was firstmate's to drive to green and present, not to merge, and the captain corrected that reading. When an instruction could be read as granting merge authority, confirm rather than assume. The 2026-07-25 rule that merges to main are the captain's exclusively is unchanged; the campaign-scoped exception for complete, current-code-matched green connector PRs is the only standing relaxation.
- 2026-08-04: Terminology - say "connector validation" for the connector definition checker (`cmd/connectorgen/validate.go`, run through `make verify`), and reserve "validation" on its own for the no-mistakes shipping pipeline every lane drives before a PR. Firstmate had been using "validation" for both, which made the parity explanation ambiguous; the captain asked for the split. This is a communication rule, not a code rename - renaming the `connectorgen validate` command itself was raised, deliberately kept off the critical path to 50 connectors, and needs its own scoping if the captain wants it.

## Website and blog

- **The latest blog post is always the header/featured post.** Stated 2026-08-06. This is now
  enforced structurally by sorting on `publishedAt` rather than by manual source order in
  `website/lib/blog.ts`, so appending a post cannot break it.
- Blog posts are authored inline as `BlogPost` data in `website/lib/blog.ts`, not as markdown.
  Images live under `website/public/blog/<slug>/`.
- The captain writes the blog in first person: plain, specific, unhurried, willing to describe what
  went wrong without drama. No marketing language. When a worker drafts one, it must read the
  existing posts as the voice specification rather than work from a description.
- The captain asked for storytelling structure over report structure: what happened, in order, with
  a turn in the middle. "Simple enough for a child" means short sentences and concrete-before-
  abstract - never explaining what a test or an API is. The reader is an engineer.
- The captain generates blog imagery themselves in ChatGPT from prompts firstmate writes. Codex here
  has no image-generation tool installed.

## Re-gate connector parity before every connector merge (2026-08-06)

The captain's standing rule: **never merge a connector PR without first re-running its parity
gate against current `main`.** Foundations land constantly, and an operation recorded as blocked
or excluded may already be reachable — its recorded reason was written against an older `main`.

The re-gate must let the tooling reclassify. **Never hand-edit an `api_surface` reason**; a
hand-edited reason destroys the only evidence that the count is real.

Report before/after executable counts. An operation may remain blocked only when it genuinely
needs foundation work that does not exist yet, and the record must name that foundation so the
remaining operations can be completed later.

The only exception is a connector whose remaining operations all depend on unbuilt foundation —
merge it, with each blocked operation naming what it waits for.

## Database replication dependencies APPROVED (2026-08-06)

The captain approved **`pglogrepl`** and the equivalent human-gated replication libraries needed for
the other database engines — MySQL, MariaDB, SQL Server, Oracle — granted together rather than one at
a time, so native CDC is no longer blocked on a per-engine approval round.

**The approval is conditional and the condition is binding: each library must be secure and tested.**
Before adding any of them, establish and record in the PR: current maintenance status and last
release, known CVEs, transitive dependency footprint, licence, and measured binary-size delta. A
library that cannot be shown maintained and clean is refused back to the captain, not added quietly
under this approval.

Scope: database replication libraries for native CDC only. It is not a general licence to add
dependencies; anything outside that set still needs its own decision.

## Deployment model: Opus orchestrates, Sonnet executes (captain ruling 2026-08-07)

Standing instruction for how fleet work is deployed, to cut cost without cutting quality.

**Opus 5 holds the lane and owns every judgement call**: reading provider artifacts and deriving
operation counts, deciding whether a regenerated diff is genuinely mechanical, answering pipeline
gates, resolving ask-user findings, deciding when to escalate a shared-gate change rather than make
it, and all review.

**Sonnet is delegated the mechanical work** through sub-agents the Opus lane spawns: generating and
editing bundle files, running commands and collecting output, regenerating catalogs and ledgers,
fixture assembly, and repetitive per-operation authoring.

The split follows where errors are expensive. The costly failure mode in this programme is a
connector that DECLARES capability it does not have; that comes from judgement, not typing. Bulk
authoring is where the tokens go and where a cheaper model is genuinely sufficient.

Rules that make the split safe:

- A sub-agent never answers a pipeline gate, never resolves an ask-user finding, and never decides
  that a diff is safe to commit. It produces work; the Opus lane accepts or rejects it.
- The Opus lane inspects every regenerated artifact diff itself before committing. It must not
  delegate "confirm this is purely mechanical" - that judgement IS the safeguard.
- Anything touching a shared gate, a test, or a capability declaration stays with Opus.
- When a sub-agent reports something is done, the Opus lane verifies it against reality rather than
  accepting the report. A report is evidence, not proof - that is the defect class this programme
  has spent two weeks removing.

Applies to the connector sweep first, then any future bulk-authoring lane. Foundation lanes stay
wholly on Opus; their work is judgement almost end to end.

## Second-opinion pass on worker decisions (captain ruling 2026-08-08, RESTATED after firstmate disobeyed it)

**EVERY decision a worker escalates goes to GPT-5.6 Sol at max reasoning BEFORE firstmate answers.
No exceptions. No firstmate-invented thresholds. No "this one is only documentation".**

On 2026-08-08 firstmate applied this rule to one of four escalations and decided the rest alone,
having privately judged the others "not worth a round". That was disobedience, not judgement. The
captain set the scope; firstmate does not get to narrow it. If firstmate believes the rule is too
broad or too costly, it says so to the captain and waits - it does not comply selectively and stay
quiet about it.

Procedure:

1. **Open the LANE'S OWN WORKTREE for Sol** - the exact directory the worker is building in, read
   from `worktree=` in `state/<id>.meta`, not the primary checkout and not origin/main. That worktree
   holds the in-flight state the decision is actually about, including work not yet pushed. Pass that
   path explicitly and name the branch. Never hand Sol a summary - a summary launders firstmate's own
   framing into the second opinion and defeats the purpose.
   `codex exec --skip-git-repo-check --model gpt-5.6-sol -c 'model_reasoning_effort="max"' '<question + the worktree path + branch + options>'`
   Invoke it from a scratch directory so the tool itself never writes into that worktree: **Sol READS
   the lane's worktree and never modifies it** - a second agent editing files under a live worker is
   how two owners corrupt one branch. It can take 5-10 minutes; run it in the background and WAIT for
   it rather than sending a decision before it returns.
2. Read what comes back. It is evidence, not the answer. Check it against the code, the captain's
   standing orders, and what the worker actually reported.
3. Firstmate makes the final decision and owns it. Never forward Sol's answer as the ruling, and never
   tell the worker a second model was consulted - the worker receives one clear decision.
4. Say so in the captain-facing message when the two disagreed, and how it was resolved.

Cost is real - roughly 9 cents and 18k tokens per call, mostly fixed overhead - and it is the
captain's cost to weigh, not firstmate's to avoid unilaterally.

Why: this programme's dominant defect is confident work that did less than it claimed, and several
instances survived because one reader accepted one account. The single Sol run that did happen caught
an assumption inside firstmate's own ruling. That is the point.

Scope: every worker-escalated decision, including pipeline ask-user findings and documentation
questions. It does not apply to the captain's own decisions - those go to the captain, never to a
model.

## Rate-limit declarations ship WITH parity, not as a separate sweep (captain ruling 2026-08-08)

Every connector brought to parity must land its `rate_limits.json` in the same work, with real
policies, not a `state: unknown` placeholder.

The active limiter already exists and is wired at `Runtime.requesterFor`, the single choke point every
request passes through, so ETL, reverse ETL, direct read, direct write and check are all covered.
`Admit` runs before a request and `Observe` consumes the response. **Never build a second mechanism.**
`internal/connectors/connsdk/rate_limits.go` owns the schema: `policies[]` with `id`, `source`
(URL + mandatory `retrieved_at`), `selector` (`endpoints`, `tiers`, `auth_types`), `scope`, `budgets`.

`selector.auth_types` exists specifically so a provider with different ceilings per auth flow can be
declared. "We cannot tell which auth flow is active" is not an acceptable reason for `unknown`.

**A declaration is not proven until it has been tested the captain's way:** deliberately drive enough
traffic to blow the configured budget, show our limiter stops us, then issue an equivalent request
with the SAME credential outside `pm` and confirm the provider still had headroom at that moment. It
passes only if we stopped ourselves BEFORE the provider would have. Stopping after a 429 is a failure
- by then a user's sync has already broken.

Reach 100% parity connector by connector, and carry the rate-limit declaration with each one rather
than deferring it to a later sweep. Deferred foundation is how 551 bundles ended up with zero
declarations while the machinery to use them sat finished and idle.

## Model split: Sol decides, Luna runs generated connector sweeps, Terra does everything else (captain ruling 2026-08-08, clarified later the same day)

**GPT-5.6 Sol is for DECISIONS ONLY** - the second-opinion pass on worker-escalated decisions, and
nothing else. It reads, analyses and advises; it never implements.

**GPT-5.6 Luna at max is reserved for generator-backed complete connector-configuration
implementation sweeps** using the current connectorgen/provider-artifact outputs, including bounded
connector implementation pilots that prove that generated path.

**GPT-5.6 Terra at max does everything else** - non-generator connector authoring, foundation
changes, investigation, planning, review, verification, documentation, and no-mistakes managed
steps. Those workers launch on `codex` / `gpt-5.6-terra` / `max`.

Never dispatch an implementation worker on Sol. Never ask Sol to write code, edit a file, or drive a
pipeline. If a second opinion concludes that work is needed, it goes to Luna only when it is part of
the generator-backed connector implementation sweep; otherwise it goes to Terra.

`config/crew-dispatch.json` keeps codex/gpt-5.6-terra/max as the default and carries one narrower
Luna/max rule for the generator-backed connector implementation sweep. A Sol implementation lane,
a Luna lane outside that rule, or a Terra lane inside that rule is a routing error to correct rather
than a preference to respect.


## Generated docs are per-connector, never repo-wide (captain ruling 2026-08-08)

A change to one connector regenerates and validates **only that connector's** generated artifacts.
Never regenerate or drift-check the whole tree for a single-connector change.

The repo carries ~1,024 generated files. A repo-wide regenerate on a one-connector PR pollutes the
diff, hides the real change from a reviewer, and lets an unrelated stale file block an unrelated lane -
or lets one lane silently rewrite another's output.

The generator must therefore expose a connector-scoped generate and check mode, and CI's drift check
must verify only the connectors a PR actually touches. Where that scoping does not exist yet, add it
in the lane that needs it rather than accepting a repo-wide regenerate "just this once".

This is the same one-owner discipline the rest of the codebase follows: a change touches what it owns,
and proves it touched nothing else.

## Map everything the API documents; never refuse for a missing foundation (captain ruling 2026-08-08)

Connector generation maps **every operation the provider's API documents**. It never refuses, drops,
or fails closed because an engine executor or other foundation is missing. That foundation gets built
when the connector is built.

What varies is **availability, not inclusion**:

- `implemented` only where the runtime can genuinely execute the command today.
- Otherwise a non-implemented availability carrying a **machine-checkable named dependency** stating
  exactly what is missing.

Nothing is silently omitted. A foundation gap becomes a visible, named item instead of a refusal.

**The one line that never moves:** never mark a command `implemented` that the runtime cannot execute.
`TestEveryImplementedCommandPassesRuntimePreflight` enforces it, and a command that validates clean
then fails on invocation is the defect class this programme exists to remove - 174 Docker Hub commands
once did exactly that.

Where our existing surface declares an endpoint the fetched artifact does not contain, **keep it and
flag it** as present-in-surface-absent-from-artifact with that reason. Published specifications are
often incomplete; deleting a working command because a provider under-documented it is worse than
recording the discrepancy.

**Source discovery is multi-source and traversal-capable.** A missing, incomplete, or unparsable
OpenAPI document does not make a provider unmappable. Try every authoritative provider-published
shape needed for that connector: OpenAPI/Swagger with referenced documents and webhooks, then
provider Postman/GraphQL/AsyncAPI/WSDL/API Blueprint or equivalent exports, then traverse the
provider's official API reference/index/operation pages or provider-owned documentation source.
Combine incomplete sources and retain exact per-operation provenance in the parity JSON. Never use
third-party wrappers or inferred SDK methods as evidence, and never invent an operation when the
authoritative sources remain ambiguous. Keep traversal cached, bounded, connector-scoped, and
resumable so Luna can generate `api_surface.json`, `operations.json`, and `cli_surface.json` quickly
across the complete connector sweep.

## Luna is scoped to the generated connector implementation sweep (captain ruling 2026-08-08, supersedes the earlier mechanical-sub-agent split)

Do not dispatch Luna merely because work is documentation, prose relocation, evidence repair,
catalog regeneration, count correction, or otherwise mechanical. Those tasks stay on Terra unless
they are an intrinsic part of the generator-backed complete connector-configuration implementation
sweep.

Within that sweep, Luna may own the implementation lane and its bounded generated-connector pilots.
The lane must still verify generated output against provider evidence and runtime behavior; generated
does not mean trusted. Capability declarations, reachability, destructive safety, connector-scoped
diff ownership, and tests remain blocking requirements.

Model roles now: **Sol decides** (second opinions only), **Luna implements the generated connector
configuration sweep**, and **Terra handles everything else**. Never ask Sol to implement, and never
route unrelated work to Luna.

## Certified-connector merge order (captain ruling 2026-08-10, supersedes the earlier GitHub -> Zoom -> Docker Hub order)

Land certified connectors in this order, one at a time, each proven at its own parity bar before the
next starts:

1. **GitHub**
2. **Docker Hub**
3. **Reddit**
4. **Zoom** — last

The earlier order put Zoom second. Firstmate measured the branches directly and Zoom was at 145 of
1,913 reachable commands (7.6%) while Docker Hub was already complete at 54 of 54 and Reddit at 220
of 230. Ordering by readiness rather than by when the work started is the standing rule: **a finished
connector never waits behind an unfinished one.**

Re-measure readiness from the branch itself, never from a worker's report or a PR body:
`git show origin/<branch>:internal/connectors/defs/<connector>/cli_surface.json`. Firstmate stated
Zoom was "progressing, not stuck" from status prose and was wrong; the commit record showed a
~24-hour gap with no implementation commits. **Commit cadence, not a busy pane, is the evidence for
whether a lane is moving.**

Zoom stays a live lane at the back of the queue, not a cancelled one.

## Force authorisation is scoped to the local validation mirror, never origin (captain ruling 2026-08-10)

The captain authorised a force-push for `cli-github-parity-extract-r1` on **2026-08-10**, scoped to
one branch ref inside the LOCAL no-mistakes mirror remote
(`~/.no-mistakes/repos/<hash>.git`), with `origin` untouched and the shared gate not ejected.

This is a precedent for the shape, not a standing grant. Before asking again, establish and state:
1. `git remote -v` proves the rejecting remote is the local mirror, not `origin`.
2. `git cherry <current-head> <stale-head>` was run, and every unmatched commit was verified
   **behaviourally** present in the current head - not inferred from commit identity.
3. The cache is rebuildable, so the action is reversible.

A force that would touch `origin`, another lane's ref, or unverified content is a different and much
larger question and must be asked as one. See [[verify-counts-before-asserting]] - firstmate asserted
"every gate-fix commit is present" from one sampled commit and was wrong; the evidence must exist
before the assurance is given.

## The goal is 50 CERTIFIED connectors, chosen for daily usefulness (captain ruling 2026-08-10)

**Target: at least 50 certified connectors.** Small is fine — the bar is that a connector is
genuinely useful to a user on a daily basis. Everything beyond those 50 may ship UNCERTIFIED via the
generated sweep.

Certified means the GitHub / Docker Hub treatment: complete documented parity, every command proven
reachable against a real built binary, zero `unsafe_or_disallowed`, rate limits declared, and a
`certification.json` record in the bundle.

**Measured baseline 2026-08-10:** main carries **17** `certification.json` records against 36
connectors with a command surface. The sweep branch carries the same 17 against 234 surfaces - it adds
198 surfaces and **zero** certifications. So the gap to the goal is **33 more certifications**.

**Selection consequence, and it is counter-intuitive:** connector SIZE is not the criterion, daily
usefulness is. Zoom has 1,913 documented endpoints, is at ~8.7%, could not be extracted by the
generator at all, and is being hand-built category by category at roughly 22 operations per three
hours. Under this ruling Zoom is a poor use of certification effort - it is the largest and slowest
and not obviously the most-used. Prefer many small everyday tools over one enormous one.

Count certifications with:
`git ls-tree -r --name-only <ref> internal/connectors/defs/ | grep -c 'certification\.json$'`

## Certification count correction: the honest baseline is ZERO, not 17 (2026-08-10)

Firstmate told the captain "17 certified, 33 to go." **That was wrong.** The 17 came from
`grep -c 'certification\.json$'` - a FILENAME match, not a validity check. Six of the matches are
Employment Hero fixture writes (`fixtures/writes/archive_certification.json` etc.), which are test
data. The remaining eleven are bundle contracts. `internal/connectors/certifications` on `origin/main`
is **empty - zero accepted live certification artifacts.**

**So the honest baseline against the captain's 50-certified goal is ZERO of 50.**

Never count certifications by filename again. A connector is certified only when the real built binary
reached every command, no `unsafe_or_disallowed` remains, rate policy is declared, parity and docs are
complete, and a fresh **accepted live artifact** exists. Adding a `certification.json` does not certify
anything.

Evidence and the ranked shortlist: `data/cli-daily-use-top50-connectors-r1/report.md`.

**The road is still short:** 33 of the recommended 50 already have a fully mapped command surface on
`origin/fm/cli-mass-artifact-materialize-r1`, so they are proving jobs rather than builds; only 17 are
unmapped. Best first targets, the only high-value candidates with complete mapped surfaces:
**Docker Hub 4/4, DocuSeal 9/9, Amazon SQS 23/23, Google Calendar 38/38** (Google Calendar gated on the
open OAuth client-ownership decision).

## Query GitHub through our own connector, not gh-axi, and stop polling (captain ruling 2026-08-10)

**Use `pm github` (REST) rather than `gh-axi` for review/check/issue lookups where practical.**
On 2026-08-10 the fleet exhausted GitHub's hourly **GraphQL** quota - 5002 of 5000 used, zero
remaining - which stopped two workers mid-task. In the same window the **REST** quota showed only 11
of 5000 used. `gh-axi`/`gh` use GraphQL for `pr checks`, `pr view` and `pr list`; our own connector
uses REST. So the exhausted axis is supervision traffic, not connector traffic, and our own connector
sidesteps it entirely while dogfooding the product. REST also parallelises better.

**Polling discipline (applies to firstmate itself, which caused much of the burn):** do local work to
a coherent stopping point, THEN check remote state once. Validation and CI take minutes - `verify`
runs around 20 minutes - so check on that cadence, never per-second. While one implementation runs,
advance the next piece of work rather than watching it.

## GitHub merges only after LIVE certification, not on green checks (captain ruling 2026-08-10)

GitHub must be proven with the real `pm` binary against the live provider and carry live certification
docs before merging. Green CI is not sufficient.

**Conflict to resolve with the captain:** the literal condition "test API to API" is currently
IMPOSSIBLE - `runConnectorETL` rejects every destination that does not implement
`synccontract.DurableETLDestination`, and no production connector implements it
(`data/cli-github-etl-reverse-etl-gap-map-r1/report.md`). Firstmate is proceeding on the achievable
reading: GitHub's own API surface proven live with `pm`. If the captain meant literal API-to-API
delivery, GitHub is blocked behind a shared foundation several waves away.

## The warehouse is ALWAYS the mediator - never source-to-destination directly (captain ruling 2026-08-10)

Every flow goes through the warehouse. An API source does ETL into the warehouse first, and the
destination is then fed FROM the warehouse. No zero-copy, no direct source-to-destination path, in
either direction.

**Why the captain wants it:** the warehouse becomes the durable record of what has been delivered,
what has not, and what remains. A Parquet materialisation is a real artefact you can diff against;
a direct pipe leaves nothing to reconcile after a failure.

This CONFIRMS the accepted design rather than changing it - see
`data/cli-cdc-bidirectional-changefeed-design-r1/report.md`: inbound preserves real source
transactions, outbound derives a keyed delta by comparing an immutable Parquet materialisation
against the last receipt-backed baseline via DuckDB. Do not derive reverse deltas from the JSONL WAL
or from a bounded reverse-plan slice; derive them from the published Parquet projection.

## NINTH worktree collision - and the first pre-loaded with foreign UNPUSHED work (2026-08-10)

`cli-pg-f1-database-foundation-r1` was spawned into a slot already sitting on branch
`fix/3579-connector-path-ownership-guardrails`, **15 commits ahead of main with 1549 uncommitted
modified and DELETED files**, and **no matching branch on origin** - that work existed nowhere else.
The worker had already reported "isolated branch created; delivery setup starting" and was about to
proceed. Its next ordinary action would have either committed 1549 foreign files into the Postgres
foundation or cleaned them away permanently. Nothing in the tooling objected.

**Standing rule: after EVERY spawn, before the worker does anything, verify the working copy is
actually clean and on the expected base:**
`git -C <worktree> rev-parse --abbrev-ref HEAD`, `git -C <worktree> status --porcelain | wc -l`,
and `git -C <worktree> log --oneline origin/main..HEAD | wc -l`.
A spawn reporting a worktree path is NOT evidence that the worktree is usable.

## `state/<id>.meta` worktree paths are UNRELIABLE - read the pane, not the record (2026-08-10)

Two workers on the same day were recorded at one worktree while actually operating in another:
`cli-pg-f1-database-foundation-r1` recorded slot 5, actually in slot 27;
`cli-github-parity-live-coverage-r2` recorded slot 7 (detached on `main`), actually in slot 28 on the
correct pinned parity commit. **Both workers were fine. The records were wrong.**

Firstmate raised two false alarms from this - including telling the captain that Wave A was about to
destroy 1549 files it had never been near.

**Verify a worker's location from its PANE (`bin/fm-peek.sh` shows the real cwd), then inspect THAT
path.** Treat `worktree=` in the metadata as a hint, never as truth. A dirty worktree found at the
recorded path may belong to nobody at all.

Corollary: an orphaned worktree can hold real unpushed work with no live owner. Slot 5 held branch
`fix/3579-connector-path-ownership-guardrails`, 15 commits, 1549 uncommitted files, **absent from
origin entirely**, owned by no live task. It contains NO CDC work - verified zero matches for
cdc/changefeed/logical/replication - and is almost entirely regenerated `docs/connectors` output.
Low value, but it would vanish silently the moment that slot is reused.

## "API to API" means API -> WAREHOUSE -> API, on ONE connector definition (captain ruling 2026-08-10)

The GitHub merge condition is now unambiguous and it is **achievable**:

**GitHub source -> warehouse (Parquet) -> GitHub destination**, using the SINGLE GitHub connector
definition with all actions mapped inside it. There is no direct source-to-destination hop; the
warehouse is always the mediator (see the warehouse-mediator ruling above).

**This unblocks GitHub.** Firstmate previously reported the merge condition as impossible because
`runConnectorETL` refuses any destination lacking `synccontract.DurableETLDestination`. That refusal
applies to DIRECT API-to-API delivery. The warehouse-mediated round trip is a different path and both
halves already exist and are tested in isolation
(`data/cli-github-etl-reverse-etl-gap-map-r1/report.md`): GitHub into the local warehouse works with
real reads landing in Parquet, and warehouse-back-to-GitHub works via the reverse plan path.

**What must still be proven before merge:** the ROUND TRIP end to end with the real `pm` binary and
live certification evidence - not the two halves separately.

**Carry this caveat into the certification record rather than hiding it:** the reverse leg is
currently a one-shot mutation facility, NOT a resumable sync. It has no immutable workset, no
destination receipt, no delivery checkpoint and no replay identity, and GitHub's actions declare no
provider idempotency key. So the round trip can be certified as WORKING; it must not be advertised as
receipt-backed or resumable delivery until the bidirectional contract lands.

## Certification must be PROOF-BEARING, with credentials FINGERPRINTED not stored (captain ruling 2026-08-10)

**Certification is the single source of truth for "does this connector work end to end."** A checker
reads the certification FIRST, then cross-checks the code against it. Code inspection alone never
certifies. A certification without embedded proof is invalid by construction and the generator must
refuse to emit one.

**Evidence is embedded, not asserted.** The certification stores the ACTUAL request and response for
each operation it certifies - endpoint, method, headers, query, body, status, response body - plus
round-trip evidence for a flow. Not "a test passed": the transcript itself. It must stay readable by
a normal user and publishable on the website.

**Credentials are FINGERPRINTED, never stored and never encrypted into the artifact.** The captain
initially proposed encrypting secrets to the local SSH key; firstmate advised against it and the
captain accepted. Reasons, recorded so this is not re-litigated:
- encrypting with a private key is signing, not secrecy - you encrypt TO a public key;
- an encrypted blob is unverifiable by CI, by reviewers, and by the same user on another machine,
  which defeats the readable/publishable requirement;
- an encrypted secret committed to git is still a secret in git history, exposed the moment a key
  leaks or the repository's visibility changes;
- **the credential proves nothing** - the endpoint, shape, status and response are what demonstrate
  the connector works.

**Mechanics:** deterministic hash, salted per repository so fingerprints are not comparable across
installations, substituted on the way IN so a raw value never reaches disk even transiently. Redact
against the ACTUAL prepared values, never by keyword - keyword scrubbing is what failed on GitHub and
persisted a token verbatim. Response bodies are credential-bearing too, not just headers. Local
re-verification works by the user supplying their own credential at replay time and matching the
fingerprint.

## Three captain rulings 2026-08-10 (OAuth, uncertified reach, certification sequence)

**OAuth client ownership: BOTH paths.** Polymetrics may operate its own OAuth application PROVIDED it
does not expose our secrets, AND user-supplied credentials remain supported. Not either/or. This
unblocks Google Calendar (38/38, one of only four complete-surface certification candidates) and bears
on the 25 connectors that need a refresh token with no interactive login flow existing today.
Design consequence: our own client must never require shipping a client secret to the user's machine -
use the installed-app/PKCE shape, not a confidential client.

**Uncertified connectors ARE reachable by users.** They ship and are usable; certification is a
QUALITY SIGNAL, not a gate on availability. Consequence: the certification record must make an
uncertified connector's status visible to the user rather than silently equivalent to a certified one.

**Certification sequence approved:** create the certificate mechanism -> research -> certificate PR ->
continue GitHub and Postgres -> certify. Design report approved in principle; captain wants it fast
and scoped to what certifies GitHub and PostgreSQL.

## GitHub merge: firstmate recommendation 2026-08-10

**Merge GitHub WITHOUT waiting for certification, once two conditions clear:**
1. the GraphQL executability answer is good, and
2. the captain triages the code-scanning alert on PR #3970.

Reasoning: the work is verified (1,571 commands, 1,225/1,225 endpoints, zero `unsafe_or_disallowed`,
measured by firstmate from the branch). Main has not moved in 12+ hours. Unmerged work rots - rebasing
damaged committed work three times in one session. Certification certifies whatever is ON MAIN, so
merging does not weaken it.

**The disqualifier:** 309 of 768 GitHub operations are GraphQL (274 mutations + 35 queries) = 40%. A
prior look at `main` found GraphQL operation rows were METADATA-ONLY with no executor. If that holds
on the parity branch, merging would ship declared-but-unexecutable operations - this programme's
dominant defect - and they must be fixed or withdrawn first.

## Certify ONCE at full parity; uncertified ships as "community build" (captain ruling 2026-08-10)

**One provenance, not two.** Certification evidence is produced against a FULL-PARITY credential. A
user whose credential has narrower scopes receives a SUBSET of the certified surface by construction -
that is expected behaviour, not a certification failure and not a second evidence class. The record
simply states that certification was produced at full scope and that a narrower credential yields a
subset. This supersedes firstmate's earlier dual-provenance suggestion, which added schema complexity
for no gain.

**Uncertified connectors: a plain warning at point of use** - not certified - and the label
**COMMUNITY BUILD, UNCERTIFIED**. No status taxonomy beyond certified versus community build. They
remain fully reachable; the warning is the honesty, not a gate.

Both choices deliberately shrink the first slice so certification ships fast for GitHub and PostgreSQL.
