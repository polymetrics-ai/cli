# Reddit full-parity completion TDD ledger

| ID | Requirement | Red evidence | Green evidence |
| --- | --- | --- | --- |
| R1 | `POST /api/vote` has a typed human-only individual command, not a bulk path. | The targeted Reddit action test fails before the action exists; the shared `batchable:false` foundation test fails before the four Reddit actions are allowlisted. | The shared foundation proves bulk refusal for every `batchable:false` action; the Reddit binary smoke proves its one-record connector-command plan and typed `--confirm destructive` seal. |
| R2 | Emoji asset and widget image upload acquire a Reddit lease then make one bounded S3 multipart POST. | Add hook tests for the two missing action names and fixture lease response; they fail before the hook handles them. | Test observes the lease request, the S3 multipart form fields/file, no Authorization header on S3, HTTPS/S3-host rejection, and no request after a rejected source. |
| R3 | Subreddit image upload is a direct typed multipart write. | Add a declaration/execute fixture test; it fails while the endpoint remains excluded. | Test proves accepted image form fields, project-root confinement, media/size validation, and command preflight. |
| R4 | The ledger has only the five captain-approved exclusions and exposes their reasons. | Count assertion fails with the pre-change 221/9 checkpoint. | Count assertion passes at 225 covered / 5 excluded, including the five named provider constraints. |
| R5 | Existing generated emoji commands resolve target scope from config. | Existing command fixture/preflight exposes the undeclared record `subreddit` field. | Regenerated command surface no longer requires that record field and config interpolation resolves correctly. |
| R6 | Dynamic conformance does not pretend fixture replay can satisfy mandatory query values. | Conformance reports required query parameter absence for the documented dynamic streams. | Per-stream `skip_dynamic` declarations produce only intentional skips and static conformance remains clean. |

## Test commands

```text
go test -timeout 20m ./internal/connectors/hooks/reddit ./internal/connectors/engine ./internal/connectors/conformance ./internal/connectors/commandrunner ./internal/app
go run ./cmd/connectorgen validate internal/connectors/defs/reddit
go run ./cmd/connectorgen surface-sync internal/connectors/defs/reddit --check
go run ./cmd/connectorgen surface-reconcile internal/connectors/defs/reddit --check
```

The initial Red step is recorded before the production declaration/hook patch;
the corresponding command output is added to `VERIFICATION.md` after each
Green step.
