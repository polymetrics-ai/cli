# VERIFICATION — issue #3867 rate-limit parking and automatic resumption

Status: planned; implementation has not started.

- [ ] Typed rate-limit classification persists a truthful park record.
- [ ] The persisted record survives reconstruction and receives its committed checkpoint unchanged.
- [ ] No same-scope send occurs before reset; unrelated scope admission remains available.
- [ ] Automatic resume occurs once at/after reset and never replays acknowledged apply work.
- [ ] Cancellation, duplicate observation, failed callback, and scheduler restart behavior are observed.
- [ ] Park/resume events assert exact reason/source and reset timestamp.
- [ ] Focused tests, required race check, targeted static/build checks, and individual repository gates pass.
- [ ] Inline GSD verify-work/code-review evidence is recorded.
- [ ] PR targets `integration/4015-mvp-flat-r1`; GitHub API reports that exact base.
