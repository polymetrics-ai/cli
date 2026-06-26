package fillout

import "polymetrics.ai/internal/connectors"

// filloutStreams returns the connector's published stream catalog.
//
//   - forms:       one record per form in the account (GET /forms).
//   - questions:   the question definitions for every form (GET /forms/{id}).
//   - submissions: the form responses, fanned out across every form
//     (GET /forms/{id}/submissions), incremental on submissionTime.
func filloutStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "forms",
			Description: "Fillout forms in the account.",
			PrimaryKey:  []string{"id"},
			Fields:      filloutFormFields(),
		},
		{
			Name:        "questions",
			Description: "Question definitions for each Fillout form.",
			PrimaryKey:  []string{"form_id", "id"},
			Fields:      filloutQuestionFields(),
		},
		{
			Name:         "submissions",
			Description:  "Form responses (submissions) across all Fillout forms.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"submissionTime"},
			Fields:       filloutSubmissionFields(),
		},
	}
}

func filloutFormFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
	}
}

func filloutQuestionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "form_id", Type: "string"},
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
	}
}

func filloutSubmissionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "form_id", Type: "string"},
		{Name: "submissionTime", Type: "string"},
		{Name: "lastUpdatedAt", Type: "string"},
		{Name: "questions", Type: "array"},
		{Name: "calculations", Type: "array"},
		{Name: "urlParameters", Type: "array"},
		{Name: "scheduling", Type: "array"},
		{Name: "payments", Type: "array"},
	}
}

// filloutFormRecord maps a raw /forms item (formId, name) onto a Record keyed on
// a normalized "id".
func filloutFormRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":   firstString(item, "formId", "id"),
		"name": item["name"],
	}
}

// filloutQuestionRecord maps a raw question object from a form's "questions"
// array, tagging it with the owning form id.
func filloutQuestionRecord(formID string, item map[string]any) connectors.Record {
	return connectors.Record{
		"form_id": formID,
		"id":      item["id"],
		"name":    item["name"],
		"type":    item["type"],
	}
}

// filloutSubmissionRecord maps a raw submission ("responses" array element),
// normalizing submissionId -> id and attaching the owning form id.
func filloutSubmissionRecord(formID string, item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             firstString(item, "submissionId", "id"),
		"form_id":        formID,
		"submissionTime": item["submissionTime"],
		"lastUpdatedAt":  item["lastUpdatedAt"],
		"questions":      item["questions"],
		"calculations":   item["calculations"],
		"urlParameters":  item["urlParameters"],
		"scheduling":     item["scheduling"],
		"payments":       item["payments"],
	}
}

// firstString returns the first non-empty string value among the given keys.
func firstString(item map[string]any, keys ...string) any {
	for _, k := range keys {
		if v, ok := item[k]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
			if v != nil {
				return v
			}
		}
	}
	return nil
}
