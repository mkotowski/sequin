package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

func main() {
	s := strings.Join(os.Args[1:], " ")
	if s == "-" || s == "" {
		bts, err := io.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}
		s = string(bts)
	}

	// Split the input string into individual escape sequences
	sequences := strings.Split(s, "\x1b")

	for _, seq := range sequences {
		if seq == "" {
			continue
		}

		// Add the escape character back to the sequence
		seq = "\x1b" + seq

		// cursor down
		if ok, no := extract(seq, `\x1b\[(\d+)B`); ok {
			fmt.Printf("Cursor down %s lines\n", no)
		} else if is(seq, `\x1b\[B`) {
			fmt.Println("Cursor down")
		}

		// cursor up
		if ok, no := extract(seq, `\x1b\[(\d+)A`); ok {
			fmt.Printf("Cursor up %s lines\n", no)
		} else if is(seq, `\x1b\[A`) {
			fmt.Println("Cursor up")
		}

		// cursor right
		if ok, no := extract(seq, `\x1b\[(\d+)C`); ok {
			fmt.Printf("Cursor right %s columns\n", no)
		} else if is(seq, `\x1b\[C`) {
			fmt.Println("Cursor right")
		}

		// cursor left
		if ok, no := extract(seq, `\x1b\[(\d+)D`); ok {
			fmt.Printf("Cursor left %s columns\n", no)
		} else if is(seq, `\x1b\[D`) {
			fmt.Println("Cursor left")
		}
	}
}

func extract(s, pattern string) (bool, string) {
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(s)
	if len(match) < 1 {
		return false, ""
	}
	return true, match[1]
}

func is(s, pattern string) bool {
	match, err := regexp.MatchString(pattern, s)
	if err != nil {
		panic(err)
	}
	return match
}
