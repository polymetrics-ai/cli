1. VERDICT

 Feasible with corrections.

 Corrections required:

 1. Do not conflate the two GSD surfaces.
     - @opengsd/gsd-pi@1.11.0 is the governed runtime; its private workflow truth is the issue-local .gsd/gsd.db.
     - scripts/gsd is a separate repository prompt adapter sourced from open-gsd/gsd-core@next; its snapshot describes file-based .planning/.
     - Shepherd must use GSD Pi query/headless APIs for dispatch, never infer canonical workflow state from .planning/.

 2. Local-only state cannot also be durable Git evidence. Existing tracked .planning/ history must move to a tracked archive, while future required plans/TDD/verification evidence
    moves to a stable tracked path outside .planning/.

 3. The directory migration should not be added to issue #389. It is a broad repository-contract and state-layout change requiring a separate sub-issue under #372.

 4. Host-local GSD must become the default qualified path. Current agent-runtime/shepherd/README.md and shepherd.example.json still default to Podman. Do not delete legacy Podman
    assets during #389 or the state migration; deprecation/removal follows a successful host canary and a human-approved cleanup issue.

 5. Current issue #389 status is inconsistent. Its implementation prompt records unfinished RED/GREEN work, while its ledger says verified. Before any implementation, reconcile
    with the real diff and rerun tests. Never reset, delete, or overwrite the current work.

 Planning-turn orchestration decision: not_spawned_human_gate — the user explicitly required read-only planning.

 ────────────────────────────────────────────────────────────────────────────────

 2. SOURCE OF TRUTH

 ┌───────────────────────────────────────────┬─────────────────────────────────────────────────────────────────────────┬──────────────────────────────────────────────────────────┐
 │ Domain                                    │ Authoritative owner and location                                        │ Non-authoritative projections                            │
 ├───────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ Source and exact work product             │ Git commit/tree and exact head SHA                                      │ Working tree before checkpoint                           │
 ├───────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ Issue scope and acceptance contract       │ GitHub issue; parent relationships and PR bases in GitHub               │ Validated immutable context snapshot in protected        │
 │                                           │                                                                         │ Shepherd state                                           │
 ├───────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ PR, review, checks, comments, human       │ GitHub                                                                  │ Local summaries and cached observations                  │
 │ decisions                                 │                                                                         │                                                          │
 ├───────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ GSD workflow state and canonical next     │ Official GSD Pi 1.11.0, one issue-local .gsd/gsd.db per worktree        │ GSD-generated Markdown and query snapshots               │
 │ unit                                      │                                                                         │                                                          │
 ├───────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ Controller authority                      │ Shepherd’s protected SQLite/WAL store outside the worktree              │ Status output and committed audit summaries              │
 ├───────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ Dispatch ownership                        │ Shepherd alone: lease, generation, unit fencing, retry decision,        │ GSD may propose next; it cannot authorize itself         │
 │                                           │ stop/human gate                                                         │                                                          │
 ├───────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ External effects                          │ Shepherd allowlisted outbox and idempotency ledger                      │ Worker effect requests                                   │
 ├───────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ Model acceptance                          │ Shepherd’s observed session/runtime evidence                            │ Desired model configuration                              │
 ├───────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ Required issue plan/TDD/verification      │ docs/agentic-delivery/issues/<issue>/<slug>/                            │ Local .planning/ compatibility projection                │
 │ evidence                                  │                                                                         │                                                          │
 ├───────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ Historical planning evidence              │ docs/agentic-delivery/archive/**                                        │ Old links resolved through the migration map             │
 ├───────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ Repository GSD adapter resources          │ .agents/gsd/prompt-adapter/**                                           │ None under runtime .gsd/                                 │
 ├───────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ Reusable GSD Pi project policy            │ .agents/gsd/shepherd-project-template/**                                │ Copied issue-local .gsd/PREFERENCES.md and               │
 │                                           │                                                                         │ .gsd/agents/**                                           │
 ├───────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ Local prompt-adapter workspace            │ Ignored .planning/                                                      │ Never controller or Shepherd workflow truth              │
 ├───────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────────┤
 │ Test/verification result                  │ Actual command exit against an exact head                               │ Ledgers claiming commands passed                         │
 └───────────────────────────────────────────┴─────────────────────────────────────────────────────────────────────────┴──────────────────────────────────────────────────────────┘

 Shepherd is therefore the sole controller, but it does not own or edit GSD Pi’s private database schema and cannot override Git/GitHub delivery facts.

 ────────────────────────────────────────────────────────────────────────────────

 3. TARGET TREE

 ### Tracked repository tree

 ```text
   .agents/
   └── gsd/
       ├── prompt-adapter/
       │   ├── commands.json
       │   ├── local-commands.json
       │   ├── upstream.lock.json
       │   ├── prompts/
       │   │   ├── programming-loop.md
       │   │   └── issue-122-rebootstrap.md
       │   └── official-docs/
       │       └── ...
       └── shepherd-project-template/
           ├── manifest.json
           ├── PREFERENCES.md
           └── agents/
               ├── polymetrics-contract-validator.md
               ├── polymetrics-parent-planner.md
               └── polymetrics-reviewer.md

   docs/
   └── agentic-delivery/
       ├── issues/
       │   ├── 49/destructive-confirmation-gate/
       │   ├── 121/github-full-certificate/
       │   ├── 372/gsd-pi-go-shepherd/
       │   └── 389/shepherd-hardening/
       ├── archive/
       │   ├── planning-v1/
       │   │   └── ...                  # preserved legacy/non-issue .planning history
       │   └── gsd-pi/
       │       └── m001-autonomous-shepherd/
       │           └── ...              # tracked current GSD projections, not gsd.db
       └── state-migration/
           ├── PATH-MAP.md
           └── tracked-manifest.json
 ```

 ### Local, ignored, per-issue worktree state

 ```text
   .gsd/
   ├── ISSUE.json
   ├── gsd.db
   ├── gsd.db-wal
   ├── gsd.db-shm
   ├── PROJECT.md
   ├── REQUIREMENTS.md
   ├── PREFERENCES.md
   ├── agents/
   ├── phases/
   └── ...                              # GSD Pi-owned runtime/projections

   .planning/
   ├── config.json                      # commit_docs=false, no secrets
   ├── PROJECT.md
   ├── ROADMAP.md
   ├── STATE.md
   ├── phases/
   └── ...                              # compatibility workspace only
 ```

 ### Protected Shepherd state outside the worktree

 ```text
   <state-root>/deliveries/<repo-id>/issue-<N>/
   ├── controller.db
   ├── validated-context.json
   ├── audit/
   ├── locks/
   ├── runtime/gsd-state/               # only if qualified for GSD Pi 1.11
   └── gsd-home/
       ├── PREFERENCES.md
       ├── agent/settings.json
       └── agent/sessions/
 ```

 GSD_STATE_DIR must be qualified by a test. If it redirects the project workflow database away from issue-local .gsd/, remove it from the host profile rather than claiming
 .gsd/gsd.db is canonical.

 ────────────────────────────────────────────────────────────────────────────────

 4. ISSUE SPLIT

 ┌───────┬───────────────────────────────────────────────────┬────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
 │ Order │ Issue/PR                                          │ Scope                                                                                                              │
 ├───────┼───────────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ 1     │ #389, existing branch/stacked PR                  │ Finish and independently verify current Shepherd hardening only. Preserve all unfinished work. No state-directory  │
 │       │                                                   │ migration.                                                                                                         │
 ├───────┼───────────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ 2     │ New sub-issue under #372: repository state-layout │ Move adapter assets, archive tracked history, update live contracts/locators, add final ignore rules and           │
 │       │ migration                                         │ compatibility bridge.                                                                                              │
 ├───────┼───────────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ 3     │ New sub-issue under #372: host-local isolated     │ Make Shepherd consume the stable template, create one fresh GSD project per issue, remove .gsd/.planning           │
 │       │ bootstrap and adoption                            │ checkpoint staging, and prove restart/collision/model behavior.                                                    │
 ├───────┼───────────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ 4     │ #379 if still open and matching scope; otherwise  │ Exact-head incident replay and merge-disabled host-local sandbox canary. No Podman requirement.                    │
 │       │ a new canary issue                                │                                                                                                                    │
 ├───────┼───────────────────────────────────────────────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
 │ 5     │ Parent #372 / PR #390                             │ Final integrated verification and human merge gate.                                                                │
 └───────┴───────────────────────────────────────────────────┴────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘

 All sub-PRs target feat/372-gsd-pi-go-shepherd, use Refs, and remain one-primary-issue PRs. Parent PR #390 into main remains human-gated.

 ────────────────────────────────────────────────────────────────────────────────

 5. MIGRATION

 ### Exact old-to-new mapping

 ┌────────────────────────────────────────────────────────────┬─────────────────────────────────────────────────────────────────────────┐
 │ Old path                                                   │ New tracked path                                                        │
 ├────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┤
 │ .gsd/commands.json                                         │ .agents/gsd/prompt-adapter/commands.json                                │
 ├────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┤
 │ .gsd/local-commands.json                                   │ .agents/gsd/prompt-adapter/local-commands.json                          │
 ├────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┤
 │ .gsd/upstream.lock.json                                    │ .agents/gsd/prompt-adapter/upstream.lock.json                           │
 ├────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┤
 │ .gsd/prompts/**                                            │ .agents/gsd/prompt-adapter/prompts/**                                   │
 ├────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┤
 │ .gsd/official-docs/**                                      │ .agents/gsd/prompt-adapter/official-docs/**                             │
 ├────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┤
 │ .gsd/PREFERENCES.md                                        │ .agents/gsd/shepherd-project-template/PREFERENCES.md                    │
 ├────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┤
 │ .gsd/agents/**                                             │ .agents/gsd/shepherd-project-template/agents/**                         │
 ├────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┤
 │ .gsd/{PROJECT,REQUIREMENTS,KNOWLEDGE}.md                   │ docs/agentic-delivery/archive/gsd-pi/m001-autonomous-shepherd/          │
 ├────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┤
 │ .gsd/M001-META.json                                        │ Same GSD Pi archive root                                                │
 ├────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┤
 │ Tracked .gsd/phases/**, .gsd/milestones/**                 │ Same GSD Pi archive root, preserving relative paths                     │
 ├────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┤
 │ .planning/phases/issue-49-destructive-confirmation-gate/** │ docs/agentic-delivery/issues/49/destructive-confirmation-gate/**        │
 ├────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┤
 │ .planning/phases/issue-121-github-full-certificate/**      │ docs/agentic-delivery/issues/121/github-full-certificate/**             │
 ├────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┤
 │ .planning/phases/issue-372-gsd-pi-go-shepherd/**           │ docs/agentic-delivery/issues/372/gsd-pi-go-shepherd/**                  │
 ├────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┤
 │ .planning/phases/issue-389-shepherd-hardening/**           │ docs/agentic-delivery/issues/389/shepherd-hardening/**                  │
 ├────────────────────────────────────────────────────────────┼─────────────────────────────────────────────────────────────────────────┤
 │ Remaining tracked .planning/**                             │ docs/agentic-delivery/archive/planning-v1/**, preserving relative paths │
 └────────────────────────────────────────────────────────────┴─────────────────────────────────────────────────────────────────────────┘

 Never commit or archive .gsd/gsd.db*, credentials, raw prompts, raw tool output, or sessions. If any database is already tracked, stop for security review and remove it from the
 index without deleting the local file.

 ### Ordered checkpoints

 #### M0 — Reconcile and inventory

 1. Finish #389 or record its exact blocker.
 2. Confirm no live Shepherd/GSD process owns the current state.
 3. Record git ls-files manifests for .gsd/** and .planning/**.
 4. Hash tracked adapter resources and historical files.
 5. Do not reset the current worktree.

 #### M1 — Add the new adapter root with a bridge

 1. Copy adapter assets to .agents/gsd/prompt-adapter/.
 2. Change scripts/gsd to:
     - prefer the new root;
     - allow the old tracked root as a temporary read-only fallback;
     - make sync-upstream write only to the new root;
     - emit a bounded deprecation warning when fallback is used.
 3. Keep all old files in place at this checkpoint.

 Working command paths: old and new.

 #### M2 — Move consumers to the new root

 Update:

 - scripts/gsd;
 - .pi/extensions/gsd/index.ts;
 - .pi/skills/gsd-core/SKILL.md;
 - .pi/prompts/gsd.md;
 - .agents/agentic-delivery/references/gsd-pi-adapter.md;
 - Shepherd adapter tests;
 - generated/help source references.

 The Pi extension must no longer depend on ignored runtime files. It may read the stable registry directly or consume scripts/gsd list --json.

 Working command paths: new primary, old fallback.

 #### M3 — Establish durable evidence paths

 1. Move issue artifacts to docs/agentic-delivery/issues/**.
 2. Move all other tracked planning history to the archive.
 3. Add PATH-MAP.md and a tracked content manifest.
 4. Update live contracts, schemas, workflows, CI path filters, and examples.
 5. Historical prose may retain old links only when the migration map explicitly resolves them.

 Future plans, TDD ledgers, verification checklists, and run-state evidence are written directly under the issue’s tracked docs directory. .planning/ is only a local projection.

 #### M4 — Make runtime directories fully ignored

 Final root rules:

 ```gitignore
   # Per-worktree GSD Pi workflow state
   /.gsd/
   /.gsd-bootstrap-*/

   # Per-worktree prompt-adapter and orchestration state
   /.planning/
   /.planning-bootstrap-*/
 ```

 Retain .gsd-worktrees/ and unrelated ignores.

 Already tracked files must be migrated with git mv. For unexpected tracked runtime files, use index-only removal; never use rm -rf, git reset, or broad cleaning.

 #### M5 — Remove hidden Git staging behavior

 In Shepherd’s Git package:

 - remove the special allowance that stages mutable tracked .gsd projections;
 - ensure .gsd/** and .planning/** cannot enter a checkpoint;
 - continue treating tracked source/policy changes outside those ignored roots normally;
 - export selected sanitized evidence only to docs/agentic-delivery/issues/**.

 #### M6 — Shepherd bootstrap and adoption

 Bootstrap:

 1. Acquire a controller lease keyed by repository identity, issue, project root, and generation.
 2. Validate real paths and ensure .gsd/.planning are absent or bound to the exact same identity.
 3. Transactionally create .gsd with ISSUE.json, template policy, and agent specs.
 4. Initialize GSD Pi through documented commands; Shepherd never creates or edits gsd.db directly.
 5. Create local .planning compatibility configuration with commit_docs=false.
 6. Persist template/adapter hashes in Shepherd’s delivery binding.
 7. Dispatch only after native query and model admission pass.

 Adoption/restart:

 - Exact identity and hashes: idempotent adoption.
 - Existing .gsd with missing protected controller binding: fail closed unless explicit --adopt-existing is authorized and independently proven.
 - Different issue, branch, base, root, head, context, version, or template hash: reject.
 - Reconcile lease, child lifecycle, outbox, attempts, GSD query, and exact Git head before dispatch.

 Collision:

 - One project root cannot be bound to two issues.
 - One issue cannot have two active controller generations.
 - No database, GSD home, session directory, or .planning tree is copied between issues.

 Upgrade:

 - Pin GSD Pi and prompt-adapter locks separately.
 - A GSD Pi upgrade is a separate qualification issue.
 - Freeze the old project; migrate only through documented query/export APIs.
 - Never merge or rewrite private databases.

 #### M7 — Remove the fallback and canary

 After clean-clone/fresh-worktree validation, remove the old .gsd adapter fallback. Run the host-local merge-disabled canary. Legacy Podman removal is a later human-approved
 cleanup; it is not part of #389.

 ### Rollback

 - Before M4: revert the latest migration commit; old resources still work.
 - After M4: revert the migration commits; local ignored state remains untouched.
 - Runtime bootstrap failure: retain the failed directory under protected same-issue quarantine; do not adopt or copy it to another issue.
 - Database rollback is allowed only for the same issue, exact identity, exact GSD version, and exact controller generation.
 - Never roll back by resetting or deleting the current #389 worktree.

 ────────────────────────────────────────────────────────────────────────────────

 6. TDD

 ┌──────────────────────┬──────────────────────────────────────────────────────────┬──────────────────────────────────────────────────────┬───────────────────────────────────────┐
 │ Checkpoint           │ RED                                                      │ GREEN                                                │ Refactor                              │
 ├──────────────────────┼──────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────┼───────────────────────────────────────┤
 │ #389 reconciliation  │ Current diff/tests reveal actual failing state           │ Existing hardening tests and module gates pass       │ Remove duplicate status claims;       │
 │                      │                                                          │                                                      │ preserve evidence                     │
 ├──────────────────────┼──────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────┼───────────────────────────────────────┤
 │ Adapter relocation   │ A clean checkout with no .gsd cannot run doctor, list,   │ All commands resolve .agents/gsd/prompt-adapter      │ Centralize path resolution; keep      │
 │                      │ sources, prompt, or Pi registration                      │                                                      │ fallback read-only                    │
 ├──────────────────────┼──────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────┼───────────────────────────────────────┤
 │ Resource             │ Manifest detects missing registry, prompt, lock, doc, or │ New tree matches tracked baseline hashes/counts      │ Remove duplicate resource readers     │
 │ preservation         │ command                                                  │                                                      │                                       │
 ├──────────────────────┼──────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────┼───────────────────────────────────────┤
 │ History migration    │ Active contracts still require tracked .planning paths   │ Contracts point to tracked issue docs; archive map   │ Classify live versus historical       │
 │                      │                                                          │ resolves history                                     │ references                            │
 ├──────────────────────┼──────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────┼───────────────────────────────────────┤
 │ Ignore migration     │ git ls-files still reports .gsd/** or .planning/**;      │ Both roots are ignored and absent from the index     │ Delete obsolete narrow ignore rules   │
 │                      │ runtime files can be staged                              │                                                      │ only                                  │
 ├──────────────────────┼──────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────┼───────────────────────────────────────┤
 │ Checkpoint isolation │ A changed .gsd or .planning file enters a Shepherd       │ Checkpoint contains only allowlisted source and      │ Remove isMutableGSDProjection         │
 │                      │ commit                                                   │ explicit tracked evidence exports                    │ behavior                              │
 ├──────────────────────┼──────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────┼───────────────────────────────────────┤
 │ Fresh project        │ Two issues can share a root/home/state/database          │ Distinct issue roots and protected delivery          │ Consolidate immutable identity        │
 │ isolation            │                                                          │ namespaces                                           │ validation                            │
 ├──────────────────────┼──────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────┼───────────────────────────────────────┤
 │ Adoption/restart     │ Changed branch/base/root/head/hash or orphaned child is  │ Exact restart resumes; mismatches fail closed        │ Keep reconciliation deterministic and │
 │                      │ silently adopted                                         │                                                      │ side-effect free                      │
 ├──────────────────────┼──────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────┼───────────────────────────────────────┤
 │ Rollback             │ Failed bootstrap leaves a half-bound live project        │ Failure quarantines state and releases/fences        │ Add typed bootstrap failure classes   │
 │                      │                                                          │ ownership                                            │                                       │
 ├──────────────────────┼──────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────┼───────────────────────────────────────┤
 │ Model routing        │ Planning/execution accept configured intent without      │ Planning/validation observe GPT-5.6 Sol/high;        │ Share exact observed-model validation │
 │                      │ observation                                              │ execution observes GPT-5.5/high                      │                                       │
 ├──────────────────────┼──────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────┼───────────────────────────────────────┤
 │ Controller ownership │ Restart dispatches before lease/outbox/attempt           │ Same generation resumes exactly once; stale          │ Separate pure policy from ports       │
 │                      │ reconciliation                                           │ controller cannot act                                │                                       │
 ├──────────────────────┼──────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────┼───────────────────────────────────────┤
 │ No state merge       │ Ignored files can be force-added or promoted wholesale   │ Only sanitized allowlisted exports reach tracked     │ Add source generation/head/hash       │
 │                      │                                                          │ issue docs                                           │ metadata                              │
 ├──────────────────────┼──────────────────────────────────────────────────────────┼──────────────────────────────────────────────────────┼───────────────────────────────────────┤
 │ Canary               │ Restart or stale head duplicates a unit/effect           │ Host-local canary reaches final human gate exactly   │ Preserve bounded audit evidence       │
 │                      │                                                          │ once                                                 │                                       │
 └──────────────────────┴──────────────────────────────────────────────────────────┴──────────────────────────────────────────────────────┴───────────────────────────────────────┘

 Tests should use temporary repositories and fake GSD runners; no Podman, network, or credentials in default tests.

 ────────────────────────────────────────────────────────────────────────────────

 7. VERIFICATION

 ### Adapter and migration

 ```bash
   node --check scripts/gsd
   scripts/gsd doctor
   scripts/gsd list --json
   scripts/gsd sources programming-loop
   scripts/gsd prompt programming-loop init --phase state-migration --dry-run
   scripts/gsd verify-pi

   git ls-files '.gsd/**' '.planning/**'
   git check-ignore -v .gsd/gsd.db .planning/config.json
   git diff --check
 ```

 Acceptance: git ls-files prints nothing for the two runtime roots.

 Also run targeted checks that:

 - compare tracked-manifest.json to the moved assets;
 - reject live contract references to old adapter paths;
 - allow explicitly classified historical references;
 - prove sync-upstream writes only to the stable adapter root.

 ### Shepherd

 ```bash
   cd agent-runtime/shepherd
   gofmt -w cmd internal
   go test ./...
   go test -race ./...
   go vet ./...
   go build ./cmd/shepherd
   make verify
   cd ../..
   scripts/tests/shepherd-module-boundary.sh
 ```

 ### Repository

 ```bash
   gofmt -w cmd internal
   go vet ./...
   go test ./...
   go build ./cmd/pm
   make verify
 ```

 ### Exact acceptance evidence

 - exact candidate head SHA;
 - old/new resource manifest and path map;
 - no tracked .gsd/** or .planning/**;
 - fresh two-issue isolation test;
 - same-issue restart test;
 - checkpoint diff proving no ignored-state merge;
 - observed model/thinking records;
 - host-local canary record ending at final_human_gate;
 - no Podman command required for default verification.

 ────────────────────────────────────────────────────────────────────────────────

 8. RISKS AND HUMAN GATES

 ┌───────────────────────────────────┬──────────────────────────────────────────────────────────────────────────────────┐
 │ Risk/gate                         │ Required handling                                                                │
 ├───────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────┤
 │ Current unfinished #389 changes   │ Never reset, clean, replace wholesale, or delete; reconcile through tests first  │
 ├───────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────┤
 │ Broad history move                │ Separate issue and PR; inspect rename manifest before integration                │
 ├───────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────┤
 │ Tracked gsd.db or sensitive state │ Stop immediately; do not print or archive contents; human security review        │
 ├───────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────┤
 │ Active GSD/Shepherd process       │ Stop migration until ownership is released or explicitly checkpointed            │
 ├───────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────┤
 │ GSD Pi private DB schema          │ Never edit, merge, or copy across issues                                         │
 ├───────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────┤
 │ GitHub sandbox/canary             │ Requires explicit auth and external-effect authorization                         │
 ├───────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────┤
 │ Dependency addition               │ No dependencies authorized; stop if one becomes necessary                        │
 ├───────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────┤
 │ Podman removal                    │ Deferred until host canary passes; deletion requires separately reviewed cutover │
 ├───────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────┤
 │ GSD Pi upgrade                    │ Separate qualification issue and exact-version review                            │
 ├───────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────┤
 │ Broad generated rewrite           │ Human review of manifest and diff required                                       │
 ├───────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────┤
 │ Parent PR merge to main           │ Always human-only                                                                │
 ├───────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────┤
 │ Automated review                  │ Exact integrated head must receive Claude, allowed fallback, or human coverage   │
 └───────────────────────────────────┴──────────────────────────────────────────────────────────────────────────────────┘

 ────────────────────────────────────────────────────────────────────────────────

 9. IMPLEMENTATION PROMPTS

 ### Prompt A — GPT-5.5/high: finish issue #389

 ```text
   Model: openai-codex/gpt-5.5
   Thinking: high

   Objective:
   Finish issue #389 only on its existing isolated branch. Reconcile the real diff and test state first.
   Preserve all intentional unfinished RED/GREEN work. Complete Shepherd runtime-contract admission,
   immutable issue identity, durable attempts, child reconciliation, exact completion proof, canonical
   single-unit supervision, model routing, and final-human-gate behavior.

   Output:
   A <=40-line worker handoff listing changed files, RED/GREEN/refactor evidence, exact commands/results,
   remaining blockers, candidate head, and required skills used.

   Tool guidance:
   Read AGENTS.md, the #389 issue artifacts, parent #372 contract, required-skills routing, GSD adapter
   reference, issue contract, and current Shepherd diff. Load gsd-core, polymetrics-issue-delivery,
   golang-how-to, golang-design-patterns, golang-structs-interfaces, golang-context,
   golang-concurrency, golang-error-handling, golang-security, golang-safety, and golang-testing.
   Use the mandatory programming-loop and run all nested Shepherd gates.

   Boundaries:
   Do not perform the .gsd/.planning migration. Do not reset, clean, delete, or overwrite current work.
   Do not add dependencies, restore Podman behavior, expose secrets, publish GitHub effects directly,
   merge any PR, or push to main. Commit coherent local green slices; route publication through the
   authorized Shepherd effect path.
 ```

 ### Prompt B — GPT-5.5/high: repository state-layout migration

 ```text
   Model: openai-codex/gpt-5.5
   Thinking: high

   Objective:
   Implement the new #372 sub-issue that relocates repository-owned GSD adapter resources, archives all
   tracked .planning and tracked GSD projection history without data loss, moves live issue evidence to
   docs/agentic-delivery/issues, updates all active locators/contracts, and makes root .gsd/ and
   .planning/ fully ignored.

   Output:
   A <=40-line handoff with the complete path map, tracked manifest result, compatibility-bridge state,
   RED/GREEN/refactor evidence, exact verification, candidate head, and human gates.

   Tool guidance:
   Use gsd-programming-loop before edits. Load gsd-core, polymetrics-issue-delivery, golang-how-to,
   golang-testing, golang-error-handling, golang-security, and golang-safety. Start with tests proving a
   clean checkout can run scripts/gsd without .gsd. Use git mv for tracked history. Update scripts/gsd,
   .pi GSD resources, live .agents contracts, CI filters, docs, .gitignore, and the bounded Shepherd
   adapter test.

   Boundaries:
   Do not modify Shepherd supervisor behavior beyond adapter path tests. Do not delete local runtime
   state, stage gsd.db, use git reset/clean, add dependencies, invoke Podman, mutate GitHub directly, or
   merge to main. Keep the old adapter path as a read-only bridge until new-path verification is green.
 ```

 ### Prompt C — GPT-5.5/high: host-local isolated bootstrap

 ```text
   Model: openai-codex/gpt-5.5
   Thinking: high

   Objective:
   Implement the subsequent #372 sub-issue that makes Shepherd bootstrap/adopt exactly one fresh local
   GSD Pi 1.11.0 project per issue worktree, consume the stable tracked project template, protect
   controller state outside the worktree, prevent .gsd/.planning from entering checkpoints, and prove
   restart-safe sole-controller ownership.

   Output:
   A <=40-line handoff listing identity invariants, RED/GREEN/refactor evidence, exact model observations,
   verification commands/results, candidate head, rollback behavior, and blockers.

   Tool guidance:
   Use the programming loop and load golang-how-to, golang-design-patterns,
   golang-structs-interfaces, golang-context, golang-concurrency, golang-error-handling,
   golang-security, golang-safety, and golang-testing. Add temporary-repository tests for two-issue
   isolation, collision rejection, exact adoption, stale generation, orphan reconciliation, bootstrap
   rollback, no ignored-state staging, and GPT-5.6/GPT-5.5 routing. Qualify whether GSD_STATE_DIR changes
   the location of gsd.db; keep the database issue-local.

   Boundaries:
   Host-local GSD only. Do not manipulate private GSD DB contents, copy state across issues, add
   dependencies, restore Podman as a requirement, delete legacy container assets, publish GitHub
   effects directly, or cross the final human merge gate.
 ```

 ### Independent validation — GPT-5.6 Sol/high

 ```text
   Model: openai-codex/gpt-5.6-sol
   Thinking: high

   Independently validate the exact integrated candidate head for #389 plus the two state-migration
   sub-issues. Read-only: do not edit, commit, push, comment, merge, or expose credentials.

   Fail unless all are proven from current commands and files:

   1. scripts/gsd and the Pi extension work in a fresh checkout with no .gsd or .planning.
   2. Repository adapter resources resolve only from .agents/gsd/prompt-adapter after bridge removal.
   3. No tracked .gsd/** or .planning/** remains, and no historical tracked evidence was discarded.
   4. Required issue plan/TDD/verification evidence is durable under docs/agentic-delivery/issues.
   5. Two issues cannot reuse a project root, database, GSD home, controller state, lease, or session.
   6. Same-issue exact restart resumes without duplicate unit, checkpoint, answer, or external effect.
   7. Ignored .gsd/.planning files cannot enter Shepherd checkpoints or PR diffs.
   8. GSD Pi owns workflow state; Shepherd owns controller authority and never edits the private DB.
   9. Planning/validation observe openai-codex/gpt-5.6-sol/high and execution observes
      openai-codex/gpt-5.5/high.
   10. Default verification and the canary require no Podman dependency.
   11. All focused, race, vet, build, nested make verify, root make verify, and module-boundary gates pass.
   12. The terminal successful state is final_human_gate; no parent-PR merge capability exists.

   Return PROCEED only with exact head, command evidence, resource-manifest result, controller/workflow
   ownership assessment, and no unresolved actionable finding. Otherwise return RETRY with bounded
   file/line findings and concrete fixes.
