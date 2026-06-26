package rss

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
)

const (
	connectorName  = "rss"
	defaultFeedURL = "https://xkcd.com/rss.xml"
	userAgent      = "polymetrics-go-cli"
)

func init() { connectors.RegisterFactory(connectorName, New) }

func New() connectors.Connector { return Connector{} }

type Connector struct{ Client *http.Client }

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

func (Connector) Name() string { return connectorName }

func (Connector) Metadata() connectors.Metadata {
	return connectors.Metadata{Name: connectorName, DisplayName: "RSS", IntegrationType: "api", Description: "Reads RSS channel metadata and feed items. Read-only and credential-free.", Capabilities: connectors.Capabilities{Check: true, Catalog: true, Read: true, Write: false}}
}

func (c Connector) Check(ctx context.Context, cfg connectors.RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if fixtureMode(cfg) {
		return nil
	}
	if _, err := c.load(ctx, cfg); err != nil {
		return fmt.Errorf("check rss: %w", err)
	}
	return nil
}

func (Connector) Catalog(ctx context.Context, cfg connectors.RuntimeConfig) (connectors.Catalog, error) {
	if err := ctx.Err(); err != nil {
		return connectors.Catalog{}, err
	}
	return connectors.Catalog{Connector: connectorName, Streams: []connectors.Stream{
		{Name: "items", Description: "RSS feed items.", PrimaryKey: []string{"id"}, CursorFields: []string{"published_at"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "title", Type: "string"}, {Name: "link", Type: "string"}, {Name: "published_at", Type: "string"}}},
		{Name: "channel", Description: "RSS channel metadata.", PrimaryKey: []string{"id"}, Fields: []connectors.Field{{Name: "id", Type: "string"}, {Name: "title", Type: "string"}, {Name: "link", Type: "string"}, {Name: "description", Type: "string"}}},
	}}, nil
}

func (c Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	stream := req.Stream
	if stream == "" {
		stream = "items"
	}
	if stream != "items" && stream != "channel" {
		return fmt.Errorf("rss stream %q not found", stream)
	}
	if fixtureMode(req.Config) {
		return readFixture(ctx, stream, emit)
	}
	doc, err := c.load(ctx, req.Config)
	if err != nil {
		return err
	}
	if stream == "channel" {
		return emit(channelRecord(doc.Channel))
	}
	for _, item := range doc.Channel.Items {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(itemRecord(item)); err != nil {
			return err
		}
	}
	return nil
}

func (Connector) Write(ctx context.Context, req connectors.WriteRequest, records []connectors.Record) (connectors.WriteResult, error) {
	return connectors.WriteResult{}, connectors.ErrUnsupportedOperation
}

func (c Connector) load(ctx context.Context, cfg connectors.RuntimeConfig) (rssDocument, error) {
	feed, err := feedURL(cfg)
	if err != nil {
		return rssDocument{}, err
	}
	r := connsdk.Requester{Client: c.Client, UserAgent: userAgent, Accept: "application/rss+xml, application/xml, text/xml, */*"}
	resp, err := r.Do(ctx, http.MethodGet, feed, nil, nil)
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

func readFixture(ctx context.Context, stream string, emit func(connectors.Record) error) error {
	if stream == "channel" {
		return emit(connectors.Record{"id": "fixture-feed", "title": "Fixture Feed", "link": "https://example.test/feed", "fixture": true})
	}
	for i := 1; i <= 2; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(connectors.Record{"id": fmt.Sprintf("fixture-item-%d", i), "title": fmt.Sprintf("Fixture item %d", i), "link": fmt.Sprintf("https://example.test/%d", i), "fixture": true}); err != nil {
			return err
		}
	}
	return nil
}

func feedURL(cfg connectors.RuntimeConfig) (string, error) {
	feed := strings.TrimSpace(cfg.Config["feed_url"])
	if feed == "" {
		feed = strings.TrimSpace(cfg.Config["base_url"])
	}
	if feed == "" {
		feed = defaultFeedURL
	}
	parsed, err := url.Parse(feed)
	if err != nil {
		return "", fmt.Errorf("rss config feed_url is invalid: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", errors.New("rss config feed_url must use http or https")
	}
	if parsed.Host == "" {
		return "", errors.New("rss config feed_url must include a host")
	}
	return feed, nil
}

func fixtureMode(cfg connectors.RuntimeConfig) bool {
	return cfg.Config != nil && strings.EqualFold(strings.TrimSpace(cfg.Config["mode"]), "fixture")
}
