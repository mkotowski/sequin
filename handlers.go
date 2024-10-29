package main

import (
	"fmt"

	"github.com/charmbracelet/x/ansi"
)

var csiHandlers = map[int]func(*ansi.Parser){
	'm':                      handleSgr,
	'A':                      handleCursor('A', false),
	'B':                      handleCursor('B', false),
	'C':                      handleCursor('C', false),
	'D':                      handleCursor('D', false),
	'E':                      handleCursor('E', false),
	'F':                      handleCursor('F', false),
	'H':                      handleCursor('H', false),
	'J':                      handleScreen('J'),
	'K':                      handleLine('K'),
	'L':                      handleLine('L'),
	'M':                      handleLine('M'),
	'S':                      handleLine('S'),
	'T':                      handleLine('T'),
	'c':                      printf("Request primary device attributes"),
	'p' | '$'<<intermedShift: handleReqMode('p', false),
	'p' | '?'<<markerShift | '$'<<intermedShift: handleReqMode('p', true),
	'h' | '?'<<markerShift:                      handleReqMode('h', true),
	'l' | '?'<<markerShift:                      handleReqMode('l', true),
	'h':                                         handleReqMode('h', false),
	'l':                                         handleReqMode('l', false),
	'n' | '?'<<markerShift:                      handleCursor('n', true),
	'n':                                         handleCursor('n', false),
	'q' | '>'<<markerShift:                      handleXT,
	'r':                                         handleScreen('r'),
	's':                                         handleCursor('s', false),
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
