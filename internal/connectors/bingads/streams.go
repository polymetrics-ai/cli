package bingads

import (
	"strings"

	"polymetrics.ai/internal/connectors"
)

// serviceKind selects which Bing Ads REST service (and therefore which base URL
// and header set) a stream targets.
type serviceKind int

const (
	serviceCustomer serviceKind = iota
	serviceCampaign
)

// streamEndpoint maps a stream to its REST POST endpoint, the JSON path to the
// records array in the response, the request-body builder, the service kind, the
// record mapper, and a deterministic fixture generator.
type streamEndpoint struct {
	// path is the endpoint path relative to the service base URL.
	path string
	// recordsPath is the dotted JSON path to the records array (e.g. "Campaigns").
	recordsPath string
	// kind selects customer vs campaign service wiring.
	kind serviceKind
	// body builds the POST request body from the runtime config.
	body func(connectors.RuntimeConfig) map[string]any
	// mapRecord flattens a raw Bing Ads object into a connectors.Record.
	mapRecord func(map[string]any) connectors.Record
	// fixture returns deterministic raw objects for credential-free conformance.
	fixture func() []map[string]any
}

// streamEndpoints is the per-stream routing table. Adding a stream means adding
// one entry here plus a Stream definition in bingStreams; the read path is fully
// data-driven from this table.
var streamEndpoints = map[string]streamEndpoint{
	"accounts": {
		path:        "AccountsInfo/Query",
		recordsPath: "AccountsInfo",
		kind:        serviceCustomer,
		body: func(cfg connectors.RuntimeConfig) map[string]any {
			b := map[string]any{"OnlyParentAccounts": false}
			if v := strings.TrimSpace(cfg.Config["customer_id"]); v != "" {
				b["CustomerId"] = v
			}
			return b
		},
		mapRecord: accountRecord,
		fixture:   accountFixture,
	},
	"users": {
		path:        "User/Query",
		recordsPath: "User",
		kind:        serviceCustomer,
		body: func(cfg connectors.RuntimeConfig) map[string]any {
			return map[string]any{"UserId": nil}
		},
		mapRecord: userRecord,
		fixture:   userFixture,
	},
	"campaigns": {
		path:        "Campaigns/QueryByAccountId",
		recordsPath: "Campaigns",
		kind:        serviceCampaign,
		body: func(cfg connectors.RuntimeConfig) map[string]any {
			b := map[string]any{}
			if v := customerAccountID(cfg); v != "" {
				b["AccountId"] = v
			}
			return b
		},
		mapRecord: campaignRecord,
		fixture:   campaignFixture,
	},
	"ad_groups": {
		path:        "AdGroups/QueryByCampaignId",
		recordsPath: "AdGroups",
		kind:        serviceCampaign,
		body: func(cfg connectors.RuntimeConfig) map[string]any {
			b := map[string]any{}
			if v := strings.TrimSpace(cfg.Config["campaign_id"]); v != "" {
				b["CampaignId"] = v
			}
			return b
		},
		mapRecord: adGroupRecord,
		fixture:   adGroupFixture,
	},
	"ads": {
		path:        "Ads/QueryByAdGroupId",
		recordsPath: "Ads",
		kind:        serviceCampaign,
		body: func(cfg connectors.RuntimeConfig) map[string]any {
			b := map[string]any{}
			if v := strings.TrimSpace(cfg.Config["ad_group_id"]); v != "" {
				b["AdGroupId"] = v
			}
			return b
		},
		mapRecord: adRecord,
		fixture:   adFixture,
	},
}

// bingStreams returns the connector's published stream catalog. Bing Ads objects
// are keyed by a numeric Id (carried as a string in the JSON REST surface), so
// the primary key is ["Id"] across the board. These are full-refresh streams
// (the upstream catalog only advertises full_refresh), so no cursor fields are
// published.
func bingStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        "accounts",
			Description: "Microsoft Advertising advertiser accounts (AccountsInfo).",
			PrimaryKey:  []string{"Id"},
			Fields:      accountFields(),
		},
		{
			Name:        "users",
			Description: "The authenticated Microsoft Advertising user.",
			PrimaryKey:  []string{"Id"},
			Fields:      userFields(),
		},
		{
			Name:        "campaigns",
			Description: "Campaigns within the configured advertiser account.",
			PrimaryKey:  []string{"Id"},
			Fields:      campaignFields(),
		},
		{
			Name:        "ad_groups",
			Description: "Ad groups within the configured campaign.",
			PrimaryKey:  []string{"Id"},
			Fields:      adGroupFields(),
		},
		{
			Name:        "ads",
			Description: "Ads within the configured ad group.",
			PrimaryKey:  []string{"Id"},
			Fields:      adFields(),
		},
	}
}

func accountFields() []connectors.Field {
	return []connectors.Field{
		{Name: "Id", Type: "string"},
		{Name: "Name", Type: "string"},
		{Name: "Number", Type: "string"},
		{Name: "AccountLifeCycleStatus", Type: "string"},
		{Name: "PauseReason", Type: "string"},
	}
}

