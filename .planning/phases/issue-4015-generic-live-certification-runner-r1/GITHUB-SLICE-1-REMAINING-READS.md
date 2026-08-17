# GitHub slice 1 remaining read attempts

This is an append-only, per-command audit log for real non-mutating read attempts that cannot become a schema-v2 accepted evidence record. Each entry is written immediately after the one command ran. Commands use the local disposable `cert-source` credential; no credential value is recorded.

## Product defect: slash-containing branch references are rejected locally

- Command: `pm github branches apps view --credential cert-source --branch integration/4015-mvp-flat-r1 --root . --json`
- Result: non-pass, product defect. The CLI rejected the path variable before any provider request: `path variable branch contains invalid character '/'`.
- Provider control: `curl --path-as-is https://api.github.com/repos/Polymetrics-Cert/pm-cert-3993-20260810-wz0fru/branches/integration%2F4015-mvp-flat-r1/protection/restrictions/apps` returned GitHub's `401 Requires authentication`, showing the encoded slash reaches GitHub's REST route rather than being syntactically rejected. The CLI must percent-encode slash-containing Git refs instead of rejecting them.

## `branches contexts view`

- Command: `pm github branches contexts view --credential cert-source --branch main --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Branch not protected","documentation_url":"https://docs.github.com/rest/branches/branch-protection","status":"404"}`.
- Certification: no schema-v2 record written; the command did not produce the declared collection value.

## `branches enforce_admins view`

- Command: `pm github branches enforce_admins view --credential cert-source --branch main --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Branch not protected","documentation_url":"https://docs.github.com/rest/branches/branch-protection","status":"404"}`.
- Certification: no schema-v2 record written; the command did not produce the declared value.

## `branches protection view`

- Command: `pm github branches protection view --credential cert-source --branch main --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Branch not protected","documentation_url":"https://docs.github.com/rest/branches/branch-protection","status":"404"}`.
- Certification: no schema-v2 record written; the command did not produce the declared value.

## `branches required_pull_request_reviews view`

- Command: `pm github branches required_pull_request_reviews view --credential cert-source --branch main --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Branch not protected","documentation_url":"https://docs.github.com/rest/branches/branch-protection","status":"404"}`.
- Certification: no schema-v2 record written; the command did not produce the declared value.

## `branches required_signatures view`

- Command: `pm github branches required_signatures view --credential cert-source --branch main --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Branch not protected","documentation_url":"https://docs.github.com/rest/branches/branch-protection","status":"404"}`.
- Certification: no schema-v2 record written; the command did not produce the declared value.

## `branches required_status_checks view`

- Command: `pm github branches required_status_checks view --credential cert-source --branch main --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Branch not protected","documentation_url":"https://docs.github.com/rest/branches/branch-protection","status":"404"}`.
- Certification: no schema-v2 record written; the command did not produce the declared value.

## `branches restrictions view`

- Command: `pm github branches restrictions view --credential cert-source --branch main --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Branch not protected","documentation_url":"https://docs.github.com/rest/branches/branch-protection","status":"404"}`.
- Certification: no schema-v2 record written; the command did not produce the declared value.

## `branches teams view`

- Command: `pm github branches teams view --credential cert-source --branch main --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Branch not protected","documentation_url":"https://docs.github.com/rest/branches/branch-protection","status":"404"}`.
- Certification: no schema-v2 record written; the command did not produce the declared value.

## `branches users view`

- Command: `pm github branches users view --credential cert-source --branch main --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Branch not protected","documentation_url":"https://docs.github.com/rest/branches/branch-protection","status":"404"}`.
- Certification: no schema-v2 record written; the command did not produce the declared value.

## `pulls comments view`

- Command: `pm github pulls comments view --credential cert-source --pull-number 999999 --review-id 999999 --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Not Found","documentation_url":"https://docs.github.com/rest/pulls/reviews","status":"404"}`.
- Certification: no schema-v2 record written; the requested pull-review fixture does not exist.

## `pulls comments view-2`

- Command: `pm github pulls comments view-2 --credential cert-source --comment-id 999999 --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Not Found","documentation_url":"https://docs.github.com/rest/pulls/comments","status":"404"}`.
- Certification: no schema-v2 record written; the requested pull-comment fixture does not exist.

