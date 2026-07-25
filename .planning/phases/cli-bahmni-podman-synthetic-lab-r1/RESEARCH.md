# Research: Bahmni Podman synthetic lab

**Task branch:** `fm/cli-bahmni-podman-synthetic-lab-r1`
**Retrieval date:** 2026-07-26
**GSD path:** `scripts/gsd doctor`; `scripts/gsd prompt plan-phase cli-bahmni-podman-synthetic-lab-r1 --skip-research`; attempted `scripts/gsd prompt programming-loop init --phase cli-bahmni-podman-synthetic-lab-r1 --dry-run` but this adapter registry does not contain `programming-loop`, so this phase records a manual GSD/TDD fallback.
**Required skills loaded:** `gsd-core`, `golang-how-to`, `golang-security`, `golang-safety`, `golang-testing`, `golang-documentation`, plus repo references `required-skills-routing.md`, `runtime-rlm-website-integration.md`, `gsd-pi-adapter.md`.

## Official Bahmni source trail

Browser verification used `chrome-devtools-axi` in session `bahmni-lab-r1`.

1. `https://www.bahmni.org/`
   - The public site describes Bahmni as an EMR/HMIS combining OpenMRS, OpenERP/Odoo, OpenELIS, DICOM/PACS, and related modules.
   - The site footer links the Bahmni Wiki and GitHub organization.
2. `https://www.bahmni.org/install`
   - The install page says Docker is the recommended way to try Bahmni and links to the wiki page “Running Bahmni on Docker”.
   - It states older RPM options are no longer supported and asks users to use Docker versions.
3. `https://bahmni.atlassian.net/wiki/spaces/BAH/pages/299630726/Running+Bahmni+on+Docker`
   - The Docker wiki page links to “Getting Started Quickly with Bahmni on Docker”, Bahmni Lite, Bahmni Standard, compose profiles, and per-service configuration pages.
4. `https://bahmni.atlassian.net/wiki/spaces/BAH/pages/5519474693/Bahmni+Security+Patch+July+02+2026+Release+Notes`
   - The July 02 2026 security patch recommends Docker deployments upgrade to patched Docker distribution tags.
   - It lists `1.0.2-standard` for Bahmni Standard and `1.0.2-lite` for Bahmni Lite.
   - For Standard it updates Bahmni Docker from `1.0.1-standard` to `1.0.2-standard`, OpenMRS image from `1.1.2` to `1.1.3`, and `bahmni-core` from `1.2.0` to `1.2.1`.
5. `https://bahmni.atlassian.net/wiki/spaces/BAH/pages/3757965318/Bahmni+Standard+Releases`
   - Standard releases page lists 1.0.1 as the current security patch line and describes Standard as the large-hospital distribution with upgraded OpenMRS, inpatient features, Odoo 16 LTS, SNOMED integration, and enhanced security.
6. GitHub via `gh-axi`:
   - Repository: `Bahmni/bahmni-docker` (`https://github.com/Bahmni/bahmni-docker`), described as “Bahmni docker compose setup to run LITE and STANDARD images”.
   - Tag `1.0.2-standard` resolves through annotated tag object `e8e98dd61bc9be36490c8b26059fd331eae9405d` to commit `1dfe62c4e5d6f3d702e65d869729726226fceb56`.
   - Commit message: `BSL-13,BSL-14 | Upgrade. Bump OpenMRS image version to 1.1.3`.

## Pin decision

Pin the local lab to **Bahmni Standard Docker tag `1.0.2-standard` / commit `1dfe62c4e5d6f3d702e65d869729726226fceb56`**.

Why this pin:

- It is the latest official security patch trail discovered from the official site and wiki.
- It is immutable by commit SHA and maps to an official Bahmni Docker tag.
- Standard exposes the broadest connector-relevant clinical surface for a local lab: Bahmni Web, OpenMRS, OpenELIS/lab, Odoo, reports, appointments, inpatient microfrontend, patient documents, PACS services, and optional SNOMED/CDSS profiles.
- It lets unsupported/unavailable modules be recorded explicitly instead of choosing Lite and silently omitting large-hospital modules.

## Podman adaptation constraints

Official Bahmni docs say Docker/Docker Compose. This lab must run with rootless Podman only, so scripts will:

- clone the official repository into a task-owned local runtime directory rather than vendoring third-party source;
- checkout the exact commit SHA;
- generate a task-owned Podman compose working copy;
- bind all published ports to `127.0.0.1` on unique non-privileged host ports;
- use a unique compose project name and task-prefixed Podman resources;
- avoid Docker Engine, Docker Desktop, and Docker CLI;
- inventory existing Podman machines/connections/containers/networks/volumes read-only before starting.

## Synthetic data source constraint

The task requires using SPARSH Hospital Hennur’s publicly advertised specialty taxonomy only as a structural example. The seed must not copy real doctor/provider identities, biographies, schedules, phone numbers, email addresses, patient data, or other personal data. The lab will record the SPARSH source URL and retrieval date, use only specialty/department structure, and seed a clearly synthetic organization.
