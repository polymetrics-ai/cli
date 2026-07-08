# Verification Results — Issue #122

**Date:** 2026-07-08

## Official GSD/Pi Adapter Verification

```bash
scripts/gsd doctor
scripts/gsd version
scripts/gsd list
scripts/gsd sources plan-phase
scripts/gsd sources issue-122-rebootstrap
scripts/gsd prompt issue-122-rebootstrap > .planning/traces/issue-122-gsd-onboarding-prompt.md
scripts/gsd prompt issue-122-rebootstrap >/tmp/issue122-regenerated.md
diff -u .planning/traces/issue-122-gsd-onboarding-prompt.md /tmp/issue122-regenerated.md
scripts/gsd prompt plan-phase 1 --skip-research >/tmp/gsd-plan-phase.md
test -s /tmp/gsd-plan-phase.md
scripts/gsd verify-pi
```

Result: PASS.

## Non-Phase Planning and Agent Refresh Verification

Commands used to generate official prompt traces:

```bash
scripts/gsd prompt map-codebase --fast > .planning/traces/gsd-map-codebase-refresh-prompt.md
scripts/gsd prompt new-project --from-existing --non-interactive > .planning/traces/gsd-new-project-refresh-prompt.md
scripts/gsd prompt roadmap --refresh > .planning/traces/gsd-roadmap-refresh-prompt.md # expected failure: not official
scripts/gsd prompt onboard --fast --skip-phases > .planning/traces/gsd-onboard-refresh-prompt.md
scripts/gsd prompt milestone-summary --planning-only > .planning/traces/gsd-milestone-summary-refresh-prompt.md
scripts/gsd prompt docs-update .planning AGENTS.md .agents --planning-only > .planning/traces/gsd-docs-update-agents-prompt.md
scripts/gsd prompt health --context > .planning/traces/gsd-health-refresh-prompt.md
```

Final verification:

```bash
scripts/gsd doctor >/tmp/gsd-doctor.txt
scripts/gsd version >/tmp/gsd-version.json
scripts/gsd list >/tmp/gsd-list.txt
test $(wc -l < /tmp/gsd-list.txt) -ge 60
scripts/gsd verify-pi >/tmp/gsd-verify-pi.txt
scripts/gsd sources map-codebase >/tmp/gsd-sources-map-codebase.txt
scripts/gsd sources docs-update >/tmp/gsd-sources-docs-update.txt
scripts/gsd prompt map-codebase --fast >/tmp/gsd-map-codebase-refresh.md
diff -u .planning/traces/gsd-map-codebase-refresh-prompt.md /tmp/gsd-map-codebase-refresh.md
node -e "JSON.parse(require('fs').readFileSync('.planning/config.json','utf8')); JSON.parse(require('fs').readFileSync('.gsd/commands.json','utf8')); JSON.parse(require('fs').readFileSync('.gsd/upstream.lock.json','utf8')); JSON.parse(require('fs').readFileSync('.pi/settings.json','utf8')); console.log('json ok')"
python3 - <<'PY'
from pathlib import Path
import yaml
for path in sorted(Path('.agents').rglob('*.yaml')):
    yaml.safe_load(path.read_text())
print('yaml ok')
PY
rg -n "gsd-pi-adapter|/gsd|scripts/gsd prompt|load_gsd_pi_adapter" AGENTS.md .agents .planning >/tmp/gsd-adapter-refs.txt
rg -n "connector parity|reverse ETL|binary|direct-read|GraphQL|XML|SOAP|CDC|queue|human-gated|certification|conformance|de-duplication|native protocol" .planning >/tmp/planning-parity-refs.txt
git diff --check
cmd_internal=$(git diff --name-only -- cmd internal)
phases=$(git diff --name-only -- .planning/phases)
test -z "$cmd_internal"
test -z "$phases"
```

Result: PASS.

Output summary:

```text
json ok
yaml ok
non-phase gsd/pi refresh verification ok; cmd/internal and phases diffs empty
```

## CLI Help / Docs / Website Parity Guidance Verification

Commands:

```bash
scripts/gsd prompt docs-update docs/cli website/content/docs .agents .planning --cli-help-parity > .planning/traces/gsd-cli-docs-help-parity-prompt.md
rg -n "CLI help/docs/website parity|cli-help-docs-website-parity|pm connectors|namespace" .planning/traces/gsd-cli-docs-help-parity-prompt.md
rg -n "cli-help-docs-website-parity|pm connectors|bare namespace|docs/cli|website" AGENTS.md .agents .pi scripts/gsd .planning
```

Result: PASS. GSD prompts, agent references, Pi skill/prompt, AGENTS.md, and non-phase planning artifacts now require CLI-visible feature work to keep runtime help, bare namespace command behavior, `docs/cli/**`, website docs, generated help/manual artifacts, and tests in parity.

## Results

- `scripts/gsd doctor`: PASS. Official docs, command registry, lock, Pi settings, Pi extension, Pi skill, Pi prompt, and commands were detected.
- `scripts/gsd version`: PASS. Reports official `github.com/open-gsd/gsd-core@20297a8ff941378b8615a5d3e8629e52c10a0f9d` and `pi-project-local` adapter.
- `scripts/gsd list`: PASS. Official command registry prints at least 60 GSD commands; current registry has 69.
- `scripts/gsd verify-pi`: PASS. Project-local Pi settings, extension, prompt, skill, command registry, and upstream lock exist.
- Deterministic map-codebase prompt diff: PASS, no diff.
- JSON parse: PASS for `.planning/config.json`, `.gsd/commands.json`, `.gsd/upstream.lock.json`, and `.pi/settings.json`.
- YAML parse: PASS for `.agents/**/*.yaml`.
- GSD/Pi adapter reference scan: PASS. Agent and planning docs contain `gsd-pi-adapter`, `/gsd`, `scripts/gsd prompt`, and `load_gsd_pi_adapter` references.
- Multi-surface parity grep: PASS. Matches include connector parity, reverse ETL, binary, direct-read, GraphQL, XML/SOAP, CDC, queue, human-gated, certification, conformance, de-duplication, and native protocol language.
- `git diff --check`: PASS.
- `git diff --name-only -- cmd internal`: PASS, no output.
- `git diff --name-only -- .planning/phases`: PASS, no output.

## Archive Check

```bash
shasum -a 256 ../planning-archives/polymetrics-cli-issue-122-pre-rebootstrap-20260708173641.tar.gz
```

Result:

```text
e0959e4c8eba6e8610255a0cd9a98b39267902ba19600515abfdab726bfd57f5  ../planning-archives/polymetrics-cli-issue-122-pre-rebootstrap-20260708173641.tar.gz
```

## Not Run

- `go test ./...`: not required because issue #122 is planning/agent/tooling-only and no Go source changed.
- Credentialed connector checks: explicitly out of scope.
- Reverse ETL execution: explicitly out of scope.
- Runtime-backed integration tests: optional and not needed for issue #122.
