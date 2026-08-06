# REVIEW — issues #3745 and #3746 truthful changefeed discovery

Mode: standard inline GSD `code-review` fallback. The task brief forbids spawning the official
reviewer role, so the worker reviewed the changed descriptor, loader, registry/manifest/definition
projections, CLI JSON serializer, PostgreSQL declaration, and focused tests directly.

## Findings and disposition

| Severity | Finding | Disposition |
| --- | --- | --- |
| Warning | A malformed evidence string could satisfy `artifact_url`, weakening the provider-evidence contract. | Fixed: `Validate` now requires an absolute `http` or `https` URL; a red test was added first. |
| Warning | An `unsupported` descriptor could still declare checkpoint or delivery semantics, which would imply a change-delivery promise. | Fixed: unsupported descriptors now reject executor, checkpoint, and delivery fields; red cases are retained. |

## Final review result

No open Critical, Warning, or Info findings. The resulting contract is fail-closed: only a valid,
implemented descriptor with an exactly matching registered executor projects `cdc: true`; an
unsupported descriptor exposes evidence and reason without an execution or delivery claim.

No dependency, live provider call, credential, redaction/masking path, generic protocol surface,
command-runner edit, connector-fleet classification, #3747 conformance work, #3748 surfacing work,
or #3749 generator enforcement was introduced.
