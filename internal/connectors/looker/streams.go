package looker

import "polymetrics.ai/internal/connectors"

type streamEndpoint struct {
	resource  string
	mapRecord func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{
	"users":      {resource: "users", mapRecord: userRecord},
	"groups":     {resource: "groups", mapRecord: groupRecord},
	"folders":    {resource: "folders", mapRecord: folderRecord},
	"looks":      {resource: "looks", mapRecord: lookRecord},
	"dashboards": {resource: "dashboards", mapRecord: dashboardRecord},
}

func streams() []connectors.Stream {
	return []connectors.Stream{
		{Name: "users", Description: "Looker users.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "display_name", Type: "string"}, {Name: "email", Type: "string"}}},
		{Name: "groups", Description: "Looker groups.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}}},
		{Name: "folders", Description: "Looker folders.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "name", Type: "string"}}},
		{Name: "looks", Description: "Looker looks.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "title", Type: "string"}, {Name: "folder_id", Type: "string"}}},
		{Name: "dashboards", Description: "Looker dashboards.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "title", Type: "string"}, {Name: "folder_id", Type: "string"}}},
	}
}

func userRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "display_name": first(item, "display_name", "displayName", "name"), "email": item["email"]}
}
func groupRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "name": item["name"]}
}
func folderRecord(item map[string]any) connectors.Record { return groupRecord(item) }
func lookRecord(item map[string]any) connectors.Record {
	return connectors.Record{"id": item["id"], "title": first(item, "title", "name"), "folder_id": first(item, "folder_id", "folderId")}
}
func dashboardRecord(item map[string]any) connectors.Record { return lookRecord(item) }
func first(item map[string]any, keys ...string) any {
	for _, key := range keys {
		if item[key] != nil {
			return item[key]
		}
	}
	return nil
}
