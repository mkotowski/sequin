package main

import (
	"fmt"

	"github.com/charmbracelet/x/ansi"
)

var csiHandlers = map[int]func(*ansi.Parser){
	'm':                      handleSgr,
	'A':                      handleCursor('A'),
	'B':                      handleCursor('B'),
	'C':                      handleCursor('C'),
	'D':                      handleCursor('D'),
	'E':                      handleCursor('E'),
	'F':                      handleCursor('F'),
	'H':                      handleCursor('H'),
	'J':                      handleScreen('J'),
	'K':                      handleLine('K'),
	'L':                      handleLine('L'),
	'M':                      handleLine('M'),
	'S':                      handleLine('S'),
	'T':                      handleLine('T'),
	'c':                      printf("Request primary device attributes"),
	'p' | '$'<<intermedShift: handleReqMode(false),
	'p' | '?'<<markerShift | '$'<<intermedShift: handleReqMode(true),
}

var oscHandlers = map[int]func(*ansi.Parser){
	0:   handleTitle,
	1:   handleTitle,
	2:   handleTitle,
	8:   handleHyperlink,
	9:   handleNotify,
	10:  handleTerminalColor,
	11:  handleTerminalColor,
	12:  handleTerminalColor,
	22:  handlePointerShape,
	52:  handleClipboard,
	110: handleResetTerminalColor,
	111: handleResetTerminalColor,
	112: handleResetTerminalColor,
}

func printf(s string) func(*ansi.Parser) {
	return func(*ansi.Parser) {
		fmt.Printf(s)
	}
}
