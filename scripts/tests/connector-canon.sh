#!/usr/bin/env bash

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

for file in \
	docs/connector-canon/INDEX.md \
	docs/connector-canon/SOURCE-LOCK-VNEXT.md \
	docs/connector-canon/IMPLEMENTATION-PROCEDURE.md \
	docs/connector-canon/connector-terminology.md \
	docs/connector-canon/REMOTE-REPRODUCIBILITY.md \
	docs/connector-canon/foundations/README.md; do
	require_file "$file"
done

require_text docs/connector-canon/INDEX.md 'Runtime and runtime validation read execution JSON only.'
require_text docs/connector-canon/SOURCE-LOCK-VNEXT.md 'schema_version: 4'
require_text docs/connector-canon/SOURCE-LOCK-VNEXT.md 'There is no second'
require_text docs/connector-canon/SOURCE-LOCK-VNEXT.md 'Every lock declares every lane.'
require_text docs/connector-canon/IMPLEMENTATION-PROCEDURE.md 'source.lock.json'
require_text docs/connector-canon/IMPLEMENTATION-PROCEDURE.md 'All saved movement is warehouse-mediated'
require_text docs/connector-canon/REMOTE-REPRODUCIBILITY.md 'runtime admission'

[[ ! -e docs/connector-canon/DECLARATION-ADMISSION.md ]] || fail 'obsolete declaration-admission procedure remains'
[[ ! -e docs/connector-canon/OPERATION-EVIDENCE.md ]] || fail 'obsolete operation-evidence procedure remains'

for connector in github gitlab asana; do
	go run ./cmd/connectorgen lock-render "$connector" --check >/dev/null || \
		fail "$connector vNext execution bundle differs from source.lock.json"
done

printf 'connector canon check: ok\n'
