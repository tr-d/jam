// +build pretty

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

const prettyVersion = "-pretty"

func banner() {
	var b, t string
	switch filepath.Base(arg0) {
	case "jam", "jam.exe":
		b = `
  	     ██    █████    ███    ███
  	     ██   ██   ██   ████  ████
  	     ██   ███████   ██ ████ ██
  	██   ██   ██   ██   ██  ██  ██
  	 █████    ██   ██   ██      ██  pretty version
`
		t = "  If structured data is scones and you are clotted cream, this is jam."
	default:
		b = `
  	██    ██████     ██   ██                 █████
  	██    ██   ██    ██  ██                      ██
  	██    ██   ██    █████                     ███
  	██    ██   ██    ██  ██    still pretty tho
  	██    ██████     ██   ██   ██   ██   ██   ██
`
		t = "  If it's not called jam I don't know what to say."
	}
	if !terminal.IsTerminal(int(os.Stderr.Fd())) {
		fmt.Fprintln(os.Stderr, b)
		fmt.Fprintln(os.Stderr, t)
		return
	}
	fmt.Fprintln(os.Stderr, "\x1b[38;2;255;175;199m"+b+"\x1b[0m")
	fmt.Fprintln(os.Stderr, "\x1b[38;2;251;249;245m"+t+"\x1b[0m")
}

func halpsPrint(s string) {
	fmt.Fprintf(os.Stderr, "Halps %[1]s:\n", arg0)
	banner()
	for _, l := range strings.Split(fmt.Sprintf(s, arg0), "\n") {
		switch {
		case l == "", l[0] == ' ', l[len(l)-1] != ':':
			fmt.Fprintln(os.Stderr, l)
		default:
			fmt.Fprintln(os.Stderr, "\x1b[38;2;183;232;240m"+l+"\x1b[0m")
		}
	}
}
