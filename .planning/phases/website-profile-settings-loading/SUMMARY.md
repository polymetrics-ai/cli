# SUMMARY — website profile settings loading

Status: implementation complete; focused website checks passed with generated-data drift recorded out of scope.

## What changed

- Added a Playwright regression proving the profile visibility checkbox stays disabled while `GET /api/profile` is still loading, then becomes editable after settings arrive.
- Updated `ProfileSettingsDialog` to reset unresolved settings on open, ignore stale profile responses, and disable profile visibility/url/Save controls while settings are loading or a save is in flight.

## What was deliberately excluded

- No release-triggered website dispatch, website-release job, release tag input, PM binary-release coupling, release documentation, deploy docs, connector code, connector generated data, or issue-guard changes were retained from the preserved release branch history.
- Running website data generation revealed current-main Bahmni connector-derived generated drift; those files were reverted because they are outside this focused task boundary.
