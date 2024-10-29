package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
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

	}
}

func parse(seq string) []string {
	var r []string

	// Request Kitty Keyboard
	if is(seq, `\x1b\[?u`) {
		r = append(r, "Request Kitty Keyboard")
	}

	// Kitty Keyboard
	if ok, params := extract(seq, `\x1b\[=(\d+;\d+)u`); ok {
		parts := strings.Split(params, ";")
		if len(parts) == 2 {
			flags, mode := parts[0], parts[1]
			r = append(r, fmt.Sprintf("Kitty Keyboard: flags=%s (%s), mode=%s (%s)",
				flags, describeKittyFlags(flags),
				mode, describeKittyMode(mode)))
		}
	}

	// Push Kitty Keyboard
	if ok, flags := extract(seq, `\x1b\[>(\d*)u`); ok {
		if flags == "" {
			r = append(r, "Push Kitty Keyboard: flags=0 (Disable all features)")
		} else {
			r = append(r, fmt.Sprintf("Push Kitty Keyboard: flags=%s", flags))
		}
	}

	// Pop Kitty Keyboard
	if ok, n := extract(seq, `\x1b\[<(\d*)u`); ok {
		if n == "" {
			r = append(r, "Pop Kitty Keyboard: n=1")
		} else {
			r = append(r, fmt.Sprintf("Pop Kitty Keyboard: n=%s", n))
		}
	}

	// cursor down
	if ok, no := extract(seq, `\x1b\[(\d+)B`); ok {
		r = append(r, fmt.Sprintf("Cursor down %s lines", no))
	} else if is(seq, `\x1b\[B`) {
		r = append(r, "Cursor down")
	}

	// cursor up
	if ok, no := extract(seq, `\x1b\[(\d+)A`); ok {
		r = append(r, fmt.Sprintf("Cursor up %s lines", no))
	} else if is(seq, `\x1b\[A`) {
		r = append(r, "Cursor up")
	}

	// cursor right
	if ok, no := extract(seq, `\x1b\[(\d+)C`); ok {
		r = append(r, fmt.Sprintf("Cursor right %s columns", no))
	} else if is(seq, `\x1b\[C`) {
		r = append(r, "Cursor right")
	}

	// cursor left
	if ok, no := extract(seq, `\x1b\[(\d+)D`); ok {
		r = append(r, fmt.Sprintf("Cursor left %s columns", no))
	} else if is(seq, `\x1b\[D`) {
		r = append(r, "Cursor left")
	}

	// save cursor position
	if is(seq, `\x1b7`) {
		r = append(r, "Save cursor position")
	}

	// restore cursor position
	if is(seq, `\x1b8`) {
		r = append(r, "Restore cursor position")
	}

	// request cursor position
	if is(seq, `\x1b\[6n`) {
		r = append(r, "Request cursor position")
	}

	// request extended cursor position
	if is(seq, `\x1b\[?6n`) {
		fmt.Println("Request extended cursor position")
	}

	// move cursor to upper left corner (origin)
	if is(seq, `\x1b\[1;1H`) {
		r = append(r, "Move cursor to upper left corner (origin)")
	}

	// save cursor position (CSI s)
	if is(seq, `\x1b\[s`) {
		r = append(r, "Save cursor position")
	}

	// restore cursor position (CSI u)
	if is(seq, `\x1b\[u`) {
		r = append(r, "Restore cursor position")
	}

	// set cursor style
	if ok, style := extract(seq, `\x1b\[(\d+) q`); ok {
		r = append(r, fmt.Sprintf("Set cursor style: %s", cursorStyle(style)))
	}

	// set pointer shape
	if ok, shape := extract(seq, `\x1b\]22;(.+)\x07`); ok {
		r = append(r, fmt.Sprintf("Set pointer shape: %s", shape))
	}

	// clipboard operations
	if ok, params := extract(seq, `\x1b\]52;([cp]);(.*)\x07`); ok {
		parts := strings.SplitN(params, ";", 2)
		if len(parts) == 2 {
			clipboardName := parts[0]
			data := parts[1]

			switch data {
			case "":
				r = append(r, fmt.Sprintf("Reset %s clipboard", describeClipboard(clipboardName)))
			case "?":
				r = append(r, fmt.Sprintf("Request %s clipboard", describeClipboard(clipboardName)))
			default:
				decodedData, err := base64.StdEncoding.DecodeString(data)
				if err != nil {
					r = append(r, fmt.Sprintf("Set %s clipboard: Invalid base64 data", describeClipboard(clipboardName)))
				} else {
					r = append(r, fmt.Sprintf("Set %s clipboard: %s", describeClipboard(clipboardName), string(decodedData)))
				}
			}
		}
	}

	// erase display
	if ok, n := extract(seq, `\x1b\[(\d*)J`); ok {
		r = append(r, fmt.Sprintf("Erase display: %s", eraseDisplayDescription(n)))
	}

	// erase line
	if ok, n := extract(seq, `\x1b\[(\d*)K`); ok {
		r = append(r, fmt.Sprintf("Erase line: %s", eraseLineDescription(n)))
	}

	// scroll up
	if ok, n := extract(seq, `\x1b\[(\d*)S`); ok {
		r = append(r, fmt.Sprintf("Scroll up: %s lines", defaultOne(n)))
	}

	// scroll down
	if ok, n := extract(seq, `\x1b\[(\d*)T`); ok {
		r = append(r, fmt.Sprintf("Scroll down: %s lines", defaultOne(n)))
	}

	// insert line
	if ok, n := extract(seq, `\x1b\[(\d*)L`); ok {
		r = append(r, fmt.Sprintf("Insert %s blank line(s)", defaultOne(n)))
	}

	// delete line
	if ok, n := extract(seq, `\x1b\[(\d*)M`); ok {
		r = append(r, fmt.Sprintf("Delete %s line(s)", defaultOne(n)))
	}

	// set scrolling region
	if ok, params := extract(seq, `\x1b\[(\d*;\d*)r`); ok {
		parts := strings.Split(params, ";")
		if len(parts) == 2 {
			r = append(r, fmt.Sprintf("Set scrolling region: top=%s, bottom=%s", parts[0], parts[1]))
		}
	}

	// hyperlink
	if ok, params := extract(seq, `\x1b\]8;(.*);(.*)\x07`); ok {
		parts := strings.SplitN(params, ";", 2)
		if len(parts) == 2 {
			attributes, uri := parts[0], parts[1]
			if uri == "" {
				r = append(r, fmt.Sprintf("Reset hyperlink (attributes: %s)", attributes))
			} else {
				r = append(r, fmt.Sprintf("Set hyperlink: URI=%s, attributes=%s", uri, attributes))
			}
		}
	}

	// set foreground color
	if ok, color := extract(seq, `\x1b\]10;(.*)\x07`); ok {
		if color == "?" {
			r = append(r, "Request foreground color")
		} else {
			r = append(r, fmt.Sprintf("Set foreground color: %s", color))
		}
	}

	// reset foreground color
	if is(seq, `\x1b\]110\x07`) {
		r = append(r, "Reset foreground color")
	}

	// set background color
	if ok, color := extract(seq, `\x1b\]11;(.*)\x07`); ok {
		if color == "?" {
			r = append(r, "Request background color")
		} else {
			r = append(r, fmt.Sprintf("Set background color: %s", color))
		}
	}

	// reset background color
	if is(seq, `\x1b\]111\x07`) {
		r = append(r, "Reset background color")
	}

	// set cursor color
	if ok, color := extract(seq, `\x1b\]12;(.*)\x07`); ok {
		if color == "?" {
			r = append(r, "Request cursor color")
		} else {
			r = append(r, fmt.Sprintf("Set cursor color: %s", color))
		}
	}

	// reset cursor color
	if is(seq, `\x1b\]112\x07`) {
		r = append(r, "Reset cursor color")
	}

	// set icon name and window title
	if ok, title := extract(seq, `\x1b\]0;(.*)\x07`); ok {
		r = append(r, fmt.Sprintf("Set icon name and window title: %s", title))
	}

	// set icon name
	if ok, name := extract(seq, `\x1b\]1;(.*)\x07`); ok {
		r = append(r, fmt.Sprintf("Set icon name: %s", name))
	}

	// set window title
	if ok, title := extract(seq, `\x1b\]2;(.*)\x07`); ok {
		r = append(r, fmt.Sprintf("Set window title: %s", title))
	}

	// Enable/Disable Cursor Keys
	if is(seq, `\x1b\[?1h`) {
		r = append(r, "Enable Cursor Keys")
	} else if is(seq, `\x1b\[?1l`) {
		r = append(r, "Disable Cursor Keys")
	} else if is(seq, `\x1b\[?1\$p`) {
		r = append(r, "Request Cursor Keys")
	}

	// Show/Hide Cursor
	if is(seq, `\x1b\[?25h`) {
		r = append(r, "Show Cursor")
	} else if is(seq, `\x1b\[?25l`) {
		r = append(r, "Hide Cursor")
	} else if is(seq, `\x1b\[?25\$p`) {
		r = append(r, "Request Cursor Visibility")
	}

	// Enable/Disable Mouse
	if is(seq, `\x1b\[?1000h`) {
		r = append(r, "Enable Mouse")
	} else if is(seq, `\x1b\[?1000l`) {
		r = append(r, "Disable Mouse")
	} else if is(seq, `\x1b\[?1000\$p`) {
		r = append(r, "Request Mouse")
	}

	// Enable/Disable Mouse Hilite
	if is(seq, `\x1b\[?1001h`) {
		r = append(r, "Enable Mouse Hilite")
	} else if is(seq, `\x1b\[?1001l`) {
		r = append(r, "Disable Mouse Hilite")
	} else if is(seq, `\x1b\[?1001\$p`) {
		r = append(r, "Request Mouse Hilite")
	}

	// Enable/Disable Mouse Cell Motion
	if is(seq, `\x1b\[?1002h`) {
		r = append(r, "Enable Mouse Cell Motion")
	} else if is(seq, `\x1b\[?1002l`) {
		r = append(r, "Disable Mouse Cell Motion")
	} else if is(seq, `\x1b\[?1002\$p`) {
		r = append(r, "Request Mouse Cell Motion")
	}

	// Enable/Disable Mouse All Motion
	if is(seq, `\x1b\[?1003h`) {
		r = append(r, "Enable Mouse All Motion")
	} else if is(seq, `\x1b\[?1003l`) {
		r = append(r, "Disable Mouse All Motion")
	} else if is(seq, `\x1b\[?1003\$p`) {
		r = append(r, "Request Mouse All Motion")
	}

	// Enable/Disable Report Focus
	if is(seq, `\x1b\[?1004h`) {
		r = append(r, "Enable Report Focus")
	} else if is(seq, `\x1b\[?1004l`) {
		r = append(r, "Disable Report Focus")
	} else if is(seq, `\x1b\[?1004\$p`) {
		r = append(r, "Request Report Focus")
	}

	// Enable/Disable Mouse SGR Ext
	if is(seq, `\x1b\[?1006h`) {
		r = append(r, "Enable Mouse SGR Ext")
	} else if is(seq, `\x1b\[?1006l`) {
		r = append(r, "Disable Mouse SGR Ext")
	} else if is(seq, `\x1b\[?1006\$p`) {
		r = append(r, "Request Mouse SGR Ext")
	}

	// Enable/Disable Alt Screen Buffer
	if is(seq, `\x1b\[?1049h`) {
		r = append(r, "Enable Alt Screen Buffer")
	} else if is(seq, `\x1b\[?1049l`) {
		r = append(r, "Disable Alt Screen Buffer")
	} else if is(seq, `\x1b\[?1049\$p`) {
		r = append(r, "Request Alt Screen Buffer")
	}

	// Enable/Disable Bracketed Paste
	if is(seq, `\x1b\[?2004h`) {
		r = append(r, "Enable Bracketed Paste")
	} else if is(seq, `\x1b\[?2004l`) {
		r = append(r, "Disable Bracketed Paste")
	} else if is(seq, `\x1b\[?2004\$p`) {
		r = append(r, "Request Bracketed Paste")
	}

	// Enable/Disable Syncd Output
	if is(seq, `\x1b\[?2026h`) {
		r = append(r, "Enable Syncd Output")
	} else if is(seq, `\x1b\[?2026l`) {
		r = append(r, "Disable Syncd Output")
	} else if is(seq, `\x1b\[?2026\$p`) {
		r = append(r, "Request Syncd Output")
	}

	// Enable/Disable Grapheme Clustering
	if is(seq, `\x1b\[?2027h`) {
		r = append(r, "Enable Grapheme Clustering")
	} else if is(seq, `\x1b\[?2027l`) {
		r = append(r, "Disable Grapheme Clustering")
	} else if is(seq, `\x1b\[?2027\$p`) {
		r = append(r, "Request Grapheme Clustering")
	}

	// Enable/Disable Win32 Input
	if is(seq, `\x1b\[?9001h`) {
		r = append(r, "Enable Win32 Input")
	} else if is(seq, `\x1b\[?9001l`) {
		r = append(r, "Disable Win32 Input")
	} else if is(seq, `\x1b\[?9001\$p`) {
		r = append(r, "Request Win32 Input")
	}

	// Reset Style
	if is(seq, `\x1b\[m`) {
		r = append(r, "Reset Style")
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
