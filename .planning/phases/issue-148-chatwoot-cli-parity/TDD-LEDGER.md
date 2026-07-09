# TDD Ledger: Chatwoot CLI Parity Parent Orchestration

## 2026-07-10 — parent planning checkpoint

Task type: parent orchestration and connector metadata planning.

GSD command evidence:

- `scripts/gsd doctor`: pass.
- `scripts/gsd verify-pi`: pass.
- `scripts/gsd list --json`: pass.
- `scripts/gsd prompt programming-loop init --phase issue-148-chatwoot-cli-parity --dry-run`: failed because `programming-loop` is not registered in `scripts/gsd`.
- Manual GSD fallback recorded in `PLAN.md` and trace files.

Red evidence: not applicable for planning-only seed; no production behavior changed.

Next red validation for #149:

```bash
python3 - <<'PY'
import json, urllib.request, collections
url = 'https://raw.githubusercontent.com/chatwoot/chatwoot/develop/swagger/swagger.json'
swagger = json.load(urllib.request.urlopen(url, timeout=30))
official = [(m.upper(), p) for p, item in swagger['paths'].items() for m in item if m.lower() in {'get','post','put','patch','delete'}]
with open('internal/connectors/defs/chatwoot/api_surface.json') as f:
    surface = json.load(f)['endpoints']
print({'official': len(official), 'surface': len(surface), 'official_methods': dict(collections.Counter(m for m, _ in official))})
raise SystemExit(0 if len(official) == len(surface) else 1)
PY
```

Expected before #149 production edits: exit 1 because official=144 and current surface=71.

## Required skill evidence

Loaded skills: `gsd-core`, `caveman`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-lint`.

## Human gates

- Parent PR merge to `main`.
- New dependencies.
- Auth scope changes.
- Secrets or credentialed connector checks.
- Destructive external actions.
- Reverse ETL execution.
- Quality gate reduction.
