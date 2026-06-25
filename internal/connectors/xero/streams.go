package xero

import "polymetrics.ai/internal/connectors"

// streamEndpoint maps a stream name to the Xero Accounting API resource path
// (relative to base_url), the JSON key the records array lives under in the
// response envelope, and the record mapper that flattens its objects.
//
// Xero list responses are shaped {"Id":...,"Status":"OK","<Resource>":[...]},
// where <Resource> is the plural resource name (e.g. "Invoices"). recordsKey is
// that name. The primary-key id field differs per resource (InvoiceID,
// ContactID, ...), so each mapper normalizes it to "id" while preserving the
// native key as well.
type streamEndpoint struct {
	// resource is the URL path segment and also the records key in the envelope,
	// e.g. "Invoices".
	resource string
	// idField is the native Xero primary-key field for this resource.
	idField string
	// mapRecord flattens a raw Xero object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
}

// xeroStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in xeroStreams; the read path
// is fully data-driven from this table.
var xeroStreamEndpoints = map[string]streamEndpoint{
	"invoices":          {resource: "Invoices", idField: "InvoiceID", mapRecord: xeroInvoiceRecord},
	"contacts":          {resource: "Contacts", idField: "ContactID", mapRecord: xeroContactRecord},
	"accounts":          {resource: "Accounts", idField: "AccountID", mapRecord: xeroAccountRecord},
	"bank_transactions": {resource: "BankTransactions", idField: "BankTransactionID", mapRecord: xeroBankTransactionRecord},
	"items":             {resource: "Items", idField: "ItemID", mapRecord: xeroItemRecord},
	"payments":          {resource: "Payments", idField: "PaymentID", mapRecord: xeroPaymentRecord},
}

// xeroStreams returns the connector's published stream catalog. Every Xero
// accounting object carries a UTC last-modified timestamp (UpdatedDateUTC) used
// as the incremental cursor, and a resource-specific GUID primary key.
func xeroStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "invoices",
			Description:  "Xero sales and purchase invoices.",
			PrimaryKey:   []string{"InvoiceID"},
			CursorFields: []string{"UpdatedDateUTC"},
			Fields:       xeroInvoiceFields(),
		},
		{
			Name:         "contacts",
			Description:  "Xero contacts (customers and suppliers).",
			PrimaryKey:   []string{"ContactID"},
			CursorFields: []string{"UpdatedDateUTC"},
			Fields:       xeroContactFields(),
		},
		{
			Name:         "accounts",
			Description:  "Xero chart of accounts.",
			PrimaryKey:   []string{"AccountID"},
			CursorFields: []string{"UpdatedDateUTC"},
			Fields:       xeroAccountFields(),
		},
		{
			Name:         "bank_transactions",
			Description:  "Xero bank transactions (spend and receive money).",
			PrimaryKey:   []string{"BankTransactionID"},
			CursorFields: []string{"UpdatedDateUTC"},
			Fields:       xeroBankTransactionFields(),
		},
		{
			Name:         "items",
			Description:  "Xero inventory items.",
			PrimaryKey:   []string{"ItemID"},
			CursorFields: []string{"UpdatedDateUTC"},
			Fields:       xeroItemFields(),
		},
		{
			Name:         "payments",
			Description:  "Xero payments applied to invoices and credit notes.",
			PrimaryKey:   []string{"PaymentID"},
			CursorFields: []string{"UpdatedDateUTC"},
			Fields:       xeroPaymentFields(),
		},
	}
}

func xeroInvoiceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "InvoiceID", Type: "string"},
		{Name: "InvoiceNumber", Type: "string"},
		{Name: "Type", Type: "string"},
		{Name: "Status", Type: "string"},
		{Name: "LineAmountTypes", Type: "string"},
		{Name: "SubTotal", Type: "number"},
		{Name: "TotalTax", Type: "number"},
		{Name: "Total", Type: "number"},
		{Name: "AmountDue", Type: "number"},
		{Name: "AmountPaid", Type: "number"},
		{Name: "CurrencyCode", Type: "string"},
		{Name: "Date", Type: "string"},
		{Name: "DueDate", Type: "string"},
		{Name: "ContactID", Type: "string"},
		{Name: "UpdatedDateUTC", Type: "string"},
	}
}

func xeroContactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "ContactID", Type: "string"},
		{Name: "ContactStatus", Type: "string"},
		{Name: "Name", Type: "string"},
		{Name: "FirstName", Type: "string"},
		{Name: "LastName", Type: "string"},
		{Name: "EmailAddress", Type: "string"},
		{Name: "IsSupplier", Type: "boolean"},
		{Name: "IsCustomer", Type: "boolean"},
		{Name: "AccountNumber", Type: "string"},
		{Name: "UpdatedDateUTC", Type: "string"},
	}
}

func xeroAccountFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "AccountID", Type: "string"},
		{Name: "Code", Type: "string"},
		{Name: "Name", Type: "string"},
		{Name: "Type", Type: "string"},
		{Name: "Status", Type: "string"},
		{Name: "Class", Type: "string"},
		{Name: "TaxType", Type: "string"},
		{Name: "EnablePaymentsToAccount", Type: "boolean"},
		{Name: "UpdatedDateUTC", Type: "string"},
	}
}

func xeroBankTransactionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "BankTransactionID", Type: "string"},
		{Name: "Type", Type: "string"},
		{Name: "Status", Type: "string"},
		{Name: "IsReconciled", Type: "boolean"},
		{Name: "SubTotal", Type: "number"},
		{Name: "TotalTax", Type: "number"},
		{Name: "Total", Type: "number"},
		{Name: "CurrencyCode", Type: "string"},
		{Name: "Date", Type: "string"},
		{Name: "ContactID", Type: "string"},
		{Name: "UpdatedDateUTC", Type: "string"},
	}
}

func xeroItemFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "ItemID", Type: "string"},
		{Name: "Code", Type: "string"},
		{Name: "Name", Type: "string"},
		{Name: "Description", Type: "string"},
		{Name: "IsSold", Type: "boolean"},
		{Name: "IsPurchased", Type: "boolean"},
		{Name: "IsTrackedAsInventory", Type: "boolean"},
		{Name: "QuantityOnHand", Type: "number"},
		{Name: "UpdatedDateUTC", Type: "string"},
	}
}

func xeroPaymentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "PaymentID", Type: "string"},
		{Name: "PaymentType", Type: "string"},
		{Name: "Status", Type: "string"},
		{Name: "Amount", Type: "number"},
		{Name: "Reference", Type: "string"},
		{Name: "CurrencyRate", Type: "number"},
		{Name: "Date", Type: "string"},
		{Name: "UpdatedDateUTC", Type: "string"},
	}
}

// contactID extracts the nested Contact.ContactID, which Xero embeds as a nested
// object on invoices and bank transactions.
func contactID(item map[string]any) any {
	if contact, ok := item["Contact"].(map[string]any); ok {
		return contact["ContactID"]
	}
	return nil
}

func xeroInvoiceRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":              item["InvoiceID"],
		"InvoiceID":       item["InvoiceID"],
		"InvoiceNumber":   item["InvoiceNumber"],
		"Type":            item["Type"],
		"Status":          item["Status"],
		"LineAmountTypes": item["LineAmountTypes"],
		"SubTotal":        item["SubTotal"],
		"TotalTax":        item["TotalTax"],
		"Total":           item["Total"],
		"AmountDue":       item["AmountDue"],
		"AmountPaid":      item["AmountPaid"],
		"CurrencyCode":    item["CurrencyCode"],
		"Date":            item["Date"],
		"DueDate":         item["DueDate"],
		"ContactID":       contactID(item),
		"UpdatedDateUTC":  item["UpdatedDateUTC"],
		// lower-cased convenience aliases used by tests and downstream mapping.
		"type":       item["Type"],
		"status":     item["Status"],
		"updated_at": item["UpdatedDateUTC"],
	}
}

func xeroContactRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["ContactID"],
		"ContactID":      item["ContactID"],
		"ContactStatus":  item["ContactStatus"],
		"Name":           item["Name"],
		"FirstName":      item["FirstName"],
		"LastName":       item["LastName"],
		"EmailAddress":   item["EmailAddress"],
		"IsSupplier":     item["IsSupplier"],
		"IsCustomer":     item["IsCustomer"],
		"AccountNumber":  item["AccountNumber"],
		"UpdatedDateUTC": item["UpdatedDateUTC"],
		"status":         item["ContactStatus"],
		"updated_at":     item["UpdatedDateUTC"],
	}
}

func xeroAccountRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                      item["AccountID"],
		"AccountID":               item["AccountID"],
		"Code":                    item["Code"],
		"Name":                    item["Name"],
		"Type":                    item["Type"],
		"Status":                  item["Status"],
		"Class":                   item["Class"],
		"TaxType":                 item["TaxType"],
		"EnablePaymentsToAccount": item["EnablePaymentsToAccount"],
		"UpdatedDateUTC":          item["UpdatedDateUTC"],
		"type":                    item["Type"],
		"status":                  item["Status"],
		"updated_at":              item["UpdatedDateUTC"],
	}
}

func xeroBankTransactionRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                item["BankTransactionID"],
		"BankTransactionID": item["BankTransactionID"],
		"Type":              item["Type"],
		"Status":            item["Status"],
		"IsReconciled":      item["IsReconciled"],
		"SubTotal":          item["SubTotal"],
		"TotalTax":          item["TotalTax"],
		"Total":             item["Total"],
		"CurrencyCode":      item["CurrencyCode"],
		"Date":              item["Date"],
		"ContactID":         contactID(item),
		"UpdatedDateUTC":    item["UpdatedDateUTC"],
		"type":              item["Type"],
		"status":            item["Status"],
		"updated_at":        item["UpdatedDateUTC"],
	}
}

func xeroItemRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":                   item["ItemID"],
		"ItemID":               item["ItemID"],
		"Code":                 item["Code"],
		"Name":                 item["Name"],
		"Description":          item["Description"],
		"IsSold":               item["IsSold"],
		"IsPurchased":          item["IsPurchased"],
		"IsTrackedAsInventory": item["IsTrackedAsInventory"],
		"QuantityOnHand":       item["QuantityOnHand"],
		"UpdatedDateUTC":       item["UpdatedDateUTC"],
		"updated_at":           item["UpdatedDateUTC"],
	}
}

func xeroPaymentRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":             item["PaymentID"],
		"PaymentID":      item["PaymentID"],
		"PaymentType":    item["PaymentType"],
		"Status":         item["Status"],
		"Amount":         item["Amount"],
		"Reference":      item["Reference"],
		"CurrencyRate":   item["CurrencyRate"],
		"Date":           item["Date"],
		"UpdatedDateUTC": item["UpdatedDateUTC"],
		"status":         item["Status"],
		"updated_at":     item["UpdatedDateUTC"],
	}
}
