# PRD: Issue-first agentic delivery foundation

## Problem

Connector CLI migration work needs a disciplined implementation foundation so every PR is
issue-backed, test-first, small enough to review, and transferable from GitHub to other connectors.

## Goals

- Require every PR to reference a GitHub issue in the PR body.
- Keep each PR incremental, testable, and non-disruptive.
- Use strict red/green/refactor evidence for behavior changes.
- Keep production connector behavior unchanged unless a specific implementation issue authorizes it.
- Build foundation pieces that generalize across connectors instead of hard-coding one connector.
- Keep reusable agent contracts isolated under `.agents/` and agent role specs under `.codex/agents/`.

## Non-goals

- Do not implement the full connector command executor in one PR.
- Do not promote GitHub CLI surface metadata in this PR.
- Do not add raw unrestricted HTTP, shell, or SQL write surfaces.
- Do not request, print, or store secrets.
- Do not create or update GitHub Projects in this PR.

## Acceptance

- A PR issue guard exists, is tested, and runs in GitHub Actions.
- The PR template makes linked issue syntax explicit.
- An agent task issue form captures objective, task type, scope, TDD, verification, and hard stops.
- Generic issue-to-PR contracts and skill mappings live under `.agents/connector-cli-parity/`.
- Repo-local YAML agent role definitions live under `.codex/agents/connector-cli-parity/`.
- TDD and verification evidence is recorded in this phase directory.
