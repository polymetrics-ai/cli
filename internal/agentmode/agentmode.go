// Package agentmode produces compact, deterministic record encodings for agent
// consumption. It avoids returning full row dumps when a count, field list,
// bounded sample, or projected NDJSON stream is enough.
package agentmode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"polymetrics.ai/internal/connectors"
)

// Summary is the compact, agent-facing envelope emitted by Summarize.
type Summary struct {
	Kind   string              `json:"kind"`
	Count  int                 `json:"count"`
	Fields []string            `json:"fields"`
	Sample []connectors.Record `json:"sample"`
}

// Summarize encodes records as compact JSON containing the kind, total count,
// sorted field union, and at most sampleN projected sample rows.
func Summarize(kind string, records []connectors.Record, sampleN int) ([]byte, error) {
	fields := collectFields(records)

	n := sampleN
	if n < 0 {
		n = 0
	}
	if n > len(records) {
		n = len(records)
	}

	sample := make([]connectors.Record, 0, n)
	for i := 0; i < n; i++ {
		sample = append(sample, projectRow(records[i], fields))
	}

	payload, err := marshalCompact(Summary{
		Kind:   kind,
		Count:  len(records),
		Fields: fields,
		Sample: sample,
	})
	if err != nil {
		return nil, fmt.Errorf("agentmode: encode summary: %w", err)
	}
	return payload, nil
}

// FieldsProjection returns copies of records containing only requested fields.
// Missing fields are omitted and source records are never mutated.
func FieldsProjection(records []connectors.Record, fields []string) []connectors.Record {
	out := make([]connectors.Record, 0, len(records))
	for _, record := range records {
		out = append(out, projectRow(record, fields))
	}
	return out
}

// EncodeStream writes records as compact NDJSON, one record per line.
func EncodeStream(w io.Writer, records []connectors.Record) error {
	for i, record := range records {
		payload, err := marshalCompact(record)
		if err != nil {
			return fmt.Errorf("agentmode: encode stream record %d: %w", i, err)
		}
		if _, err := w.Write(payload); err != nil {
			return fmt.Errorf("agentmode: write stream record %d: %w", i, err)
		}
		if _, err := w.Write([]byte{'\n'}); err != nil {
			return fmt.Errorf("agentmode: write stream newline %d: %w", i, err)
		}
	}
	return nil
}

// EstimateTokens approximates token usage with the common 4-bytes-per-token
// heuristic. It is intentionally cheap and deterministic.
func EstimateTokens(payload []byte) int {
	return len(payload) / 4
}

// VerboseEnvelope is a pretty-printed, full-record baseline used to compare
// agent-mode output against a fat MCP-style response shape.
func VerboseEnvelope(kind string, records []connectors.Record) ([]byte, error) {
	rows := records
	if rows == nil {
		rows = []connectors.Record{}
	}

	payload, err := json.MarshalIndent(map[string]any{
		"object":          "connector_record_collection",
		"kind":            kind,
		"description":     "Full record collection returned by the connector tool. Every field of every row is included verbatim for the agent to read.",
		"record_count":    len(records),
		"truncated":       false,
		"schema_included": true,
		"fields":          collectFields(records),
		"records":         rows,
		"metadata": map[string]any{
			"encoding":      "json",
			"pretty":        true,
			"source":        "mcp-style-baseline",
			"per_tool":      true,
			"includes_keys": true,
		},
	}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("agentmode: encode verbose envelope: %w", err)
	}
	return payload, nil
}

func collectFields(records []connectors.Record) []string {
	seen := make(map[string]struct{})
	for _, record := range records {
		for key := range record {
			seen[key] = struct{}{}
		}
	}
	fields := make([]string, 0, len(seen))
	for key := range seen {
		fields = append(fields, key)
	}
	sort.Strings(fields)
	return fields
}

func projectRow(record connectors.Record, fields []string) connectors.Record {
	row := make(connectors.Record, len(fields))
	for _, field := range fields {
		if value, ok := record[field]; ok {
			row[field] = value
		}
	}
	return row
}

func marshalCompact(v any) ([]byte, error) {
	payload, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return bytes.TrimRight(payload, "\n"), nil
}
