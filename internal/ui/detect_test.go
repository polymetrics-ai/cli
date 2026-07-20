package ui

import "testing"

func TestDetectModeUsesADRGate(t *testing.T) {
	tests := []struct {
		name string
		opts DetectOptions
		want Mode
	}{
		{
			name: "dual tty default enables tui",
			opts: DetectOptions{StdinTTY: true, StdoutTTY: true, Env: map[string]string{"TERM": "xterm-256color"}},
			want: ModeTUI,
		},
		{
			name: "stdin pipe forces plain",
			opts: DetectOptions{StdinTTY: false, StdoutTTY: true, Env: map[string]string{"TERM": "xterm-256color"}},
			want: ModePlain,
		},
		{
			name: "stdout pipe forces plain",
			opts: DetectOptions{StdinTTY: true, StdoutTTY: false, Env: map[string]string{"TERM": "xterm-256color"}},
			want: ModePlain,
		},
		{
			name: "json forces plain",
			opts: DetectOptions{StdinTTY: true, StdoutTTY: true, JSON: true, Env: map[string]string{"TERM": "xterm-256color"}},
			want: ModePlain,
		},
		{
			name: "plain flag forces plain",
			opts: DetectOptions{StdinTTY: true, StdoutTTY: true, Plain: true, Env: map[string]string{"TERM": "xterm-256color"}},
			want: ModePlain,
		},
		{
			name: "no input forces plain",
			opts: DetectOptions{StdinTTY: true, StdoutTTY: true, NoInput: true, Env: map[string]string{"TERM": "xterm-256color"}},
			want: ModePlain,
		},
		{
			name: "pm no tui forces plain",
			opts: DetectOptions{StdinTTY: true, StdoutTTY: true, Env: map[string]string{"TERM": "xterm-256color", "PM_NO_TUI": "1"}},
			want: ModePlain,
		},
		{
			name: "pm no tui zero still forces plain",
			opts: DetectOptions{StdinTTY: true, StdoutTTY: true, Env: map[string]string{"TERM": "xterm-256color", "PM_NO_TUI": "0"}},
			want: ModePlain,
		},
		{
			name: "pm no tui false still forces plain",
			opts: DetectOptions{StdinTTY: true, StdoutTTY: true, Env: map[string]string{"TERM": "xterm-256color", "PM_NO_TUI": "false"}},
			want: ModePlain,
		},
		{
			name: "ci forces plain",
			opts: DetectOptions{StdinTTY: true, StdoutTTY: true, Env: map[string]string{"TERM": "xterm-256color", "CI": "true"}},
			want: ModePlain,
		},
		{
			name: "ci zero still forces plain",
			opts: DetectOptions{StdinTTY: true, StdoutTTY: true, Env: map[string]string{"TERM": "xterm-256color", "CI": "0"}},
			want: ModePlain,
		},
		{
			name: "ci false still forces plain",
			opts: DetectOptions{StdinTTY: true, StdoutTTY: true, Env: map[string]string{"TERM": "xterm-256color", "CI": "false"}},
			want: ModePlain,
		},
		{
			name: "dumb term forces plain",
			opts: DetectOptions{StdinTTY: true, StdoutTTY: true, Env: map[string]string{"TERM": "dumb"}},
			want: ModePlain,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Detect(tt.opts)
			if got.Mode != tt.want {
				t.Fatalf("Detect(%+v).Mode = %q, want %q (reasons=%v)", tt.opts, got.Mode, tt.want, got.Reasons)
			}
		})
	}
}

func TestDetectCapabilitiesDegradeForColorAndASCII(t *testing.T) {
	tests := []struct {
		name      string
		opts      DetectOptions
		wantColor bool
		wantASCII bool
	}{
		{
			name:      "dual tty color unicode default",
			opts:      DetectOptions{StdinTTY: true, StdoutTTY: true, Env: map[string]string{"TERM": "xterm-256color"}},
			wantColor: true,
			wantASCII: false,
		},
		{
			name:      "no color disables color",
			opts:      DetectOptions{StdinTTY: true, StdoutTTY: true, Env: map[string]string{"TERM": "xterm-256color", "NO_COLOR": "1"}},
			wantColor: false,
			wantASCII: false,
		},
		{
			name:      "no color zero disables color",
			opts:      DetectOptions{StdinTTY: true, StdoutTTY: true, Env: map[string]string{"TERM": "xterm-256color", "NO_COLOR": "0"}},
			wantColor: false,
			wantASCII: false,
		},
		{
			name:      "no color false disables color",
			opts:      DetectOptions{StdinTTY: true, StdoutTTY: true, Env: map[string]string{"TERM": "xterm-256color", "NO_COLOR": "false"}},
			wantColor: false,
			wantASCII: false,
		},
		{
			name:      "clicolor zero disables color",
			opts:      DetectOptions{StdinTTY: true, StdoutTTY: true, Env: map[string]string{"TERM": "xterm-256color", "CLICOLOR": "0"}},
			wantColor: false,
			wantASCII: false,
		},
		{
			name:      "pipe disables color and uses ascii",
			opts:      DetectOptions{StdoutTTY: false, Env: map[string]string{"TERM": "xterm-256color"}},
			wantColor: false,
			wantASCII: true,
		},
		{
			name:      "pm ascii requests ascii",
			opts:      DetectOptions{StdinTTY: true, StdoutTTY: true, Env: map[string]string{"TERM": "xterm-256color", "PM_ASCII": "1"}},
			wantColor: true,
			wantASCII: true,
		},
		{
			name:      "dumb term disables color and uses ascii",
			opts:      DetectOptions{StdinTTY: true, StdoutTTY: true, Env: map[string]string{"TERM": "dumb"}},
			wantColor: false,
			wantASCII: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Detect(tt.opts)
			if got.Color != tt.wantColor || got.ASCII != tt.wantASCII {
				t.Fatalf("Detect(%+v) Color/ASCII = %t/%t, want %t/%t", tt.opts, got.Color, got.ASCII, tt.wantColor, tt.wantASCII)
			}
		})
	}
}
