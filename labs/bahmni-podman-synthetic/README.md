# Bahmni Podman synthetic clinical lab

Task-owned assets for a reproducible local Bahmni Standard lab for connector testing.

## Pin and safety

- Pinned source: `Bahmni/bahmni-docker` tag `1.0.2-standard`, commit `1dfe62c4e5d6f3d702e65d869729726226fceb56`.
- Rootless Podman only; no Docker CLI/Engine/Desktop.
- Loopback-only service binds.
- Task-prefixed resources: Podman machine `fm-bahmni-lab-r1-machine`, compose project `fm_bahmni_lab_r1`, runtime dir `/tmp/fm-bahmni-lab-r1`. Every explicit compose `container_name` is rewritten with the `fm-bahmni-lab-r1-` prefix, and generation fails closed if any unscoped name remains.
- `reset`/`cleanup` require `--yes`, target only task-owned resources, and never create or stop the Podman machine. If the task-owned connection is unavailable they report that there is nothing to reset instead of provisioning a machine.
- Local credentials stay under `/tmp/fm-bahmni-lab-r1` and are not committed.

## Commands

```bash
labs/bahmni-podman-synthetic/bin/bahmni-lab inventory --json
labs/bahmni-podman-synthetic/bin/bahmni-lab prepare
labs/bahmni-podman-synthetic/bin/bahmni-lab start
labs/bahmni-podman-synthetic/bin/bahmni-lab health --json
labs/bahmni-podman-synthetic/bin/bahmni-lab seed --json
labs/bahmni-podman-synthetic/bin/bahmni-lab verify
```

Local URLs:

- Bahmni: `https://127.0.0.1:18443/bahmni/home/index.html`
- OpenMRS REST: `https://127.0.0.1:18443/openmrs/ws/rest/v1`
- FHIR R4: `https://127.0.0.1:18443/openmrs/ws/fhir2/R4`
- Odoo: `http://127.0.0.1:18069`
- DCM4CHEE HTTP: `http://127.0.0.1:18055`
- DICOM: `127.0.0.1:11112`

Use a private terminal for local credentials:

```bash
labs/bahmni-podman-synthetic/bin/bahmni-lab credentials --show
```

## Synthetic dataset

Fixture: `fixtures/synthetic-seed.json`.

- Hospital/location name is exactly `Chikitsalayaḥ`.
- Patients use `SYN-HEN-*` identifiers and obviously synthetic family names such as `Syntheticcase`, `Testpatient`, and `Democase`; `check-synthetic` fails if a family name carries no synthetic marker.
- Contact defaults are deliberately invalid placeholders: `000-000-0000` and `.invalid` emails.
- Karthik test patient: `SYN-HEN-0009 - Karthik Syntheticcase`.
- Karthik has a completed OPD cold/fever visit, fever temperature observation, chief complaint text, diagnosis text, completed visit stop time, appointments, lab/procedure/radiology/medication orders, allergy placeholder, and FHIR condition presence.

The dataset uses SPARSH Hennur-style taxonomy only structurally. It does not use real people, phone numbers, emails, addresses, or contacts.

## Local-only real contact override path

Real-contact testing is manual and local-only. Do not commit or log real contacts.

1. Copy an override to a gitignored file, e.g. `labs/bahmni-podman-synthetic/local-contact-overrides.json`.
2. Add human-only values there for manual UI tests.
3. Keep fixture/config/report output on invalid placeholders by default. The fixture
   `synthetic_contact` blocks are assertion-only: `check-synthetic` enforces them, and the seed
   never writes any phone or email attribute to Bahmni.
4. Delete the local override before handoff.

The repository `.gitignore` and lab `.gitignore` exclude `local-contact-overrides*.json`, `.local-credentials.json`, `.env.local`, and runtime output.

## Verification

Recommended local checks:

```bash
bash -n labs/bahmni-podman-synthetic/bin/bahmni-lab
python3 -m py_compile labs/bahmni-podman-synthetic/lib/labctl.py
labs/bahmni-podman-synthetic/bin/bahmni-lab check-synthetic --json
labs/bahmni-podman-synthetic/bin/bahmni-lab verify --offline
labs/bahmni-podman-synthetic/bin/bahmni-lab seed --json
labs/bahmni-podman-synthetic/bin/bahmni-lab seed --json   # idempotency rerun
labs/bahmni-podman-synthetic/bin/bahmni-lab verify
```

`verify` checks the fixture, live OpenMRS/FHIR records, Unicode hospital name, Karthik patient, completed visit, fever observation, cold/fever note, and FHIR condition presence.

Known API limit: the pinned OpenMRS FHIR2 Condition endpoint accepts code/text on create but does not echo condition code/text back on read. The seed script therefore uses condition presence as the idempotency guard and records cold/fever detail in encounter observations.
