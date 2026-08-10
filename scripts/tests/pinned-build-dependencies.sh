#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

ruby - "$ROOT_DIR" <<'RUBY'
require "pathname"
require "psych"

root = Pathname.new(ARGV.fetch(0))
workflow_glob = root.join(".github", "workflows", "*.{yml,yaml}").to_s
action_sha = /\A[^@\s]+@[0-9a-f]{40}\z/
version_comment = /\s#\s*v[0-9]/
image_digest = /@sha256:[0-9a-f]{64}\z/
errors = []

def each_mapping_entry(node, &block)
  case node
  when Psych::Nodes::Mapping
    node.children.each_slice(2) do |key, value|
      yield key, value
      each_mapping_entry(value, &block)
    end
  when Psych::Nodes::Sequence, Psych::Nodes::Document, Psych::Nodes::Stream
    node.children.each { |child| each_mapping_entry(child, &block) }
  end
end

Dir[workflow_glob].sort.each do |path|
  lines = File.readlines(path, chomp: true)
  workflow = Psych.parse_file(path)
  relative_path = Pathname.new(path).relative_path_from(root)

  each_mapping_entry(workflow) do |key, value|
    next unless key.is_a?(Psych::Nodes::Scalar) && key.value == "uses"

    line_number = value.start_line + 1
    unless value.is_a?(Psych::Nodes::Scalar)
      errors << "#{relative_path}:#{line_number}: action reference must be a scalar"
      next
    end

    ref = value.value
    next if ref.start_with?("./")

    source_line = lines.fetch(value.start_line)
    if ref.start_with?("docker://")
      unless image_digest.match?(ref.delete_prefix("docker://"))
        errors << "#{relative_path}:#{line_number}: Docker action must use a sha256 digest: #{ref}"
      end
    elsif !action_sha.match?(ref)
      errors << "#{relative_path}:#{line_number}: action must use a full 40-character commit SHA: #{ref}"
    end

    unless version_comment.match?(source_line)
      errors << "#{relative_path}:#{line_number}: action must retain its version in a trailing '# v…' comment: #{ref}"
    end
  end
end

unless errors.empty?
  warn "pinned build dependencies check failed:"
  errors.each { |error| warn "  #{error}" }
  exit 1
end
RUBY

python3 - "$ROOT_DIR" <<'PY'
from pathlib import Path
import re
import sys

root = Path(sys.argv[1])
workflow_dir = root / ".github" / "workflows"
workflows = sorted([*workflow_dir.glob("*.yml"), *workflow_dir.glob("*.yaml")])

image_digest = re.compile(r"@sha256:[0-9a-f]{64}$")
image_line = re.compile(r"^\s*image:\s*(.*?)\s*$")
from_line = re.compile(r"^\s*FROM\s+(?:--platform=[^\s]+\s+)?([^\s]+)(?:\s+AS\s+[^\s]+)?\s*(?:#.*)?$", re.IGNORECASE)

errors = []

for workflow in workflows:
    for line_number, line in enumerate(workflow.read_text().splitlines(), start=1):
        match = image_line.match(line)
        if match:
            image = match.group(1).split("#", 1)[0].strip().strip("'\"")
            if not image or image.startswith("${{"):
                continue
            if not image_digest.search(image):
                errors.append(f"{workflow.relative_to(root)}:{line_number}: literal workflow image must use a sha256 digest: {image}")

for dockerfile in sorted(root.rglob("Dockerfile*")):
    if not dockerfile.is_file():
        continue
    for line_number, line in enumerate(dockerfile.read_text().splitlines(), start=1):
        match = from_line.match(line)
        if not match:
            continue
        image = match.group(1)
        if not image_digest.search(image):
            errors.append(f"{dockerfile.relative_to(root)}:{line_number}: Dockerfile base image must use a sha256 digest: {image}")

if errors:
    print("pinned build dependencies check failed:", file=sys.stderr)
    print("\n".join(f"  {error}" for error in errors), file=sys.stderr)
    raise SystemExit(1)
PY

printf 'pinned build dependencies: all workflow actions and literal build images are immutable\n'