func userFields() []connectors.Field {
	return []connectors.Field{
		{Name: "Id", Type: "string"},
		{Name: "UserName", Type: "string"},
		{Name: "CustomerId", Type: "string"},
		{Name: "JobTitle", Type: "string"},
		{Name: "UserLifeCycleStatus", Type: "string"},
		{Name: "LastModifiedTime", Type: "string"},
	}
}

func campaignFields() []connectors.Field {
	return []connectors.Field{
		{Name: "Id", Type: "string"},
		{Name: "Name", Type: "string"},
		{Name: "Status", Type: "string"},
		{Name: "CampaignType", Type: "string"},
		{Name: "BudgetType", Type: "string"},
		{Name: "DailyBudget", Type: "number"},
		{Name: "TimeZone", Type: "string"},
	}
}

func adGroupFields() []connectors.Field {
	return []connectors.Field{
		{Name: "Id", Type: "string"},
		{Name: "Name", Type: "string"},
		{Name: "Status", Type: "string"},
		{Name: "AdRotation", Type: "string"},
		{Name: "Network", Type: "string"},
		{Name: "StartDate", Type: "string"},
		{Name: "EndDate", Type: "string"},
	}
}

func adFields() []connectors.Field {
	return []connectors.Field{
		{Name: "Id", Type: "string"},
		{Name: "Type", Type: "string"},
		{Name: "Status", Type: "string"},
		{Name: "EditorialStatus", Type: "string"},
		{Name: "DevicePreference", Type: "string"},
	}
}

func accountRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"Id":                     item["Id"],
		"Name":                   item["Name"],
		"Number":                 item["Number"],
		"AccountLifeCycleStatus": item["AccountLifeCycleStatus"],
		"PauseReason":            item["PauseReason"],
	}
}

func userRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"Id":                  item["Id"],
		"UserName":            item["UserName"],
		"CustomerId":          item["CustomerId"],
		"JobTitle":            item["JobTitle"],
		"UserLifeCycleStatus": item["UserLifeCycleStatus"],
		"LastModifiedTime":    item["LastModifiedTime"],
	}
}

func campaignRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"Id":           item["Id"],
		"Name":         item["Name"],
		"Status":       item["Status"],
		"CampaignType": item["CampaignType"],
		"BudgetType":   item["BudgetType"],
		"DailyBudget":  item["DailyBudget"],
		"TimeZone":     item["TimeZone"],
	}
}

func adGroupRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"Id":         item["Id"],
		"Name":       item["Name"],
		"Status":     item["Status"],
		"AdRotation": item["AdRotation"],
		"Network":    item["Network"],
		"StartDate":  item["StartDate"],
		"EndDate":    item["EndDate"],
	}
}

func adRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"Id":               item["Id"],
		"Type":             item["Type"],
		"Status":           item["Status"],
		"EditorialStatus":  item["EditorialStatus"],
		"DevicePreference": item["DevicePreference"],
	}
}

func accountFixture() []map[string]any {
	return []map[string]any{
		{"Id": "1000001", "Name": "Fixture Account 1", "Number": "F0000001", "AccountLifeCycleStatus": "Active", "PauseReason": nil},
		{"Id": "1000002", "Name": "Fixture Account 2", "Number": "F0000002", "AccountLifeCycleStatus": "Paused", "PauseReason": "PausedByUser"},
	}
}

func userFixture() []map[string]any {
	return []map[string]any{
		{"Id": "2000001", "UserName": "fixture-user", "CustomerId": "3000001", "JobTitle": "Account Manager", "UserLifeCycleStatus": "Active", "LastModifiedTime": "2026-01-01T00:00:00Z"},
	}
}

func campaignFixture() []map[string]any {
	return []map[string]any{
		{"Id": "4000001", "Name": "Fixture Brand Campaign", "Status": "Active", "CampaignType": "Search", "BudgetType": "DailyBudgetStandard", "DailyBudget": 50.0, "TimeZone": "PacificTimeUSCanadaTijuana"},
		{"Id": "4000002", "Name": "Fixture Shopping Campaign", "Status": "Paused", "CampaignType": "Shopping", "BudgetType": "DailyBudgetStandard", "DailyBudget": 25.0, "TimeZone": "PacificTimeUSCanadaTijuana"},
	}
}

func adGroupFixture() []map[string]any {
	return []map[string]any{
		{"Id": "5000001", "Name": "Fixture Ad Group", "Status": "Active", "AdRotation": "OptimizeForClicks", "Network": "OwnedAndOperatedAndSyndicatedSearch", "StartDate": "2026-01-01", "EndDate": nil},
	}
}

func adFixture() []map[string]any {
	return []map[string]any{
		{"Id": "6000001", "Type": "ResponsiveSearch", "Status": "Active", "EditorialStatus": "Active", "DevicePreference": "All"},
	}
}
