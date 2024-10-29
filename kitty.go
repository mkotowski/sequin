package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

func handleKitty(parser *ansi.Parser) {
	flagDesc := func(flag int) string {
		var r []string
		if flag&1 != 0 {
			r = append(r, "Disambiguate escape codes")
		}
		if flag&2 != 0 {
			r = append(r, "Report event types")
		}
		if flag&4 != 0 {
			r = append(r, "Report alternate keys")
		}
		if flag&8 != 0 {
			r = append(r, "Report all keys as escape codes")
		}
		if flag&16 != 0 {
			r = append(r, "Report associated text")
		}
		return strings.Join(r, ", ")
	}
	modeDesc := func(mode int) string {
		switch mode {
		case 1:
			return "Set given flags and unset all others"
		case 2:
			return "Set given flags and keep existing flags unchanged"
		case 3:
			return "Unset given flags and keep existing flags unchanged"
		default:
			return "Unknown mode"
		}
	}

	var first int
	if parser.ParamsLen > 0 {
		first = ansi.Param(parser.Params[0]).Param()
	}

	cmd := ansi.Cmd(parser.Cmd)
	switch cmd.Marker() {
	case '?':
		fmt.Printf("Request Kitty keyboard")
	case '>':
		if first == 0 {
			fmt.Printf("Disable Kitty keyboard")
		} else {
			fmt.Printf("Push %q Kitty keyboard flag", flagDesc(first))
		}
	case '<':
		fmt.Printf("Pop %d Kitty keyboard flags", first)
	case '=':
		if parser.ParamsLen > 1 {
			second := ansi.Param(parser.Params[1]).Param()
			fmt.Printf("Set %q Kitty keyboard flags to %q", flagDesc(first), modeDesc(second))
		}
	}
}
