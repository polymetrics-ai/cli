# Bahmni Podman synthetic clinical lab

Task-owned assets for a reproducible local Bahmni Standard lab for connector testing.

## Pin and safety

- Pinned source: `Bahmni/bahmni-docker` tag `1.0.2-standard`, commit `1dfe62c4e5d6f3d702e65d869729726226fceb56`.
- Rootless Podman only; no Docker CLI/Engine/Desktop.
- Loopback-only service binds.
- Task-prefixed resources: Podman machine/connection `fm-bahmni-lab-r1-machine`, compose project `fm_bahmni_lab_r1`, runtime dir `/tmp/fm-bahmni-lab-r1`. The first safe start atomically writes private ownership state under `${XDG_STATE_HOME:-~/.local/state}/fm-bahmni-lab-r1/ownership.json`, binding those exact values and any recovered network/volume identities. Cleanup retains that proof while the owned machine remains; later machine, connection, compose, reset, cleanup, and restart paths refuse missing or mismatched markers.
- Every explicit compose `container_name` is rewritten with the `fm-bahmni-lab-r1-` prefix, and every service, named volume, and default network receives the `io.polymetrics.bahmni-lab.owner` label. Project actions verify labels or exact recovered identities before they run. Generation fails closed if names or labels are unscoped.
- `reset`/`cleanup` require `--yes`, target only marker-bound resources, and never create, start, or stop the Podman machine. If the task-owned connection is unavailable they report that there is nothing to reset.
- `ownership-recover --yes` recreates lost durable state only after the exact machine/connection and every task-project resource can be inspected; containers without the explicit owner label are never adopted. `ownership-forget --yes` removes the state only after its exact machine and connection no longer exist.
- Local credentials stay under `/tmp/fm-bahmni-lab-r1` and are not committed.

## Commands

```bash
labs/bahmni-podman-synthetic/bin/bahmni-lab inventory --json
labs/bahmni-podman-synthetic/bin/bahmni-lab prepare
labs/bahmni-podman-synthetic/bin/bahmni-lab start
labs/bahmni-podman-synthetic/bin/bahmni-lab health --json
labs/bahmni-podman-synthetic/bin/bahmni-lab seed --json
labs/bahmni-podman-synthetic/bin/bahmni-lab verify
labs/bahmni-podman-synthetic/bin/bahmni-lab ownership-recover --yes  # only after durable state loss
```

Local URLs:

- Bahmni: `https://127.0.0.1:18443/bahmni/home/index.html`
- OpenMRS REST: `https://127.0.0.1:18443/openmrs/ws/rest/v1`
- FHIR R4: `https://127.0.0.1:18443/openmrs/ws/fhir2/R4`
- Odoo: `http://127.0.0.1:18069`
- DCM4CHEE HTTP: `http://127.0.0.1:18055`
- DICOM: `127.0.0.1:11112`

Use a private terminal for local API credentials:

```bash
labs/bahmni-podman-synthetic/bin/bahmni-lab credentials --show
```

For browser inspection, open the Bahmni URL, accept the loopback-only self-signed certificate, log in with the pinned Bahmni Standard bundled demo `doctor` user using the private local demo password, and select `OPD-1` as the login location. Do not paste credentials into issues, PRs, status messages, or shell transcripts.

## Synthetic dataset

Fixture: `fixtures/synthetic-seed.json`.

- Hospital/location name is exactly `Chikitsalayaḥ`.
- Patients use `SYN-HEN-*` identifiers, invalid placeholder contacts, and clearly fictional Indian names. The visible clinical names are realistic for UI/API testing; the synthetic proof is carried by identifiers, fixture metadata, and automated checks rather than fake-looking family-name markers.
- Clinical providers use `SYN-PROV-*`; the extended non-patient staff roster uses `SYN-STAFF-*` and is represented as OpenMRS Provider/Person records for connector inspection. Names are fictional and checked against forbidden/source-collision identity patterns.
- Staff coverage includes consultant/resident doctors, nursing, allied health, lab, radiology, pharmacy, front office, billing, medical records, administration, HR, finance, procurement, IT, biomedical engineering, facilities, housekeeping, security, dietary, ambulance/transport, CSSD, infection control, quality/compliance, and patient relations.
- Contact defaults are deliberately invalid placeholders: `000-000-0000` and `.invalid` emails.
- The pre-seed fixture scan and the API write boundary both reject formatted Indian mobile/landline variants and every email outside the reserved `.invalid` domain.
- Karthik test patient: `SYN-HEN-0009 - Karthik Iyer`.
- Rohit test patient: `SYN-HEN-0010 - Rohit Nair`; contact data remains the invalid placeholder and no real phone number is stored in fixture/config/log output.
- Karthik has a completed OPD cold/fever visit, fever temperature observation, chief complaint text, diagnosis text, completed visit stop time, appointments, lab/procedure/radiology/medication orders, allergy placeholder, and FHIR condition presence.
- Karthik and Rohit also have longitudinal fictional history events for intake, diagnostics/lab results, medication notes, follow-up/discharge, document metadata, and simulated billing notes. Billing is represented as OpenMRS document/encounter text for connector testing, not as a real Odoo invoice.

The dataset uses SPARSH Hennur-style taxonomy only structurally. It does not use real people, phone numbers, emails, addresses, or contacts.

The seed path does not accept a real-contact override. Local ignored files remain outside the fixture, but the recursive preflight and API write guard refuse their realistic contacts if they are ever merged into a generated payload.

## Verification

Recommended local checks. During active connector live verification, use only the static/offline commands until the connector owner clears a live reseed:

```bash
bash -n labs/bahmni-podman-synthetic/bin/bahmni-lab
python3 -m py_compile labs/bahmni-podman-synthetic/lib/labctl.py
python3 -m unittest discover -s labs/bahmni-podman-synthetic/tests -v
labs/bahmni-podman-synthetic/bin/bahmni-lab check-synthetic --json
labs/bahmni-podman-synthetic/bin/bahmni-lab check-synthetic --json --online-source  # optional network collision check against the public taxonomy page
labs/bahmni-podman-synthetic/bin/bahmni-lab verify --offline
labs/bahmni-podman-synthetic/bin/bahmni-lab seed --dry-run --json
# Live mutation gates; run only after explicit clearance from active connector verification:
# labs/bahmni-podman-synthetic/bin/bahmni-lab seed --json
# labs/bahmni-podman-synthetic/bin/bahmni-lab seed --json   # idempotency rerun
# labs/bahmni-podman-synthetic/bin/bahmni-lab verify
```

`verify` checks exact expected provider/patient display names and allergy comments, one canonical stable appointment per synthetic patient, stable history-event markers, the fixture, live OpenMRS/FHIR records, Unicode hospital name, Karthik patient, completed visit, fever observation, cold/fever note, and FHIR condition presence.

Known API limit: the pinned OpenMRS FHIR2 Condition endpoint accepts code/text on create but does not echo condition code/text back on read. The seed script therefore uses condition presence as the idempotency guard and records cold/fever detail in encounter observations.
