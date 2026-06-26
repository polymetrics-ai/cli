package ip2whois

import "polymetrics.ai/internal/connectors"

// ip2whoisStreams is the published stream catalog. IP2WHOIS exposes a single
// domain-lookup endpoint that returns one WHOIS object per domain, so the
// connector fans that object out into three derived streams:
//
//   - whois:       one flattened record per looked-up domain.
//   - nameservers: one record per (domain, nameserver) pair.
//   - contacts:    one record per (domain, contact role) — registrant, admin,
//     tech, billing.
//
// The natural cursor is update_date (the WHOIS last-modified timestamp); the
// primary key is the domain (plus the discriminator for the fan-out streams).
func ip2whoisStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "whois",
			Description:  "Flattened WHOIS record for each configured domain.",
			PrimaryKey:   []string{"domain"},
			CursorFields: []string{"update_date"},
			Fields:       whoisFields(),
		},
		{
			Name:        "nameservers",
			Description: "One record per nameserver for each configured domain.",
			PrimaryKey:  []string{"domain", "nameserver"},
			Fields: []connectors.Field{
				{Name: "domain", Type: "string"},
				{Name: "nameserver", Type: "string"},
			},
		},
		{
			Name:        "contacts",
			Description: "WHOIS contact records (registrant, admin, tech, billing) per domain.",
			PrimaryKey:  []string{"domain", "role"},
			Fields: []connectors.Field{
				{Name: "domain", Type: "string"},
				{Name: "role", Type: "string"},
				{Name: "name", Type: "string"},
				{Name: "organization", Type: "string"},
				{Name: "street_address", Type: "string"},
				{Name: "city", Type: "string"},
				{Name: "region", Type: "string"},
				{Name: "zip_code", Type: "string"},
				{Name: "country", Type: "string"},
				{Name: "phone", Type: "string"},
				{Name: "fax", Type: "string"},
				{Name: "email", Type: "string"},
			},
		},
	}
}

func whoisFields() []connectors.Field {
	return []connectors.Field{
		{Name: "domain", Type: "string"},
		{Name: "domain_id", Type: "string"},
		{Name: "status", Type: "string"},
		{Name: "create_date", Type: "timestamp"},
		{Name: "update_date", Type: "timestamp"},
		{Name: "expire_date", Type: "timestamp"},
		{Name: "domain_age", Type: "integer"},
		{Name: "whois_server", Type: "string"},
		{Name: "registrar_iana_id", Type: "string"},
		{Name: "registrar_name", Type: "string"},
		{Name: "registrar_url", Type: "string"},
		{Name: "registrant_name", Type: "string"},
		{Name: "registrant_organization", Type: "string"},
		{Name: "registrant_email", Type: "string"},
		{Name: "registrant_country", Type: "string"},
		{Name: "admin_name", Type: "string"},
		{Name: "admin_email", Type: "string"},
		{Name: "tech_name", Type: "string"},
		{Name: "tech_email", Type: "string"},
		{Name: "billing_name", Type: "string"},
		{Name: "billing_email", Type: "string"},
		{Name: "nameservers", Type: "string"},
	}
}

// contactRoles is the fixed set of WHOIS contact roles IP2WHOIS returns as
// nested objects.
var contactRoles = []string{"registrant", "admin", "tech", "billing"}

// whoisRecord flattens a raw IP2WHOIS lookup object into the whois stream's
// connectors.Record. Nested registrar/contact objects are flattened with a
// role prefix; nameservers are joined for the flat view.
func whoisRecord(item map[string]any) connectors.Record {
	registrar := nestedObject(item["registrar"])
	registrant := nestedObject(item["registrant"])
	admin := nestedObject(item["admin"])
	tech := nestedObject(item["tech"])
	billing := nestedObject(item["billing"])

	return connectors.Record{
		"domain":                  item["domain"],
		"domain_id":               item["domain_id"],
		"status":                  item["status"],
		"create_date":             item["create_date"],
		"update_date":             item["update_date"],
		"expire_date":             item["expire_date"],
		"domain_age":              item["domain_age"],
		"whois_server":            item["whois_server"],
		"registrar_iana_id":       registrar["iana_id"],
		"registrar_name":          registrar["name"],
		"registrar_url":           registrar["url"],
		"registrant_name":         registrant["name"],
		"registrant_organization": registrant["organization"],
		"registrant_email":        registrant["email"],
		"registrant_country":      registrant["country"],
		"admin_name":              admin["name"],
		"admin_email":             admin["email"],
		"tech_name":               tech["name"],
		"tech_email":              tech["email"],
		"billing_name":            billing["name"],
		"billing_email":           billing["email"],
		"nameservers":             joinNameservers(item["nameservers"]),
	}
}

// nameserverRecords fans a single lookup object out into one record per
// nameserver.
func nameserverRecords(item map[string]any) []connectors.Record {
	domain := item["domain"]
	out := make([]connectors.Record, 0)
	for _, ns := range nameserverList(item["nameservers"]) {
		out = append(out, connectors.Record{
			"domain":     domain,
			"nameserver": ns,
		})
	}
	return out
}

// contactRecords fans a single lookup object out into one record per populated
// contact role.
func contactRecords(item map[string]any) []connectors.Record {
	domain := item["domain"]
	out := make([]connectors.Record, 0, len(contactRoles))
	for _, role := range contactRoles {
		contact := nestedObject(item[role])
		if len(contact) == 0 {
			continue
		}
		out = append(out, connectors.Record{
			"domain":         domain,
			"role":           role,
			"name":           contact["name"],
			"organization":   contact["organization"],
			"street_address": contact["street_address"],
			"city":           contact["city"],
			"region":         contact["region"],
			"zip_code":       contact["zip_code"],
			"country":        contact["country"],
			"phone":          contact["phone"],
			"fax":            contact["fax"],
			"email":          contact["email"],
		})
	}
	return out
}

// nestedObject coerces an arbitrary JSON value into a map, returning an empty
// map when the value is absent or not an object.
func nestedObject(v any) map[string]any {
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return map[string]any{}
}

// nameserverList normalizes the nameservers field (a JSON array of strings)
// into a []string.
func nameserverList(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok && s != "" {
			out = append(out, s)
		}
	}
	return out
}

// joinNameservers renders the nameserver array as a comma-separated string for
// the flat whois stream.
func joinNameservers(v any) string {
	list := nameserverList(v)
	out := ""
	for i, ns := range list {
		if i > 0 {
			out += ","
		}
		out += ns
	}
	return out
}
