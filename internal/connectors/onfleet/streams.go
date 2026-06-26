package onfleet

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Onfleet API resource path (relative
// to base_url) it reads from, how the response is shaped, and the record mapper
// that flattens its objects.
//
// Onfleet has two response shapes among the core streams:
//   - list endpoints (workers, teams, hubs, admins) return a top-level JSON
//     array; arrayPath is "" (root) and they are not paginated.
//   - the tasks endpoint (tasks/all) returns {lastId, tasks:[...]} and is
//     paginated via the lastId cursor; arrayPath is "tasks" and paginated=true.
type streamEndpoint struct {
	// resource is the Onfleet endpoint path segment (e.g. "workers", "tasks/all").
	resource string
	// arrayPath is the dotted JSON path to the records array ("" = root array).
	arrayPath string
	// paginated reports whether the endpoint uses Onfleet lastId pagination.
	paginated bool
	// mapRecord flattens a raw Onfleet object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// onfleetStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in onfleetStreams; the read
// path is fully data-driven from this table.
var onfleetStreamEndpoints = map[string]streamEndpoint{
	"tasks":          {resource: "tasks/all", arrayPath: "tasks", paginated: true, mapRecord: onfleetTaskRecord},
	"workers":        {resource: "workers", arrayPath: "", paginated: false, mapRecord: onfleetWorkerRecord},
	"teams":          {resource: "teams", arrayPath: "", paginated: false, mapRecord: onfleetTeamRecord},
	"hubs":           {resource: "hubs", arrayPath: "", paginated: false, mapRecord: onfleetHubRecord},
	"administrators": {resource: "admins", arrayPath: "", paginated: false, mapRecord: onfleetAdminRecord},
}

// onfleetStreams returns the connector's published stream catalog. Every Onfleet
// object exposes a string id and a unix-milliseconds timeCreated/timeLastModified
// timestamp, so the primary key is ["id"] and the cursor field is
// ["timeLastModified"] where the resource tracks modification time.
func onfleetStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "tasks",
			Description:  "Onfleet delivery tasks.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"timeLastModified"},
			Fields:       onfleetTaskFields(),
		},
		{
			Name:         "workers",
			Description:  "Onfleet workers (drivers).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"timeLastModified"},
			Fields:       onfleetWorkerFields(),
		},
		{
			Name:         "teams",
			Description:  "Onfleet teams.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"timeLastModified"},
			Fields:       onfleetTeamFields(),
		},
		{
			Name:         "hubs",
			Description:  "Onfleet hubs (depots).",
			PrimaryKey:   []string{"id"},
			CursorFields: nil,
			Fields:       onfleetHubFields(),
		},
		{
			Name:         "administrators",
			Description:  "Onfleet organization administrators.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"timeLastModified"},
			Fields:       onfleetAdminFields(),
		},
	}
}

func onfleetTaskFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "shortId", Type: "string"},
		{Name: "state", Type: "integer"},
		{Name: "completed", Type: "boolean"},
		{Name: "trackingURL", Type: "string"},
		{Name: "worker", Type: "string"},
		{Name: "merchant", Type: "string"},
		{Name: "executor", Type: "string"},
		{Name: "creator", Type: "string"},
		{Name: "timeCreated", Type: "integer"},
		{Name: "timeLastModified", Type: "integer"},
	}
}

func onfleetWorkerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "onDuty", Type: "boolean"},
		{Name: "activeTask", Type: "string"},
		{Name: "timeLastSeen", Type: "integer"},
		{Name: "timeCreated", Type: "integer"},
		{Name: "timeLastModified", Type: "integer"},
	}
}

func onfleetTeamFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "hub", Type: "string"},
		{Name: "timeCreated", Type: "integer"},
		{Name: "timeLastModified", Type: "integer"},
	}
}

func onfleetHubFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "address", Type: "string"},
	}
}

func onfleetAdminFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "isActive", Type: "boolean"},
		{Name: "timeCreated", Type: "integer"},
		{Name: "timeLastModified", Type: "integer"},
	}
}

func onfleetTaskRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"shortId":          item["shortId"],
		"state":            item["state"],
		"completed":        item["completed"],
		"trackingURL":      item["trackingURL"],
		"worker":           item["worker"],
		"merchant":         item["merchant"],
		"executor":         item["executor"],
		"creator":          item["creator"],
		"timeCreated":      item["timeCreated"],
		"timeLastModified": item["timeLastModified"],
	}
}

func onfleetWorkerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"name":             item["name"],
		"phone":            item["phone"],
		"onDuty":           item["onDuty"],
		"activeTask":       item["activeTask"],
		"timeLastSeen":     item["timeLastSeen"],
		"timeCreated":      item["timeCreated"],
		"timeLastModified": item["timeLastModified"],
	}
}

func onfleetTeamRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"name":             item["name"],
		"hub":              item["hub"],
		"timeCreated":      item["timeCreated"],
		"timeLastModified": item["timeLastModified"],
	}
}

func onfleetHubRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":      item["id"],
		"name":    item["name"],
		"address": item["address"],
	}
}

func onfleetAdminRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"name":             item["name"],
		"email":            item["email"],
		"type":             item["type"],
		"isActive":         item["isActive"],
		"timeCreated":      item["timeCreated"],
		"timeLastModified": item["timeLastModified"],
	}
}
