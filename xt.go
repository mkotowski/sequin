package main

import (
	"fmt"

	"github.com/charmbracelet/x/ansi"
)

func handleXT(parser *ansi.Parser) {
	var count int
	if parser.ParamsLen > 0 {
		count = ansi.Param(parser.Params[0]).Param()
	}

	if count != 0 {
		fmt.Printf("unknown")
	}

	fmt.Printf("Request XT Version")
}
