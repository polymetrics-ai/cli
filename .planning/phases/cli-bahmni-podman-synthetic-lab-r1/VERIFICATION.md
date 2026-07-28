# Verification checklist: Bahmni Podman synthetic lab

## Static/local checks

- [x] `bash -n labs/bahmni-podman-synthetic/bin/bahmni-lab` passed.
- [x] `python3 -m py_compile labs/bahmni-podman-synthetic/lib/labctl.py` passed.
- [x] `labs/bahmni-podman-synthetic/bin/bahmni-lab check-synthetic --json` passed: 9 patients, 12 providers, exact `Chikitsalayaḥ`, placeholder contacts only.
- [x] `labs/bahmni-podman-synthetic/bin/bahmni-lab verify --offline` passed: Unicode hospital name and Karthik fixture present.
- [x] Runnable script review confirms no Docker CLI invocation; wrapper uses `podman` and `podman-compose --podman-path` only.
- [x] Generated compose fail-closed check passed: all published ports found in `/tmp/fm-bahmni-lab-r1/.../docker-compose.podman.yml` begin with `127.0.0.1:`.
- [x] Removed generated Python `__pycache__`; local credentials/upstream runtime remain under `/tmp/fm-bahmni-lab-r1`.

## Podman safety checks

- [x] Initial read-only inventory of Podman machines/connections/resources completed before mutations.
- [x] Created/used only task-owned rootless machine `fm-bahmni-lab-r1-machine`; default Podman connection unchanged.
- [x] Running compose project is task-scoped: `fm_bahmni_lab_r1`.
- [x] Services are loopback-only on local ports: `18443`, `18069`, `18055`, `11112`.
- [x] `bahmni-lab health --json` passed against local loopback endpoints.
- [x] `bahmni-lab ps` passed through the rootless-connection guard and showed only `fm_bahmni_lab_r1_*` project containers.
- [x] Unsafe cleanup guard test passed: `BAHMNI_LAB_HOME=/tmp bahmni-lab cleanup --yes` exited 2 before reset/delete.
- [x] Reset profile list includes optional `snowstorm-lite` and `cdss` profiles as well as the Standard profiles.
- [x] No unrelated containers/networks/volumes intentionally stopped/renamed/recreated.

## Bahmni app and API checks

- [x] Browser verification with `chrome-devtools-axi` opened local Bahmni.
- [x] Real local Bahmni login succeeded with `doctor` user and local-only credential procedure; dashboard showed OPD-1 and app tiles.
- [x] Unicode rendering verified in Chrome accessibility snapshot: heading `Chikitsalayaḥ` and patient `SYN-HEN-0009 - Karthik Syntheticcase` rendered from live OpenMRS REST data. Code point for final `ḥ` was `1e25`.
- [x] Live seed passed after condition/allergy idempotency fixes.
- [x] Seed idempotency rerun passed with no nonzero `*_created` counters; final rerun after review fixes reported `conditions_existing=9`, `conditions_created=0`.
- [x] Live verify passed:
  - `ok=true`
  - `unicode_location_present=true`
  - patients `9`, providers `12`, appointments `9`, allergies `9`, orders `45`, encounters `37`, conditions `36` in the currently-running lab.
  - Karthik live checks all true: patient found, completed visit found, fever observation found, cold/fever condition+note found.
- [x] Known API limitation recorded: pinned FHIR2 Condition create accepts code/text but read/search does not echo code; seed treats condition presence as idempotency guard and stores cold/fever detail in encounter observations.

## Final repo gates

- [x] `gofmt -w cmd internal` not applicable: docs/scripts/lab fixture only, no Go edits.
- [x] `go test ./...` not run: no Go code changed; targeted script/API/browser verification above is the relevant gate.
- [x] `make verify` considered but not run because this branch adds an external Podman lab and no production Go code; targeted gates above passed.
- [x] Commit on `fm/cli-bahmni-podman-synthetic-lab-r1` completed.
- [x] Status file got terminal `done:` after commit.
