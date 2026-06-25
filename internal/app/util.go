package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"polymetrics/internal/connectors"
)

func cloneStringMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneRecord(in connectors.Record) connectors.Record {
	out := make(connectors.Record, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneRecords(in []connectors.Record) []connectors.Record {
	out := make([]connectors.Record, 0, len(in))
	for _, record := range in {
		out = append(out, cloneRecord(record))
	}
	return out
}

func mapReverseRecords(records []connectors.Record, mappings map[string]string, planID string) []connectors.Record {
	mapped := make([]connectors.Record, 0, len(records))
	for _, record := range records {
		out := connectors.Record{}
		for source, dest := range mappings {
			out[dest] = record[source]
		}
		if planID != "" {
			out["_polymetrics_reverse_plan_id"] = planID
		}
		mapped = append(mapped, out)
	}
	return mapped
}

func hashString(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func hashJSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return hashString(string(b)), nil
}

func parseSelectAll(sql string) (string, int, error) {
	fields := strings.Fields(strings.TrimSpace(strings.TrimSuffix(sql, ";")))
	if len(fields) < 4 {
		return "", 0, errors.New("only SELECT * FROM <table> [LIMIT n] is supported in the MVP")
	}
	if !strings.EqualFold(fields[0], "select") || fields[1] != "*" || !strings.EqualFold(fields[2], "from") {
		return "", 0, errors.New("only SELECT * FROM <table> [LIMIT n] is supported in the MVP")
	}
	table := fields[3]
	limit := 100
	if len(fields) > 4 {
		if len(fields) != 6 || !strings.EqualFold(fields[4], "limit") {
			return "", 0, errors.New("only SELECT * FROM <table> [LIMIT n] is supported in the MVP")
		}
		n, err := strconv.Atoi(fields[5])
		if err != nil || n <= 0 {
			return "", 0, fmt.Errorf("invalid limit %q", fields[5])
		}
		limit = n
	}
	return table, limit, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func findCatalogStream(catalog connectors.Catalog, name string) (connectors.Stream, bool) {
	for _, stream := range catalog.Streams {
		if stream.Name == name {
			return stream, true
		}
	}
	return connectors.Stream{}, false
}
