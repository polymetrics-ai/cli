# UAT — Issue #4083

Status: Docker live proof passed; Podman live proof remains unavailable.

| Deliverable | Evidence | Verdict |
| --- | --- | --- |
| Explicit runtime selection | Focused dbtest green selector passes for Docker and Podman; unknown and omitted values are rejected. | Pass |
| Explicit endpoint pinning | Unit contract proves Docker uses `--host` and Podman uses `--url`, both with the configured Unix endpoint. | Pass |
| No false enabled skip | The tagged MySQL selector with opt-in but no runtime inputs exits 1 and names both required variables. | Pass |
| Target safety | Docker daemon-ID/data-root and Podman target checks remain fail-closed; Docker VM capacity uses only a pre-cached locked-down probe; dbtest passes under `-race`. | Pass |
| Maintainer contract | README and `AGENTS.md` document exact variables, endpoint restrictions, resource ownership, and the single `Config` owner. | Pass |
| Live Docker proof | Exact tagged MySQL proof passed through Colima's explicit Unix socket, including catalog/read/TLS/CDC and cleanup capacity report. | Pass |
| Live Podman proof | No local Podman VM exposed an explicit direct Unix API socket; no global default was read. | Honest gap |

The external PostgreSQL endpoint option remains deliberately unimplemented: it
would no longer prove the managed image/version/settings/extensions/durability
or cleanliness and must receive a separate opt-in contract if needed.
