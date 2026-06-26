package googleforms

import "polymetrics.ai/internal/connectors"

// Stream names published by the connector.
const (
	streamForms     = "forms"
	streamResponses = "responses"
	streamFormItems = "form_items"
)

// googleFormsStreams returns the connector's published stream catalog.
//
//   - forms: one record per configured form_id (form metadata).
//   - responses: every submitted response across the configured forms.
//   - form_items: the questions/items that make up each configured form.
func googleFormsStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         streamForms,
			Description:  "Google Forms metadata, one record per configured form.",
			PrimaryKey:   []string{"form_id"},
			CursorFields: nil,
			Fields:       formFields(),
		},
		{
			Name:         streamResponses,
			Description:  "Submitted responses for the configured forms.",
			PrimaryKey:   []string{"response_id"},
			CursorFields: []string{"last_submitted_time"},
			Fields:       responseFields(),
		},
		{
			Name:         streamFormItems,
			Description:  "Questions and items that make up the configured forms.",
			PrimaryKey:   []string{"form_id", "item_id"},
			CursorFields: nil,
			Fields:       formItemFields(),
		},
	}
}

func formFields() []connectors.Field {
	return []connectors.Field{
		{Name: "form_id", Type: "string"},
		{Name: "revision_id", Type: "string"},
		{Name: "responder_uri", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "document_title", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "item_count", Type: "integer"},
	}
}

func responseFields() []connectors.Field {
	return []connectors.Field{
		{Name: "response_id", Type: "string"},
		{Name: "form_id", Type: "string"},
		{Name: "create_time", Type: "timestamp"},
		{Name: "last_submitted_time", Type: "timestamp"},
		{Name: "respondent_email", Type: "string"},
		{Name: "total_score", Type: "number"},
		{Name: "answers", Type: "object"},
	}
}

func formItemFields() []connectors.Field {
	return []connectors.Field{
		{Name: "form_id", Type: "string"},
		{Name: "item_id", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "question_id", Type: "string"},
	}
}

// mapResponseRecord flattens a Forms API FormResponse object into a record. The
// form_id is supplied by the caller because list responses omit it.
func mapResponseRecord(formID string, item map[string]any) connectors.Record {
	rec := connectors.Record{
		"response_id":         item["responseId"],
		"form_id":             formID,
		"create_time":         item["createTime"],
		"last_submitted_time": item["lastSubmittedTime"],
		"respondent_email":    item["respondentEmail"],
		"total_score":         item["totalScore"],
		"answers":             item["answers"],
	}
	if rec["form_id"] == nil || rec["form_id"] == "" {
		rec["form_id"] = item["formId"]
	}
	return rec
}

// mapFormRecord flattens a Forms API Form object into a record.
func mapFormRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"form_id":       item["formId"],
		"revision_id":   item["revisionId"],
		"responder_uri": item["responderUri"],
	}
	if info, ok := item["info"].(map[string]any); ok {
		rec["title"] = info["title"]
		rec["document_title"] = info["documentTitle"]
		rec["description"] = info["description"]
	}
	if items, ok := item["items"].([]any); ok {
		rec["item_count"] = len(items)
	}
	return rec
}

// mapFormItemRecords expands a Form object's items[] into one record per item.
func mapFormItemRecords(formID string, item map[string]any) []connectors.Record {
	rawItems, ok := item["items"].([]any)
	if !ok {
		return nil
	}
	out := make([]connectors.Record, 0, len(rawItems))
	for _, raw := range rawItems {
		obj, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		rec := connectors.Record{
			"form_id":     formID,
			"item_id":     obj["itemId"],
			"title":       obj["title"],
			"description": obj["description"],
		}
		// questionItem.question.questionId, when present, identifies the answer key.
		if qi, ok := obj["questionItem"].(map[string]any); ok {
			if q, ok := qi["question"].(map[string]any); ok {
				rec["question_id"] = q["questionId"]
			}
		}
		out = append(out, rec)
	}
	return out
}
