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
		{
			name: "a page size that never reached the wire says completeness is unproved",
			page: connectors.DirectReadPage{Strategy: "page_number", Records: 30, Number: 1, Reason: connectors.DirectReadPageReasonSizeNotRequested},
			want: "names no page-size parameter",
		},
		{
			name: "more pages with no engine-addressable next points at the caller's own parameter",
			page: connectors.DirectReadPage{Strategy: "offset_limit", Records: 50, HasMore: true, Reason: connectors.DirectReadPageReasonMorePages},
			want: "advance the paging parameter you supplied",
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

// connectorCommandPage is the CLI half of "refused, never answered quietly".
// --page 0 names no page, and treating it as unset both returned page one and
// let the mutual-exclusion check below be walked past.
func TestConnectorCommandPageRefusesAPageNumberThatNamesNoPage(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  string
		wantPage int
		wantCurs string
	}{
		{name: "absent page is unset", args: []string{"list"}},
		{name: "page zero is refused", args: []string{"list", "--page", "0"}, wantErr: "want a positive page number"},
		{name: "negative page is refused", args: []string{"list", "--page", "-1"}, wantErr: "want a positive page number"},
		{name: "page zero cannot smuggle a cursor past the exclusion check", args: []string{"list", "--page", "0", "--page-cursor", "abc"}, wantErr: "invalid --page 0"},
		{name: "page and cursor together are refused", args: []string{"list", "--page", "2", "--page-cursor", "abc"}, wantErr: "mutually exclusive"},
		{name: "a positive page is accepted", args: []string{"list", "--page", "3"}, wantPage: 3},
		{name: "a cursor alone is accepted", args: []string{"list", "--page-cursor", "abc"}, wantCurs: "abc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, cursor, err := connectorCommandPage(parseFlags(tt.args))
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("connectorCommandPage(%v) = (%d, %q, nil), want a refusal", tt.args, page, cursor)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %q, want it to contain %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("connectorCommandPage(%v): %v", tt.args, err)
			}
			if page != tt.wantPage || cursor != tt.wantCurs {
				t.Fatalf("connectorCommandPage(%v) = (%d, %q), want (%d, %q)", tt.args, page, cursor, tt.wantPage, tt.wantCurs)
			}
		})
	}
}
