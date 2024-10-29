package main

import (
	"fmt"

	"github.com/charmbracelet/x/ansi"
)

func handleReqMode(cmd int, isPrivate bool) func(*ansi.Parser) {
	// DECRQM - Request Mode
	return func(parser *ansi.Parser) {
		if parser.ParamsLen == 0 {
			// Invalid, ignore
			return
		}

		mode := modeDesc(ansi.Param(parser.Params[0]).Param())
		switch cmd {
		case 'p':
			if isPrivate {
				fmt.Printf("Request private mode %q", mode)
			} else {
				fmt.Printf("Request mode %q", mode)
			}
		case 'h':
			if isPrivate {
				fmt.Printf("Enable private mode %q", mode)
			} else {
				fmt.Printf("Enable mode %q", mode)
			}
		case 'l':
			if isPrivate {
				fmt.Printf("Disable private mode %q", mode)
			} else {
				fmt.Printf("Disable mode %q", mode)
			}
		}
	}
}

func modeDesc(mode int) string {
	switch mode {
	case 1:
		return "cursor keys"
	case 25:
		return "cursor visibility"
	case 1000:
		return "show mouse"
	case 1001:
		return "mouse hilite"
	case 1002:
		return "mouse cell motion"
	case 1003:
		return "mouse all motion"
	case 1004:
		return "report focus"
	case 1006:
		return "mouse SGR ext"
	case 1049:
		return "altscreen"
	case 2004:
		return "bracketed paste"
	case 2026:
		return "synchronized output"
	case 2027:
		return "grapheme clustering"
	case 9001:
		return "win32 input"
	default:
		return "unknown"
	}
}
