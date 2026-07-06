// Package slack implements the slack bundle's two Tier-2 hooks
// (conventions.md §1's 2-interface cap): CheckHook + StreamHook.
//
// Slack's Web API always returns HTTP 200 even for logical failures,
// signaling errors solely via a JSON body field ({"ok":false,"error":"<code>"})
// -- auth failures, invalid_auth, token_revoked, missing_scope, etc. all
// surface this way, never as a non-2xx HTTP status
// (docs/migration/quarantine.json's original slack blocker). The engine's
// declarative read/check paths only ever treat a non-2xx HTTP status as a
// failure, with no mechanism to inspect a response BODY field as a
// stop/error condition, so both Check and every stream's read are hook-side
// here, porting legacy internal/connectors/slack/{slack.go,streams.go}'s
// harvest/slackOK/record-mapper logic verbatim.
//
// Auth stays declarative (base.auth's bearer candidates in streams.json):
// legacy's real, catalog-documented auth is a static bot/user/API token
// (website/.enrich/enr/source-slack.json), not an OAuth 2.0 refresh-token
// grant -- there is no repeatable token-exchange for an AuthHook to
// implement, so this package implements no AuthHook (see defs/slack/docs.md
// "Auth setup").
package slack

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

const (
	defaultPageSize = 200
	maxPageSize     = 1000
	safetyMaxPages  = 10000
)

func init() {
	engine.RegisterHooks("slack", func() engine.Hooks { return Hooks{} })
}

// Hooks is the slack hook set: CheckHook + StreamHook.
type Hooks struct{}

func (Hooks) ConnectorName() string { return "slack" }

// --- CheckHook --------------------------------------------------------------

// Check implements engine.CheckHook: calls auth.test and inspects the
// response body's ok field (Slack's primary error-signaling mechanism),
// exactly like legacy's Check (slack.go:63-88).
func (Hooks) Check(ctx context.Context, cfg connectors.RuntimeConfig, rt *engine.Runtime) (bool, error) {
	resp, err := rt.Requester.Do(ctx, http.MethodGet, "auth.test", nil, nil)
	if err != nil {
		return true, fmt.Errorf("check slack: %w", err)
	}
	if err := slackOK(resp.Body, "auth.test"); err != nil {
		return true, fmt.Errorf("check slack: %w", err)
	}
	return true, nil
}

// --- StreamHook ---------------------------------------------------------

// streamEndpoint maps a stream name to the Slack Web API method path, the
// JSON list key its records live under, and the record mapper (ported
// verbatim from legacy streams.go's streamEndpoint/slackStreamEndpoints).
type streamEndpoint struct {
	resource  string
	listKey   string
	mapRecord func(map[string]any) connectors.Record
}

var streamEndpoints = map[string]streamEndpoint{
	"users":            {resource: "users.list", listKey: "members", mapRecord: userRecord},
	"channels":         {resource: "conversations.list", listKey: "channels", mapRecord: channelRecord},
	"channel_messages": {resource: "conversations.history", listKey: "messages", mapRecord: messageRecord},
}

// ReadStream implements engine.StreamHook, handling every declared stream
// with handled=true; an unknown stream name returns handled=false so the
// engine can surface its own "stream not found" error.
func (h Hooks) ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (bool, error) {
	name := stream.Name
	if name == "" {
		name = "users"
	}
	endpoint, ok := streamEndpoints[name]
	if !ok {
		return false, nil
	}

	pageSize, err := pageSizeFrom(req.Config)
	if err != nil {
		return true, err
	}
	maxPages, err := maxPagesFrom(req.Config)
	if err != nil {
		return true, err
	}

	base := url.Values{}
	base.Set("limit", strconv.Itoa(pageSize))
	if name == "channel_messages" {
		channel := strings.TrimSpace(req.Config.Config["channel_id"])
		if channel == "" {
			return true, errors.New("slack stream channel_messages requires config channel_id")
		}
		base.Set("channel", channel)
	}

	return true, h.harvest(ctx, rt.Requester, endpoint, base, maxPages, emit)
}

