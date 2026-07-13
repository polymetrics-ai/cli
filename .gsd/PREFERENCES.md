---
version: 1
mode: team
always_use_skills:
  - polymetrics-issue-delivery
models:
  validation:
    model: openai-codex/gpt-5.6-sol
    thinking: high
dynamic_routing:
  enabled: false
git:
  auto_push: false
  push_branches: false
  snapshots: false
  pre_merge_check: true
  main_branch: main
  isolation: worktree
  manage_gitignore: false
  auto_pr: false
  merge_strategy: squash
reactive_execution:
  enabled: false
parallel:
  enabled: false
  max_workers: 4
  auto_merge: manual
context_management:
  observation_masking: true
  observation_mask_turns: 6
  compaction_threshold_percent: 0.60
  tool_result_max_chars: 800
verification_auto_fix: false
verification_max_retries: 0
uat_dispatch: true
pre_dispatch_hooks:
  - name: polymetrics-unit-policy
    before:
      - research-milestone
      - plan-milestone
      - research-slice
      - plan-slice
      - execute-task
      - complete-slice
      - replan-slice
      - reassess-roadmap
      - run-uat
    action: modify
    prepend: |
      POLYMETRICS POLICY: State Objective, Output format, Tool guidance, and Boundaries before acting.
      Read AGENTS.md and the polymetrics-issue-delivery skill. Stay inside the isolated issue
      worktree and declared write scope. Record RED before behavior changes. Do not use GitHub
      credentials or publish Git/GitHub effects; emit a typed external-effect intent instead.
    enabled: true
post_unit_hooks:
  - name: polymetrics-task-contract
    after:
      - execute-task
      - complete-slice
      - run-uat
    prompt: |
      Validate the current exact head against the Polymetrics delivery contract. Reject missing RED
      evidence, scope drift, false completion, direct external effects, stale evidence, or missing
      four-field dispatches. Start the artifact with verdict frontmatter and include candidate head,
      policy hash, independent checks, observed model, and observed thinking.
    max_cycles: 1
    model:
      model: openai-codex/gpt-5.6-sol
      thinking: high
    artifact: POLYMETRICS-VALIDATION.md
    criticality: blocking
    on_block:
      action: pause
    agent: polymetrics-contract-validator
    enabled: true
planning_subagent_registry:
  polymetrics-parent-planner:
    read_only_specialist: true
planning_subagents:
  plan-milestone:
    allowed:
      - polymetrics-parent-planner
  plan-slice:
    allowed:
      - polymetrics-parent-planner
---

# GSD Skill Preferences

Polymetrics pins validation to GPT-5.6 Sol with high reasoning and no fallback. GSD Pi may plan and
execute work, but Go Shepherd remains the only publisher of GitHub effects and merge to `main`
always requires a human.
