#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)
selector="$root/.github/workflows/runner-selection.yml"

test -f "$selector"

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
require 'echo "runner=polymetrics-website" >> "$GITHUB_OUTPUT"'
require 'echo "runner=ubuntu-latest" >> "$GITHUB_OUTPUT"'

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
test "$selected_linux_jobs" -eq 19 || {
  printf 'expected 19 Linux jobs to consume the shared selector, found %s\n' "$selected_linux_jobs" >&2
  exit 1
}

windows="$root/.github/workflows/windows-package-check.yml"
require "runs-on: windows-latest" "$windows"
if grep -Fq "runner-selection.yml" "$windows"; then
  printf 'Windows workflow must not consume the Linux runner selector\n' >&2
  exit 1
fi

website="$root/.github/workflows/website.yml"
require "runs-on: [self-hosted, linux, tailscale, polymetrics-website]" "$website"

printf 'CI runner routing contract passed\n'
