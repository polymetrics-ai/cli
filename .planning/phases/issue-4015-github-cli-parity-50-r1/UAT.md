# UAT — Issue #4015 GitHub declared-command parity

GSD `verify-work` was performed through the documented inline/manual fallback because the canonical
single-worker delivery contract forbids spawning a verifier role. These deliverables are binary,
provider, and artifact assertions; no visual or subjective sign-off is required.

| ID | Deliverable | Automated or directly observed evidence | Result |
| --- | --- | --- | --- |
| D1 | Every supplied declaration has one verdict | The external 27 + 23 inventory joins to 50 unique CLI rows with no missing path; `TestGitHubDeclaredParityVerdicts` pins the same exact inventory. | pass |
| D2 | Every promotion is a fixed executable command | 25 promoted rows each have exactly one `api_surface`; the real `commandrunner.Preflight` accepts all 25, and the binary reachability report proves all 1,546 implemented commands reach the shared missing-credential boundary. | pass |
| D3 | Every retained command has concrete evidence | 25 rows retain a declaration and name their provider, composite, local-executor, cryptographic, interactive, or safety boundary; the parity test asserts an evidence fragment for each. | pass |
| D4 | Binary downloads preserve the bounded transport contract | Unit tests cover declaration-owned `Accept`, header-injection refusal, and redirect policy. Live `pr diff` returned a diff; live release download matched the exact 51-byte fixture through a credential-stripped allowed redirect. | pass |
| D5 | Runtime help, manuals, skills, and website stay aligned | `pm help github`, bare `pm github`, exact command help, tracked manuals/skill generation, website script tests, lint, typecheck, and production build pass. | pass |
| D6 | Provider certification leaves no residue | Variable, autolink, issue, workflow file, temporary PR branch, release, and release tag have independent 404 proofs; the PR fixture is closed/non-draft, no codespace exists, and all sealed local certification projects/download roots were deleted. | pass |

Verdict: **verified** — 25 commands are newly executable and 25 remain honestly declared with exact
provider/runtime evidence, for a complete 50-command accounting.
