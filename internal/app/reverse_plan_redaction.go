package app

import (
	"strings"

	"polymetrics.ai/internal/connectors"
)

const reversePlanRedactedValue = "***"

var whatsappReversePlanRedactFields = map[string][]string{
	"send_text_message":        {"to", "text"},
	"send_image_message":       {"to", "image"},
	"send_audio_message":       {"to", "audio"},
	"send_video_message":       {"to", "video"},
	"send_document_message":    {"to", "document"},
	"send_sticker_message":     {"to", "sticker"},
	"send_location_message":    {"to", "location"},
	"send_contacts_message":    {"to", "contacts"},
	"send_interactive_message": {"to", "interactive"},
	"send_template_message":    {"to", "template"},
	"send_reaction_message":    {"to", "reaction"},
	"mark_message_read":        {"message_id"},
	"send_typing_indicator":    {"message_id", "typing_indicator"},
	"upload_media":             {"media_file"},
	"create_message_template":  {"components"},
	"update_message_template":  {"components"},
	"create_qr_code":           {"prefilled_message"},
}

func redactedReversePlan(plan ReversePlan) ReversePlan {
	plan.Sample = redactedReversePlanSample(plan.DestinationConnector, plan.Action, plan.Sample)
	return plan
}

func redactedReversePlanSample(connectorName, actionName string, records []connectors.Record) []connectors.Record {
	out := cloneRecords(records)
	fields := reversePlanSampleRedactFields(connectorName, actionName)
	if len(fields) == 0 {
		return out
	}
	for _, rec := range out {
		for _, field := range fields {
			if _, ok := rec[field]; ok {
				rec[field] = reversePlanRedactedValue
			}
		}
	}
	return out
}

func reversePlanSampleRedactFields(connectorName, actionName string) []string {
	if !strings.EqualFold(strings.TrimSpace(connectorName), "whatsapp") {
		return nil
	}
	return whatsappReversePlanRedactFields[strings.ToLower(strings.TrimSpace(actionName))]
}
