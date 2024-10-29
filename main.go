package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
)

const (
	markerShift   = parser.MarkerShift
	intermedShift = parser.IntermedShift
)

var buf bytes.Buffer

func flushPrint() {
	if buf.Len() == 0 {
		return
	}
	fmt.Printf("Text: %q\n", buf.String())
	buf.Reset()
}

func main() {
	in := strings.Join(os.Args[1:], " ")
	if in == "-" || in == "" {
		bts, err := io.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}
		in = string(bts)
	}

	var state byte
	p := ansi.GetParser()
	defer ansi.PutParser(p)

	for len(in) > 0 {
		seq, width, n, newState := ansi.DecodeSequence(in, state, p)

		switch {
		case ansi.HasCsiPrefix(seq):
			flushPrint()
			fmt.Printf("CSI %q: ", seq)

			handler, ok := csiHandlers[p.Cmd]
			if ok {
				handler(p)
			}

			fmt.Println()

		case ansi.HasDcsPrefix(seq):
			// TODO: add common DCS handlers
			flushPrint()

		case ansi.HasOscPrefix(seq):
			flushPrint()
			fmt.Printf("OSC %q: ", seq)
			handler, ok := oscHandlers[p.Cmd]
			if ok {
				handler(p)
			}

			fmt.Println()

		case ansi.HasApcPrefix(seq):
			flushPrint()
			fmt.Printf("APC %q", seq)

			switch {
			case ansi.HasPrefix(p.Data, []byte("G")):
				// Kitty graphics
			}

			fmt.Println()

		case ansi.HasEscPrefix(seq):
			flushPrint()

			if len(seq) == 1 {
				// just an ESC
				fmt.Println("Control code ESC")
				break
			}

			fmt.Printf("ESC: %q", seq)
			switch p.Cmd {
			case 7:
				// save cursor
			case 8:
				// restore cursor
			}

			fmt.Println()

		case width == 0 && len(seq) == 1:
			flushPrint()
			// control code
			fmt.Printf("Control code %q: %s\n", seq, ctrlCodes[seq[0]])

		case width > 0:
			// Text
			buf.WriteString(seq)

		default:
			flushPrint()
			fmt.Printf("Unknown %q\n", seq)
		}

		in = in[n:]
		state = newState
	}

	flushPrint()
}

var ctrlCodes = map[byte]string{
	0:  "null",
	1:  "start of heading",
	2:  "start of text",
	3:  "end of text",
	4:  "end of transmission",
	5:  "enquiry",
	6:  "acknowledge",
	7:  "bell",
	8:  "backspace",
	9:  "horizontal tab",
	10: "line feed",
	11: "vertical tab",
	12: "form feed",
	13: "carriage return",
	14: "shift out",
	15: "shift in",
	16: "data link escape",
	17: "device control 1",
	18: "device control 2",
	19: "device control 3",
	20: "device control 4",
	21: "negative acknowledge",
	22: "synchronous idle",
	23: "end of transmission block",
	24: "cancel",
	25: "end of medium",
	26: "substitute",
	27: "escape",
	28: "file separator",
	29: "group separator",
	30: "record separator",
	31: "unit separator",
}
