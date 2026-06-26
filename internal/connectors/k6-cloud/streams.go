package k6cloud

import "polymetrics.ai/internal/connectors"

// streamSpec describes how a k6-cloud stream is read. k6 Cloud has three shapes:
//   - a flat single-page list (organizations) read from /v3/organizations,
//   - a substream (projects) read per organization id, and
//   - a page-incremented list (k6-tests) read from loadtests/v2/tests.
//
// The read path in k6_cloud.go is driven entirely from this table plus the
// boolean flags below, so adding a stream is: add a streamSpec entry and a
// Stream definition in k6Streams.
type streamSpec struct {
	// resource is the API path (relative to base_url). For substreams it is a
	// template with a single %d that the parent id is substituted into.
	resource string
	// recordsPath is the dotted JSON path to the records array in the response.
	recordsPath string
	// paginated is true for endpoints that support PageIncrement pagination.
	paginated bool
	// perOrganization is true for substreams that must be read once per
	// organization id (projects).
	perOrganization bool
	// mapRecord flattens a raw API object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// k6StreamSpecs is the per-stream routing table.
var k6StreamSpecs = map[string]streamSpec{
	"organizations": {
		resource:    "/v3/organizations",
		recordsPath: "organizations",
		mapRecord:   organizationRecord,
	},
	"projects": {
		resource:        "/v3/organizations/%d/projects",
		recordsPath:     "projects",
		paginated:       true,
		perOrganization: true,
		mapRecord:       projectRecord,
	},
	"k6-tests": {
		resource:    "loadtests/v2/tests",
		recordsPath: "k6-tests",
		paginated:   true,
		mapRecord:   testRecord,
	},
}

// k6Streams returns the connector's published stream catalog. k6 Cloud supports
// only full_refresh sync, so no CursorFields are declared.
func k6Streams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "organizations",
			Description: "k6 Cloud organizations the API token can access.",
			PrimaryKey:  []string{"id"},
			Fields:      organizationFields(),
		},
		{
			Name:        "projects",
			Description: "k6 Cloud projects across all accessible organizations.",
			PrimaryKey:  []string{"id"},
			Fields:      projectFields(),
		},
		{
			Name:        "k6-tests",
			Description: "k6 Cloud load tests.",
			PrimaryKey:  []string{"id"},
			Fields:      testFields(),
		},
	}
}

func organizationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "owner_id", Type: "integer"},
		{Name: "billing_address", Type: "string"},
		{Name: "billing_country", Type: "string"},
		{Name: "billing_email", Type: "string"},
		{Name: "vat_number", Type: "string"},
		{Name: "is_default", Type: "boolean"},
		{Name: "is_saml_org", Type: "boolean"},
		{Name: "created", Type: "string"},
		{Name: "updated", Type: "string"},
	}
}

func projectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "organization_id", Type: "integer"},
		{Name: "is_default", Type: "boolean"},
		{Name: "created", Type: "string"},
		{Name: "updated", Type: "string"},
	}
}

func testFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "project_id", Type: "integer"},
		{Name: "user_id", Type: "integer"},
		{Name: "last_test_run_id", Type: "string"},
		{Name: "test_run_ids", Type: "array"},
		{Name: "script", Type: "string"},
		{Name: "created", Type: "string"},
		{Name: "updated", Type: "string"},
	}
}

func organizationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"name":            item["name"],
		"description":     item["description"],
		"owner_id":        item["owner_id"],
		"billing_address": item["billing_address"],
		"billing_country": item["billing_country"],
		"billing_email":   item["billing_email"],
		"vat_number":      item["vat_number"],
		"is_default":      item["is_default"],
		"is_saml_org":     item["is_saml_org"],
		"created":         item["created"],
		"updated":         item["updated"],
	}
}

func projectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"name":            item["name"],
		"description":     item["description"],
		"organization_id": item["organization_id"],
		"is_default":      item["is_default"],
		"created":         item["created"],
		"updated":         item["updated"],
	}
}

func testRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"name":             item["name"],
		"project_id":       item["project_id"],
		"user_id":          item["user_id"],
		"last_test_run_id": item["last_test_run_id"],
		"test_run_ids":     item["test_run_ids"],
		"script":           item["script"],
		"created":          item["created"],
		"updated":          item["updated"],
	}
}
