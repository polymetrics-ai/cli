// Package rss implements the rss bundle's Tier-2 StreamHook + CheckHook
// (docs/migration/conventions.md §1/§6): every RSS feed response is an
// XML document (RFC-shaped RSS 2.0), decoded via encoding/xml — the
// engine's declarative read path (internal/connectors/engine/read.go) only
// ever decodes a JSON response body, so an XML-bodied read/check cannot be
// expressed in streams.json alone. This ports
// internal/connectors/rss/rss.go's load/itemRecord/channelRecord logic
// almost verbatim via StreamHook.ReadStream (both streams: items, channel)
// and CheckHook.Check, both reusing rt.Requester (the engine's already-built
// *connsdk.Requester: base URL/headers/auth already resolved from
// streams.json's base.url/base.headers).
package rss

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/engine"
)

func init() {
	engine.RegisterHooks("rss", func() engine.Hooks { return Hooks{} })
}

// Hooks is rss's Tier-2 hook set: StreamHook (both reads) + CheckHook.
type Hooks struct{}

func (Hooks) ConnectorName() string { return "rss" }

// rssDocument/rssChannel/rssItem mirror legacy rss.go's identically named
// types field-for-field.
type rssDocument struct {
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Updated     string    `xml:"lastBuildDate"`
	Items       []rssItem `xml:"item"`
}

type rssItem struct {
	GUID        string `xml:"guid"`
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Published   string `xml:"pubDate"`
}

// ReadStream implements engine.StreamHook, handling both declared streams
// (items, channel) with handled=true always — the declarative streams.json
// fallback is never exercised in production (docs.md "Declarative path").
func (h Hooks) ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (bool, error) {
	name := stream.Name
	if name == "" {
		name = "items"
	}
	if name != "items" && name != "channel" {
		return false, nil
	}
	if err := ctx.Err(); err != nil {
		return true, err
	}

	doc, err := load(ctx, rt.Requester)
	if err != nil {
		return true, err
	}

	if name == "channel" {
		return true, emit(channelRecord(doc.Channel))
	}
	for _, item := range doc.Channel.Items {
		if err := ctx.Err(); err != nil {
			return true, err
		}
		if err := emit(itemRecord(item)); err != nil {
			return true, err
		}
	}
	return true, nil
}

// Check implements engine.CheckHook: legacy's Check also decodes the feed as
// XML (rss.go:55-66's Check calls the same load helper reads use), so a
// declarative JSON check_fixture replay cannot exercise it either.
func (h Hooks) Check(ctx context.Context, cfg connectors.RuntimeConfig, rt *engine.Runtime) (bool, error) {
	if err := ctx.Err(); err != nil {
		return true, err
	}
	if _, err := load(ctx, rt.Requester); err != nil {
		return true, fmt.Errorf("check rss: %w", err)
	}
	return true, nil
}

// load fetches the feed via r (the engine-built Requester, whose BaseURL is
// already resolved from streams.json's base.url = {{ config.feed_url }})
// and decodes it as RSS/XML, mirroring legacy rss.go's load.
func load(ctx context.Context, r *connsdk.Requester) (rssDocument, error) {
	resp, err := r.Do(ctx, http.MethodGet, "", nil, nil)
	if err != nil {
		return rssDocument{}, fmt.Errorf("read rss feed: %w", err)
	}
	var doc rssDocument
	dec := xml.NewDecoder(bytes.NewReader(resp.Body))
	if err := dec.Decode(&doc); err != nil {
		return rssDocument{}, fmt.Errorf("decode rss feed: %w", err)
	}
	if strings.TrimSpace(doc.Channel.Title) == "" && len(doc.Channel.Items) == 0 {
		return rssDocument{}, errors.New("rss feed missing channel data")
	}
	return doc, nil
}

// channelRecord/itemRecord port legacy rss.go's identically named functions
// verbatim, including their id fallback chains.
func channelRecord(channel rssChannel) connectors.Record {
	id := strings.TrimSpace(channel.Link)
	if id == "" {
		id = strings.TrimSpace(channel.Title)
	}
	return connectors.Record{"id": id, "title": channel.Title, "link": channel.Link, "description": channel.Description, "updated_at": channel.Updated}
}

func itemRecord(item rssItem) connectors.Record {
	id := strings.TrimSpace(item.GUID)
	if id == "" {
		id = strings.TrimSpace(item.Link)
	}
	if id == "" {
		id = strings.TrimSpace(item.Title)
	}
	return connectors.Record{"id": id, "title": item.Title, "link": item.Link, "description": item.Description, "published_at": item.Published}
}
