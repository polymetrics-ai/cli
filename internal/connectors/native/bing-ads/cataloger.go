package bingads

import (
	"context"
	"fmt"

	"polymetrics.ai/internal/connectors"
)

// Catalog returns the connector's published stream catalog. Bing Ads'
// 5-stream schema is statically known ahead of time (not discovered at
// runtime from the live API, unlike postgres' information_schema
// discovery), matching legacy bing_ads/streams.go's bingStreams(). Bing Ads
// objects are keyed by a numeric Id (carried as a string in the JSON REST
// surface), so the primary key is ["Id"] across the board. These are
// full-refresh streams (the upstream catalog only advertises full_refresh),
// so no cursor fields are published.
func (c Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: c.Name(), Streams: catalogStreams()}, nil
}

func catalogStreams() []connectors.Stream {
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
			Description: "Campaigns within the configured advertiser accounts.",
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

// mapRecord flattens a raw Bing Ads object into a connectors.Record for the
// named stream (mirrors legacy streams.go's per-stream mapRecord table).
func mapRecord(stream string, item map[string]any) connectors.Record {
	switch stream {
	case "users":
		return userRecord(item)
	case "campaigns":
		return campaignRecord(item)
	case "ad_groups":
		return adGroupRecord(item)
	case "ads":
		return adRecord(item)
	default:
		return accountRecord(item)
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

// readFixture emits deterministic records without any network access so the
// conformance harness can exercise bing-ads credential-free (mirrors legacy
// bing_ads.go:177-193).
func (c Connector) readFixture(ctx context.Context, stream string, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	items, err := fixtureRecords(stream)
	if err != nil {
		return err
	}
	for _, item := range items {
		if err := ctx.Err(); err != nil {
			return err
		}
		record := mapRecord(stream, item)
		if cursor := req.State["cursor"]; cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

func fixtureRecords(stream string) ([]map[string]any, error) {
	switch stream {
	case "accounts":
		return accountFixture(), nil
	case "users":
		return userFixture(), nil
	case "campaigns":
		return campaignFixture(), nil
	case "ad_groups":
		return adGroupFixture(), nil
	case "ads":
		return adFixture(), nil
	default:
		return nil, fmt.Errorf("bing-ads stream %q not found", stream)
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
