package cli

import "testing"

func TestParseMaxInFlightBatchesRequiresExecutableBound(t *testing.T) {
	for _, tt := range []struct {
		raw     string
		want    int
		wantErr bool
	}{
		{raw: "", want: 0},
		{raw: "1", want: 1},
		{raw: "8", want: 8},
		{raw: "0", wantErr: true},
		{raw: "9", wantErr: true},
		{raw: "workers", wantErr: true},
	} {
		got, err := parseMaxInFlightBatches(tt.raw)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("parseMaxInFlightBatches(%q) = (%d, nil), want error", tt.raw, got)
			}
			continue
		}
		if err != nil || got != tt.want {
			t.Fatalf("parseMaxInFlightBatches(%q) = (%d, %v), want (%d, nil)", tt.raw, got, err, tt.want)
		}
	}
}

func TestParseTargetCopyWorkersRequiresGlobalBound(t *testing.T) {
	for _, tt := range []struct {
		raw     string
		want    int
		wantErr bool
	}{
		{raw: "", want: 0},
		{raw: "1", want: 1},
		{raw: "8", want: 8},
		{raw: "0", wantErr: true},
		{raw: "9", wantErr: true},
	} {
		got, err := parseTargetCopyWorkers(tt.raw)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("parseTargetCopyWorkers(%q) = (%d, nil), want error", tt.raw, got)
			}
			continue
		}
		if err != nil || got != tt.want {
			t.Fatalf("parseTargetCopyWorkers(%q) = (%d, %v), want (%d, nil)", tt.raw, got, err, tt.want)
		}
	}
}
