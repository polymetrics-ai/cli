#!/usr/bin/env python3
"""query-validates-query: validate an RLM result table against the input using
DuckDB SQL assertions only (no LLM). Fail-CLOSED: every check is wrapped in
COALESCE(passed, false) so a NULL-evaluating check fails.

Usage: validate.py <input.ndjson> <output.ndjson>
Prints a JSON verdict {"all_passed": bool, "failures": ["check", ...]} and exits
0 on pass, 1 on failure.

Known ceiling: this proves the result is well-formed and 1:1 with the input
(every input id scored exactly once, scores in [0,1], non-degenerate). It does
NOT prove the scores are *meaningful* — score = hash(id) would pass.
"""
import json
import sys

import duckdb

CHECKS = """
WITH checks AS (
  SELECT 'schema.required' AS chk, (
    (SELECT COUNT(*) FROM (DESCRIBE result) WHERE column_name = '_polymetrics_raw_id') = 1
    AND (SELECT COUNT(*) FROM (DESCRIBE result) WHERE column_name = '_rlm_score') = 1
  ) AS passed
  UNION ALL SELECT 'score.range',
    (SELECT COUNT(*) FROM result
     WHERE _rlm_score < 0 OR _rlm_score > 1 OR _rlm_score IS NULL
        OR isnan(_rlm_score) OR isinf(_rlm_score)) = 0
  UNION ALL SELECT 'result.rowcount', (SELECT COUNT(*) FROM result) > 0
  UNION ALL SELECT 'card.equality',
    (SELECT COUNT(*) FROM input) = (SELECT COUNT(*) FROM result)
  UNION ALL SELECT 'ref.coverage',
    (SELECT COUNT(*) FROM input i
     LEFT JOIN result r ON i._polymetrics_raw_id = r._polymetrics_raw_id
     WHERE r._polymetrics_raw_id IS NULL) = 0
  UNION ALL SELECT 'ref.no_dupes',
    (SELECT COUNT(*) FROM (
       SELECT _polymetrics_raw_id FROM result GROUP BY 1 HAVING COUNT(*) > 1)) = 0
  UNION ALL SELECT 'ref.no_orphans',
    (SELECT COUNT(*) FROM result r
     LEFT JOIN input i ON r._polymetrics_raw_id = i._polymetrics_raw_id
     WHERE i._polymetrics_raw_id IS NULL) = 0
  UNION ALL SELECT 'dist.not_degenerate',
    ((SELECT COUNT(DISTINCT _rlm_score) FROM result) >= 2
     AND (SELECT stddev_pop(_rlm_score) FROM result) > 1e-9)
)
SELECT chk, COALESCE(passed, false) AS passed FROM checks;
"""


def main() -> int:
    if len(sys.argv) != 3:
        print(json.dumps({"all_passed": False, "failures": ["usage"]}))
        return 1
    in_path, out_path = sys.argv[1], sys.argv[2]
    con = duckdb.connect()
    # Input is warehouse-enveloped {"_polymetrics_raw_id":..,"record":{..}};
    # flatten the id alongside the record so checks can reference it.
    con.execute(
        "CREATE TABLE input AS "
        "SELECT _polymetrics_raw_id FROM read_json_auto(?)", [in_path]
    )
    # Output is a flat record per line (the agent's final scored rows).
    con.execute("CREATE TABLE result AS SELECT * FROM read_json_auto(?)", [out_path])
    rows = con.execute(CHECKS).fetchall()
    failures = [name for (name, passed) in rows if not passed]
    verdict = {"all_passed": len(failures) == 0, "failures": failures}
    print(json.dumps(verdict))
    return 0 if not failures else 1


if __name__ == "__main__":
    sys.exit(main())
