package gainsightpx

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Gainsight PX API resource path
// (relative to base_url), the JSON key under which the response nests its record
// array (the manifest's DpathExtractor field_path), and the record mapper that
// flattens its objects.
type streamEndpoint struct {
	// resource is the API list endpoint path segment (e.g. "accounts").
	resource string
	// recordsKey is the top-level JSON key holding the record array
	// (e.g. "accounts", "users", "features", "segments").
	recordsKey string
	// mapRecord flattens a raw object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// gainsightStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in gainsightStreams; the read
// path is fully data-driven from this table.
//
// Endpoints, record keys, and primary keys come from the Airbyte
// source-gainsight-px declarative manifest (api.aptrinsic.com/v1).
var gainsightStreamEndpoints = map[string]streamEndpoint{
	"accounts": {resource: "accounts", recordsKey: "accounts", mapRecord: gainsightAccountRecord},
	"users":    {resource: "users", recordsKey: "users", mapRecord: gainsightUserRecord},
	"feature":  {resource: "feature", recordsKey: "features", mapRecord: gainsightFeatureRecord},
	"segments": {resource: "segment", recordsKey: "segments", mapRecord: gainsightSegmentRecord},
}

// gainsightStreams returns the connector's published stream catalog. Every
// Gainsight PX object exposes a string id, so the primary key is ["id"] across
// the board. The API supports only full-refresh sync, so no incremental cursor
// fields are published.
func gainsightStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "accounts",
			Description: "Gainsight PX accounts (companies tracked in PX).",
			PrimaryKey:  []string{"id"},
			Fields:      gainsightAccountFields(),
		},
		{
			Name:        "users",
			Description: "Gainsight PX users (end users tracked in PX).",
			PrimaryKey:  []string{"id"},
			Fields:      gainsightUserFields(),
		},
		{
			Name:        "feature",
			Description: "Gainsight PX features and modules.",
			PrimaryKey:  []string{"id"},
			Fields:      gainsightFeatureFields(),
		},
		{
			Name:        "segments",
			Description: "Gainsight PX segments.",
			PrimaryKey:  []string{"id"},
			Fields:      gainsightSegmentFields(),
		},
	}
}

func gainsightAccountFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "trackedSubscriptionId", Type: "string"},
		{Name: "sfdcId", Type: "string"},
		{Name: "industry", Type: "string"},
		{Name: "numberOfEmployees", Type: "string"},
		{Name: "numberOfUsers", Type: "string"},
		{Name: "plan", Type: "string"},
		{Name: "location", Type: "string"},
		{Name: "website", Type: "string"},
		{Name: "createDate", Type: "string"},
		{Name: "lastModifiedDate", Type: "string"},
		{Name: "lastSeenDate", Type: "string"},
	}
}

func gainsightUserFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "accountId", Type: "string"},
		{Name: "aptrinsicId", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "firstName", Type: "string"},
		{Name: "lastName", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "role", Type: "string"},
		{Name: "score", Type: "number"},
		{Name: "numberOfVisits", Type: "integer"},
		{Name: "signUpDate", Type: "integer"},
		{Name: "createDate", Type: "integer"},
		{Name: "lastSeenDate", Type: "integer"},
		{Name: "lastModifiedDate", Type: "integer"},
	}
}

func gainsightFeatureFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "parentFeatureId", Type: "string"},
		{Name: "propertyKey", Type: "string"},
		{Name: "status", Type: "string"},
	}
}

func gainsightSegmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "priority", Type: "string"},
		{Name: "productId", Type: "string"},
		{Name: "productName", Type: "string"},
		{Name: "createdBy", Type: "string"},
		{Name: "createdDate", Type: "string"},
		{Name: "modifiedBy", Type: "string"},
		{Name: "modifiedDate", Type: "string"},
	}
}

func gainsightAccountRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                    item["id"],
		"name":                  item["name"],
		"trackedSubscriptionId": item["trackedSubscriptionId"],
		"sfdcId":                item["sfdcId"],
		"industry":              item["industry"],
		"numberOfEmployees":     item["numberOfEmployees"],
		"numberOfUsers":         item["numberOfUsers"],
		"plan":                  item["plan"],
		"location":              item["location"],
		"website":               item["website"],
		"createDate":            item["createDate"],
		"lastModifiedDate":      item["lastModifiedDate"],
		"lastSeenDate":          item["lastSeenDate"],
	}
}

func gainsightUserRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"type":             item["type"],
		"accountId":        item["accountId"],
		"aptrinsicId":      item["aptrinsicId"],
		"email":            item["email"],
		"firstName":        item["firstName"],
		"lastName":         item["lastName"],
		"title":            item["title"],
		"role":             item["role"],
		"score":            item["score"],
		"numberOfVisits":   item["numberOfVisits"],
		"signUpDate":       item["signUpDate"],
		"createDate":       item["createDate"],
		"lastSeenDate":     item["lastSeenDate"],
		"lastModifiedDate": item["lastModifiedDate"],
	}
}

func gainsightFeatureRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["id"],
		"name":            item["name"],
		"type":            item["type"],
		"parentFeatureId": item["parentFeatureId"],
		"propertyKey":     item["propertyKey"],
		"status":          item["status"],
	}
}

func gainsightSegmentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":           item["id"],
		"name":         item["name"],
		"description":  item["description"],
		"status":       item["status"],
		"priority":     item["priority"],
		"productId":    item["productId"],
		"productName":  item["productName"],
		"createdBy":    item["createdBy"],
		"createdDate":  item["createdDate"],
		"modifiedBy":   item["modifiedBy"],
		"modifiedDate": item["modifiedDate"],
	}
}
