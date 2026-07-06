package ashby

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Ashby API resource path (relative to
// base_url) it reads from, and the record mapper that flattens its objects.
//
// Ashby list endpoints are POSTed at "<resource>/list" (e.g. "candidate/list")
// and return {success, results:[...], moreDataAvailable, nextCursor}.
type streamEndpoint struct {
	// resource is the Ashby list endpoint path (e.g. "candidate/list").
	resource string
	// mapRecord flattens a raw Ashby object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// ashbyStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in ashbyStreams; the read path
// is fully data-driven from this table.
var ashbyStreamEndpoints = map[string]streamEndpoint{
	"candidates":   {resource: "candidate/list", mapRecord: ashbyCandidateRecord},
	"jobs":         {resource: "job/list", mapRecord: ashbyJobRecord},
	"applications": {resource: "application/list", mapRecord: ashbyApplicationRecord},
	"users":        {resource: "user/list", mapRecord: ashbyUserRecord},
}

// ashbyStreams returns the connector's published stream catalog. Every Ashby
// object exposes a string id and updatedAt/createdAt timestamps, so the primary
// key is ["id"] across the board and updatedAt is the incremental cursor field
// where present.
func ashbyStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "candidates",
			Description:  "Ashby candidates.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       ashbyCandidateFields(),
		},
		{
			Name:         "jobs",
			Description:  "Ashby jobs.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       ashbyJobFields(),
		},
		{
			Name:         "applications",
			Description:  "Ashby applications.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       ashbyApplicationFields(),
		},
		{
			Name:         "users",
			Description:  "Ashby users (recruiters, hiring managers, and other org members).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       ashbyUserFields(),
		},
	}
}

func ashbyCandidateFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "primaryEmailAddress", Type: "object"},
		{Name: "primaryPhoneNumber", Type: "object"},
		{Name: "company", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "locationSummary", Type: "string"},
		{Name: "timezone", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func ashbyJobFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "employmentType", Type: "string"},
		{Name: "locationId", Type: "string"},
		{Name: "departmentId", Type: "string"},
		{Name: "defaultInterviewPlanId", Type: "string"},
		{Name: "customFields", Type: "array"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func ashbyApplicationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "candidateId", Type: "string"},
		{Name: "jobId", Type: "string"},
		{Name: "currentInterviewStageId", Type: "string"},
		{Name: "source", Type: "object"},
		{Name: "archiveReason", Type: "object"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func ashbyUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "firstName", Type: "string"},
		{Name: "lastName", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "globalRole", Type: "string"},
		{Name: "isEnabled", Type: "boolean"},
		{Name: "updatedAt", Type: "string"},
	}
}

func ashbyCandidateRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"name":                item["name"],
		"primaryEmailAddress": item["primaryEmailAddress"],
		"primaryPhoneNumber":  item["primaryPhoneNumber"],
		"company":             item["company"],
		"title":               item["title"],
		"locationSummary":     item["locationSummary"],
		"timezone":            item["timezone"],
		"createdAt":           item["createdAt"],
		"updatedAt":           item["updatedAt"],
	}
}

func ashbyJobRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                     item["id"],
		"title":                  item["title"],
		"status":                 item["status"],
		"employmentType":         item["employmentType"],
		"locationId":             item["locationId"],
		"departmentId":           item["departmentId"],
		"defaultInterviewPlanId": item["defaultInterviewPlanId"],
		"customFields":           item["customFields"],
		"createdAt":              item["createdAt"],
		"updatedAt":              item["updatedAt"],
	}
}

func ashbyApplicationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                      item["id"],
		"status":                  item["status"],
		"candidateId":             item["candidateId"],
		"jobId":                   item["jobId"],
		"currentInterviewStageId": item["currentInterviewStageId"],
		"source":                  item["source"],
		"archiveReason":           item["archiveReason"],
		"createdAt":               item["createdAt"],
		"updatedAt":               item["updatedAt"],
	}
}

func ashbyUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"firstName":  item["firstName"],
		"lastName":   item["lastName"],
		"email":      item["email"],
		"globalRole": item["globalRole"],
		"isEnabled":  item["isEnabled"],
		"updatedAt":  item["updatedAt"],
	}
}
