package nativeset

import (
	"context"
	"fmt"
	"io/fs"
	"strings"

	"polymetrics.ai/internal/connectors"
	alphavantage "polymetrics.ai/internal/connectors/native/alpha-vantage"
	apifydataset "polymetrics.ai/internal/connectors/native/apify-dataset"
	"polymetrics.ai/internal/connectors/native/ashby"
	awscloudtrail "polymetrics.ai/internal/connectors/native/aws-cloudtrail"
	"polymetrics.ai/internal/connectors/native/babelforce"
	"polymetrics.ai/internal/connectors/native/basecamp"
	bunnyinc "polymetrics.ai/internal/connectors/native/bunny-inc"
	"polymetrics.ai/internal/connectors/native/canny"
	"polymetrics.ai/internal/connectors/native/copper"
	"polymetrics.ai/internal/connectors/native/dixa"
	"polymetrics.ai/internal/connectors/native/fastbill"
	"polymetrics.ai/internal/connectors/native/feishu"
	freeagentconnector "polymetrics.ai/internal/connectors/native/free-agent-connector"
	"polymetrics.ai/internal/connectors/native/freightview"
	googleanalyticsdataapi "polymetrics.ai/internal/connectors/native/google-analytics-data-api"
	googlecalendar "polymetrics.ai/internal/connectors/native/google-calendar"
	googleclassroom "polymetrics.ai/internal/connectors/native/google-classroom"
	googlepagespeedinsights "polymetrics.ai/internal/connectors/native/google-pagespeed-insights"
	lessannoyingcrm "polymetrics.ai/internal/connectors/native/less-annoying-crm"
	"polymetrics.ai/internal/connectors/native/lokalise"
	"polymetrics.ai/internal/connectors/native/mendeley"
	mercadoads "polymetrics.ai/internal/connectors/native/mercado-ads"
	"polymetrics.ai/internal/connectors/native/metabase"
	nativemode "polymetrics.ai/internal/connectors/native/mode"
	myhours "polymetrics.ai/internal/connectors/native/my-hours"
	"polymetrics.ai/internal/connectors/native/pocket"
	"polymetrics.ai/internal/connectors/native/prestashop"
	"polymetrics.ai/internal/connectors/native/rootly"
	"polymetrics.ai/internal/connectors/native/safetyculture"
	yahoofinanceprice "polymetrics.ai/internal/connectors/native/yahoo-finance-price"

	"polymetrics.ai/internal/connectors/defs"
	googlecalendardefs "polymetrics.ai/internal/connectors/defs/google-calendar"
	"polymetrics.ai/internal/connectors/engine"
)

type definitionConnector struct {
	connectors.Connector
	base engine.Base
}

type fixtureModeEngineConnector struct {
	*engine.Connector
	fixture connectors.Connector
}

func (c definitionConnector) Definition() connectors.Definition {
	return c.base.Definition()
}

// HasConfigurationConstraints forwards the wrapped bundle's actual
// configuration-constraint declaration. It prevents promoted native
// connectors from advertising a generic validator when their bundle has no
// constraint to evaluate.
func (c definitionConnector) HasConfigurationConstraints() bool {
	return c.base.HasConfigurationConstraints()
}

// ValidateConfiguration delegates declared configuration checks to the
// wrapped bundle's compiled schema.
func (c definitionConnector) ValidateConfiguration(config map[string]string) error {
	return c.base.ValidateConfiguration(config)
}

func (c *fixtureModeEngineConnector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if fixtureMode(cfg) {
		return c.fixture.Check(ctx, cfg)
	}
	return c.Connector.Check(ctx, cfg)
}

func (c *fixtureModeEngineConnector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if fixtureMode(req.Config) {
		return c.ReadFixture(ctx, req, emit)
	}
	return c.Connector.Read(ctx, req, emit)
}

func (c *fixtureModeEngineConnector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	if fixtureMode(req.Config) {
		return connectors.WriteResult{RecordsFailed: len(records)}, fixtureModeUnsupported(ctx, c.Name(), "writes")
	}
	return c.Connector.Write(ctx, req, records)
}

func (c *fixtureModeEngineConnector) ValidateWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) error {
	if fixtureMode(req.Config) {
		return fixtureModeUnsupported(ctx, c.Name(), "write validation")
	}
	return c.Connector.ValidateWrite(ctx, req, records)
}

func (c *fixtureModeEngineConnector) DryRunWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WritePreview, error) {
	if fixtureMode(req.Config) {
		return connectors.WritePreview{}, fixtureModeUnsupported(ctx, c.Name(), "write previews")
	}
	return c.Connector.DryRunWrite(ctx, req, records)
}

func (c *fixtureModeEngineConnector) DirectRead(ctx context.Context, req connectors.DirectReadRequest) (connectors.DirectReadResult, error) {
	if fixtureMode(req.Config) {
		return connectors.DirectReadResult{}, fixtureModeUnsupported(ctx, c.Name(), "direct reads")
	}
	return c.Connector.DirectRead(ctx, req)
}

func (c *fixtureModeEngineConnector) OperationDirectRead(ctx context.Context, req connectors.OperationDirectReadRequest) (connectors.DirectReadResult, error) {
	if fixtureMode(req.Config) {
		return connectors.DirectReadResult{}, fixtureModeUnsupported(ctx, c.Name(), "operation direct reads")
	}
	return c.Connector.OperationDirectRead(ctx, req)
}