// harvest ports legacy slack.go's harvest loop verbatim: Slack's cursor
// pagination (response_metadata.next_cursor -> cursor param), with every
// page checked via slackOK before its records/cursor are trusted.
func (h Hooks) harvest(ctx context.Context, r *connsdk.Requester, endpoint streamEndpoint, base url.Values, maxPages int, emit func(connectors.Record) error) error {
	limit := maxPages
	if limit <= 0 {
		limit = safetyMaxPages
	}
	cursor := ""
	for page := 0; page < limit; page++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		query := cloneValues(base)
		if cursor != "" {
			query.Set("cursor", cursor)
		}
		resp, err := r.Do(ctx, http.MethodGet, endpoint.resource, query, nil)
		if err != nil {
			return fmt.Errorf("read slack %s: %w", endpoint.resource, err)
		}
		if err := slackOK(resp.Body, endpoint.resource); err != nil {
			return err
		}
		records, err := connsdk.RecordsAt(resp.Body, endpoint.listKey)
		if err != nil {
			return fmt.Errorf("decode slack %s page: %w", endpoint.resource, err)
		}
		for _, item := range records {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := emit(endpoint.mapRecord(item)); err != nil {
				return err
			}
		}
		next, err := connsdk.StringAt(resp.Body, "response_metadata.next_cursor")
		if err != nil {
			return fmt.Errorf("decode slack %s next_cursor: %w", endpoint.resource, err)
		}
		if strings.TrimSpace(next) == "" {
			return nil
		}
		cursor = next
	}
	return nil
}

// slackOK inspects a Slack Web API response body for the ok flag. Slack
// returns HTTP 200 even for logical failures, carrying {ok:false,
// error:"<code>"} (ported verbatim from legacy slack.go:255-268).
func slackOK(body []byte, method string) error {
	ok, err := connsdk.StringAt(body, "ok")
	if err != nil {
		return fmt.Errorf("decode slack %s response: %w", method, err)
	}
	if ok == "true" {
		return nil
	}
	code, _ := connsdk.StringAt(body, "error")
	if strings.TrimSpace(code) == "" {
		code = "unknown_error"
	}
	return fmt.Errorf("slack %s returned error: %s", method, code)
}

func pageSizeFrom(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(cfg.Config["page_size"])
	if raw == "" {
		return defaultPageSize, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("slack config page_size must be an integer: %w", err)
	}
	if value < 1 || value > maxPageSize {
		return 0, fmt.Errorf("slack config page_size must be between 1 and %d", maxPageSize)
	}
	return value, nil
}

func maxPagesFrom(cfg connectors.RuntimeConfig) (int, error) {
	raw := strings.TrimSpace(strings.ToLower(cfg.Config["max_pages"]))
	if raw == "" || raw == "all" || raw == "unlimited" {
		return 0, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, errors.New("slack config max_pages must be a non-negative integer, all, or unlimited")
	}
	return value, nil
}

func cloneValues(in url.Values) url.Values {
	out := url.Values{}
	for k, vs := range in {
		for _, v := range vs {
			out.Add(k, v)
		}
	}
	return out
}

// --- record mappers (ported verbatim from legacy streams.go) --------------

func userRecord(item map[string]any) connectors.Record {
	profile, _ := item["profile"].(map[string]any)
	return connectors.Record{
		"id":           item["id"],
		"team_id":      item["team_id"],
		"name":         item["name"],
		"real_name":    item["real_name"],
		"display_name": profileField(profile, "display_name"),
		"email":        profileField(profile, "email"),
		"deleted":      item["deleted"],
		"is_admin":     item["is_admin"],
		"is_bot":       item["is_bot"],
		"tz":           item["tz"],
		"updated":      item["updated"],
	}
}

func channelRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"id":          item["id"],
		"name":        item["name"],
		"is_channel":  item["is_channel"],
		"is_group":    item["is_group"],
		"is_private":  item["is_private"],
		"is_archived": item["is_archived"],
		"is_general":  item["is_general"],
		"created":     item["created"],
		"creator":     item["creator"],
		"num_members": item["num_members"],
		"topic":       nestedValue(item, "topic", "value"),
		"purpose":     nestedValue(item, "purpose", "value"),
	}
}

func messageRecord(item map[string]any) connectors.Record {
	return connectors.Record{
		"ts":          item["ts"],
		"type":        item["type"],
		"subtype":     item["subtype"],
		"user":        item["user"],
		"text":        item["text"],
		"thread_ts":   item["thread_ts"],
		"reply_count": item["reply_count"],
		"team":        item["team"],
	}
}

func profileField(profile map[string]any, key string) any {
	if profile == nil {
		return nil
	}
	return profile[key]
}

func nestedValue(item map[string]any, outer, inner string) any {
	obj, ok := item[outer].(map[string]any)
	if !ok {
		return nil
	}
	return obj[inner]
}
