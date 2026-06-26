package dolibarr

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Dolibarr REST resource path (relative
// to base_url, e.g. "thirdparties") and the record mapper that flattens its
// objects into a connectors.Record. The read path is fully data-driven from this
// table: adding a stream means adding one entry here plus a Stream definition in
// dolibarrStreams.
type streamEndpoint struct {
	resource  string
	mapRecord func(map[string]any) connectors.Record
}

// dolibarrStreamEndpoints is the per-stream routing table for the core set of
// Dolibarr objects most useful for analytics (third parties / customers,
// contacts, products, customer invoices, and sales orders).
var dolibarrStreamEndpoints = map[string]streamEndpoint{
	"thirdparties": {resource: "thirdparties", mapRecord: dolibarrThirdPartyRecord},
	"contacts":     {resource: "contacts", mapRecord: dolibarrContactRecord},
	"products":     {resource: "products", mapRecord: dolibarrProductRecord},
	"invoices":     {resource: "invoices", mapRecord: dolibarrInvoiceRecord},
	"orders":       {resource: "orders", mapRecord: dolibarrOrderRecord},
}

// dolibarrStreams returns the connector's published stream catalog. Dolibarr
// objects are keyed by a string "id" (rowid) and most expose a unix
// "date_modification" timestamp used as the incremental cursor where present.
func dolibarrStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "thirdparties",
			Description:  "Dolibarr third parties (customers, prospects, and suppliers).",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_modification"},
			Fields:       dolibarrThirdPartyFields(),
		},
		{
			Name:         "contacts",
			Description:  "Dolibarr contacts/addresses attached to third parties.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_modification"},
			Fields:       dolibarrContactFields(),
		},
		{
			Name:         "products",
			Description:  "Dolibarr products and services.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_modification"},
			Fields:       dolibarrProductFields(),
		},
		{
			Name:         "invoices",
			Description:  "Dolibarr customer invoices.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_modification"},
			Fields:       dolibarrInvoiceFields(),
		},
		{
			Name:         "orders",
			Description:  "Dolibarr customer (sales) orders.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_modification"},
			Fields:       dolibarrOrderFields(),
		},
	}
}

func dolibarrThirdPartyFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "name_alias", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "town", Type: "string"},
		{Name: "zip", Type: "string"},
		{Name: "country_code", Type: "string"},
		{Name: "client", Type: "string"},
		{Name: "fournisseur", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "date_creation", Type: "integer"},
		{Name: "date_modification", Type: "integer"},
	}
}

func dolibarrContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "socid", Type: "string"},
		{Name: "lastname", Type: "string"},
		{Name: "firstname", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone_pro", Type: "string"},
		{Name: "phone_mobile", Type: "string"},
		{Name: "town", Type: "string"},
		{Name: "zip", Type: "string"},
		{Name: "country_code", Type: "string"},
		{Name: "statut", Type: "string"},
		{Name: "date_creation", Type: "integer"},
		{Name: "date_modification", Type: "integer"},
	}
}

func dolibarrProductFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "ref", Type: "string"},
		{Name: "label", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "price", Type: "string"},
		{Name: "price_ttc", Type: "string"},
		{Name: "tva_tx", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "status_buy", Type: "string"},
		{Name: "stock_reel", Type: "string"},
		{Name: "date_creation", Type: "integer"},
		{Name: "date_modification", Type: "integer"},
	}
}

func dolibarrInvoiceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "ref", Type: "string"},
		{Name: "socid", Type: "string"},
		{Name: "type", Type: "string"},
		{Name: "total_ht", Type: "string"},
		{Name: "total_tva", Type: "string"},
		{Name: "total_ttc", Type: "string"},
		{Name: "paye", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "date", Type: "integer"},
		{Name: "date_creation", Type: "integer"},
		{Name: "date_modification", Type: "integer"},
	}
}

func dolibarrOrderFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "ref", Type: "string"},
		{Name: "socid", Type: "string"},
		{Name: "total_ht", Type: "string"},
		{Name: "total_tva", Type: "string"},
		{Name: "total_ttc", Type: "string"},
		{Name: "billed", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "date", Type: "integer"},
		{Name: "date_creation", Type: "integer"},
		{Name: "date_modification", Type: "integer"},
	}
}

func dolibarrThirdPartyRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"name":              item["name"],
		"name_alias":        item["name_alias"],
		"email":             item["email"],
		"phone":             item["phone"],
		"town":              item["town"],
		"zip":               item["zip"],
		"country_code":      item["country_code"],
		"client":            item["client"],
		"fournisseur":       item["fournisseur"],
		"status":            item["status"],
		"date_creation":     item["date_creation"],
		"date_modification": item["date_modification"],
	}
}

func dolibarrContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"socid":             item["socid"],
		"lastname":          item["lastname"],
		"firstname":         item["firstname"],
		"email":             item["email"],
		"phone_pro":         item["phone_pro"],
		"phone_mobile":      item["phone_mobile"],
		"town":              item["town"],
		"zip":               item["zip"],
		"country_code":      item["country_code"],
		"statut":            item["statut"],
		"date_creation":     item["date_creation"],
		"date_modification": item["date_modification"],
	}
}

func dolibarrProductRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"ref":               item["ref"],
		"label":             item["label"],
		"type":              item["type"],
		"price":             item["price"],
		"price_ttc":         item["price_ttc"],
		"tva_tx":            item["tva_tx"],
		"status":            item["status"],
		"status_buy":        item["status_buy"],
		"stock_reel":        item["stock_reel"],
		"date_creation":     item["date_creation"],
		"date_modification": item["date_modification"],
	}
}

func dolibarrInvoiceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"ref":               item["ref"],
		"socid":             item["socid"],
		"type":              item["type"],
		"total_ht":          item["total_ht"],
		"total_tva":         item["total_tva"],
		"total_ttc":         item["total_ttc"],
		"paye":              item["paye"],
		"status":            item["status"],
		"date":              item["date"],
		"date_creation":     item["date_creation"],
		"date_modification": item["date_modification"],
	}
}

func dolibarrOrderRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["id"],
		"ref":               item["ref"],
		"socid":             item["socid"],
		"total_ht":          item["total_ht"],
		"total_tva":         item["total_tva"],
		"total_ttc":         item["total_ttc"],
		"billed":            item["billed"],
		"status":            item["status"],
		"date":              item["date"],
		"date_creation":     item["date_creation"],
		"date_modification": item["date_modification"],
	}
}
