package intruder

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Intruder API resource path (relative
// to base_url) it reads from, and the record mapper that flattens its objects.
// substream marks the occurrences stream, which is sliced by parent issue id and
// reaches its records via the parameterized path /issues/{id}/occurrences.
type streamEndpoint struct {
	// resource is the Intruder list endpoint path segment (e.g. "targets").
	resource string
	// mapRecord flattens a raw Intruder object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
	// substream is true for occurrences, which is read per parent issue id.
	substream bool
}

// intruderStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in intruderStreams; the read
// path is fully data-driven from this table.
var intruderStreamEndpoints = map[string]streamEndpoint{
	"issues":      {resource: "issues", mapRecord: intruderIssueRecord},
	"scans":       {resource: "scans", mapRecord: intruderScanRecord},
	"targets":     {resource: "targets", mapRecord: intruderTargetRecord},
	"occurrences": {resource: "issues", mapRecord: intruderOccurrenceRecord, substream: true},
}

// intruderStreams returns the connector's published stream catalog. Every
// Intruder object exposes an integer id; the API is full-refresh only (no
// incremental cursor), so CursorFields are empty across the board.
func intruderStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "issues",
			Description: "Current security issues detected by Intruder.",
			PrimaryKey:  []string{"id"},
			Fields:      intruderIssueFields(),
		},
		{
			Name:        "occurrences",
			Description: "Per-issue occurrences (affected target/port) for each Intruder issue.",
			PrimaryKey:  []string{"id"},
			Fields:      intruderOccurrenceFields(),
		},
		{
			Name:        "scans",
			Description: "Intruder vulnerability scans.",
			PrimaryKey:  []string{"id"},
			Fields:      intruderScanFields(),
		},
		{
			Name:        "targets",
			Description: "Targets monitored by Intruder.",
			PrimaryKey:  []string{"id"},
			Fields:      intruderTargetFields(),
		},
	}
}

func intruderIssueFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "title", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "severity", Type: "string"},
		{Name: "remediation", Type: "string"},
		{Name: "occurrences", Type: "string"},
		{Name: "snoozed", Type: "boolean"},
		{Name: "snooze_reason", Type: "string"},
		{Name: "snooze_until", Type: "string"},
	}
}

func intruderOccurrenceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "issue_id", Type: "integer"},
		{Name: "target", Type: "string"},
		{Name: "port", Type: "integer"},
		{Name: "age", Type: "string"},
		{Name: "extra_info", Type: "object"},
		{Name: "snoozed", Type: "boolean"},
		{Name: "snooze_reason", Type: "string"},
		{Name: "snooze_until", Type: "string"},
	}
}

func intruderScanFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "status", Type: "string"},
		{Name: "created_at", Type: "string"},
	}
}

func intruderTargetFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "address", Type: "string"},
		{Name: "tags", Type: "array"},
	}
}

func intruderIssueRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"title":         item["title"],
		"description":   item["description"],
		"severity":      item["severity"],
		"remediation":   item["remediation"],
		"occurrences":   item["occurrences"],
		"snoozed":       item["snoozed"],
		"snooze_reason": item["snooze_reason"],
		"snooze_until":  item["snooze_until"],
	}
}

func intruderOccurrenceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"issue_id":      item["issue_id"],
		"target":        item["target"],
		"port":          item["port"],
		"age":           item["age"],
		"extra_info":    item["extra_info"],
		"snoozed":       item["snoozed"],
		"snooze_reason": item["snooze_reason"],
		"snooze_until":  item["snooze_until"],
	}
}

func intruderScanRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"status":     item["status"],
		"created_at": item["created_at"],
	}
}

func intruderTargetRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":      item["id"],
		"address": item["address"],
		"tags":    item["tags"],
	}
}
