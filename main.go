package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

func main() {
	s := strings.Join(os.Args[1:], " ")
	if s == "-" || s == "" {
		bts, err := io.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}
		s = string(bts)
	}

	// Split the input string into individual escape sequences
	sequences := strings.Split(s, "\x1b")

	for _, seq := range sequences {
		if seq == "" {
			continue
		}

		// Add the escape character back to the sequence
		seq = "\x1b" + seq

		for _, desc := range parse(seq) {
			fmt.Println(desc)
		}

		if txt := ansi.Strip(seq); strings.TrimSpace(txt) != "" {
			fmt.Println(txt)
		}

	}
}

func parse(seq string) []string {
	var r []string

	// Request Kitty Keyboard
	if is(seq, `\x1b\[?u`) {
		r = append(r, "ESC [?u: Request Kitty Keyboard")
	}

	// Kitty Keyboard
	if ok, params := extract(seq, `\x1b\[=(\d+;\d+)u`); ok {
		parts := strings.Split(params, ";")
		if len(parts) == 2 {
			flags, mode := parts[0], parts[1]
			r = append(r, fmt.Sprintf("ESC [=%s;%su: Kitty Keyboard: flags=%s (%s), mode=%s (%s)",
				flags, mode, flags, describeKittyFlags(flags),
				mode, describeKittyMode(mode)))
		}
	}

	// Push Kitty Keyboard
	if ok, flags := extract(seq, `\x1b\[>(\d*)u`); ok {
		if flags == "" {
			r = append(r, "ESC [>u: Push Kitty Keyboard: flags=0 (Disable all features)")
		} else {
			r = append(r, fmt.Sprintf("ESC [>%su: Push Kitty Keyboard: flags=%s", flags, flags))
		}
	}

	// Pop Kitty Keyboard
	if ok, n := extract(seq, `\x1b\[<(\d*)u`); ok {
		if n == "" {
			r = append(r, "ESC [<u: Pop Kitty Keyboard: n=1")
		} else {
			r = append(r, fmt.Sprintf("ESC [<%su: Pop Kitty Keyboard: n=%s", n, n))
		}
	}

	// cursor down
	if ok, no := extract(seq, `\x1b\[(\d+)B`); ok {
		r = append(r, fmt.Sprintf("CSI %sB: Cursor down %s lines", no, no))
	} else if is(seq, `\x1b\[B`) {
		r = append(r, "CSI B: Cursor down 1 line")
	}

	// cursor up
	if ok, no := extract(seq, `\x1b\[(\d+)A`); ok {
		r = append(r, fmt.Sprintf("CSI %sA: Cursor up %s lines", no, no))
	} else if is(seq, `\x1b\[A`) {
		r = append(r, "CSI A: Cursor up 1 line")
	}

	// cursor right
	if ok, no := extract(seq, `\x1b\[(\d+)C`); ok {
		r = append(r, fmt.Sprintf("CSI %sC: Cursor right %s columns", no, no))
	} else if is(seq, `\x1b\[C`) {
		r = append(r, "CSI C: Cursor right 1 column")
	}

	// cursor left
	if ok, no := extract(seq, `\x1b\[(\d+)D`); ok {
		r = append(r, fmt.Sprintf("CSI %sD: Cursor left %s columns", no, no))
	} else if is(seq, `\x1b\[D`) {
		r = append(r, "CSI D: Cursor left 1 column")
	}

	// save cursor position
	if is(seq, `\x1b7`) {
		r = append(r, "ESC 7: Save cursor position")
	}

	// restore cursor position
	if is(seq, `\x1b8`) {
		r = append(r, "ESC 8: Restore cursor position")
	}

	// request cursor position
	if is(seq, `\x1b\[6n`) {
		r = append(r, "CSI 6n: Request cursor position (CPR)")
	}

	// request extended cursor position
	if is(seq, `\x1b\[?6n`) {
		r = append(r, "CSI ?6n: Request extended cursor position")
	}

	// move cursor to upper left corner (origin)
	if is(seq, `\x1b\[1;1H`) {
		r = append(r, "CSI 1;1H: Move cursor to upper left corner (origin)")
	}

	// save cursor position (CSI s)
	if is(seq, `\x1b\[s`) {
		r = append(r, "CSI s: Save cursor position")
	}

	// restore cursor position (CSI u)
	if is(seq, `\x1b\[u`) {
		r = append(r, "CSI u: Restore cursor position")
	}

	// set cursor style
	if ok, style := extract(seq, `\x1b\[(\d+) q`); ok {
		r = append(r, fmt.Sprintf("CSI %s q: Set cursor style: %s", style, cursorStyle(style)))
	}

	// set pointer shape
	if ok, shape := extract(seq, `\x1b\]22;(.+)\x07`); ok {
		r = append(r, fmt.Sprintf("OSC 22;%s BEL: Set pointer shape: %s", shape, shape))
	}

	// clipboard operations
	if ok, params := extract(seq, `\x1b\]52;([cp]);(.*)\x07`); ok {
		parts := strings.SplitN(params, ";", 2)
		if len(parts) == 2 {
			clipboardName := parts[0]
			data := parts[1]

			switch data {
			case "":
				r = append(r, fmt.Sprintf("OSC 52;%s; BEL: Reset %s clipboard", clipboardName, describeClipboard(clipboardName)))
			case "?":
				r = append(r, fmt.Sprintf("OSC 52;%s;? BEL: Request %s clipboard", clipboardName, describeClipboard(clipboardName)))
			default:
				decodedData, err := base64.StdEncoding.DecodeString(data)
				if err != nil {
					r = append(r, fmt.Sprintf("OSC 52;%s;%s BEL: Set %s clipboard: Invalid base64 data", clipboardName, data, describeClipboard(clipboardName)))
				} else {
					r = append(r, fmt.Sprintf("OSC 52;%s;%s BEL: Set %s clipboard: %s", clipboardName, data, describeClipboard(clipboardName), string(decodedData)))
				}
			}
		}
	}

	// erase display
	if ok, n := extract(seq, `\x1b\[(\d*)J`); ok {
		r = append(r, fmt.Sprintf("CSI %sJ: Erase display: %s", n, eraseDisplayDescription(n)))
	}

	// erase line
	if ok, n := extract(seq, `\x1b\[(\d*)K`); ok {
		r = append(r, fmt.Sprintf("CSI %sK: Erase line: %s", n, eraseLineDescription(n)))
	}

	// scroll up
	if ok, n := extract(seq, `\x1b\[(\d*)S`); ok {
		r = append(r, fmt.Sprintf("CSI %sS: Scroll up: %s lines", n, defaultOne(n)))
	}

	// scroll down
	if ok, n := extract(seq, `\x1b\[(\d*)T`); ok {
		r = append(r, fmt.Sprintf("CSI %sT: Scroll down: %s lines", n, defaultOne(n)))
	}

	// insert line
	if ok, n := extract(seq, `\x1b\[(\d*)L`); ok {
		r = append(r, fmt.Sprintf("CSI %sL: Insert %s blank line(s)", n, defaultOne(n)))
	}

	// delete line
	if ok, n := extract(seq, `\x1b\[(\d*)M`); ok {
		r = append(r, fmt.Sprintf("CSI %sM: Delete %s line(s)", n, defaultOne(n)))
	}

	// set scrolling region
	if ok, params := extract(seq, `\x1b\[(\d*;\d*)r`); ok {
		parts := strings.Split(params, ";")
		if len(parts) == 2 {
			r = append(r, fmt.Sprintf("CSI %s;%sr: Set scrolling region: top=%s, bottom=%s", parts[0], parts[1], parts[0], parts[1]))
		}
	}

	// hyperlink
	if ok, params := extract(seq, `\x1b\]8;(.*);(.*)\x07`); ok {
		parts := strings.SplitN(params, ";", 2)
		if len(parts) == 2 {
			attributes, uri := parts[0], parts[1]
			if uri == "" {
				r = append(r, fmt.Sprintf("OSC 8;%s; BEL: Reset hyperlink (attributes: %s)", attributes, attributes))
			} else {
				r = append(r, fmt.Sprintf("OSC 8;%s;%s BEL: Set hyperlink: URI=%s, attributes=%s", attributes, uri, uri, attributes))
			}
		}
	}

	// set foreground color
	if ok, color := extract(seq, `\x1b\]10;(.*)\x07`); ok {
		if color == "?" {
			r = append(r, "OSC 10;? BEL: Request foreground color")
		} else {
			r = append(r, fmt.Sprintf("OSC 10;%s BEL: Set foreground color: %s", color, color))
		}
	}

	// reset foreground color
	if is(seq, `\x1b\]110\x07`) {
		r = append(r, "OSC 110 BEL: Reset foreground color")
	}

	// set background color
	if ok, color := extract(seq, `\x1b\]11;(.*)\x07`); ok {
		if color == "?" {
			r = append(r, "OSC 11;? BEL: Request background color")
		} else {
			r = append(r, fmt.Sprintf("OSC 11;%s BEL: Set background color: %s", color, color))
		}
	}

	// reset background color
	if is(seq, `\x1b\]111\x07`) {
		r = append(r, "OSC 111 BEL: Reset background color")
	}

	// set cursor color
	if ok, color := extract(seq, `\x1b\]12;(.*)\x07`); ok {
		if color == "?" {
			r = append(r, "OSC 12;? BEL: Request cursor color")
		} else {
			r = append(r, fmt.Sprintf("OSC 12;%s BEL: Set cursor color: %s", color, color))
		}
	}

	// reset cursor color
	if is(seq, `\x1b\]112\x07`) {
		r = append(r, "OSC 112 BEL: Reset cursor color")
	}

	// set icon name and window title
	if ok, title := extract(seq, `\x1b\]0;(.*)\x07`); ok {
		r = append(r, fmt.Sprintf("OSC 0;%s BEL: Set icon name and window title: %s", title, title))
	}

	// set icon name
	if ok, name := extract(seq, `\x1b\]1;(.*)\x07`); ok {
		r = append(r, fmt.Sprintf("OSC 1;%s BEL: Set icon name: %s", name, name))
	}

	// set window title
	if ok, title := extract(seq, `\x1b\]2;(.*)\x07`); ok {
		r = append(r, fmt.Sprintf("OSC 2;%s BEL: Set window title: %s", title, title))
	}

	// Enable/Disable Cursor Keys
	if is(seq, `\x1b\[?1h`) {
		r = append(r, "CSI ?1h: Enable Cursor Keys")
	} else if is(seq, `\x1b\[?1l`) {
		r = append(r, "CSI ?1l: Disable Cursor Keys")
	} else if is(seq, `\x1b\[?1\$p`) {
		r = append(r, "CSI ?1$p: Request Cursor Keys")
	}

	// Show/Hide Cursor
	if is(seq, `\x1b\[?25h`) {
		r = append(r, "CSI ?25h: Show Cursor")
	} else if is(seq, `\x1b\[?25l`) {
		r = append(r, "CSI ?25l: Hide Cursor")
	} else if is(seq, `\x1b\[?25\$p`) {
		r = append(r, "CSI ?25$p: Request Cursor Visibility")
	}

	// Enable/Disable Mouse
	if is(seq, `\x1b\[?1000h`) {
		r = append(r, "CSI ?1000h: Enable Mouse")
	} else if is(seq, `\x1b\[?1000l`) {
		r = append(r, "CSI ?1000l: Disable Mouse")
	} else if is(seq, `\x1b\[?1000\$p`) {
		r = append(r, "CSI ?1000$p: Request Mouse")
	}

	// Enable/Disable Mouse Hilite
	if is(seq, `\x1b\[?1001h`) {
		r = append(r, "CSI ?1001h: Enable Mouse Hilite")
	} else if is(seq, `\x1b\[?1001l`) {
		r = append(r, "CSI ?1001l: Disable Mouse Hilite")
	} else if is(seq, `\x1b\[?1001\$p`) {
		r = append(r, "CSI ?1001$p: Request Mouse Hilite")
	}

	// Enable/Disable Mouse Cell Motion
	if is(seq, `\x1b\[?1002h`) {
		r = append(r, "CSI ?1002h: Enable Mouse Cell Motion")
	} else if is(seq, `\x1b\[?1002l`) {
		r = append(r, "CSI ?1002l: Disable Mouse Cell Motion")
	} else if is(seq, `\x1b\[?1002\$p`) {
		r = append(r, "CSI ?1002$p: Request Mouse Cell Motion")
	}

	// Enable/Disable Mouse All Motion
	if is(seq, `\x1b\[?1003h`) {
		r = append(r, "CSI ?1003h: Enable Mouse All Motion")
	} else if is(seq, `\x1b\[?1003l`) {
		r = append(r, "CSI ?1003l: Disable Mouse All Motion")
	} else if is(seq, `\x1b\[?1003\$p`) {
		r = append(r, "CSI ?1003$p: Request Mouse All Motion")
	}

	// Enable/Disable Report Focus
	if is(seq, `\x1b\[?1004h`) {
		r = append(r, "CSI ?1004h: Enable Report Focus")
	} else if is(seq, `\x1b\[?1004l`) {
		r = append(r, "CSI ?1004l: Disable Report Focus")
	} else if is(seq, `\x1b\[?1004\$p`) {
		r = append(r, "CSI ?1004$p: Request Report Focus")
	}

	// Enable/Disable Mouse SGR Ext
	if is(seq, `\x1b\[?1006h`) {
		r = append(r, "CSI ?1006h: Enable Mouse SGR Ext")
	} else if is(seq, `\x1b\[?1006l`) {
		r = append(r, "CSI ?1006l: Disable Mouse SGR Ext")
	} else if is(seq, `\x1b\[?1006\$p`) {
		r = append(r, "CSI ?1006$p: Request Mouse SGR Ext")
	}

	// Enable/Disable Alt Screen Buffer
	if is(seq, `\x1b\[?1049h`) {
		r = append(r, "CSI ?1049h: Enable Alt Screen Buffer")
	} else if is(seq, `\x1b\[?1049l`) {
		r = append(r, "CSI ?1049l: Disable Alt Screen Buffer")
	} else if is(seq, `\x1b\[?1049\$p`) {
		r = append(r, "CSI ?1049$p: Request Alt Screen Buffer")
	}

	// Enable/Disable Bracketed Paste
	if is(seq, `\x1b\[?2004h`) {
		r = append(r, "CSI ?2004h: Enable Bracketed Paste")
	} else if is(seq, `\x1b\[?2004l`) {
		r = append(r, "CSI ?2004l: Disable Bracketed Paste")
	} else if is(seq, `\x1b\[?2004\$p`) {
		r = append(r, "CSI ?2004$p: Request Bracketed Paste")
	}

	// Enable/Disable Syncd Output
	if is(seq, `\x1b\[?2026h`) {
		r = append(r, "CSI ?2026h: Enable Syncd Output")
	} else if is(seq, `\x1b\[?2026l`) {
		r = append(r, "CSI ?2026l: Disable Syncd Output")
	} else if is(seq, `\x1b\[?2026\$p`) {
		r = append(r, "CSI ?2026$p: Request Syncd Output")
	}

	// Enable/Disable Grapheme Clustering
	if is(seq, `\x1b\[?2027h`) {
		r = append(r, "CSI ?2027h: Enable Grapheme Clustering")
	} else if is(seq, `\x1b\[?2027l`) {
		r = append(r, "CSI ?2027l: Disable Grapheme Clustering")
	} else if is(seq, `\x1b\[?2027\$p`) {
		r = append(r, "CSI ?2027$p: Request Grapheme Clustering")
	}

	// Enable/Disable Win32 Input
	if is(seq, `\x1b\[?9001h`) {
		r = append(r, "CSI ?9001h: Enable Win32 Input")
	} else if is(seq, `\x1b\[?9001l`) {
		r = append(r, "CSI ?9001l: Disable Win32 Input")
	} else if is(seq, `\x1b\[?9001\$p`) {
		r = append(r, "CSI ?9001$p: Request Win32 Input")
	}

	// Reset Style
	if is(seq, `\x1b\[m`) {
		r = append(r, "CSI m: Reset Style")
	}

	// Parse SGR (Select Graphic Rendition) sequences
	if ok, params := extract(seq, `\x1b\[(\d*(?:;\d*)*)m`); ok {
		r = append(r, parseSGR(params)...)
	}

	return r
}

