package main

import (
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

		// Request Kitty Keyboard
		if is(seq, `\x1b\[?u`) {
			fmt.Println("Request Kitty Keyboard")
		}

		// Kitty Keyboard
		if ok, params := extract(seq, `\x1b\[=(\d+;\d+)u`); ok {
			parts := strings.Split(params, ";")
			if len(parts) == 2 {
				flags, mode := parts[0], parts[1]
				fmt.Printf("Kitty Keyboard: flags=%s (%s), mode=%s (%s)\n",
					flags, describeKittyFlags(flags),
					mode, describeKittyMode(mode))
			}
		}

		// Push Kitty Keyboard
		if ok, flags := extract(seq, `\x1b\[>(\d*)u`); ok {
			if flags == "" {
				fmt.Println("Push Kitty Keyboard: flags=0 (Disable all features)")
			} else {
				fmt.Printf("Push Kitty Keyboard: flags=%s\n", flags)
			}
		}

		// Pop Kitty Keyboard
		if ok, n := extract(seq, `\x1b\[<(\d*)u`); ok {
			if n == "" {
				fmt.Println("Pop Kitty Keyboard: n=1")
			} else {
				fmt.Printf("Pop Kitty Keyboard: n=%s\n", n)
			}
		}

		// cursor down
		if ok, no := extract(seq, `\x1b\[(\d+)B`); ok {
			fmt.Printf("Cursor down %s lines\n", no)
		} else if is(seq, `\x1b\[B`) {
			fmt.Println("Cursor down")
		}

		// cursor up
		if ok, no := extract(seq, `\x1b\[(\d+)A`); ok {
			fmt.Printf("Cursor up %s lines\n", no)
		} else if is(seq, `\x1b\[A`) {
			fmt.Println("Cursor up")
		}

		// cursor right
		if ok, no := extract(seq, `\x1b\[(\d+)C`); ok {
			fmt.Printf("Cursor right %s columns\n", no)
		} else if is(seq, `\x1b\[C`) {
			fmt.Println("Cursor right")
		}

		// cursor left
		if ok, no := extract(seq, `\x1b\[(\d+)D`); ok {
			fmt.Printf("Cursor left %s columns\n", no)
		} else if is(seq, `\x1b\[D`) {
			fmt.Println("Cursor left")
		}

		// save cursor position
		if is(seq, `\x1b7`) {
			fmt.Println("Save cursor position")
		}

		// restore cursor position
		if is(seq, `\x1b8`) {
			fmt.Println("Restore cursor position")
		}

		// request cursor position
		if is(seq, `\x1b\[6n`) {
			fmt.Println("Request cursor position")
		}

		// request extended cursor position
		if is(seq, `\x1b\[?6n`) {
			fmt.Println("Request extended cursor position")
		}

		// move cursor to upper left corner (origin)
		if is(seq, `\x1b\[1;1H`) {
			fmt.Println("Move cursor to upper left corner (origin)")
		}

		// save cursor position (CSI s)
		if is(seq, `\x1b\[s`) {
			fmt.Println("Save cursor position")
		}

		// restore cursor position (CSI u)
		if is(seq, `\x1b\[u`) {
			fmt.Println("Restore cursor position")
		}

		// set cursor style
		if ok, style := extract(seq, `\x1b\[(\d+) q`); ok {
			fmt.Printf("Set cursor style: %s\n", cursorStyle(style))
		}

		// set pointer shape
		if ok, shape := extract(seq, `\x1b\]22;(.+)\x07`); ok {
			fmt.Printf("Set pointer shape: %s\n", shape)
		}

		// erase display
		if ok, n := extract(seq, `\x1b\[(\d*)J`); ok {
			fmt.Printf("Erase display: %s\n", eraseDisplayDescription(n))
		}

		// erase line
		if ok, n := extract(seq, `\x1b\[(\d*)K`); ok {
			fmt.Printf("Erase line: %s\n", eraseLineDescription(n))
		}

		// scroll up
		if ok, n := extract(seq, `\x1b\[(\d*)S`); ok {
			fmt.Printf("Scroll up: %s lines\n", defaultOne(n))
		}

		// scroll down
		if ok, n := extract(seq, `\x1b\[(\d*)T`); ok {
			fmt.Printf("Scroll down: %s lines\n", defaultOne(n))
		}

		// insert line
		if ok, n := extract(seq, `\x1b\[(\d*)L`); ok {
			fmt.Printf("Insert %s blank line(s)\n", defaultOne(n))
		}

		// delete line
		if ok, n := extract(seq, `\x1b\[(\d*)M`); ok {
			fmt.Printf("Delete %s line(s)\n", defaultOne(n))
		}

		// set scrolling region
		if ok, params := extract(seq, `\x1b\[(\d*;\d*)r`); ok {
			parts := strings.Split(params, ";")
			if len(parts) == 2 {
				fmt.Printf("Set scrolling region: top=%s, bottom=%s\n", parts[0], parts[1])
			}
		}
	}
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