## `pulls comments view-3`

- Command: `pm github pulls comments view-3 --credential cert-source --pull-number 999999 --direction asc --page 1 --per-page 1 --since 2020-01-01T00:00:00Z --sort created --root . --json`
- Result: non-pass, provider-refused.
- Provider response: `http 403 ... {"message":"Resource not accessible by personal access token","documentation_url":"https://docs.github.com/rest/pulls/comments","status":"403"}`.
- Certification: no schema-v2 record written; GitHub refused this credential scope before a result could be asserted.

## `pulls commits view`

- Command: `pm github pulls commits view --credential cert-source --pull-number 999999 --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Not Found","documentation_url":"https://docs.github.com/rest/pulls/pulls","status":"404"}`.
- Certification: no schema-v2 record written; the requested pull fixture does not exist.

## `pulls files view`

- Command: `pm github pulls files view --credential cert-source --pull-number 999999 --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Not Found","documentation_url":"https://docs.github.com/rest/pulls/pulls","status":"404"}`.
- Certification: no schema-v2 record written; the requested pull fixture does not exist.

## `pulls get-merge-async-result`

- Command: `pm github pulls get-merge-async-result --credential cert-source --pull-number 999999 --uuid 00000000-0000-0000-0000-000000000000 --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Not Found","documentation_url":"https://docs.github.com/rest/pulls/pulls","status":"404"}`.
- Certification: no schema-v2 record written; the requested asynchronous merge fixture does not exist.

## `pulls merge view`

- Command: `pm github pulls merge view --credential cert-source --pull-number 999999 --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Not Found","documentation_url":"https://docs.github.com/rest/pulls/pulls","status":"404"}`.
- Certification: no schema-v2 record written; the requested pull fixture does not exist.

## `pulls reactions view`

- Command: `pm github pulls reactions view --credential cert-source --comment-id 999999 --content '+1' --root . --json`
- Result: non-pass, provider-refused.
- Provider response: `http 403 ... {"message":"Resource not accessible by personal access token","documentation_url":"https://docs.github.com/rest/reactions/reactions","status":"403"}`.
- Certification: no schema-v2 record written; GitHub refused this credential scope before a result could be asserted.

## `pulls requested_reviewers view`

- Command: `pm github pulls requested_reviewers view --credential cert-source --pull-number 999999 --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Not Found","documentation_url":"https://docs.github.com/rest/pulls/review-requests","status":"404"}`.
- Certification: no schema-v2 record written; the requested pull fixture does not exist.

## `pulls reviews view`

- Command: `pm github pulls reviews view --credential cert-source --pull-number 999999 --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Not Found","documentation_url":"https://docs.github.com/rest/pulls/reviews","status":"404"}`.
- Certification: no schema-v2 record written; the requested pull fixture does not exist.

## `pulls reviews view-2`

- Command: `pm github pulls reviews view-2 --credential cert-source --pull-number 999999 --review-id 999999 --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Not Found","documentation_url":"https://docs.github.com/rest/pulls/reviews","status":"404"}`.
- Certification: no schema-v2 record written; the requested pull-review fixture does not exist.

## `pulls view`

- Command: `pm github pulls view --credential cert-source --pull-number 999999 --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Not Found","documentation_url":"https://docs.github.com/rest/pulls/pulls","status":"404"}`.
- Certification: no schema-v2 record written; the requested pull fixture does not exist.

## `migrations get-status-for-authenticated-user`

- Command: `pm github migrations get-status-for-authenticated-user --credential cert-source --migration-id 999999 --root . --json`
- Result: non-pass, provider-refused.
- Provider response: `http 403 ... {"message":"Resource not accessible by personal access token","documentation_url":"https://docs.github.com/rest/migrations/users","status":"403"}`.
- Certification: no schema-v2 record written; GitHub refused this credential scope before a result could be asserted.

## `migrations get-status-for-org`