func (c *fixtureModeEngineConnector) OperationBinaryDownload(ctx context.Context, req connectors.OperationBinaryDownloadRequest) (connectors.OperationBinaryDownloadResult, error) {
	if fixtureMode(req.Config) {
		return connectors.OperationBinaryDownloadResult{}, fixtureModeUnsupported(ctx, c.Name(), "binary downloads")
	}
	return c.Connector.OperationBinaryDownload(ctx, req)
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}

func fixtureModeUnsupported(ctx context.Context, name, operation string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return fmt.Errorf("%s fixture mode does not support %s: %w", name, operation, connectors.ErrUnsupportedOperation)
}

func loadPromotedBundle(name string) engine.Bundle {
	bundle, err := engine.Load(defs.FS, name)
	if err != nil {
		panic("native/" + name + ": failed to load defs/" + name + " bundle: " + err.Error())
	}
	return bundle
}

func withBundleDefinition(name string, c connectors.Connector) connectors.Connector {
	bundle := loadPromotedBundle(name)
	return definitionConnector{Connector: c, base: engine.NewBase(bundle)}
}

func withFixtureModeEngineBundle(name string, fixture connectors.Connector, fixtures fs.FS) connectors.Connector {
	bundle := loadPromotedBundle(name)
	bundle.Fixtures = fixtures
	return &fixtureModeEngineConnector{Connector: engine.New(bundle, engine.HooksFor(name)), fixture: fixture}
}

func promotedFactories() []Factory {
	return []Factory{
		{Name: "alpha-vantage", New: func() connectors.Connector { return withBundleDefinition("alpha-vantage", alphavantage.New()) }},
		{Name: "apify-dataset", New: func() connectors.Connector { return withBundleDefinition("apify-dataset", apifydataset.New()) }},
		{Name: "ashby", New: func() connectors.Connector { return ashby.New() }},
		{Name: "aws-cloudtrail", New: func() connectors.Connector { return withBundleDefinition("aws-cloudtrail", awscloudtrail.New()) }},
		{Name: "babelforce", New: func() connectors.Connector { return withBundleDefinition("babelforce", babelforce.New()) }},
		{Name: "basecamp", New: func() connectors.Connector { return withBundleDefinition("basecamp", basecamp.New()) }},
		{Name: "bunny-inc", New: func() connectors.Connector { return withBundleDefinition("bunny-inc", bunnyinc.New()) }},
		{Name: "canny", New: func() connectors.Connector { return withBundleDefinition("canny", canny.New()) }},
		{Name: "copper", New: func() connectors.Connector { return withBundleDefinition("copper", copper.New()) }},
		{Name: "dixa", New: func() connectors.Connector { return withBundleDefinition("dixa", dixa.New()) }},
		{Name: "fastbill", New: func() connectors.Connector { return withBundleDefinition("fastbill", fastbill.New()) }},
		{Name: "feishu", New: func() connectors.Connector { return withBundleDefinition("feishu", feishu.New()) }},
		{Name: "free-agent-connector", New: func() connectors.Connector {
			return withBundleDefinition("free-agent-connector", freeagentconnector.New())
		}},
		{Name: "freightview", New: func() connectors.Connector { return withBundleDefinition("freightview", freightview.New()) }},
		{Name: "google-analytics-data-api", New: func() connectors.Connector {
			return withBundleDefinition("google-analytics-data-api", googleanalyticsdataapi.New())
		}},
		{Name: "google-calendar", New: func() connectors.Connector {
			fixtures, err := googlecalendardefs.Fixtures()
			if err != nil {
				panic("native/google-calendar: failed to load embedded fixtures: " + err.Error())
			}
			return withFixtureModeEngineBundle("google-calendar", googlecalendar.New(), fixtures)
		}},
		{Name: "google-classroom", New: func() connectors.Connector { return withBundleDefinition("google-classroom", googleclassroom.New()) }},
		{Name: "google-pagespeed-insights", New: func() connectors.Connector {
			return withBundleDefinition("google-pagespeed-insights", googlepagespeedinsights.New())
		}},
		{Name: "less-annoying-crm", New: func() connectors.Connector { return withBundleDefinition("less-annoying-crm", lessannoyingcrm.New()) }},
		{Name: "lokalise", New: func() connectors.Connector { return withBundleDefinition("lokalise", lokalise.New()) }},
		{Name: "mendeley", New: func() connectors.Connector { return withBundleDefinition("mendeley", mendeley.New()) }},
		{Name: "mercado-ads", New: func() connectors.Connector { return withBundleDefinition("mercado-ads", mercadoads.New()) }},
		{Name: "metabase", New: func() connectors.Connector { return withBundleDefinition("metabase", metabase.New()) }},
		{Name: "mode", New: func() connectors.Connector { return withBundleDefinition("mode", nativemode.New()) }},
		{Name: "my-hours", New: func() connectors.Connector { return withBundleDefinition("my-hours", myhours.New()) }},
		{Name: "pocket", New: func() connectors.Connector { return withBundleDefinition("pocket", pocket.New()) }},
		{Name: "prestashop", New: func() connectors.Connector { return withBundleDefinition("prestashop", prestashop.New()) }},
		{Name: "rootly", New: func() connectors.Connector { return withBundleDefinition("rootly", rootly.New()) }},
		{Name: "safetyculture", New: func() connectors.Connector { return withBundleDefinition("safetyculture", safetyculture.New()) }},
		{Name: "yahoo-finance-price", New: func() connectors.Connector {
			return withBundleDefinition("yahoo-finance-price", yahoofinanceprice.New())
		}},
	}
}
