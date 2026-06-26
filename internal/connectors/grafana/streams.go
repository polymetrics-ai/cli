package grafana

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Grafana API resource path (relative
// to base_url) it reads from, the record mapper that flattens its objects, an
// optional fixed query (e.g. type=dash-db on /api/search), and whether the
// endpoint supports page/limit pagination. Grafana endpoints all return
// top-level JSON arrays, so recordsPath is the root ("").
type streamEndpoint struct {
	// resource is the API path segment relative to base_url (e.g. "api/search").
	resource string
	// fixedQuery holds query params applied to every request for the stream
	// (e.g. {"type": "dash-db"} to restrict /api/search to dashboards).
	fixedQuery map[string]string
	// paginated is true for endpoints that honor page/limit; false for
	// list-everything endpoints (datasources, org users, alert rules).
	paginated bool
	// mapRecord flattens a raw Grafana object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// grafanaStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in grafanaStreams; the read
// path is fully data-driven from this table.
var grafanaStreamEndpoints = map[string]streamEndpoint{
	"dashboards": {
		resource:   "api/search",
		fixedQuery: map[string]string{"type": "dash-db"},
		paginated:  true,
		mapRecord:  grafanaDashboardRecord,
	},
	"folders": {
		resource:   "api/search",
		fixedQuery: map[string]string{"type": "dash-folder"},
		paginated:  true,
		mapRecord:  grafanaFolderRecord,
	},
	"datasources": {
		resource:  "api/datasources",
		paginated: false,
		mapRecord: grafanaDatasourceRecord,
	},
	"org_users": {
		resource:  "api/org/users",
		paginated: false,
		mapRecord: grafanaOrgUserRecord,
	},
	"alert_rules": {
		resource:  "api/v1/provisioning/alert-rules",
		paginated: false,
		mapRecord: grafanaAlertRuleRecord,
	},
}

// grafanaStreams returns the connector's published stream catalog.
func grafanaStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "dashboards",
			Description: "Grafana dashboards (GET /api/search?type=dash-db).",
			PrimaryKey:  []string{"uid"},
			Fields:      grafanaDashboardFields(),
		},
		{
			Name:        "folders",
			Description: "Grafana dashboard folders (GET /api/search?type=dash-folder).",
			PrimaryKey:  []string{"uid"},
			Fields:      grafanaFolderFields(),
		},
		{
			Name:        "datasources",
			Description: "Configured Grafana data sources (GET /api/datasources).",
			PrimaryKey:  []string{"uid"},
			Fields:      grafanaDatasourceFields(),
		},
		{
			Name:        "org_users",
			Description: "Users in the current Grafana organization (GET /api/org/users).",
			PrimaryKey:  []string{"userId"},
			Fields:      grafanaOrgUserFields(),
		},
		{
			Name:        "alert_rules",
			Description: "Provisioned Grafana alert rules (GET /api/v1/provisioning/alert-rules).",
			PrimaryKey:  []string{"uid"},
			Fields:      grafanaAlertRuleFields(),
		},
	}
}

func grafanaDashboardFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "uid", Type: "string"},
		{Name: "orgId", Type: "integer"},
		{Name: "title", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "tags", Type: "array"},
		{Name: "isStarred", Type: "boolean"},
		{Name: "folderId", Type: "integer"},
		{Name: "folderUid", Type: "string"},
		{Name: "folderTitle", Type: "string"},
	}
}

func grafanaFolderFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "uid", Type: "string"},
		{Name: "orgId", Type: "integer"},
		{Name: "title", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "tags", Type: "array"},
	}
}

func grafanaDatasourceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "uid", Type: "string"},
		{Name: "orgId", Type: "integer"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "access", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "isDefault", Type: "boolean"},
		{Name: "readOnly", Type: "boolean"},
	}
}

func grafanaOrgUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "orgId", Type: "integer"},
		{Name: "userId", Type: "integer"},
		{Name: "email", Type: "string"},
		{Name: "login", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "lastSeenAt", Type: "string"},
	}
}

func grafanaAlertRuleFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "uid", Type: "string"},
		{Name: "orgID", Type: "integer"},
		{Name: "folderUID", Type: "string"},
		{Name: "ruleGroup", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "condition", Type: "string"},
		{Name: "noDataState", Type: "string"},
		{Name: "execErrState", Type: "string"},
		{Name: "for", Type: "string"},
	}
}

func grafanaDashboardRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"uid":         item["uid"],
		"orgId":       item["orgId"],
		"title":       item["title"],
		"url":         item["url"],
		"type":        item["type"],
		"tags":        item["tags"],
		"isStarred":   item["isStarred"],
		"folderId":    item["folderId"],
		"folderUid":   item["folderUid"],
		"folderTitle": item["folderTitle"],
	}
}

func grafanaFolderRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":    item["id"],
		"uid":   item["uid"],
		"orgId": item["orgId"],
		"title": item["title"],
		"url":   item["url"],
		"type":  item["type"],
		"tags":  item["tags"],
	}
}

func grafanaDatasourceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":        item["id"],
		"uid":       item["uid"],
		"orgId":     item["orgId"],
		"name":      item["name"],
		"type":      item["type"],
		"access":    item["access"],
		"url":       item["url"],
		"isDefault": item["isDefault"],
		"readOnly":  item["readOnly"],
	}
}

func grafanaOrgUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"orgId":      item["orgId"],
		"userId":     item["userId"],
		"email":      item["email"],
		"login":      item["login"],
		"role":       item["role"],
		"lastSeenAt": item["lastSeenAt"],
	}
}

func grafanaAlertRuleRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"uid":          item["uid"],
		"orgID":        item["orgID"],
		"folderUID":    item["folderUID"],
		"ruleGroup":    item["ruleGroup"],
		"title":        item["title"],
		"condition":    item["condition"],
		"noDataState":  item["noDataState"],
		"execErrState": item["execErrState"],
		"for":          item["for"],
	}
}
