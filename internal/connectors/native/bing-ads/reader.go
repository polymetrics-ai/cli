package bingads

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

// streamRoute maps a stream to its REST POST endpoint, the JSON path to the
// records array in the response, the request-body builder, and the service
// kind. Mirrors legacy streams.go's streamEndpoint (minus mapRecord/fixture,
// which live in cataloger.go alongside the rest of the catalog/fixture
// logic).
//
// perAccount marks campaign-scoped streams whose query is keyed by
// AccountId; the read path iterates the configured account ids and merges
// the results so a sync can span every account in the configured set.
type streamRoute struct {
	path        string
	recordsPath string
	kind        serviceKind
	perAccount  bool
	body        func(cfg connectors.RuntimeConfig, accountID string) map[string]any
}

// streamRoutes is the per-stream routing table. Adding a stream means
// adding one entry here plus a Stream definition in catalogStreams
// (cataloger.go); the read path is fully data-driven from this table,
// mirroring legacy's identical design (streams.go's doc comment).
var streamRoutes = map[string]streamRoute{
	"accounts": {
		path:        "/AccountsInfo/Query",
		recordsPath: "AccountsInfo",
		kind:        serviceCustomer,
		body: func(cfg connectors.RuntimeConfig, _ string) map[string]any {
			b := map[string]any{"OnlyParentAccounts": false}
			if v := strings.TrimSpace(cfg.Config["customer_id"]); v != "" {
				b["CustomerId"] = v
			}
			return b
		},
	},
	"users": {
		path:        "/User/Query",
		recordsPath: "User",
		kind:        serviceCustomer,
		body: func(_ connectors.RuntimeConfig, _ string) map[string]any {
			return map[string]any{"UserId": nil}
		},
	},
	"campaigns": {
		path:        "/Campaigns/QueryByAccountId",
		recordsPath: "Campaigns",
		kind:        serviceCampaign,
		perAccount:  true,
		body: func(_ connectors.RuntimeConfig, accountID string) map[string]any {
			b := map[string]any{}
			if accountID != "" {
				b["AccountId"] = accountID
			}
			return b
		},
	},
	"ad_groups": {
		path:        "/AdGroups/QueryByCampaignId",
		recordsPath: "AdGroups",
		kind:        serviceCampaign,
		body: func(cfg connectors.RuntimeConfig, _ string) map[string]any {
			b := map[string]any{}
			if v := strings.TrimSpace(cfg.Config["campaign_id"]); v != "" {
				b["CampaignId"] = v
			}
			return b
		},
	},
	"ads": {
		path:        "/Ads/QueryByAdGroupId",
		recordsPath: "Ads",
		kind:        serviceCampaign,
		body: func(cfg connectors.RuntimeConfig, _ string) map[string]any {
			b := map[string]any{}
			if v := strings.TrimSpace(cfg.Config["ad_group_id"]); v != "" {
				b["AdGroupId"] = v
			}
			return b
		},
	},
}

// InitialState satisfies connectors.StatefulReader: every Bing Ads stream is
// full_refresh only (legacy's catalog publishes no CursorFields), so the
// cursor is always empty.
func (c Connector) InitialState(ctx context.Context, stream string, cfg connectors.RuntimeConfig) (map[string]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return connsdk.WithCursor(map[string]string{"stream": stream}, ""), nil
}

// Read dispatches to the fixture path (credential-free conformance) or the
// live per-stream routing table, mirroring legacy bing_ads.go:113-168.
func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "accounts"
	}
	route, ok := streamRoutes[stream]
	if !ok {
		return fmt.Errorf("bing-ads stream %q not found", stream)
	}

	if fixtureMode(req.Config) {
		return c.readFixture(ctx, stream, req, emit)
	}

	r, err := c.requester(req.Config, route.kind)
	if err != nil {
		return err
	}

	// Per-account streams (e.g. campaigns) iterate the configured account
	// ids so a sync can span every account in the set. A single empty id is
	// used when no account ids are configured, which lets the request body
	// carry whatever account the headers already scope it to.
	accountIDs := []string{""}
	if route.perAccount {
		if ids := accountIDList(req.Config); len(ids) > 0 {
			accountIDs = ids
		}
	}

	for _, accountID := range accountIDs {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := r.Do(ctx, http.MethodPost, route.path, nil, route.body(req.Config, accountID))
		if err != nil {
			return fmt.Errorf("read bing-ads %s: %w", stream, err)
		}
		records, err := connsdk.RecordsAt(resp.Body, route.recordsPath)
		if err != nil {
			return fmt.Errorf("decode bing-ads %s: %w", stream, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(mapRecord(stream, item)); err != nil {
				return err
			}
		}
	}
	return nil
}
