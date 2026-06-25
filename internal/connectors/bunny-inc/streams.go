package bunnyinc

import "polymetrics/internal/connectors"

// streamEndpoint describes how to read one Bunny GraphQL collection. Bunny exposes
// a single POST /graphql endpoint; each stream is a top-level connection field
// (e.g. `accounts`) queried with cursor pagination. gqlField is that connection
// name (used both in the query and to walk the response: data.<gqlField>.nodes,
// data.<gqlField>.pageInfo). query is the GraphQL document with a single $after
// String variable. mapRecord flattens a node into a connectors.Record.
type streamEndpoint struct {
	gqlField  string
	query     string
	mapRecord func(map[string]any) connectors.Record
}

// bunnyStreamEndpoints is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in bunnyStreams; the read path is
// fully data-driven from this table.
var bunnyStreamEndpoints = map[string]streamEndpoint{
	"accounts":      {gqlField: "accounts", query: accountsQuery, mapRecord: accountRecord},
	"contacts":      {gqlField: "contacts", query: contactsQuery, mapRecord: contactRecord},
	"invoices":      {gqlField: "invoices", query: invoicesQuery, mapRecord: invoiceRecord},
	"payments":      {gqlField: "payments", query: paymentsQuery, mapRecord: paymentRecord},
	"subscriptions": {gqlField: "subscriptions", query: subscriptionsQuery, mapRecord: subscriptionRecord},
}

// bunnyStreams returns the connector's published stream catalog. Every Bunny node
// exposes a string id (primary key) and createdAt/updatedAt timestamps; updatedAt
// is the natural incremental cursor where present.
func bunnyStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:         "accounts",
			Description:  "Bunny customer accounts.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       accountFields(),
		},
		{
			Name:         "contacts",
			Description:  "Bunny account contacts.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       contactFields(),
		},
		{
			Name:         "invoices",
			Description:  "Bunny invoices.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       invoiceFields(),
		},
		{
			Name:         "payments",
			Description:  "Bunny payments.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       paymentFields(),
		},
		{
			Name:         "subscriptions",
			Description:  "Bunny subscriptions.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"updatedAt"},
			Fields:       subscriptionFields(),
		},
	}
}

// The GraphQL documents below mirror the upstream Airbyte manifest (one $after
// String variable, first: pageSize injected at request time, and a pageInfo block
// for cursor pagination). The selected fields are the stable subset the connector
// maps; %d is replaced by the configured page size before sending.
const (
	accountsQuery = `query($after: String) { accounts(first: %d, after: $after) { nodes { id accountTypeId annualRevenue billingCity billingContactId billingCountry billingDay billingState billingStreet billingZip code createdAt currencyId description employees entityId groupId industryId invoiceTemplateId name netPaymentDays ownerUserId payingStatus phone taxNumber timezone updatedAt website } pageInfo { startCursor endCursor hasNextPage hasPreviousPage } } }`

	contactsQuery = `query($after: String) { contacts(first: %d, after: $after) { nodes { id accountId code createdAt description email entityId firstName fullName lastName linkedinUrl mailingCity mailingCountry mailingState mailingStreet mailingZip mobile phone portalAccess salutation title updatedAt } pageInfo { startCursor endCursor hasNextPage hasPreviousPage } } }`

	invoicesQuery = `query($after: String) { invoices(first: %d, after: $after) { nodes { id accountId amount amountDue amountPaid createdAt credits currencyId description dueAt netPaymentDays number paidAt payableId poNumber portalUrl quoteId smallUnitAmountDue subtotal taxAmount updatedAt url uuid } pageInfo { startCursor endCursor hasNextPage hasPreviousPage } } }`

	paymentsQuery = `query($after: String) { payments(first: %d, after: $after) { nodes { id accountId amount amountUnapplied baseCurrencyCash baseCurrencyId createdAt currencyId description isLegacy memo receivedAt updatedAt } pageInfo { startCursor endCursor hasNextPage hasPreviousPage } } }`

	subscriptionsQuery = `query($after: String) { subscriptions(first: %d, after: $after) { nodes { id accountId cancelationDate createdAt currencyId endDate name period priceListId rampIntervalMonths startDate trialEndDate trialPeriod trialStartDate updatedAt } pageInfo { startCursor endCursor hasNextPage hasPreviousPage } } }`
)

func accountFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "code", Type: "string"},
		{Name: "accountTypeId", Type: "string"},
		{Name: "currencyId", Type: "string"},
		{Name: "entityId", Type: "string"},
		{Name: "ownerUserId", Type: "string"},
		{Name: "payingStatus", Type: "string"},
		{Name: "billingCountry", Type: "string"},
		{Name: "annualRevenue", Type: "number"},
		{Name: "employees", Type: "integer"},
		{Name: "netPaymentDays", Type: "integer"},
		{Name: "phone", Type: "string"},
		{Name: "website", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func contactFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "accountId", Type: "string"},
		{Name: "entityId", Type: "string"},
		{Name: "code", Type: "string"},
		{Name: "firstName", Type: "string"},
		{Name: "lastName", Type: "string"},
		{Name: "fullName", Type: "string"},
		{Name: "email", Type: "string"},
		{Name: "phone", Type: "string"},
		{Name: "mobile", Type: "string"},
		{Name: "title", Type: "string"},
		{Name: "portalAccess", Type: "boolean"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func invoiceFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "uuid", Type: "string"},
		{Name: "accountId", Type: "string"},
		{Name: "currencyId", Type: "string"},
		{Name: "quoteId", Type: "string"},
		{Name: "number", Type: "string"},
		{Name: "amount", Type: "number"},
		{Name: "amountDue", Type: "number"},
		{Name: "amountPaid", Type: "number"},
		{Name: "subtotal", Type: "number"},
		{Name: "taxAmount", Type: "number"},
		{Name: "credits", Type: "number"},
		{Name: "netPaymentDays", Type: "integer"},
		{Name: "dueAt", Type: "string"},
		{Name: "paidAt", Type: "string"},
		{Name: "url", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func paymentFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "accountId", Type: "string"},
		{Name: "currencyId", Type: "string"},
		{Name: "baseCurrencyId", Type: "string"},
		{Name: "amount", Type: "number"},
		{Name: "amountUnapplied", Type: "number"},
		{Name: "baseCurrencyCash", Type: "number"},
		{Name: "description", Type: "string"},
		{Name: "memo", Type: "string"},
		{Name: "isLegacy", Type: "boolean"},
		{Name: "receivedAt", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func subscriptionFields() []connectors.Field {
	return []connectors.Field{
		{Name: "id", Type: "string"},
		{Name: "accountId", Type: "string"},
		{Name: "currencyId", Type: "string"},
		{Name: "priceListId", Type: "string"},
		{Name: "name", Type: "string"},
		{Name: "period", Type: "string"},
		{Name: "rampIntervalMonths", Type: "integer"},
		{Name: "trialPeriod", Type: "integer"},
		{Name: "startDate", Type: "string"},
		{Name: "endDate", Type: "string"},
		{Name: "cancelationDate", Type: "string"},
		{Name: "trialStartDate", Type: "string"},
		{Name: "trialEndDate", Type: "string"},
		{Name: "createdAt", Type: "string"},
		{Name: "updatedAt", Type: "string"},
	}
}

func accountRecord(item map[string]any) connectors.Record {
	return pick(item,
		"id", "name", "code", "accountTypeId", "currencyId", "entityId",
		"ownerUserId", "payingStatus", "billingCountry", "annualRevenue",
		"employees", "netPaymentDays", "phone", "website", "createdAt", "updatedAt",
	)
}

func contactRecord(item map[string]any) connectors.Record {
	return pick(item,
		"id", "accountId", "entityId", "code", "firstName", "lastName",
		"fullName", "email", "phone", "mobile", "title", "portalAccess",
		"createdAt", "updatedAt",
	)
}

func invoiceRecord(item map[string]any) connectors.Record {
	return pick(item,
		"id", "uuid", "accountId", "currencyId", "quoteId", "number", "amount",
		"amountDue", "amountPaid", "subtotal", "taxAmount", "credits",
		"netPaymentDays", "dueAt", "paidAt", "url", "createdAt", "updatedAt",
	)
}

func paymentRecord(item map[string]any) connectors.Record {
	return pick(item,
		"id", "accountId", "currencyId", "baseCurrencyId", "amount",
		"amountUnapplied", "baseCurrencyCash", "description", "memo", "isLegacy",
		"receivedAt", "createdAt", "updatedAt",
	)
}

func subscriptionRecord(item map[string]any) connectors.Record {
	return pick(item,
		"id", "accountId", "currencyId", "priceListId", "name", "period",
		"rampIntervalMonths", "trialPeriod", "startDate", "endDate",
		"cancelationDate", "trialStartDate", "trialEndDate", "createdAt", "updatedAt",
	)
}

// pick projects the named keys from a raw GraphQL node into a Record, preserving
// nil for absent keys so the record shape stays stable across pages.
func pick(item map[string]any, keys ...string) connectors.Record {
	out := make(connectors.Record, len(keys))
	for _, k := range keys {
		out[k] = item[k]
	}
	return out
}