func eraseDisplayDescription(n string) string {
	switch n {
	case "", "0":
		return "Clear from cursor to end of screen"
	case "1":
		return "Clear from cursor to beginning of the screen"
	case "2":
		return "Clear entire screen"
	case "3":
		return "Clear entire display including scrollback buffer"
	default:
		return "Unknown erase display command"
	}
}

func eraseLineDescription(n string) string {
	switch n {
	case "", "0":
		return "Clear from cursor to end of line"
	case "1":
		return "Clear from cursor to beginning of the line"
	case "2":
		return "Clear entire line"
	default:
		return "Unknown erase line command"
	}
}

func defaultOne(n string) string {
	if n == "" {
		return "1"
	}
	return n
}

func cursorStyle(style string) string {
	switch style {
	case "0", "1":
		return "Blinking block"
	case "2":
		return "Steady block"
	case "3":
		return "Blinking underline"
	case "4":
		return "Steady underline"
	case "5":
		return "Blinking bar"
	case "6":
		return "Steady bar"
	default:
		return "Unknown style"
	}
}

func extract(s, pattern string) (bool, string) {
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(s)
	if len(match) < 1 {
		return false, ""
	}
	return true, match[1]
}

func is(s, pattern string) bool {
	match, err := regexp.MatchString(pattern, s)
	if err != nil {
		panic(err)
	}
	return match
}

