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
| F | `seed --json` after staff expansion | new staff records fail OpenMRS API create/search | new `providers_created` count equals unseeded staff records; rerun creates zero records | green live |
| F | `verify` after staff expansion | provider count remains old or staff missing | live provider count equals doctors plus extended staff roster | green live |
| G | `check-synthetic --json` after Rohit/history expansion | real phone or non-synthetic identity appears in fixture | 10 synthetic patients, no real-looking phone, invalid placeholders only | green offline |
| G | `verify --offline` after Rohit/history expansion | patient/history counts stale | offline reports Rohit present and Karthik/Rohit history events planned | green offline |
| G | `seed --dry-run --json` after Rohit/history expansion | dry-run omits history events | reports planned history event count without live API writes | green offline |
| G | live `seed --json` and idempotency | live mutation while connector proof is active or records duplicate | live seed adds Rohit/history once and rerun creates no records | green live |
| H | `check-synthetic --json` after realistic-name pass | Indian-style names fail old fake-family marker check or real phone leaks into fixtures | 10 patients, 87 providers/staff pass via `SYN-*` IDs, no real-looking phone numbers, and clinical-facing text excludes repeated `synthetic` wording | green offline |
| H | clinical-facing text scan | patient history/record notes contain repeated `synthetic` wording | patient history/record notes use realistic fictional clinical wording while metadata/docs retain safeguards | green offline |
| H | `verify --offline` and `seed --dry-run --json` | offline counts regress | offline counts remain 10 patients, 87 providers/staff, 7 history events | green offline |
| I | exact owned-identifier name reconciliation | existing stable IDs retain obsolete display names | only exact task-owned identifiers update through the Person name subresource and verify exact expected displays | green focused tests |
| I | legacy encounter/appointment upgrade | mutable prose creates replacement records and retains old marker text | legacy/current markers migrate to one stable canonical identity without deletes or new duplicates | green focused tests |
| I | adversarial contact validation | spaced/grouped phones, landlines, and non-`.invalid` emails pass pre-seed checks | every fixture and outgoing write string rejects those variants while `.invalid` placeholders pass | green focused tests |
| I | Podman ownership guard | environment overrides can adopt an unrelated machine, connection, or compose project | invalid names and missing/mismatched markers fail before resource actions; compose labels are verified before project operations | green focused tests |
| J | pinned appointment update contract | legacy reconciliation posts to an unsupported plural UUID route and overwrites live status/provider state | singular `/appointment` receives the UUID and preserves every existing mutable appointment field | green focused tests |
| J | cleanup/restart ownership continuity | cleanup deletes the proof needed to reuse the intentionally retained machine/connection | ownership proof survives runtime cleanup and safely authorizes the same exact resources on restart | green focused tests |
| J | normalized contact candidates | sentence-punctuated emails and compact country-code landlines bypass regex-only checks | normalized candidates reject every non-`.invalid` email and realistic compact/formatted Indian contact | green focused tests |
| J | authenticated API origin boundary | remote, unsafe, ambiguous, or redirected targets can receive the local Basic credential | only the configured HTTPS IPv4/IPv6/localhost loopback REST origin is accepted and redirects never receive credentials | green focused tests |
| K | lossless appointment reconciliation | the pinned mapper can reactivate voided provider associations submitted from a lossy representation | reconciliation refuses before writing whenever any provider association cannot be represented losslessly | green focused tests |
| K | punctuation-normalized contacts | Unicode dashes and slash-separated Indian contacts bypass the candidate extractor | common punctuation and compact country-code forms normalize and fail before API writes | green focused tests |
| K | proxy-free authenticated transport | environment/system proxies can receive loopback Basic credentials | authenticated REST/FHIR openers install an empty proxy handler and refuse redirects | green focused tests |
| K | durable exact Podman ownership | project-name collisions or temporary-file purges can authorize unrelated resources or strand the owned machine | private atomic XDG proof binds exact identities; containers require owner labels and networks/volumes require labels or recorded IDs | green focused tests |
| K | allergy prose upgrade | a coded-allergen match retains obsolete clinical prose | only the exact owned patient/allergen legacy comment updates through the supported allergy subresource and live verification requires the exact new comment | green focused tests |

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
- 2026-07-26: Slice G accepted for offline-only Rohit/history fixture work. The user-provided real phone is intentionally not stored; live application remains blocked pending connector clearance.
- 2026-07-26: Offline Rohit/history validation passed: `check-synthetic --json` ok for 10 patients and 87 providers; `verify --offline` ok with Rohit present and 7 history events (3 Karthik, 4 Rohit); `seed --dry-run --json` planned 7 history events without live writes.
- 2026-07-26: Slice H accepted for offline-only clinical realism pass. Boundary recorded: make fictional records clinically plausible and Indian-named, but keep `SYN-*` identifiers, no real phone/contact data, and no live mutation until cleared.
- 2026-08-02: Recovery validation passed: `check-synthetic --json` ok for 10 patients / 87 providers-staff with `clinical_texts_checked=100`; `check-synthetic --json --online-source` ok without source collisions; `verify --offline` ok with 7 history events; `seed --dry-run --json` reports planned counts only.
- 2026-08-02: Live recovery validation passed after read-only Podman inventory and task-owned restart: health ok, first seed added missing Slice F/G/H records, second seed had `created_nonzero={}`, live verify ok with 10 patients, 87 providers/staff, 10 appointments, 50 orders, 100 encounters, FHIR exact Patient count 10, and Karthik/Rohit history checks true.
- 2026-08-02: Cold start initially showed OpenMRS health not ready while Bahmni Web was already serving; `bahmni-lab start` now retries health before returning.
- 2026-08-02: Review Slice I planned under the recorded manual GSD fallback after `scripts/gsd prompt programming-loop ...` remained unavailable. Required skills loaded: `no-mistakes`, `gsd-programming-loop`, `golang-how-to`, `golang-security`, `golang-safety`, `golang-error-handling`, `golang-lint`, and `golang-testing`. The assigned one-verification rule supersedes an executable pre-fix red run; regression tests are written first and remain an explicit red specification until the complete fix round.
- 2026-08-02: Slice I focused run initially exposed unsupported parsing of Bahmni's `+0000` appointment offset in the new stable key. The parser was normalized, and the confirmation run `python3 -m unittest discover -s labs/bahmni-podman-synthetic/tests -v` passed all 17 tests. No live Podman or Bahmni operation ran.
- 2026-08-02: Review Slice J uses the manual GSD fallback because `scripts/gsd doctor` passed but `scripts/gsd prompt programming-loop init --phase cli-bahmni-podman-synthetic-lab-r1 --dry-run` reported an unknown command. The pinned 1.8.0-AGPLv3 Bahmni controllers confirm plural `/appointments` creates and singular `/appointment` updates by request UUID. Round-two regression tests were written as an unexecuted red specification under the assigned one-verification rule.
- 2026-08-02: Slice J's focused run first stopped at a test-module indentation error before executing behavior. After that test-only correction, the confirmation run `python3 -m unittest discover -s labs/bahmni-podman-synthetic/tests -v` passed all 23 tests. No live Podman or Bahmni operation ran.
- 2026-08-02: Review Slice K uses the manual GSD fallback because `scripts/gsd doctor` passed but `scripts/gsd prompt programming-loop init --phase cli-bahmni-podman-synthetic-lab-r1 --dry-run` reported an unknown command. The pinned appointment mapper and OpenMRS patient-allergy resource were inspected before the regression specification. The assigned one-verification rule supersedes a pre-fix red run.
- 2026-08-02: Slice K's first focused run passed 29 of 32 tests and exposed two representation edges before confirmation: maximal phone candidates had to start before ISO timestamp separators, and Python does not retain an empty proxy handler in the opener's public handler list even though the explicit handler suppresses proxy discovery. The extractor now consumes maximal numeric punctuation sequences for classification, and the client retains its explicit empty handler for regression inspection.
- 2026-08-02: Slice K focused confirmation `python3 -m unittest discover -s labs/bahmni-podman-synthetic/tests -v` passed all 32 tests. No live Podman, Bahmni, push, PR, CI, or outer-pipeline action ran.
