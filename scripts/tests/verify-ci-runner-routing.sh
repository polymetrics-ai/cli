#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)
selector="$root/.github/workflows/runner-selection.yml"
policy="$root/docs/security/self-hosted-ci-runner-policy.md"

test -f "$selector"
test -f "$policy"

require() {
  local pattern=$1
  local file=${2:-$selector}
  grep -Fq -- "$pattern" "$file" || {
    printf 'missing required routing contract in %s: %s\n' "$file" "$pattern" >&2
    exit 1
  }
}

require "workflow_call:"
require "runs-on: ubuntu-latest"
require "github.event_name == 'pull_request'"
require "github.event.pull_request.head.repo.full_name == github.repository"
require "github.event.pull_request.user.login == 'karthik-sivadas'"
require "github.event.pull_request.user.login == 'alfred-polymetrics-ai'"
require 'echo "runner=polymetrics-cli" >> "$GITHUB_OUTPUT"'
require 'echo "runner=ubuntu-latest" >> "$GITHUB_OUTPUT"'
require 'CI routing is unsafe until' "$policy"
require 'Require approval for all external contributors' "$policy"
require 'Repository access' "$policy"
require 'Selected repositories' "$policy"
require 'polymetrics-ai/cli' "$policy"
if grep -Fq 'separate `polymetrics-website` runner' "$policy"; then
  printf 'routing policy must not retain a non-PR self-hosted website exception\n' >&2
  exit 1
fi

claude="$root/.github/workflows/claude-review.yml"
claude_selector=$(awk '
  /^  select-runner:/ { in_selector=1 }
  in_selector && /^  auto-review:/ { exit }
  in_selector { print }
' "$claude")
require_claude_selector() {
  local pattern=$1
  grep -Fq -- "$pattern" <<<"$claude_selector" || {
    printf 'missing Claude selector eligibility condition: %s\n' "$pattern" >&2
    exit 1
  }
}

require_claude_selector 'if: |'
require_claude_selector "github.event_name == 'pull_request'"
require_claude_selector 'github.event.pull_request.draft == false'
require_claude_selector '["OWNER","MEMBER","COLLABORATOR","CONTRIBUTOR"]'
require_claude_selector "github.event_name == 'issue_comment'"
require_claude_selector 'github.event.issue.pull_request != null'
require_claude_selector "contains(github.event.comment.body, '@claude')"
require_claude_selector "github.event_name == 'pull_request_review_comment'"
require_claude_selector '["OWNER","MEMBER","COLLABORATOR"]'

linux_workflows=(
  scorecard.yml website.yml gsd-workflow.yml connector-boundary.yml conventions.yml
  verify.yml security.yml website-data.yml release.yml pr-issue-guard.yml claude-review.yml
)
for workflow in "${linux_workflows[@]}"; do
  require "uses: ./.github/workflows/runner-selection.yml" "$root/.github/workflows/$workflow"
  require 'runs-on: ${{ needs.select-runner.outputs.runner }}' "$root/.github/workflows/$workflow"
done

if rg -n --glob '*.yml' 'runs-on: ubuntu-latest' "$root/.github/workflows" | grep -Fv 'runner-selection.yml:'; then
  printf 'all Linux GitHub-hosted jobs must consume the shared selector\n' >&2
  exit 1
fi

selected_linux_jobs=$(rg --glob '*.yml' -c 'runs-on: \$\{\{ needs\.select-runner\.outputs\.runner \}\}' "$root/.github/workflows" | awk -F: '{ total += $NF } END { print total + 0 }')
test "$selected_linux_jobs" -eq 20 || {
  printf 'expected 20 Linux jobs to consume the shared selector, found %s\n' "$selected_linux_jobs" >&2
  exit 1
}

windows="$root/.github/workflows/windows-package-check.yml"
require "runs-on: windows-latest" "$windows"
if grep -Fq "runner-selection.yml" "$windows"; then
  printf 'Windows workflow must not consume the Linux runner selector\n' >&2
  exit 1
fi

website="$root/.github/workflows/website.yml"
website_deploy=$(awk '
  /^  deploy:/ { in_deploy=1 }
  in_deploy { print }
' "$website")
require_website_deploy() {
  local pattern=$1
  grep -Fq -- "$pattern" <<<"$website_deploy" || {
    printf 'missing website deployment routing contract: %s\n' "$pattern" >&2
    exit 1
  }
}

require_website_deploy 'runs-on: ${{ needs.select-runner.outputs.runner }}'
require_website_deploy "if: github.ref_name == 'main' && (github.event_name == 'push' || github.event_name == 'workflow_dispatch') && vars.WEBSITE_DEPLOY_ENABLED == 'true'"
if ! grep -Eq '^[[:space:]]+needs:.*select-runner' <<<"$website_deploy"; then
  printf 'website deployment must depend on the shared runner selector\n' >&2
  exit 1
fi
if grep -Fq 'self-hosted' <<<"$website_deploy"; then
  printf 'website deployment must not route non-PR events to a self-hosted runner\n' >&2
  exit 1
fi

printf 'CI runner routing contract passed\n'