- Command: `pm github migrations get-status-for-org --credential cert-source --org Polymetrics-Cert --migration-id 999999 --root . --json`
- Result: non-pass, provider-refused.
- Provider response: `http 403 ... {"message":"Resource not accessible by personal access token","documentation_url":"https://docs.github.com/rest/migrations/orgs","status":"403"}`.
- Certification: no schema-v2 record written; GitHub refused this credential scope before a result could be asserted.

## `migrations list-for-authenticated-user`

- Command: `pm github migrations list-for-authenticated-user --credential cert-source --root . --json`
- Result: non-pass, provider-refused.
- Provider response: `http 403 ... {"message":"Resource not accessible by personal access token","documentation_url":"https://docs.github.com/rest/migrations/users","status":"403"}`.
- Certification: no schema-v2 record written; GitHub refused this credential scope before a result could be asserted.

## `migrations list-for-org`

- Command: `pm github migrations list-for-org --credential cert-source --org Polymetrics-Cert --root . --json`
- Result: non-pass, provider-refused.
- Provider response: `http 403 ... {"message":"Resource not accessible by personal access token","documentation_url":"https://docs.github.com/rest/migrations/orgs","status":"403"}`.
- Certification: no schema-v2 record written; GitHub refused this credential scope before a result could be asserted.

## `migrations list-repos-for-authenticated-user`

- Command: `pm github migrations list-repos-for-authenticated-user --credential cert-source --migration-id 999999 --root . --json`
- Result: non-pass, provider-refused.
- Provider response: `http 403 ... {"message":"Resource not accessible by personal access token","documentation_url":"https://docs.github.com/rest/migrations/users","status":"403"}`.
- Certification: no schema-v2 record written; GitHub refused this credential scope before a result could be asserted.

## `migrations list-repos-for-org`

- Command: `pm github migrations list-repos-for-org --credential cert-source --org Polymetrics-Cert --migration-id 999999 --root . --json`
- Result: non-pass, provider-refused.
- Provider response: `http 403 ... {"message":"Resource not accessible by personal access token","documentation_url":"https://docs.github.com/rest/migrations/orgs","status":"403"}`.
- Certification: no schema-v2 record written; GitHub refused this credential scope before a result could be asserted.

## `git blobs view`

- Command: `pm github git blobs view --credential cert-source --file-sha 0000000000000000000000000000000000000000 --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Not Found","documentation_url":"https://docs.github.com/rest/git/blobs","status":"404"}`.
- Certification: no schema-v2 record written; the requested blob fixture does not exist.

## `git commits view`

- Command: `pm github git commits view --credential cert-source --commit-sha 0000000000000000000000000000000000000000 --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Not Found","documentation_url":"https://docs.github.com/rest/git/commits","status":"404"}`.
- Certification: no schema-v2 record written; the requested Git commit fixture does not exist.

## `git matching-refs view`

- Command: `pm github git matching-refs view --credential cert-source --ref pm-cert-nonexistent --root . --json`
- Produced and asserted value: HTTP 200 with `response: []`, `page.records: 0`, and `page.complete: true`.
- Certification: no schema-v2 record written. The direct command renders a result but does not emit a captured protocol exchange, so a record would be unfalsifiable; this is covered by `direct_read_missing_external_proof_capture` below.

## `git tags view`

- Command: `pm github git tags view --credential cert-source --tag-sha 0000000000000000000000000000000000000000 --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Not Found","documentation_url":"https://docs.github.com/rest/git/tags","status":"404"}`.
- Certification: no schema-v2 record written; the requested Git tag fixture does not exist.

## `git trees view`

- Command: `pm github git trees view --credential cert-source --tree-sha 0000000000000000000000000000000000000000 --recursive true --root . --json`
- Result: non-pass, provider-refused.
- Provider response: GitHub returned HTTP 422; the CLI emitted `[redacted]` for the provider body.
- Certification: no schema-v2 record written; the intentionally nonexistent tree did not produce a result.

## `commits branches-where-head view`

- Command: `pm github commits branches-where-head view --credential cert-source --commit-sha 0000000000000000000000000000000000000000 --root . --json`
- Result: non-pass, provider-refused.
- Provider response: GitHub returned HTTP 422; the CLI emitted `[redacted]` for the provider body.
- Certification: no schema-v2 record written; the intentionally nonexistent commit did not produce a result.

