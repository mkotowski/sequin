package main

import (
	"fmt"

	"github.com/charmbracelet/x/ansi"
)

func handleScreen(cmd int) func(*ansi.Parser) {
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
		case 'J':
			fmt.Printf(must([]string{
				"Erase screen above",
				"Erase screen bellow",
				"Erase entire screen",
				"Erase entire display",
			}, count))
		}
	}
}
