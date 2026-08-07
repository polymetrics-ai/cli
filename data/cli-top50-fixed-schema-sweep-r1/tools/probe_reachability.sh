#!/bin/bash
# Reachability probe for one command path, e.g.  ./probe_reachability.sh "orgs get"
#
# A namespace miss renders `pm <connector>` group help and STILL EXITS 0 -- that
# is the documented namespace behaviour, and it makes the exit code worthless as
# proof. The rendered NAME line is the only evidence the router resolved this
# exact command path. Prints nothing on success; one FAIL line otherwise.
BIN=${PM_BIN:-/tmp/pm-gh}
CONNECTOR=${PM_CONNECTOR:-github}

out=$("$BIN" "$CONNECTOR" $1 --help 2>&1)
rc=$?
if [ $rc -ne 0 ]; then echo "FAIL(exit $rc): $1"; exit 0; fi
if ! printf '%s' "$out" | grep -qF "pm $CONNECTOR $1 - "; then
  echo "FAIL(unrouted): $1 :: $(printf '%s' "$out" | sed -n '2p')"
fi