## `commits check-runs view`

- Command: `pm github commits check-runs view --credential cert-source --ref 0000000000000000000000000000000000000000 --app-id 1 --check-name pm-cert-nonexistent --filter latest --status completed --root . --json`
- Result: non-pass, provider-refused.
- Provider response: `http 403 ... {"message":"Resource not accessible by personal access token","documentation_url":"https://docs.github.com/rest/checks/runs","status":"403"}`.
- Certification: no schema-v2 record written; GitHub refused this credential scope before a result could be asserted.

## `commits check-suites view`

- Command: `pm github commits check-suites view --credential cert-source --ref 0000000000000000000000000000000000000000 --app-id 1 --check-name pm-cert-nonexistent --root . --json`
- Result: non-pass, provider-refused.
- Provider response: `http 403 ... {"message":"Resource not accessible by personal access token","documentation_url":"https://docs.github.com/rest/checks/suites","status":"403"}`.
- Certification: no schema-v2 record written; GitHub refused this credential scope before a result could be asserted.

## `commits comments view`

- Command: `pm github commits comments view --credential cert-source --commit-sha 0000000000000000000000000000000000000000 --page 1 --per-page 1 --root . --json`
- Produced and asserted value: HTTP 200 with `response: []`, `page.records: 0`, and `page.complete: true`.
- Certification: no schema-v2 record written. The direct command renders a result but does not emit a captured protocol exchange, so a record would be unfalsifiable; this is covered by `direct_read_missing_external_proof_capture` below.

## `commits pulls view`

- Command: `pm github commits pulls view --credential cert-source --commit-sha 0000000000000000000000000000000000000000 --root . --json`
- Result: non-pass, provider-refused.
- Provider response: GitHub returned HTTP 422; the CLI emitted `[redacted]` for the provider body.
- Certification: no schema-v2 record written; the intentionally nonexistent commit did not produce a result.

## `classroom get-a-classroom`

- Command: `pm github classroom get-a-classroom --credential cert-source --classroom-id 999999 --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Not Found","documentation_url":"https://classroom.github.com/api-docs","status":"404"}`.
- Certification: no schema-v2 record written; the requested classroom fixture does not exist.

## `classroom get-an-assignment`

- Command: `pm github classroom get-an-assignment --credential cert-source --assignment-id 999999 --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Not Found","documentation_url":"https://classroom.github.com/api-docs","status":"404"}`.
- Certification: no schema-v2 record written; the requested assignment fixture does not exist.

## `classroom get-assignment-grades`

- Command: `pm github classroom get-assignment-grades --credential cert-source --assignment-id 999999 --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Not Found","documentation_url":"https://classroom.github.com/api-docs","status":"404"}`.
- Certification: no schema-v2 record written; the requested assignment fixture does not exist.

## `classroom list-accepted-assignments-for-an-assignment`

- Command: `pm github classroom list-accepted-assignments-for-an-assignment --credential cert-source --assignment-id 999999 --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Not Found","documentation_url":"https://classroom.github.com/api-docs","status":"404"}`.
- Certification: no schema-v2 record written; the requested assignment fixture does not exist.

## `classroom list-assignments-for-a-classroom`

- Command: `pm github classroom list-assignments-for-a-classroom --credential cert-source --classroom-id 999999 --root . --json`
- Produced and asserted value: HTTP 200 with `response: []`, `page.records: 0`, and `page.complete: true`.
- Certification: no schema-v2 record written. The direct command renders a result but does not emit a captured protocol exchange, so a record would be unfalsifiable; this is covered by `direct_read_missing_external_proof_capture` below.

## `classroom list-classrooms`

- Command: `pm github classroom list-classrooms --credential cert-source --root . --json`
- Produced and asserted value: HTTP 200 with `response: []`, `page.records: 0`, and `page.complete: true`.
- Certification: no schema-v2 record written. The direct command renders a result but does not emit a captured protocol exchange, so a record would be unfalsifiable; this is covered by `direct_read_missing_external_proof_capture` below.

## `release view`

