package typeform

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the Typeform API resource path it reads
// from and the record mapper that flattens its objects. perForm marks streams
// (responses) that are scoped to a form id and fan out across the configured
// form_ids rather than hitting a single collection endpoint.
type streamEndpoint struct {
	// resource is the collection path segment (e.g. "forms", "workspaces").
	// For per-form streams it is the trailing segment (e.g. "responses") that
	// hangs off /forms/{form_id}/.
	resource string
	// perForm is true for streams read per form id (responses).
	perForm bool
	// mapRecord flattens a raw Typeform object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// typeformStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in typeformStreams; the read
// path is fully data-driven from this table.
var typeformStreamEndpoints = map[string]streamEndpoint{
	"forms":      {resource: "forms", mapRecord: typeformFormRecord},
	"responses":  {resource: "responses", perForm: true, mapRecord: typeformResponseRecord},
	"workspaces": {resource: "workspaces", mapRecord: typeformWorkspaceRecord},
	"themes":     {resource: "themes", mapRecord: typeformThemeRecord},
	"images":     {resource: "images", mapRecord: typeformImageRecord},
}

// typeformStreams returns the connector's published stream catalog.
func typeformStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "forms",
			Description:  "Typeform forms in the account.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"last_updated_at"},
			Fields:       typeformFormFields(),
		},
		{
			Name:         "responses",
			Description:  "Submitted responses, read per configured form id.",
			PrimaryKey:   []string{"response_id"},
			CursorFields: []string{"submitted_at"},
			Fields:       typeformResponseFields(),
		},
		{
			Name:        "workspaces",
			Description: "Workspaces that organize forms.",
			PrimaryKey:  []string{"id"},
			Fields:      typeformWorkspaceFields(),
		},
		{
			Name:        "themes",
			Description: "Themes available to the account.",
			PrimaryKey:  []string{"id"},
			Fields:      typeformThemeFields(),
		},
		{
			Name:        "images",
			Description: "Images in the account image collection.",
			PrimaryKey:  []string{"id"},
			Fields:      typeformImageFields(),
		},
	}
}

func typeformFormFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "created_at", Type: "timestamp"},
		{Name: "last_updated_at", Type: "timestamp"},
		{Name: "is_public", Type: "boolean"},
		{Name: "theme_href", Type: "string"},
		{Name: "self_href", Type: "string"},
	}
}

func typeformResponseFields() []connectors.Field {
	return []connectors.Field{
		{Name: "response_id", Type: "string"},
		{Name: "form_id", Type: "string"},
		{Name: "token", Type: "string"},
		{Name: "landing_id", Type: "string"},
		{Name: "landed_at", Type: "timestamp"},
		{Name: "submitted_at", Type: "timestamp"},
		{Name: "answers", Type: "array"},
		{Name: "hidden", Type: "object"},
		{Name: "calculated", Type: "object"},
		{Name: "metadata", Type: "object"},
	}
}

func typeformWorkspaceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "account_id", Type: "string"},
		{Name: "shared", Type: "boolean"},
		{Name: "default", Type: "boolean"},
		{Name: "self_href", Type: "string"},
	}
}

func typeformThemeFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "visibility", Type: "string"},
		{Name: "background", Type: "object"},
		{Name: "colors", Type: "object"},
		{Name: "font", Type: "string"},
	}
}

func typeformImageFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "file_name", Type: "string"},
		{Name: "src", Type: "string"},
		{Name: "media_type", Type: "string"},
		{Name: "has_alpha", Type: "boolean"},
		{Name: "width", Type: "integer"},
		{Name: "height", Type: "integer"},
	}
}

func typeformFormRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":              item["id"],
		"title":           item["title"],
		"type":            item["type"],
		"created_at":      item["created_at"],
		"last_updated_at": item["last_updated_at"],
	}
	if settings, ok := item["settings"].(map[string]any); ok {
		rec["is_public"] = settings["is_public"]
	}
	if theme, ok := item["theme"].(map[string]any); ok {
		rec["theme_href"] = theme["href"]
	}
	if links, ok := item["_links"].(map[string]any); ok {
		rec["self_href"] = links["display"]
	}
	if self, ok := item["self"].(map[string]any); ok {
		rec["self_href"] = self["href"]
	}
	return rec
}

func typeformResponseRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"response_id":  item["response_id"],
		"form_id":      item["form_id"],
		"token":        item["token"],
		"landing_id":   item["landing_id"],
		"landed_at":    item["landed_at"],
		"submitted_at": item["submitted_at"],
		"answers":      item["answers"],
		"hidden":       item["hidden"],
		"calculated":   item["calculated"],
		"metadata":     item["metadata"],
	}
}

func typeformWorkspaceRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"account_id": item["account_id"],
		"shared":     item["shared"],
		"default":    item["default"],
	}
	if self, ok := item["self"].(map[string]any); ok {
		rec["self_href"] = self["href"]
	}
	return rec
}

func typeformThemeRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"name":       item["name"],
		"visibility": item["visibility"],
		"background": item["background"],
		"colors":     item["colors"],
		"font":       item["font"],
	}
}

func typeformImageRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":         item["id"],
		"file_name":  item["file_name"],
		"src":        item["src"],
		"media_type": item["media_type"],
		"has_alpha":  item["has_alpha"],
		"width":      item["width"],
		"height":     item["height"],
	}
}
