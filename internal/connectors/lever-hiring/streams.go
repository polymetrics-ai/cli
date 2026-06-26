package leverhiring

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Lever Data API resource path (relative
// to base_url) it reads from, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Lever list endpoint path segment (e.g. "opportunities").
	resource string
	// mapRecord flattens a raw Lever object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// leverStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in leverStreams; the read path
// is fully data-driven from this table.
var leverStreamEndpoints = map[string]streamEndpoint{
	"opportunities": {resource: "opportunities", mapRecord: leverOpportunityRecord},
	"postings":      {resource: "postings", mapRecord: leverPostingRecord},
	"users":         {resource: "users", mapRecord: leverUserRecord},
	"requisitions":  {resource: "requisitions", mapRecord: leverRequisitionRecord},
	"stages":        {resource: "stages", mapRecord: leverStageRecord},
}

// leverStreams returns the connector's published stream catalog. Every Lever
// object exposes a string id; most list objects carry a `createdAt` unix-millis
// timestamp used as the incremental cursor where available.
func leverStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "opportunities",
			Description:  "Lever opportunities (candidate profiles and their applications).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"createdAt"},
			Fields:       leverOpportunityFields(),
		},
		{
			Name:         "postings",
			Description:  "Lever job postings.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"createdAt"},
			Fields:       leverPostingFields(),
		},
		{
			Name:         "users",
			Description:  "Lever users (team members).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"createdAt"},
			Fields:       leverUserFields(),
		},
		{
			Name:         "requisitions",
			Description:  "Lever hiring requisitions.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"createdAt"},
			Fields:       leverRequisitionFields(),
		},
		{
			Name:        "stages",
			Description: "Lever pipeline stages.",
			PrimaryKey:  []string{"id"},
			Fields:      leverStageFields(),
		},
	}
}

func leverOpportunityFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "headline", Type: "string"},
		{Name: "stage", Type: "string"},
		{Name: "origin", Type: "string"},
		{Name: "sources", Type: "array"},
		{Name: "tags", Type: "array"},
		{Name: "emails", Type: "array"},
		{Name: "createdAt", Type: "integer"},
		{Name: "updatedAt", Type: "integer"},
		{Name: "lastInteractionAt", Type: "integer"},
		{Name: "archivedAt", Type: "integer"},
	}
}

func leverPostingFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "text", Type: "string"},
		{Name: "state", Type: "string"},
		{Name: "user", Type: "string"},
		{Name: "owner", Type: "string"},
		{Name: "hiringManager", Type: "string"},
		{Name: "categories", Type: "object"},
		{Name: "createdAt", Type: "integer"},
		{Name: "updatedAt", Type: "integer"},
	}
}

func leverUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "username", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "accessRole", Type: "string"},
		{Name: "deactivatedAt", Type: "integer"},
		{Name: "createdAt", Type: "integer"},
	}
}

func leverRequisitionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "requisitionCode", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "headcountTotal", Type: "integer"},
		{Name: "headcountHired", Type: "integer"},
		{Name: "owner", Type: "string"},
		{Name: "createdAt", Type: "integer"},
		{Name: "updatedAt", Type: "integer"},
	}
}

func leverStageFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "text", Type: "string"},
	}
}

func leverOpportunityRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"name":              item["name"],
		"headline":          item["headline"],
		"stage":             item["stage"],
		"origin":            item["origin"],
		"sources":           item["sources"],
		"tags":              item["tags"],
		"emails":            item["emails"],
		"createdAt":         item["createdAt"],
		"updatedAt":         item["updatedAt"],
		"lastInteractionAt": item["lastInteractionAt"],
		"archivedAt":        item["archivedAt"],
	}
}

func leverPostingRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"text":          item["text"],
		"state":         item["state"],
		"user":          item["user"],
		"owner":         item["owner"],
		"hiringManager": item["hiringManager"],
		"categories":    item["categories"],
		"createdAt":     item["createdAt"],
		"updatedAt":     item["updatedAt"],
	}
}

func leverUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":            item["id"],
		"name":          item["name"],
		"username":      item["username"],
		"email":         item["email"],
		"accessRole":    item["accessRole"],
		"deactivatedAt": item["deactivatedAt"],
		"createdAt":     item["createdAt"],
	}
}

func leverRequisitionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"requisitionCode": item["requisitionCode"],
		"name":            item["name"],
		"status":          item["status"],
		"headcountTotal":  item["headcountTotal"],
		"headcountHired":  item["headcountHired"],
		"owner":           item["owner"],
		"createdAt":       item["createdAt"],
		"updatedAt":       item["updatedAt"],
	}
}

func leverStageRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":   item["id"],
		"text": item["text"],
	}
}
