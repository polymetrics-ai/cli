package cli

import (
	"bytes"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

// A direct read's page notice is the human-readable half of the completeness
// claim, so it may not say anything the page did not measure.
func TestWriteDirectReadPageNotice(t *testing.T) {
	tests := []struct {
		name   string
		page   connectors.DirectReadPage
		want   string
		silent bool
	}{
		{
			name:   "unset page from a non-engine connector says nothing",
			page:   connectors.DirectReadPage{},
			silent: true,
		},
		{
			name:   "complete page says nothing",
			page:   connectors.DirectReadPage{Strategy: "page_number", Records: 7, Complete: true},
			silent: true,
		},
		{
			name: "more pages point at the next page number",
			page: connectors.DirectReadPage{Strategy: "page_number", Records: 100, Number: 1, NextNumber: 2, HasMore: true, Reason: connectors.DirectReadPageReasonMorePages},
			want: "--page 2",
		},
		{
			name: "no declaration says so",
			page: connectors.DirectReadPage{Strategy: "none", Records: 30, Reason: connectors.DirectReadPageReasonNoPagination},
			want: "declares no pagination strategy",
		},
		{
			name: "an explicit none is not reported as no declaration",
			page: connectors.DirectReadPage{Strategy: "none", Records: 30, Reason: connectors.DirectReadPageReasonDeclaredNone},
			want: `declares pagination type "none"`,
		},
		{
			name: "a strategy that cannot page this request says which",
			page: connectors.DirectReadPage{Strategy: "cursor", Records: 30, Reason: connectors.DirectReadPageReasonNotAddressable},
			want: `the declared "cursor" pagination cannot page this request`,
		},
		{
			name: "an unusable declaration names the strategy",
			page: connectors.DirectReadPage{Strategy: "cursor", Records: 30, Reason: connectors.DirectReadPageReasonInvalidSpec},
			want: `declared "cursor" pagination is unusable`,
		},
		{
			name: "an ambiguous envelope reports the rows it did receive",
			page: connectors.DirectReadPage{Strategy: "offset_limit", Records: 4, Number: 1, Reason: connectors.DirectReadPageReasonAmbiguous},
			want: "4 array elements returned",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			writeDirectReadPageNotice(&stderr, tt.page)
			got := stderr.String()
			if tt.silent {
				if got != "" {
					t.Fatalf("notice = %q, want nothing written", got)
				}
				return
			}
			if !strings.Contains(got, tt.want) {
				t.Fatalf("notice = %q, want it to contain %q", got, tt.want)
			}
		})
	}
}

// The empty parenthetical was the visible half of the same defect: a read that
// measured nothing was rendered as one that measured incompleteness.
func TestWriteDirectReadPageNoticeNeverPrintsAnEmptyReason(t *testing.T) {
	for _, page := range []connectors.DirectReadPage{
		{},
		{Strategy: "page_number", Records: 3},
	} {
		var stderr bytes.Buffer
		writeDirectReadPageNotice(&stderr, page)
		if strings.Contains(stderr.String(), "()") {
			t.Fatalf("notice = %q, want no empty parenthetical", stderr.String())
		}
	}
}
