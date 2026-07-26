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
| F | `check-synthetic --json` after staff expansion | staff fixture absent, non-synthetic identifiers, real contacts, or forbidden identity tokens | validates expanded doctors/staff roster as synthetic and contact-free | green offline |
| F | `seed --dry-run --json` after staff expansion | planned counts do not include staff/providers | dry-run reports 87 providers = 23 clinical + 64 staff, 46 departments, 9 patients | green offline |
| F | `seed --json` after staff expansion | new staff records fail OpenMRS API create/search | new `providers_created` count equals unseeded staff records; rerun creates zero records | blocked pending connector clearance |
| F | `verify` after staff expansion | provider count remains old or staff missing | live provider count equals doctors plus extended staff roster | blocked pending connector clearance |

## Evidence log

- 2026-07-26: planning artifacts created before production lab edits.
- 2026-07-26: offline checks passed: shell syntax, Python compile, `check-synthetic --json`, `verify --offline`.
- 2026-07-26: live seeding initially exposed two useful red states: duplicate visit matching by date only and allergy/condition idempotency gaps. Fixed by exact visit start matching, allergy coded-allergen guard, and FHIR Condition presence guard.
- 2026-07-26: final idempotency rerun passed with `created_nonzero={}` and `conditions_created=0`.
- 2026-07-26: live verify passed for Unicode location, Karthik completed cold/fever visit, fever observation, and condition+note presence.
- 2026-07-26: `chrome-devtools-axi` verified Bahmni real local login and rendered `Chikitsalayaḥ` with final code point `U+1E25`.
- 2026-07-26: New staff-roster request accepted under Slice F; plan/TDD/verification updated before fixture/seeder edits.
- 2026-07-26: User requested coordination with active connector verification; no live restart/reseed/container or current dataset mutation is allowed until `cli-bahmni-connector-r1` finishes. Offline-only work continues.
- 2026-07-26: Offline staff validation passed: `check-synthetic --json` ok for 87 providers (23 clinical + 64 staff) and 9 patients; `verify --offline` ok; `seed --dry-run --json` planned 87 providers and 46 departments without live API writes.
