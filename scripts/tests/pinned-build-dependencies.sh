#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

python3 - "$ROOT_DIR" <<'PY'
from pathlib import Path
import re
import sys

root = Path(sys.argv[1])
workflow_dir = root / ".github" / "workflows"
workflows = sorted([*workflow_dir.glob("*.yml"), *workflow_dir.glob("*.yaml")])

action_sha = re.compile(r"^[^@\s]+@[0-9a-f]{40}$")
version_comment = re.compile(r"^\s*#\s*v[0-9]")
image_digest = re.compile(r"@sha256:[0-9a-f]{64}$")
uses_line = re.compile(r"^\s*uses:\s*([^\s#]+)(\s*#.*)?\s*$")
image_line = re.compile(r"^\s*image:\s*(.*?)\s*$")
from_line = re.compile(r"^\s*FROM\s+(?:--platform=[^\s]+\s+)?([^\s]+)(?:\s+AS\s+[^\s]+)?\s*(?:#.*)?$", re.IGNORECASE)

errors = []

for workflow in workflows:
    for line_number, line in enumerate(workflow.read_text().splitlines(), start=1):
        match = uses_line.match(line)
        if match:
            ref, comment = match.groups()
            if ref.startswith("./"):
                continue
            if ref.startswith("docker://"):
                image = ref.removeprefix("docker://")
                if not image_digest.search(image):
                    errors.append(f"{workflow.relative_to(root)}:{line_number}: Docker action must use a sha256 digest: {ref}")
                continue
            if not action_sha.fullmatch(ref):
                errors.append(f"{workflow.relative_to(root)}:{line_number}: action must use a full 40-character commit SHA: {ref}")
            if comment is None or not version_comment.match(comment):
                errors.append(f"{workflow.relative_to(root)}:{line_number}: action must retain its version in a trailing '# v…' comment: {ref}")
            continue

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

ruby - "$ROOT_DIR" <<'RUBY'
require "yaml"

Dir[File.join(ARGV.fetch(0), ".github", "workflows", "*.{yml,yaml}")].sort.each do |path|
  YAML.load_file(path)
end
RUBY

printf 'pinned build dependencies: all workflow actions and literal build images are immutable\n'
