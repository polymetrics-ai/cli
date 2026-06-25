package algolia

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the Algolia API resource path (relative
// to base_url), the JSON path to its records array, and the record mapper that
// flattens its objects.
type streamEndpoint struct {
	// resource is the Algolia API path segment, e.g. "1/indexes".
	resource string
	// recordsPath is the dotted JSON path to the array of records in the
	// response body, e.g. "items" for indexes or "keys" for api keys.
	recordsPath string
	// mapRecord flattens a raw Algolia object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
	// paged is true when the endpoint supports page-number pagination via the
	// "page"/"nbPages" convention (only the indices endpoint does).
	paged bool
}

// algoliaStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in algoliaStreams; the read
// path is fully data-driven from this table.
var algoliaStreamEndpoints = map[string]streamEndpoint{
	"indices":        {resource: "1/indexes", recordsPath: "items", mapRecord: algoliaIndexRecord, paged: true},
	"api_keys":       {resource: "1/keys", recordsPath: "keys", mapRecord: algoliaKeyRecord},
	"index_settings": {resource: "1/indexes/%s/settings", recordsPath: ".", mapRecord: algoliaSettingsRecord},
}

// algoliaStreams returns the connector's published stream catalog. Algolia's
// REST API is a configuration/management API (indices, API keys, settings)
// supporting full-refresh sync only, so streams carry no incremental cursor.
func algoliaStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "indices",
			Description: "Algolia indices (name, record counts, replicas) for the application.",
			PrimaryKey:  []string{"name"},
			Fields:      algoliaIndexFields(),
		},
		{
			Name:        "api_keys",
			Description: "Algolia API keys configured for the application (ACL, validity, scope).",
			PrimaryKey:  []string{"value"},
			Fields:      algoliaKeyFields(),
		},
		{
			Name:        "index_settings",
			Description: "Settings for a single Algolia index (selected via config index_name).",
			PrimaryKey:  []string{"index_name"},
			Fields:      algoliaSettingsFields(),
		},
	}
}

func algoliaIndexFields() []connectors.Field {
	return []connectors.Field{
		{Name: "name", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "updated_at", Type: "string"},
		{Name: "entries", Type: "integer"},
		{Name: "data_size", Type: "integer"},
		{Name: "file_size", Type: "integer"},
		{Name: "last_build_time_s", Type: "integer"},
		{Name: "number_of_pending_tasks", Type: "integer"},
		{Name: "pending_task", Type: "boolean"},
		{Name: "primary", Type: "string"},
		{Name: "replicas", Type: "array"},
	}
}

func algoliaKeyFields() []connectors.Field {
	return []connectors.Field{
		{Name: "value", Type: "string"},
		{Name: "created_at", Type: "integer"},
		{Name: "description", Type: "string"},
		{Name: "acl", Type: "array"},
		{Name: "indexes", Type: "array"},
		{Name: "validity", Type: "integer"},
		{Name: "max_hits_per_query", Type: "integer"},
		{Name: "max_queries_per_ip_per_hour", Type: "integer"},
		{Name: "referers", Type: "array"},
	}
}

func algoliaSettingsFields() []connectors.Field {
	return []connectors.Field{
		{Name: "index_name", Type: "string"},
		{Name: "searchable_attributes", Type: "array"},
		{Name: "attributes_for_faceting", Type: "array"},
		{Name: "ranking", Type: "array"},
		{Name: "custom_ranking", Type: "array"},
		{Name: "replicas", Type: "array"},
		{Name: "hits_per_page", Type: "integer"},
		{Name: "pagination_limited_to", Type: "integer"},
	}
}

func algoliaIndexRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"name":                    item["name"],
		"created_at":              item["createdAt"],
		"updated_at":              item["updatedAt"],
		"entries":                 item["entries"],
		"data_size":               item["dataSize"],
		"file_size":               item["fileSize"],
		"last_build_time_s":       item["lastBuildTimeS"],
		"number_of_pending_tasks": item["numberOfPendingTasks"],
		"pending_task":            item["pendingTask"],
		"primary":                 item["primary"],
		"replicas":                item["replicas"],
	}
}

func algoliaKeyRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"value":                       item["value"],
		"created_at":                  item["createdAt"],
		"description":                 item["description"],
		"acl":                         item["acl"],
		"indexes":                     item["indexes"],
		"validity":                    item["validity"],
		"max_hits_per_query":          item["maxHitsPerQuery"],
		"max_queries_per_ip_per_hour": item["maxQueriesPerIPPerHour"],
		"referers":                    item["referers"],
	}
}

func algoliaSettingsRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"index_name":              item["index_name"],
		"searchable_attributes":   item["searchableAttributes"],
		"attributes_for_faceting": item["attributesForFaceting"],
		"ranking":                 item["ranking"],
		"custom_ranking":          item["customRanking"],
		"replicas":                item["replicas"],
		"hits_per_page":           item["hitsPerPage"],
		"pagination_limited_to":   item["paginationLimitedTo"],
	}
}