func describeKittyFlags(flags string) string {
	flagInt, _ := strconv.Atoi(flags)
	var descriptions []string
	if flagInt&1 != 0 {
		descriptions = append(descriptions, "Disambiguate escape codes")
	}
	if flagInt&2 != 0 {
		descriptions = append(descriptions, "Report event types")
	}
	if flagInt&4 != 0 {
		descriptions = append(descriptions, "Report alternate keys")
	}
	if flagInt&8 != 0 {
		descriptions = append(descriptions, "Report all keys as escape codes")
	}
	if flagInt&16 != 0 {
		descriptions = append(descriptions, "Report associated text")
	}
	return strings.Join(descriptions, ", ")
}

func describeKittyMode(mode string) string {
	switch mode {
	case "1":
		return "Set given flags and unset all others"
	case "2":
		return "Set given flags and keep existing flags unchanged"
	case "3":
		return "Unset given flags and keep existing flags unchanged"
	default:
		return "Unknown mode"
	}
}

func describeClipboard(c string) string {
	switch c {
	case "c":
		return "system"
	case "p":
		return "primary"
	default:
		return "unknown"
	}
}

func parseSGR(params string) []string {
	var r []string
	if params == "" {
		return []string{"CSI m: Reset all attributes"}
	}
	for _, param := range strings.Split(params, ";") {
		switch param {
		case "0":
			r = append(r, "CSI 0m: Reset all attributes")
		case "1":
			r = append(r, "CSI 1m: Set bold")
		case "2":
			r = append(r, "CSI 2m: Set faint")
		case "3":
			r = append(r, "CSI 3m: Set italic")
		case "4":
			r = append(r, "CSI 4m: Set underline")
		case "5":
			r = append(r, "CSI 5m: Set slow blink")
		case "6":
			r = append(r, "CSI 6m: Set rapid blink")
		case "7":
			r = append(r, "CSI 7m: Set reverse video")
		case "8":
			r = append(r, "CSI 8m: Set concealed")
		case "9":
			r = append(r, "CSI 9m: Set crossed-out")
		case "21":
			r = append(r, "CSI 21m: Set double underline")
		case "22":
			r = append(r, "CSI 22m: Reset bold and faint")
		case "23":
			r = append(r, "CSI 23m: Reset italic")
		case "24":
			r = append(r, "CSI 24m: Reset underline")
		case "25":
			r = append(r, "CSI 25m: Reset blink")
		case "27":
			r = append(r, "CSI 27m: Reset reverse video")
		case "28":
			r = append(r, "CSI 28m: Reset concealed")
		case "29":
			r = append(r, "CSI 29m: Reset crossed-out")
		default:
			if strings.HasPrefix(param, "3") && len(param) == 2 {
				r = append(r, fmt.Sprintf("CSI %sm: Set foreground color to %s", param, colorName(param[1])))
			} else if strings.HasPrefix(param, "4") && len(param) == 2 {
				r = append(r, fmt.Sprintf("CSI %sm: Set background color to %s", param, colorName(param[1])))
			} else if strings.HasPrefix(param, "38;5;") {
				r = append(r, fmt.Sprintf("CSI 38;5;%sm: Set foreground color to 8-bit color %s", param[5:], param[5:]))
			} else if strings.HasPrefix(param, "48;5;") {
				r = append(r, fmt.Sprintf("CSI 48;5;%sm: Set background color to 8-bit color %s", param[5:], param[5:]))
			} else if strings.HasPrefix(param, "38;2;") {
				parts := strings.Split(param, ";")
				if len(parts) == 5 {
					r = append(r, fmt.Sprintf("CSI 38;2;%s;%s;%sm: Set foreground color to RGB(%s,%s,%s)", parts[2], parts[3], parts[4], parts[2], parts[3], parts[4]))
				}
			} else if strings.HasPrefix(param, "48;2;") {
				parts := strings.Split(param, ";")
				if len(parts) == 5 {
					r = append(r, fmt.Sprintf("CSI 48;2;%s;%s;%sm: Set background color to RGB(%s,%s,%s)", parts[2], parts[3], parts[4], parts[2], parts[3], parts[4]))
				}
			} else {
				r = append(r, fmt.Sprintf("CSI %sm: Unknown SGR parameter", param))
			}
		}
	}
	return r
}

func colorName(c byte) string {
	switch c {
	case '0':
		return "Black"
	case '1':
		return "Red"
	case '2':
		return "Green"
	case '3':
		return "Yellow"
	case '4':
		return "Blue"
	case '5':
		return "Magenta"
	case '6':
		return "Cyan"
	case '7':
		return "White"
	default:
		return "Unknown"
	}
}
