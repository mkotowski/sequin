package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

func handleTermcap(p *ansi.Parser) {
	if p.ParamsLen != 0 || p.DataLen == 0 {
		return
	}

	parts := bytes.Split(p.Data[:p.DataLen], []byte{';'})

	var caps []string
	for _, part := range parts {
		capName, err := hex.DecodeString(string(part))
		if err != nil {
			return
		}
		caps = append(caps, string(capName))
	}

	fmt.Printf("Request termcap entry for %q", strings.Join(caps, ", "))
}
