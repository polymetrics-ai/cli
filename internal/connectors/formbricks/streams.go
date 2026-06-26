package formbricks

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Formbricks management API resource
// path (relative to base_url) it reads from, the record mapper that flattens its
// objects, and whether it supports offset (skip) pagination.
type streamEndpoint struct {
	// resource is the API path segment (e.g. "management/surveys").
	resource string
	// mapRecord flattens a raw Formbricks object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
	// paginated is true when the endpoint accepts limit/skip offset pagination.
	// Only the responses stream paginates in the Formbricks management API; the
	// rest return their full collection in a single data[] payload.
	paginated bool
}

// formbricksStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in formbricksStreams; the
// read path is fully data-driven from this table. Paths and field selection
// follow the Formbricks management API (records live under "data").
var formbricksStreamEndpoints = map[string]streamEndpoint{
	"surveys":           {resource: "management/surveys", mapRecord: formbricksSurveyRecord},
	"responses":         {resource: "management/responses", mapRecord: formbricksResponseRecord, paginated: true},
	"action_classes":    {resource: "management/action-classes", mapRecord: formbricksActionClassRecord},
	"attribute_classes": {resource: "management/attribute-classes", mapRecord: formbricksAttributeClassRecord},
	"webhooks":          {resource: "webhooks", mapRecord: formbricksWebhookRecord},
}

// formbricksStreams returns the connector's published stream catalog. Every
// Formbricks object exposes a string id and most carry createdAt/updatedAt
// RFC3339 timestamps, so the primary key is ["id"] across the board and the
// cursor field is ["updated_at"] where present.
func formbricksStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "surveys",
			Description:  "Formbricks surveys.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       formbricksSurveyFields(),
		},
		{
			Name:         "responses",
			Description:  "Formbricks survey responses (offset paginated).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       formbricksResponseFields(),
		},
		{
			Name:         "action_classes",
			Description:  "Formbricks action classes (tracked user actions).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       formbricksActionClassFields(),
		},
		{
			Name:         "attribute_classes",
			Description:  "Formbricks contact attribute classes.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updated_at"},
			Fields:       formbricksAttributeClassFields(),
		},
		{
			Name:        "webhooks",
			Description: "Formbricks webhooks.",
			PrimaryKey:  []string{"id"},
			Fields:      formbricksWebhookFields(),
		},
	}
}

func formbricksSurveyFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "environment_id", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func formbricksResponseFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "survey_id", Type: "string"},
		{Name: "contact_id", Type: "string"},
		{Name: "finished", Type: "boolean"},
		{Name: "data", Type: "object"},
		{Name: "meta", Type: "object"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func formbricksActionClassFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "environment_id", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func formbricksAttributeClassFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "environment_id", Type: "string"},
		{Name: "archived", Type: "boolean"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func formbricksWebhookFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "source", Type: "string"},
		{Name: "environment_id", Type: "string"},
		{Name: "surveyIds", Type: "array"},
		{Name: "triggers", Type: "array"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "updated_at", Type: "timestamp"},
	}
}

func formbricksSurveyRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"name":           item["name"],
		"type":           item["type"],
		"status":         item["status"],
		"environment_id": item["environmentId"],
		"created_at":     item["createdAt"],
		"updated_at":     item["updatedAt"],
	}
}

func formbricksResponseRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"survey_id":  item["surveyId"],
		"contact_id": item["contactId"],
		"finished":   item["finished"],
		"data":       item["data"],
		"meta":       item["meta"],
		"created_at": item["createdAt"],
		"updated_at": item["updatedAt"],
	}
}

func formbricksActionClassRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"name":           item["name"],
		"description":    item["description"],
		"type":           item["type"],
		"environment_id": item["environmentId"],
		"created_at":     item["createdAt"],
		"updated_at":     item["updatedAt"],
	}
}

func formbricksAttributeClassRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"name":           item["name"],
		"description":    item["description"],
		"type":           item["type"],
		"environment_id": item["environmentId"],
		"archived":       item["archived"],
		"created_at":     item["createdAt"],
		"updated_at":     item["updatedAt"],
	}
}

func formbricksWebhookRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["id"],
		"url":            item["url"],
		"source":         item["source"],
		"environment_id": item["environmentId"],
		"surveyIds":      item["surveyIds"],
		"triggers":       item["triggers"],
		"created_at":     item["createdAt"],
		"updated_at":     item["updatedAt"],
	}
}
