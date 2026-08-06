package awscloudtrail

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const legacyFixtureEventTime int64 = 1767225600

type legacyEventStreamSpec struct {
	filterKey   string
	filterValue string
}

var legacyEventStreamSpecs = map[string]legacyEventStreamSpec{
	"management_events": {},
	"read_only_events":  {filterKey: "ReadOnly", filterValue: "true"},
	"write_only_events": {filterKey: "ReadOnly", filterValue: "false"},
	"console_logins":    {filterKey: "EventName", filterValue: "ConsoleLogin"},
}

func legacyEventStream(stream string) (legacyEventStreamSpec, bool) {
	spec, ok := legacyEventStreamSpecs[stream]
	return spec, ok
}

func (c Connector) readLegacyEventStream(ctx context.Context, stream string, spec legacyEventStreamSpec, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if fixtureMode(req.Config) {
		return c.readLegacyEventFixture(ctx, stream, req, emit)
	}
	r, err := c.requester(req.Config, "LookupEvents")
	if err != nil {
		return err
	}
	maxItems, err := pageSizeForAction("LookupEvents", req.Config)
	if err != nil {
		return err
	}
	startTime, err := legacyEventStartTime(req)
	if err != nil {
		return err
	}
	return c.harvestLegacyEventStream(ctx, r, spec, maxItems, startTime, req.Config, emit)
}

func (c Connector) harvestLegacyEventStream(ctx context.Context, r *connsdk.Requester, spec legacyEventStreamSpec, maxItems int, startTime *time.Time, cfg connectors.RuntimeConfig, emit func(connectors.Record) error) error {
	lookupAttributes := legacyLookupAttributes(spec, cfg)
	maxPageLimit, err := maxPages(cfg)
	if err != nil {
		return err
	}
	nextToken := ""
	seenTokens := map[string]bool{}
	for page := 0; maxPageLimit == 0 || page < maxPageLimit; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		body := map[string]any{"MaxResults": maxItems}
		if len(lookupAttributes) > 0 {
			body["LookupAttributes"] = lookupAttributes
		}
		if startTime != nil {
			body["StartTime"] = startTime.Unix()
		}
		if nextToken != "" {
			body["NextToken"] = nextToken
		}
		resp, err := r.Do(ctx, http.MethodPost, "/", nil, body)
		if err != nil {
			return fmt.Errorf("read aws-cloudtrail LookupEvents: %w", err)
		}
		records, err := connsdk.RecordsAt(resp.Body, "Events")
		if err != nil {
			return fmt.Errorf("decode aws-cloudtrail Events page: %w", err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(legacyEventRecord(map[string]any(item))); err != nil {
				return err
			}
		}
		nextToken, err = connsdk.StringAt(resp.Body, "NextToken")
		if err != nil {
			return fmt.Errorf("decode aws-cloudtrail NextToken: %w", err)
		}
		nextToken = strings.TrimSpace(nextToken)
		if nextToken == "" || maxPageLimit > 0 && page+1 >= maxPageLimit {
			return nil
		}
		if seenTokens[nextToken] {
			return fmt.Errorf("read aws-cloudtrail LookupEvents: %w %q", errRepeatedPaginationToken, nextToken)
		}
		seenTokens[nextToken] = true
	}
	return nil
}

func (c Connector) readLegacyEventFixture(ctx context.Context, stream string, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	for index := 1; index <= 2; index++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		record := legacyEventRecord(map[string]any{
			"EventId":         fmt.Sprintf("%s_fixture_%d", stream, index),
			"EventName":       "ConsoleLogin",
			"EventSource":     "signin.amazonaws.com",
			"EventTime":       legacyFixtureEventTime + int64(index),
			"Username":        fmt.Sprintf("fixture-user-%d", index),
			"AccessKeyId":     "AKIAFIXTURE000000000",
			"ReadOnly":        "true",
			"Resources":       []any{},
			"CloudTrailEvent": `{"eventVersion":"1.08","fixture":true}`,
		})
		if cursor := connsdk.Cursor(req.State); cursor != "" {
			record["previous_cursor"] = cursor
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	return nil
}

func legacyLookupAttributes(spec legacyEventStreamSpec, cfg connectors.RuntimeConfig) []map[string]string {
	key, value := spec.filterKey, spec.filterValue
	if cfg.Config != nil {
		if override := strings.TrimSpace(cfg.Config["lookup_attribute_key"]); override != "" {
			key = override
			value = strings.TrimSpace(cfg.Config["lookup_attribute_value"])
		}
	}
	if key == "" {
		return nil
	}
	return []map[string]string{{"AttributeKey": key, "AttributeValue": value}}
}

func legacyEventStartTime(req connectors.ReadRequest) (*time.Time, error) {
	if cursor := connsdk.Cursor(req.State); cursor != "" {
		seconds, err := strconv.ParseInt(cursor, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("aws-cloudtrail cursor must be unix seconds: %w", err)
		}
		value := time.Unix(seconds, 0).UTC()
		return &value, nil
	}
	startDate := ""
	if req.Config.Config != nil {
		startDate = strings.TrimSpace(req.Config.Config["start_date"])
	}
	if startDate == "" {
		return nil, nil
	}
	value, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		value, err = time.Parse(time.RFC3339, startDate)
		if err != nil {
			return nil, fmt.Errorf("aws-cloudtrail config start_date must be YYYY-MM-DD: %w", err)
		}
	}
	value = value.UTC()
	return &value, nil
}

func legacyEventRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"EventId":         item["EventId"],
		"EventName":       item["EventName"],
		"EventSource":     item["EventSource"],
		"EventTime":       item["EventTime"],
		"Username":        item["Username"],
		"AccessKeyId":     item["AccessKeyId"],
		"ReadOnly":        item["ReadOnly"],
		"Resources":       item["Resources"],
		"CloudTrailEvent": item["CloudTrailEvent"],
	}
}
