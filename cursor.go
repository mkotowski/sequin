package main

import (
	"fmt"

	"github.com/charmbracelet/x/ansi"
)

func handleCursor(cmd int, isPrivate bool) func(*ansi.Parser) {
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
		case 'A':
			// CUU - Cursor Up
			fmt.Printf("Cursor up %d", count)
		case 'B':
			// CUD - Cursor Down
			fmt.Printf("Cursor down %d", count)
		case 'C':
			// CUF - Cursor Forward
			fmt.Printf("Cursor right %d", count)
		case 'D':
			// CUB - Cursor Back
			fmt.Printf("Cursor left %d", count)
		case 'E':
			fmt.Printf("Cursor next line %d", count)
		case 'F':
			fmt.Printf("Cursor previous line %d", count)
		case 'H':
			row := count
			col := 1
			if parser.ParamsLen > 1 {
				col = parser.Params[1]
			}
			fmt.Printf("Set cursor position row=%[1]d col=%[2]d", row, col)
		case 'n':
			if count != 6 {
				fmt.Printf("unknown")
				return
			}
			if isPrivate {
				fmt.Printf("Request extended cursor position")
			} else {
				fmt.Printf("Request cursor position")
			}
		case 's':
			fmt.Printf("Save cursor position")
		}
	}
}
