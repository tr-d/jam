// +build !pretty

package main

import (
	"fmt"
	"os"
)

func banner() {
	var b, t string
	switch arg0 {
	case "jam", "jam.exe":
		b = `
  	     ██    █████    ███    ███
  	     ██   ██   ██   ████  ████
  	     ██   ███████   ██ ████ ██
  	██   ██   ██   ██   ██  ██  ██
  	 █████    ██   ██   ██      ██
`
		t = "  If structured data is scones and you are clotted cream, this is jam."
	default:
		b = `
  	██    ██████     ██   ██                 █████
  	██    ██   ██    ██  ██                      ██
  	██    ██   ██    █████                     ███
  	██    ██   ██    ██  ██                                         
  	██    ██████     ██   ██   ██   ██   ██   ██
`
		t = "  If it's not called jam I don't know what to say."
	}
	fmt.Fprint(os.Stderr, b)
	fmt.Fprintln(os.Stderr, t)
}

func halpsPrint(s string) {
	fmt.Fprintf(os.Stderr, "Halps %[1]s:\n", arg0)
	banner()
	fmt.Fprintf(os.Stderr, s, arg0)
}

func versionPrint() {
	fmt.Println(arg0 + "-" + version)
}
