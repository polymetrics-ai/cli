#!/usr/bin/env python3
"""Refuse to clear a connector whose red was never actually observed.

    python3 check_red_observed.py <connector> [<connector> ...]
    python3 check_red_observed.py --all

**This tool is described in PROGRESS.md's handoff and in `gen_planning_slice.py`'s
contract, and it was never committed.** Three workers were told it "enforces this
and rejects placeholder text"; running it raised `No such file or directory`. It
is written here so the enforcement the handoff claims is real, and so the next
worker inherits a check rather than a claim -- which is the same failure mode as
the `tools/` directory that died with its worktree.

It reads `.planning/phases/<connector>-parity-sweep-r1/RUN-STATE.json` and fails
unless `red_failure` holds output that could only have come from a real `go test`
run against the real bundle:

  * `--- FAIL:` -- the test ran and failed, rather than erroring or being skipped.
  * `<file>_test.go:<line>:` -- at least one assertion fired with a source
    location, which a hand-written summary does not have.
  * a `want <n>` / `want map[...]` clause -- the failure carries the expected
    value, so the number in the test is the number that was actually asserted.
  * no placeholder text (`TODO`, `TBD`, `<paste`, `...output...`, `see above`).

`red_confirmed: true` with anything else in `red_failure` is the exact shape of
"I authored first and wrote the ledger afterwards", which two agents in this
sweep did.
"""

import json
import os
import re
import sys

PHASES = os.path.join(".planning", "phases")

PLACEHOLDERS = (
    "todo", "tbd", "<paste", "...output...", "see above", "xxx",
    "placeholder", "n/a", "same as", "captured separately",
)


def check(connector):
    path = os.path.join(PHASES, "%s-parity-sweep-r1" % connector, "RUN-STATE.json")
    if not os.path.exists(path):
        return ["%s: no RUN-STATE.json at %s" % (connector, path)]

    with open(path) as fh:
        state = json.load(fh)

    failures = []
    red = state.get("red_failure")
    if not state.get("red_confirmed"):
        # Not an error: a connector may honestly record red as unobserved with a
        # stated blocker (help-scout did, when the shared build cache was
        # corrupt). It is only an error to claim red and not have it.
        if red:
            failures.append("%s: red_confirmed is false but red_failure is populated" % connector)
        return failures

    if not isinstance(red, str) or not red.strip():
        return ["%s: red_confirmed is true but red_failure is empty" % connector]

    lowered = red.lower()
    for token in PLACEHOLDERS:
        # Anchored at both ends, because a bare substring search reports a
        # placeholder inside real output: `n/a` matches "expression/analyse",
        # which is a documented Jira endpoint, and the tool would then reject a
        # genuinely observed failure -- a check that cries wolf gets disabled.
        if re.search(r"(?<![\w/])" + re.escape(token) + r"(?![\w])", lowered):
            failures.append("%s: red_failure contains placeholder text %r" % (connector, token))

    if "--- FAIL:" not in red:
        failures.append("%s: red_failure has no '--- FAIL:' line; the test did not run and fail"
                        % connector)
    if not re.search(r"\w+_test\.go:\d+:", red):
        failures.append("%s: red_failure has no '<file>_test.go:<line>:' assertion location"
                        % connector)
    if not re.search(r"\bwant [\w\[\-]", red):
        failures.append("%s: red_failure carries no 'want <expected>' clause, so it does not show "
                        "which number was asserted" % connector)
    return failures


def main():
    argv = sys.argv[1:]
    if not argv:
        raise SystemExit(__doc__)
    if argv == ["--all"]:
        argv = sorted(
            d[: -len("-parity-sweep-r1")]
            for d in os.listdir(PHASES)
            if d.endswith("-parity-sweep-r1")
        )

    failures = []
    for connector in argv:
        found = check(connector)
        failures.extend(found)
        print("%-18s %s" % (connector, "FAIL" if found else "red observed"))
    for line in failures:
        print("  " + line, file=sys.stderr)
    raise SystemExit(1 if failures else 0)


if __name__ == "__main__":
    main()
