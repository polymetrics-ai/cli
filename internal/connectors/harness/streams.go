package harness

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Harness NextGen API resource path
// (relative to base_url) it reads from, the JSON wrapper key inside each
// data.content[] element (NextGen wraps each list item under a singular key,
// e.g. "organization"), and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the NextGen list endpoint path (e.g. "ng/api/organizations").
	resource string
	// wrapper is the singular key each content element is nested under
	// (e.g. content[i].organization). Empty means the element is the record.
	wrapper string
	// mapRecord flattens a raw Harness object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// harnessStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in harnessStreams; the read
// path is fully data-driven from this table.
var harnessStreamEndpoints = map[string]streamEndpoint{
	"organizations": {resource: "ng/api/organizations", wrapper: "organization", mapRecord: harnessOrganizationRecord},
	"projects":      {resource: "ng/api/projects", wrapper: "project", mapRecord: harnessProjectRecord},
	"services":      {resource: "ng/api/servicesV2", wrapper: "service", mapRecord: harnessServiceRecord},
	"connectors":    {resource: "ng/api/connectors", wrapper: "connector", mapRecord: harnessConnectorRecord},
	"pipelines":     {resource: "pipeline/api/pipelines/list", wrapper: "", mapRecord: harnessPipelineRecord},
}

// harnessStreams returns the connector's published stream catalog. Harness
// NextGen resources are addressed by a string identifier, so the primary key is
// ["identifier"] across the board. These endpoints do not expose a reliable
// monotonic cursor for full-refresh sources, so CursorFields is left empty.
func harnessStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "organizations",
			Description: "Harness NextGen organizations within the account.",
			PrimaryKey:  []string{"identifier"},
			Fields:      harnessOrganizationFields(),
		},
		{
			Name:        "projects",
			Description: "Harness NextGen projects across organizations.",
			PrimaryKey:  []string{"identifier"},
			Fields:      harnessProjectFields(),
		},
		{
			Name:        "services",
			Description: "Harness NextGen services (servicesV2).",
			PrimaryKey:  []string{"identifier"},
			Fields:      harnessServiceFields(),
		},
		{
			Name:        "connectors",
			Description: "Harness NextGen connectors.",
			PrimaryKey:  []string{"identifier"},
			Fields:      harnessConnectorFields(),
		},
		{
			Name:        "pipelines",
			Description: "Harness pipelines (pipeline service list).",
			PrimaryKey:  []string{"identifier"},
			Fields:      harnessPipelineFields(),
		},
	}
}

func harnessOrganizationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "identifier", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "account_identifier", Type: "string"},
		{Name: "description", Type: "string"},
	}
}

func harnessProjectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "identifier", Type: "string"},
		{Name: "org_identifier", Type: "string"},
		{Name: "account_identifier", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "color", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "modules", Type: "array"},
	}
}

func harnessServiceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "identifier", Type: "string"},
		{Name: "org_identifier", Type: "string"},
		{Name: "project_identifier", Type: "string"},
		{Name: "account_identifier", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "deleted", Type: "boolean"},
	}
}

func harnessConnectorFields() []connectors.Field {
	return []connectors.Field{
		{Name: "identifier", Type: "string"},
		{Name: "org_identifier", Type: "string"},
		{Name: "project_identifier", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "description", Type: "string"},
	}
}

func harnessPipelineFields() []connectors.Field {
	return []connectors.Field{
		{Name: "identifier", Type: "string"},
		{Name: "org_identifier", Type: "string"},
		{Name: "project_identifier", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "stage_count", Type: "integer"},
	}
}

func harnessOrganizationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"identifier":         item["identifier"],
		"name":               item["name"],
		"account_identifier": item["accountIdentifier"],
		"description":        item["description"],
	}
}

func harnessProjectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"identifier":         item["identifier"],
		"org_identifier":     item["orgIdentifier"],
		"account_identifier": item["accountIdentifier"],
		"name":               item["name"],
		"color":              item["color"],
		"description":        item["description"],
		"modules":            item["modules"],
	}
}

func harnessServiceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"identifier":         item["identifier"],
		"org_identifier":     item["orgIdentifier"],
		"project_identifier": item["projectIdentifier"],
		"account_identifier": item["accountIdentifier"],
		"name":               item["name"],
		"description":        item["description"],
		"deleted":            item["deleted"],
	}
}

func harnessConnectorRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"identifier":         item["identifier"],
		"org_identifier":     item["orgIdentifier"],
		"project_identifier": item["projectIdentifier"],
		"name":               item["name"],
		"type":               item["type"],
		"description":        item["description"],
	}
}

func harnessPipelineRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"identifier":         item["identifier"],
		"org_identifier":     item["orgIdentifier"],
		"project_identifier": item["projectIdentifier"],
		"name":               item["name"],
		"description":        item["description"],
		"stage_count":        item["numOfStages"],
	}
}
