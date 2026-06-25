package prestashop

import "polymetrics/internal/connectors"

// streamEndpoint maps a stream name to the PrestaShop Webservice resource path
// (relative to <shop_url>/api) it reads from, and the record mapper that
// flattens its objects.
type streamEndpoint struct {
	// resource is the PrestaShop list endpoint path segment (e.g. "customers").
	resource string
	// mapRecord flattens a raw PrestaShop object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// prestashopStreamEndpoints is the per-stream routing table. Adding a stream
// means adding one entry here plus a Stream definition in prestashopStreams;
// the read path is fully data-driven from this table.
var prestashopStreamEndpoints = map[string]streamEndpoint{
	"customers": {resource: "customers", mapRecord: prestashopCustomerRecord},
	"orders":    {resource: "orders", mapRecord: prestashopOrderRecord},
	"products":  {resource: "products", mapRecord: prestashopProductRecord},
	"addresses": {resource: "addresses", mapRecord: prestashopAddressRecord},
	"carts":     {resource: "carts", mapRecord: prestashopCartRecord},
}

// prestashopStreams returns the connector's published stream catalog. Every
// PrestaShop resource exposes a numeric id and a date_upd timestamp, so the
// primary key is ["id"] and the incremental cursor field is ["date_upd"].
func prestashopStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "customers",
			Description:  "PrestaShop customers.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_upd"},
			Fields:       prestashopCustomerFields(),
		},
		{
			Name:         "orders",
			Description:  "PrestaShop orders.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_upd"},
			Fields:       prestashopOrderFields(),
		},
		{
			Name:         "products",
			Description:  "PrestaShop products.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_upd"},
			Fields:       prestashopProductFields(),
		},
		{
			Name:         "addresses",
			Description:  "PrestaShop addresses.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_upd"},
			Fields:       prestashopAddressFields(),
		},
		{
			Name:         "carts",
			Description:  "PrestaShop carts.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"date_upd"},
			Fields:       prestashopCartFields(),
		},
	}
}

func prestashopCustomerFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "id_default_group", Type: "integer"},
		{Name: "id_lang", Type: "integer"},
		{Name: "firstname", Type: "string"},
		{Name: "lastname", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "company", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "newsletter", Type: "boolean"},
		{Name: "date_add", Type: "timestamp"},
		{Name: "date_upd", Type: "timestamp"},
	}
}

func prestashopOrderFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "id_customer", Type: "integer"},
		{Name: "id_address_delivery", Type: "integer"},
		{Name: "id_address_invoice", Type: "integer"},
		{Name: "current_state", Type: "integer"},
		{Name: "reference", Type: "string"},
		{Name: "payment", Type: "string"},
		{Name: "total_paid", Type: "string"},
		{Name: "total_paid_real", Type: "string"},
		{Name: "valid", Type: "boolean"},
		{Name: "date_add", Type: "timestamp"},
		{Name: "date_upd", Type: "timestamp"},
	}
}

func prestashopProductFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "id_manufacturer", Type: "integer"},
		{Name: "id_supplier", Type: "integer"},
		{Name: "id_category_default", Type: "integer"},
		{Name: "reference", Type: "string"},
		{Name: "price", Type: "string"},
		{Name: "active", Type: "boolean"},
		{Name: "quantity", Type: "integer"},
		{Name: "date_add", Type: "timestamp"},
		{Name: "date_upd", Type: "timestamp"},
	}
}

func prestashopAddressFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "id_customer", Type: "integer"},
		{Name: "id_country", Type: "integer"},
		{Name: "id_state", Type: "integer"},
		{Name: "firstname", Type: "string"},
		{Name: "lastname", Type: "string"},
		{Name: "city", Type: "string"},
		{Name: "postcode", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "date_add", Type: "timestamp"},
		{Name: "date_upd", Type: "timestamp"},
	}
}

func prestashopCartFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "integer"},
		{Name: "id_customer", Type: "integer"},
		{Name: "id_address_delivery", Type: "integer"},
		{Name: "id_address_invoice", Type: "integer"},
		{Name: "id_currency", Type: "integer"},
		{Name: "id_lang", Type: "integer"},
		{Name: "id_carrier", Type: "integer"},
		{Name: "date_add", Type: "timestamp"},
		{Name: "date_upd", Type: "timestamp"},
	}
}

func prestashopCustomerRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":               item["id"],
		"id_default_group": item["id_default_group"],
		"id_lang":          item["id_lang"],
		"firstname":        item["firstname"],
		"lastname":         item["lastname"],
		"email":            item["email"],
		"company":          item["company"],
		"active":           item["active"],
		"newsletter":       item["newsletter"],
		"date_add":         item["date_add"],
		"date_upd":         item["date_upd"],
	}
}

func prestashopOrderRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"id_customer":         item["id_customer"],
		"id_address_delivery": item["id_address_delivery"],
		"id_address_invoice":  item["id_address_invoice"],
		"current_state":       item["current_state"],
		"reference":           item["reference"],
		"payment":             item["payment"],
		"total_paid":          item["total_paid"],
		"total_paid_real":     item["total_paid_real"],
		"valid":               item["valid"],
		"date_add":            item["date_add"],
		"date_upd":            item["date_upd"],
	}
}

func prestashopProductRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"id_manufacturer":     item["id_manufacturer"],
		"id_supplier":         item["id_supplier"],
		"id_category_default": item["id_category_default"],
		"reference":           item["reference"],
		"price":               item["price"],
		"active":              item["active"],
		"quantity":            item["quantity"],
		"date_add":            item["date_add"],
		"date_upd":            item["date_upd"],
	}
}

func prestashopAddressRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"id_customer": item["id_customer"],
		"id_country":  item["id_country"],
		"id_state":    item["id_state"],
		"firstname":   item["firstname"],
		"lastname":    item["lastname"],
		"city":        item["city"],
		"postcode":    item["postcode"],
		"phone":       item["phone"],
		"date_add":    item["date_add"],
		"date_upd":    item["date_upd"],
	}
}

func prestashopCartRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                  item["id"],
		"id_customer":         item["id_customer"],
		"id_address_delivery": item["id_address_delivery"],
		"id_address_invoice":  item["id_address_invoice"],
		"id_currency":         item["id_currency"],
		"id_lang":             item["id_lang"],
		"id_carrier":          item["id_carrier"],
		"date_add":            item["date_add"],
		"date_upd":            item["date_upd"],
	}
}
