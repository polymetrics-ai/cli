package nativeset

import (
	"context"
	"fmt"

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
	"polymetrics.ai/internal/connectors/engine"
)

type definitionConnector struct {
	connectors.Connector
	base engine.Base
}

func (c definitionConnector) Definition() connectors.Definition {
	return c.base.Definition()
}

type cloudTrailDefinitionConnector struct{ definitionConnector }

func (c cloudTrailDefinitionConnector) Manifest() connectors.Manifest {
	return c.base.Manifest()
}

func (c cloudTrailDefinitionConnector) CommandSurface() *connectors.CommandSurface {
	return c.base.CommandSurface()
}

func (c cloudTrailDefinitionConnector) OperationDirectRead(ctx context.Context, req connectors.OperationDirectReadRequest) (connectors.DirectReadResult, error) {
	reader, ok := c.Connector.(connectors.OperationDirectReader)
	if !ok {
		return connectors.DirectReadResult{}, fmt.Errorf("connector %s does not support operation direct read", c.Name())
	}
	return reader.OperationDirectRead(ctx, req)
}

func (c cloudTrailDefinitionConnector) ValidateWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) error {
	validator, ok := c.Connector.(connectors.WriteValidator)
	if !ok {
		return nil
	}
	return validator.ValidateWrite(ctx, req, records)
}

func (c cloudTrailDefinitionConnector) DryRunWrite(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WritePreview, error) {
	if dryRun, ok := c.Connector.(connectors.DryRunWriter); ok {
		return dryRun.DryRunWrite(ctx, req, records)
	}
	return connectors.WritePreview{}, fmt.Errorf("connector %s does not support reverse ETL previews", c.Name())
}

func (c cloudTrailDefinitionConnector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	stateful, ok := c.Connector.(connectors.StatefulReader)
	if !ok {
		return nil, nil
	}
	return stateful.InitialState(ctx, stream, cfg)
}

func newDefinitionConnector(name string, c connectors.Connector) definitionConnector {
	bundle, err := engine.Load(defs.FS, name)
	if err != nil {
		panic("native/" + name + ": failed to load defs/" + name + " bundle: " + err.Error())
	}
	return definitionConnector{Connector: c, base: engine.NewBase(bundle)}
}

func withBundleDefinition(name string, c connectors.Connector) connectors.Connector {
	return newDefinitionConnector(name, c)
}

func promotedFactories() []Factory {
	return []Factory{
		{Name: "alpha-vantage", New: func() connectors.Connector { return withBundleDefinition("alpha-vantage", alphavantage.New()) }},
		{Name: "apify-dataset", New: func() connectors.Connector { return withBundleDefinition("apify-dataset", apifydataset.New()) }},
		{Name: "ashby", New: func() connectors.Connector { return withBundleDefinition("ashby", ashby.New()) }},
		{Name: "aws-cloudtrail", New: func() connectors.Connector {
			return cloudTrailDefinitionConnector{definitionConnector: newDefinitionConnector("aws-cloudtrail", awscloudtrail.New())}
		}},
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
		{Name: "google-calendar", New: func() connectors.Connector { return withBundleDefinition("google-calendar", googlecalendar.New()) }},
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
