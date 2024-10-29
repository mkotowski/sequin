package main

import (
	"fmt"

	"github.com/charmbracelet/x/ansi"
)

func handleLine(cmd int) func(*ansi.Parser) {
	return func(parser *ansi.Parser) {
		var count int
		if parser.ParamsLen > 0 {
			count = ansi.Param(parser.Params[0]).Param()
		}

		if count == 0 {
			// Default value is 1
			count = 1
		}

		switch cmd {
		case 'K':
			fmt.Printf(must([]string{
				"Erase line right",
				"Erase line left",
				"Erase entire line",
			}, count))
		case 'L':
			fmt.Printf("CSI %d L: Insert %[1]d blank lines", count)
		case 'M':
			fmt.Printf("CSI %d M: Delete %[1]d lines", count)
		case 'S':
			fmt.Printf("CSI %d S: Scroll up %[1]d lines", count)
		case 'T':
			fmt.Printf("CSI %d T: Scroll down %[1]d lines", count)
		}
	}
}

func must(ss []string, i int) string {
	if len(ss) <= i {
		return "invalid"
	}
	return ss[i]
}
