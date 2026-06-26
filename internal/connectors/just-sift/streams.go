package justsift

import "polymetrics.ai/internal/connectors"

// paginationKind selects how a stream walks pages.
type paginationKind int

const (
	// pageIncrement advances an integer page parameter (1-based) until a short
	// or empty page is returned (the JustSift /search/people behaviour).
	pageIncrement paginationKind = iota
	// linkCursor follows the links.next cursor token in the response body until
	// it is absent (the JustSift /fields/person behaviour).
	linkCursor
)

// streamEndpoint maps a stream name to the JustSift API resource path (relative
// to base_url), the record mapper that flattens its objects, and its pagination
// style. The read path is fully data-driven from this table.
type streamEndpoint struct {
	resource   string
	pagination paginationKind
	mapRecord  func(map[string]any) connectors.Record
}

// streamEndpoints is the per-stream routing table. Adding a stream means adding
// one entry here plus a Stream definition in streams().
var streamEndpoints = map[string]streamEndpoint{
	"peoples": {resource: "search/people", pagination: pageIncrement, mapRecord: peopleRecord},
	"fields":  {resource: "fields/person", pagination: linkCursor, mapRecord: fieldRecord},
}

// streams returns the connector's published stream catalog. Both JustSift
// streams key on a string id and only support full refresh (no incremental
// cursor), matching the upstream manifest.
func streams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "peoples",
			Description: "JustSift people directory profiles (GET /search/people).",
			PrimaryKey:  []string{"id"},
			Fields:      peopleFields(),
		},
		{
			Name:        "fields",
			Description: "JustSift person field definitions (GET /fields/person).",
			PrimaryKey:  []string{"id"},
			Fields:      fieldFields(),
		},
	}
}

func peopleFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "displayName", Type: "string"},
		{Name: "firstName", Type: "string"},
		{Name: "lastName", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "department", Type: "string"},
		{Name: "companyName", Type: "string"},
		{Name: "officeCity", Type: "string"},
		{Name: "officeState", Type: "string"},
		{Name: "directoryId", Type: "string"},
		{Name: "isTeamLeader", Type: "boolean"},
		{Name: "directReportCount", Type: "number"},
		{Name: "pictureUrl", Type: "string"},
	}
}

func fieldFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "displayName", Type: "string"},
		{Name: "objectKey", Type: "string"},
		{Name: "filterable", Type: "boolean"},
		{Name: "searchable", Type: "boolean"},
	}
}

func peopleRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":                item["id"],
		"displayName":       item["displayName"],
		"firstName":         item["firstName"],
		"lastName":          item["lastName"],
		"email":             item["email"],
		"phone":             item["phone"],
		"title":             item["title"],
		"department":        item["department"],
		"companyName":       item["companyName"],
		"officeCity":        item["officeCity"],
		"officeState":       item["officeState"],
		"directoryId":       item["directoryId"],
		"isTeamLeader":      item["isTeamLeader"],
		"directReportCount": item["directReportCount"],
		"pictureUrl":        item["pictureUrl"],
	}
	rec["connector"] = registryName
	return rec
}

func fieldRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":          item["id"],
		"type":        item["type"],
		"displayName": item["displayName"],
		"objectKey":   item["objectKey"],
		"filterable":  item["filterable"],
		"searchable":  item["searchable"],
	}
	rec["connector"] = registryName
	return rec
}
