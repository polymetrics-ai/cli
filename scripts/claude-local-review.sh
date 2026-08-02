#!/usr/bin/env bash
# Local first-party Claude review for a PR — the `claude_local` route in
# .agents/agentic-delivery/workflows/automated-review-routing-loop.md.
#
# Use when the claude-review GitHub Action cannot provide coverage (missing
# ANTHROPIC_API_KEY secret, errored runs, quota) but review coverage is required
# before integrating a sub-PR or finalizing the parent PR. Runs on the operator's
# Claude subscription via the first-party `claude` CLI — never a third-party
# gateway — and posts ONE consolidated review comment via `gh`, plus a durable
# copy under .planning/reviews/.
#
# Usage:
#   scripts/claude-local-review.sh <pr-number>
#
# Config (env; defaults shown):
#   REVIEW_MODEL=claude-opus-4-8
#   CLAUDE_BIN=claude
#   REPO=            # owner/repo; defaults to the current repo per gh
set -euo pipefail

PR="${1:?usage: scripts/claude-local-review.sh <pr-number>}"
REVIEW_MODEL="${REVIEW_MODEL:-claude-opus-4-8}"
CLAUDE_BIN="${CLAUDE_BIN:-claude}"
REPO="${REPO:-$(gh repo view --json nameWithOwner -q .nameWithOwner)}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="$REPO_ROOT/.planning/reviews"
mkdir -p "$OUT_DIR"
OUT_FILE="$OUT_DIR/pr${PR}-claude-local.md"

HEAD_SHA="$(gh pr view "$PR" --repo "$REPO" --json headRefOid -q .headRefOid)"
BASE_REF="$(gh pr view "$PR" --repo "$REPO" --json baseRefName -q .baseRefName)"

"$CLAUDE_BIN" -p --model "$REVIEW_MODEL" \
  --allowedTools "Read,Grep,Glob,Bash" --permission-mode acceptEdits \
  "You are the repository's automated code reviewer, running as the sanctioned \`claude_local\`
review route (first-party claude CLI, model $REVIEW_MODEL) because the claude-review GitHub
Action is unavailable. Review PR #$PR on $REPO (head $HEAD_SHA, base $BASE_REF).

Follow .agents/agentic-delivery/workflows/claude-review-loop.md and
.agents/connector-migration/validation-gates.md. Adversarial standard: hunt for correctness bugs,
contract violations (write scopes, coverage-manifest reconciliation, secret hygiene, gate
integrity), and convention drift against the golden bundles — do not rubber-stamp.

Steps:
1. \`gh pr view $PR --repo $REPO\` and \`gh pr diff $PR --repo $REPO\` — review the FULL diff; read
   surrounding repo files as needed to judge context.
2. Write the review to $OUT_FILE with: a summary verdict (clean | findings), every finding as
   severity (blocking | important | minor) + file:line + a one-line fix direction, the reviewed
   range (base $BASE_REF...$HEAD_SHA), and the footer:
   'route: claude_local — first-party claude CLI, model $REVIEW_MODEL, reviewed head $HEAD_SHA'.
3. Post EXACTLY that file as one PR comment:
   \`gh pr review $PR --repo $REPO --comment --body-file $OUT_FILE\`
   (comment, never approve — approval stays human).
4. Print the verdict line as your final output.

Never print or store secret values. Do not push, merge, resolve threads, or edit code." \
  || { echo "claude_local review failed for PR #$PR" >&2; exit 1; }

echo "claude_local review posted for PR #$PR (head $HEAD_SHA); durable copy: $OUT_FILE"
