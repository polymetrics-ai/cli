package nativeset

import (
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestFactoriesExposeDefinitions(t *testing.T) {
	want := map[string]bool{
		"alpha-vantage":             false,
		"amazon-sqs":                false,
		"apify-dataset":             false,
		"ashby":                     false,
		"aws-cloudtrail":            false,
		"babelforce":                false,
		"basecamp":                  false,
		"bing-ads":                  false,
		"bunny-inc":                 false,
		"canny":                     false,
		"copper":                    false,
		"dixa":                      false,
		"dynamodb":                  false,
		"faker":                     false,
		"fastbill":                  false,
		"feishu":                    false,
		"free-agent-connector":      false,
		"freightview":               false,
		"google-analytics-data-api": false,
		"google-calendar":           false,
		"google-classroom":          false,
		"google-pagespeed-insights": false,
		"less-annoying-crm":         false,
		"lokalise":                  false,
		"mendeley":                  false,
		"mercado-ads":               false,
		"metabase":                  false,
		"mode":                      false,
		"my-hours":                  false,
		"pocket":                    false,
		"postgres":                  false,
		"prestashop":                false,
		"rootly":                    false,
		"safetyculture":             false,
		"tally-prime":               false,
		"yahoo-finance-price":       false,
	}

	for _, factory := range Factories() {
		if factory.New == nil {
			t.Fatalf("factory %q New = nil", factory.Name)
		}
		c := factory.New()
		if c.Name() != factory.Name {
			t.Fatalf("factory %q New().Name() = %q", factory.Name, c.Name())
		}
		def, ok := connectors.DefinitionOf(c)
		if !ok {
			t.Fatalf("factory %q connector does not implement DefinitionProvider", factory.Name)
		}
		if def.Name != factory.Name {
			t.Fatalf("factory %q Definition().Name = %q", factory.Name, def.Name)
		}
		if _, tracked := want[factory.Name]; tracked {
			want[factory.Name] = true
		}
	}

	for name, seen := range want {
		if !seen {
			t.Fatalf("Factories() missing %q", name)
		}
	}
}
