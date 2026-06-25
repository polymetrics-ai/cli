package configcat

import "polymetrics.ai/internal/connectors"

// streamDef describes how to read one ConfigCat stream. ConfigCat's Public
// Management API returns each list as a top-level JSON array (no pagination, no
// wrapping), so the record path is always "" (root). Some resources are nested
// under a product (configs, environments, tags); those set nestedUnderProduct so
// the read path fans the call out across every product id.
type streamDef struct {
	name string
	// path is the absolute API path. For nested streams it contains a single
	// "%s" placeholder for the product id.
	path string
	// nestedUnderProduct, when true, makes Read first list products then call
	// path for each product id, annotating each record with product_id.
	nestedUnderProduct bool
	mapRecord          func(map[string]any) connectors.Record
}

// configcatStreamDefs is the per-stream routing table. Adding a stream means one
// entry here plus a Stream definition in configcatStreams.
var configcatStreamDefs = map[string]streamDef{
	"organizations": {name: "organizations", path: "/v1/organizations", mapRecord: organizationRecord},
	"products":      {name: "products", path: "/v1/products", mapRecord: productRecord},
	"configs":       {name: "configs", path: "/v1/products/%s/configs", nestedUnderProduct: true, mapRecord: configRecord},
	"environments":  {name: "environments", path: "/v1/products/%s/environments", nestedUnderProduct: true, mapRecord: environmentRecord},
	"tags":          {name: "tags", path: "/v1/products/%s/tags", nestedUnderProduct: true, mapRecord: tagRecord},
}

// configcatStreams returns the connector's published stream catalog.
func configcatStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "organizations",
			Description: "ConfigCat organizations the credentials can access.",
			PrimaryKey:  []string{"organization_id"},
			Fields:      organizationFields(),
		},
		{
			Name:        "products",
			Description: "ConfigCat products across accessible organizations.",
			PrimaryKey:  []string{"product_id"},
			Fields:      productFields(),
		},
		{
			Name:        "configs",
			Description: "ConfigCat configs, fanned out across every accessible product.",
			PrimaryKey:  []string{"config_id"},
			Fields:      configFields(),
		},
		{
			Name:        "environments",
			Description: "ConfigCat environments, fanned out across every accessible product.",
			PrimaryKey:  []string{"environment_id"},
			Fields:      environmentFields(),
		},
		{
			Name:        "tags",
			Description: "ConfigCat tags, fanned out across every accessible product.",
			PrimaryKey:  []string{"tag_id"},
			Fields:      tagFields(),
		},
	}
}

func organizationFields() []connectors.Field {
	return []connectors.Field{
		{Name: "organization_id", Type: "string"},
		{Name: "name", Type: "string"},
	}
}

func productFields() []connectors.Field {
	return []connectors.Field{
		{Name: "product_id", Type: "string"},
		{Name: "organization_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "order", Type: "integer"},
		{Name: "reason_required", Type: "boolean"},
		{Name: "approve_required", Type: "boolean"},
	}
}

func configFields() []connectors.Field {
	return []connectors.Field{
		{Name: "config_id", Type: "string"},
		{Name: "product_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "order", Type: "integer"},
		{Name: "evaluation_version", Type: "string"},
		{Name: "migrated_config_id", Type: "string"},
	}
}

func environmentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "environment_id", Type: "string"},
		{Name: "product_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "color", Type: "string"},
		{Name: "description", Type: "string"},
		{Name: "order", Type: "integer"},
		{Name: "reason_required", Type: "boolean"},
		{Name: "approve_required", Type: "boolean"},
	}
}

func tagFields() []connectors.Field {
	return []connectors.Field{
		{Name: "tag_id", Type: "integer"},
		{Name: "product_id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "color", Type: "string"},
	}
}

func organizationRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"organization_id": item["organizationId"],
		"name":            item["name"],
	}
}

func productRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"product_id":       item["productId"],
		"name":             item["name"],
		"description":      item["description"],
		"order":            item["order"],
		"reason_required":  item["reasonRequired"],
		"approve_required": item["approveRequired"],
	}
	if org, ok := item["organization"].(map[string]any); ok {
		rec["organization_id"] = org["organizationId"]
	}
	return rec
}

func configRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"config_id":          item["configId"],
		"name":               item["name"],
		"description":        item["description"],
		"order":              item["order"],
		"evaluation_version": item["evaluationVersion"],
		"migrated_config_id": item["migratedConfigId"],
	}
	if prod, ok := item["product"].(map[string]any); ok {
		rec["product_id"] = prod["productId"]
	}
	return rec
}

func environmentRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"environment_id":   item["environmentId"],
		"name":             item["name"],
		"color":            item["color"],
		"description":      item["description"],
		"order":            item["order"],
		"reason_required":  item["reasonRequired"],
		"approve_required": item["approveRequired"],
	}
	if prod, ok := item["product"].(map[string]any); ok {
		rec["product_id"] = prod["productId"]
	}
	return rec
}

func tagRecord(item map[string]any) connectors.Record {
	rec := connectors.Record{
		"tag_id": item["tagId"],
		"name":   item["name"],
		"color":  item["color"],
	}
	if prod, ok := item["product"].(map[string]any); ok {
		rec["product_id"] = prod["productId"]
	}
	return rec
}