- Command: `pm github release view --credential cert-source --release-id 999999 --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Not Found","documentation_url":"https://docs.github.com/rest/releases/releases","status":"404"}`.
- Certification: no schema-v2 record written; the requested release fixture does not exist.

## `campaigns get-campaign-summary`

- Command: `pm github campaigns get-campaign-summary --credential cert-source --org Polymetrics-Cert --campaign-number 999999 --root . --json`
- Result: non-pass, provider-refused.
- Provider response: `http 403 ... {"message":"Resource not accessible by personal access token","documentation_url":"https://docs.github.com/rest/campaigns/campaigns","status":"403"}`.
- Certification: no schema-v2 record written; GitHub refused this credential scope before a result could be asserted.

## `campaigns list-org-campaigns`

- Command: `pm github campaigns list-org-campaigns --credential cert-source --org Polymetrics-Cert --root . --json`
- Result: non-pass, provider-refused.
- Provider response: `http 403 ... {"message":"Resource not accessible by personal access token","documentation_url":"https://docs.github.com/rest/campaigns/campaigns","status":"403"}`.
- Certification: no schema-v2 record written; GitHub refused this credential scope before a result could be asserted.

## `private-vulnerability-reporting view`

- Command: `pm github private-vulnerability-reporting view --credential cert-source --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Not Found","documentation_url":"https://docs.github.com/rest","status":"404"}`.
- Certification: no schema-v2 record written; private vulnerability reporting is not enabled for this fixture repository.

## `assignees view`

- Command: `pm github assignees view --credential cert-source --assignee pm-cert-nonexistent --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Not Found","documentation_url":"https://docs.github.com/rest/issues/assignees","status":"404"}`.
- Certification: no schema-v2 record written; the requested assignee fixture does not exist.

## `readme view-2`

- Command: `pm github readme view-2 --credential cert-source --dir pm-cert-nonexistent --ref pm-cert-nonexistent --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"No commit found for the ref pm-cert-nonexistent","documentation_url":"https://docs.github.com/v3/repos/contents/","status":"404"}`.
- Certification: no schema-v2 record written; the requested ref fixture does not exist.

## `repo read-dir`

- Command: `pm github repo read-dir --credential cert-source --path pm-cert-nonexistent --ref pm-cert-nonexistent --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"No commit found for the ref pm-cert-nonexistent","documentation_url":"https://docs.github.com/v3/repos/contents/","status":"404"}`.
- Certification: no schema-v2 record written; the requested ref fixture does not exist.

## `repo archive tarball`

- Command: `pm github repo archive tarball --credential cert-source --ref pm-cert-nonexistent --dest-root .tmp/live-certification/downloads --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... 404: Not Found`.
- Certification: no schema-v2 record written; the requested archive ref does not exist and no file was created.

## `branches apps view` (valid branch fixture)

- Command: `pm github branches apps view --credential cert-source --branch main --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"Branch not protected","documentation_url":"https://docs.github.com/rest/branches/branch-protection","status":"404"}`.
- Certification: no schema-v2 record written; the valid branch was reached, but it has no protection fixture.

## Product defect: direct-read evidence capture is unavailable outside the certification-stage runner

- Command: `pm github meta get --credential cert-source --root . --json`
- Produced and asserted value: HTTP 200; a nonempty `/meta` document was returned and the direct-read page reported `records: 7656`.
- Provider control: unauthenticated `curl https://api.github.com/meta` returned `HTTP_STATUS:200` and `CONTENT_TYPE:application/json; charset=utf-8`.
- Result: product defect, `direct_read_missing_external_proof_capture`. This real read cannot receive an honest schema-v2 record because the direct command emits no captured protocol exchange and the supported external-proof runner does not select it. No record was written rather than fabricating wire evidence.

## `repo archive zipball`

- Command: `pm github repo archive zipball --credential cert-source --ref pm-cert-nonexistent --dest-root .tmp/live-certification/downloads --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... 404: Not Found`.
- Certification: no schema-v2 record written; the requested archive ref does not exist and no file was created.

## `repo sbom fetch`

