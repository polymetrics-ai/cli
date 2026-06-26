package insightly

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Insightly API resource path (relative
// to base_url) it reads from, the resource's primary-key field name in the raw
// payload, and the record mapper that flattens its objects.
type streamEndpoint struct {
	// resource is the Insightly list endpoint path segment (e.g. "Contacts").
	// Insightly resource paths are PascalCase.
	resource string
	// idField is the raw payload's primary-key field (e.g. "CONTACT_ID"). It is
	// surfaced as a normalized "id" on every mapped record.
	idField string
	// mapRecord flattens a raw Insightly object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// insightlyStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in insightlyStreams; the read
// path is fully data-driven from this table.
var insightlyStreamEndpoints = map[string]streamEndpoint{
	"contacts":      {resource: "Contacts", idField: "CONTACT_ID", mapRecord: insightlyContactRecord},
	"organisations": {resource: "Organisations", idField: "ORGANISATION_ID", mapRecord: insightlyOrganisationRecord},
	"opportunities": {resource: "Opportunities", idField: "OPPORTUNITY_ID", mapRecord: insightlyOpportunityRecord},
	"leads":         {resource: "Leads", idField: "LEAD_ID", mapRecord: insightlyLeadRecord},
	"projects":      {resource: "Projects", idField: "PROJECT_ID", mapRecord: insightlyProjectRecord},
	"tasks":         {resource: "Tasks", idField: "TASK_ID", mapRecord: insightlyTaskRecord},
}

// insightlyStreams returns the connector's published stream catalog. Every
// Insightly object exposes a numeric `{RESOURCE}_ID` primary key and a
// DATE_UPDATED_UTC timestamp, which is the incremental cursor field.
func insightlyStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "contacts",
			Description:  "Insightly contacts (people).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_updated_utc"},
			Fields:       insightlyContactFields(),
		},
		{
			Name:         "organisations",
			Description:  "Insightly organisations (companies).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_updated_utc"},
			Fields:       insightlyOrganisationFields(),
		},
		{
			Name:         "opportunities",
			Description:  "Insightly sales opportunities.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_updated_utc"},
			Fields:       insightlyOpportunityFields(),
		},
		{
			Name:         "leads",
			Description:  "Insightly leads.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_updated_utc"},
			Fields:       insightlyLeadFields(),
		},
		{
			Name:         "projects",
			Description:  "Insightly projects.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_updated_utc"},
			Fields:       insightlyProjectFields(),
		},
		{
			Name:         "tasks",
			Description:  "Insightly tasks.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_updated_utc"},
			Fields:       insightlyTaskFields(),
		},
	}
}

func insightlyContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "contact_id", Type: "integer"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "email_address", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "organisation_id", Type: "integer"},
		{Name: "title", Type: "string"},
		{Name: "date_created_utc", Type: "string"},
		{Name: "date_updated_utc", Type: "string"},
	}
}

func insightlyOrganisationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "organisation_id", Type: "integer"},
		{Name: "organisation_name", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "website", Type: "string"},
		{Name: "owner_user_id", Type: "integer"},
		{Name: "date_created_utc", Type: "string"},
		{Name: "date_updated_utc", Type: "string"},
	}
}

func insightlyOpportunityFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "opportunity_id", Type: "integer"},
		{Name: "opportunity_name", Type: "string"},
		{Name: "opportunity_state", Type: "string"},
		{Name: "opportunity_value", Type: "number"},
		{Name: "probability", Type: "integer"},
		{Name: "bid_currency", Type: "string"},
		{Name: "pipeline_id", Type: "integer"},
		{Name: "stage_id", Type: "integer"},
		{Name: "date_created_utc", Type: "string"},
		{Name: "date_updated_utc", Type: "string"},
	}
}

func insightlyLeadFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "lead_id", Type: "integer"},
		{Name: "first_name", Type: "string"},
		{Name: "last_name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "organisation_name", Type: "string"},
		{Name: "lead_status_id", Type: "integer"},
		{Name: "lead_source_id", Type: "integer"},
		{Name: "converted", Type: "boolean"},
		{Name: "date_created_utc", Type: "string"},
		{Name: "date_updated_utc", Type: "string"},
	}
}

func insightlyProjectFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "project_id", Type: "integer"},
		{Name: "project_name", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "pipeline_id", Type: "integer"},
		{Name: "stage_id", Type: "integer"},
		{Name: "owner_user_id", Type: "integer"},
		{Name: "date_created_utc", Type: "string"},
		{Name: "date_updated_utc", Type: "string"},
	}
}

func insightlyTaskFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "task_id", Type: "integer"},
		{Name: "title", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "priority", Type: "integer"},
		{Name: "completed", Type: "boolean"},
		{Name: "due_date", Type: "string"},
		{Name: "owner_user_id", Type: "integer"},
		{Name: "date_created_utc", Type: "string"},
		{Name: "date_updated_utc", Type: "string"},
	}
}

func insightlyContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["CONTACT_ID"],
		"contact_id":       item["CONTACT_ID"],
		"first_name":       item["FIRST_NAME"],
		"last_name":        item["LAST_NAME"],
		"email_address":    item["EMAIL_ADDRESS"],
		"phone":            item["PHONE"],
		"organisation_id":  item["ORGANISATION_ID"],
		"title":            item["TITLE"],
		"date_created_utc": item["DATE_CREATED_UTC"],
		"date_updated_utc": item["DATE_UPDATED_UTC"],
	}
}

func insightlyOrganisationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["ORGANISATION_ID"],
		"organisation_id":   item["ORGANISATION_ID"],
		"organisation_name": item["ORGANISATION_NAME"],
		"phone":             item["PHONE"],
		"website":           item["WEBSITE"],
		"owner_user_id":     item["OWNER_USER_ID"],
		"date_created_utc":  item["DATE_CREATED_UTC"],
		"date_updated_utc":  item["DATE_UPDATED_UTC"],
	}
}

func insightlyOpportunityRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["OPPORTUNITY_ID"],
		"opportunity_id":    item["OPPORTUNITY_ID"],
		"opportunity_name":  item["OPPORTUNITY_NAME"],
		"opportunity_state": item["OPPORTUNITY_STATE"],
		"opportunity_value": item["OPPORTUNITY_VALUE"],
		"probability":       item["PROBABILITY"],
		"bid_currency":      item["BID_CURRENCY"],
		"pipeline_id":       item["PIPELINE_ID"],
		"stage_id":          item["STAGE_ID"],
		"date_created_utc":  item["DATE_CREATED_UTC"],
		"date_updated_utc":  item["DATE_UPDATED_UTC"],
	}
}

func insightlyLeadRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["LEAD_ID"],
		"lead_id":           item["LEAD_ID"],
		"first_name":        item["FIRST_NAME"],
		"last_name":         item["LAST_NAME"],
		"email":             item["EMAIL"],
		"organisation_name": item["ORGANISATION_NAME"],
		"lead_status_id":    item["LEAD_STATUS_ID"],
		"lead_source_id":    item["LEAD_SOURCE_ID"],
		"converted":         item["CONVERTED"],
		"date_created_utc":  item["DATE_CREATED_UTC"],
		"date_updated_utc":  item["DATE_UPDATED_UTC"],
	}
}

func insightlyProjectRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["PROJECT_ID"],
		"project_id":       item["PROJECT_ID"],
		"project_name":     item["PROJECT_NAME"],
		"status":           item["STATUS"],
		"pipeline_id":      item["PIPELINE_ID"],
		"stage_id":         item["STAGE_ID"],
		"owner_user_id":    item["OWNER_USER_ID"],
		"date_created_utc": item["DATE_CREATED_UTC"],
		"date_updated_utc": item["DATE_UPDATED_UTC"],
	}
}

func insightlyTaskRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["TASK_ID"],
		"task_id":          item["TASK_ID"],
		"title":            item["TITLE"],
		"status":           item["STATUS"],
		"priority":         item["PRIORITY"],
		"completed":        item["COMPLETED"],
		"due_date":         item["DUE_DATE"],
		"owner_user_id":    item["OWNER_USER_ID"],
		"date_created_utc": item["DATE_CREATED_UTC"],
		"date_updated_utc": item["DATE_UPDATED_UTC"],
	}
}
