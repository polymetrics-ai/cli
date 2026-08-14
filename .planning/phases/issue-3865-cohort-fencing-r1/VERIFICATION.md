# VERIFICATION — issue #3865 verified-auth cohort fencing

Status: focused implementation verification passed; broader local gates and review remain.

- [x] Typed verified-auth result is the sole fence trigger.
- [x] A verified fence cancels same-cohort active contexts and rejects all later admissions before a fake sender increments.
- [x] An unrelated opaque cohort remains admitted.
- [x] Verified repair/test increases the epoch, refuses stale members, and permits the new healthy epoch.
- [x] Restart/race proof observes zero post-fence admissions and sends under `-race`.
- [ ] Formatter, vet/build, and applicable individual repository gates pass.
- [ ] Inline GSD verify-work and code-review evidence is recorded.
- [ ] PR targets `integration/4015-mvp-flat-r1`; the GitHub API reports that exact base.
