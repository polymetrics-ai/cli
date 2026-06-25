package boldsign

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the BoldSign API resource path (relative
// to base_url), the JSON path to the records array in the list response, and the
// record mapper that flattens its objects.
//
// BoldSign list endpoints are page-numbered (Page + PageSize) and wrap the
// records in a "result" envelope. The teams endpoint is the documented
// exception: it uses "results". recordsPath captures that per-stream difference.
type streamEndpoint struct {
	// resource is the list endpoint path (e.g. "v1/document/list").
	resource string
	// recordsPath is the dotted JSON path to the records array (e.g. "result").
	recordsPath string
	// mapRecord flattens a raw BoldSign object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// boldsignStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in boldsignStreams; the read
// path is fully data-driven from this table.
var boldsignStreamEndpoints = map[string]streamEndpoint{
	"documents": {resource: "v1/document/list", recordsPath: "result", mapRecord: boldsignDocumentRecord},
	"templates": {resource: "v1/template/list", recordsPath: "result", mapRecord: boldsignTemplateRecord},
	"teams":     {resource: "v1/teams/list", recordsPath: "results", mapRecord: boldsignTeamRecord},
	"contacts":  {resource: "v1/contacts/list", recordsPath: "result", mapRecord: boldsignContactRecord},
	"brands":    {resource: "v1/brand/list", recordsPath: "result", mapRecord: boldsignBrandRecord},
}

// boldsignStreams returns the connector's published stream catalog.
func boldsignStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "documents",
			Description:  "BoldSign documents (e-signature requests) and their signing status.",
			PrimaryKey:   []string{"document_id"},
			CursorFields: []string{"created_date"},
			Fields:       boldsignDocumentFields(),
		},
		{
			Name:         "templates",
			Description:  "BoldSign reusable document templates.",
			PrimaryKey:   []string{"document_id"},
			CursorFields: []string{"created_date"},
			Fields:       boldsignTemplateFields(),
		},
		{
			Name:         "teams",
			Description:  "BoldSign teams.",
			PrimaryKey:   []string{"team_id"},
			CursorFields: []string{"created_date"},
			Fields:       boldsignTeamFields(),
		},
		{
			Name:        "contacts",
			Description: "BoldSign contacts (recipients address book).",
			PrimaryKey:  []string{"id"},
			Fields:      boldsignContactFields(),
		},
		{
			Name:        "brands",
			Description: "BoldSign branding profiles.",
			PrimaryKey:  []string{"brand_id"},
			Fields:      boldsignBrandFields(),
		},
	}
}

func boldsignDocumentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "document_id", Type: "string"},
		{Name: "sender_email", Type: "string"},
		{Name: "message_title", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "created_date", Type: "string"},
		{Name: "expiry_date", Type: "string"},
		{Name: "is_deleted", Type: "boolean"},
		{Name: "enable_signing_order", Type: "boolean"},
		{Name: "sender_detail", Type: "object"},
		{Name: "signer_details", Type: "array"},
		{Name: "labels", Type: "array"},
	}
}

func boldsignTemplateFields() []connectors.Field {
	return []connectors.Field{
		{Name: "document_id", Type: "string"},
		{Name: "template_name", Type: "string"},
		{Name: "template_description", Type: "string"},
		{Name: "created_date", Type: "string"},
		{Name: "sender_email", Type: "string"},
		{Name: "is_shared_template", Type: "boolean"},
		{Name: "labels", Type: "array"},
	}
}

func boldsignTeamFields() []connectors.Field {
	return []connectors.Field{
		{Name: "team_id", Type: "string"},
		{Name: "team_name", Type: "string"},
		{Name: "created_date", Type: "string"},
		{Name: "users", Type: "array"},
	}
}

func boldsignContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone_number", Type: "string"},
		{Name: "company_name", Type: "string"},
	}
}

func boldsignBrandFields() []connectors.Field {
	return []connectors.Field{
		{Name: "brand_id", Type: "string"},
		{Name: "brand_name", Type: "string"},
		{Name: "is_default", Type: "boolean"},
		{Name: "background_color", Type: "string"},
		{Name: "button_color", Type: "string"},
	}
}

func boldsignDocumentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"document_id":          firstString(item, "documentId", "documentID"),
		"sender_email":         item["senderEmail"],
		"message_title":        item["messageTitle"],
		"status":               item["status"],
		"created_date":         item["createdDate"],
		"expiry_date":          item["expiryDate"],
		"is_deleted":           item["isDeleted"],
		"enable_signing_order": item["enableSigningOrder"],
		"sender_detail":        item["senderDetail"],
		"signer_details":       item["signerDetails"],
		"labels":               item["labels"],
	}
}

func boldsignTemplateRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"document_id":          firstString(item, "documentId", "documentID"),
		"template_name":        item["templateName"],
		"template_description": item["templateDescription"],
		"created_date":         item["createdDate"],
		"sender_email":         item["senderEmail"],
		"is_shared_template":   item["isSharedTemplate"],
		"labels":               item["labels"],
	}
}

func boldsignTeamRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"team_id":      firstString(item, "teamId", "teamID"),
		"team_name":    item["teamName"],
		"created_date": item["createdDate"],
		"users":        item["users"],
	}
}

func boldsignContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           firstString(item, "id", "contactId"),
		"name":         item["name"],
		"email":        item["email"],
		"phone_number": item["phoneNumber"],
		"company_name": item["companyName"],
	}
}

func boldsignBrandRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"brand_id":         firstString(item, "brandId", "brandID"),
		"brand_name":       item["brandName"],
		"is_default":       item["isDefault"],
		"background_color": item["backgroundColor"],
		"button_color":     item["buttonColor"],
	}
}

// firstString returns the first present non-nil value among keys as a value,
// accommodating BoldSign's mixed-case id field names across endpoints.
func firstString(item map[string]any, keys ...string) any {
	for _, k := range keys {
		if v, ok := item[k]; ok && v != nil {
			return v
		}
	}
	return nil
}
