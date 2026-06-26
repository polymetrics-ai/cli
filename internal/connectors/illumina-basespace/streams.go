package basespace

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the BaseSpace v1pre3 resource path
// (relative to base_url, after the /v1pre3/users/<user> prefix) it reads from,
// and the record mapper that flattens its objects.
//
// BaseSpace list endpoints are user-scoped: GET /v1pre3/users/<user>/<resource>.
// The resource field holds just the trailing segment(s); the user prefix is
// composed at read time so the configured `user` (default "current") applies
// uniformly.
type streamEndpoint struct {
	// resource is the trailing path under /v1pre3/users/<user>/ (e.g. "projects").
	resource string
	// mapRecord flattens a raw BaseSpace object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// basespaceStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in basespaceStreams; the
// read path is fully data-driven from this table.
var basespaceStreamEndpoints = map[string]streamEndpoint{
	"projects":    {resource: "projects", mapRecord: basespaceProjectRecord},
	"runs":        {resource: "runs", mapRecord: basespaceRunRecord},
	"samples":     {resource: "samples", mapRecord: basespaceSampleRecord},
	"appsessions": {resource: "appsessions", mapRecord: basespaceAppSessionRecord},
	"datasets":    {resource: "datasets", mapRecord: basespaceDatasetRecord},
}

// basespaceStreams returns the connector's published stream catalog. Every
// BaseSpace object exposes a string Id and a DateCreated timestamp, so the
// primary key is ["id"] and the cursor field is ["date_created"] across the
// board (the upstream Airbyte source is full-refresh only; cursor_fields are
// published for downstream incremental use but reads do not require them).
func basespaceStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "projects",
			Description:  "BaseSpace projects owned by or shared with the user.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_created"},
			Fields:       basespaceProjectFields(),
		},
		{
			Name:         "runs",
			Description:  "BaseSpace sequencing runs.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_created"},
			Fields:       basespaceRunFields(),
		},
		{
			Name:         "samples",
			Description:  "BaseSpace samples.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_created"},
			Fields:       basespaceSampleFields(),
		},
		{
			Name:         "appsessions",
			Description:  "BaseSpace application sessions (analyses).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_created"},
			Fields:       basespaceAppSessionFields(),
		},
		{
			Name:         "datasets",
			Description:  "BaseSpace datasets.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_created"},
			Fields:       basespaceDatasetFields(),
		},
	}
}

func basespaceProjectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "href", Type: "string"},
		{Name: "date_created", Type: "timestamp"},
		{Name: "date_modified", Type: "timestamp"},
		{Name: "total_size", Type: "integer"},
		{Name: "user_owned_by", Type: "object"},
	}
}

func basespaceRunFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "href", Type: "string"},
		{Name: "experiment_name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "date_created", Type: "timestamp"},
		{Name: "date_modified", Type: "timestamp"},
		{Name: "total_size", Type: "integer"},
		{Name: "instrument_name", Type: "string"},
	}
}

func basespaceSampleFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "href", Type: "string"},
		{Name: "sample_id", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "date_created", Type: "timestamp"},
		{Name: "num_reads_raw", Type: "integer"},
		{Name: "num_reads_pf", Type: "integer"},
		{Name: "total_size", Type: "integer"},
	}
}

func basespaceAppSessionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "href", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "status_summary", Type: "string"},
		{Name: "date_created", Type: "timestamp"},
		{Name: "date_completed", Type: "timestamp"},
		{Name: "total_size", Type: "integer"},
		{Name: "application", Type: "object"},
	}
}

func basespaceDatasetFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "href", Type: "string"},
		{Name: "date_created", Type: "timestamp"},
		{Name: "total_size", Type: "integer"},
		{Name: "dataset_type", Type: "object"},
		{Name: "project", Type: "object"},
	}
}

func basespaceProjectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["Id"],
		"name":          item["Name"],
		"href":          item["Href"],
		"date_created":  item["DateCreated"],
		"date_modified": item["DateModified"],
		"total_size":    item["TotalSize"],
		"user_owned_by": item["UserOwnedBy"],
	}
}

func basespaceRunRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["Id"],
		"name":            item["Name"],
		"href":            item["Href"],
		"experiment_name": item["ExperimentName"],
		"status":          item["Status"],
		"date_created":    item["DateCreated"],
		"date_modified":   item["DateModified"],
		"total_size":      item["TotalSize"],
		"instrument_name": item["InstrumentName"],
	}
}

func basespaceSampleRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["Id"],
		"name":          item["Name"],
		"href":          item["Href"],
		"sample_id":     item["SampleId"],
		"status":        item["Status"],
		"date_created":  item["DateCreated"],
		"num_reads_raw": item["NumReadsRaw"],
		"num_reads_pf":  item["NumReadsPF"],
		"total_size":    item["TotalSize"],
	}
}

func basespaceAppSessionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["Id"],
		"name":           item["Name"],
		"href":           item["Href"],
		"status":         item["Status"],
		"status_summary": item["StatusSummary"],
		"date_created":   item["DateCreated"],
		"date_completed": item["DateCompleted"],
		"total_size":     item["TotalSize"],
		"application":    item["Application"],
	}
}

func basespaceDatasetRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["Id"],
		"name":         item["Name"],
		"href":         item["Href"],
		"date_created": item["DateCreated"],
		"total_size":   item["TotalSize"],
		"dataset_type": item["DatasetType"],
		"project":      item["Project"],
	}
}
