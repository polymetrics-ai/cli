package youtubeanalytics

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the YouTube Reporting API resource path
// (relative to base_url), the JSON field the records array lives under, and the
// record mapper that flattens each item into a connectors.Record.
type streamEndpoint struct {
	// resource is the path segment under base_url, e.g. "jobs". For "reports"
	// the {jobId} segment is filled from the job_id config at read time.
	resource string
	// recordsPath is the dotted JSON path to the records array in the response.
	recordsPath string
	// mapRecord flattens a raw Reporting API object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
	// needsJobID is true for the reports stream, which is scoped to a job.
	needsJobID bool
}

// streamEndpoints is the per-stream routing table. The read path is fully
// data-driven from this table: adding a stream means adding one entry here plus
// a Stream definition in youtubeStreams.
var streamEndpoints = map[string]streamEndpoint{
	"jobs":         {resource: "jobs", recordsPath: "jobs", mapRecord: jobRecord},
	"report_types": {resource: "reportTypes", recordsPath: "reportTypes", mapRecord: reportTypeRecord},
	"reports":      {resource: "jobs/%s/reports", recordsPath: "reports", mapRecord: reportRecord, needsJobID: true},
}

// youtubeStreams returns the connector's published stream catalog. These are the
// three core resources of the YouTube Reporting API (the bulk-report data plane
// the Airbyte source is built on): the scheduled reporting jobs, the available
// report types, and the generated reports for a job.
func youtubeStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "jobs",
			Description:  "Scheduled YouTube Reporting API jobs for the channel or content owner.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"create_time"},
			Fields:       jobFields(),
		},
		{
			Name:        "report_types",
			Description: "Report types available to be scheduled as reporting jobs.",
			PrimaryKey:  []string{"id"},
			Fields:      reportTypeFields(),
		},
		{
			Name:         "reports",
			Description:  "Generated reports for a reporting job (requires job_id config).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"create_time"},
			Fields:       reportFields(),
		},
	}
}

func jobFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "report_type_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "create_time", Type: "timestamp"},
		{Name: "expire_time", Type: "timestamp"},
		{Name: "system_managed", Type: "boolean"},
	}
}

func reportTypeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "deprecate_time", Type: "timestamp"},
		{Name: "system_managed", Type: "boolean"},
	}
}

func reportFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "job_id", Type: "string"},
		{Name: "start_time", Type: "timestamp"},
		{Name: "end_time", Type: "timestamp"},
		{Name: "create_time", Type: "timestamp"},
		{Name: "job_expire_time", Type: "timestamp"},
		{Name: "download_url", Type: "string"},
	}
}

func jobRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"report_type_id": item["reportTypeId"],
		"name":           item["name"],
		"create_time":    item["createTime"],
		"expire_time":    item["expireTime"],
		"system_managed": item["systemManaged"],
	}
}

func reportTypeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"name":           item["name"],
		"deprecate_time": item["deprecateTime"],
		"system_managed": item["systemManaged"],
	}
}

func reportRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"job_id":          item["jobId"],
		"start_time":      item["startTime"],
		"end_time":        item["endTime"],
		"create_time":     item["createTime"],
		"job_expire_time": item["jobExpireTime"],
		"download_url":    item["downloadUrl"],
	}
}
