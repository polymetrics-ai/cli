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

def collect_anchors(node, anchors = {}, duplicate_anchors = [])
  if !node.is_a?(Psych::Nodes::Alias) && node.respond_to?(:anchor) && node.anchor
    if anchors.key?(node.anchor)
      duplicate_anchors << node
    else
      anchors[node.anchor] = node
    end
  end

  case node
  when Psych::Nodes::Mapping, Psych::Nodes::Sequence, Psych::Nodes::Document, Psych::Nodes::Stream
    node.children.each { |child| collect_anchors(child, anchors, duplicate_anchors) }
  end

  anchors
end

def resolve_alias(node, anchors)
  seen = {}
  while node.is_a?(Psych::Nodes::Alias)
    return nil if seen[node.anchor]

    seen[node.anchor] = true
    node = anchors[node.anchor]
    return nil unless node
  end

  node
end

Dir[workflow_glob].sort.each do |path|
  lines = File.readlines(path, chomp: true)
  workflow = Psych.parse_file(path)
  relative_path = Pathname.new(path).relative_path_from(root)
  duplicate_anchors = []
  anchors = collect_anchors(workflow, {}, duplicate_anchors)
  duplicate_anchors.each do |anchor|
    errors << "#{relative_path}:#{anchor.start_line + 1}: duplicate YAML anchor is not supported: #{anchor.anchor}"
  end

  each_mapping_entry(workflow) do |key, value|
    if key.is_a?(Psych::Nodes::Alias)
      errors << "#{relative_path}:#{key.start_line + 1}: workflow mapping keys must not use YAML aliases"
      next
    end
    next unless key.is_a?(Psych::Nodes::Scalar)

    case key.value
    when "uses"
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
    when "image"
      if value.is_a?(Psych::Nodes::Alias)
        image_node = resolve_alias(value, anchors)
        unless image_node.is_a?(Psych::Nodes::Scalar)
          errors << "#{relative_path}:#{value.start_line + 1}: workflow image alias must resolve to a scalar"
          next
        end
      else
        next unless value.is_a?(Psych::Nodes::Scalar)

        image_node = value
      end

      image = image_node.value.strip
      next if image.empty? || image.start_with?("${{")
      unless image_digest.match?(image)
        errors << "#{relative_path}:#{value.start_line + 1}: literal workflow image must use a sha256 digest: #{image}"
      end
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

image_digest = re.compile(r"@sha256:[0-9a-f]{64}$")
from_line = re.compile(r"^\s*FROM\s+(?:--platform=[^\s]+\s+)?([^\s]+)(?:\s+AS\s+[^\s]+)?\s*(?:#.*)?$", re.IGNORECASE)

errors = []

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
