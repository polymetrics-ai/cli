package freshcaller

import "polymetrics.ai/internal/connectors"

type streamEndpoint struct {
	resource    string
	recordsPath string
	mapRecord   func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{
	"calls":   {resource: "calls", recordsPath: "calls", mapRecord: callRecord},
	"agents":  {resource: "agents", recordsPath: "agents", mapRecord: agentRecord},
	"teams":   {resource: "teams", recordsPath: "teams", mapRecord: teamRecord},
	"numbers": {resource: "numbers", recordsPath: "numbers", mapRecord: numberRecord},
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "calls", Description: "Freshcaller call records.", PrimaryKey: []string{"id"}, CursorFields: []string{"call_time"}, Fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "direction", Type: "string"}, {Name: "status", Type: "string"}, {Name: "call_time", Type: "timestamp"}, {Name: "duration", Type: "integer"}, {Name: "agent_id", Type: "integer"}, {Name: "phone_number", Type: "string"}}},
		{Name: "agents", Description: "Freshcaller agents.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "name", Type: "string"}, {Name: "email", Type: "string"}, {Name: "status", Type: "string"}}},
		{Name: "teams", Description: "Freshcaller teams.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "name", Type: "string"}}},
		{Name: "numbers", Description: "Freshcaller phone numbers.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "integer"}, {Name: "phone_number", Type: "string"}, {Name: "name", Type: "string"}}},
	}
}

func callRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "direction": item["direction"], "status": item["status"], "call_time": item["call_time"], "duration": item["duration"], "agent_id": item["agent_id"], "phone_number": item["phone_number"]}
}

func agentRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"], "email": item["email"], "status": item["status"]}
}

func teamRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"]}
}

func numberRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "phone_number": item["phone_number"], "name": item["name"]}
}
