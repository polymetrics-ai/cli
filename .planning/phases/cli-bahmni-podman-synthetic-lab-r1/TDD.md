# TDD ledger: Bahmni Podman synthetic lab

**Manual GSD/TDD fallback:** `scripts/gsd prompt programming-loop ...` was unavailable in the adapter registry. This ledger captures red/green intent and evidence for shell/Python lab assets.

## Test targets

| Slice | Test or check | Red expectation | Green expectation | Status |
|---|---|---|---|---|
| B | `bash -n labs/bahmni-podman-synthetic/bin/bahmni-lab` | script absent/fails parse | no shell syntax errors | green |
| B | `python3 -m py_compile labs/bahmni-podman-synthetic/lib/labctl.py` | module absent/fails parse | no Python syntax errors | green |
| C | `bahmni-lab inventory --json` | command absent | read-only machine/connection summary; no mutations | green |
| C | `bahmni-lab prepare --dry-run` | command absent | prints planned clone/patch paths and pin without secrets | green |
| C | generated compose path | compose absent | loopback-only ports and task-prefixed runtime files | green from running lab |
| C | generated compose port scan | upstream compose can leak broad binds | every published port starts with `127.0.0.1:` | green |
| C | unsafe cleanup guard | env override can point at broad path | `BAHMNI_LAB_HOME=/tmp cleanup --yes` exits before deletion | green |
| C | rootless connection guard | root/rootful connection could be accepted | `podman info` must report rootless before start/ps/reset | green |
| D | `check-synthetic --json` | fixture absent or real identity/contact leaks | exact synthetic hospital, 9 synthetic patients, 12 synthetic providers, invalid contact placeholders | green |
| D | `seed --json` | no seeder or API failures | supported REST/FHIR writes complete without printing secrets | green |
| D | second `seed --json` | duplicates or non-idempotent counters | no nonzero `*_created` counters on rerun | green |
| E | `verify --offline` | verifier absent | validates fixture, Unicode name, Karthik fixture | green |
| E | live `health`, `seed`, `verify` | stack absent/API failures | loopback services healthy and endpoints return synthetic records | green |
| E | browser Unicode/login check | Chrome cannot open/login/render | Bahmni dashboard login succeeds; Unicode heading `Chikitsalayaḥ` renders | green |

## Evidence log

- 2026-07-26: planning artifacts created before production lab edits.
- 2026-07-26: offline checks passed: shell syntax, Python compile, `check-synthetic --json`, `verify --offline`.
- 2026-07-26: live seeding initially exposed two useful red states: duplicate visit matching by date only and allergy/condition idempotency gaps. Fixed by exact visit start matching, allergy coded-allergen guard, and FHIR Condition presence guard.
- 2026-07-26: final idempotency rerun passed with `created_nonzero={}` and `conditions_created=0`.
- 2026-07-26: live verify passed for Unicode location, Karthik completed cold/fever visit, fever observation, and condition+note presence.
- 2026-07-26: `chrome-devtools-axi` verified Bahmni real local login and rendered `Chikitsalayaḥ` with final code point `U+1E25`.
