package googlepagespeedinsights

import "polymetrics.ai/internal/connectors"

// streamName is the single synthetic stream exposed by the connector. The
// PageSpeed Insights API has no list endpoints; one runPagespeed request yields
// one report per (url, strategy) pair, so the connector folds the cartesian
// product of the configured urls and strategies into a single report stream.
const streamName = "pagespeed_reports"

// pagespeedStreams returns the connector's published stream catalog. The primary
// key is the composite of the analyzed url and strategy, which uniquely
// identifies one report row in a sync. There is no incremental cursor field: the
// API only supports full refresh, but analysis_utc_timestamp is exposed so a
// downstream model can dedupe by recency.
func pagespeedStreams() []connectors.Stream {
	return []connectors.Stream{
		{
			Name:        streamName,
			Description: "Lighthouse PageSpeed Insights reports, one row per analyzed URL and strategy.",
			PrimaryKey:  []string{"url", "strategy"},
			Fields:      pagespeedFields(),
		},
	}
}

func pagespeedFields() []connectors.Field {
	return []connectors.Field{
		{Name: "url", Type: "string"},
		{Name: "strategy", Type: "string"},
		{Name: "id", Type: "string"},
		{Name: "kind", Type: "string"},
		{Name: "requested_url", Type: "string"},
		{Name: "final_url", Type: "string"},
		{Name: "lighthouse_version", Type: "string"},
		{Name: "fetch_time", Type: "string"},
		{Name: "analysis_utc_timestamp", Type: "string"},
		{Name: "overall_loading_experience", Type: "string"},
		{Name: "performance_score", Type: "number"},
		{Name: "accessibility_score", Type: "number"},
		{Name: "best_practices_score", Type: "number"},
		{Name: "seo_score", Type: "number"},
		{Name: "pwa_score", Type: "number"},
	}
}

// pagespeedRecord flattens a raw runPagespeed response object into a record,
// injecting the url and strategy that produced it (the API echoes neither the
// requested strategy nor a stable id we can rely on for the row identity). The
// per-category Lighthouse scores are pulled out of lighthouseResult.categories
// into flat <category>_score columns.
func pagespeedRecord(analyzedURL, strategy string, body map[string]any) connectors.Record {
	rec := connectors.Record{
		"url":      analyzedURL,
		"strategy": strategy,
		"id":       body["id"],
		"kind":     body["kind"],
	}
	if ts, ok := body["analysisUTCTimestamp"]; ok {
		rec["analysis_utc_timestamp"] = ts
	}

	if le, ok := body["loadingExperience"].(map[string]any); ok {
		rec["overall_loading_experience"] = le["overall_category"]
	}

	lhr, _ := body["lighthouseResult"].(map[string]any)
	if lhr != nil {
		rec["requested_url"] = lhr["requestedUrl"]
		rec["final_url"] = lhr["finalUrl"]
		rec["lighthouse_version"] = lhr["lighthouseVersion"]
		rec["fetch_time"] = lhr["fetchTime"]
		if cats, ok := lhr["categories"].(map[string]any); ok {
			rec["performance_score"] = categoryScore(cats, "performance")
			rec["accessibility_score"] = categoryScore(cats, "accessibility")
			rec["best_practices_score"] = categoryScore(cats, "best-practices")
			rec["seo_score"] = categoryScore(cats, "seo")
			rec["pwa_score"] = categoryScore(cats, "pwa")
		}
	}
	return rec
}

// categoryScore returns the Lighthouse score for a category id, or nil when the
// category was not requested/run. The score is left as the raw decoded value
// (json.Number) so integer/float fidelity is preserved downstream.
func categoryScore(categories map[string]any, id string) any {
	cat, ok := categories[id].(map[string]any)
	if !ok {
		return nil
	}
	return cat["score"]
}
