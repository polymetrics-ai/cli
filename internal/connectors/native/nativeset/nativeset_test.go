package nativeset

import (
	"testing"

	"polymetrics.ai/internal/connectors"
)

func TestPromotedFactoryManifestUsesBundleDefinition(t *testing.T) {
	var connector connectors.Connector
	for _, factory := range Factories() {
		if factory.Name == "aws-cloudtrail" {
			connector = factory.New()
			break
		}
	}
	if connector == nil {
		t.Fatal("Factories() missing aws-cloudtrail")
	}

	def, ok := connectors.DefinitionOf(connector)
	if !ok {
		t.Fatal("aws-cloudtrail connector does not implement DefinitionProvider")
	}

	manifest := connectors.ManifestOf(connector)
	if len(manifest.Streams) != len(def.Streams) {
		t.Fatalf("aws-cloudtrail manifest streams = %d, want %d (from bundle definition)", len(manifest.Streams), len(def.Streams))
	}
	if !hasConfigField(manifest.ConfigFields, "aws_region_name") {
		t.Fatalf("aws-cloudtrail manifest config fields = %+v, want aws_region_name", manifest.ConfigFields)
	}
	if !hasSecretField(manifest.SecretFields, "aws_key_id") || !hasSecretField(manifest.SecretFields, "aws_secret_key") {
		t.Fatalf("aws-cloudtrail manifest secret fields = %+v, want aws_key_id and aws_secret_key", manifest.SecretFields)
	}
}

func hasConfigField(fields []connectors.ConfigField, name string) bool {
	for _, field := range fields {
		if field.Name == name {
			return true
		}
	}
	return false
}

func hasSecretField(fields []connectors.SecretField, name string) bool {
	for _, field := range fields {
		if field.Name == name {
			return true
		}
	}
	return false
}

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
