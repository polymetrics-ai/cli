package genesys

import "polymetrics.ai/internal/connectors"

type streamEndpoint struct {
	resource  string
	mapRecord func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{
	"users":     {resource: "users", mapRecord: userRecord},
	"queues":    {resource: "routing/queues", mapRecord: queueRecord},
	"groups":    {resource: "groups", mapRecord: groupRecord},
	"divisions": {resource: "authorization/divisions", mapRecord: divisionRecord},
}

func streams() []connectors.Stream {
	base := []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}, {Name: "display_name", Type: "string"}, {Name: "email", Type: "string"}, {Name: "state", Type: "string"}, {Name: "description", Type: "string"}}
	return []connectors.Stream{
		{Name: "users", Description: "Genesys Cloud users.", PrimaryKey: []string{"id"}, Fields: base},
		{Name: "queues", Description: "Genesys Cloud routing queues.", PrimaryKey: []string{"id"}, Fields: base},
		{Name: "groups", Description: "Genesys Cloud groups.", PrimaryKey: []string{"id"}, Fields: base},
		{Name: "divisions", Description: "Genesys Cloud authorization divisions.", PrimaryKey: []string{"id"}, Fields: base},
	}
}

func userRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"], "display_name": item["name"], "email": item["email"], "state": item["state"]}
}

func queueRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"], "description": item["description"]}
}

func groupRecord(item map[string]any) connectors.Record { return queueRecord(item) }

func divisionRecord(item map[string]any) connectors.Record { return queueRecord(item) }
