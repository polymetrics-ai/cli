#!/usr/bin/env bash
# Regenerate all recurly command-surface artifacts from the connector bundles.
# Usage: scripts/regen-recurly-surface.sh
set -euo pipefail
cd "$(dirname "$0")/.."

python3 scripts/gen-recurly-cli-surface.py
go build -o pm ./cmd/pm

rm -rf /tmp/recurly-conn-docs
./pm docs generate --dir /tmp/recurly-cli-docs --connectors-dir /tmp/recurly-conn-docs >/dev/null
cp /tmp/recurly-conn-docs/recurly/MANUAL.md docs/connectors/recurly/MANUAL.md
cp /tmp/recurly-conn-docs/recurly/SKILL.md docs/connectors/recurly/SKILL.md
cp /tmp/recurly-conn-docs/catalog/all-connectors.md docs/connectors/catalog/all-connectors.md
cp /tmp/recurly-conn-docs/catalog/all-connectors.json docs/connectors/catalog/all-connectors.json
./pm docs validate --connectors-dir docs/connectors

( cd website && node scripts/gen-connector-bundles.mjs >/dev/null && \
  node scripts/gen-connectors.mjs >/dev/null && \
  node scripts/gen-connector-catalog.mjs >/dev/null && \
  node scripts/gen-docs-data.mjs >/dev/null )

echo "regen complete"
