package rlm

import (
	"fmt"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors"
)

// scoreFloat coerces a record's _rlm_score to float64, tolerating the float64,
// json.Number, and numeric-string shapes a JSON round-trip can produce.
func scoreFloat(rec connectors.Record) float64 {
	switch v := rec["_rlm_score"].(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return 0
	}
}

// sortScored orders records by _rlm_score descending, then _polymetrics_raw_id
// ascending. It is the single sort used by every backend so the materialized
// OutTable ordering is identical regardless of how scores were produced
// (deterministic spec scoring or an agent-authored program).
func sortScored(records []connectors.Record) {
	sort.SliceStable(records, func(i, j int) bool {
		si, sj := scoreFloat(records[i]), scoreFloat(records[j])
		if si != sj {
			return si > sj
		}
		ri, _ := records[i]["_polymetrics_raw_id"].(string)
		rj, _ := records[j]["_polymetrics_raw_id"].(string)
		return ri < rj
	})
}

// validateOutTable rejects table names that would escape the warehouse directory
// when joined as "<table>.ndjson". Table names must be bare identifiers.
func validateOutTable(table string) error {
	return validateTableName(table, "OutTable")
}

func validateInTable(table string) error {
	return validateTableName(table, "InTable")
}

func validateTableName(table, field string) error {
	if table == "" {
		return fmt.Errorf("rlm: empty %s name", field)
	}
	if table == "." || table == ".." ||
		strings.ContainsAny(table, "/\\") ||
		strings.Contains(table, "..") {
		return fmt.Errorf("rlm: invalid %s name %q (must be a bare name)", field, table)
	}
	return nil
}
