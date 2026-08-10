#!/usr/bin/env bash

# Makes the connector delivery canon discoverable and source-pinned. Keep this
# intentionally cheap: runtime executability is checked by the separate
# connector-runtime-preflight Make target, which calls the real Go test.
set -euo pipefail

repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd -P)"
cd "$repo_root"

fail() {
	printf 'connector canon check: %s\n' "$*" >&2
	exit 1
}

require_file() {
	[[ -f "$1" ]] || fail "missing required file: $1"
}

require_text() {
	local file=$1
	local text=$2
	grep -Fq -- "$text" "$file" || fail "missing required text in $file: $text"
}

require_absent() {
	local file=$1
	local text=$2
	if grep -Fq -- "$text" "$file"; then
		fail "stale or unsupported claim remains in $file: $text"
	fi
}

for file in \
	data/captain.md \
	data/CURRENT-CORRECTIONS.md \
	data/cli-database-connector-framework-design-r1/report.md \
	data/cli-cdc-large-transaction-strategy-r1/report.md \
	data/cli-cdc-bidirectional-changefeed-design-r1/report.md \
	data/cli-postgres-parity-issue-tree-r2/report.md \
	data/archive/cli-postgres-parity-issue-tree-r1/report.md \
	data/archive/cli-postgres-parity-issue-tree-r1/SUPERSEDED.md \
	data/cli-daily-use-top50-connectors-r1/report.md \
	data/archive/cli-github-etl-reverse-etl-gap-map-r1/report.md \
	data/archive/cli-github-etl-reverse-etl-gap-map-r1/SUPERSEDED.md \
	data/archive/cli-blocked-source-recovery-tiers-r1/brief.md \
	data/archive/cli-blocked-source-recovery-tiers-r1/SUPERSEDED.md \
	data/CANON-MANIFEST.sha256 \
	docs/connector-canon/INDEX.md \
	docs/connector-canon/IMPLEMENTATION-PROCEDURE.md \
	docs/connector-canon/REMOTE-REPRODUCIBILITY.md; do
	require_file "$file"
done

[[ ! -e data/cli-github-etl-reverse-etl-gap-map-r1 ]] || \
	fail 'void GitHub gap map must remain under data/archive/'
[[ ! -e data/cli-blocked-source-recovery-tiers-r1 ]] || \
	fail 'superseded blocker brief must remain under data/archive/'
[[ ! -e data/cli-postgres-parity-issue-tree-r1 ]] || \
	fail 'superseded PostgreSQL r1 tree must remain under data/archive/'

shasum -a 256 -c data/CANON-MANIFEST.sha256 >/dev/null || \
	fail 'source-pinned canon report changed; deliberately update the manifest and index'

require_text docs/connector-canon/INDEX.md 'zero**'
require_text docs/connector-canon/INDEX.md '15 entries'
require_text docs/connector-canon/INDEX.md 'ACTIVELY WRONG'
require_text docs/connector-canon/INDEX.md 'warehouse is always the mediator'
require_text docs/connector-canon/IMPLEMENTATION-PROCEDURE.md 'FOUNDATION CHECK'
require_text docs/connector-canon/IMPLEMENTATION-PROCEDURE.md 'API → warehouse → API'
require_text docs/connector-canon/IMPLEMENTATION-PROCEDURE.md 'API → warehouse → database'
require_text docs/connector-canon/IMPLEMENTATION-PROCEDURE.md 'Database → warehouse → API'
require_text docs/connector-canon/IMPLEMENTATION-PROCEDURE.md 'Database → warehouse → database'
require_text docs/connector-canon/REMOTE-REPRODUCIBILITY.md 'zero accepted live-certification artifacts'
require_text data/cli-postgres-parity-issue-tree-r2/report.md '#3987'
require_text data/cli-postgres-parity-issue-tree-r2/report.md 'API read → warehouse → API write'
require_text data/cli-postgres-parity-issue-tree-r2/report.md 'incremental_dedupe_history'
require_text data/archive/cli-postgres-parity-issue-tree-r1/SUPERSEDED.md '11-child count'
require_text data/archive/cli-github-etl-reverse-etl-gap-map-r1/SUPERSEDED.md 'Every coverage number in the report is wrong'
require_text data/archive/cli-blocked-source-recovery-tiers-r1/SUPERSEDED.md '195 genuinely blocked'
require_text data/CURRENT-CORRECTIONS.md 'accepted quarantine list has 15 entries'
require_text data/CURRENT-CORRECTIONS.md 'baseline is **zero**'

require_absent internal/connectors/defs/github/metadata.json 'full-surface certified'
require_absent internal/connectors/defs/github/docs.md 'full certification passed'
require_absent docs/connectors/github/MANUAL.md 'full-surface certified'
require_absent docs/connectors/github/SKILL.md 'full-surface certified'
require_text website/content/docs/github-cli-surface.mdx 'Current-head live proof'
require_text website/content/docs/github-cli-surface.mdx 'current source-inventory evidence'

printf 'connector canon check: ok\n'
