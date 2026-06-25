package concord

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Concord REST path it reads from, the
// JSON path where its record array lives, and the mapper that flattens each
// object into a connectors.Record.
//
// Concord paths split into two shapes:
//   - org-scoped paths contain "{org}" which is substituted with the resolved
//     organization id at read time (agreements, folders, reports).
//   - flat paths (user/me/organizations, tags) need no organization id.
//
// recordsPath uses connsdk dotted notation: "" selects a root array, "tags"
// selects body["tags"], etc.
type streamEndpoint struct {
	// resource is the path template relative to base_url. "{org}" is replaced
	// with the organization id.
	resource string
	// recordsPath is the connsdk.RecordsAt selector for the records array.
	recordsPath string
	// orgScoped is true when resource references "{org}" and therefore needs a
	// resolved organization id.
	orgScoped bool
	// mapRecord flattens a raw Concord object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// concordStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in concordStreams; the read
// path is fully data-driven from this table.
var concordStreamEndpoints = map[string]streamEndpoint{
	"agreements":         {resource: "organizations/{org}/agreements", recordsPath: "", orgScoped: true, mapRecord: agreementRecord},
	"user_organizations": {resource: "user/me/organizations", recordsPath: "organizations", orgScoped: false, mapRecord: organizationRecord},
	"folders":            {resource: "organizations/{org}/folders", recordsPath: "", orgScoped: true, mapRecord: folderRecord},
	"reports":            {resource: "organizations/{org}/reports", recordsPath: "reports", orgScoped: true, mapRecord: reportRecord},
	"tags":               {resource: "tags", recordsPath: "tags", orgScoped: false, mapRecord: tagRecord},
}

// concordStreams returns the connector's published stream catalog. Concord only
// supports full-refresh sync, so no cursor fields are declared.
func concordStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "agreements",
			Description: "Concord agreements (contracts) within an organization.",
			PrimaryKey:  []string{"uid"},
			Fields:      agreementFields(),
		},
		{
			Name:        "user_organizations",
			Description: "Organizations the authenticated user belongs to.",
			PrimaryKey:  []string{"id"},
			Fields:      organizationFields(),
		},
		{
			Name:        "folders",
			Description: "Folders within an organization.",
			PrimaryKey:  []string{"id"},
			Fields:      folderFields(),
		},
		{
			Name:        "reports",
			Description: "Saved reports within an organization.",
			PrimaryKey:  []string{"id"},
			Fields:      reportFields(),
		},
		{
			Name:        "tags",
			Description: "Tags available to the authenticated user.",
			PrimaryKey:  []string{"id"},
			Fields:      tagFields(),
		},
	}
}

func agreementFields() []connectors.Field {
	return []connectors.Field{
		{Name: "uid", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "stage", Type: "string"},
		{Name: "organizationId", Type: "integer"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func organizationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "type", Type: "string"},
	}
}

func folderFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "parentId", Type: "integer"},
		{Name: "organizationId", Type: "integer"},
	}
}

func reportFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "organizationId", Type: "integer"},
	}
}

func tagFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "color", Type: "string"},
	}
}

func agreementRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"uid":            item["uid"],
		"title":          item["title"],
		"status":         item["status"],
		"stage":          item["stage"],
		"organizationId": item["organizationId"],
		"createdAt":      item["createdAt"],
		"updatedAt":      item["updatedAt"],
	}
}

func organizationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":   item["id"],
		"name": item["name"],
		"role": item["role"],
		"type": item["type"],
	}
}

func folderRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"name":           item["name"],
		"parentId":       item["parentId"],
		"organizationId": item["organizationId"],
	}
}

func reportRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"name":           item["name"],
		"type":           item["type"],
		"organizationId": item["organizationId"],
	}
}

func tagRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":    item["id"],
		"name":  item["name"],
		"color": item["color"],
	}
}
