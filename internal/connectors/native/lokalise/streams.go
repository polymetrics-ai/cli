package lokalise

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Lokalise project-scoped resource path
// segment it reads from, the JSON key under which the array of records lives in
// the response body, the record's primary key field, and the record mapper.
//
// Lokalise list endpoints are all shaped as
// projects/{project_id}/{resource} and return {"<recordsKey>":[...]} with
// offset pagination (page/limit) reported via the X-Pagination-Page-Count and
// X-Pagination-Page response headers.
type streamEndpoint struct {
	// resource is the path segment under projects/{project_id}/ (e.g. "keys").
	resource string
	// recordsKey is the top-level JSON key holding the records array.
	recordsKey string
	// mapRecord flattens a raw Lokalise object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// lokaliseStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in lokaliseStreams; the read
// path is fully data-driven from this table.
var lokaliseStreamEndpoints = map[string]streamEndpoint{
	"keys":         {resource: "keys", recordsKey: "keys", mapRecord: lokaliseKeyRecord},
	"languages":    {resource: "languages", recordsKey: "languages", mapRecord: lokaliseLanguageRecord},
	"translations": {resource: "translations", recordsKey: "translations", mapRecord: lokaliseTranslationRecord},
	"contributors": {resource: "contributors", recordsKey: "contributors", mapRecord: lokaliseContributorRecord},
	"comments":     {resource: "comments", recordsKey: "comments", mapRecord: lokaliseCommentRecord},
}

// lokaliseStreams returns the connector's published stream catalog.
func lokaliseStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "keys",
			Description:  "Lokalise translation keys for the project.",
			PrimaryKey:   []string{"key_id"},
			CursorFields: []string{"modified_at_timestamp"},
			Fields:       lokaliseKeyFields(),
		},
		{
			Name:        "languages",
			Description: "Languages configured on the project.",
			PrimaryKey:  []string{"lang_id"},
			Fields:      lokaliseLanguageFields(),
		},
		{
			Name:         "translations",
			Description:  "Translations for the project's keys and languages.",
			PrimaryKey:   []string{"translation_id"},
			CursorFields: []string{"modified_at_timestamp"},
			Fields:       lokaliseTranslationFields(),
		},
		{
			Name:        "contributors",
			Description: "Users with access to the project.",
			PrimaryKey:  []string{"user_id"},
			Fields:      lokaliseContributorFields(),
		},
		{
			Name:        "comments",
			Description: "Comments attached to the project's keys.",
			PrimaryKey:  []string{"comment_id"},
			Fields:      lokaliseCommentFields(),
		},
	}
}

func lokaliseKeyFields() []connectors.Field {
	return []connectors.Field{
		{Name: "key_id", Type: "integer"},
		{Name: "created_at", Type: "string"},
		{Name: "created_at_timestamp", Type: "integer"},
		{Name: "modified_at", Type: "string"},
		{Name: "modified_at_timestamp", Type: "integer"},
		{Name: "key_name", Type: "object"},
		{Name: "description", Type: "string"},
		{Name: "platforms", Type: "array"},
		{Name: "tags", Type: "array"},
		{Name: "is_plural", Type: "boolean"},
		{Name: "is_hidden", Type: "boolean"},
		{Name: "is_archived", Type: "boolean"},
	}
}

func lokaliseLanguageFields() []connectors.Field {
	return []connectors.Field{
		{Name: "lang_id", Type: "integer"},
		{Name: "lang_iso", Type: "string"},
		{Name: "lang_name", Type: "string"},
		{Name: "is_rtl", Type: "boolean"},
		{Name: "plural_forms", Type: "array"},
	}
}

func lokaliseTranslationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "translation_id", Type: "integer"},
		{Name: "key_id", Type: "integer"},
		{Name: "language_iso", Type: "string"},
		{Name: "translation", Type: "string"},
		{Name: "modified_at", Type: "string"},
		{Name: "modified_at_timestamp", Type: "integer"},
		{Name: "modified_by", Type: "integer"},
		{Name: "modified_by_email", Type: "string"},
		{Name: "is_reviewed", Type: "boolean"},
		{Name: "is_unverified", Type: "boolean"},
		{Name: "reviewed_by", Type: "integer"},
	}
}

func lokaliseContributorFields() []connectors.Field {
	return []connectors.Field{
		{Name: "user_id", Type: "integer"},
		{Name: "email", Type: "string"},
		{Name: "fullname", Type: "string"},
		{Name: "created_at", Type: "string"},
		{Name: "created_at_timestamp", Type: "integer"},
		{Name: "is_admin", Type: "boolean"},
		{Name: "is_reviewer", Type: "boolean"},
		{Name: "languages", Type: "array"},
		{Name: "role_id", Type: "integer"},
	}
}

func lokaliseCommentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "comment_id", Type: "integer"},
		{Name: "key_id", Type: "integer"},
		{Name: "comment", Type: "string"},
		{Name: "added_by", Type: "integer"},
		{Name: "added_by_email", Type: "string"},
		{Name: "added_at", Type: "string"},
		{Name: "added_at_timestamp", Type: "integer"},
	}
}

func lokaliseKeyRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"key_id":                item["key_id"],
		"created_at":            item["created_at"],
		"created_at_timestamp":  item["created_at_timestamp"],
		"modified_at":           item["modified_at"],
		"modified_at_timestamp": item["modified_at_timestamp"],
		"key_name":              item["key_name"],
		"description":           item["description"],
		"platforms":             item["platforms"],
		"tags":                  item["tags"],
		"is_plural":             item["is_plural"],
		"is_hidden":             item["is_hidden"],
		"is_archived":           item["is_archived"],
	}
}

func lokaliseLanguageRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"lang_id":      item["lang_id"],
		"lang_iso":     item["lang_iso"],
		"lang_name":    item["lang_name"],
		"is_rtl":       item["is_rtl"],
		"plural_forms": item["plural_forms"],
	}
}

func lokaliseTranslationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"translation_id":        item["translation_id"],
		"key_id":                item["key_id"],
		"language_iso":          item["language_iso"],
		"translation":           item["translation"],
		"modified_at":           item["modified_at"],
		"modified_at_timestamp": item["modified_at_timestamp"],
		"modified_by":           item["modified_by"],
		"modified_by_email":     item["modified_by_email"],
		"is_reviewed":           item["is_reviewed"],
		"is_unverified":         item["is_unverified"],
		"reviewed_by":           item["reviewed_by"],
	}
}

func lokaliseContributorRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"user_id":              item["user_id"],
		"email":                item["email"],
		"fullname":             item["fullname"],
		"created_at":           item["created_at"],
		"created_at_timestamp": item["created_at_timestamp"],
		"is_admin":             item["is_admin"],
		"is_reviewer":          item["is_reviewer"],
		"languages":            item["languages"],
		"role_id":              item["role_id"],
	}
}

func lokaliseCommentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"comment_id":         item["comment_id"],
		"key_id":             item["key_id"],
		"comment":            item["comment"],
		"added_by":           item["added_by"],
		"added_by_email":     item["added_by_email"],
		"added_at":           item["added_at"],
		"added_at_timestamp": item["added_at_timestamp"],
	}
}
