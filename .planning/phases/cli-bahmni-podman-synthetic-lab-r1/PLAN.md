# Plan: Bahmni Podman synthetic lab

**Branch:** `fm/cli-bahmni-podman-synthetic-lab-r1`
**Linked issue:** `Refs #516` — this lab is the local Bahmni Standard validation environment for the Bahmni connector CLI parity roadmap; it does not close that roadmap.
**Mode:** Manual GSD programming-loop fallback because `scripts/gsd prompt programming-loop ...` is not in this repo-local adapter registry.
**Safety posture:** rootless Podman only; loopback ports only; no Docker CLI/Engine/Desktop; no production credentials; no mutation of unrelated Podman workloads, machines, networks, volumes, or default connection.

## Deliverables

1. Task-owned lab assets under `labs/bahmni-podman-synthetic/`:
   - operator README and troubleshooting notes;
   - pinned source manifest;
   - `.env.example` with placeholders only;
   - Podman control script for `inventory`, `start`, `health`, `seed`, `verify`, `reset`, and `cleanup`;
   - deterministic synthetic fixtures/generator and API seeder;
   - checks that seed data remains synthetic and does not copy real provider identities.
2. Documentation of official Bahmni research, pin rationale, SPARSH taxonomy source, supported modules, and known gaps.
3. Local runtime in a task-owned temp directory with official Bahmni deployment material cloned at the pinned commit and generated Podman compose adaptations.
4. Healthy loopback-only lab left running for captain inspection.
5. Synthetic hospital name exactly `Chikitsalayaḥ`, plus Karthik synthetic cold/fever completed visit fixture and live verification.
6. Commit containing only reproducible lab assets, not cloned upstream source or local secrets.

## Work slices

### Slice A — Discovery and planning

- Verify worktree isolation and branch.
- Run `no-mistakes doctor` without restarting the daemon.
- Read AGENTS and required skills/references.
- Research official Bahmni install path with `chrome-devtools-axi` and GitHub metadata with `gh-axi`.
- Inventory Podman machines/connections/workloads read-only.
- Produce `RESEARCH.md`, this `PLAN.md`, `TDD.md`, and `VERIFICATION.md` before production edits.

### Slice B — Lab asset skeleton

- Add a task-owned directory `labs/bahmni-podman-synthetic/`.
- Add `bahmni-lab` shell script that never calls Docker and uses `podman`/`podman compose` or `podman-compose` only.
- Script defaults:
  - project name `fm_bahmni_lab_r1`;
  - machine name `fm-bahmni-lab-r1-machine` only when a task-owned machine must be created;
  - runtime root `${TMPDIR:-/tmp}/fm-bahmni-lab-r1` unless `BAHMNI_LAB_HOME` is set;
  - loopback ports `127.0.0.1:18080`, `18443`, `18069`, `18055`, `11112` remapped from official compose.
- Generate local credentials into the runtime directory only; never commit or print secrets by default.

### Slice C — Podman compose adaptation and lifecycle

- Clone `Bahmni/bahmni-docker` at commit `1dfe62c4e5d6f3d702e65d869729726226fceb56` into runtime root.
- Generate an adapted compose file in runtime root that preserves official service definitions while:
  - replacing published ports with loopback-only unique ports;
  - removing or overriding non-unique `container_name` values;
  - setting a unique compose project;
  - optionally reducing memory flags only through env values documented in the local generated env.
- Implement lifecycle commands:
  - `inventory`: read-only Podman inventory;
  - `start`: prepare runtime, create/start only task-owned Podman machine if required, run compose profile;
  - `health`: service and HTTP health checks;
  - `reset`: explicit task-only compose down + volumes for this project;
  - `cleanup`: explicit task-only clone/cache cleanup plus compose resources;
  - never stop the Podman machine automatically.

### Slice D — Synthetic data and supported API seeding

- Generate deterministic JSON fixture data for:
  - synthetic facility exactly named `Chikitsalayaḥ`, departments/specialties, locations;
  - fictional providers;
  - fictional patients and identifiers, including `SYN-HEN-0009 - Karthik Syntheticcase` with invalid placeholder contacts;
  - appointments, visits, encounters;
  - diagnoses/conditions, allergies, observations/vitals, including Karthik cold/fever chief complaint, fever observation, and completed visit stop time;
  - orders, medications, lab tests/results, procedures;
  - documents/attachments metadata;
  - module support inventory and honest gaps.
- Use OpenMRS/Bahmni REST/FHIR APIs where possible.
- Do not use direct DB writes unless an official importer is the only path; document any such exception and reset safety.
- Make seeding idempotent by deterministic external IDs/names, preflight searches, and update-or-skip behavior.
- Include a synthetic identity manifest and automated check ensuring no copied SPARSH doctor/provider identities are present and no real phone/email contacts are in fixture defaults.

### Slice E — Validation and evidence

- Add focused shell/Python validation scripts.
- Run static checks on scripts and JSON fixtures.
- Run Podman stack health, seed, verify endpoints, and browser login/navigation verification with `chrome-devtools-axi`.
- Record dataset counts, known gaps, URLs, pin, and local login procedure without exposing credentials.
- Commit lab assets on the task branch.

## Risks and mitigations

- **Podman machine absent/stopped:** create/start only the task-owned machine if needed; never change default connection or unrelated machines.
- **Compose incompatibility:** generate a patched task-owned compose file; if Podman cannot run official services safely, report a blocker rather than altering unrelated resources.
- **Official images/ports require privileged or public binds:** use non-privileged loopback host ports.
- **Secrets in git/status:** generated credentials stay under runtime root; committed `.env.example` uses placeholders.
- **Synthetic data duplication:** deterministic IDs and preflight searches make seeding idempotent.
- **Unsupported modules:** verification report records modules that are disabled, inaccessible, or unavailable in the selected profile.
