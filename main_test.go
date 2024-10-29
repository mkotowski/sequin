package main

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		sequence string
		want     []string
	}{
		{
			name:     "Request Kitty Keyboard",
			sequence: "\x1b[?u",
			want:     []string{"ESC [?u: Request Kitty Keyboard"},
		},
		{
			name:     "Kitty Keyboard",
			sequence: "\x1b[=1;2u",
			want:     []string{"ESC [=1;2u: Kitty Keyboard: flags=1 (Disambiguate escape codes), mode=2 (Set given flags and keep existing flags unchanged)"},
		},
		{
			name:     "Push Kitty Keyboard",
			sequence: "\x1b[>3u",
			want:     []string{"ESC [>3u: Push Kitty Keyboard: flags=3"},
		},
		{
			name:     "Pop Kitty Keyboard",
			sequence: "\x1b[<2u",
			want:     []string{"ESC [<2u: Pop Kitty Keyboard: n=2"},
		},
		{
			name:     "Cursor down",
			sequence: "\x1b[3B",
			want:     []string{"CSI 3B: Cursor down 3 lines"},
		},
		{
			name:     "Cursor up",
			sequence: "\x1b[2A",
			want:     []string{"CSI 2A: Cursor up 2 lines"},
		},
		{
			name:     "Cursor right",
			sequence: "\x1b[4C",
			want:     []string{"CSI 4C: Cursor right 4 columns"},
		},
		{
			name:     "Cursor left",
			sequence: "\x1b[1D",
			want:     []string{"CSI 1D: Cursor left 1 column"},
		},
		{
			name:     "Save cursor position",
			sequence: "\x1b7",
			want:     []string{"ESC 7: Save cursor position"},
		},
		{
			name:     "Restore cursor position",
			sequence: "\x1b8",
			want:     []string{"ESC 8: Restore cursor position"},
		},
		{
			name:     "Request cursor position",
			sequence: "\x1b[6n",
			want:     []string{"CSI 6n: Request cursor position (CPR)"},
		},
		{
			name:     "Request extended cursor position",
			sequence: "\x1b[?6n",
			want:     []string{"CSI ?6n: Request extended cursor position"},
		},
		{
			name:     "Move cursor to upper left corner",
			sequence: "\x1b[1;1H",
			want:     []string{"CSI 1;1H: Move cursor to upper left corner (origin)"},
		},
		{
			name:     "Save cursor position (CSI s)",
			sequence: "\x1b[s",
			want:     []string{"CSI s: Save cursor position"},
		},
		{
			name:     "Restore cursor position (CSI u)",
			sequence: "\x1b[u",
			want:     []string{"CSI u: Restore cursor position"},
		},
		{
			name:     "Set cursor style",
			sequence: "\x1b[3 q",
			want:     []string{"CSI 3 q: Set cursor style: Blinking underline"},
		},
		{
			name:     "Set pointer shape",
			sequence: "\x1b]22;beam\x07",
			want:     []string{"OSC 22;beam BEL: Set pointer shape: beam"},
		},
		{
			name:     "Clipboard operation (set)",
			sequence: "\x1b]52;c;SGVsbG8sIFdvcmxkIQ==\x07",
			want:     []string{"OSC 52;c;SGVsbG8sIFdvcmxkIQ== BEL: Set system clipboard: Hello, World!"},
		},
		{
			name:     "Erase display",
			sequence: "\x1b[2J",
			want:     []string{"CSI 2J: Erase display: Clear entire screen"},
		},
		{
			name:     "Erase line",
			sequence: "\x1b[K",
			want:     []string{"CSI K: Erase line: Clear from cursor to end of line"},
		},
		{
			name:     "Scroll up",
			sequence: "\x1b[2S",
			want:     []string{"CSI 2S: Scroll up: 2 lines"},
		},
		{
			name:     "Scroll down",
			sequence: "\x1b[3T",
			want:     []string{"CSI 3T: Scroll down: 3 lines"},
		},
		{
			name:     "Insert line",
			sequence: "\x1b[L",
			want:     []string{"CSI L: Insert 1 blank line(s)"},
		},
		{
			name:     "Delete line",
			sequence: "\x1b[2M",
			want:     []string{"CSI 2M: Delete 2 line(s)"},
		},
		{
			name:     "Set scrolling region",
			sequence: "\x1b[1;24r",
			want:     []string{"CSI 1;24r: Set scrolling region: top=1, bottom=24"},
		},
		{
			name:     "Hyperlink",
			sequence: "\x1b]8;;https://example.com\x07",
			want:     []string{"OSC 8;;https://example.com BEL: Set hyperlink: URI=https://example.com, attributes="},
		},
		{
			name:     "Set foreground color",
			sequence: "\x1b]10;#FF0000\x07",
			want:     []string{"OSC 10;#FF0000 BEL: Set foreground color: #FF0000"},
		},
		{
			name:     "Reset foreground color",
			sequence: "\x1b]110\x07",
			want:     []string{"OSC 110 BEL: Reset foreground color"},
		},
		{
			name:     "Set background color",
			sequence: "\x1b]11;#00FF00\x07",
			want:     []string{"OSC 11;#00FF00 BEL: Set background color: #00FF00"},
		},
		{
			name:     "Reset background color",
			sequence: "\x1b]111\x07",
			want:     []string{"OSC 111 BEL: Reset background color"},
		},
		{
			name:     "Set cursor color",
			sequence: "\x1b]12;#0000FF\x07",
			want:     []string{"OSC 12;#0000FF BEL: Set cursor color: #0000FF"},
		},
		{
			name:     "Reset cursor color",
			sequence: "\x1b]112\x07",
			want:     []string{"OSC 112 BEL: Reset cursor color"},
		},
		{
			name:     "Set icon name and window title",
			sequence: "\x1b]0;My Window\x07",
			want:     []string{"OSC 0;My Window BEL: Set icon name and window title: My Window"},
		},
		{
			name:     "Set icon name",
			sequence: "\x1b]1;My Icon\x07",
			want:     []string{"OSC 1;My Icon BEL: Set icon name: My Icon"},
		},
		{
			name:     "Set window title",
			sequence: "\x1b]2;My Title\x07",
			want:     []string{"OSC 2;My Title BEL: Set window title: My Title"},
		},
		{
			name:     "Enable Cursor Keys",
			sequence: "\x1b[?1h",
			want:     []string{"CSI ?1h: Enable Cursor Keys"},
		},
		{
			name:     "Disable Cursor Keys",
			sequence: "\x1b[?1l",
			want:     []string{"CSI ?1l: Disable Cursor Keys"},
		},
		{
			name:     "Show Cursor",
			sequence: "\x1b[?25h",
			want:     []string{"CSI ?25h: Show Cursor"},
		},
		{
			name:     "Hide Cursor",
			sequence: "\x1b[?25l",
			want:     []string{"CSI ?25l: Hide Cursor"},
		},
		{
			name:     "Enable Mouse",
			sequence: "\x1b[?1000h",
			want:     []string{"CSI ?1000h: Enable Mouse"},
		},
		{
			name:     "Disable Mouse",
			sequence: "\x1b[?1000l",
			want:     []string{"CSI ?1000l: Disable Mouse"},
		},
		{
			name:     "Enable Alt Screen Buffer",
			sequence: "\x1b[?1049h",
			want:     []string{"CSI ?1049h: Enable Alt Screen Buffer"},
		},
		{
			name:     "Disable Alt Screen Buffer",
			sequence: "\x1b[?1049l",
			want:     []string{"CSI ?1049l: Disable Alt Screen Buffer"},
		},
		{
			name:     "Enable Bracketed Paste",
			sequence: "\x1b[?2004h",
			want:     []string{"CSI ?2004h: Enable Bracketed Paste"},
		},
		{
			name:     "Disable Bracketed Paste",
			sequence: "\x1b[?2004l",
			want:     []string{"CSI ?2004l: Disable Bracketed Paste"},
		},
		{
			name:     "Reset Style",
			sequence: "\x1b[m",
			want:     []string{"CSI m: Reset Style"},
		},
		{
			name:     "Set multiple SGR attributes",
			sequence: "\x1b[1;4;31m",
			want: []string{
				"CSI 1m: Set bold",
				"CSI 4m: Set underline",
				"CSI 31m: Set foreground color to Red",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parse(tt.sequence); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
