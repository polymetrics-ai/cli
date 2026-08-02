# Verification checklist: Bahmni Podman synthetic lab

## Static/local checks

- [x] `bash -n labs/bahmni-podman-synthetic/bin/bahmni-lab` passed.
- [x] `python3 -m py_compile labs/bahmni-podman-synthetic/lib/labctl.py` passed.
- [x] `labs/bahmni-podman-synthetic/bin/bahmni-lab check-synthetic --json` passed: 10 patients, 87 providers/staff, exact `Chikitsalayaḥ`, placeholder contacts only, and clinical-facing text checked for no repeated `synthetic` wording.
- [x] `labs/bahmni-podman-synthetic/bin/bahmni-lab check-synthetic --json --online-source` passed: no synthetic provider/staff name collides with the public SPARSH Hennur source text.
- [x] `labs/bahmni-podman-synthetic/bin/bahmni-lab verify --offline` passed: Unicode hospital name, Karthik fixture, Rohit fixture, and 7 Karthik/Rohit history events present.
- [x] Runnable script review confirms no Docker CLI invocation; wrapper uses `podman` and `podman-compose --podman-path` only.
- [x] Generated compose fail-closed check passed: all published ports found in `/tmp/fm-bahmni-lab-r1/.../docker-compose.podman.yml` begin with `127.0.0.1:`.
- [x] Removed generated Python `__pycache__`; local credentials/upstream runtime remain under `/tmp/fm-bahmni-lab-r1`.

## Podman safety checks

- [x] Initial read-only inventory of Podman machines/connections/resources completed before mutations.
- [x] Created/used only task-owned rootless machine `fm-bahmni-lab-r1-machine`; default Podman connection unchanged.
- [x] Running compose project is task-scoped: `fm_bahmni_lab_r1`.
- [x] Services are loopback-only on local ports: `18443`, `18069`, `18055`, `11112`.
- [x] `bahmni-lab health --json` passed against local loopback endpoints after recovery restart. `bahmni-lab start` now waits/retries health so cold OpenMRS startup does not leave a false failed start.
- [x] `bahmni-lab ps` passed through the rootless-connection guard and showed only `fm_bahmni_lab_r1_*` project containers.
- [x] Unsafe cleanup guard test passed: `BAHMNI_LAB_HOME=/tmp bahmni-lab cleanup --yes` exited 2 before reset/delete.
- [x] Reset profile list includes optional `snowstorm-lite` and `cdss` profiles as well as the Standard profiles.
- [x] No unrelated containers/networks/volumes intentionally stopped/renamed/recreated.

## Bahmni app and API checks

- [x] Browser verification with `chrome-devtools-axi` opened local Bahmni.
- [x] Real local Bahmni login succeeded with the bundled `doctor` demo user and private local credential entry; dashboard showed OPD-1 and app tiles. Recovery run also opened OpenMRS REST pages in Chrome showing `SYN-HEN-0010 - Rohit Nair` and provider `SYN-STAFF-NURS-001 - Aditi Rao`.
- [x] Unicode rendering verified in Chrome accessibility snapshot: heading `Chikitsalayaḥ` and the then-seeded Karthik patient rendered from live OpenMRS REST data. Code point for final `ḥ` was `1e25`. Slice H updates the reproducible fixture display name to `SYN-HEN-0009 - Karthik Iyer`; live UI rerender remains blocked pending connector clearance.
- [x] Live seed passed after condition/allergy idempotency fixes.
- [x] Seed idempotency rerun passed with no nonzero `*_created` counters; final rerun after review fixes reported `conditions_existing=9`, `conditions_created=0`.
- [x] Live verify passed after recovery reseed:
  - `ok=true`
  - `unicode_location_present=true`
  - patients `10`, providers/staff `87`, appointments `10`, allergies `10`, orders `50`, encounters `100`, conditions `37` in the currently-running lab.
  - FHIR exact Patient readback count `10`; FHIR Observation readback total for the configured weight concept `28`.
  - Karthik live checks all true: patient found, completed visit found, fever observation found, cold/fever condition+note found.
  - Karthik/Rohit history checks all true: `SYN-HEN-0009` found `3/3`; `SYN-HEN-0010` found `4/4`.
- [x] Known API limitation recorded: pinned FHIR2 Condition create accepts code/text but read/search does not echo code; seed treats condition presence as idempotency guard and stores cold/fever detail in encounter observations.

