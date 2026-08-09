# Reddit full-parity completion plan

## Scope

Complete the nine rows left by the rebased Reddit checkpoint. Four become
executable write actions and five retain their captain-approved named
exclusions. This phase is deliberately connector-scoped, except for generated
catalog output required by repository checks.

## Delivery mode

This is an inline/manual GSD execution fallback. The canonical Pi worker
runtime is not available as an isolated compatible worker in this task lane,
and the captain supplied the required product decisions in the 2026-08-10
resume order. Required skills are recorded in `CONTEXT.md`.

## Work plan

1. Establish executable contracts before mutations.
   - Add targeted regression tests for Reddit S3-lease uploads, direct
     subreddit image upload, and the human-only vote command.
   - Prove a bulk reverse-ETL plan cannot select `vote`, while a one-record
     connector command is still able to traverse plan, preview, approval, and
     execute with its typed confirmation.
   - **Red:** run the targeted Reddit hook/engine/conformance tests before
     the new declarations and hook exist; the expected missing-action or
     unsupported-hook assertions fail.
   - **Green:** add the declared actions and bounded hook, then rerun the same
     tests until they pass.

2. Make the four executable rows truthful.
   - Add `vote` as a form write action with `batchable: false` and the closed
     `destructive` confirmation kind. The generated `pm reddit vote` command
     stays one-record-only and the normal command lifecycle supplies explicit
     per-invocation approval.
   - Add `/r/{subreddit}/api/upload_sr_img` as a bounded, typed multipart
     write. This endpoint is a direct Reddit multipart request, not an S3
     lease; the official API page distinguishes it from the two S3 endpoints.
   - Add the emoji-asset and widget-image lease actions. Extend the existing
     Reddit hook set with the one narrow `WriteHook` needed to acquire the
     lease from Reddit and send the bounded, approved local file to a
     HTTPS-only Amazon S3 lease host without propagating OAuth headers.
   - Keep data transfer bound to the project directory, regular files,
     declared media types, and an approved content digest.

3. Close ledger and generated-surface drift.
   - Change the four covered rows in `api_surface.json`; retain exactly the
     five captain-approved exclusions with their provider reasons.
   - Correct the three emoji paths to resolve the target subreddit through
     `config.subreddit`, not an undeclared record field.
   - Apply `conformance.skip_dynamic` only to the thirteen documented streams
     whose required provider query parameters cannot be replayed from a
     static fixture.
   - Regenerate `cli_surface.json`, catalog/manual/website generated artifacts,
     and hook registration with the repository generator; never hand-merge
     generated catalog JSON.

4. Verify the shipped surface.
   - Run connector validation, surface sync/reconciliation, focused Go tests,
     conformance, a built `pm` binary's representative commands from stream,
     moderator, live, upload, and vote groups, and the documented generated
     catalog checks.
   - Update phase verification evidence with actual commands and results.
   - Update the parent/sub-issue execution record, open the requested draft
     PR only after local gates pass, and run the required no-mistakes delivery
     workflow rather than manually resolving its findings.

## Acceptance criteria

- `api_surface.json` has 230 dispositioned rows: 225 covered and exactly five
  named captain-approved exclusions.
- No row uses `unsafe_or_disallowed` and no coverage/exclusion is blank.
- `vote` is non-batchable, explicit-confirmation protected, and available only
  through its own single-record connector command lifecycle.
- The three upload endpoints are executable with bounded typed inputs; the two
  lease actions cannot redirect, leak Reddit authentication to S3, follow an
  arbitrary host, or read a file outside the project root.
- Generated command surfaces pass real runtime preflight and website/catalog
  checks; documented restriction text remains visible at the command and in
  connector docs.