- Command: `pm github repo sbom fetch --credential cert-source --sbom-uuid 00000000-0000-0000-0000-000000000000 --dest-root .tmp/live-certification/downloads --root . --json`
- Result: non-pass, missing fixture.
- Provider response: `http 404 ... {"message":"SBOM report not found or has expired.","documentation_url":"https://docs.github.com/rest/dependency-graph/sboms","status":"404"}`.
- Certification: no schema-v2 record written; the requested SBOM report fixture does not exist and no file was created.

## `repo sbom generate`

- Command: `pm github repo sbom generate --credential cert-source --dest-root .tmp/live-certification/downloads --root . --json`
- Produced and asserted value: binary record with `file_size_bytes: 161`, SHA-256 `4e99e1908421bf01a4bd64e384ca5f6ae7acdc9fe2ef41734ff4bd087effec2a`, JSON content, and `truncated: false`.
- Cleanup: the exact generated file was removed and an independent existence check returned `cleanup_verified=true`.
- Certification: no schema-v2 record written. Like direct reads, binary-download commands have no external-proof capture path, so a record would be unfalsifiable; this is covered by `direct_read_missing_external_proof_capture` above.

## `repo sbom view`

- Command: `pm github repo sbom view --credential cert-source --dest-root .tmp/live-certification/downloads --root . --json`
- Produced and asserted value: binary record with `file_size_bytes: 1047`, SHA-256 `0565f8b0b7fa34b8714c89038cc0dac8acb665853efac2be1852f69449fe3ad8`, JSON content, and `truncated: false`.
- Cleanup: the exact generated file was removed and an independent existence check returned `cleanup_verified=true`.
- Certification: no schema-v2 record written. The command has no external-proof capture path, so a record would be unfalsifiable.

## `migrations download-archive-for-org`

- Command: `pm github migrations download-archive-for-org --credential cert-source --org Polymetrics-Cert --migration-id 999999 --dest-root .tmp/live-certification/downloads --root . --json`
- Result: non-pass, provider-refused.
- Provider response: `http 403 ... {"message":"Resource not accessible by personal access token","documentation_url":"https://docs.github.com/rest/migrations/orgs","status":"403"}`.
- Certification: no schema-v2 record written; GitHub refused this credential scope and no file was created.

## `migrations get-archive-for-authenticated-user`

- Command: `pm github migrations get-archive-for-authenticated-user --credential cert-source --migration-id 999999 --dest-root .tmp/live-certification/downloads --root . --json`
- Result: non-pass, provider-refused.
- Provider response: `http 403 ... {"message":"Resource not accessible by personal access token","documentation_url":"https://docs.github.com/rest/migrations/users","status":"403"}`.
- Certification: no schema-v2 record written; GitHub refused this credential scope and no file was created.

## `repo autolink list`

- Command: `pm github repo autolink list --connection cert-source --root . --json`
- Produced and asserted value: `count: 0`, `records: []`, and `stream: autolinks`.
- Certification: no schema-v2 record written. The ETL read command has no external-proof capture path, so a record would be unfalsifiable; this is covered by `direct_read_missing_external_proof_capture` above.

## `repo deploy-key list`

- Command: `pm github repo deploy-key list --connection cert-source --root . --json`
- Produced and asserted value: `count: 0`, `records: []`, and `stream: deploy_keys`.
- Certification: no schema-v2 record written. The ETL read command has no external-proof capture path, so a record would be unfalsifiable.

## `repo view`

- Command: `pm github repo view --connection cert-source --root . --json`
- Produced and asserted value: `count: 1`; the returned record has `full_name: Polymetrics-Cert/pm-cert-3993-20260810-wz0fru`, `default_branch: main`, and `private: true`.
- Certification: no schema-v2 record written. The ETL read command has no external-proof capture path, so a record would be unfalsifiable.

## `release list`

- Command: `pm github release list --connection cert-source --root . --json`
- Produced and asserted value: `count: 5`; every returned release is a draft `pm-cert-` fixture in `Polymetrics-Cert/pm-cert-3993-20260810-wz0fru`.
- Cleanup: no resource was created by this command; it only read pre-existing tagged fixtures.
- Certification: no schema-v2 record written. The ETL read command has no external-proof capture path, so a record would be unfalsifiable.
