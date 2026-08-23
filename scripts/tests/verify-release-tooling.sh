#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)

ruby - "$root" <<'RUBY'
require "yaml"

root = ARGV.fetch(0)
workflow_path = File.join(root, ".github", "workflows", "verify.yml")
workflow = YAML.safe_load(File.read(workflow_path), aliases: false)
steps = workflow.fetch("jobs").fetch("verify").fetch("steps")
verify_index = steps.index { |step| step["name"] == "Verify" && step["run"] == "make verify" }

abort "verify release tooling check failed: Verify job must run make verify" unless verify_index

required_install = "go install github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.43.0"
required_path = 'echo "$(go env GOPATH)/bin" >> "$GITHUB_PATH"'
required_version = '"$(go env GOPATH)/bin/nfpm" --version'
nfpm_index = steps.index do |step|
  run = step["run"]
  run.is_a?(String) && run.include?(required_install)
end

unless nfpm_index && nfpm_index < verify_index
  abort "verify release tooling check failed: Verify job must provision pinned nfpm before make verify"
end

run = steps.fetch(nfpm_index).fetch("run")
missing = [required_install, required_path, required_version].reject { |requirement| run.include?(requirement) }
unless missing.empty?
  abort "verify release tooling check failed: nfpm setup is incomplete: #{missing.join(', ')}"
end
RUBY

printf '%s\n' 'verify release tooling: nfpm is provisioned in the owning Verify job'
