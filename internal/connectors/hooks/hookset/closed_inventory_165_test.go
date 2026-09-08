package hookset

import (
	"slices"
	"sort"
	"testing"
)

// Expected members are frozen pre-edit at29b790d7, not generated from Factories.
func TestClosedHookInventory165(t *testing.T) {
	expected := []string{"akeneo", "amazon-ads", "amazon-seller-partner", "apple-search-ads", "appsflyer", "blogger", "chift", "ebay-fulfillment", "elasticsearch", "feishu", "github", "gmail", "google-ads", "google-calendar", "google-forms", "google-search-console", "hoorayhr", "jamf-pro", "keka", "microsoft-dataverse", "microsoft-entra-id", "microsoft-lists", "microsoft-teams", "mixpanel", "monday", "netsuite", "nexus-datasets", "notion", "outlook", "paypal-transaction", "pinterest", "plaid", "quickbooks", "rss", "salesloft", "sentry", "serpstat", "slack", "smartsheets", "snapchat-marketing", "stigg", "strava", "twilio", "uptick", "us-census", "wasabi-stats-api", "youtube-analytics", "zoho-analytics-metadata-api", "zoho-bigin"}
	var actual []string
	for _, f := range Factories() {
		actual = append(actual, f.Connector)
		if f.ID != "hook/"+f.Connector+".v1" || f.New == nil {
			t.Fatalf("invalid closed factory binding: %+v", f)
		}
		t.Run(f.Connector, func(t *testing.T) {
			h := f.New()
			if h == nil || h.ConnectorName() != f.Connector {
				t.Fatalf("constructor returned wrong owner for %s", f.ID)
			}
		})
	}
	sort.Strings(actual)
	if !slices.Equal(actual, expected) {
		t.Fatalf("closed hook membership changed: got=%v want=%v", actual, expected)
	}
	wrong := slices.Clone(actual)
	wrong[0] = "same-count-wrong-owner"
	if slices.Equal(wrong, expected) {
		t.Fatal("same-count membership substitution escaped oracle")
	}
}