## Extended staff roster checks (Slice F)

- [x] `check-synthetic --json` reports expanded staff/provider totals with no identity/contact failures: 87 providers checked (23 clinical + 64 staff), 9 patients.
- [x] `verify --offline` reports expanded provider count and unchanged 9 synthetic patients.
- [x] `seed --dry-run --json` reports planned non-mutating counts: 87 providers, 23 clinical providers, 64 staff, 46 departments, 9 patients.
- [x] Live `seed --json` created missing synthetic staff/provider records without printing secrets: recovery first run created `75` providers/staff on top of the prior 12.
- [x] Second live `seed --json` is idempotent with no nonzero `*_created` counters and reports `providers_existing=87`.
- [x] Live `verify` passes with provider count equal to fixture total and Karthik checks still true.
- [x] Spot-check OpenMRS/Bahmni provider search for a synthetic staff role passed in Chrome/OpenMRS REST: `SYN-STAFF-NURS-001 - Aditi Rao`.

## Rohit and longitudinal history checks (Slice G)

- [x] `check-synthetic --json` reports 10 synthetic patients and rejects real-looking phone/contact data.
- [x] `verify --offline` reports Rohit present and Karthik/Rohit history event counts: 7 total, 3 Karthik, 4 Rohit.
- [x] `seed --dry-run --json` reports planned history events without mutating live Bahmni.
- [x] Live `seed --json` creates Rohit/history records without printing secrets or real contact data: recovery first run created `patients_created=1` and history/clinical encounters; REST spot-check showed `SYN-HEN-0010 - Rohit Nair`.
- [x] Second live `seed --json` is idempotent with no nonzero `*_created` counters.
- [x] Live `verify` passes with 10 patients and Karthik/Rohit history checks true.

## Realistic fictional clinical wording checks (Slice H)

- [x] `check-synthetic --json` passes with Indian fictional names and `SYN-*` identifiers; recovery run reported 10 patients, 87 providers/staff, and `clinical_texts_checked=100`.
- [x] Fixture/code scan confirms the user-provided real phone number, real-looking phone patterns, and removed fake-family marker names are absent from `labs/bahmni-podman-synthetic/{fixtures,lib,config,README.md}`.
- [x] Patient history/chief-complaint/diagnosis/billing/medication notes no longer rely on repeated `synthetic` wording in clinical-facing text; `check-synthetic` now enforces this for patient-record-facing fixture strings.
- [x] `verify --offline` preserves 10 patients, 87 providers/staff, Rohit present, and 7 Karthik/Rohit history events.
- [x] `seed --dry-run --json` reports planned counts only and performs no live writes.
- [x] Live `seed --json`, live idempotency rerun, and live `verify` passed after read-only Podman inventory showed no running unrelated workloads and only task-owned resources were started/mutated.

## Final repo gates

- [x] Slice J focused standard-library regression tests passed: 23 tests cover the pinned singular appointment update route, preserved appointment state, legacy marker migration, cleanup/restart ownership continuity, normalized contact candidates, exact loopback API origins, and authenticated redirect refusal.
- [x] Slice J source review confirms REST and FHIR authentication can target only the configured HTTPS loopback origin, and existing runtime ownership proof migrates forward before lifecycle actions.
- [x] Slice J performed no live Podman, Bahmni, push, PR, CI, or outer-pipeline action.
- [x] Slice I focused standard-library regression tests passed: 17 tests covering exact-name reconciliation, legacy/ambiguous record upgrades, adversarial contacts, shell parsing, fixture validation, and ownership guards.
- [x] Slice I source review confirms every Podman mutation path requires matching ownership state and compose resource labels.
- [x] Slice I live reconciliation is intentionally left to the outer validation/runtime phase; this review phase did not mutate the running lab.

- [x] `gofmt -w cmd internal` not applicable: docs/scripts/lab fixture only, no Go edits.
- [x] `go test ./...` not run: no Go code changed; targeted script/API/browser verification above is the relevant gate.
- [x] `make verify` considered but not run because this branch adds an external Podman lab and no production Go code; targeted gates above passed.
- [x] Slice H recovery commit on `fm/cli-bahmni-podman-synthetic-lab-r1` prepared with final offline/live validation evidence.
- [x] Status file terminal `done:` to be appended immediately after commit.
